package campagne

import (
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// ProdDeps returns production dependencies for the campagne runner.
// Interfaces are left blank where the testbed topology will be provisioned
// by the deploy bootstrap (veth pair, addrs). Until M1.5 these run with
// FakeRunner semantics on the host; on the VM they use ExecRunner.
func ProdDeps() Deps {
	return Deps{
		TC:       qdisc.ExecRunner{},
		CliIf:    "veth-c",
		ShaperIf: "veth-s",
		Target:   "127.0.0.1",
		SmallURL: "http://127.0.0.1:8081/small",
		BulkAddr: "127.0.0.1:5201",
	}
}
