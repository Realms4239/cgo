// Package api — REST + SSE surface (docs/SPEC.md §2.4–2.5). stdlib only.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const frameHz = 10

// Hub broadcasts snapshots at 10 Hz with delta-compressed structural
// fields, Last-Event-ID replay (ring 2048) and named backpressure events.
type Hub struct {
	mu       sync.Mutex
	lastID   uint64
	ring     []frame // cap 2048
	subs     map[*sub]struct{}
	last     map[string]any // last structural values for delta
	lastFull string         // full snapshot json (structural included) for fresh-client sync
	ticker   *time.Ticker
}

type frame struct {
	id   uint64
	raw  string
	name string // "" = metric message; "backpressure" = named
}

type sub struct {
	ch      chan string
	dropped uint64
}

func NewHub() *Hub {
	h := &Hub{subs: map[*sub]struct{}{}, last: map[string]any{}}
	return h
}

// Serve starts the 10 Hz broadcast loop over the snapshot provider.
func (h *Hub) Serve(next func() any) {
	if h.ticker != nil {
		return
	}
	h.ticker = time.NewTicker(time.Second / frameHz)
	go func() {
		for range h.ticker.C {
			h.Publish(next())
		}
	}()
}

func (h *Hub) Close() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
}

// Publish marshals one snapshot, delta-encodes structural keys and fans out.
var structuralKeys = []string{"profile", "qdisc", "cc", "repetition", "event_id", "phase"}

func (h *Hub) Publish(snap any) {
	src, err := toMap(snap)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make(map[string]any, len(src))
	ts := time.Now().UnixMilli()
	for k, v := range src {
		if contains(structuralKeys, k) {
			if eq, ok := h.last[k]; ok && fmt.Sprint(eq) == fmt.Sprint(v) {
				continue // unchanged structural field: omitted, client retains
			}
			(h.last)[k] = v
		}
		out[k] = v
	}
	out["ts"] = ts
	raw, err := json.Marshal(out)
	if err != nil {
		return
	}
	// full copy (structural included) — fresh clients sync state from it
	full := make(map[string]any, len(src)+1)
	for k, v := range src {
		full[k] = v
	}
	full["ts"] = ts
	if fullRaw, err := json.Marshal(full); err == nil {
		h.lastFull = string(fullRaw)
	}
	h.lastID++
	f := frame{id: h.lastID, raw: string(raw)}
	h.ring = append(h.ring, f)
	if len(h.ring) > 2048 {
		h.ring = h.ring[1:]
	}
	payload := "id: " + strconv.FormatUint(f.id, 10) + "\ndata: " + f.raw + "\n\n"
	for s := range h.subs {
		select {
		case s.ch <- payload:
		default:
			s.dropped++
			select { // never stack more than one backpressure notice
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
		}
	}
}

// SSE handles GET /api/stream: replay then live.
func (h *Hub) SSE(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	h.mu.Lock()
	if len(h.subs) >= maxSubs {
		h.mu.Unlock()
		http.Error(w, "too many streams", http.StatusServiceUnavailable)
		return
	}
	h.mu.Unlock()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	// no manual Connection header — hop-by-hop, illegal over HTTP/2 (broke SSE behind cloudflared/CF edge)
	fmt.Fprint(w, "retry: 2000\n")

	h.mu.Lock()
	var from int
	if les := r.Header.Get("Last-Event-ID"); les != "" {
		if id, err := strconv.ParseUint(les, 10, 64); err == nil {
			for i := len(h.ring) - 1; i >= 0; i-- {
				if h.ring[i].id <= id {
					from = i + 1
					break
				}
			}
		}
	} else {
		// No Last-Event-ID: cap replay to last 5 frames to avoid 2048-frame
		// storm (was 2s stale with 20). 5 = 0.5s warm-up, still gives client
		// a recent warm frame but not a 2-second backlog.
		if len(h.ring) > 5 {
			from = len(h.ring) - 5
		}
	}
	replay := make([]string, 0, len(h.ring)-from)
	for _, f := range h.ring[from:] {
		replay = append(replay, "id: "+strconv.FormatUint(f.id, 10)+"\ndata: "+f.raw+"\n\n")
	}
	s := &sub{ch: make(chan string, 16)}
	h.subs[s] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.subs, s)
		h.mu.Unlock()
	}()

	for _, p := range replay {
		if _, err := fmt.Fprint(w, p); err != nil {
			return
		}
	}
	// fresh client (no Last-Event-ID): delta frames omit unchanged structural
	// fields the client never received — sync state with one full frame first
	h.mu.Lock()
	lastFull := h.lastFull
	h.mu.Unlock()
	if r.Header.Get("Last-Event-ID") == "" && lastFull != "" {
		fmt.Fprintf(w, "data: %s\n\n", lastFull)
	}
	fl.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			if _, err := fmt.Fprint(w, msg); err != nil {
				return
			}
			fl.Flush()
		}
	}
}

// ---- helpers ----

func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	return m, json.Unmarshal(b, &m)
}

func contains(ks []string, k string) bool {
	for _, x := range ks {
		if strings.EqualFold(x, k) {
			return true
		}
	}
	return false
}
