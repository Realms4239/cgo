package metrics

import "testing"

func TestMedianIQR(t *testing.T) {
	if got := Median([]float64{3, 1, 2}); got != 2 {
		t.Fatalf("median=%v want 2", got)
	}
	if got := IQR([]float64{1, 2, 3, 4}); got != 1.5 {
		t.Fatalf("iqr=%v want 1.5", got)
	}
}

func TestCliffDeltaObvious(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{4, 5, 6}
	if d := CliffDelta(a, b); d != 1 {
		t.Fatalf("delta=%v want 1", d)
	}
	if s := CliffInterp(1); s != "grand" {
		t.Fatalf("interp=%q want grand", s)
	}
}

func TestMannWhitneySeparated(t *testing.T) {
	a := []float64{60.2, 55.6, 58.1, 57.0, 59.3}
	b := []float64{96.2, 98.1, 97.5, 98.0, 96.8}
	_, p := MannWhitneyTwoSided(a, b)
	if p >= 0.05 {
		t.Fatalf("p=%v want <0.05", p)
	}
}

func TestBootstrapCISeeded(t *testing.T) {
	xs := []float64{60.2, 55.6, 58.1, 57.0, 59.3, 61.0}
	lo1, hi1 := BootstrapMedianCI95(xs, 10000, 42)
	lo2, hi2 := BootstrapMedianCI95(xs, 10000, 42)
	if lo1 != lo2 || hi1 != hi2 || !(lo1 <= 58.1 && 58.1 <= hi1) {
		t.Fatalf("ci not seeded/stable: [%v,%v] vs [%v,%v]", lo1, hi1, lo2, hi2)
	}
}
