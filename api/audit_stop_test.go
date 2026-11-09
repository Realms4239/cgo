package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// TestAuditStopDeliversCancelOnce — un arrêt demandé deux fois ne doit pas
// répondre « stopped » deux fois : la deuxième requête n'a plus rien à
// annuler. Le défaut constaté en vrai : deux POST /api/audit/stop séparés de
// quelques secondes renvoyaient tous deux {"stopped":true}, alors que le
// statut disait encore running tant que la goroutine n'avait pas rendu la main.
func TestAuditStopDeliversCancelOnce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	auditMu.Lock()
	auditRunning = true
	auditCancel = cancel
	auditMu.Unlock()
	defer func() {
		auditMu.Lock()
		auditRunning = false
		auditCancel = nil
		auditMu.Unlock()
	}()

	h := handleAuditStop()
	call := func() bool {
		rr := httptest.NewRecorder()
		h(rr, httptest.NewRequest("POST", "/api/audit/stop", nil))
		var body struct {
			Stopped bool `json:"stopped"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("réponse illisible: %v (%s)", err, rr.Body.String())
		}
		return body.Stopped
	}

	if !call() {
		t.Fatal("premier arrêt : stopped=false, want true")
	}
	if ctx.Err() == nil {
		t.Fatal("premier arrêt : le contexte n'a pas été annulé")
	}
	if call() {
		t.Fatal("deuxième arrêt : stopped=true encore, une seule livraison attendue")
	}
}
