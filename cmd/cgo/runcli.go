package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/capabilities"
	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/results"
)

// runCLI — campagne en ligne de commande : la même matrice que le dashboard,
// imprimée en une ligne de progression par frame, arrêt gracieux CTRL-C
// (gel propre), résumé final depuis le CSV gelé. L'opérateur sans navigateur
// peut campagner par SSH.
func runCLI(profilesStr string, qdiscsStr, ccsStr string, reps, deadlineMs, flows int, target, direction, dataDir, runID, bench, channelModel string) int {
	// Un --run-id hostile ne touche JAMAIS le disque : même garde que le
	// banc (un élément de chemin simple). Sans ça, le verrou fichier serait
	// écrit hors du dossier de données.
	if err := results.ValidateRunID(runID); runID != "" && err != nil {
		fmt.Fprintf(os.Stderr, "run: run-id %q invalide — un élément de chemin, sans separateur ni point, 128 octets max\n", runID)
		return 2
	}
	profiles := strings.Split(profilesStr, ",")
	for i := range profiles {
		profiles[i] = strings.TrimSpace(profiles[i])
	}
	splitAxis := func(s string) []string {
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	qdiscs := splitAxis(qdiscsStr)
	ccs := splitAxis(ccsStr)
	if (qdiscs == nil) != (ccs == nil) {
		fmt.Fprintln(os.Stderr, "run: --qdiscs et --cc se filtrent ensemble")
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Garde anti-double-launch (camp-final : deux `run` concurrents se sont
	// battus sur les mêmes qdiscs + aqm_eval.csv, 36 lignes pour 18 events).
	// Verrou fichier O_EXCL dans le dataDir : le second meurt avec un message.
	var lockPath string
	if dataDir != "" {
		lockPath = dataDir + "/." + runID + ".lock"
	} else {
		lockPath = ".cgo-run-" + runID + ".lock"
	}
	lockF, lockErr := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if lockErr != nil {
		fmt.Fprintf(os.Stderr, "run: %s déjà en cours (verrou %s) — ps count == 1 avant chaque launch\n", runID, lockPath)
		return 3
	}
	defer func() {
		lockF.Close()
		os.Remove(lockPath)
	}()

	lastLine := func(s string) {
		fmt.Printf("\r\033[K%s", s)
	}
	switch direction {
	case "", "up", "down", "both":
	default:
		fmt.Fprintln(os.Stderr, "run: --direction doit être up|down|both")
		return 2
	}
	// Sélection du banc : dualns = chemin de production (quatre sauts,
	// puits dans ns_server) ; legacy = reproduction du corpus gelé.
	if bench == "" {
		bench = os.Getenv("CGO_BENCH")
	}
	switch bench {
	case "", "legacy":
		bench = "legacy"
	case "dualns":
	default:
		fmt.Fprintln(os.Stderr, "run: --bench doit être legacy|dualns")
		return 2
	}
	var deps campagne.Deps
	if bench == "dualns" {
		deps = campagne.ProdDepsDualNS(direction)
	} else {
		deps = campagne.ProdDeps()
		if direction == "down" {
			deps = campagne.ProdDepsDown()
		}
	}
	deps.Bench = bench
	deps.ChannelModel = channelModel
	deps.Direction = direction
	deps.Target = target
	deps.DeadlineMs = float64(deadlineMs)
	if flows < 1 || flows > 8 {
		fmt.Fprintln(os.Stderr, "run: --flows doit être 1–8")
		return 2
	}
	deps.Flows = flows

	// Preflight banc (camp220 : matrice lancée sur banc mort = campagne à
	// zéros). Réutilise le script durci : check → up si besoin → re-check ;
	// échec persistant = on refuse de geler du vide (exit 4).
	if code := preflightTestbed(bench); code != 0 {
		return code
	}

	// progression : chaque instantané rafraîchit la ligne courante
	deps.OnSnap = func(s campagne.Snapshot) {
		gates := ""
		for _, g := range s.Gates {
			switch {
			case g == nil:
				gates += "·"
			case *g:
				gates += "✓"
			default:
				gates += "✗"
			}
		}
		lastLine(fmt.Sprintf("[%s %s/%s rep%d] %s — small p95 %.1fms — rtt %.1fms — G[%s] %d/%d",
			s.Profile, s.Qdisc, s.CC, s.Repetition, s.Phase,
			s.Smallp95Ms, s.RTTp95Ms, gates, s.EventID, s.TotalEvents))
	}

	if qdiscs == nil {
		qs := make([]string, len(model.AllQdiscs))
		for i, q := range model.AllQdiscs {
			qs[i] = string(q)
		}
		qdiscs = qs
	}
	if ccs == nil {
		ccs = capabilities.DefaultCCs()
	}

	var m *campagne.Matrix
	var err error
	if runID != "" {
		m, err = campagne.StartMatrixFilteredWithID(ctx, runID, profiles, qdiscs, ccs, reps, deps, dataDir)
	} else {
		m, err = campagne.StartMatrixFiltered(ctx, profiles, qdiscs, ccs, reps, deps, dataDir)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "run:", err)
		return 2
	}

	// attente active : CTRL-C → Stop gracieux, la matrice draine et gèle
	for m.IsRunning() {
		select {
		case <-ctx.Done():
			fmt.Println("\narrêt demandé — gel en cours…")
			m.Stop()
			for i := 0; i < 30 && m.IsRunning(); i++ {
				time.Sleep(200 * time.Millisecond)
			}
			fmt.Println("gel terminé.")
			printSummary(dataDir, runID)
			return 0
		case <-time.After(300 * time.Millisecond):
		}
	}
	fmt.Println()
	printSummary(dataDir, runID)
	return 0
}

// printSummary — médianes gelées de la matrice + verdict + hash. Filtrée au
// run exécuté : agréger tout dataDir mélange des échéances différentes
// (résumé P3-1500 pollué par le P3-600 historique — vu en prod).
// preflightTestbed — refuse de lancer une matrice sur un banc mort.
// check → up si besoin → re-check via le script durci (sudo -n). Script
// absent (poste non-banc) = avertissement, pas de blocage. 0 = banc prêt.
func preflightTestbed(bench string) int {
	wd, _ := os.Getwd()
	// Le pré-contrôle suit le banc choisi : dualns vérifie le banc durci
	// (cgo-bench-dualns.sh, fermé-échec sur les offloads), legacy vérifie
	// testbed.sh. Un dualns qui précontrôlerait le banc legacy validerait
	// la mauvaise topologie.
	script := filepath.Join(wd, "kit", "testbed.sh")
	if bench == "dualns" {
		script = filepath.Join(wd, "kit", "cgo-bench-dualns.sh")
	}
	if _, err := os.Stat(script); err != nil {
		fmt.Fprintf(os.Stderr, "run: %s absent — pré-contrôle banc sauté (poste non-banc ?)\n", script)
		return 0
	}
	run := func(arg string, timeout time.Duration) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "bash", script, arg)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}
	fmt.Println("run: pré-contrôle du banc de mesure…")
	if run("check", 90*time.Second) == nil {
		return 0
	}
	fmt.Println("run: banc non prêt — repose (up) puis re-contrôle…")
	if err := run("up", 180*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "run: %s up ÉCHEC — console VM : sudo bash %s up\n", filepath.Base(script), script)
		return 4
	}
	if err := run("check", 90*time.Second); err != nil {
		fmt.Fprintln(os.Stderr, "run: banc toujours non prêt après up — campagne REFUSÉE (pas de zéros gelés)")
		return 4
	}
	return 0
}

func printSummary(dataDir, runID string) {

	groups, _ := results.Scan(dataDir, runID)
	if len(groups) == 0 {
		fmt.Println("aucune ligne gelée.")
		return
	}
	fmt.Printf("\n%-8s %-12s %-6s %10s %10s %10s %10s\n", "profil", "qdisc", "cc", "small p95", "rtt p95", "goodput", "deadline")
	for _, g := range groups {
		mark := ""
		if g.Best {
			mark = " ★"
		}
		fmt.Printf("%-8s %-12s %-6s %9.1fms %9.1fms %8.1fMb %9.0f%%%s\n",
			g.Profile, g.Qdisc, g.CC, g.Smallp95Median, g.RTTp95Median, g.GoodputMedian, g.DeadlineMedian, mark)
	}
}
