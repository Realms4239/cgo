package campagne

import (
	"os"
	"testing"
)

// TestMain — les tests de matrice comptent des lignes ; un hôte BBRv3 ne
// doit pas changer le défaut. Les tests qui veulent bbr3 le stubbent.
func TestMain(m *testing.M) {
	readCCList = func() []string { return []string{"cubic", "bbr"} }
	os.Exit(m.Run())
}
