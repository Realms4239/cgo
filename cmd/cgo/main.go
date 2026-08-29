// CGO — LIEN instrument. Commands per LIEN.md Partie III:
//
//	cgo --serve                          dashboard + API (default port :9090)
//	cgo audit --link-type T --site S --duration N
//	cgo run  --matrix full|reduced --profiles P1,P2 --reps 3
//	cgo verify                           manifest integrity check
//	cgo figures                          regenerate SVGs from frozen CSVs
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
	"syscall"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/doctor"
	"github.com/Realms4239/cgo/pkg/figures"
)

// version — overridden at release time: -X main.version={{.Version}}.
var version = "1.0.6"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "--serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		def := os.Getenv("CGO_DASHBOARD__ADDR")
		if def == "" {
			def = ":9090"
		}
		addr := fs.String("addr", def, "listen address")
		mode := fs.String("mode", "auto", "observe | full | auto (défaut: OS décide — Windows observe, Linux complet)")
		_ = fs.Parse(os.Args[2:])
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := runServer(ctx, *addr, doctor.Mode(*mode)); err != nil {
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
ExecStart=/usr/local/bin/cgo --serve --addr 0.0.0.0:9090
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
		linkType := fs.String("link-type", "5g", "link_type: fiber, 5g, 4g, vsat, other")
		site := fs.String("site", "Site X", "site / département")
		provider := fs.String("provider", "", "provider")
		duration := fs.Int("duration", 60, "duration seconds")
		target := fs.String("target", "8.8.8.8", "ping target")
		_ = fs.Parse(os.Args[2:])
		p := audit.Params{
			AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()),
			Site: *site, LinkType: *linkType, Provider: *provider,
			Duration: *duration, Target: *target,
		}
		fmt.Printf("audit %s — %s %s %ds → %s\n", p.AuditID, p.LinkType, p.Site, p.Duration, p.Target)
		res, err := audit.Run(context.Background(), p, audit.Deps{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit:", err)
			os.Exit(1)
		}
		if err := audit.AppendLinkAudit("data", res); err != nil {
			fmt.Fprintln(os.Stderr, "audit write:", err)
			os.Exit(1)
		}
		fmt.Printf("audit ok — p50 %.1f p95 %.1f small %.1f → data/link_audit.csv\n", res.RTTIdleP50, res.RTTIdleP95, res.HTTPSmallP95)
	case "run":
		fmt.Fprintln(os.Stderr, "run: use API POST /api/run/start {profiles,reps} (CLI run direct en M3+)")
		os.Exit(2)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: cgo --serve [--addr host:port] [--mode observe|full|auto]
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
	runs, _ := filepath.Glob(filepath.Join(dataDir, "*", "manifest.json"))
	if len(runs) == 0 {
		return true, nil // no runs yet -> ok
	}
	for _, mf := range runs {
		raw, err := os.ReadFile(mf)
		if err != nil {
			return false, err
		}
		var doc struct {
			Files []struct{ File, Sha256 string } `json:"files"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return false, err
		}
		dir := filepath.Dir(mf)
		for _, f := range doc.Files {
			b, err := os.ReadFile(filepath.Join(dir, f.File))
			if err != nil {
				return false, err
			}
			sum := sha256.Sum256(b)
			if hex.EncodeToString(sum[:]) != f.Sha256 {
				return false, fmt.Errorf("sha mismatch %s", f.File)
			}
		}
	}
	return true, nil
}
