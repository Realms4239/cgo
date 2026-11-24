package metrics

import "math"

// SLO déclaré : disponibilité d'un rapport d'audit complet et part des
// observations tenant dans le plafond EMD déclaré. Ce sont des cibles, pas des
// mesures : aucun budget n'est annoncé vert sans observations réelles.
const (
	SLOAuditAvailability     = 0.99
	SLOFidelityWithinCeiling = 0.99
)

// ErrorBudget — burn rate et budget restant dérivés d'observations réelles.
// observedOK ne peut dépasser observedTotal et la cible doit être finie dans
// (0,1) ; sans observation ou avec une entrée invalide, refuse (ok=false) et
// ne fabrique aucun vert. Le budget restant peut être négatif : c'est un
// dépassement mesuré, pas une erreur d'arrondi.
func ErrorBudget(observedOK, observedTotal uint64, target float64) (burnRate, remaining float64, ok bool) {
	if observedTotal == 0 || observedOK > observedTotal ||
		math.IsNaN(target) || math.IsInf(target, 0) || target <= 0 || target >= 1 {
		return 0, 0, false
	}
	allowed := (1 - target) * float64(observedTotal)
	burnRate = float64(observedTotal-observedOK) / allowed
	remaining = 1 - burnRate
	return burnRate, remaining, true
}
