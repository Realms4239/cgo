package audit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditRefusesEmptyProbes — garde anti-ligne-fantôme : cible injoignable
// (zéro sonde) = erreur, pas une ligne de zéros gelée.
func TestAuditRefusesEmptyProbes(t *testing.T) {
	deps := Deps{
		Ping:  func(context.Context, string, int) []float64 { return nil },
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	_, err := Run(context.Background(), Params{AuditID: "a-empty", Site: "s", Duration: 2, Target: "10.0.0.0"}, deps)
	if err == nil {
		t.Fatal("audit aveugle accepté : ligne de zéros gelable")
	}
}

// TestAuditMatchProfile — rapprochement au référentiel : 22 ms idle → P1,
// 628 ms → P3, avec delta lisible. Sans débit : rapprochement au délai seul,
// dit explicitement.
func TestAuditMatchProfile(t *testing.T) {
	deps := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{620, 628, 635} },
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), Params{AuditID: "a-vsat", Site: "s", Duration: 2, Target: "t"}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.ProfileMatch != "P3" {
		t.Fatalf("match=%q want P3 (628 ms idle)", res.ProfileMatch)
	}
	if res.MatchDelta == "" {
		t.Fatal("delta vide")
	}
}

// TestAuditMatchFrozen — le rapprochement voyage dans le CSV (même API que
// la liste : extract-stats et la vue Audits lisent par nom).
func TestAuditMatchFrozen(t *testing.T) {
	dir := t.TempDir()
	deps := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{20, 21, 22} },
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), Params{AuditID: "a-p1", Site: "s", Duration: 2, Target: "t"}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.ProfileMatch != "P1" {
		t.Fatalf("match=%q want P1", res.ProfileMatch)
	}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "link_audit.csv"))
	found := false
	for _, ln := range strings.Split(string(raw), "\n") {
		if len(ln) > 0 && contains(ln, "a-p1") && contains(ln, "P1") {
			found = true
		}
	}
	if !found {
		t.Fatal("ligne gelée sans rapprochement P1")
	}
}

// TestAuditPresets — le protocole RQ1 est matérialisé : 3 audits opérateur,
// durées et sites figés par le plan (pas de recopie manuelle).
func TestAuditPresets(t *testing.T) {
	ps := Presets()
	if len(ps) != 3 {
		t.Fatalf("%d presets want 3 (fibre + 4G creuse/pointe)", len(ps))
	}
	want := map[string]bool{"fibre-siege": false, "4g-creuse": false, "4g-pointe": false}
	for _, p := range ps {
		if _, ok := want[p.ID]; !ok {
			t.Fatalf("preset inattendu %q", p.ID)
		}
		want[p.ID] = true
		if p.Duration != 60 || p.Site == "" || p.Target == "" {
			t.Fatalf("preset %q incomplet", p.ID)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
