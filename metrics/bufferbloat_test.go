package metrics

import (
	"math"
	"testing"
)

// Bandes Waveform (waveform.com/tools/bufferbloat, vérifiées 2026-09-03) :
// note = pire écart de latence moyenne idle → chargé, download et upload
// confondus. A+ <5 ms · A <30 · B <60 · C <200 · D <400 · F ≥400.
func TestBufferbloatGrade(t *testing.T) {
	cases := []struct {
		name string
		ms   float64
		want string
	}{
		{"A+ sous 5 ms", 4.9, "A+"},
		{"A juste sous 30", 29.9, "A"},
		{"B juste sous 60", 59.9, "B"},
		{"C juste sous 200", 199.9, "C"},
		{"D juste sous 400", 399.9, "D"},
		{"F à 400", 400, "F"},
		{"F au-delà", 800, "F"},
		{"négatif clampé A+", -3, "A+"},
		{"zéro = A+", 0, "A+"},
	}
	for _, c := range cases {
		if got := BufferbloatGrade(c.ms); got != c.want {
			t.Fatalf("%s: BufferbloatGrade(%v) = %q, want %q", c.name, c.ms, got, c.want)
		}
	}
}

// La note d'un lien = le PIRE écart mesuré, download et upload confondus —
// un sens excellent ne rachète pas l'autre (méthode Waveform).
func TestBufferbloatWorst(t *testing.T) {
	if got := BufferbloatWorst(10, 55); got != 55 {
		t.Fatalf("worst(10,55) = %v, want 55 (l'upload domine)", got)
	}
	if got := BufferbloatWorst(120, 15); got != 120 {
		t.Fatalf("worst(120,15) = %v, want 120 (le download domine)", got)
	}
	// sens non mesuré (NaN) : l'autre sens décide seul
	if got := BufferbloatWorst(math.NaN(), 42); got != 42 {
		t.Fatalf("worst(NaN,42) = %v, want 42", got)
	}
	if got := BufferbloatWorst(42, math.NaN()); got != 42 {
		t.Fatalf("worst(42,NaN) = %v, want 42", got)
	}
	if !math.IsNaN(BufferbloatWorst(math.NaN(), math.NaN())) {
		t.Fatal("worst(NaN,NaN) doit rester NaN — indisponible honnête")
	}
}

// Verdict en langage opérateur — chaque note porte son remède, pas une
// description décorative.
func TestBufferbloatVerdict(t *testing.T) {
	for _, c := range []struct {
		grade, want string
	}{
		{"A+", "excellente"},
		{"A", "bonne"},
		{"B", "légère"},
		{"C", "modérée"},
		{"D", "sévère"},
		{"F", "critique"},
	} {
		v := BufferbloatVerdict(c.grade)
		if v == "" {
			t.Fatalf("verdict(%s) vide", c.grade)
		}
		if len(v) < 20 {
			t.Fatalf("verdict(%s) trop court: %q", c.grade, v)
		}
	}
	if BufferbloatVerdict("X") != "" {
		t.Fatal("grade inconnu → verdict vide")
	}
}
