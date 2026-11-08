package metrics

import "testing"

// TestJFI — index d'équité de Jain sur les contributions per-flow.
// Égalité parfaite = 100 % ; un flux affamé fait chuter l'index même à
// total identique ; mono-flux = 0 (non applicable, pas fabriquée).
func TestJFI(t *testing.T) {
	if got := JFI([]uint64{100, 100, 100}); got != 1 {
		t.Fatalf("égalité parfaite: %v, want 1", got)
	}
	// total identique, un flux affamé : 3 flux (200,200,0) vs (133,133,134)
	if got := JFI([]uint64{200, 200, 0}); got >= 0.9 {
		t.Fatalf("famine masquée: JFI(200,200,0) = %v", got)
	}
	if got := JFI([]uint64{133, 133, 134}); got < 0.999 {
		t.Fatalf("réparti: JFI = %v, want ~1", got)
	}
	if got := JFI([]uint64{50000}); got != 0 {
		t.Fatalf("mono-flux: %v, want 0 (n.a., jamais 100 fabriqué)", got)
	}
	if got := JFI(nil); got != 0 {
		t.Fatalf("vide: %v, want 0", got)
	}
}
