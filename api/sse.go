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

const frameHz = 20

// ringCap — taille fixe de l'anneau : 4096 frames soit ~205 s à 20 Hz.
// Un tableau statique évite le décalage O(n) de l'ancienne tranche glissante.
const ringCap = 4096

// subCap — capacité du canal de chaque abonné SSE. Au-delà, l'éviction
// prioritaire garantit qu'un jalon n'est jamais sacrifié à un tick.
const subCap = 32

// milestoneKinds — les six jalons garantis (docs/MUSTREAD/13-chunk-3).
var milestoneKinds = map[string]bool{
	"CALIBRATION_BASELINE":  true,
	"BULK_SATURATION_START": true,
	"SLA_BREACH_220MS":      true,
	"QDISC_DROP_BURST":      true,
	"AQM_STABILIZED":        true,
	"DUAL_SEAL_COMMITTED":   true,
}

func isMilestonePayload(payload string) bool {
	for k := range milestoneKinds {
		if strings.Contains(payload, "event: "+k) {
			return true
		}
	}
	return false
}

// Hub diffuse des instantanés à 20 Hz avec champs structurels compressés en
// delta, rejeu Last-Event-ID (anneau 4096) et événements backpressure nommés.
type Hub struct {
	mu        sync.Mutex
	lastID    uint64
	ring      [ringCap]frame // anneau O(1) : head/count modulo ringCap
	head      int            // prochaine position d'écriture
	count     int            // nombre de frames valides dans l'anneau
	subs      map[*sub]struct{}
	last      map[string]any // last structural values for delta
	lastFull  string         // full snapshot json (structural included) for fresh-client sync
	ticker    *time.Ticker
	done      chan struct{} // fermé par Close : la goroutine Serve s'arrête (pas de fuite)
	closeOnce sync.Once
}

type frame struct {
	id   uint64
	raw  string
	name string // "" = message métrique ; sinon jalon rejoué avec une ligne event:
}

type sub struct {
	ch      chan string
	dropped uint64
}

func NewHub() *Hub {
	h := &Hub{subs: map[*sub]struct{}{}, last: map[string]any{}, done: make(chan struct{})}
	return h
}

// Serve démarre la boucle de diffusion 20 Hz sur le fournisseur d'instantanés.
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

// Subscribers — nombre d'abonnés SSE vivants (série mesurée de /metrics).
func (h *Hub) Subscribers() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
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
	h.ring[h.head] = f
	h.head = (h.head + 1) % ringCap
	if h.count < ringCap {
		h.count++
	}
	payload := "id: " + strconv.FormatUint(f.id, 10) + "\ndata: " + f.raw + "\n\n"
	h.deliver(payload)
}

// deliver diffuse un payload à tous les abonnés (appelé sous h.mu). Les jalons
// sont prioritaires : sur un canal plein, ils évincent un tick périmé plutôt
// que d'être abandonnés.
func (h *Hub) deliver(payload string) {
	priority := isMilestonePayload(payload)
	for s := range h.subs {
		select {
		case s.ch <- payload:
			continue
		default:
		}
		if !priority {
			s.dropped++
			select { // ne jamais empiler plus d'une notification backpressure
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
			continue
		}
		// Jalon : évincer un tick périmé, jamais un autre jalon.
		for i := 0; i < cap(s.ch); i++ {
			old, ok := <-s.ch
			if !ok {
				break
			}
			if isMilestonePayload(old) {
				s.ch <- old
				continue
			}
			s.dropped++
			break
		}
		select {
		case s.ch <- payload:
		default: // aucun tick évincable : dernier recours, signaler backpressure
			s.dropped++
			select {
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
		}
	}
}

// PublishEvent écrit un jalon garanti dans l'anneau et le diffuse. Le nom du
// jalon est conservé dans la frame pour le rejeu Last-Event-ID.
func (h *Hub) PublishEvent(kind string, detail map[string]any) {
	raw, err := json.Marshal(detail)
	if err != nil {
		raw = []byte("{}")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastID++
	f := frame{id: h.lastID, raw: string(raw), name: kind}
	h.ring[h.head] = f
	h.head = (h.head + 1) % ringCap
	if h.count < ringCap {
		h.count++
	}
	h.deliver("event: " + kind + "\ndata: " + string(raw) + "\n\n")
}

// replayFrom — frames encore présents dans l'anneau, du plus ancien au plus
// récent, en commençant à id >= from. Ignorer les IDs périmés évite de
// rejouer des deltas dont la base structurelle a été évincée.
func (h *Hub) replayFrom(from uint64) []string {
	out := make([]string, 0, h.count)
	start := (h.head - h.count + ringCap) % ringCap
	for i := 0; i < h.count; i++ {
		f := h.ring[(start+i)%ringCap]
		if f.id < from {
			continue
		}
		if f.name != "" { // jalon : le rejeu doit porter le nom de l'événement
			out = append(out, "event: "+f.name+"\nid: "+strconv.FormatUint(f.id, 10)+"\ndata: "+f.raw+"\n\n")
			continue
		}
		out = append(out, "id: "+strconv.FormatUint(f.id, 10)+"\ndata: "+f.raw+"\n\n")
	}
	return out
}

// SSE gère GET /api/stream : rejeu puis direct.
func (h *Hub) SSE(w http.ResponseWriter, r *http.Request) {
	// Certains ResponseWriter (tests, proxys) ne supportent pas la deadline :
	// on continue sans échouer le flux pour autant.
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
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
	var replay []string
	lastEventID := r.Header.Get("Last-Event-ID")
	stale := false
	if lastEventID != "" {
		if id, err := strconv.ParseUint(lastEventID, 10, 64); err == nil {
			if h.count > 0 {
				start := (h.head - h.count + ringCap) % ringCap
				// ID antérieur au début de l'anneau (rotation) : le client a raté
				// la base structurelle des deltas — resynchroniser sur le frame
				// complet, sinon son état reste corrompu.
				if id < h.ring[start].id {
					stale = true
				}
			}
			replay = h.replayFrom(id + 1)
		} else {
			lastEventID = "" // ID illisible ⇒ traiter comme client frais
		}
	} else {
		// Sans Last-Event-ID : 5 derniers frames seulement (0,25 s à 20 Hz).
		if h.count > 5 {
			start := (h.head - h.count + ringCap) % ringCap
			replay = h.replayFrom(h.ring[(start+h.count-5)%ringCap].id)
		} else {
			replay = h.replayFrom(0)
		}
	}
	lastFull := h.lastFull
	s := &sub{ch: make(chan string, subCap)}
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
