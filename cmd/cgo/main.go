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
	"syscall"

	"github.com/Realms4239/cgo/pkg/figures"
)

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
		_ = fs.Parse(os.Args[2:])
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := runServer(ctx, *addr); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
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
	case "audit", "run":
		fmt.Fprintf(os.Stderr, "%s: not implemented until M3 (audit) / use API POST /api/run/start\n", os.Args[1])
		os.Exit(2)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: cgo --serve [--addr host:port]
       cgo audit --link-type 5g --site "Dept X" --duration 300
       cgo run --matrix full|reduced --profiles P1,P2 --reps 3
       cgo verify
       cgo figures
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
