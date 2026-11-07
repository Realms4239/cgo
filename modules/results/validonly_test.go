package results

import "testing"

// TestValidOnlyFields — degraded exclus des champs valid-only (décision spec),
// filtres historiques (Count/Quarantined/médianes mixtes) inchangés.
func TestValidOnlyFields(t *testing.T) {
	dir := t.TempDir()
	writeRun(t, dir, "run-v", header18,
		"run-v,1,P2,cake,bbr,1,100.0,114.0,14.0,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
		"run-v,2,P2,cake,bbr,1,101.0,115.0,15.0,211.0,97.0,18.3,90,0,120000,12.50,0.0,valid",
		"run-v,3,P2,cake,bbr,1,102.0,116.0,16.0,500.0,10.0,18.0,95,0,130000,13.00,0.0,degraded",
		"run-v,4,P2,cake,bbr,1,103.0,117.0,17.0,600.0,5.0,2.0,99,0,140000,14.00,0.0,invalid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	g := findGroup(groups, "P2", "cake", "bbr")
	if g == nil {
		t.Fatal("missing P2|cake|bbr")
	}
	if g.Smallp95ValidN != 2 {
		t.Fatalf("validN=%d want 2 (degraded+invalid exclus)", g.Smallp95ValidN)
	}
	if g.Smallp95ValidMedian != 210.1 { // médiane linéaire de {209.2,211.0}
		t.Fatalf("validMedian=%v want 210.1", g.Smallp95ValidMedian)
	}
	if g.DeadlineValidN != 2 || g.DeadlineValidMedian != 97.55 {
		t.Fatalf("deadline valid n=%d med=%v want 2/97.55", g.DeadlineValidN, g.DeadlineValidMedian)
	}
	lo, hi := g.Smallp95ValidCI95[0], g.Smallp95ValidCI95[1]
	if !(lo <= 210.1 && 210.1 <= hi) {
		t.Fatalf("CI95=[%v,%v] ne couvre pas la médiane", lo, hi)
	}
	// historique intact : Count compte tout, Quarantined = invalid seul.
	if g.Count != 4 || g.Quarantined != 1 {
		t.Fatalf("count=%d quar=%d want 4/1", g.Count, g.Quarantined)
	}
}
