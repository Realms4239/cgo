package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	frontend "github.com/Realms4239/cgo/web/frontend"
	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/figures"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/results"
)

// Deps wires the server to the campagne core.
type Deps struct {
	GetSnap  func() any
	StartFn  func(profiles []string, reps int) error
	StopFn   func()
}

var auditMu sync.Mutex
var lastAudit *audit.Result
var auditRunning bool

// New builds the full handler with SPA fallback.
func New(d Deps) http.Handler {
	hub := NewHub()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, d.GetSnap())
	})
	mux.HandleFunc("POST /api/run/start", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Profiles []string `json:"profiles"`
			Reps     int      `json:"reps"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if err := d.StartFn(body.Profiles, body.Reps); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		writeJSON(w, map[string]any{"started": true})
	})
	mux.HandleFunc("POST /api/run/stop", func(w http.ResponseWriter, _ *http.Request) {
		d.StopFn()
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
		writeJSON(w, map[string]any{"available": true, "runs": len(runs), "manifests": manifests, "valid": valid, "quarantined": quarantined, "run_ids": runIDs})
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
		w.Header().Set("Connection", "keep-alive")
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
		var p model.Profile
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&p); err != nil {
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
		// also update in-memory
		model.Profiles[p.ID] = p
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	})
	mux.HandleFunc("GET /api/profile/list", func(w http.ResponseWriter, _ *http.Request) {
		profile.Load()
		writeJSON(w, model.Profiles)
	})

	mux.Handle("GET /api/stream", http.HandlerFunc(hub.SSE))
	hub.Serve(func() any { return d.GetSnap() })

	spa := http.FileServer(http.FS(frontend.FS))
	mux.Handle("/", spa)
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
