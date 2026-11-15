package figures

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// Tâche 5.1 — les figures exigent des données (erreur honnête, pas de SVG vide) et
// intègrent la provenance RDF (sha256 du corpus entier + date + créateur) dans les deux SVG.
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

// ProvenanceHash8 — le sha256 8 caractères du corpus entier (chaque
// */aqm_eval.csv haché, liste hex rejointe re-hachée), stable sur
// Wall/Drawer/Archives/Report (triple provenance, registre maître).
func TestProvenanceHash8(t *testing.T) {
	if h := ProvenanceHash8(filepath.Join(t.TempDir(), "empty")); h != "" {
		t.Fatalf("no data expected empty hash, got %q", h)
	}
	dataDir := filepath.Join(t.TempDir(), "runs")
	csv := strings.Join(model.AQMEvalHeader, ",") + "\nrun-x,1,P1,pfifo_fast,bbr,1,20,22.5,46.9,100,235.1,0,0,1,11.6,12,valid\n"
	csv2 := strings.Join(model.AQMEvalHeader, ",") + "\nrun-y,1,P1,cake,bbr,1,20,22.5,46.9,100,235.1,0,0,1,11.6,12,valid\n"
	for _, r := range []struct{ run, body string }{{"run-x", csv}, {"run-y", csv2}} {
		if err := os.MkdirAll(filepath.Join(dataDir, r.run), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dataDir, r.run, "aqm_eval.csv"), []byte(r.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// attendu : chaque fichier haché (trié par chemin), liste rejointe "\n" re-hachée
	s1 := sha256.Sum256([]byte(csv))
	s2 := sha256.Sum256([]byte(csv2))
	inner := hex.EncodeToString(s1[:]) + "\n" + hex.EncodeToString(s2[:]) // run-x < run-y
	sum := sha256.Sum256([]byte(inner))
	want := hex.EncodeToString(sum[:])[:8]
	if h := ProvenanceHash8(dataDir); h != want {
		t.Fatalf("hash8 = %q, want %q (corpus entier)", h, want)
	}
	if again := ProvenanceHash8(dataDir); again != want {
		t.Fatalf("hash instable : %q puis %q", want, again)
	}
	// un octet changé dans UN run change le hash du corpus entier
	mutated := strings.Replace(csv, "22.5", "22.6", 1)
	if err := os.WriteFile(filepath.Join(dataDir, "run-x", "aqm_eval.csv"), []byte(mutated), 0o644); err != nil {
		t.Fatal(err)
	}
	if h := ProvenanceHash8(dataDir); h == want {
		t.Fatalf("mutation d'un run invisible dans le hash du corpus")
	}
}
