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

	// Start reduced matrix: P2 only = 3 qdisc *2 CC *3 reps =18
	m, err := StartMatrixWithID(context.Background(), "run-resume-test", []string{"P2"}, 3, deps, NewLive(), dir)
	if err != nil {
		t.Fatal(err)
	}
	// wait for completion (instant windows -> fast, 18 events ~ <500ms)
	for i := 0; i < 50; i++ {
		if m.Done == m.Total {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	// We expect 18 rows in CSV
	data, err := os.ReadFile(filepath.Join(dir, "run-resume-test", "aqm_eval.csv"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(data))
	// count non-empty data rows (skip header and trailing empty)
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
	// order check: first rows should be P2 pfifo_fast cubic rep1, rep2, rep3, then bbr, then fq_codel...
	if len(lines) < 2 {
		t.Fatal("no data")
	}
	first := splitCSV(lines[1])
	if first[2] != "P2" || first[3] != "pfifo_fast" || first[4] != "cubic" {
		t.Fatalf("first row order wrong: %v", first)
	}

	// Resume: reopen same runID with same matrix, add 0 new rows (all seen)
	m2, err := StartMatrixWithID(context.Background(), "run-resume-test", []string{"P2"}, 3, deps, NewLive(), dir)
	if err != nil {
		t.Fatal(err)
	}
	_ = m2
	time.Sleep(200 * time.Millisecond)
	// after resume, row count still 18 (no duplicates)
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
