package api

import (
	"encoding/json"
	"net/http"

	frontend "github.com/Realms4239/cgo/web/frontend"
)

// Deps wires the server to the campagne core.
type Deps struct {
	GetSnap  func() any
	StartFn  func(profiles []string, reps int) error
	StopFn   func()
}

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
	// Honest M2 placeholders — never fabricated data.
	mux.HandleFunc("GET /api/results", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"available": false, "reason": "résultats disponibles après gel (jalon M2)"})
	})
	mux.HandleFunc("GET /api/integrity", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"available": false, "reason": "intégrité disponible après gel (jalon M2)"})
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
