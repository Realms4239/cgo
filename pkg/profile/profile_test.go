package profile

import (
	"os"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// P3 VSAT built-in values (5/600/30/1) per LIEN Tableau 2 — import layer must not shadow them.
func TestP3Values(t *testing.T) {
	p, ok := model.Profiles["P3"]
	if !ok {
		t.Fatal("P3 missing")
	}
	if p.CapacityMbps != 5 || p.DelayMs != 600 || p.JitterMs != 30 || p.LossPct != 1 {
		t.Fatalf("P3 wrong values: %+v", p)
	}
}

// Import registers AND persists a custom profile (single source — callers get both).
func TestImportCustom(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	p := model.Profile{ID: "PX", CapacityMbps: 20, DelayMs: 100, JitterMs: 15, LossPct: 0.5}
	if err := Import(p); err != nil {
		t.Fatal(err)
	}
	if _, ok := model.Profiles["PX"]; !ok {
		t.Fatal("PX not registered in model.Profiles by Import")
	}
	model.Profiles["PX"] = model.Profile{} // simulate fresh process
	Load()
	if got := model.Profiles["PX"]; got.CapacityMbps != 20 || got.ID != "PX" {
		t.Fatalf("persisted values lost after Load: %+v", got)
	}
}