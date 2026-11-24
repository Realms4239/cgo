package metrics

import (
	"math"
	"testing"
)

// TestErrorBudget_NoSampleIsNoGreen — oracle hostile : zéro observation ne doit
// jamais produire un vert, même avec une cible généreuse. L'absence est refusée
// explicitement, pas arrondie à un budget plein.
func TestErrorBudget_NoSampleIsNoGreen(t *testing.T) {
	burn, remaining, ok := ErrorBudget(0, 0, 0.99)
	if ok {
		t.Fatal("zéro observation acceptée : un budget sans échantillon ne peut pas être vert")
	}
	if remaining > 0 {
		t.Fatalf("budget restant %g sans observation : aucun vert ne doit être fabriqué", remaining)
	}
	if math.IsNaN(burn) || math.IsNaN(remaining) || math.IsInf(burn, 0) || math.IsInf(remaining, 0) {
		t.Fatalf("refus doit être fini, got burn=%g remaining=%g", burn, remaining)
	}
}

// TestErrorBudget_BurnRateFromRealMisses — 95/100 à 0.99 : cinq erreurs pour
// une seule autorisée. Le burn rate doit dépasser 1 et le budget être dépassé.
func TestErrorBudget_BurnRateFromRealMisses(t *testing.T) {
	burn, remaining, ok := ErrorBudget(95, 100, 0.99)
	if !ok {
		t.Fatal("observation réelle refusée à tort")
	}
	if burn <= 1 {
		t.Fatalf("burn rate %g, doit dépasser 1 pour 5 erreurs sur 1 autorisée", burn)
	}
	if math.Abs(burn-5) > 1e-9 {
		t.Fatalf("burn rate = %g, want 5", burn)
	}
	if remaining >= 0 {
		t.Fatalf("budget restant %g, doit être négatif après dépassement", remaining)
	}
	if math.Abs(remaining+4) > 1e-9 {
		t.Fatalf("budget restant = %g, want -4", remaining)
	}
}

// TestErrorBudget_PositiveControl — le helper ne refuse pas tout : un échantillon
// conforme donne un budget plein, et une erreur exactement autorisée donne burn=1.
func TestErrorBudget_PositiveControl(t *testing.T) {
	burn, remaining, ok := ErrorBudget(100, 100, 0.99)
	if !ok || burn != 0 || remaining != 1 {
		t.Fatalf("100/100 : got (%g, %g, %v), want (0, 1, true)", burn, remaining, ok)
	}
	burn, remaining, ok = ErrorBudget(99, 100, 0.99)
	if !ok || math.Abs(burn-1) > 1e-9 || math.Abs(remaining) > 1e-9 {
		t.Fatalf("99/100 : got (%g, %g, %v), want (1, 0, true)", burn, remaining, ok)
	}
}

// TestErrorBudget_RefusesInvalidTargetsAndCounts — NaN, Inf, cible hors de
// (0,1) et OK > total sont refusés : aucune de ces entrées ne peut produire un
// burn rate ou un budget présentable comme mesuré.
func TestErrorBudget_RefusesInvalidTargetsAndCounts(t *testing.T) {
	cases := []struct {
		name      string
		ok, total uint64
		target    float64
	}{
		{"NaN", 95, 100, math.NaN()},
		{"+Inf", 95, 100, math.Inf(1)},
		{"-Inf", 95, 100, math.Inf(-1)},
		{"zéro", 95, 100, 0},
		{"un", 95, 100, 1},
		{"au-dessus de un", 95, 100, 1.5},
		{"négatif", 95, 100, -0.5},
		{"OK supérieur au total", 101, 100, 0.99},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			burn, remaining, ok := ErrorBudget(c.ok, c.total, c.target)
			if ok {
				t.Fatalf("entrée invalide acceptée : burn=%g remaining=%g", burn, remaining)
			}
			if remaining > 0 || math.IsNaN(burn) || math.IsNaN(remaining) {
				t.Fatalf("refus non sûr : burn=%g remaining=%g", burn, remaining)
			}
		})
	}
}
