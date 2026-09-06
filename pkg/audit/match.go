package audit

import (
	"fmt"

	"github.com/Realms4239/cgo/pkg/model"
)

// MatchProfile — quel profil P1–P4 le lien audité ressemble-t-il ?
// (perspective « calibration des profils par site » du mémoire, outillée.)
// Distance normalisée sur le RTT idle p95 vs délai du profil (le RTT mesuré
// au banc vaut ≈ DelayMs : netem sur un seul saut, cf. P3 600→628 ms) et,
// quand le débit a été mesuré (bulk sink), sur la capacité. Sans débit, le
// rapprochement se fait au délai seul et le dit — un audit sans sink ne
// calibre pas la capacité, il ne prétend pas le faire.
func MatchProfile(r *Result) (id, delta string) {
	model.ProfilesMu.RLock()
	defer model.ProfilesMu.RUnlock()
	best, bestScore := "", 1e18
	var bestD string
	for pid, p := range model.Profiles {
		if len(pid) > 2 || pid[0] != 'P' {
			continue // profils importés (P-audit-…) hors course : on calibre VERS le référentiel
		}
		ref := p.DelayMs
		if ref <= 0 {
			ref = 20
		}
		dRTT := (r.RTTIdleP95 - p.DelayMs) / ref
		if dRTT < 0 {
			dRTT = -dRTT
		}
		score := dRTT
		dCap := ""
		if r.ThroughputMbps > 0 && p.CapacityMbps > 0 {
			d := (r.ThroughputMbps - p.CapacityMbps) / p.CapacityMbps
			if d < 0 {
				d = -d
			}
			score += d
			dCap = fmt.Sprintf(", débit %.1f vs %.0f Mbit/s", r.ThroughputMbps, p.CapacityMbps)
		}
		if score < bestScore {
			bestScore = score
			best = pid
			if r.ThroughputMbps > 0 {
				bestD = fmt.Sprintf("rtt %.0f vs %g ms%s", r.RTTIdleP95, p.DelayMs, dCap)
			} else {
				bestD = fmt.Sprintf("rtt %.0f vs %g ms (débit non mesuré — pas de sink)", r.RTTIdleP95, p.DelayMs)
			}
		}
	}
	if best == "" {
		return "", "aucun profil de référence"
	}
	return best, bestD
}
