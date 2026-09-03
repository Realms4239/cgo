package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSSEStaleIDResync — un client qui revient avec un Last-Event-ID antérieur
// à l'anneau (rotation des 2048 frames) doit recevoir le frame complet de
// resynchronisation : sans lui, les deltas rejoués référencent une base
// structurelle jamais reçue (état client corrompu).
func TestSSEStaleIDResync(t *testing.T) {
	h := NewHub()
	snap := map[string]any{"phase": "charge", "profile": "P1", "rtt_p95_ms": 100.0}
	for i := 0; i < 2050; i++ {
		h.Publish(snap)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/api/stream", nil)
	req.Header.Set("Last-Event-ID", "1") // évincé depuis longtemps
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"profile":"P1"`) {
		t.Fatalf("stale client got no full resync frame (no profile in %d bytes)", len(body))
	}
}

// TestSSESubLimitAtomic — 70 connexions simultanées : exactement 64 admises,
// 6 refusées 503. Le test-puis-insertion en deux temps laissait passer > 64.
func TestSSESubLimitAtomic(t *testing.T) {
	h := NewHub()
	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	const n = 70
	var admitted, refused atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+"/api/stream", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusServiceUnavailable {
				refused.Add(1)
				return
			}
			if resp.StatusCode == http.StatusOK {
				admitted.Add(1)
				_, _ = io.Copy(io.Discard, resp.Body)
			}
		}()
	}
	close(start)
	wg.Wait()
	if admitted.Load() != maxSubs || refused.Load() != n-maxSubs {
		t.Fatalf("admitted=%d refused=%d, want %d/%d", admitted.Load(), refused.Load(), maxSubs, n-maxSubs)
	}
}

// TestHubCloseStopsTicks — après Close, le fournisseur n'est plus appelé :
// la goroutine Serve s'arrête (fuite corrigée : `for range ticker.C` ne sort
// jamais d'un ticker stoppé puisque Stop ne ferme pas le canal).
func TestHubCloseStopsTicks(t *testing.T) {
	h := NewHub()
	var calls atomic.Int32
	h.Serve(func() any { calls.Add(1); return map[string]any{"phase": "idle"} })
	time.Sleep(250 * time.Millisecond) // ≥ 2 ticks à 10 Hz
	n1 := calls.Load()
	if n1 < 2 {
		t.Fatalf("hub did not tick: %d calls", n1)
	}
	h.Close()
	h.Close() // idempotent, ne panique pas
	time.Sleep(300 * time.Millisecond)
	if n2 := calls.Load(); n2 > n1+1 {
		t.Fatalf("hub still ticking after Close: %d → %d", n1, n2)
	}
}
