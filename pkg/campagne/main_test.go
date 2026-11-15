package campagne

import (
	"os"
	"testing"

	"github.com/Realms4239/cgo/capabilities"
)

// TestMain — les tests de matrice comptent des lignes ; un hôte BBRv3 ne
// doit pas changer le défaut. Les tests qui veulent bbr3 le stubbent.
// La provenance est aussi épinglée « stock » : sans ça, un hôte provisionné
// (marqueur + version -g90210de4b779) ferait entrer bbr3 dans le défaut via
// BBRv3Token(), même avec la liste CC réduite à cubic/bbr.
func TestMain(m *testing.M) {
	restore := capabilities.StubHost([]string{"cubic", "bbr"}, "")
	restoreProv := capabilities.StubProvisioned("{}", "Linux version 0.0.0-stock")
	code := m.Run()
	restoreProv()
	restore()
	os.Exit(code)
}
