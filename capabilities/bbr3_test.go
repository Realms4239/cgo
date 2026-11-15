package capabilities

import (
	"errors"
	"testing"
)

// Le pin épinglé est testé par rapport à la constante, jamais recopie à la
// main : un pin qui bouge dans bbr3.go fait échouer ces tests.
const pinnedMarker = `{"branch":"v3","tag":"bbrv3-2025-03-18","commit":"90210de4b779d40496dee0b89081780eeddf2a60","kernel":"6.13.7","built":"2026-11-12T00:00:00+00:00"}`

func TestProvisionedBBRv3TrueOnPinnedHost(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	defer StubProvisioned(pinnedMarker, "Linux version 6.13.7-g90210de4b779 #1 SMP")()
	if !ProvisionedBBRv3() {
		t.Fatal("hôte épinglé non reconnu : marqueur + version g90210de4b779 attendus vrais")
	}
	if got := BBRv3Token(); got != "bbr" {
		t.Fatalf("token = %q, want « bbr » (la v3 est enregistrée sous bbr sur cet arbre)", got)
	}
}

func TestBBRv3TokenBareTokenWins(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3") // marqueur/version restent « stock » (stubCCs)
	if ProvisionedBBRv3() {
		t.Fatal("hôte stock lu comme provisionné")
	}
	if got := BBRv3Token(); got != "bbr3" {
		t.Fatalf("token = %q, want « bbr3 » (token nu, sans provenance)", got)
	}
}

func TestBBRv3TokenStockWithoutProvenance(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	old := readMarker
	readMarker = func() (string, error) { return "", errors.New("marqueur absent") }
	t.Cleanup(func() { readMarker = old })
	if ProvisionedBBRv3() {
		t.Fatal("hôte stock sans provenance lu comme provisionné")
	}
	if got := BBRv3Token(); got != "" {
		t.Fatalf("token = %q, want « » (refus plutôt que repli)", got)
	}
}

func TestProvisionedBBRv3WrongCommitMarker(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	wrong := `{"branch":"v3","tag":"bbrv3-2025-03-18","commit":"1111111111111111111111111111111111111111","kernel":"6.13.7","built":"2026-11-12T00:00:00+00:00"}`
	defer StubProvisioned(wrong, "Linux version 6.13.7-g111111111111 #1 SMP")()
	if ProvisionedBBRv3() {
		t.Fatal("commit hors pin accepté comme provenance")
	}
	if got := BBRv3Token(); got != "" {
		t.Fatalf("token = %q, want « »", got)
	}
}

func TestProvisionedBBRv3VersionMismatch(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	// Marqueur épinglé correct, release sans le suffixe -g90210de4b779.
	defer StubProvisioned(pinnedMarker, "Linux version 6.13.7-generic #1 SMP")()
	if ProvisionedBBRv3() {
		t.Fatal("version sans suffixe g acceptée comme provenance")
	}
	if got := BBRv3Token(); got != "" {
		t.Fatalf("token = %q, want « »", got)
	}
}

func TestProvisionedBBRv3MarkerMissing(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	oldM := readMarker
	readMarker = func() (string, error) { return "", errors.New("open /var/lib/cgo/bbr3-kernel.json: no such file or directory") }
	t.Cleanup(func() { readMarker = oldM })
	if ProvisionedBBRv3() {
		t.Fatal("marqueur absent accepté comme provenance")
	}
	if got := BBRv3Token(); got != "" {
		t.Fatalf("token = %q, want « »", got)
	}
}
