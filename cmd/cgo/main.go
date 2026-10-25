// Meteolink (binaire cgo) — commandes principales:
//
//	cgo --serve                          dashboard + API (https://meteolink.dev:9090 par défaut, certificat local auto-signé)
//	cgo setup                            wizard première installation (tout guidé)
//	cgo kit <action>                     moteur de déploiement (doctor, scan, ensure, deploy…)
//	cgo audit --link-type T --site S --duration N
//	cgo run --profiles P1,P2 --reps 3    campagne CLI (progression ligne à ligne)
//	cgo tui                              tableau de bord terminal complet
//	cgo verify                           manifest integrity check
//	cgo figures                          régénère les SVG depuis les CSV gelés
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/doctor"
	"github.com/Realms4239/cgo/pkg/figures"
	"github.com/Realms4239/cgo/pkg/remediate"
)

// version — overridden at release time: -X main.version={{.Version}}.
// RÈGLE VERSION (source unique) : kit.KitVersion est canonique — `version`
// vaut KitVersion par défaut et les GUI comparent le dashboard distant à
// kit.KitVersion directement ; une release -X doit donc reprendre exactement
// le tag (= KitVersion), sinon CLI (`kit next`) et GUI divergent.
var version = kit.KitVersion

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "kit":
		os.Exit(runKit(os.Args[2:]))
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "tui", "top":
		os.Exit(runTUI(os.Args[2:]))
	case "--serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		def := os.Getenv("CGO_DASHBOARD__ADDR")
		if def == "" {
			def = "meteolink.dev:9090" // interface canonique (hosts → 127.0.0.1) ; --tls=false + 0.0.0.0 pour la VM
		}
		addr := fs.String("addr", def, "listen address")
		tlsOn := fs.Bool("tls", true, "HTTPS avec le certificat local meteolink.dev (false = HTTP brut, VM/systemd)")
		httpAddr := fs.String("http-addr", "127.0.0.1:9080", "redirection HTTP→HTTPS (vide = désactivée)")
		mode := fs.String("mode", "auto", "observe | full | auto (défaut: OS décide — Windows observe, Linux complet)")
		_ = fs.Parse(os.Args[2:])
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := runServer(ctx, *addr, doctor.Mode(*mode), *tlsOn, *httpAddr); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
	case "doctor":
		doctor.Print()
	case "version":
		fmt.Printf("meteolink (cgo) %s — %s/%s, mode %s\n", version, runtime.GOOS, runtime.GOARCH, doctor.Mode("auto"))
	case "shape":
		fs := flag.NewFlagSet("shape", flag.ExitOnError)
		restore := fs.Bool("restore", false, "supprime tous les qdiscs root posés par meteolink sur l'interface de façonnage")
		_ = fs.Parse(os.Args[2:])
		if !*restore {
			fmt.Fprintln(os.Stderr, "shape: utilisez --restore (le levier complet vit dans le tableau de bord)")
			os.Exit(2)
		}
		if runtime.GOOS == "windows" {
			fmt.Fprintln(os.Stderr, "shape: le façonnage nécessite Linux (mode observation sous Windows)")
			os.Exit(1)
		}
		deps := campagne.ProdDeps()
		if _, err := deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root"); err != nil {
			fmt.Println("shape: rien à restaurer (ou déjà propre)")
			return
		}
		fmt.Printf("shape: qdiscs root supprimés sur %s\n", deps.ShaperIf)
	case "service":
		if len(os.Args) < 3 || os.Args[2] != "install" {
			fmt.Fprintln(os.Stderr, "service: utilisez `cgo service install` (Linux, root)")
			os.Exit(2)
		}
		if runtime.GOOS != "linux" {
			fmt.Fprintln(os.Stderr, "service: systemd nécessite Linux")
			os.Exit(1)
		}
		const unit = `[Unit]
Description=Meteolink — banc d'audit et de comparaison AQM/BBR
After=network-online.target
Wants=network-online.target

[Service]
		ExecStart=/usr/local/bin/cgo --serve --addr 0.0.0.0:9090 --tls=false --http-addr=
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
`
		if err := os.WriteFile("/etc/systemd/system/meteolink.service", []byte(unit), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "service:", err, "(lancez en root)")
			os.Exit(1)
		}
		fmt.Println("service: /etc/systemd/system/meteolink.service écrit — puis:")
		fmt.Println("  systemctl daemon-reload && systemctl enable --now meteolink")
	case "figures":
		if err := figures.Generate("data/runs", "data/figures"); err != nil {
			fmt.Fprintln(os.Stderr, "figures:", err)
			os.Exit(1)
		}
		fmt.Println("figures générées dans data/figures")
	case "verify":
		ok, err := verifyManifests("data/runs")
		if err != nil {
			fmt.Fprintln(os.Stderr, "verify:", err)
			os.Exit(1)
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "verify: échec — manifeste incohérent")
			os.Exit(1)
		}
		fmt.Println("verify: ok — tous les manifests valides")
	case "testbedsrv":
		fs := flag.NewFlagSet("testbedsrv", flag.ExitOnError)
		httpAddr := fs.String("http", "10.200.0.1:8081", "small-object http listen")
		bulkAddr := fs.String("bulk", "10.200.0.1:5201", "bulk sink listen")
		_ = fs.Parse(os.Args[2:])
		if err := testbedSrvFn(*httpAddr, *bulkAddr); err != nil {
			fmt.Fprintln(os.Stderr, "testbedsrv:", err)
			os.Exit(1)
		}
	case "audit":
		fs := flag.NewFlagSet("audit", flag.ExitOnError)
		// Surface opérateur : quatre réglages suffisent à un audit terrain.
		linkType := fs.String("link-type", "5g", "link_type: fiber, 5g, 4g, vsat, other")
		site := fs.String("site", "Site X", "site / département")
		duration := fs.Int("duration", 60, "duration seconds (60 = 10/40/10)")
		direction := fs.String("direction", "up", "sens de charge : up | down | both")
		provider := fs.String("provider", "", "provider")
		flows := fs.Int("flows", 4, "flux de charge concurrents (1–8 ; 1 = régime historique)")
		// Avancé : overrides explicites de terrain. Par défaut, la détection
		// du banc (dualns → legacy → external) choisit cible et puits ; en
		// external il n'y a PAS de charge implicite — jamais de note sans
		// charge réelle.
		target := fs.String("target", "", "cible ping (défaut : détection banc)")
		small := fs.String("small", "", "URL petit objet (défaut : détection banc)")
		bulk := fs.String("bulk", "", "puits bulk host:port (défaut : détection banc ; vide en externe = pas de charge)")
		// Remediation needs to know which device to configure. These are
		// deliberately explicit: the tool will not guess an interface name and
		// emit a tc command against the wrong device.
		iface := fs.String("iface", "", "Linux egress device to shape in the remediation (e.g. eth0)")
		rosIface := fs.String("routeros-iface", "", "RouterOS interface for the remediation script (e.g. ether1-wan)")
		reportPath := fs.String("report", "", "additional report path (.md or .json); the standard data/audits/<id>/ artifacts are always written")
		qdisc := fs.String("qdisc", "", "qdisc for the remediation: cake, fq_codel, pfifo_fast (empty = recommended)")
		asymmetric := fs.Bool("asymmetric", true, "link is asymmetric (uplink << downlink); drives the qdisc recommendation")
		cpuBound := fs.Bool("cpu-constrained", false, "gateway is CPU-constrained; drives the qdisc recommendation")
		_ = fs.Parse(os.Args[2:])
		sinks := ResolveSinks()
		if *target == "" {
			*target = sinks.Target
		}
		if *small == "" {
			*small = sinks.SmallURL
		}
		if *bulk == "" {
			*bulk = sinks.envOrBulk()
		}
		p := audit.Params{
			AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()),
			Site:    *site, LinkType: *linkType, Provider: *provider,
			Duration: *duration, Target: *target, Direction: *direction, Flows: *flows,
			SmallURL: *small, BulkAddr: *bulk,
		}
		fmt.Printf("audit %s — %s %s %ds sens %s flux %d (banc %s) → %s (bulk %s)\n",
			p.AuditID, p.LinkType, p.Site, p.Duration, p.Direction, p.Flows, sinks.Mode, p.Target, p.BulkAddr)
		res, err := audit.Run(context.Background(), p, audit.Deps{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit:", err)
			os.Exit(1)
		}
		// Ordre des preuves : la série brute d'abord (telemetry.csv), puis
		// la ligne gelée (link_audit.csv), puis le rapport qui scelle les
		// deux — chaque maillon existe avant d'être scellé.
		telPath, err := audit.WriteTelemetry("data", res)
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit telemetry:", err)
			os.Exit(1)
		}
		if err := audit.AppendLinkAudit("data", res); err != nil {
			fmt.Fprintln(os.Stderr, "audit write:", err)
			os.Exit(1)
		}
		fmt.Printf("audit ok — p50 %.1f p95 %.1f small %.1f débit %.1f Mbit/s note %q → data/link_audit.csv\n",
			res.RTTIdleP50, res.RTTIdleP95, res.HTTPSmallP95, res.ThroughputMbps, res.BloatGrade)
		if res.BloatGrade == "" {
			fmt.Fprintln(os.Stderr, "attention: bufferbloat NON MESURÉ (charge bulk absente/échouée) — ne pas citer cette ligne comme audit terrain")
		}
		// The report is part of the audit, not an opt-in extra: it is the
		// artifact an operator pastes into a ticket and the evidence whose
		// hashes a reviewer can re-check. It is written even when no
		// remediation can be emitted; the "not measured" section says why.
		link := remediate.Link{Interface: *iface, RouterOSInterface: *rosIface}
		rep, err := audit.BuildReport(res, link, *qdisc, *asymmetric, *cpuBound)
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit report:", err)
			os.Exit(1)
		}
		// Double sceau : la série brute, puis la ligne CSV que l'audit vient
		// d'ajouter — le rapport prouve les deux maillons.
		if err := rep.SealTelemetry(telPath); err != nil {
			fmt.Fprintln(os.Stderr, "audit telemetry seal:", err)
			os.Exit(1)
		}
		if err := rep.SealSource(filepath.Join("data", "link_audit.csv")); err != nil {
			fmt.Fprintln(os.Stderr, "audit report seal:", err)
			os.Exit(1)
		}
		jsonPath, mdPath, err := audit.WriteArtifacts(rep, "data")
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit artifacts:", err)
			os.Exit(1)
		}
		fmt.Printf("rapport → %s · %s\n", jsonPath, mdPath)
		if *reportPath != "" {
			var out []byte
			if strings.HasSuffix(*reportPath, ".json") {
				out, err = rep.RenderJSON()
			} else {
				out = []byte(rep.RenderMarkdown())
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "audit report render:", err)
				os.Exit(1)
			}
			if err := os.WriteFile(*reportPath, out, 0644); err != nil {
				fmt.Fprintln(os.Stderr, "audit report write:", err)
				os.Exit(1)
			}
			fmt.Printf("rapport additionnel → %s\n", *reportPath)
		}
		printAuditSummary(rep)
		// Code de sortie 2 : audit PARTIEL — le rapport existe et cite ses
		// manques ; l'appelant (script, CI, soutenance) sait ne pas citer
		// les absences comme des mesures.
		partial := false
		for _, u := range rep.Unavailable {
			switch u.Metric {
			case "bufferbloat grade", "economics", "throughput":
				partial = true
			}
		}
		if partial {
			os.Exit(2)
		}
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		profiles := fs.String("profiles", "P2", "profils à tester (P1,P2,P3-importé)")
		qdiscs := fs.String("qdiscs", "", "files à tester (pfifo_fast,fq_codel,cake) — vide = toutes")
		ccs := fs.String("cc", "", "contrôles de congestion (cubic,bbr) — vide = tous")
		direction := fs.String("direction", "up", "sens de charge : up | down | both (RRUL séquentiel)")
		reps := fs.Int("reps", 3, "répétitions par cellule (1–5)")
		flows := fs.Int("flows", 4, "flux de charge concurrents (1–8 ; 1 = régime historique gelé)")
		deadline := fs.Int("deadline", 1000, "objectif small p95 en ms (200–5000)")
		target := fs.String("target", "1.1.1.1", "cible de mesure")
		dataDir := fs.String("data", "data/runs", "répertoire des runs gelés")
		runID := fs.String("run-id", "", "reprendre ce run (cellules déjà gelées sautées)")
		bench := fs.String("bench", "", "banc : legacy (reproduction corpus) | dualns (production, défaut CGO_BENCH)")
		channelModel := fs.String("channel-model", "clean", "canal : clean | bursty (P2/P3 seulement — refus sinon)")
		_ = fs.Parse(os.Args[2:])
		if code := runCLI(*profiles, *qdiscs, *ccs, *reps, *deadline, *flows, *target, *direction, *dataDir, *runID, *bench, *channelModel); code != 0 {
			os.Exit(code)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func envDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// printAuditSummary — table 80 colonnes, lisible en 24 lignes : note, Δp95,
// QDI, drain, VoIP R, JFI, SLA CAP, sens, Ar/h primaire + B2B, sceau,
// scripts. Aucune valeur fabriquée : chaque absence est dite.
func printAuditSummary(rep *audit.Report) {
	line := strings.Repeat("─", 78)
	fmt.Println(line)
	grade := rep.Grade
	if grade == "" {
		grade = "—"
	}
	fmt.Printf(" AUDIT %-28s %-8s flux:%d sens:%s\n", rep.AuditID, grade, rep.Measured.FlowCount, rep.Measured.Direction)
	fmt.Println(line)
	fmt.Printf(" Δp95 %6.1f ms   QDI %6.1f ms   drain %6.1f ms\n", rep.Measured.BloatDeltaMs, rep.Measured.QDIMs, rep.Measured.DrainMs)
	fmt.Printf(" VoIP R %4.1f        JFI %5.1f %%      débit %6.2f Mbit/s\n", rep.Measured.VoIPR, rep.Measured.JFIPct, rep.Measured.ThroughputMbps)
	if rep.CAPSLA.Measured {
		etat := "conforme"
		if rep.CAPSLA.Breach {
			etat = "MANQUEMENT SLA"
		}
		fmt.Printf(" SLA CAP (220 ms) %5.1f %%  %s\n", rep.CAPSLA.OKPct, etat)
	} else {
		fmt.Printf(" SLA CAP (220 ms) non mesuré\n")
	}
	if rep.Economics != nil && rep.Economics.Primary != nil {
		fmt.Printf(" Coût primaire (%s) %10.0f Ar/h\n", rep.Economics.Primary.Name, rep.Economics.Primary.ARPerHour)
		for _, s := range rep.Economics.Scenarios {
			if s.Model == "B" {
				fmt.Printf("   B2B %-22s %10.0f Ar/h\n", s.Name, s.ARPerHour)
			}
		}
	} else {
		fmt.Printf(" Coût primaire non chiffré (ratio de réémission absent)\n")
	}
	if rep.Provenance.TelemetrySHA256 != "" {
		fmt.Printf(" sceau télémétrie %s…\n", rep.Provenance.TelemetrySHA256[:16])
	}
	if rep.Provenance.SourceSHA256 != "" {
		fmt.Printf(" sceau lignes    %s…\n", rep.Provenance.SourceSHA256[:16])
	}
	if rep.Remediation != nil {
		fmt.Printf(" scripts : tc + RouterOS v7+ → %s\n", rep.Provenance.SourceFile)
	} else {
		fmt.Println(" remédiation : non émise — capacité, RTT ou interface non mesurés")
	}
	fmt.Println(line)
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: cgo --serve [--addr host:port] [--tls=false] [--http-addr 127.0.0.1:9080] [--mode observe|full|auto]
       cgo audit --link-type 5g --site "Dept X" --duration 300
       cgo run --matrix full|reduced --profiles P1,P2 --reps 3
       cgo verify
       cgo figures
       cgo doctor                # rapport de capacités (tc, BBR, CAP_NET_ADMIN)
       cgo shape --restore       # retire les qdiscs root posés par meteolink
       cgo service install       # unité systemd (Linux, root)
       cgo version
       cgo testbedsrv --http 10.200.0.1:8081 --bulk 10.200.0.1:5201
`)
}

func verifyManifests(dataDir string) (bool, error) {
	// Un manifeste par run mesuré : l'ancien glob sur les manifestes laissait
	// passer les runs sans manifeste (25/160 dans le corpus gelé).
	runs, _ := filepath.Glob(filepath.Join(dataDir, "*", "aqm_eval.csv"))
	if len(runs) == 0 {
		return true, nil // no runs yet -> ok
	}
	for _, csvPath := range runs {
		dir := filepath.Dir(csvPath)
		name := filepath.Base(dir)
		raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if err != nil {
			return false, fmt.Errorf("%s: manifeste manquant: %w", name, err)
		}
		var doc struct {
			ConfigSHA256 string                          `json:"config_sha256"`
			Config       string                          `json:"config"`
			Files        []struct{ File, Sha256 string } `json:"files"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return false, fmt.Errorf("%s: manifeste illisible: %w", name, err)
		}
		if doc.Config == "" || !json.Valid([]byte(doc.Config)) {
			return false, fmt.Errorf("%s: configuration absente ou invalide", name)
		}
		sum := sha256.Sum256([]byte(doc.Config))
		if hex.EncodeToString(sum[:]) != doc.ConfigSHA256 {
			return false, fmt.Errorf("%s: config_sha256 ne correspond pas à la configuration", name)
		}
		listed := map[string]bool{}
		for _, f := range doc.Files {
			b, err := os.ReadFile(filepath.Join(dir, f.File))
			if err != nil {
				return false, fmt.Errorf("%s: %w", name, err)
			}
			s := sha256.Sum256(b)
			if hex.EncodeToString(s[:]) != f.Sha256 {
				return false, fmt.Errorf("%s: sha mismatch %s", name, f.File)
			}
			listed[f.File] = true
		}
		// Aucun fichier du run ne doit échapper au manifeste.
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false, fmt.Errorf("%s: %w", name, err)
		}
		for _, e := range entries {
			if e.IsDir() || e.Name() == "manifest.json" {
				continue
			}
			if !listed[e.Name()] {
				return false, fmt.Errorf("%s: fichier non couvert par le manifeste: %s", name, e.Name())
			}
		}
	}
	return true, nil
}
