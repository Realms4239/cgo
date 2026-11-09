// Commande cgo-audit — audit de liaison phasé (60 s par défaut).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Realms4239/cgo/audit"
)

func main() {
	fs := flag.NewFlagSet("cgo-audit", flag.ExitOnError)
	// Surface opérateur minimale : site, type de lien, durée — la détection
	// du banc choisit cible et puits, comme le CLI historique.
	site := fs.String("site", "Site X", "site / département")
	linkType := fs.String("link-type", "5g", "link_type: fiber, 5g, 4g, vsat, other")
	duration := fs.Int("duration", 60, "durée en secondes (60 = 10/40/10)")
	_ = fs.Parse(os.Args[1:])

	sinks := audit.ResolveSinks()
	p := audit.Params{
		AuditID:  fmt.Sprintf("audit-%d", time.Now().Unix()),
		Site:     *site,
		LinkType: *linkType,
		Duration: *duration,
		Target:   sinks.Target,
		SmallURL: sinks.SmallURL,
		BulkAddr: audit.EnvOrBulk(sinks),
		// Sens et flux par défaut du banc durci (4 flux) ; le mode de banc
		// est un fait de détection, consigné avec l'audit.
		Direction: "up",
		Flows:     4,
		SinkMode:  sinks.Mode,
		Origin:    "cli",
	}
	res, err := audit.Run(context.Background(), p, audit.Deps{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit:", err)
		os.Exit(1)
	}
	fmt.Printf("audit ok — p50 %.1f p95 %.1f débit %.1f Mbit/s note %q\n",
		res.RTTIdleP50, res.RTTIdleP95, res.ThroughputMbps, res.BloatGrade)
}
