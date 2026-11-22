package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Hypothesis oracle (Task 3, Phase 3). A reconnecting browser sends the
// Last-Event-ID it held on the previous hub instance. Hub IDs restart at 1 on
// every process start, so after a restart the old ID can collide with the new
// ID space. The client can only interpret delta frames against a structural
// base it has actually received: without a full resync frame it keeps the old
// phase/qdisc/cc values forever. The oracle observes only the SSE wire output.
func TestSSEReconnectAfterHubRestartGetsStructuralResync(t *testing.T) {
	const oldLastID = "200"

	// Old instance: 200 frames, last event id = 200, phase "charge".
	old := NewHub()
	for i := 0; i < 200; i++ {
		old.Publish(map[string]any{"phase": "charge", "rtt_p50_ms": 42})
	}
	if old.lastID != 200 {
		t.Fatalf("old hub lastID = %d, want 200", old.lastID)
	}

	// Fresh instance: IDs restart at 1 and advance past the old client's ID.
	// Structural values differ ("recup"): deltas at ids <= 200 carried the
	// change, but the client never saw those frames on this instance.
	h := NewHub()
	for i := 0; i < 200; i++ {
		h.Publish(map[string]any{"phase": "recup", "rtt_p50_ms": 7})
	}
	if h.lastID != 200 {
		t.Fatalf("fresh hub lastID = %d, want 200", h.lastID)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/api/stream", nil)
	req.Header.Set("Last-Event-ID", oldLastID)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// The client must receive a full frame carrying the CURRENT structural
	// value. Without it, it silently keeps "charge" while the hub is "recup".
	if !strings.Contains(string(body), `"phase":"recup"`) {
		t.Fatalf("no structural resync after hub restart (client keeps stale phase): %q", body)
	}
}
