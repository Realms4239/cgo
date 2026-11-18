package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestQuarantineSummaryReportsMalformedRows — le résumé expose malformed_rows
// sans retirer ni renommer aucune clé historique. Oracle indépendant : le
// nombre de lignes dépassant l'en-tête, compté à la main dans les fixtures.
func TestQuarantineSummaryReportsMalformedRows(t *testing.T) {
	withDevNoAuth(t)
	dir := chdirTemp(t)

	// run-a : 1 ligne courte (tolérée) + 1 ligne avec EXTRA (malformée) → 1.
	// run-b : 1 ligne courte + 2 lignes avec EXTRA → 2. Total attendu : 3.
	fixtures := map[string]string{
		"run-a": "profile,gate_status\nP2,valid\nP2\nP2,valid,EXTRA\n",
		"run-b": "profile,gate_status\nP2\nP2,x,EXTRA\nP2,y,EXTRA\n",
	}
	for run, csv := range fixtures {
		r := filepath.Join(dir, "data", "runs", run)
		if err := os.MkdirAll(r, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r, "aqm_eval.csv"), []byte(csv), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/quarantine/summary")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&got); err != nil {
		t.Fatal(err)
	}

	// Oracle positif : agrégation sur les deux runs, courtes exclues.
	if got["malformed_rows"] != float64(3) {
		t.Fatalf("malformed_rows = %v, want 3 : %v", got["malformed_rows"], got)
	}
	// Oracle hostile : jeu de clés exact, aucune clé historique perdue ou ajoutée.
	wantKeys := []string{
		"total_invalid", "g4_low", "g4_high", "empty_probes",
		"g3_implausible", "note", "malformed_rows",
	}
	if len(got) != len(wantKeys) {
		t.Fatalf("clés = %v, want exactement %v", got, wantKeys)
	}
	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Fatalf("clé absente %q : %v", k, got)
		}
	}
}
