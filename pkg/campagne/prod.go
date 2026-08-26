package campagne

import (
	"os"

	"github.com/Realms4239/cgo/pkg/qdisc"
)

// ProdDeps returns production dependencies for the campagne runner,
// wired to the testbed topology from deploy/testbed.sh:
// netem hop veth-c (main ns) · shaping+aqm hop veth-s (netns cgo-srv).
// Every address is env-overridable for exotic setups.
//
// Shaper placement: bulk is an UPLOAD (BulkSendTo client→sink), so the
// rate cap must sit on the client egress — veth-c, stacked as child of
// the netem root at 1: (ApplyShaper parent 1: fallback). Shaping on
// veth-s egress only covers the download direction and leaves the
// upload unshaped (G4 fail → invalid rows).
func ProdDeps() Deps {
	target := env("CGO_TARGET", "10.200.0.1")
	small := env("CGO_SMALL_URL", "http://10.200.0.1:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.200.0.1:5201")
	ns := env("CGO_NS", "cgo-srv")
	cliIf := env("CGO_CLI_IF", "veth-c")
	shaperIf := env("CGO_SHAPER_IF", cliIf) // same hop as netem — upload egress
	return Deps{
		TC:       qdisc.ExecRunner{},     // netem on veth-c (main ns)
		TCShaper: qdisc.ExecRunner{},     // shaper stacked on veth-c too (upload egress)
		CliIf:    cliIf,
		ShaperIf: shaperIf,
		Target:   target,
		SmallURL: small,
		BulkAddr: bulk,
		StatsFn: func() []qdisc.Stats {
			// poll both hops for drops/bytes delta
			sts1, err1 := qdisc.PollStats(qdisc.ExecRunner{}, cliIf)
			sts2, err2 := qdisc.PollStats(qdisc.NsRunner{Ns: ns}, "veth-s")
			var all []qdisc.Stats
			if err1 == nil {
				all = append(all, sts1...)
			}
			if err2 == nil {
				all = append(all, sts2...)
			}
			return all
		},
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
