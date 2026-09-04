package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/results"
)

// runCLI — campagne en ligne de commande : la même matrice que le dashboard,
// imprimée en une ligne de progression par frame, arrêt gracieux CTRL-C
// (gel propre), résumé final depuis le CSV gelé. L'opérateur sans navigateur
// peut campagner par SSH.
func runCLI(profilesStr string, qdiscsStr, ccsStr string, reps, deadlineMs int, target, direction, dataDir string) int {
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

	lastLine := func(s string) {
		fmt.Printf("\r\033[K%s", s)
	}
	switch direction {
	case "", "up", "down", "both":
	default:
		fmt.Fprintln(os.Stderr, "run: --direction doit être up|down|both")
		return 2
	}
	deps := campagne.ProdDeps()
	if direction == "down" {
		deps = campagne.ProdDepsDown()
	}
	deps.Direction = direction
	deps.Target = target
	deps.DeadlineMs = float64(deadlineMs)

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
		cs := make([]string, len(model.AllCC))
		for i, c := range model.AllCC {
			cs[i] = string(c)
		}
		ccs = cs
	}

	m, err := campagne.StartMatrixFiltered(ctx, profiles, qdiscs, ccs, reps, deps, dataDir)
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
			printSummary(dataDir)
			return 0
		case <-time.After(300 * time.Millisecond):
		}
	}
	fmt.Println()
	printSummary(dataDir)
	return 0
}

// printSummary — médianes gelées de la dernière matrice + verdict + hash.
func printSummary(dataDir string) {
	groups, _ := results.Scan(dataDir, "")
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
