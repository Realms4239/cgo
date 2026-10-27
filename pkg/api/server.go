package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/figures"
	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/remediate"
	"github.com/Realms4239/cgo/pkg/results"
	frontend "github.com/Realms4239/cgo/web/frontend"
)

// envDefault — petit utilitaire de configuration par environnement ;
// les défauts du banc restent surchargeables sans toucher au code.
func envDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// readAuditReport loads the archived JSON artifact. The id is validated with
// the same guard as writes — a malformed id never silently resolves another
// audit's file. Empty id reads the global last-report copy.
func readAuditReport(id string) (*audit.Report, error) {
	path := filepath.Join("data", "audit_report.json")
	if id != "" {
		clean, err := audit.ArtifactID(id)
		if err != nil {
			return nil, err
		}
		path = filepath.Join("data", "audits", clean, "audit_report.json")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rep audit.Report
	if err := json.Unmarshal(b, &rep); err != nil {
		return nil, err
	}
	if rep.AuditID == "" {
		return nil, fmt.Errorf("%s: audit_id absent", path)
	}
	return &rep, nil
}

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
// (vides = matrice pleine), répétitions, deadline small p95, cible et sens
// de la charge ("up" défaut, "down" download, "both" RRUL combiné).
// RunID : reprise d'un run interrompu (PC/VM éteints en pleine matrice) —
// vide = run frais. La reprise rejoue les cellules manquantes (moteur).
type RunOpts struct {
	Profiles   []string `json:"profiles"`
	Qdiscs     []string `json:"qdiscs"`
	CCs        []string `json:"ccs"`
	Reps       int      `json:"reps"`
	DeadlineMs int      `json:"deadline_ms"`
	Target     string   `json:"target"`
	Direction  string   `json:"direction"`
	RunID      string   `json:"run_id"`
	// Flows — charge multi-flux. 0 = défaut du banc (4, décision 6 — le
	// câblage cmd/cgo applique le défaut) ; 1 = one stream (historical
	// cells); 2–8 saturates a lossy path where a single CUBIC collapses
	// below the link rate. The value travels into the manifest config.
	Flows int `json:"flows"`
	// Bench — étiquette de provenance : "" (défaut CGO_BENCH → legacy),
	// "legacy" (reproduction corpus) ou "dualns" (banc durci production).
	Bench string `json:"bench"`
	// ChannelModel — canal : "" (clean côté campagne), "clean" ou "bursty"
	// (rafales Gilbert-Elliott, P2/P3 uniquement — 400 sinon).
	ChannelModel string `json:"channel_model"`
}

// Deps relie le serveur au noyau campagne.
type Deps struct {
	GetSnap func() any
	// NextSnap — fournisseur du hub 10 Hz. Nil ⇒ GetSnap. Le hub lit un
	// échantillon frais ; /api/state garde GetSnap (dernier échantillon).
	NextSnap func() any
	StartFn  func(o RunOpts) error
	StopFn   func()
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
	// SinksFn — puits d'audit détectés (dualns → legacy → external) ;
	// nil ⇒ détection réelle. Les tests injectent un stub pour rester hors
	// réseau : les défauts d'audit/start suivent le banc, pas un 10.200.0.1
	// en dur (M4).
	SinksFn func() audit.Sinks
	// AuditDepsFn — dépendances de l'audit (tests : ping/small/bulk
	// instantanés) ; nil ⇒ implémentations réelles. OnProgress est posé
	// par le handler, jamais par l'injection.
	AuditDepsFn func() audit.Deps
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
var lastAuditReport *audit.Report
var auditRunning bool
var auditCancel context.CancelFunc
var auditStartedAt time.Time
var auditTotalSec int
var auditProgress atomic.Value // audit.Progress : écrit par l'audit, lu sans verrou

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
// compareCache — le bootstrap/MW coûte ~ms par cellule : précalcul + cache,
// invalidé par le hash de provenance (tout nouveau gel change la clé).
var compareCache struct {
	sync.Mutex
	key string
	val map[string]any
}

// readValidMetric — valeurs valid-only d'une métrique pour une cellule
// profile|qdisc|cc + direction (défaut up historique). Ignore run-smoke.
func readValidMetric(cell, metric, direction string) ([]float64, int) {
	files, _ := filepath.Glob(filepath.Join("data", "runs", "*", "aqm_eval.csv"))
	var out []float64
	n := 0
	for _, f := range files {
		if strings.HasPrefix(filepath.Base(filepath.Dir(f)), "run-smoke") {
			continue
		}
		header, rows, err := results.ReadAQM(f)
		if err != nil {
			continue
		}
		iProf, iQdisc, iCC := results.ColIndex(header, "profile"), results.ColIndex(header, "qdisc"), results.ColIndex(header, "cc")
		iMet := results.ColIndex(header, metric)
		iGate := results.ColIndex(header, "gate_status")
		iDir := results.ColIndex(header, "direction")
		get := func(row []string, i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return row[i]
		}
		for _, row := range rows {
			if get(row, iProf)+"|"+get(row, iQdisc)+"|"+get(row, iCC) != cell {
				continue
			}
			d := get(row, iDir)
			if d == "" {
				d = "up"
			}
			if d != direction || get(row, iGate) != "valid" {
				continue
			}
			n++
			// small_p95_ms : 0 = sonde vide (convention book) ; deadline 0 %
			// est une vraie mesure et se garde.
			if v, err := strconv.ParseFloat(get(row, iMet), 64); err == nil {
				if metric != "small_p95_ms" || v > 0 {
					out = append(out, v)
				}
			}
		}
	}
	return out, n
}

func New(d Deps) Handler {
	hub := NewHub()
	mux := http.NewServeMux()

	// fournisseur nil ⇒ instantané idle — vide honnête, jamais de panic sur appel nil
	snap := d.GetSnap
	if snap == nil {
		snap = func() any { return map[string]any{"running": false} }
	}
	next := d.NextSnap
	if next == nil {
		next = snap
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
		writeJSON(w, map[string]any{"tiers": metrics.Tiers, "default": metrics.ActiveTierName()})
	})
	// catalogue entreprise — plans, tarifs B2B, pénalité de secours et
	// mapping primaire : TOUT vient des vars de pkg/metrics (source unique,
	// un tarif changé au paquet se retrouve ici) ; aucun nombre recopié.
	mux.HandleFunc("GET /api/cost/enterprise", func(w http.ResponseWriter, _ *http.Request) {
		plans := make([]map[string]any, 0, 2)
		for _, p := range []metrics.EnterprisePlan{metrics.PlanGOLD, metrics.PlanPLATINE} {
			m := map[string]any{"name": p.Name, "monthly_opex_mga": p.MonthlyOpExMGA}
			// capacité ou FUP — jamais les deux zéros : un plan CIR n'a pas
			// de FUP, un plan FUP n'a pas de CIR (pas de zéro qui lit « mesuré »)
			if p.CIRMbps > 0 {
				m["cir_mbps"] = p.CIRMbps
			}
			if p.FUPBytes > 0 {
				m["fup_bytes"] = p.FUPBytes
			}
			plans = append(plans, m)
		}
		tariffs := make([]map[string]any, 0, 2)
		for _, tr := range []metrics.B2BTariff{metrics.TariffWholesale, metrics.TariffCorporateAPN} {
			t := map[string]any{"name": tr.Name, "ar_per_gb": tr.ARPerGB}
			if tr.Quoted {
				t["quoted"] = true // devis opérateur — le rapport doit le dire (invariant 4)
			}
			tariffs = append(tariffs, t)
		}
		writeJSON(w, map[string]any{
			"plans":               plans,
			"tariffs":             tariffs,
			"emergency_ar_per_gb": metrics.RetailEmergencyARPerGB,
			"primary_mapping":     map[string]string{"fiber": metrics.PrimaryName("fiber"), "other": metrics.PrimaryName("other"), "4g": metrics.PrimaryName("4g"), "5g": metrics.PrimaryName("5g"), "vsat": metrics.PrimaryName("vsat")},
			"disclaimer":          metrics.EconomicsDisclaimer,
		})
	})
	// palier actif — le forfait où vit l'institution (les coûts affichés
	// suivent ; inconnu = 400, pas de défaut silencieux)
	mux.HandleFunc("POST /api/cost/tier", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tier string `json:"tier"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if err := metrics.SetDefaultTier(body.Tier); err != nil {
			writeErr(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		recordEvent("tarif", fmt.Sprintf("palier actif : %s", body.Tier))
		writeJSON(w, map[string]any{"ok": true, "tier": body.Tier})
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
		profile.Load() // fusion persistée : un import survit au redémarrage
		model.ProfilesMu.RLock()
		ids := make([]string, 0, len(model.Profiles))
		for id := range model.Profiles {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		type prof struct {
			ID         string  `json:"id"`
			Capacity   float64 `json:"capacity_mbps"`
			CapacityUp float64 `json:"capacity_up_mbps,omitempty"`
			DelayMs    float64 `json:"delay_ms"`
			JitterMs   float64 `json:"jitter_ms"`
			LossPct    float64 `json:"loss_pct"`
			LossP      float64 `json:"loss_burst_p,omitempty"`
			LossR      float64 `json:"loss_burst_r,omitempty"`
			LossH      float64 `json:"loss_burst_h,omitempty"`
			LossK      float64 `json:"loss_burst_k,omitempty"`
			FromImport bool    `json:"imported"`
		}
		out := make([]prof, 0, len(ids))
		for _, id := range ids {
			p := model.Profiles[id]
			out = append(out, prof{ID: id, Capacity: p.CapacityMbps, CapacityUp: p.CapacityUpMbps, DelayMs: p.DelayMs, JitterMs: p.JitterMs, LossPct: p.LossPct, LossP: p.LossBurstP, LossR: p.LossBurstR, LossH: p.LossBurstH, LossK: p.LossBurstK, FromImport: !model.BuiltinProfileIDs[id]})
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
		if opts.Flows != 0 && (opts.Flows < 1 || opts.Flows > 8) {
			writeErr(w, r, "flows must be 1–8", http.StatusBadRequest)
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
		// provenance déclarée — validée ici même sans moteur câblé (le refus
		// doit tenir sur les hôtes en observation) ; bursty est une physique
		// de canal radio : P2/P3 uniquement (décision 7), refusé avant toute
		// mutation. La campagne reste le dernier mot (même règle, même message).
		switch opts.Bench {
		case "", "legacy", "dualns":
		default:
			writeErr(w, r, "bench must be legacy | dualns", http.StatusBadRequest)
			return
		}
		switch opts.ChannelModel {
		case "", "clean", "bursty":
		default:
			writeErr(w, r, "channel_model must be clean | bursty", http.StatusBadRequest)
			return
		}
		if opts.ChannelModel == "bursty" {
			var hors []string
			for _, id := range opts.Profiles {
				if id != "P2" && id != "P3" {
					hors = append(hors, id)
				}
			}
			if len(hors) > 0 {
				writeErr(w, r, "channel_model bursty invalide pour "+strings.Join(hors, ", ")+" — rafales sur P2/P3 uniquement (choisir clean)", http.StatusBadRequest)
				return
			}
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
		switch opts.Direction {
		case "", "up", "down", "both":
		default:
			writeErr(w, r, "direction must be up|down|both", http.StatusBadRequest)
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
			writeJSON(w, map[string]any{"available": false, "groups": []any{}, "reason": "résultats disponibles après gel (jalon M2)"})
			return
		}
		writeJSON(w, map[string]any{"available": true, "groups": groups})
	})
	// delta — même cellule (profile|qdisc|cc) du run précédent : la dérive
	// temporelle sans re-campagne. Ex: GET /api/results/delta?cell=P2|cake|bbr
	mux.HandleFunc("GET /api/results/delta", func(w http.ResponseWriter, r *http.Request) {
		cell := r.URL.Query().Get("cell")
		// cellule "P|q|cc" (up, historique) ou "P|q|cc|down" (explicite) :
		// les sens ne se comparent jamais entre eux
		parts := strings.Split(cell, "|")
		direction := "up"
		if len(parts) == 4 {
			direction = parts[3]
			cell = strings.Join(parts[:3], "|")
		}
		if cell == "" || len(parts) < 3 || len(parts) > 4 || (direction != "up" && direction != "down") {
			writeErr(w, r, "cell required: profile|qdisc|cc[|up|down]", http.StatusBadRequest)
			return
		}
		cellKey := cell + "|" + direction
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
			iDir := results.ColIndex(header, "direction")
			get := func(row []string, i int) string {
				if i < 0 || i >= len(row) {
					return ""
				}
				return row[i]
			}
			dirOf := func(row []string) string {
				if d := get(row, iDir); d != "" {
					return d
				}
				return "up" // historique sans colonne
			}
			var smalls, rtts, gds []float64
			for _, row := range rows {
				if get(row, iProf)+"|"+get(row, iQdisc)+"|"+get(row, iCC) != cell || dirOf(row) != direction || get(row, iGate) != "valid" {
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
			"available": true, "cell": cellKey, "previous_run": filepath.Base(prev), "current_run": filepath.Base(cur),
			"delta": map[string]any{"small_p95_pct": pct(pv.SmallP95, cv.SmallP95), "rtt_p95_pct": pct(pv.RTTp95, cv.RTTp95), "goodput_pct": pct(pv.Goodput, cv.Goodput)},
		})
	})
	// compare — preuve pilote : Mann-Whitney bilatéral + Cliff sur deux
	// cellules valid-only. Ex: GET /api/results/compare?a=P2|pfifo_fast|bbr&b=P2|cake|bbr&metric=small_p95_ms
	// metric ∈ {small_p95_ms, deadline_ok_pct}. deadline agrégée mélange des
	// échéances opérateur → flag deadline_mixed_D (indicatif seul).
	mux.HandleFunc("GET /api/results/compare", func(w http.ResponseWriter, r *http.Request) {
		a, b, metric := r.URL.Query().Get("a"), r.URL.Query().Get("b"), r.URL.Query().Get("metric")
		direction := r.URL.Query().Get("direction")
		if direction == "" {
			direction = "up"
		}
		if a == "" || b == "" {
			writeErr(w, r, "paramètres a et b requis (profile|qdisc|cc)", http.StatusBadRequest)
			return
		}
		if metric != "small_p95_ms" && metric != "deadline_ok_pct" {
			writeErr(w, r, "metric ∈ {small_p95_ms, deadline_ok_pct}", http.StatusBadRequest)
			return
		}
		key := figures.ProvenanceHash8("data/runs") + "|" + a + "|" + b + "|" + metric + "|" + direction
		compareCache.Lock()
		if compareCache.key == key {
			v := compareCache.val
			compareCache.Unlock()
			writeJSON(w, v)
			return
		}
		compareCache.Unlock()
		xa, na := readValidMetric(a, metric, direction)
		xb, nb := readValidMetric(b, metric, direction)
		if len(xa) == 0 || len(xb) == 0 {
			writeJSON(w, map[string]any{"available": false, "reason": "cellule vide en valid-only"})
			return
		}
		_, p := metrics.MannWhitneyTwoSided(xa, xb)
		d := metrics.CliffDelta(xa, xb)
		out := map[string]any{
			"available": true, "a": a, "b": b, "metric": metric, "direction": direction,
			"n_a": na, "n_b": nb, "mw_p_two_sided": p,
			"cliff_delta": d, "cliff_interp": metrics.CliffInterp(d),
			"deadline_mixed_D": metric == "deadline_ok_pct",
		}
		compareCache.Lock()
		compareCache.key, compareCache.val = key, out
		compareCache.Unlock()
		writeJSON(w, out)
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
		// script tc déployable — la recette de la meilleure cellule par profil,
		// rejouable en l'état sur l'équipement Linux cible (NetDevOps : le
		// verdict devient un artefact versionnable, pas une ligne isolée)
		if fmtParam == "sh" {
			groups, _ := results.Scan("data/runs", "")
			if len(groups) == 0 {
				writeErr(w, r, "no data", http.StatusNotFound)
				return
			}
			profByID := map[string]string{}
			model.ProfilesMu.RLock()
			ids := make([]string, 0, len(model.Profiles))
			for id := range model.Profiles {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				profByID[id] = id
			}
			model.ProfilesMu.RUnlock()
			var b strings.Builder
			b.WriteString("#!/bin/sh\n# Meteolink — recette AQM dérivée des mesures gelées\n")
			b.WriteString("# provenance : sha256:" + figures.ProvenanceSHA("data/runs") + "\n")
			b.WriteString("# usage : sh aqm-recipe.sh [interface] (lancer en root)\n")
			b.WriteString("IF=${1:-eth0}\n\n")
			for _, id := range ids {
				var best *results.Group
				for i := range groups {
					g := &groups[i]
					if g.Profile == id && g.Best {
						best = g
						break
					}
				}
				if best == nil {
					continue // profil jamais campagné : pas de recette fabriquée
				}
				model.ProfilesMu.RLock()
				p := model.Profiles[id]
				model.ProfilesMu.RUnlock()
				b.WriteString("# ---- " + id + " : best mesuré " + best.Qdisc + "/" + best.CC + " (" + fmt.Sprintf("small p95 %.1f ms", best.Smallp95Median) + ")\n")
				b.WriteString("tc qdisc del dev $IF root 2>/dev/null\n")
				b.WriteString(fmt.Sprintf("tc qdisc replace dev $IF root handle 1: netem delay %gms %gms", p.DelayMs, p.JitterMs))
				if p.LossPct > 0 {
					b.WriteString(fmt.Sprintf(" loss %g%%", p.LossPct))
				}
				b.WriteString("\n")
				// cake porte sa bande passante ; tbf pour pfifo/fq_codel (même
				// empilement que le banc, qdisc.ApplyShaper). Bande passante =
				// goodput mesuré de la cellule ×0,9 — même règle que la
				// suggestion CLI de translate (le lien réel, pas le profil).
				shaperMbps := float64(int(best.GoodputMedian * 0.9))
				if shaperMbps < 1 {
					shaperMbps = 1
				}
				switch best.Qdisc {
				case "cake":
					b.WriteString(fmt.Sprintf("tc qdisc replace dev $IF parent 1: handle 10: cake bandwidth %gmbit rtt %gms\n", shaperMbps, p.DelayMs))
				default:
					b.WriteString(fmt.Sprintf("tc qdisc replace dev $IF parent 1: handle 10: tbf rate %gmbit burst 256kbit latency 400ms\n", shaperMbps))
					b.WriteString("tc qdisc replace dev $IF parent 10:1 handle 20: " + best.Qdisc + "\n")
				}
			}
			w.Header().Set("Content-Type", "text/x-shellscript")
			w.Header().Set("Content-Disposition", "attachment; filename=aqm-recipe.sh")
			w.Write([]byte(b.String()))
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
			Site           string `json:"site"`
			LinkType       string `json:"link_type"`
			Provider       string `json:"provider"`
			Duration       int    `json:"duration"`
			Target         string `json:"target"`
			Direction      string `json:"direction"`
			Flows          int    `json:"flows"`
			SmallURL       string `json:"small_url"`
			BulkAddr       string `json:"bulk_addr"`
			Iface          string `json:"iface"`
			RouterOSIface  string `json:"routeros_iface"`
			Qdisc          string `json:"qdisc"`
			Asymmetric     *bool  `json:"asymmetric"`
			CPUConstrained bool   `json:"cpu_constrained"`
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
			body.Site = truncateUTF8(body.Site, 80)
		}
		// une cible en blanc ferait échouer le ping sans dire pourquoi :
		// elle retombe sur le banc détecté, comme une cible absente.
		body.Target = strings.TrimSpace(body.Target)
		if len(body.Target) > 64 {
			writeErr(w, r, "cible trop longue", http.StatusBadRequest)
			return
		}
		// puits décidés par le banc détecté (M4) — cible, petit objet et
		// charge par défaut viennent du banc vivant, pas d'un 10.200.0.1
		// en dur. En external, le petit objet reste VIDE : jamais de charge
		// implicite (un A+ sans charge est un mensonge). L'opérateur garde
		// la main : champs explicites > env (CGO_SMALL_URL/CGO_BULK_ADDR) > banc.
		sinks := audit.Sinks{Mode: "external", Target: "8.8.8.8"}
		if d.SinksFn != nil {
			sinks = d.SinksFn()
		} else {
			sinks = audit.ResolveSinks()
		}
		if body.Target == "" {
			body.Target = sinks.Target
		}
		// Sens et flux : défauts up/4, bornes imposées — un « both » débridé
		// double la charge, un « flows » hors 1–8 sort du régime documenté.
		switch body.Direction {
		case "", "up":
			body.Direction = "up"
		case "down", "both":
		default:
			writeErr(w, r, "direction invalide (up|down|both)", http.StatusBadRequest)
			return
		}
		if body.Flows <= 0 {
			body.Flows = 4
		}
		if body.Flows > 8 {
			writeErr(w, r, "flows hors bornes (1–8)", http.StatusBadRequest)
			return
		}
		if body.SmallURL == "" {
			// external laisse le petit objet vide (pas de charge implicite) ;
			// l'env garde la main sur les liaisons de terrain.
			body.SmallURL = envDefault("CGO_SMALL_URL", sinks.SmallURL)
		}
		if body.BulkAddr == "" {
			body.BulkAddr = audit.EnvOrBulk(sinks)
		}
		if body.Iface == "" {
			body.Iface = os.Getenv("CGO_AUDIT_IFACE")
		}
		if body.RouterOSIface == "" {
			body.RouterOSIface = os.Getenv("CGO_AUDIT_ROUTEROS_IFACE")
		}
		auditMu.Lock()
		if auditRunning {
			auditMu.Unlock()
			writeErr(w, r, "audit already running", http.StatusConflict)
			return
		}
		// contexte détaché de la requête HTTP mais annulable par l'opérateur
		runCtx, cancel := context.WithCancel(context.Background())
		auditRunning = true
		auditCancel = cancel
		auditStartedAt = time.Now()
		auditTotalSec = body.Duration
		auditProgress.Store(audit.Progress{Phase: "repos"})
		// Le rapport en mémoire appartient à l'audit PRÉCÉDENT : le garder
		// pendant ce run ferait lire un ancien rapport comme le résultat de
		// celui-ci si le nouveau échoue. L'archive disque reste disponible.
		lastAuditReport = nil
		auditMu.Unlock()
		asymmetric := true
		if body.Asymmetric != nil {
			asymmetric = *body.Asymmetric
		}
		go func() {
			defer func() {
				// Une panne de dépendance ne doit pas tuer le dashboard :
				// net/http ne récupère pas les goroutines détachées. On
				// journalise la pile, on efface le rapport de ce run et on
				// rend la main (auditRunning=false) pour le suivant.
				if rec := recover(); rec != nil {
					log.Printf("audit: panne interne : %v\n%s", rec, debug.Stack())
					recordEvent("audit", fmt.Sprintf("panne interne — %s (%s) : %v", body.Site, body.LinkType, rec))
					auditMu.Lock()
					lastAuditReport = nil
					auditMu.Unlock()
				}
				cancel()
				auditMu.Lock()
				auditRunning = false
				auditCancel = nil
				auditMu.Unlock()
			}()
			p := auditParamsFromRequest(map[string]any{
				"site": body.Site, "link_type": body.LinkType, "provider": body.Provider,
				"duration": float64(body.Duration), "target": body.Target,
				"direction": body.Direction, "flows": float64(body.Flows),
				"small_url": body.SmallURL, "bulk_addr": body.BulkAddr,
			})
			deps := audit.Deps{}
			if d.AuditDepsFn != nil {
				deps = d.AuditDepsFn()
			}
			deps.OnProgress = func(pr audit.Progress) { auditProgress.Store(pr) }
			res, err := audit.Run(runCtx, p, deps)
			// Annulation opérateur : Run rend context.Canceled sans résultat
			// partiel. Le rapport précédent a été effacé au démarrage : un
			// arrêt ne peut donc pas être maquillé en audit terminé. Un audit
			// achevé juste avant l'annulation garde son résultat (err == nil).
			if errors.Is(err, context.Canceled) {
				recordEvent("audit", fmt.Sprintf("annulé par l'opérateur — %s (%s)", p.Site, p.LinkType))
				return
			}
			if err != nil {
				recordEvent("audit", fmt.Sprintf("échec — %s (%s) : %s", p.Site, p.LinkType, err))
				return
			}
			// Evidence first: the report is sealed from the row this audit
			// appended, so the artifact and the CSV cannot drift apart.
			if err := audit.AppendLinkAudit("data", res); err != nil {
				recordEvent("audit", fmt.Sprintf("échec écriture CSV — %s : %s", p.Site, err))
			}
			link := remediate.Link{Interface: body.Iface, RouterOSInterface: body.RouterOSIface}
			rep, rerr := audit.BuildReport(res, link, body.Qdisc, asymmetric, body.CPUConstrained)
			if rerr != nil {
				recordEvent("audit", fmt.Sprintf("rapport impossible — %s : %s", p.Site, rerr))
			} else {
				if err := rep.SealSource(filepath.Join("data", "link_audit.csv")); err != nil {
					recordEvent("audit", fmt.Sprintf("scellement impossible — %s : %s", p.Site, err))
				}
				if _, _, err := audit.WriteArtifacts(rep, "data"); err != nil {
					recordEvent("audit", fmt.Sprintf("artefacts non écrits — %s : %s", p.Site, err))
				}
			}
			auditMu.Lock()
			lastAudit = res
			lastAuditReport = rep
			auditMu.Unlock()
			grade := res.BloatGrade
			if grade == "" {
				grade = "non mesuré"
			}
			recordEvent("audit", fmt.Sprintf("terminé — %s (%s), p95 %g ms, bufferbloat %s", p.Site, p.LinkType, res.RTTIdleP95, grade))
		}()
		// la ligne d'ouverture nomme le banc détecté (M4) : l'opérateur voit
		// le chemin qui va mesurer AVANT de lire un verdict.
		smallNote, bulkNote := body.SmallURL, body.BulkAddr
		if smallNote == "" {
			smallNote = "—"
		}
		if bulkNote == "" {
			bulkNote = "—"
		}
		recordEvent("audit", fmt.Sprintf("démarré — %s (%s) %ds, banc %s, cible %s, charge %s, small %s",
			body.Site, body.LinkType, body.Duration, sinks.Mode, body.Target, bulkNote, smallNote))
		writeJSON(w, map[string]any{"started": true})
	})
	mux.HandleFunc("GET /api/audit/status", func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		defer auditMu.Unlock()
		writeJSON(w, map[string]any{"running": auditRunning, "last": lastAudit, "report": lastAuditReport})
	})
	// Audit vivant — la carte « Observer » sonde ceci toutes les 2 s. Pas
	// d'audit ⇒ running:false seul : aucun zéro ne se lit comme mesuré.
	mux.HandleFunc("GET /api/audit/live", func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		running, started, total := auditRunning, auditStartedAt, auditTotalSec
		auditMu.Unlock()
		if !running {
			writeJSON(w, map[string]any{"running": false})
			return
		}
		pr, _ := auditProgress.Load().(audit.Progress)
		rttSamples, smallSamples := pr.SamplesRTT, pr.SamplesSmall
		if rttSamples == nil {
			rttSamples = []float64{}
		}
		if smallSamples == nil {
			smallSamples = []float64{}
		}
		writeJSON(w, map[string]any{
			"running":          true,
			"phase":            pr.Phase,
			"elapsed_sec":      int(time.Since(started).Seconds()),
			"total_sec":        total,
			"phase_remain_sec": pr.PhaseRemainSec,
			"last_rtt_ms":      pr.LastRTTms,
			"last_small_ms":    pr.LastSmallMs,
			"samples":          map[string]any{"rtt_ms": rttSamples, "small_ms": smallSamples},
		})
	})
	// Arrêt opérateur — la goroutine d'audit voit son contexte annulé,
	// n'écrit rien de partiel et journalise « annulé par l'opérateur ».
	mux.HandleFunc("POST /api/audit/stop", func(w http.ResponseWriter, _ *http.Request) {
		auditMu.Lock()
		cancel, running := auditCancel, auditRunning
		auditMu.Unlock()
		if !running || cancel == nil {
			writeJSON(w, map[string]any{"stopped": false})
			return
		}
		cancel()
		writeJSON(w, map[string]any{"stopped": true})
	})
	// Rapport d'audit — artefact persistant, pas seulement l'état en mémoire.
	// Sans audit_id, le dernier rapport ; avec, la copie archivée.
	mux.HandleFunc("GET /api/audit/report", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.URL.Query().Get("audit_id"))
		var rep *audit.Report
		auditMu.Lock()
		if id == "" || (lastAuditReport != nil && lastAuditReport.AuditID == id) {
			rep = lastAuditReport
		}
		auditMu.Unlock()
		if rep == nil {
			var err error
			rep, err = readAuditReport(id)
			if err != nil {
				writeErr(w, r, "rapport introuvable : "+id, http.StatusNotFound)
				return
			}
		}
		if r.URL.Query().Get("format") == "md" || r.URL.Query().Get("format") == "markdown" {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			_, _ = io.WriteString(w, rep.RenderMarkdown())
			return
		}
		writeJSON(w, rep)
	})
	// Remédiation — le script est déjà dans le rapport scellé ; l'endpoint
	// existe pour que le dashboard copie sans parser le Markdown. Un iface
	// fourni ici ne réécrit pas les mesures : BuildReport repart du Result
	// mesuré et ne fait que nommer l'équipement.
	mux.HandleFunc("GET /api/remediate", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := strings.TrimSpace(q.Get("audit_id"))
		var rep *audit.Report
		auditMu.Lock()
		if id == "" || (lastAuditReport != nil && lastAuditReport.AuditID == id) {
			rep = lastAuditReport
		}
		auditMu.Unlock()
		if rep == nil {
			var err error
			rep, err = readAuditReport(id)
			if err != nil {
				writeErr(w, r, "aucun rapport d'audit — lancez cgo audit", http.StatusNotFound)
				return
			}
		}
		// Rebuild only when the operator names an interface and the measured
		// result is still in memory; otherwise return the archived report's
		// remediation verbatim.
		if iface := strings.TrimSpace(q.Get("iface")); iface != "" {
			auditMu.Lock()
			res := lastAudit
			auditMu.Unlock()
			if res != nil && res.AuditID == rep.AuditID {
				link := remediate.Link{Interface: iface, RouterOSInterface: strings.TrimSpace(q.Get("routeros_iface"))}
				asym := q.Get("asymmetric") != "false"
				rebuilt, err := audit.BuildReport(res, link, q.Get("qdisc"), asym, q.Get("cpu_constrained") == "true")
				if err == nil {
					rep = rebuilt
				}
			}
		}
		writeJSON(w, map[string]any{"audit_id": rep.AuditID, "remediation": rep.Remediation, "unavailable": rep.Unavailable})
	})
	// protocole RQ1 matérialisé : l'opérateur choisit au lieu de recopier.
	mux.HandleFunc("GET /api/audit/presets", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"presets": audit.Presets()})
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
	// capacity = throughput, delay = rtt_idle_p95, loss = loss_pct mesurés.
	// ?audit_id= choisit la ligne (défaut : dernière) — l'opérateur importe
	// la fibre sans relancer la pointe.
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
		// ?audit_id= : la ligne choisie, sinon la dernière
		if want := r.URL.Query().Get("audit_id"); want != "" {
			found := false
			for _, row := range rows[1:] {
				for i, h := range hdr {
					if h == "audit_id" && i < len(row) && row[i] == want {
						last, found = row, true
					}
				}
				if found {
					break
				}
			}
			if !found {
				writeErr(w, r, "audit introuvable : "+want, http.StatusNotFound)
				return
			}
		}
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
		loss, _ := strconv.ParseFloat(get("loss_pct"), 64)
		id := "P-audit"
		if v := get("audit_id"); v != "" {
			id = "P-" + v
		}
		p := model.Profile{ID: id, CapacityMbps: cap, DelayMs: delay, LossPct: loss}
		// Un profil importé est rejoué comme un lien réel : une capacité ou un
		// délai non mesuré ne peut pas devenir 20 Mbit/s / 100 ms. L'ancien
		// repli fabriquait un lien de laboratoire et l'appelait « audit ».
		if p.CapacityMbps <= 0 || p.DelayMs <= 0 {
			writeErr(w, r, "audit sans capacité ou RTT idle mesurés — profil non importable (relancez l'audit avec charge bulk et cible joignable)", http.StatusUnprocessableEntity)
			return
		}
		if err := profile.Import(p); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		recordEvent("profil", fmt.Sprintf("%s importé depuis audit — %g Mbit/s, %g ms, perte %g %%", id, p.CapacityMbps, p.DelayMs, p.LossPct))
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
			// Ligne CSV: id,capacity,delay,jitter,loss[,capacity_up,loss_burst_p,r,h,k]
			// — positions 5..9 optionnelles (profils asymétriques + rafales) ;
			// absentes = symétrique + perte uniforme (historique inchangé)
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
			if len(row) > 5 {
				p.CapacityUpMbps = num(row[5])
			}
			if len(row) > 6 {
				p.LossBurstP = num(row[6])
			}
			if len(row) > 7 {
				p.LossBurstR = num(row[7])
			}
			if len(row) > 8 {
				p.LossBurstH = num(row[8])
			}
			if len(row) > 9 {
				p.LossBurstK = num(row[9])
			}
		} else if err := json.Unmarshal(body, &p); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if p.ID == "" {
			writeErr(w, r, "id required", http.StatusBadRequest)
			return
		}
		if profile.Reserved(p.ID) {
			writeErr(w, r, fmt.Sprintf("profil %s réservé (référentiel P1–P4)", p.ID), http.StatusBadRequest)
			return
		}
		if err := profile.Import(p); err != nil {
			writeErr(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "profile": p})
	})
	// Suppression d'un profil importé (le référentiel P1–P4 est intouchable —
	// refusé par profile.Delete, pas par le vide).
	mux.HandleFunc("POST /api/profile/delete", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, r, "bad json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.ID) == "" {
			writeErr(w, r, "id required", http.StatusBadRequest)
			return
		}
		if err := profile.Delete(strings.TrimSpace(body.ID)); err != nil {
			writeErr(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		recordEvent("profil", fmt.Sprintf("%s supprimé (import retiré)", strings.TrimSpace(body.ID)))
		writeJSON(w, map[string]any{"ok": true})
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
	hub.Serve(func() any { return next() })

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
			// FailedGates — portes en échec (runs futurs ; absent des archives
			// passées, champ omis alors — additif).
			FailedGates []string `json:"failed_gates,omitempty"`
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
	// Quarantaine/summary — reconstruction live des causes (même heuristique
	// que l'annexe : low/high/zero/G3, non-exclusives). Aucune constante
	// gelée : fonctionne sur tout jeu de runs.
	mux.HandleFunc("GET /api/quarantine/summary", func(w http.ResponseWriter, _ *http.Request) {
		files, _ := filepath.Glob(filepath.Join("data", "runs", "*", "aqm_eval.csv"))
		var total, low, high, zero, g3 int
		for _, f := range files {
			if strings.HasPrefix(filepath.Base(filepath.Dir(f)), "run-smoke") {
				continue
			}
			header, rows, err := results.ReadAQM(f)
			if err != nil {
				continue
			}
			iProf := results.ColIndex(header, "profile")
			iGood := results.ColIndex(header, "bulk_goodput_mbps")
			iRTT := results.ColIndex(header, "rtt_p95_ms")
			iGate := results.ColIndex(header, "gate_status")
			iDir := results.ColIndex(header, "direction")
			get := func(row []string, i int) string {
				if i < 0 || i >= len(row) {
					return ""
				}
				return row[i]
			}
			for _, row := range rows {
				if get(row, iGate) != "invalid" {
					continue
				}
				total++
				gp, _ := strconv.ParseFloat(get(row, iGood), 64)
				rtt, _ := strconv.ParseFloat(get(row, iRTT), 64)
				model.ProfilesMu.RLock()
				prof, ok := model.Profiles[get(row, iProf)]
				model.ProfilesMu.RUnlock()
				if !ok {
					continue
				}
				capUp, delay := prof.CapUp(), prof.DelayMs
				floor := capUp * 0.5
				if d := get(row, iDir); d == "down" {
					floor = prof.CapDown() * 0.01
				}
				if gp == 0 {
					zero++
				}
				if gp < floor || gp > capUp*1.1+0.5 {
					if gp < floor {
						low++
					} else {
						high++
					}
				}
				if rtt > delay*10+200 {
					g3++
				}
			}
		}
		writeJSON(w, map[string]any{
			"total_invalid": total, "g4_low": low, "g4_high": high,
			"empty_probes": zero, "g3_implausible": g3,
			"note": "catégories non-exclusives ; G4 = porte de régime (Mathis), pas de qualité",
		})
	})
	mux.HandleFunc("GET /api/events", func(w http.ResponseWriter, _ *http.Request) {
		eventsMu.Lock()
		defer eventsMu.Unlock()
		out := make([]map[string]any, len(eventsRing))
		copy(out, eventsRing)
		writeJSON(w, map[string]any{"events": out})
	})
	mux.Handle("/", spa)
	return Handler{Handler: crossSiteGuard(mux), hubClose: func() { hub.Close() }}
}

// crossSiteGuard — un navigateur venu d'un autre site ne déclenche pas de
// mutation (CSRF) : Sec-Fetch-Site: cross-site est refusé sur les verbes non
// sûrs. Les clients non-navigateur (curl, CLI, tests) n'envoient pas cet
// en-tête et restent inchangés ; aucun jeton n'est introduit.
func crossSiteGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				writeErr(w, r, "requête cross-site refusée", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// truncateUTF8 — coupe à max octets sans fendre un caractère : un site coupé
// au milieu d'un « é » laisserait un U+FFFD dans le CSV et le rapport.
func truncateUTF8(s string, max int) string {
	if len(s) <= max {
		return s
	}
	for max > 0 && s[max]&0xC0 == 0x80 {
		max--
	}
	return s[:max]
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	// Jamais d'échec silencieux (un NaN donnait un 200 vide : vu en prod).
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
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

// auditParamsFromRequest — le mapping requête → audit.Params, défauts
// up/4 (le régime du banc durci). Les champs de terrain fournis passent
// tels quels : cette fonction mappe, le handler valide et borne.
func auditParamsFromRequest(body map[string]any) audit.Params {
	get := func(k string) string {
		if v, ok := body[k].(string); ok {
			return v
		}
		return ""
	}
	num := func(k string, def int) int {
		// JSON décode en float64 ; un appel Go direct fournit souvent un
		// int — les deux sont acceptés, jamais réécrits en silence.
		switch v := body[k].(type) {
		case float64:
			if v > 0 {
				return int(v)
			}
		case int:
			if v > 0 {
				return v
			}
		}
		return def
	}
	dir := get("direction")
	if dir == "" {
		dir = "up"
	}
	return audit.Params{
		AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()),
		Site:    get("site"), LinkType: get("link_type"), Provider: get("provider"),
		Duration: num("duration", 60), Target: get("target"),
		Direction: dir, Flows: num("flows", 4),
		SmallURL: get("small_url"), BulkAddr: get("bulk_addr"),
	}
}
