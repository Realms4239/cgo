package campagne

import (
	"os"

	"github.com/Realms4239/cgo/pkg/qdisc"
)

// ProdDeps returns production dependencies for the campagne runner,
// wired to the testbed topology from deploy/testbed.sh:
// netem hop veth-c (main ns) · shaping+aqm hop veth-s (netns cgo-srv).
// Every address is env-overridable for exotic setups.
func ProdDeps() Deps {
	target := env("CGO_TARGET", "10.200.0.1")
	small := env("CGO_SMALL_URL", "http://10.200.0.1:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.200.0.1:5201")
	ns := env("CGO_NS", "cgo-srv")
	return Deps{
		TC:       qdisc.ExecRunner{},            // netem on veth-c (main ns)
		TCShaper: qdisc.NsRunner{Ns: ns},        // tbf/cake on veth-s (inside ns)
		CliIf:    env("CGO_CLI_IF", "veth-c"),
		ShaperIf: env("CGO_SHAPER_IF", "veth-s"),
		Target:   target,
		SmallURL: small,
		BulkAddr: bulk,
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
