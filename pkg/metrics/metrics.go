// Package metrics — percentile, deadline et coût.
// Le navigateur ne recalcule rien; Go est le seul calculateur.
package metrics

import (
	"math"
	"sort"
)

// Percentile of a pre-sorted slice (linear interpolation, nearest-rank-ish).
func Percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	switch n {
	case 0:
		return 0
	case 1:
		return sorted[0]
	}
	rank := (p / 100) * float64(n-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	frac := rank - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

type Summary struct {
	P50, P95, Median float64
	IQRLow, IQRHigh  float64
}

// Summarize trie une copie et rend médiane + IQR + percentiles.
func Summarize(v []float64) Summary {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return Summary{
		Median:  Percentile(s, 50),
		P50:     Percentile(s, 50),
		P95:     Percentile(s, 95),
		IQRLow:  Percentile(s, 25),
		IQRHigh: Percentile(s, 75),
	}
}

// DeadlineOKPct — share of completions strictly under dueMs.
func DeadlineOKPct(completionsMs []float64, dueMs float64) float64 {
	if len(completionsMs) == 0 {
		return 0
	}
	ok := 0
	for _, v := range completionsMs {
		if v < dueMs {
			ok++
		}
	}
	return float64(ok) / float64(len(completionsMs)) * 100
}

// Coût: gaspillé / 4,5 Gio × 30 000 Ar/h.
const (
	planBytes = 4.5 * 1024 * 1024 * 1024
	planAR    = 30000
)

// JFI — Jain's fairness index (Σx)² / (n·Σx²), 0..1 (1 = perfectly fair).
// Inutilisé tant que l'API n'expose pas les valeurs par répétition (computeJFI côté front).
func JFI(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
	}
	if sumSq == 0 {
		return 0
	}
	return (sum * sum) / (float64(len(values)) * sumSq)
}

func CostARPerH(wastedBytes uint64) float64 {
	return float64(wastedBytes) / planBytes * planAR
}
