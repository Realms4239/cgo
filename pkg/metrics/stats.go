// stats.go — inférence minimale pour la preuve pilote (seed fixe 42).
package metrics

import (
	"math"
	"sort"
)

func sortedCopy(xs []float64) []float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	return s
}

func quantileSorted(s []float64, q float64) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	if len(s) == 1 {
		return s[0]
	}
	pos := float64(len(s)-1) * q
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return s[lo]
	}
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}

// Median — médiane (interpolation linéaire, même convention que Summarize).
func Median(xs []float64) float64 { return quantileSorted(sortedCopy(xs), 0.5) }

// Quartiles — Q1, médiane, Q3.
func Quartiles(xs []float64) (q1, med, q3 float64) {
	s := sortedCopy(xs)
	return quantileSorted(s, 0.25), quantileSorted(s, 0.5), quantileSorted(s, 0.75)
}

// IQR — Q3−Q1.
func IQR(xs []float64) float64 {
	q1, _, q3 := Quartiles(xs)
	return q3 - q1
}

// lcg — Park-Miller minimal pour un bootstrap seedé sans dépendance.
type lcg struct{ s uint64 }

func (r *lcg) next(n int) int {
	r.s = (r.s * 48271) % 2147483647
	return int(r.s % uint64(n))
}

// BootstrapMedianCI95 — IC95 percentile de la médiane (seed fixe → stable).
func BootstrapMedianCI95(xs []float64, resamples int, seed uint64) (lo, hi float64) {
	if len(xs) == 0 || resamples <= 0 {
		return math.NaN(), math.NaN()
	}
	r := &lcg{s: seed % 2147483647}
	if r.s == 0 {
		r.s = 1
	}
	meds := make([]float64, resamples)
	buf := make([]float64, len(xs))
	for i := 0; i < resamples; i++ {
		for j := range buf {
			buf[j] = xs[r.next(len(xs))]
		}
		meds[i] = quantileSorted(sortedCopy(buf), 0.5)
	}
	sort.Float64s(meds)
	return quantileSorted(meds, 0.025), quantileSorted(meds, 0.975)
}

// MannWhitneyTwoSided — U + p bilatéral (approx. normale, correction de continuité).
func MannWhitneyTwoSided(a, b []float64) (u, p float64) {
	type v struct {
		x float64
		g int
	}
	all := make([]v, 0, len(a)+len(b))
	for _, x := range a {
		all = append(all, v{x, 0})
	}
	for _, x := range b {
		all = append(all, v{x, 1})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].x < all[j].x })
	var ra float64
	i := 0
	for i < len(all) {
		j := i
		for j+1 < len(all) && all[j+1].x == all[i].x {
			j++
		}
		avg := float64(i+1+j+1) / 2 // rangs 1-based moyens sur ex æquo
		for k := i; k <= j; k++ {
			if all[k].g == 0 {
				ra += avg
			}
		}
		i = j + 1
	}
	n1, n2 := float64(len(a)), float64(len(b))
	u1 := ra - n1*(n1+1)/2
	u = u1
	mu := n1 * n2 / 2
	sd := math.Sqrt(n1 * n2 * (n1 + n2 + 1) / 12)
	if sd == 0 {
		return u, 1
	}
	z := (math.Abs(u1-mu) - 0.5) / sd // continuité
	p = math.Erfc(z / math.Sqrt2)
	return u, p
}

// CliffDelta — delta de Cliff (b−a dominant → +1).
func CliffDelta(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return math.NaN()
	}
	var more, less float64
	for _, x := range a {
		for _, y := range b {
			switch {
			case y > x:
				more++
			case y < x:
				less++
			}
		}
	}
	return (more - less) / (float64(len(a)) * float64(len(b)))
}

// CliffInterp — seuils Vargha-Delaney sur |d|.
func CliffInterp(d float64) string {
	a := math.Abs(d)
	switch {
	case a < 0.11:
		return "négligeable"
	case a < 0.28:
		return "petit"
	case a < 0.43:
		return "moyen"
	default:
		return "grand"
	}
}
