package audit

import (
	"fmt"

	"github.com/Realms4239/cgo/audit/profile"
	"github.com/Realms4239/cgo/model"
)

// MatchProfile — quel profil P1–P4 le lien audité ressemble-t-il ?
// (perspective « calibration des profils par site » du mémoire, outillée.)
// La distance est normalisée sur le RTT idle p95 vs le délai du profil ET
// sur le débit mesuré vs la capacité. Sans débit (pas de sink bulk), aucun
// profil n'est revendiqué : un délai seul a déjà fait passer une 4G réelle
// pour Starlink (P4) — le rapprochement doit exiger les deux axes.
func MatchProfile(r *Result) (id, delta string) {
	if r.ThroughputMbps <= 0 {
		return "", "débit non mesuré — pas de sink bulk ; profil P1–P4 non déterminé"
	}
	profile.Load() // référentiel persisté : les imports survivent au redémarrage
	model.ProfilesMu.RLock()
	defer model.ProfilesMu.RUnlock()
	best, bestScore := "", 1e18
	var bestD string
	for pid, p := range model.Profiles {
		if len(pid) > 2 || pid[0] != 'P' {
			continue // profils importés (P-audit-…) hors course : on calibre VERS le référentiel
		}
		if p.CapacityMbps <= 0 {
			continue // pas de capacité de référence : le délai seul ne suffit pas
		}
		ref := p.DelayMs
		if ref <= 0 {
			ref = 20
		}
		dRTT := (r.RTTIdleP95 - p.DelayMs) / ref
		if dRTT < 0 {
			dRTT = -dRTT
		}
		dCapAbs := (r.ThroughputMbps - p.CapacityMbps) / p.CapacityMbps
		if dCapAbs < 0 {
			dCapAbs = -dCapAbs
		}
		score := dRTT + dCapAbs
		if score < bestScore {
			bestScore = score
			best = pid
			bestD = fmt.Sprintf("rtt %.0f vs %g ms, débit %.1f vs %.0f Mbit/s",
				r.RTTIdleP95, p.DelayMs, r.ThroughputMbps, p.CapacityMbps)
		}
	}
	if best == "" {
		return "", "aucun profil de référence avec capacité mesurée"
	}
	return best, bestD
}
