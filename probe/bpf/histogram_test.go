package bpf

import "testing"

// Oracle positif — 1000 échantillons dans le bucket 10 donnent un p99.9
// contenu dans les bornes du bucket : [2^10, 2^11) = [1024, 2048) us.
func TestPercentileP999WithinBucketBounds(t *testing.T) {
	var h Log2Histogram
	h[10] = 1000
	got, ok := Percentile(h, 0.999, 1000)
	if !ok {
		t.Fatal("p99.9 refusé avec 1000 échantillons")
	}
	if got < 1024 || got >= 2048 {
		t.Fatalf("p99.9 = %v us, hors des bornes du bucket 10 [1024, 2048)", got)
	}
}

// Oracle hostile — sous le plancher de 1000, aucun percentile n'est affirmé :
// 999 échantillons restent une absence explicite, jamais une extrapolation.
func TestPercentileRefusesBelowSampleFloor(t *testing.T) {
	var h Log2Histogram
	h[10] = 999
	if got, ok := Percentile(h, 0.999, 999); ok || got != 0 {
		t.Fatalf("999 échantillons ont produit un percentile: %v, ok=%v", got, ok)
	}
}

// Oracle hostile — un quantile hors de (0, 1] est refusé, y compris 1.5.
func TestPercentileRefusesInvalidQuantile(t *testing.T) {
	var h Log2Histogram
	h[10] = 1000
	for _, q := range []float64{1.5, 0, -0.1} {
		if got, ok := Percentile(h, q, 1000); ok || got != 0 {
			t.Fatalf("q=%v accepté: %v, ok=%v", q, got, ok)
		}
	}
}

// Oracle hostile — un histogramme vide (ou incohérent avec le compteur)
// n'invente aucun percentile.
func TestPercentileEmptyOrShortfallReturnsAbsence(t *testing.T) {
	var empty Log2Histogram
	if got, ok := Percentile(empty, 0.999, 5000); ok || got != 0 {
		t.Fatalf("histogramme vide a produit: %v, ok=%v", got, ok)
	}
	var short Log2Histogram
	short[10] = 1000 // 1000 présents mais 5000 annoncés : cumul insuffisant
	if got, ok := Percentile(short, 0.999, 5000); ok || got != 0 {
		t.Fatalf("cumul insuffisant a produit: %v, ok=%v", got, ok)
	}
}

// Oracle de bord — les bins se cumulent : p50 et p95 tombent dans des buckets
// distincts et croissants.
func TestPercentileCumulativeBuckets(t *testing.T) {
	var h Log2Histogram
	h[4] = 500 // [16, 32)
	h[7] = 450 // [128, 256)
	h[10] = 50 // [1024, 2048)
	p50, ok := Percentile(h, 0.50, 1000)
	if !ok || p50 != 16 {
		t.Fatalf("p50 = %v, ok=%v ; voulu 16", p50, ok)
	}
	p95, ok := Percentile(h, 0.95, 1000)
	if !ok || p95 != 128 {
		t.Fatalf("p95 = %v, ok=%v ; voulu 128", p95, ok)
	}
}
