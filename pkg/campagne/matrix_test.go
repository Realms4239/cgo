package campagne

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/model"
)

func TestMatrixReducedOrderAndResume(t *testing.T) {
	dir := t.TempDir()
	deps := fastDeps()

	// Démarrer la matrice réduite : P2 seul = 3 qdisc *2 CC *3 rép =18
	m, err := StartMatrixWithID(context.Background(), "run-resume-test", []string{"P2"}, 3, deps, dir)
	if err != nil {
		t.Fatal(err)
	}
	// attendre la fin (fenêtres instantanées → rapide, 18 événements ~ <500 ms)
	for i := 0; i < 50; i++ {
		if m.Done == m.Total {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	// Nous attendons 18 lignes dans le CSV
	data, err := os.ReadFile(filepath.Join(dir, "run-resume-test", "aqm_eval.csv"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(data))
	// compter les lignes de données non vides (sauter l'en-tête et le vide final)
	rows := 0
	for i, ln := range lines {
		if i == 0 {
			continue
		}
		if ln != "" {
			rows++
		}
	}
	if rows != 18 {
		t.Fatalf("want 18 rows, got %d (lines=%v)", rows, lines)
	}
	// vérification d'ordre : les premières lignes doivent être P2 pfifo_fast cubic rép1, rép2, rép3, puis bbr, puis fq_codel...
	if len(lines) < 2 {
		t.Fatal("no data")
	}
	first := splitCSV(lines[1])
	if first[2] != "P2" || first[3] != "pfifo_fast" || first[4] != "cubic" {
		t.Fatalf("first row order wrong: %v", first)
	}

	// Reprise : rouvrir le même runID avec la même matrice, 0 nouvelle ligne (tout déjà vu)
	m2, err := StartMatrixWithID(context.Background(), "run-resume-test", []string{"P2"}, 3, deps, dir)
	if err != nil {
		t.Fatal(err)
	}
	_ = m2
	time.Sleep(200 * time.Millisecond)
	// après reprise, le nombre de lignes reste 18 (pas de doublons)
	data2, _ := os.ReadFile(filepath.Join(dir, "run-resume-test", "aqm_eval.csv"))
	lines2 := splitLines(string(data2))
	rows2 := 0
	for i, ln := range lines2 {
		if i == 0 {
			continue
		}
		if ln != "" {
			rows2++
		}
	}
	if rows2 != 18 {
		t.Fatalf("resume duplicated rows: %d", rows2)
	}
	if len(model.AllQdiscs) != 3 || len(model.AllCC) != 2 {
		t.Fatal("model constants changed")
	}
}
