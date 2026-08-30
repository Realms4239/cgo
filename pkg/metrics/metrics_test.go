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

// Référence: un plan complet de gaspillage coûte exactement 30 000 Ar/h.
func TestCostARPerHGolden(t *testing.T) {
	if got := CostARPerH(uint64(4.5 * 1024 * 1024 * 1024)); got != 30000 {
		t.Fatalf("cost = %v, want 30000", got)
	}
	if got := CostARPerH(0); got != 0 {
		t.Fatalf("zero waste must be zero")
	}
}
