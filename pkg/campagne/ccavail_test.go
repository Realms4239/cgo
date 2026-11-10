package campagne

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/model"
)

func stubCCs(t *testing.T, list ...string) {
	t.Helper()
	old := readCCList
	readCCList = func() []string { return list }
	t.Cleanup(func() { readCCList = old })
}

func TestDefaultCCsWithoutBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	got := DefaultCCs()
	if len(got) != 2 || got[0] != "cubic" || got[1] != "bbr" {
		t.Fatalf("défaut sans bbr3 : %v", got)
	}
}

func TestDefaultCCsWithBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
	got := DefaultCCs()
	if len(got) != 3 || got[2] != "bbr3" {
		t.Fatalf("défaut avec bbr3 : %v", got)
	}
}

func TestExplicitBBRv3RefusedWithoutKernel(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	_, err := StartMatrixFiltered(context.Background(), []string{"P2"},
		[]string{"cake"}, []string{"bbr3"}, 1, fastDeps(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "install-bbr3") {
		t.Fatalf("bbr3 explicite sans noyau doit refuser : %v", err)
	}
}

func TestExplicitBBRv3AcceptedWhenPresent(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
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
	old := readKernelRelease
	readKernelRelease = func() string { return "7.1.0-rc5-cgo-bbr3-test" }
	t.Cleanup(func() { readKernelRelease = old })
	cfg, err := buildMatrixConfig([]string{"P1"}, []string{"cake"}, []string{string(model.BBRv3)}, 1, Deps{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), `"kernel":"7.1.0-rc5-cgo-bbr3-test"`) {
		t.Fatalf("kernel absent du gel : %s", cfg)
	}
}
