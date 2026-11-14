package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestObserveModeReturns501 — chaque route de mutation répond 501 en mode
// observation (l'hôte observe ; le banc mesure). Les quatre sites déjà
// gardés et les neuf complétés sont épinglés par la même table.
func TestObserveModeReturns501(t *testing.T) {
	paths := []string{
		"/api/cost/tier", "/api/burst", "/api/watch", "/api/shape",
		"/api/run/start", "/api/run/stop", "/api/run/skip",
		"/api/figures/regen",
		"/api/audit/start", "/api/audit/stop", "/api/audit/toprofile",
		"/api/profile/import", "/api/profile/delete",
	}
	h := New(Deps{Mode: "observe"})
	for _, p := range paths {
		req := httptest.NewRequest("POST", p, strings.NewReader("{}"))
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotImplemented {
			t.Fatalf("POST %s en mode observe : %d, want 501 (corps %q)", p, rec.Code, rec.Body.String())
		}
	}
}

// TestObserveModeDoesNotRunSideEffects — la garde précède tout : en mode
// observe, une mutation refusée n'exécute ni moteur ni écriture.
func TestObserveModeDoesNotRunSideEffects(t *testing.T) {
	called := false
	h := New(Deps{Mode: "observe", StopFn: func() { called = true }})
	req := httptest.NewRequest("POST", "/api/run/stop", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("run/stop observe : %d, want 501", rec.Code)
	}
	if called {
		t.Fatal("StopFn exécuté malgré le mode observation")
	}
}
