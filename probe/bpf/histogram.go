package bpf

// Log2Histogram — 64 buckets en mémoire bornée. Le bucket i couvre
// [2^i, 2^(i+1)) microsecondes (sortie de bpf_log2l sur un échantillon en us).
type Log2Histogram [64]uint64

// Percentile rend le percentile q (0 < q <= 1) en microsecondes, ou
// (0, false) quand moins de 1000 échantillons le soutiennent, quand q est
// invalide, ou quand le cumul des buckets n'atteint pas la cible. Aucun
// percentile n'est extrapolé : l'absence est un mot explicite, pas un zéro.
func Percentile(h Log2Histogram, q float64, samples uint64) (float64, bool) {
	if samples < 1000 || q <= 0 || q > 1 {
		return 0, false
	}
	target := uint64(float64(samples)*q + 0.999999)
	var seen uint64
	for i, n := range h {
		seen += n
		if seen >= target {
			return float64(uint64(1) << uint(i)), true // bord bas du bucket, us
		}
	}
	return 0, false
}
