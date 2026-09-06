package profile

import (
	"os"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

func chdirTemp(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// TestReservedRefused — P1–P4 ni importables ni supprimables, et Load ne les
// laisse plus masquer par un vieux fichier (P3 35/60 fantôme vu en prod).
func TestReservedRefused(t *testing.T) {
	chdirTemp(t)
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		if !Reserved(id) {
			t.Fatalf("%s devrait être réservé", id)
		}
		if err := Import(model.Profile{ID: id, CapacityMbps: 99}); err == nil {
			t.Fatalf("import %s accepté : le référentiel serait masqué", id)
		}
		if err := Delete(id); err == nil {
			t.Fatalf("delete %s accepté", id)
		}
	}
	// P3 garde ses valeurs malgré un fichier qui le contrefait
	if err := os.MkdirAll("data", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("data/profiles.json",
		[]byte(`{"P3":{"id":"P3","capacity_mbps":35,"delay_ms":60}}`), 0644); err != nil {
		t.Fatal(err)
	}
	Load()
	if got := model.Profiles["P3"]; got.CapacityMbps != 5 || got.DelayMs != 600 {
		t.Fatalf("P3 masqué par le fichier : %+v", got)
	}
	if err := Delete("PX-absent"); err == nil {
		t.Fatal("delete d'un inconnu accepté")
	}
}

// TestDeleteRoundTrip — importé puis supprimé : ni carte, ni fichier.
func TestDeleteRoundTrip(t *testing.T) {
	chdirTemp(t)
	p := model.Profile{ID: "PX-DEL", CapacityMbps: 10, DelayMs: 50}
	if err := Import(p); err != nil {
		t.Fatal(err)
	}
	if err := Delete("PX-DEL"); err != nil {
		t.Fatal(err)
	}
	if _, ok := model.Profiles["PX-DEL"]; ok {
		t.Fatal("toujours en carte après delete")
	}
	Load()
	if _, ok := model.Profiles["PX-DEL"]; ok {
		t.Fatal("ressuscité par Load après delete")
	}
}
