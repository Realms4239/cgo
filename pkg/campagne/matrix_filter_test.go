package campagne

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMatrixFilteredSubmatrix — filtres qdiscs/CC : la matrice ne parcourt
// QUE les axes demandés (P2 × cake × bbr × 2 reps = 2 événements), total
// recalculé. Vide = refus honnête, pas de matrice fantôme.
func TestMatrixFilteredSubmatrix(t *testing.T) {
	dir := t.TempDir()
	deps := fastDeps()
	m, err := StartMatrixFiltered(context.Background(), []string{"P2"}, []string{"cake"}, []string{"bbr"}, 2, deps, dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40 && m.Done != m.Total; i++ {
		time.Sleep(50 * time.Millisecond)
	}
	if m.Total != 2 {
		t.Fatalf("total = %d, want 2 (cake×bbr×2)", m.Total)
	}
	data, _ := os.ReadFile(filepath.Join(dir, m.RunID, "aqm_eval.csv"))
	lines := 0
	for i, ln := range splitLines(string(data)) {
		if i == 0 || ln == "" {
			continue
		}
		lines++
		if cols := splitCSV(ln); cols[3] != "cake" || cols[4] != "bbr" {
			t.Fatalf("ligne hors filtre: %v", cols)
		}
	}
	if lines != 2 {
		t.Fatalf("lignes gélées = %d, want 2", lines)
	}

	// filtres vides (aucun qdisc, aucune CC) → erreur, pas de run muet
	if _, err := StartMatrixFiltered(context.Background(), []string{"P2"}, nil, []string{"bbr"}, 1, deps, dir); err == nil {
		t.Fatal("filtre qdisc vide doit refuser")
	}
	if _, err := StartMatrixFiltered(context.Background(), []string{"P2"}, []string{"cake"}, nil, 1, deps, dir); err == nil {
		t.Fatal("filtre CC vide doit refuser")
	}
}
