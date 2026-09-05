package results

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/pkg/metrics"
)

// Group is one aggregated cell profile×qdisc×cc.
type Group struct {
	Profile string `json:"profile"`
	Qdisc   string `json:"qdisc"`
	CC      string `json:"cc"`
	// Direction — sens de charge ("up" défaut historique, "down"). Les
	// groupes up/down ne fusionnent jamais : même cellule, physique opposée.
	Direction   string `json:"direction"`
	Count       int    `json:"count"`
	Quarantined int    `json:"quarantined"`

	RTTp95Median           float64    `json:"rtt_p95_median"`
	RTTp95IQR              [2]float64 `json:"rtt_p95_iqr"`
	Smallp95Median         float64    `json:"small_p95_median"`
	Smallp95IQR            [2]float64 `json:"small_p95_iqr"`
	GoodputMedian          float64    `json:"goodput_median"`
	DeadlineMedian         float64    `json:"deadline_median"`
	WastedMedian           float64    `json:"wasted_median"`
	CostMedian             float64    `json:"cost_median"`
	// Valid-only strict (degraded exclus) — preuve pilote : mêmes fonctions
	// que pkg/metrics/stats.go (miroir de extract-stats.js, seed 42).
	Smallp95ValidN      int       `json:"small_p95_valid_n"`
	Smallp95ValidMedian float64   `json:"small_p95_valid_median"`
	Smallp95ValidIQR    [2]float64 `json:"small_p95_valid_iqr"`
	Smallp95ValidCI95   [2]float64 `json:"small_p95_valid_ci95"`
	DeadlineValidN      int       `json:"deadline_valid_n"`
	DeadlineValidMedian float64   `json:"deadline_valid_median"`
	DeadlineValidCI95   [2]float64 `json:"deadline_valid_ci95"`
	Best                   bool       `json:"best,omitempty"`
	HardwareRecommendation string     `json:"hardware_recommendation"`
}

// Scan agrège tous les aqm_eval.csv sous dataDir (dernier run, sinon tous).
// Parcourt les gelés, groupe, médianes. Vide -> nil, nil.
func Scan(dataDir, runFilter string) ([]Group, error) {
	pattern := filepath.Join(dataDir, "*", "aqm_eval.csv")
	if runFilter != "" {
		pattern = filepath.Join(dataDir, runFilter, "aqm_eval.csv")
	}
	files, _ := filepath.Glob(pattern)
	if len(files) == 0 {
		return nil, nil
	}
	type bucket struct {
		rtts, smalls, goodputs, deadlines, wasteds, costs []float64
		vSmalls, vDeadlines                               []float64 // valid-only
		quarantined                                       int
		count                                             int
		profile, qdisc, cc, direction                     string
	}
	buckets := map[string]*bucket{}
	for _, f := range files {
		// run-smoke* = fixture de test du pipeline, pas une mesure : exclue
		// du corpus publié (parité avec extract-stats.js — sinon sa ligne
		// valid factice fait médiane à elle seule côté UI/rapport/figures)
		if strings.HasPrefix(filepath.Base(filepath.Dir(f)), "run-smoke") {
			continue
		}
		// schémas 17/18/19 colonnes (insertions qdi_ms puis voip_r), chaque
		// fichier porte son en-tête. Plus aucun index positionnel.
		header, rows, err := ReadAQM(f)
		if err != nil {
			continue
		}
		iProf, iQdisc, iCC := ColIndex(header, "profile"), ColIndex(header, "qdisc"), ColIndex(header, "cc")
		if iProf < 0 || iQdisc < 0 || iCC < 0 {
			continue
		}
		iRTT := ColIndex(header, "rtt_p95_ms")
		iSmall := ColIndex(header, "small_p95_ms")
		iGood := ColIndex(header, "bulk_goodput_mbps")
		iDead := ColIndex(header, "deadline_ok_pct")
		iWaste := ColIndex(header, "wasted_bytes")
		iCost := ColIndex(header, "cost_ar_per_h")
		iGate := ColIndex(header, "gate_status")
		iDir := ColIndex(header, "direction")
		get := func(r []string, i int) string {
			if i < 0 || i >= len(r) {
				return ""
			}
			return r[i]
		}
		add := func(dst *[]float64, v string) {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				*dst = append(*dst, f)
			}
		}
		for _, r := range rows {
			profile, qdisc, cc := get(r, iProf), get(r, iQdisc), get(r, iCC)
			if profile == "" || qdisc == "" || cc == "" {
				continue
			}
			// direction gelée, défaut "up" pour les 150 runs historiques sans
			// la colonne : up et down ne fusionnent JAMAIS dans un groupe
			direction := get(r, iDir)
			if direction == "" {
				direction = "up"
			}
			key := profile + "|" + qdisc + "|" + cc + "|" + direction
			b := buckets[key]
			if b == nil {
				b = &bucket{profile: profile, qdisc: qdisc, cc: cc, direction: direction}
				buckets[key] = b
			}
			b.count++
			// Quarantaine seulement pour invalid — pas dégradé; 7 portes honnêtes.
			// Les médianes portent les lignes exploitables (non invalid), comme
			// extract-stats.js — sinon une ligne invalidée par les portes
			// fausse l'agrégat affiché.
		if get(r, iGate) == "invalid" {
			b.quarantined++
			continue
		}
		if get(r, iGate) == "valid" {
			add(&b.vSmalls, get(r, iSmall))
			add(&b.vDeadlines, get(r, iDead))
		}
			add(&b.rtts, get(r, iRTT))
			add(&b.smalls, get(r, iSmall))
			add(&b.goodputs, get(r, iGood))
			add(&b.deadlines, get(r, iDead))
			add(&b.wasteds, get(r, iWaste))
			add(&b.costs, get(r, iCost))
		}
	}
	var out []Group
	for _, b := range buckets {
		rs := metrics.Summarize(b.rtts)
		ss := metrics.Summarize(b.smalls)
		gs := metrics.Summarize(b.goodputs)
		ds := metrics.Summarize(b.deadlines)
		ws := metrics.Summarize(b.wasteds)
		cs := metrics.Summarize(b.costs)
		vq1, _, vq3 := metrics.Quartiles(b.vSmalls)
		vlo, vhi := metrics.BootstrapMedianCI95(b.vSmalls, 10000, 42)
		dlo, dhi := metrics.BootstrapMedianCI95(b.vDeadlines, 10000, 42)
		out = append(out, Group{
			Profile: b.profile, Qdisc: b.qdisc, CC: b.cc, Direction: b.direction,
			Count: b.count, Quarantined: b.quarantined,
			RTTp95Median: rs.Median, RTTp95IQR: [2]float64{rs.IQRLow, rs.IQRHigh},
			Smallp95Median: ss.Median, Smallp95IQR: [2]float64{ss.IQRLow, ss.IQRHigh}, GoodputMedian: gs.Median, DeadlineMedian: ds.Median,
			WastedMedian: ws.Median, CostMedian: cs.Median,
			Smallp95ValidN: len(b.vSmalls), Smallp95ValidMedian: metrics.Median(b.vSmalls),
			Smallp95ValidIQR: [2]float64{vq1, vq3}, Smallp95ValidCI95: [2]float64{vlo, vhi},
			DeadlineValidN: len(b.vDeadlines), DeadlineValidMedian: metrics.Median(b.vDeadlines),
			DeadlineValidCI95: [2]float64{dlo, dhi},
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Profile != out[j].Profile {
			return out[i].Profile < out[j].Profile
		}
		if out[i].Qdisc != out[j].Qdisc {
			return out[i].Qdisc < out[j].Qdisc
		}
		if out[i].CC != out[j].CC {
			return out[i].CC < out[j].CC
		}
		// direction : up (montée, défaut historique) avant down. Sans ce
		// tiebreak, l'itération aléatoire des maps + sort.Slice non stable
		// mélangeait les lignes liées à chaque appel, et le constat UI
		// (find du premier pfifo) basculait entre deux chargements.
		if out[i].Direction != out[j].Direction {
			return out[i].Direction == "up"
		}
		return false
	})
	// mark best smallest small_p95 per profile
	bestForProfile := map[string]int{}
	bestVal := map[string]float64{}
	for i, g := range out {
		if v, ok := bestVal[g.Profile]; !ok || g.Smallp95Median < v {
			bestVal[g.Profile] = g.Smallp95Median
			bestForProfile[g.Profile] = i
		}
	}
	for p, idx := range bestForProfile {
		if bestVal[p] > 0 {
			out[idx].Best = true
		}
	}
	// hardware recommendation per profile best (not per row) — deduped source pkg/results/hardware.go
	bestQdiscForProfile := map[string]string{}
	for _, g := range out {
		if g.Best {
			bestQdiscForProfile[g.Profile] = g.Qdisc
		}
	}
	for i := range out {
		if best, ok := bestQdiscForProfile[out[i].Profile]; ok {
			out[i].HardwareRecommendation = HardwareRecommendation(best, out[i].Profile)
		} else {
			out[i].HardwareRecommendation = HardwareRecommendation(out[i].Qdisc, out[i].Profile)
		}
	}
	return out, nil
}
