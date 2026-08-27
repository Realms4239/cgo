package results

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/Realms4239/cgo/pkg/metrics"
)

// Group is one aggregated cell profile×qdisc×cc.
type Group struct {
	Profile  string  `json:"profile"`
	Qdisc    string  `json:"qdisc"`
	CC       string  `json:"cc"`
	Count    int     `json:"count"`
	Quarantined int  `json:"quarantined"`

	RTTp95Median   float64 `json:"rtt_p95_median"`
	RTTp95IQR      [2]float64 `json:"rtt_p95_iqr"`
	Smallp95Median float64 `json:"small_p95_median"`
	GoodputMedian  float64 `json:"goodput_median"`
	DeadlineMedian float64 `json:"deadline_median"`
	WastedMedian float64 `json:"wasted_median"`
	CostMedian   float64 `json:"cost_median"`
	Best           bool    `json:"best,omitempty"`
	HardwareRecommendation string `json:"hardware_recommendation"`
}

// Scan aggregates all aqm_eval.csv under dataDir (latest run if multiple, else all).
// Ponytail: scan all files, group, compute medians. Empty -> nil, nil.
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
		quarantined int
		count int
		profile, qdisc, cc string
	}
	buckets := map[string]*bucket{}
	for _, f := range files {
		rows, err := readCSV(f)
		if err != nil {
			continue
		}
		for _, r := range rows {
			key := r[2]+"|"+r[3]+"|"+r[4] // profile|qdisc|cc
			b := buckets[key]
			if b == nil {
				b = &bucket{profile: r[2], qdisc: r[3], cc: r[4]}
				buckets[key] = b
			}
			b.count++
			// ponytail: quarantine only invalid — not degraded, not covariable; spec gates 7 honest
			if r[16] == "invalid" {
				b.quarantined++
			}
			if v, err := strconv.ParseFloat(r[7], 64); err == nil {
				b.rtts = append(b.rtts, v)
			}
			if v, err := strconv.ParseFloat(r[8], 64); err == nil {
				b.smalls = append(b.smalls, v)
			}
			if v, err := strconv.ParseFloat(r[10], 64); err == nil {
				b.goodputs = append(b.goodputs, v)
			}
			if v, err := strconv.ParseFloat(r[9], 64); err == nil {
				b.deadlines = append(b.deadlines, v)
			}
			if v, err := strconv.ParseFloat(r[13], 64); err == nil {
				b.wasteds = append(b.wasteds, v)
			}
			if v, err := strconv.ParseFloat(r[14], 64); err == nil {
				b.costs = append(b.costs, v)
			}
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
		out = append(out, Group{
			Profile: b.profile, Qdisc: b.qdisc, CC: b.cc,
			Count: b.count, Quarantined: b.quarantined,
			RTTp95Median: rs.Median, RTTp95IQR: [2]float64{rs.IQRLow, rs.IQRHigh},
			Smallp95Median: ss.Median, GoodputMedian: gs.Median, DeadlineMedian: ds.Median,
			WastedMedian: ws.Median, CostMedian: cs.Median,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Profile != out[j].Profile { return out[i].Profile < out[j].Profile }
		if out[i].Qdisc != out[j].Qdisc { return out[i].Qdisc < out[j].Qdisc }
		return out[i].CC < out[j].CC
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

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil { return nil, err }
	if len(rows) <= 1 { return nil, nil }
	return rows[1:], nil // skip header
}
