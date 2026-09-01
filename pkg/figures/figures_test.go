package figures

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// Tâche 5.1 — les figures exigent des données (erreur honnête, pas de SVG vide) et
// intègrent la provenance RDF (sha256 du dernier aqm_eval.csv + date + créateur) dans les deux SVG.
func TestFiguresRDF(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "figures")

	// pas de données → erreur honnête, rien d'écrit
	if err := Generate(filepath.Join(dir, "runs"), out); err == nil {
		t.Fatal("expected error with no data")
	}

	// un run, deux lignes même profile|qdisc|cc — en-tête selon model.AQMEvalHeader
	run := filepath.Join(dir, "runs", "run-20260828-0001")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := []string{
		strings.Join(model.AQMEvalHeader, ","),
		"run-20260828-0001,1,P1,pfifo_fast,bbr,1,20,22.5,2.5,85.0,46.9,100,235.1,0,0,1872264,11.6,12,valid",
		"run-20260828-0001,2,P1,pfifo_fast,bbr,2,21,22.7,1.7,84.5,47.2,100,233.8,0,0,1873302,11.7,12,valid",
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

// ProvenanceHash8 — le sha256 8 caractères du dernier aqm_eval.csv, stable
// sur Wall/Drawer/Archives/Report (triple provenance, registre maître).
func TestProvenanceHash8(t *testing.T) {
	if h := ProvenanceHash8(filepath.Join(t.TempDir(), "empty")); h != "" {
		t.Fatalf("no data expected empty hash, got %q", h)
	}
	run := filepath.Join(t.TempDir(), "runs", "run-x")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	csv := strings.Join(model.AQMEvalHeader, ",") + "\nrun-x,1,P1,pfifo_fast,bbr,1,20,22.5,46.9,100,235.1,0,0,1,11.6,12,valid\n"
	if err := os.WriteFile(filepath.Join(run, "aqm_eval.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	h := ProvenanceHash8(filepath.Dir(run))
	sum := sha256.Sum256([]byte(csv))
	want := hex.EncodeToString(sum[:])[:8]
	if h != want {
		t.Fatalf("hash8 = %q, want %q", h, want)
	}
}
