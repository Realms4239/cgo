package metrics

import "testing"

func TestPercentile(t *testing.T) {
	s := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	if got := Percentile(s, 95); got != 95.5 {
		t.Fatalf("p95 = %v, want 95.5", got)
	}
	if got := Percentile([]float64{42}, 99); got != 42 {
		t.Fatalf("single = %v", got)
	}
}

func TestDeadlineOKPct(t *testing.T) {
	got := DeadlineOKPct([]float64{10, 20, 30, 200}, 100)
	if got != 75 {
		t.Fatalf("deadline pct = %v, want 75", got)
	}
}

// Référence : 4,5 Go gaspillés au palier Yas Net Month (25 000 Ar / 4,5 Go =
// 25 002 Ar pour le bundle complet), fenêtre 3 min extrapolée à l'heure
// (×20) = 500 040 Ar/h. L'ancien golden (30 000 Ar) mélangeait bundle et
// heure sur un tarif inexistant — corrigé via docs/data-prices.md (yas.mg).
func TestCostARPerHGolden(t *testing.T) {
	got := CostARPerH(uint64(4.5 * 1024 * 1024 * 1024))
	if got < 500039 || got > 500041 {
		t.Fatalf("cost = %v, want ~500040 (25 002 Ar bundle × 20 fenêtre/heure)", got)
	}
	if got := CostARPerH(0); got != 0 {
		t.Fatalf("zero waste must be zero")
	}
}

// Paliers réels — le catalogue doit refléter les tarifs publiés.
func TestTiers(t *testing.T) {
	byName := map[string]PriceTier{}
	for _, tr := range Tiers {
		byName[tr.Name] = tr
	}
	if tr, ok := byName["yas-day-1gb"]; !ok || tr.ARGB != 1000 {
		t.Fatalf("yas-day-1gb = %+v, want 1000 Ar/Go (Ye'low One)", tr)
	}
	if tr, ok := byName["yas-month-4.5gb"]; !ok || tr.ARGB != 5556 {
		t.Fatalf("yas-month-4.5gb = %+v, want 5556 Ar/Go (25 000 Ar / 4,5 Go)", tr)
	}
	if tr, ok := byName["yas-ftth-100gb"]; !ok || tr.ARGB != 490 {
		t.Fatalf("yas-ftth-100gb = %+v, want 490 Ar/Go (49 000 Ar / 100 Go)", tr)
	}
	// le quotidien Ye'low One est le MEILLEUR rapport du catalogue Yas (1 000 Ar/Go)
	// vs le mensuel 4,5 Go qui est le pire (5 556 Ar/Go) — écart 5,6×
	small := byName["yas-day-1gb"].ARGB
	worst := byName["yas-month-4.5gb"].ARGB
	if worst/small < 5 {
		t.Fatalf("le mensuel 4,5 Go (%v Ar/Go) doit être >5× le Ye'low One (%v)", worst, small)
	}
}

// Palier explicite — chaque calcul doit suivre le forfait choisi.
func TestCostARPerHTier(t *testing.T) {
	w := uint64(100 * MB)
	day := CostARPerHTier(w, PriceTier{Name: "d", ARGB: 1000})
	ftth := CostARPerHTier(w, PriceTier{Name: "f", ARGB: 490})
	if day <= ftth {
		t.Fatalf("mobile (%v) doit coûter plus cher que fibre (%v) à volume égal", day, ftth)
	}
}

// VoIP R (E-model simplifié) — repères calculés depuis G.107 : fibre calme
// ~93, VSAT propre ~85 (le E-model tolère le délai < 400 ms), catastrophe
// (délai max + grosse gigue + perte) chute franchement. R>50 = supportable.
func TestVoIPR(t *testing.T) {
	fiber := VoIPR(10, 2, 0)
	if fiber < 90 || fiber > 93.2 {
		t.Fatalf("fibre calme R = %v, want 90..93.2", fiber)
	}
	vsat := VoIPR(300, 30, 1)
	if vsat < 68 || vsat > 75 {
		t.Fatalf("VSAT R = %v, want 68..75 (sévérité ITU au-delà de 177 ms)", vsat)
	}
	if vsat >= fiber {
		t.Fatalf("VSAT (%v) ne doit pas battre la fibre (%v)", vsat, fiber)
	}
	// bufferbloat : 400 ms aller + 100 ms de gigue + 5 % de perte
	bloat := VoIPR(400, 100, 5)
	if bloat > vsat-10 {
		t.Fatalf("bufferbloat R=%v doit chute ≥10 pts sous le VSAT R=%v", bloat, vsat)
	}
	if bloat < 0 {
		t.Fatalf("R borné à 0, obtenu %v", bloat)
	}
}
