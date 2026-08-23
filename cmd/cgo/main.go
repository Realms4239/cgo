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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	case "audit", "run", "verify", "figures":
		fmt.Fprintf(os.Stderr, "%s: not implemented until M1/M2/M3\n", os.Args[1])
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
