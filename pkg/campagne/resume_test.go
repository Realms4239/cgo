package campagne

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestResumeWithID — reprise d'un run interrompu : les cellules déjà gelées
// (run/event vus dans le CSV) sont sautées, seules les manquantes tournent.
// Prouvé : la 2e passe avec le même runID ne double aucune ligne (G5).
func TestResumeWithID(t *testing.T) {
	dir := t.TempDir()
	deps := fastDeps()
	deps.BaselineSec, deps.ChargeSec, deps.RecupSec = -1, -1, -1

	// passe 1 : matrice complète 2 qdiscs × 1 cc × 1 rep
	m1, err := StartMatrixFilteredWithID(context.Background(), "run-resume", []string{"P2"}, []string{"pfifo_fast", "fq_codel"}, []string{"bbr"}, 1, deps, dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100 && m1.IsRunning(); i++ {
		time.Sleep(20 * time.Millisecond)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "run-resume", "aqm_eval.csv"))
	lines1 := 0
	for _, l := range splitLines(string(b)) {
		if l != "" {
			lines1++
		}
	}
	if lines1 != 3 { // en-tête + 2 cellules
		t.Fatalf("passe 1 : %d lignes want 3 (en-tête + 2)", lines1)
	}

	// passe 2 : MÊME runID, une cellule manquante en plus (cake) — seules
	// les non-gelées s'exécutent ; les 2 gelées sont sautées.
	m2, err := StartMatrixFilteredWithID(context.Background(), "run-resume", []string{"P2"}, []string{"pfifo_fast", "fq_codel", "cake"}, []string{"bbr"}, 1, deps, dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100 && m2.IsRunning(); i++ {
		time.Sleep(20 * time.Millisecond)
	}
	b2, _ := os.ReadFile(filepath.Join(dir, "run-resume", "aqm_eval.csv"))
	lines2 := 0
	for _, l := range splitLines(string(b2)) {
		if l != "" {
			lines2++
		}
	}
	if lines2 != 4 { // en-tête + 3 cellules (2 reprises + 1 nouvelle)
		t.Fatalf("passe 2 : %d lignes want 4 (reprise sans doublon)", lines2)
	}
	if m2.Done != 3 {
		t.Fatalf("done=%d want 3", m2.Done)
	}

	// runID vide : refus net, pas de matrice fantôme.
	if _, err := StartMatrixFilteredWithID(context.Background(), "", []string{"P2"}, nil, nil, 1, deps, dir); err == nil {
		t.Fatal("runID vide accepté")
	}
}
