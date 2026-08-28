package figures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// Task 5.1 — figures need data (honest error, no empty SVG) and embed RDF
// provenance (sha256 of latest aqm_eval.csv + date + creator) in both SVGs.
func TestFiguresRDF(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "figures")

	// no data → honest error, nothing written
	if err := Generate(filepath.Join(dir, "runs"), out); err == nil {
		t.Fatal("expected error with no data")
	}

	// one run, two rows same profile|qdisc|cc — header per model.AQMEvalHeader
	run := filepath.Join(dir, "runs", "run-20260828-0001")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := []string{
		strings.Join(model.AQMEvalHeader, ","),
		"run-20260828-0001,1,P1,pfifo_fast,bbr,1,20,22.5,46.9,100,235.1,0,0,1872264,11.6,12,valid",
		"run-20260828-0001,2,P1,pfifo_fast,bbr,2,21,22.7,47.2,100,233.8,0,0,1873302,11.7,12,valid",
	}
	if err := os.WriteFile(filepath.Join(run, "aqm_eval.csv"), []byte(strings.Join(rows, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(filepath.Join(dir, "runs"), out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"small_p95.svg", "scatter.svg"} {
		b, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.Contains(s, "rdf:RDF") || !strings.Contains(s, "sha256:") || !strings.Contains(s, "dc:creator") {
			t.Fatalf("%s missing RDF provenance", name)
		}
	}
}
