package results

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Realms4239/cgo/metrics"
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
	Degraded    int    `json:"degraded"`

	RTTp95Median   float64    `json:"rtt_p95_median"`
	RTTp95IQR      [2]float64 `json:"rtt_p95_iqr"`
	Smallp95Median float64    `json:"small_p95_median"`
	Smallp95IQR    [2]float64 `json:"small_p95_iqr"`
	GoodputMedian  float64    `json:"goodput_median"`
	DeadlineMedian float64    `json:"deadline_median"`
	WastedMedian   float64    `json:"wasted_median"`
	CostMedian     float64    `json:"cost_median"`
	// Valid-only strict (degraded exclus) — preuve pilote : mêmes fonctions
	// que pkg/metrics/stats.go (miroir de extract-stats.js, seed 42).
	Smallp95ValidN         int        `json:"small_p95_valid_n"`
	Smallp95ValidMedian    float64    `json:"small_p95_valid_median"`
	Smallp95ValidIQR       [2]float64 `json:"small_p95_valid_iqr"`
	Smallp95ValidCI95      [2]float64 `json:"small_p95_valid_ci95"`
	DeadlineValidN         int        `json:"deadline_valid_n"`
	DeadlineValidMedian    float64    `json:"deadline_valid_median"`
	DeadlineValidCI95      [2]float64 `json:"deadline_valid_ci95"`
	Best                   bool       `json:"best,omitempty"`
	HardwareRecommendation string     `json:"hardware_recommendation"`
	// Runs — les identifiants des runs qui ont nourri les mesures de la cellule
	// (au moins une ligne non invalid), triés. Une médiane sans provenance n'est
	// pas une preuve ; un run entièrement en quarantaine n'en est pas une.
	Runs []string `json:"runs,omitempty"`
}

// Scan agrège tous les aqm_eval.csv sous dataDir (dernier run, sinon tous).
// Parcourt les gelés, groupe, médianes. Vide -> nil, nil.
// runIDs — les runs d'une cellule, triés : le client peut les citer sans
// hériter de l'ordre d'itération d'une map.
func runIDs(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

var (
	scanCacheMu  sync.Mutex
	scanCacheKey string
	scanCacheVal []Group
)

// scanFingerprint — clé de cache : emprunte (taille, mtime) de chaque CSV du
// corpus. Change dès qu'un fichier bouge/part/apparaît ; stable sinon.
func scanFingerprint(files []string) string {
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		if st, err := os.Stat(f); err == nil {
			fmt.Fprintf(h, "%s|%d|%d;", f, st.Size(), st.ModTime().UnixNano())
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Fingerprint — la même clé, exposée pour l'ETag /api/results.
func Fingerprint(dataDir, runFilter string) string {
	pattern := filepath.Join(dataDir, "*", "aqm_eval.csv")
	if runFilter != "" {
		pattern = filepath.Join(dataDir, runFilter, "aqm_eval.csv")
	}
	files, _ := filepath.Glob(pattern)
	if len(files) == 0 {
		return ""
	}
	return scanFingerprint(files)
}

// Scan — mémoïsée tant que le corpus est statique (taille+mtime). Un
// /api/results coûtait ~140–300 ms de parse+stats par appel ; après le
// premier appel, ~zéro. Le filtre est dans la clé ; l'invalidation est
// conservatrice (aucune écriture sans changer un stat de fichier).
func Scan(dataDir, runFilter string) ([]Group, error) {
	if runFilter != "" && !ValidRunID(runFilter) {
		return nil, ErrInvalidRunID
	}
	pattern := filepath.Join(dataDir, "*", "aqm_eval.csv")
	if runFilter != "" {
		pattern = filepath.Join(dataDir, runFilter, "aqm_eval.csv")
	}
	files, _ := filepath.Glob(pattern)
	if len(files) == 0 {
		return nil, nil
	}
	key := dataDir + "|" + runFilter + "|" + scanFingerprint(files)
	scanCacheMu.Lock()
	if key == scanCacheKey {
		v := scanCacheVal
		scanCacheMu.Unlock()
		return v, nil
	}
	scanCacheMu.Unlock()
	out, err := scanUncached(dataDir, runFilter, files)
	if err == nil {
		scanCacheMu.Lock()
		scanCacheKey, scanCacheVal = key, out
		scanCacheMu.Unlock()
	}
	return out, err
}

// scanUncached — le vrai parse+stats (pas de cache ; benchmark + Scan).
func scanUncached(dataDir, runFilter string, files []string) ([]Group, error) {
	type bucket struct {
		rtts, smalls, goodputs, deadlines, wasteds, costs []float64
		vSmalls, vDeadlines                               []float64 // valid-only
		quarantined                                       int
		degraded                                          int
		count                                             int
		profile, qdisc, cc, direction                     string
		runs                                              map[string]bool
	}
	type indexes struct {
		prof, qdisc, cc, rtt, small, good, dead, waste, cost, gate, dir int
	}
	type parsedRun struct {
		f      string
		header []string
		rows   [][]string
		ix     indexes
		ok     bool
	}
	// Parse parallèle des CSV (ordre de glob réintroduit au pliage — zéro
	// drift) : chaque worker interprète son fichier sans état partagé.
	parsed := make([]parsedRun, len(files))
	{
		var wg sync.WaitGroup
		sem := make(chan struct{}, runtime.NumCPU())
		for i, f := range files {
			// run-smoke* = fixture de test du pipeline, pas une mesure : exclue
			// du corpus publié (parité avec extract-stats.js — sinon sa ligne
			// valid factice fait médiane à elle seule côté UI/rapport/figures)
			if strings.HasPrefix(filepath.Base(filepath.Dir(f)), "run-smoke") {
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(i int, f string) {
				defer wg.Done()
				defer func() { <-sem }()
				// schémas 17/18/19 colonnes, chaque fichier porte son en-tête.
				header, rows, err := ReadAQM(f)
				if err != nil {
					return
				}
				ix := indexes{
					prof: ColIndex(header, "profile"), qdisc: ColIndex(header, "qdisc"), cc: ColIndex(header, "cc"),
					rtt: ColIndex(header, "rtt_p95_ms"), small: ColIndex(header, "small_p95_ms"),
					good: ColIndex(header, "bulk_goodput_mbps"), dead: ColIndex(header, "deadline_ok_pct"),
					waste: ColIndex(header, "wasted_bytes"), cost: ColIndex(header, "cost_ar_per_h"),
					gate: ColIndex(header, "gate_status"), dir: ColIndex(header, "direction"),
				}
				if ix.prof < 0 || ix.qdisc < 0 || ix.cc < 0 {
					return
				}
				parsed[i] = parsedRun{f: f, header: header, rows: rows, ix: ix, ok: true}
			}(i, f)
		}
		wg.Wait()
	}
	buckets := map[string]*bucket{}
	for _, p := range parsed {
		if !p.ok {
			continue
		}
		f, rows := p.f, p.rows
		iProf, iQdisc, iCC := p.ix.prof, p.ix.qdisc, p.ix.cc
		iRTT, iSmall, iGood, iDead := p.ix.rtt, p.ix.small, p.ix.good, p.ix.dead
		iWaste, iCost, iGate, iDir := p.ix.waste, p.ix.cost, p.ix.gate, p.ix.dir
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
				b = &bucket{profile: profile, qdisc: qdisc, cc: cc, direction: direction, runs: map[string]bool{}}
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
			// Après le filtre : un run dont toutes les lignes sont en quarantaine
			// n'a nourri aucune mesure et ne peut pas figurer dans la provenance.
			b.runs[filepath.Base(filepath.Dir(f))] = true
			if get(r, iGate) == "degraded" {
				b.degraded++
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
	summarizeBucket := func(b *bucket) Group {
		rs := metrics.Summarize(b.rtts)
		ss := metrics.Summarize(b.smalls)
		gs := metrics.Summarize(b.goodputs)
		ds := metrics.Summarize(b.deadlines)
		ws := metrics.Summarize(b.wasteds)
		cs := metrics.Summarize(b.costs)
		vq1, _, vq3 := metrics.Quartiles(b.vSmalls)
		vlo, vhi := metrics.BootstrapMedianCI95(b.vSmalls, 10000, 42)
		dlo, dhi := metrics.BootstrapMedianCI95(b.vDeadlines, 10000, 42)
		return Group{
			Profile: b.profile, Qdisc: b.qdisc, CC: b.cc, Direction: b.direction,
			Count: b.count, Quarantined: b.quarantined, Degraded: b.degraded,
			RTTp95Median: rs.Median, RTTp95IQR: [2]float64{rs.IQRLow, rs.IQRHigh},
			Smallp95Median: ss.Median, Smallp95IQR: [2]float64{ss.IQRLow, ss.IQRHigh}, GoodputMedian: gs.Median, DeadlineMedian: ds.Median,
			WastedMedian: ws.Median, CostMedian: cs.Median,
			Smallp95ValidN: len(b.vSmalls), Smallp95ValidMedian: metrics.Median(b.vSmalls),
			Smallp95ValidIQR: [2]float64{vq1, vq3}, Smallp95ValidCI95: [2]float64{vlo, vhi},
			DeadlineValidN: len(b.vDeadlines), DeadlineValidMedian: metrics.Median(b.vDeadlines),
			DeadlineValidCI95: [2]float64{dlo, dhi},
			Runs:              runIDs(b.runs),
		}
	}
	// Résumé par cellule : les deux bootstraps 10 000 × 35 groupes sont la
	// hot path (~150–250 ms sur 4 cœurs). Chaque cellule est indépendante
	// (LCG seedé, sort final) → parallélisation sans drift de résultat.
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	out := make([]Group, len(keys))
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())
	for i, k := range keys {
		b := buckets[k]
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, b *bucket) {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = summarizeBucket(b)
		}(i, b)
	}
	wg.Wait()
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
