// Package api — surface REST + SSE (docs/SPEC.md §2.4–2.5). stdlib seul.
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

// Hub diffuse des instantanés à 10 Hz avec champs structurels compressés en
// delta, rejeu Last-Event-ID (anneau 2048) et événements backpressure nommés.
type Hub struct {
	mu       sync.Mutex
	lastID   uint64
	ring     []frame // cap 2048
	subs     map[*sub]struct{}
	last     map[string]any // last structural values for delta
	lastFull string         // full snapshot json (structural included) for fresh-client sync
	ticker   *time.Ticker
	done     chan struct{} // fermé par Close : la goroutine Serve s'arrête (pas de fuite)
	closeOnce sync.Once
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
	h := &Hub{subs: map[*sub]struct{}{}, last: map[string]any{}, done: make(chan struct{})}
	return h
}

// Serve démarre la boucle de diffusion 10 Hz sur le fournisseur d'instantanés.
func (h *Hub) Serve(next func() any) {
	if h.ticker != nil {
		return
	}
	h.ticker = time.NewTicker(time.Second / frameHz)
	go func() {
		for {
			select {
			case <-h.done:
				h.ticker.Stop()
				return
			case <-h.ticker.C:
				h.Publish(next())
			}
		}
	}()
}

// Close arrête la boucle de diffusion. Idempotent — Serve ne redémarre pas
// après Close (un seul Serve par New, câblé dans api.New).
func (h *Hub) Close() {
	h.closeOnce.Do(func() { close(h.done) })
}

// Publish sérialise un instantané, encode en delta les clés structurelles et diffuse.
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
				continue // champ structurel inchangé : omis, le client conserve la valeur
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
	// copie complète (structurel inclus) — les nouveaux clients synchronisent leur état depuis elle
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
			select { // ne jamais empiler plus d'une notification backpressure
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
		}
	}
}

// SSE gère GET /api/stream : rejeu puis direct.
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
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	// pas d'en-tête Connection manuel — hop-by-hop, illégal en HTTP/2 (cassait le SSE derrière cloudflared/bord CF)
	fmt.Fprint(w, "retry: 2000\n")

	// Section atomique : le comptage et l'insertion sont sous le même verrou
	// (le test-puis-insertion en deux temps laissait passer > 64 abonnés en
	// rafale). L'anneau est copié ici, le rejeu s'écrit hors verrou.
	var from int
	lastEventID := r.Header.Get("Last-Event-ID")
	stale := false
	if lastEventID != "" {
		if id, err := strconv.ParseUint(lastEventID, 10, 64); err == nil {
			for i := len(h.ring) - 1; i >= 0; i-- {
				if h.ring[i].id <= id {
					from = i + 1
					break
				}
			}
			// ID antérieur au début de l'anneau (rotation) : le client a raté
			// la base structurelle des deltas — resynchroniser sur le frame
			// complet, sinon son état reste corrompu.
			if from == 0 && len(h.ring) > 0 {
				stale = true
			}
		} else {
			lastEventID = "" // ID illisible ⇒ traiter comme client frais
		}
	} else {
		// Sans Last-Event-ID : rejeu limité aux 5 derniers frames pour éviter une
		// tempête de 2048 frames (2 s de retard avec 20). 5 = 0,5 s de préchauffage,
		// donne un frame récent mais pas un arriéré de 2 secondes.
		if len(h.ring) > 5 {
			from = len(h.ring) - 5
		}
	}
	replay := make([]string, 0, len(h.ring)-from)
	for _, f := range h.ring[from:] {
		replay = append(replay, "id: "+strconv.FormatUint(f.id, 10)+"\ndata: "+f.raw+"\n\n")
	}
	lastFull := h.lastFull
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
	// nouveau client (sans Last-Event-ID) ou client périmé (ID antérieur à
	// l'anneau) : les frames delta omettent les champs structurels jamais
	// reçus — synchroniser avec un frame complet d'abord
	if (lastEventID == "" || stale) && lastFull != "" {
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

// ---- aides ----

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
