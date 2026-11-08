package api

import (
	"context"
	"net/http/httptest"
	"testing"
)

// TestResultsRejectsUnsafeRun — l'identifiant de run atteint un glob
// (data/runs/<run>/aqm_eval.csv) : il doit être refusé comme sur run/rows et
// replay/stream, au lieu de partir dans le chemin.
func TestResultsRejectsUnsafeRun(t *testing.T) {
	h := New(Deps{})
	for _, bad := range []string{"..", "../x", `..\x`, "a/b"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/results?run="+bad, nil)
		h.ServeHTTP(rr, req.WithContext(context.Background()))
		if rr.Code != 400 {
			t.Fatalf("run=%q → %d, want 400 (corps %q)", bad, rr.Code, rr.Body.String())
		}
	}
	// Un run inconnu mais sûr reste une réponse normale : pas de résultats.
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/results?run=run-inexistant", nil))
	if rr.Code != 200 {
		t.Fatalf("run sûr inconnu → %d, want 200", rr.Code)
	}
}
