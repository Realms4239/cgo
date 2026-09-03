package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/results"
	frontend "github.com/Realms4239/cgo/web/frontend"
)

// Schema — source unique de vérité pour chaque contrôle paramétré :
// le serveur valide depuis ce registre, GET /api/schema le sert, le client
// rend les champs depuis ce registre (placeholders = valeurs par défaut,
// min/max d'ici). Les bornes ne peuvent plus diverger entre l'UI et le backend.
type Param struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Min     float64  `json:"min,omitempty"`
	Max     float64  `json:"max,omitempty"`
	Default any      `json:"default"`
	Unit    string   `json:"unit,omitempty"`
	Step    float64  `json:"step,omitempty"`
	Enum    []string `json:"enum,omitempty"`
	Desc    string   `json:"desc"`
}

var schemaParams = []Param{
	{Key: "reps", Label: "Répétitions", Min: 1, Max: 5, Default: 3, Unit: "×", Desc: "répétitions de chaque cellule profile×qdisc×CC"},
	{Key: "deadline_ms", Label: "Deadline small p95", Min: 200, Max: 5000, Default: 1000, Unit: "ms", Step: 100, Desc: "objectif de latence qui voyage avec la campagne"},
	{Key: "target", Label: "Cible de mesure", Max: 64, Default: "1.1.1.1", Desc: "hôte sondé par la campagne et l'audit"},
	{Key: "audit_duration_s", Label: "Durée d'audit", Min: 10, Max: 600, Default: 30, Unit: "s", Step: 10, Desc: "durée de l'audit client-side du lien"},
	{Key: "warn_ms", Label: "Seuil dégradé", Min: 10, Max: 200, Default: 40, Unit: "ms", Step: 5, Desc: "latence au-delà de laquelle une carte passe en jaune"},
	{Key: "crit_ms", Label: "Seuil critique", Min: 50, Max: 500, Default: 100, Unit: "ms", Step: 5, Desc: "latence au-delà de laquelle une carte passe en rouge"},
	{Key: "ring_max", Label: "Fenêtre live", Min: 600, Max: 3600, Default: 1800, Unit: "pts", Step: 100, Desc: "profondeur d'historique du tableau live"},
	{Key: "watch_ms", Label: "Cadence surveillance", Min: 200, Max: 2000, Default: 500, Unit: "ms", Step: 100, Desc: "intervalle des sondes de surveillance"},
	{Key: "shape_capacity_mbps", Label: "Capacité du bord", Min: 1, Max: 1000, Default: 20, Unit: "Mbit/s", Desc: "capacité appliquée par le levier de façonnage"},
	{Key: "link_delay_ms", Label: "Délai du lien", Min: 0, Max: 600, Default: 20, Unit: "ms", Desc: "délai proposé par défaut aux conditions du lien"},
	{Key: "link_jitter_ms", Label: "Gigue", Min: 0, Max: 100, Default: 2, Unit: "ms", Desc: "gigue proposée par défaut"},
	{Key: "link_loss_pct", Label: "Perte", Min: 0, Max: 10, Default: 0, Unit: "%", Step: 0.1, Desc: "perte proposée par défaut"},
	{Key: "shape_qdisc", Label: "File d'attente", Default: "cake", Enum: []string{"cake", "fq_codel", "pfifo_fast", "none"}, Desc: "discipline appliquée au bord"},
	{Key: "burst_cc", Label: "Contrôle de congestion du burst", Default: "bbr", Enum: []string{"cubic", "bbr"}, Desc: "CC du burst de test à travers le bord façonné"},
	{Key: "burst_seconds", Label: "Durée du burst", Min: 2, Max: 10, Default: 4, Unit: "s", Desc: "durée d'un burst de test"},
	{Key: "price_tier", Label: "Palier tarifaire", Default: "yas-month-4.5gb", Enum: []string{"yas-day-1gb", "yas-month-4.5gb", "airtel-month-4.5gb", "orange-month-5gb", "yas-month-100gb", "yas-ftth-100gb"}, Desc: "forfait de référence du coût du gaspillage (tarifs réels 2026 — docs/data-prices.md)"},
}

// paramBounds — la validation côté serveur lit ce même registre.
func paramBounds(key string) (min, max float64, ok bool) {
	for _, p := range schemaParams {
		if p.Key == key {
			return p.Min, p.Max, true
		}
	}
	return 0, 0, false
}

// SchemaParams expose le registre des paramètres (défauts côté hôte).
func SchemaParams() []Param { return schemaParams }

// ShapeReq — levier complet: file d'attente + conditions du lien.
type ShapeReq struct {
	Qdisc    string  `json:"qdisc"`
	CapMbps  float64 `json:"capacity_mbps"`
	DelayMs  float64 `json:"delay_ms"`
	JitterMs float64 `json:"jitter_ms"`
	LossPct  float64 `json:"loss_pct"`
}

// RunOpts — ce que l'opérateur décide réellement : profils, axes qdisc/CC
// (vides = matrice pleine), répétitions, deadline small p95 et cible.
type RunOpts struct {
	Profiles   []string `json:"profiles"`
	Qdiscs     []string `json:"qdiscs"`
	CCs        []string `json:"ccs"`
	Reps       int      `json:"reps"`
	DeadlineMs int      `json:"deadline_ms"`
	Target     string   `json:"target"`
}

// Deps relie le serveur au noyau campagne.
type Deps struct {
	GetSnap func() any
	StartFn func(o RunOpts) error
	StopFn  func()
	// SkipFn coupe la cellule en cours sans arrêter la matrice — l'event
	// n'est pas gelé, il reste re-exécutable à la reprise.
	SkipFn func()
	// ShapeFn applique la discipline de file + les conditions du lien au bord
	// (le façonnage du bord est le levier que la DSI contrôle).
	// "none" efface.
	ShapeFn func(r ShapeReq) error
	// RunningFn signale une campagne active — le levier de façonnage et une
	// matrice en cours se disputent le même shaper, donc le façonnage manuel
	// est refusé en cours de run.
	RunningFn func() bool
	// WatchFn active/désactive la boucle de surveillance non intrusive
	// (ping+small, sans bulk) : le mur reste vivant hors campagne. La
	// surveillance n'entre PAS en conflit avec le façonnage — cette
	// combinaison est le produit.
	WatchFn func(on bool) error
	// Mode — les hôtes en observation refusent les endpoints de contrôle :
	// audit et consultation fonctionnent, campagne/façonnage/surveillance
	// répondent 501. "" = mode complet.
	Mode string
	// Version — estampillée au build (-X main.version), exposée par /api/health.
	Version string
	// DoctorFn — rapport de capacités pour GET /api/doctor ; nil ⇒ mode seul.
	DoctorFn func() any
	// BurstFn — une sonde bulk avec la CC choisie à travers le bord façonné :
	// façonnage seul, sans campagne. Bloquant (2–10 s).
	BurstFn func(cc string, seconds int) error
}

// observeBlocked — les hôtes Windows observent : l'audit fonctionne, les
// endpoints de contrôle répondent 501 avec le renvoi vers le banc. True = répondu.
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

// RecordEvent exporte le journal opérateur vers les hôtes qui câblent les
// callbacks campagne (les cellules quarantaine arrivent ici depuis la boucle matrice).
func RecordEvent(kind, msg string) { recordEvent(kind, msg) }

// Handler porte le cycle de vie du hub afin que les hôtes puissent arrêter
// le ticker 10 Hz à l'arrêt (Server.Close seul ne le fait jamais).
type Handler struct {
	http.Handler
	hubClose func()
}

// CloseHub arrête le ticker de diffusion du hub. Idempotent.
func (h Handler) CloseHub() {
	if h.hubClose != nil {
		h.hubClose()
	}
}

var auditMu sync.Mutex
var lastAudit *audit.Result
var auditRunning bool

// Journal — mémoire opérateur: anneau des 50 derniers événements.
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

// shapeState — dernier façonnage du bord appliqué (observabilité du contrôle).
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
	capMn, capMx, _ := paramBounds("shape_capacity_mbps")
	if cap < capMn || cap > capMx {
		return http.StatusBadRequest, map[string]any{"error": "capacity_mbps must be 1–1000"}
	}
	dMn, dMx, _ := paramBounds("link_delay_ms")
	jMn, jMx, _ := paramBounds("link_jitter_ms")
	lMn, lMx, _ := paramBounds("link_loss_pct")
	if req.DelayMs < dMn || req.DelayMs > dMx || req.JitterMs < jMn || req.JitterMs > jMx || req.LossPct < lMn || req.LossPct > lMx {
		return http.StatusBadRequest, map[string]any{"error": "conditions du lien: délai 0–600 ms, gigue 0–100 ms, perte 0–10 %"}
	}
	if err := fn(req); err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	shapeMu.Lock()
	shapeQdisc, shapeCap, shapeSince = q, cap, time.Now()
	since := shapeSince.Format(time.RFC3339)
	shapeMu.Unlock()
	cond := ""
	if req.DelayMs > 0 || req.JitterMs > 0 || req.LossPct > 0 {
		cond = fmt.Sprintf(", conditions %gms/%gms/%g%%", req.DelayMs, req.JitterMs, req.LossPct)
	}
	recordEvent("façonnage", fmt.Sprintf("%s @ %g Mbit/s%s", q, cap, cond))
	return http.StatusOK, map[string]any{"qdisc": q, "capacity_mbps": cap, "since": since}
}

const maxSubs = 64

// New construit le handler complet avec repli SPA. Le Handler retourné expose
// CloseHub pour un arrêt propre (stoppe le ticker de diffusion 10 Hz).
func New(d Deps) Handler {
	hub := NewHub()
	mux := http.NewServeMux()

	// fournisseur nil ⇒ instantané idle — vide honnête, jamais de panic sur appel nil
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
	// le contrat de paramètres — le client rend les champs depuis ce registre
	mux.HandleFunc("GET /api/schema", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"params": schemaParams})
	})
	// paliers tarifaires réels — le client rend le choix de forfait honnêtement
	mux.HandleFunc("GET /api/cost/tiers", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"tiers": metrics.Tiers, "default": metrics.DefaultTier.Name})
	})
	// test burst — une sonde bulk avec la CC choisie à travers le bord
	// actuellement façonné, pendant que la surveillance veille. Refusé en
	// cours de campagne : la matrice possède alors le shaper.
	mux.HandleFunc("POST /api/burst", func(w http.ResponseWriter, r *http.Request) {
		if d.observeBlocked(w) {
			return
		}
		var body struct {
			CC      string `json:"cc"`
			Seconds int    `json:"seconds"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		validCC := false
		for _, p := range schemaParams {
			if p.Key == "burst_cc" {
				for _, e := range p.Enum {
					if e == body.CC {
						validCC = true
					}
				}
			}
		}
		if !validCC {
			writeErr(w, r, "cc must be cubic | bbr", http.StatusBadRequest)
			return
		}
		if body.Seconds == 0 {
			body.Seconds = 4
		}
		if mn, mx, ok := paramBounds("burst_seconds"); ok && (body.Seconds < int(mn) || body.Seconds > int(mx)) {
			writeErr(w, r, "seconds must be 2–10", http.StatusBadRequest)
			return
		}
		if d.RunningFn != nil && d.RunningFn() {
			writeErr(w, r, "campagne active — le burst se lance hors campagne", http.StatusConflict)
			return
		}
		if d.BurstFn == nil {
			writeErr(w, r, "burst engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		if err := d.BurstFn(body.CC, body.Seconds); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		recordEvent("burst", fmt.Sprintf("burst %s %ds — traverse le bord façonné", body.CC, body.Seconds))
		writeJSON(w, map[string]any{"ok": true, "cc": body.CC, "seconds": body.Seconds})
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, snap())
	})
	// Profils dynamiques (profil importé inclus):
	mux.HandleFunc("GET /api/profiles", func(w http.ResponseWriter, _ *http.Request) {
		model.ProfilesMu.RLock()
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
			out = append(out, prof{ID: id, Capacity: p.CapacityMbps, DelayMs: p.DelayMs, JitterMs: p.JitterMs, LossPct: p.LossPct, FromImport: !model.BuiltinProfileIDs[id]})
		}
		model.ProfilesMu.RUnlock()
		writeJSON(w, map[string]any{"profiles": out})
	})
	// Façonnage du bord — appliquer / observer:
	mux.HandleFunc("GET /api/shape", func(w http.ResponseWriter, _ *http.Request) {
		shapeMu.Lock()
		defer shapeMu.Unlock()
		if shapeQdisc == "" {
			writeJSON(w, map[string]any{"applied": false})
			return
		}
		writeJSON(w, map[string]any{"applied": true, "qdisc": shapeQdisc, "capacity_mbps": shapeCap, "since": shapeSince.Format(time.RFC3339)})
	})
	// surveillance continue — non intrusive, composable avec le levier
	mux.HandleFunc("POST /api/watch", func(w http.ResponseWriter, r *http.Request) {
		if d.observeBlocked(w) {
			return
		}
		var body struct {
			On bool `json:"on"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if d.WatchFn == nil {
			writeErr(w, r, "watch engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		if body.On && d.RunningFn != nil && d.RunningFn() {
			writeErr(w, r, "campagne active — la surveillance se lance hors campagne", http.StatusConflict)
			return
		}
		if err := d.WatchFn(body.On); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
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
			writeErr(w, r, "bad json", http.StatusBadRequest)
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
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		// prévention — le client est une indication, jamais le contrat (§6) ; l'entrée
		// est validée même sur les hôtes sans moteur câblé — bornes issues du
		// même registre que sert GET /api/schema
		if mn, mx, ok := paramBounds("reps"); ok && (opts.Reps < int(mn) || opts.Reps > int(mx)) {
			writeErr(w, r, "reps must be 1–5", http.StatusBadRequest)
			return
		}
		// bornes — la deadline voyage avec le run, dans la même fenêtre
		// que le tiroir Réglages impose côté client
		if opts.DeadlineMs != 0 {
			if mn, mx, ok := paramBounds("deadline_ms"); ok && (float64(opts.DeadlineMs) < mn || float64(opts.DeadlineMs) > mx) {
				writeErr(w, r, "deadline_ms must be 200–5000", http.StatusBadRequest)
				return
			}
		}
		if len(opts.Target) > 64 {
			writeErr(w, r, "target must be ≤ 64 characters", http.StatusBadRequest)
			return
		}
		if len(opts.Profiles) == 0 {
			writeErr(w, r, "aucun profil sélectionné", http.StatusBadRequest)
			return
		}
		var unknown []string
		model.ProfilesMu.RLock()
		for _, id := range opts.Profiles {
			if _, ok := model.Profiles[id]; !ok {
				unknown = append(unknown, id)
			}
		}
		model.ProfilesMu.RUnlock()
		if len(unknown) > 0 {
			writeErr(w, r, "profils inconnus: "+strings.Join(unknown, ", "), http.StatusBadRequest)
			return
		}
		// axes qdisc/CC : validés ici même sans moteur câblé (prévention §6) ;
		// vides = matrice pleine (comportement historique)
		if len(opts.Qdiscs) > 0 {
			validQ := map[string]bool{}
			for _, p := range schemaParams {
				if p.Key == "shape_qdisc" {
					for _, e := range p.Enum {
						if e != "none" {
							validQ[e] = true
						}
					}
				}
			}
			for _, q := range opts.Qdiscs {
				if !validQ[q] {
					writeErr(w, r, "qdisc inconnu: "+q, http.StatusBadRequest)
					return
				}
			}
		}
		if len(opts.CCs) > 0 {
			validCC := map[string]bool{}
			for _, p := range schemaParams {
				if p.Key == "burst_cc" {
					for _, e := range p.Enum {
						validCC[e] = true
					}
				}
			}
			for _, c := range opts.CCs {
				if !validCC[c] {
					writeErr(w, r, "cc inconnu: "+c, http.StatusBadRequest)
					return
				}
			}
		}
		if len(opts.Qdiscs) == 0 && len(opts.CCs) > 0 || len(opts.Qdiscs) > 0 && len(opts.CCs) == 0 {
			writeErr(w, r, "qdiscs et ccs se filtrent ensemble", http.StatusBadRequest)
			return
		}
		if d.StartFn == nil {
			writeErr(w, r, "run engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		if err := d.StartFn(opts); err != nil {
			writeErr(w, r, err.Error(), http.StatusConflict)
			return
		}
		recordEvent("campagne", fmt.Sprintf("démarrée — %s ×%d, deadline %d ms, cible %s", strings.Join(opts.Profiles, "/"), opts.Reps, opts.DeadlineMs, opts.Target))
		writeJSON(w, map[string]any{"started": true})
	})
	mux.HandleFunc("POST /api/run/stop", func(w http.ResponseWriter, r *http.Request) {
		if d.StopFn == nil {
			writeErr(w, r, "run engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		d.StopFn()
		recordEvent("campagne", "arrêtée")
		writeJSON(w, map[string]any{"stopped": true})
	})
	mux.HandleFunc("POST /api/run/skip", func(w http.ResponseWriter, r *http.Request) {
		if d.SkipFn == nil {
			writeErr(w, r, "run engine not wired on this host", http.StatusServiceUnavailable)
			return
		}
		d.SkipFn()
		recordEvent("campagne", "cellule courante skippée — reprise possible")
		writeJSON(w, map[string]any{"skipped": true})
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
	// delta — même cellule (profile|qdisc|cc) du run précédent : la dérive
	// temporelle sans re-campagne. Ex: GET /api/results/delta?cell=P2|cake|bbr
	mux.HandleFunc("GET /api/results/delta", func(w http.ResponseWriter, r *http.Request) {
		cell := r.URL.Query().Get("cell")
		if cell == "" || strings.Count(cell, "|") != 2 {
			writeErr(w, r, "cell required: profile|qdisc|cc", http.StatusBadRequest)
			return
		}
		runs, _ := filepath.Glob("data/runs/*")
		if len(runs) < 2 {
			writeJSON(w, map[string]any{"available": false})
			return
		}
		sort.Strings(runs)
		prev, cur := runs[len(runs)-2], runs[len(runs)-1]
		type cellVals struct {
			SmallP95 float64
			RTTp95   float64
			Goodput  float64
		}
		readCell := func(runDir string) (cellVals, bool) {
			// Lecture par NOM de colonne : small_p95_ms, pas le voisin qdi_ms
			// (le delta calculait la dérive du QDI en l'étiquetant small_p95).
			header, rows, err := results.ReadAQM(filepath.Join(runDir, "aqm_eval.csv"))
			if err != nil || len(rows) == 0 {
				return cellVals{}, false
			}
			iProf, iQdisc, iCC := results.ColIndex(header, "profile"), results.ColIndex(header, "qdisc"), results.ColIndex(header, "cc")
			iSmall := results.ColIndex(header, "small_p95_ms")
			iRTT := results.ColIndex(header, "rtt_p95_ms")
			iGood := results.ColIndex(header, "bulk_goodput_mbps")
			iGate := results.ColIndex(header, "gate_status")
			get := func(row []string, i int) string {
				if i < 0 || i >= len(row) {
					return ""
				}
				return row[i]
			}
			var smalls, rtts, gds []float64
			for _, row := range rows {
				if get(row, iProf)+"|"+get(row, iQdisc)+"|"+get(row, iCC) != cell || get(row, iGate) != "valid" {
					continue
				}
				if v, err := strconv.ParseFloat(get(row, iSmall), 64); err == nil {
					smalls = append(smalls, v)
				}
				if v, err := strconv.ParseFloat(get(row, iRTT), 64); err == nil {
					rtts = append(rtts, v)
				}
				if v, err := strconv.ParseFloat(get(row, iGood), 64); err == nil {
					gds = append(gds, v)
				}
			}
			if len(smalls) == 0 {
				return cellVals{}, false
			}
			sm := func(xs []float64) float64 {
				if len(xs) == 0 {
					return 0
				}
				return xs[len(xs)/2]
			}
			return cellVals{SmallP95: sm(smalls), RTTp95: sm(rtts), Goodput: sm(gds)}, true
		}
		pv, ok1 := readCell(prev)
		cv, ok2 := readCell(cur)
		if !ok1 || !ok2 {
			writeJSON(w, map[string]any{"available": false, "reason": "cellule absente d'un run"})
			return
		}
		pct := func(a, b float64) *int {
			if a <= 0 {
				return nil
			}
			p := int(math.Round((b - a) / a * 100))
			return &p
		}
		writeJSON(w, map[string]any{
			"available": true, "previous_run": filepath.Base(prev), "current_run": filepath.Base(cur),
			"delta": map[string]any{"small_p95_pct": pct(pv.SmallP95, cv.SmallP95), "rtt_p95_pct": pct(pv.RTTp95, cv.RTTp95), "goodput_pct": pct(pv.Goodput, cv.Goodput)},
		})
	})
	mux.HandleFunc("GET /api/integrity", func(w http.ResponseWriter, _ *http.Request) {
		runs, _ := filepath.Glob("data/runs/*")
		sort.Strings(runs)
		valid, quarantined := 0, 0
		manifests := 0
		var runIDs []string
		type runBreak struct {
			Run         string `json:"run"`
			Rows        int    `json:"rows"`
			Valid       int    `json:"valid"`
			Quarantined int    `json:"quarantined"`
		}
		breakdown := []runBreak{}
		for _, p := range runs {
			if st, err := os.Stat(filepath.Join(p, "manifest.json")); err == nil && !st.IsDir() {
				manifests++
			}
			gs, _ := results.Scan("data/runs", filepath.Base(p))
			b := runBreak{Run: filepath.Base(p)}
			for _, g := range gs {
				valid += g.Count - g.Quarantined
				quarantined += g.Quarantined
				b.Rows += g.Count
				b.Valid += g.Count - g.Quarantined
				b.Quarantined += g.Quarantined
			}
			breakdown = append(breakdown, b)
			runIDs = append(runIDs, filepath.Base(p))
		}
		// récents d'abord (run-<unix> : l'ordre lexical inverse = chronologique inverse)
		for i, j := 0, len(breakdown)-1; i < j; i, j = i+1, j-1 {
			breakdown[i], breakdown[j] = breakdown[j], breakdown[i]
		}
		for i, j := 0, len(runIDs)-1; i < j; i, j = i+1, j-1 {
			runIDs[i], runIDs[j] = runIDs[j], runIDs[i]
		}
		updated := ""
		if files, _ := filepath.Glob(filepath.Join("data/runs", "*", "aqm_eval.csv")); len(files) > 0 {
			sort.Strings(files)
			if st, err := os.Stat(files[len(files)-1]); err == nil {
				updated = st.ModTime().UTC().Format(time.RFC3339)
			}
		}
		if len(runs) == 0 {
			writeJSON(w, map[string]any{"available": false, "reason": "intégrité disponible après gel (jalon M2)"})
			return
		}
		writeJSON(w, map[string]any{"available": true, "runs": len(runs), "manifests": manifests, "valid": valid, "quarantined": quarantined, "run_ids": runIDs, "breakdown": breakdown, "updated": updated, "sha256": figures.ProvenanceSHA("data/runs"), "hash8": figures.ProvenanceHash8("data/runs")})
	})
	mux.HandleFunc("GET /api/report/export", func(w http.ResponseWriter, r *http.Request) {
		fmtParam := r.URL.Query().Get("format")
		if fmtParam == "" {
			fmtParam = "md"
		}
		groups, _ := results.Scan("data/runs", "")
		if len(groups) == 0 {
			writeErr(w, r, "no data", http.StatusNotFound)
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
		w.Write([]byte("# Meteolink Report\n\n| profil | qdisc | cc | n | small p95 | rtt p95 | goodput | quarant. | best |\n|---|---|---|---|---|---|---|---|---|\n"))
		for _, g := range groups {
			mark := ""
			if g.Best {
				mark = "★"
			}
			w.Write([]byte(fmt.Sprintf("| %s | %s | %s | %d | %.1f | %.1f | %.1f | %d | %s |\n", g.Profile, g.Qdisc, g.CC, g.Count, g.Smallp95Median, g.RTTp95Median, g.GoodputMedian, g.Quarantined, mark)))
		}
		// triple provenance — même hash8 que le tiroir, les archives et les figures RDF
		if hash8 := figures.ProvenanceHash8("data/runs"); hash8 != "" {
			fmt.Fprintf(w, "\nprovenance : sha256:%s · source data/runs/*/aqm_eval.csv · généré %s\n", hash8, time.Now().UTC().Format(time.RFC3339))
		}
	})

	mux.HandleFunc("POST /api/figures/regen", func(w http.ResponseWriter, r *http.Request) {
		if err := figures.Generate("data/runs", "data/figures"); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
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
		// même garde anti-traversal que /api/run/rows — pas de sortie de data/runs
		if run == "" || strings.ContainsAny(run, `/\.`) {
			writeErr(w, r, "id de run invalide", http.StatusBadRequest)
			return
		}
		// Lecture par NOM de colonne : small_p95_ms depuis small_p95_ms (pas
		// qdi_ms), goodput depuis bulk_goodput_mbps (pas deadline_ok_pct).
		header, rows, err := results.ReadAQM(filepath.Join("data/runs", run, "aqm_eval.csv"))
		if err != nil {
			writeErr(w, r, "not found", http.StatusNotFound)
			return
		}
		if len(rows) == 0 {
			writeErr(w, r, "no data", http.StatusNotFound)
			return
		}
		fl, ok := w.(http.Flusher)
		if !ok {
			writeErr(w, r, "stream unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		// no manual Connection header — hop-by-hop, illegal over HTTP/2 (broke SSE behind cloudflared/CF edge)
		for i, row := range rows {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			payload, _ := json.Marshal(map[string]any{
				"run_id": results.Field(header, row, "run_id"), "event_id": results.Field(header, row, "event_id"),
				"profile": results.Field(header, row, "profile"), "qdisc": results.Field(header, row, "qdisc"), "cc": results.Field(header, row, "cc"),
				"repetition": results.Field(header, row, "repetition"),
				"rtt_p50_ms": results.Field(header, row, "rtt_p50_ms"), "rtt_p95_ms": results.Field(header, row, "rtt_p95_ms"),
				"small_p95_ms":      results.Field(header, row, "small_p95_ms"),
				"bulk_goodput_mbps": results.Field(header, row, "bulk_goodput_mbps"), "drops": results.Field(header, row, "drops"),
				"gate_status": results.Field(header, row, "gate_status"),
				"ts":          time.Now().UnixMilli(), "running": true, "phase": "replay",
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
			SmallURL string `json:"small_url"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if body.Duration <= 0 {
			body.Duration = 30
		}
		// prévention — bornes imposées; un audit débridé coûte cher
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
			writeErr(w, r, "cible trop longue", http.StatusBadRequest)
			return
		}
		if body.Target == "" {
			body.Target = "8.8.8.8"
		}
		auditMu.Lock()
		if auditRunning {
			auditMu.Unlock()
			writeErr(w, r, "audit already running", http.StatusConflict)
			return
		}
		auditRunning = true
		auditMu.Unlock()
		go func() {
			defer func() { auditMu.Lock(); auditRunning = false; auditMu.Unlock() }()
			p := audit.Params{AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()), Site: body.Site, LinkType: body.LinkType, Provider: body.Provider, Duration: body.Duration, Target: body.Target, SmallURL: body.SmallURL}
			// contexte détaché: l'audit survit à cette requête HTTP
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
	// boucle RQ1→RQ2 : l'audit réel devient profil rejouable sur le banc.
	// Dernière ligne de link_audit.csv → capacity = throughput, delay = rtt_idle_p95.
	mux.HandleFunc("POST /api/audit/toprofile", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("data/link_audit.csv")
		if err != nil {
			writeErr(w, r, "aucun audit — lancez d'abord cgo audit", http.StatusNotFound)
			return
		}
		defer f.Close()
		rd := csv.NewReader(f)
		rows, _ := rd.ReadAll()
		if len(rows) < 2 {
			writeErr(w, r, "audit vide", http.StatusNotFound)
			return
		}
		hdr := rows[0]
		last := rows[len(rows)-1]
		get := func(col string) string {
			for i, h := range hdr {
				if h == col && i < len(last) {
					return last[i]
				}
			}
			return ""
		}
		cap, _ := strconv.ParseFloat(get("throughput_mbps"), 64)
		delay, _ := strconv.ParseFloat(get("rtt_idle_p95_ms"), 64)
		id := "P-audit"
		if v := get("audit_id"); v != "" {
			id = "P-" + v
		}
		p := model.Profile{ID: id, CapacityMbps: cap, DelayMs: delay}
		if p.CapacityMbps <= 0 {
			p.CapacityMbps = 20 // repli documenté — banc par défaut si débit non mesuré
		}
		if p.DelayMs <= 0 {
			p.DelayMs = 100
		}
		if err := profile.Import(p); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		recordEvent("profil", fmt.Sprintf("%s importé depuis audit — %g Mbit/s, %g ms", id, p.CapacityMbps, p.DelayMs))
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	})
	mux.HandleFunc("POST /api/profile/import", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			writeErr(w, r, "bad body", http.StatusBadRequest)
			return
		}
		ct := r.Header.Get("Content-Type")
		var p model.Profile
		// Ligne CSV: id,capacity_mbps,delay_ms,jitter_ms,loss_pct — en-tête optionnel sauté
		if strings.Contains(ct, "text/csv") || (len(body) > 0 && body[0] != '{') {
			rows, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
			if err != nil || len(rows) == 0 {
				writeErr(w, r, "bad csv", http.StatusBadRequest)
				return
			}
			row := rows[0]
			if len(row) > 0 && (row[0] == "id" || row[0] == "ID") && len(rows) > 1 {
				row = rows[1] // saute la ligne d'en-tête
			}
			if len(row) < 2 {
				writeErr(w, r, "bad csv row", http.StatusBadRequest)
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
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if p.ID == "" {
			writeErr(w, r, "id required", http.StatusBadRequest)
			return
		}
		if err := profile.Import(p); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	})
	mux.HandleFunc("GET /api/profile/list", func(w http.ResponseWriter, _ *http.Request) {
		profile.Load()
		// copie sous verrou : l'encodage JSON vers un client lent ne doit pas
		// retenir le verrou pendant l'écriture réseau
		model.ProfilesMu.RLock()
		cp := make(map[string]model.Profile, len(model.Profiles))
		for k, v := range model.Profiles {
			cp[k] = v
		}
		model.ProfilesMu.RUnlock()
		writeJSON(w, cp)
	})
	mux.HandleFunc("GET /api/diagnostics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"hub": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("GET /api/hardware/translate", HandleTranslate)
	// 404 JSON pour /api/* inconnu — doit précéder le catch-all SPA "/"
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found", "path": r.URL.Path})
	})

	mux.Handle("GET /api/stream", http.HandlerFunc(hub.SSE))
	hub.Serve(func() any { return snap() })

	spa := http.FileServer(http.FS(frontend.FS))
	// Vue Compare — lignes gelées brutes d'un run, protégée contre le cheminement.
	mux.HandleFunc("GET /api/run/rows", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("run")
		if id == "" || strings.ContainsAny(id, `/\.`) {
			writeErr(w, r, "id de run invalide", http.StatusBadRequest)
			return
		}
		path := filepath.Join("data", "runs", id, "aqm_eval.csv")
		f, err := os.Open(path)
		if err != nil {
			writeErr(w, r, "run introuvable: "+id, http.StatusNotFound)
			return
		}
		defer f.Close()
		rows, err := csv.NewReader(f).ReadAll()
		if err != nil || len(rows) < 2 {
			writeErr(w, r, "run vide: "+id, http.StatusNotFound)
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
	// Quarantaine réelle — lignes quarantine.json gelées, tous runs si ?run= vide.
	mux.HandleFunc("GET /api/quarantine", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("run")
		if strings.ContainsAny(id, `/\.`) {
			writeErr(w, r, "id de run invalide", http.StatusBadRequest)
			return
		}
		type qrow struct {
			Run     string `json:"run"`
			EventID int    `json:"event_id"`
			Profile string `json:"profile"`
			Qdisc   string `json:"qdisc"`
			CC      string `json:"cc"`
			Status  string `json:"gate_status"`
		}
		out := []qrow{}
		collect := func(run string) {
			b, err := os.ReadFile(filepath.Join("data", "runs", run, "quarantine.json"))
			if err != nil {
				return // aucun gel de quarantaine pour ce run : pas une erreur
			}
			var rows []qrow
			if err := json.Unmarshal(b, &rows); err != nil {
				return
			}
			for _, q := range rows {
				q.Run = run
				out = append(out, q)
			}
		}
		if id != "" {
			if st, err := os.Stat(filepath.Join("data", "runs", id)); err != nil || !st.IsDir() {
				writeErr(w, r, "run introuvable: "+id, http.StatusNotFound)
				return
			}
			collect(id)
		} else {
			runs, _ := filepath.Glob(filepath.Join("data", "runs", "*"))
			sort.Strings(runs)
			for i := len(runs) - 1; i >= 0; i-- {
				collect(filepath.Base(runs[i]))
			}
		}
		writeJSON(w, map[string]any{"quarantines": out})
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

// writeErr — erreur HTTP négociée : JSON pour les clients fetch().json()
// du frontend, text/plain inchangé pour curl et les scripts bash.
func writeErr(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if strings.Contains(r.Header.Get("Accept"), "application/json") ||
		strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": msg})
		return
	}
	http.Error(w, msg, code)
}
