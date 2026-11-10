package campagne

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/capabilities"
	"github.com/Realms4239/cgo/model"
)

// Ces tests portent sur la MATRICE (le bbr3 explicite est joué uniquement si
// le noyau l'enregistre) — ils vivent ici, pas dans le module capabilities.
// Le noyau simulé traverse la seam de test du module.

func TestExplicitBBRv3RefusedWithoutKernel(t *testing.T) {
	defer capabilities.StubHost([]string{"cubic", "bbr"}, "")()
	_, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "install-bbr3") {
		t.Fatalf("bbr3 explicite sans noyau doit refuser : %v", err)
	}
}

func TestExplicitBBRv3AcceptedWhenPresent(t *testing.T) {
	defer capabilities.StubHost([]string{"cubic", "bbr", "bbr3"}, "")()
	m, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer stopWait(t, m)
	for i := 0; i < 40 && m.Done != m.Total; i++ {
		time.Sleep(50 * time.Millisecond)
	}
	if m.Total != 1 {
		t.Fatalf("total = %d, want 1", m.Total)
	}
}

func TestKernelReleaseIsFrozen(t *testing.T) {
	defer capabilities.StubHost([]string{"cubic", "bbr"}, "7.1.0-rc5-cgo-bbr3-test")()
	cfg, err := buildMatrixConfig([]string{"P1"}, []string{"cake"}, []string{string(model.BBRv3)}, 1, Deps{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), `"kernel":"7.1.0-rc5-cgo-bbr3-test"`) {
		t.Fatalf("kernel absent du gel : %s", cfg)
	}
}
