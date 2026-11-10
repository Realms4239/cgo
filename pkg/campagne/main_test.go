package campagne

import (
	"os"
	"testing"

	"github.com/Realms4239/cgo/capabilities"
)

// TestMain — les tests de matrice comptent des lignes ; un hôte BBRv3 ne
// doit pas changer le défaut. Les tests qui veulent bbr3 le stubbent.
func TestMain(m *testing.M) {
	restore := capabilities.StubHost([]string{"cubic", "bbr"}, "")
	code := m.Run()
	restore()
	os.Exit(code)
}
