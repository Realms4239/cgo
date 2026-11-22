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

// TestSSEEpochMismatchMidRingEmitsFullFirst — oracle for the residual left by
// the "empty/stale replay" fix. A reconnect whose Last-Event-ID still lands
// MID-RING on a restarted hub gets non-empty deltas from a different
// generation; the browser then keeps stale structural state forever. The only
// honest repair is the full snapshot (already built in lastFull), emitted
// before any delta. Observation is the SSE wire only — no mocked Hub.SSE.
func TestSSEEpochMismatchMidRingEmitsFullFirst(t *testing.T) {
	old := NewHub()
	for i := 0; i < 100; i++ {
		old.Publish(map[string]any{"phase": "charge", "profile": "OLD", "rtt_p50_ms": 42})
	}
	oldEpoch := old.epoch
	if oldEpoch == "" {
		t.Fatal("old hub has no epoch — generation marker missing")
	}

	// Fresh hub: IDs restart at 1 and advance past the client's old ID. Its
	// epoch is a different generation even though the ID space overlaps.
	h := NewHub()
	if h.epoch == oldEpoch {
		t.Fatalf("fresh hub reused epoch %q — no generation boundary", h.epoch)
	}
	for i := 0; i < 200; i++ {
		h.Publish(map[string]any{"phase": "recup", "profile": "NEW", "rtt_p50_ms": 7})
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Last-Event-ID 100 is mid-ring (ids 1..200): replayFrom(101) is non-empty.
	req, _ := http.NewRequest("GET", srv.URL+"/api/stream?epoch="+oldEpoch, nil)
	req.Header.Set("Last-Event-ID", "100")
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)

	// Hostile precondition: the replay really is non-empty. Without it the old
	// len(replay)==0 branch would resync anyway and mask the bug.
	if !strings.Contains(body, "id: 101\n") {
		t.Fatalf("replay was empty — oracle invalid: %q", body)
	}

	// The full snapshot must precede every delta, carry the CURRENT epoch and
	// structural values, and never leak the previous generation's values.
	fullIdx := strings.Index(body, "data: {")
	deltaIdx := strings.Index(body, "id: ")
	if fullIdx < 0 {
		t.Fatalf("no full snapshot emitted on epoch mismatch: %q", body)
	}
	if fullIdx > deltaIdx {
		t.Fatalf("full snapshot emitted after deltas (full=%d delta=%d): %q", fullIdx, deltaIdx, body)
	}
	if !strings.Contains(body, `"epoch":"`+h.epoch+`"`) {
		t.Fatalf("full frame missing current epoch %q: %q", h.epoch, body)
	}
	if !strings.Contains(body, `"profile":"NEW"`) || !strings.Contains(body, `"phase":"recup"`) {
		t.Fatalf("full structural snapshot missing: %q", body)
	}
	if strings.Contains(body, `"profile":"OLD"`) {
		t.Fatalf("stale structural value leaked across generations: %q", body)
	}
}

// TestSSEEpochMatchMidRingNoFullFrame — hostile twin: on a SAME-generation
// reconnect with a mid-ring ID, no full snapshot is needed. Deltas omit
// unchanged structural keys, so "profile" must be absent from the wire. This
// fails closed if the fix degenerates into "always resync".
func TestSSEEpochMatchMidRingNoFullFrame(t *testing.T) {
	h := NewHub()
	for i := 0; i < 200; i++ {
		h.Publish(map[string]any{"phase": "recup", "profile": "NEW", "rtt_p50_ms": 7})
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/api/stream?epoch="+h.epoch, nil)
	req.Header.Set("Last-Event-ID", "100")
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "id: 101\n") {
		t.Fatalf("replay was empty — oracle invalid: %q", body)
	}
	if strings.Contains(string(body), `"profile"`) {
		t.Fatalf("full snapshot emitted on matching epoch (unneeded resync): %q", body)
	}
}
