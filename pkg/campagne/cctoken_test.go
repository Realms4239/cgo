package campagne

import (
	"context"
	"testing"

	"github.com/Realms4239/cgo/capabilities"
)

// pinnedMarkerJSON — le marqueur que kit/install-bbr3.sh dépose après une
// installation épinglée (commit épinglé + version -g90210de4b779).
const pinnedMarkerJSON = `{"branch":"v3","tag":"bbrv3-2025-03-18","commit":"90210de4b779d40496dee0b89081780eeddf2a60","kernel":"6.13.7","built":"2026-11-12T00:00:00+00:00"}`

// TestBBRv3ProvisionedGateAccepts — la trappe défaite par ce task : sur un
// hôte provisionné, la v3 est enregistrée sous « bbr », jamais sous « bbr3 ».
// Une demande explicite bbr3 doit passer la porte du moteur (le noyau
// n'expose pas le token bbr3, mais le pin est là) — là où elle était
// refusée avant le gate de provenance. Porte uniquement : l'exécution
// complète est couverte par TestExplicitBBRv3AcceptedOnProvisionedHost.
func TestBBRv3ProvisionedGateAccepts(t *testing.T) {
	defer capabilities.StubHost([]string{"cubic", "bbr"}, "")()
	defer capabilities.StubProvisioned(pinnedMarkerJSON, "Linux version 6.13.7-g90210de4b779 #1 SMP")()
	m, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err != nil {
		t.Fatalf("porte refusant bbr3 sur hôte provisionné : %v", err)
	}
	defer stopWait(t, m)
}
