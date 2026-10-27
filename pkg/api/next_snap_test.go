package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Le hub 10 Hz doit lire NextSnap (échantillon frais) ; /api/state doit lire
// GetSnap (dernier échantillon). Un poll d'état ne raccourcit jamais une
// fenêtre SSE.
func TestNextSnapDrivesHubAndNotState(t *testing.T) {
	var hubCalls, stateCalls atomic.Int64
	h := New(Deps{
		GetSnap: func() any {
			stateCalls.Add(1)
			return map[string]any{"running": false, "src": "state"}
		},
		NextSnap: func() any {
			hubCalls.Add(1)
			return map[string]any{"running": false, "src": "hub"}
		},
	})
	defer h.CloseHub()
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `"src":"state"`) {
		t.Fatalf("/api/state n'a pas utilisé GetSnap : %s", body)
	}

	time.Sleep(250 * time.Millisecond)
	if got := hubCalls.Load(); got < 2 {
		t.Fatalf("hub NextSnap calls = %d, want >= 2", got)
	}
	if got := stateCalls.Load(); got != 1 {
		t.Fatalf("GetSnap calls = %d, want 1 (le hub ne doit pas l'appeler)", got)
	}
}
