package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/figures"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/results"
	frontend "github.com/Realms4239/cgo/web/frontend"
)

// ShapeReq — the full lever: file d'attente + conditions du lien (Q8).
type ShapeReq struct {
	Qdisc    string  `json:"qdisc"`
	CapMbps  float64 `json:"capacity_mbps"`
	DelayMs  float64 `json:"delay_ms"`
	JitterMs float64 `json:"jitter_ms"`
	LossPct  float64 `json:"loss_pct"`
}

// RunOpts — what an operator actually decides (Q10): profils, répétitions,
// deadline small p95 et cible de mesure.
type RunOpts struct {
	Profiles   []string `json:"profiles"`
	Reps       int      `json:"reps"`
	DeadlineMs int      `json:"deadline_ms"`
	Target     string   `json:"target"`
}

// Deps wires the server to the campagne core.
type Deps struct {
	GetSnap func() any
	StartFn func(o RunOpts) error
	StopFn  func()
	// ShapeFn applies the queue discipline + link conditions to the edge
	// gateway (ARG.md pivot: edge shaping is the lever the DSI controls).
	// "none" clears.
	ShapeFn func(r ShapeReq) error
	// RunningFn reports an active campagne — the shape lever and a running
	// matrix fight over the same shaper, so manual shaping is refused mid-run.
	RunningFn func() bool
	// WatchFn toggles the non-intrusive surveillance loop (ping+small, no
	// bulk): the wall stays alive outside a campagne. Watching does NOT
	// conflict with shaping — that combination is the product.
	WatchFn func(on bool) error
	// Mode — observation hosts refuse the control endpoints (Q10): audit and
	// consultation work, campagne/façonnage/surveillance answer 501. "" is full.
	Mode string
	// Version — stamped at build (-X main.version), surfaced by /api/health.
	Version string
	// DoctorFn — capability report for GET /api/doctor; nil ⇒ mode only.
	DoctorFn func() any
}

// observeBlocked — Windows hosts observe (Q10): the audit works, the control
// endpoints answer 501 with the pointer to the bench. True = answered.
func (d Deps) observeBlocked(w http.ResponseWriter) bool {
	if d.Mode != "observe" {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": "mode observation — campagne/façonnage/surveillance nécessitent l'hôte Linux (voir docs/deploy.md)"})
	return true
}

// RecordEvent exports the operator journal to hosts wiring campagne
// callbacks (quarantined cells land here from the matrix loop).
func RecordEvent(kind, msg string) { recordEvent(kind, msg) }

// Handler carries the hub lifecycle so hosts can stop the 10 Hz ticker on
// shutdown (Server.Close alone never does).
type Handler struct {
	http.Handler
	hubClose func()
}

// CloseHub stops the hub broadcast ticker. Idempotent.
func (h Handler) CloseHub() {
	if h.hubClose != nil {
		h.hubClose()
	}
}

var auditMu sync.Mutex
var lastAudit *audit.Result
var auditRunning bool

// Journal — mémoire opérateur (Q13): anneau des 50 derniers événements.
var eventsMu sync.Mutex
var eventsRing []map[string]any

func recordEvent(kind, msg string) {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	eventsRing = append(eventsRing, map[string]any{"ts": time.Now().Format(time.RFC3339), "kind": kind, "msg": msg})
	if len(eventsRing) > 50 {
		eventsRing = eventsRing[len(eventsRing)-50:]
	}
}

// shapeState — last applied edge shaping (observability of the control).
var shapeMu sync.Mutex
var shapeQdisc string
var shapeCap float64
var shapeSince time.Time

func shapeApply(fn func(r ShapeReq) error, req ShapeReq, running func() bool) (int, any) {
	q, cap := req.Qdisc, req.CapMbps
	if running != nil && running() {
		return http.StatusConflict, map[string]any{"error": "campagne active — arrêtez la mesure avant de façonner le bord"}
	}
	if fn == nil {
		return http.StatusServiceUnavailable, map[string]any{"error": "shape engine not wired on this host"}
	}
	valid := map[string]bool{"cake": true, "fq_codel": true, "pfifo_fast": true, "none": true}
	if !valid[q] {
		return http.StatusBadRequest, map[string]any{"error": "qdisc must be cake | fq_codel | pfifo_fast | none"}
	}
	if cap < 1 || cap > 1000 {
		return http.StatusBadRequest, map[string]any{"error": "capacity_mbps must be 1–1000"}
	}
	if req.DelayMs < 0 || req.DelayMs > 600 || req.JitterMs < 0 || req.JitterMs > 100 || req.LossPct < 0 || req.LossPct > 10 {
		return http.StatusBadRequest, map[string]any{"error": "conditions du lien: délai 0–600 ms, gigue 0–100 ms, perte 0–10 %"}
	}
	if err := fn(req); err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	shapeMu.Lock()
	shapeQdisc, shapeCap, shapeSince = q, cap, time.Now()
	shapeMu.Unlock()
	cond := ""
	if req.DelayMs > 0 || req.JitterMs > 0 || req.LossPct > 0 {
		cond = fmt.Sprintf(", conditions %gms/%gms/%g%%", req.DelayMs, req.JitterMs, req.LossPct)
	}
	recordEvent("façonnage", fmt.Sprintf("%s @ %g Mbit/s%s", q, cap, cond))
	return http.StatusOK, map[string]any{"qdisc": q, "capacity_mbps": cap, "since": shapeSince.Format(time.RFC3339)}
}

const maxSubs = 64

// New builds the full handler with SPA fallback. The returned Handler exposes
// CloseHub for graceful shutdown (stops the 10 Hz broadcast ticker).
func New(d Deps) Handler {
	hub := NewHub()
	mux := http.NewServeMux()

	// nil provider defaults to an idle snapshot — honest emptiness, never a nil-call panic
	snap := d.GetSnap
	if snap == nil {
		snap = func() any { return map[string]any{"running": false} }
	}

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "version": d.Version, "mode": d.Mode})
	})
	mux.HandleFunc("GET /api/doctor", func(w http.ResponseWriter, _ *http.Request) {
		if d.DoctorFn != nil {
			writeJSON(w, d.DoctorFn())
			return
		}
		writeJSON(w, map[string]any{"mode": d.Mode, "checks": []any{}})
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, snap())
	})
	// ARG.md edge control — dynamic profile registry (imported profiles included)
	mux.HandleFunc("GET /api/profiles", func(w http.ResponseWriter, _ *http.Request) {
		ids := make([]string, 0, len(model.Profiles))
		for id := range model.Profiles {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		type prof struct {
			ID         string  `json:"id"`
			Capacity   float64 `json:"capacity_mbps"`
			DelayMs    float64 `json:"delay_ms"`
			JitterMs   float64 `json:"jitter_ms"`
			LossPct    float64 `json:"loss_pct"`
			FromImport bool    `json:"imported"`
		}
		out := make([]prof, 0, len(ids))
		for _, id := range ids {
			p := model.Profiles[id]
			out = append(out, prof{ID: id, Capacity: p.CapacityMbps, DelayMs: p.DelayMs, JitterMs: p.JitterMs, LossPct: p.LossPct, FromImport: id != "P1" && id != "P2" && id != "P3"})
		}
		writeJSON(w, map[string]any{"profiles": out})
	})
	// ARG.md edge control — apply / observe shaping on the gateway
	mux.HandleFunc("GET /api/shape", func(w http.ResponseWriter, _ *http.Request) {
		shapeMu.Lock()
		defer shapeMu.Unlock()
		if shapeQdisc == "" {
			writeJSON(w, map[string]any{"applied": false})
			return
		}
		writeJSON(w, map[string]any{"applied": true, "qdisc": shapeQdisc, "capacity_mbps": shapeCap, "since": shapeSince.Format(time.RFC3339)})
	})
	// surveillance continue — non-intrusive, composable with the shape lever
	mux.HandleFunc("POST /api/watch", func(w http.ResponseWriter, r *http.Request) {
		if d.observeBlocked(w) {
			return
		}
		var body struct {
			On bool `json:"on"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if d.WatchFn == nil {
			http.Error(w, "watch engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		if body.On && d.RunningFn != nil && d.RunningFn() {
			http.Error(w, "campagne active — la surveillance se lance hors campagne", http.StatusConflict)
			return
		}
		if err := d.WatchFn(body.On); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"watch": body.On})
	})
	mux.HandleFunc("POST /api/shape", func(w http.ResponseWriter, r *http.Request) {
		if d.observeBlocked(w) {
			return
		}
		var req ShapeReq
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		code, out := shapeApply(d.ShapeFn, req, d.RunningFn)
		if code != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		writeJSON(w, out)
	})
	mux.HandleFunc("POST /api/run/start", func(w http.ResponseWriter, r *http.Request) {
		if d.observeBlocked(w) {
			return
		}
		var opts RunOpts
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&opts); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		// prevention — the client is a hint, never the contract (§6); input is
		// validated even on hosts without a wired engine
		if opts.Reps < 1 || opts.Reps > 5 {
			http.Error(w, "reps must be 1–5", http.StatusBadRequest)
			return
		}
		// Q10 bounds — the deadline travels with the run, inside the same
		// window the Réglages drawer enforces client-side
		if opts.DeadlineMs != 0 && (opts.DeadlineMs < 200 || opts.DeadlineMs > 5000) {
			http.Error(w, "deadline_ms must be 200–5000", http.StatusBadRequest)
			return
		}
		if len(opts.Target) > 64 {
			http.Error(w, "target must be ≤ 64 characters", http.StatusBadRequest)
			return
		}
		if len(opts.Profiles) == 0 {
			http.Error(w, "aucun profil sélectionné", http.StatusBadRequest)
			return
		}
		var unknown []string
		for _, id := range opts.Profiles {
			if _, ok := model.Profiles[id]; !ok {
				unknown = append(unknown, id)
			}
		}
		if len(unknown) > 0 {
			http.Error(w, "profils inconnus: "+strings.Join(unknown, ", "), http.StatusBadRequest)
			return
		}
		if d.StartFn == nil {
			http.Error(w, "run engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		if err := d.StartFn(opts); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		recordEvent("campagne", fmt.Sprintf("démarrée — %s ×%d, deadline %d ms, cible %s", strings.Join(opts.Profiles, "/"), opts.Reps, opts.DeadlineMs, opts.Target))
		writeJSON(w, map[string]any{"started": true})
	})
	mux.HandleFunc("POST /api/run/stop", func(w http.ResponseWriter, _ *http.Request) {
		if d.StopFn == nil {
			http.Error(w, "run engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		d.StopFn()
		recordEvent("campagne", "arrêtée")
		writeJSON(w, map[string]any{"stopped": true})
	})
	mux.HandleFunc("GET /api/results", func(w http.ResponseWriter, r *http.Request) {
		run := r.URL.Query().Get("run")
		groups, _ := results.Scan("data/runs", run)
		if len(groups) == 0 {
			writeJSON(w, map[string]any{"available": false, "reason": "résultats disponibles après gel (jalon M2)"})
			return
		}
		writeJSON(w, map[string]any{"available": true, "groups": groups})
	})
	mux.HandleFunc("GET /api/integrity", func(w http.ResponseWriter, _ *http.Request) {
		runs, _ := filepath.Glob("data/runs/*")
		valid, quarantined := 0, 0
		manifests := 0
		var runIDs []string
		for _, p := range runs {
			if st, err := os.Stat(filepath.Join(p, "manifest.json")); err == nil && !st.IsDir() {
				manifests++
			}
			gs, _ := results.Scan("data/runs", filepath.Base(p))
			for _, g := range gs {
				valid += g.Count - g.Quarantined
				quarantined += g.Quarantined
			}
			runIDs = append(runIDs, filepath.Base(p))
		}
		if len(runs) == 0 {
			writeJSON(w, map[string]any{"available": false, "reason": "intégrité disponible après gel (jalon M2)"})
			return
		}
		writeJSON(w, map[string]any{"available": true, "runs": len(runs), "manifests": manifests, "valid": valid, "quarantined": quarantined, "run_ids": runIDs, "sha256": figures.ProvenanceSHA("data/runs"), "hash8": figures.ProvenanceHash8("data/runs")})
	})
	mux.HandleFunc("GET /api/report/export", func(w http.ResponseWriter, r *http.Request) {
		fmtParam := r.URL.Query().Get("format")
		if fmtParam == "" {
			fmtParam = "md"
		}
		groups, _ := results.Scan("data/runs", "")
		if len(groups) == 0 {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}
		if fmtParam == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
			w.Write([]byte("profile,qdisc,cc,count,small_p95_median,rtt_p95_median,goodput_median,quarantined,best\n"))
			for _, g := range groups {
				best := ""
				if g.Best {
					best = "1"
				}
				w.Write([]byte(fmt.Sprintf("%s,%s,%s,%d,%.1f,%.1f,%.1f,%d,%s\n", g.Profile, g.Qdisc, g.CC, g.Count, g.Smallp95Median, g.RTTp95Median, g.GoodputMedian, g.Quarantined, best)))
			}
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write([]byte("# CGO Report — LIEN\n\n| profil | qdisc | cc | n | small p95 | rtt p95 | goodput | quarant. | best |\n|---|---|---|---|---|---|---|---|---|\n"))
		for _, g := range groups {
			mark := ""
			if g.Best {
				mark = "★"
			}
			w.Write([]byte(fmt.Sprintf("| %s | %s | %s | %d | %.1f | %.1f | %.1f | %d | %s |\n", g.Profile, g.Qdisc, g.CC, g.Count, g.Smallp95Median, g.RTTp95Median, g.GoodputMedian, g.Quarantined, mark)))
		}
		// triple-provenance — same hash8 as Wall drawer / Archives / figures RDF
		if hash8 := figures.ProvenanceHash8("data/runs"); hash8 != "" {
			fmt.Fprintf(w, "\nprovenance : sha256:%s · source data/runs/*/aqm_eval.csv · généré %s\n", hash8, time.Now().UTC().Format(time.RFC3339))
		}
	})

	mux.HandleFunc("POST /api/figures/regen", func(w http.ResponseWriter, _ *http.Request) {
		if err := figures.Generate("data/runs", "data/figures"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	})
	mux.Handle("/api/figures/", http.StripPrefix("/api/figures/", http.FileServer(http.Dir("data/figures"))))
	mux.HandleFunc("GET /api/replay/list", func(w http.ResponseWriter, _ *http.Request) {
		runs, _ := filepath.Glob("data/runs/*")
		var ids []string
		for _, p := range runs {
			if st, err := os.Stat(filepath.Join(p, "aqm_eval.csv")); err == nil && !st.IsDir() {
				ids = append(ids, filepath.Base(p))
			}
		}
		writeJSON(w, map[string]any{"runs": ids})
	})
	mux.HandleFunc("GET /api/replay/stream", func(w http.ResponseWriter, r *http.Request) {
		run := r.URL.Query().Get("run")
		if run == "" {
			http.Error(w, "run required", http.StatusBadRequest)
			return
		}
		f, err := os.Open(filepath.Join("data/runs", run, "aqm_eval.csv"))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		rd := csv.NewReader(f)
		rows, _ := rd.ReadAll()
		if len(rows) <= 1 {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		// no manual Connection header — hop-by-hop, illegal over HTTP/2 (broke SSE behind cloudflared/CF edge)
		for i, row := range rows[1:] {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			// row order per model.AQMEvalHeader
			payload, _ := json.Marshal(map[string]any{
				"run_id": row[0], "event_id": row[1], "profile": row[2], "qdisc": row[3], "cc": row[4],
				"repetition": row[5], "rtt_p50_ms": row[6], "rtt_p95_ms": row[7], "small_p95_ms": row[8],
				"bulk_goodput_mbps": row[10], "drops": row[11], "gate_status": row[16],
				"ts": time.Now().UnixMilli(), "running": true, "phase": "replay",
			})
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", i+1, payload)
			fl.Flush()
			time.Sleep(400 * time.Millisecond)
		}
	})

	mux.HandleFunc("POST /api/audit/start", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Site     string `json:"site"`
			LinkType string `json:"link_type"`
			Provider string `json:"provider"`
			Duration int    `json:"duration"`
			Target   string `json:"target"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if body.Duration <= 0 {
			body.Duration = 30
		}
		// prevention — clamp to the §6 window; a runaway audit is a cost
		if body.Duration < 10 {
			body.Duration = 10
		}
		if body.Duration > 600 {
			body.Duration = 600
		}
		if len(body.Site) > 80 {
			body.Site = body.Site[:80]
		}
		if len(body.Target) > 64 {
			http.Error(w, "cible trop longue", http.StatusBadRequest)
			return
		}
		if body.Target == "" {
			body.Target = "8.8.8.8"
		}
		auditMu.Lock()
		if auditRunning {
			auditMu.Unlock()
			http.Error(w, "audit already running", http.StatusConflict)
			return
		}
		auditRunning = true
		auditMu.Unlock()
		go func() {
			defer func() { auditMu.Lock(); auditRunning = false; auditMu.Unlock() }()
			p := audit.Params{AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()), Site: body.Site, LinkType: body.LinkType, Provider: body.Provider, Duration: body.Duration, Target: body.Target}
			// detached ctx: the audit must outlive this HTTP request
			res, err := audit.Run(context.Background(), p, audit.Deps{})
			if err != nil {
				return
			}
			auditMu.Lock()
			lastAudit = res
			auditMu.Unlock()
			recordEvent("audit", fmt.Sprintf("terminé — %s (%s), p95 %g ms", p.Site, p.LinkType, res.RTTIdleP95))
			_ = audit.AppendLinkAudit("data", res)
		}()
		writeJSON(w, map[string]any{"started": true})
	})
	mux.HandleFunc("GET /api/audit/status", func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		defer auditMu.Unlock()
		writeJSON(w, map[string]any{"running": auditRunning, "last": lastAudit})
	})
	mux.HandleFunc("GET /api/audit/list", func(w http.ResponseWriter, _ *http.Request) {
		// read link_audit.csv
		f, err := os.Open("data/link_audit.csv")
		if err != nil {
			writeJSON(w, map[string]any{"audits": []any{}})
			return
		}
		defer f.Close()
		rd := csv.NewReader(f)
		rows, _ := rd.ReadAll()
		if len(rows) <= 1 {
			writeJSON(w, map[string]any{"audits": []any{}})
			return
		}
		var out []map[string]string
		hdr := rows[0]
		for _, row := range rows[1:] {
			m := map[string]string{}
			for i, h := range hdr {
				if i < len(row) {
					m[h] = row[i]
				}
			}
			out = append(out, m)
		}
		writeJSON(w, map[string]any{"audits": out})
	})
	mux.HandleFunc("POST /api/profile/import", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		ct := r.Header.Get("Content-Type")
		var p model.Profile
		// CSV row (LIEN III.IV): id,capacity_mbps,delay_ms,jitter_ms,loss_pct — optional header line skipped
		if strings.Contains(ct, "text/csv") || (len(body) > 0 && body[0] != '{') {
			rows, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
			if err != nil || len(rows) == 0 {
				http.Error(w, "bad csv", http.StatusBadRequest)
				return
			}
			row := rows[0]
			if len(row) > 0 && (row[0] == "id" || row[0] == "ID") && len(rows) > 1 {
				row = rows[1] // skip header line
			}
			if len(row) < 2 {
				http.Error(w, "bad csv row", http.StatusBadRequest)
				return
			}
			num := func(s string) float64 { v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64); return v }
			p = model.Profile{ID: strings.TrimSpace(row[0])}
			if len(row) > 1 {
				p.CapacityMbps = num(row[1])
			}
			if len(row) > 2 {
				p.DelayMs = num(row[2])
			}
			if len(row) > 3 {
				p.JitterMs = num(row[3])
			}
			if len(row) > 4 {
				p.LossPct = num(row[4])
			}
		} else if err := json.Unmarshal(body, &p); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if p.ID == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		if err := profile.Import(p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	})
	mux.HandleFunc("GET /api/profile/list", func(w http.ResponseWriter, _ *http.Request) {
		profile.Load()
		writeJSON(w, model.Profiles)
	})
	mux.HandleFunc("GET /api/diagnostics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"hub": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("GET /api/hardware/translate", HandleTranslate)
	// JSON 404 for unknown /api/* — must be before "/" SPA catch-all
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found", "path": r.URL.Path})
	})

	mux.Handle("GET /api/stream", http.HandlerFunc(hub.SSE))
	hub.Serve(func() any { return snap() })

	spa := http.FileServer(http.FS(frontend.FS))
	// Compare view (Q14) — raw frozen rows of one run, path-traversal safe.
	mux.HandleFunc("GET /api/run/rows", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("run")
		if id == "" || strings.ContainsAny(id, `/\.`) {
			http.Error(w, "id de run invalide", http.StatusBadRequest)
			return
		}
		path := filepath.Join("data", "runs", id, "aqm_eval.csv")
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "run introuvable: "+id, http.StatusNotFound)
			return
		}
		defer f.Close()
		rows, err := csv.NewReader(f).ReadAll()
		if err != nil || len(rows) < 2 {
			http.Error(w, "run vide: "+id, http.StatusNotFound)
			return
		}
		hdr := rows[0]
		out := make([]map[string]any, 0, len(rows)-1)
		for _, row := range rows[1:] {
			m := map[string]any{}
			for i, col := range hdr {
				if i < len(row) {
					m[col] = row[i]
				}
			}
			out = append(out, m)
		}
		writeJSON(w, map[string]any{"run": id, "rows": out})
	})
	mux.HandleFunc("GET /api/events", func(w http.ResponseWriter, _ *http.Request) {
		eventsMu.Lock()
		defer eventsMu.Unlock()
		out := make([]map[string]any, len(eventsRing))
		copy(out, eventsRing)
		writeJSON(w, map[string]any{"events": out})
	})
	mux.Handle("/", spa)
	return Handler{Handler: mux, hubClose: func() { hub.Close() }}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
