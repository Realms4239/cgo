package metrics

import (
	"fmt"
	"math"
)

// Seuils de fidélité, figés dans le code : déclarés dans chaque rapport, jamais
// élargis pour verdir un verdict. FidelityMinSamples et les états vivent dans
// verdict.go et sont réutilisés ici.
const (
	FidelityEMDMax         = 0.10 // EMD normalisé <= 10 % du délai live moyen
	FidelityGammaMax       = 0.05 // au plus 5 % de paires hors borne
	FidelityEpsilonFracMax = 0.05 // epsilon <= 5 % du délai modèle
)

// DistributionEMD — distance de Wasserstein 1-D entre les CDF empiriques de
// live et emulated, divisée par la moyenne de live (grandeur sans unité).
// Les séries sont triées : l'ordre d'arrivée des paquets ne change rien.
// Un jeu vide, une moyenne non positive ou un échantillon non fini est refusé :
// un NaN se lirait comme une mesure valide.
func DistributionEMD(live, emulated []float64) (float64, error) {
	if len(live) == 0 || len(emulated) == 0 {
		return 0, fmt.Errorf("EMD: empty sample set")
	}
	for _, v := range live {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("EMD: non-finite live sample")
		}
	}
	for _, v := range emulated {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("EMD: non-finite emulated sample")
		}
	}

	a, b := sortedCopy(live), sortedCopy(emulated)
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		d := a[i] - b[i]
		if d < 0 {
			d = -d
		}
		sum += d
	}

	meanLive := 0.0
	for _, v := range a {
		meanLive += v
	}
	meanLive /= float64(len(a))
	if meanLive <= 0 || math.IsNaN(meanLive) || math.IsInf(meanLive, 0) {
		return 0, fmt.Errorf("EMD: invalid live mean")
	}
	return sum / float64(n) / meanLive, nil
}

// DelayAccuracy — critère (γ, ε) « probablement approximativement exact » sur
// des paires de paquets (RTD mesuré contre le délai modèle δ = l/v + |P|/B +
// |Q|/B). epsilonMs est le quantile (1−γ) des écarts |RTD−δ| ; gamma est la
// fraction des paires hors de la borne εFrac×δ. Les quantiles remplacent MAE/
// MPAE : une moyenne masquerait les queues.
//
// trust exige au moins FidelityMinSamples paires finies, gamma <= gammaMax et
// epsilon <= epsilonFrac×δ. Les plafonds FidelityGammaMax/FidelityEpsilonFracMax
// sont appliqués avant tout calcul : un appelant ne peut pas élargir un seuil.
// Un écart contraire ou un seuil élargi donne discard, pas de preuves suffisantes
// donne bounded ; jamais trust par défaut.
func DelayAccuracy(samples []float64, modelMs, gammaMax, epsilonFrac float64) (state string, epsilonMs float64, ok bool) {
	if math.IsNaN(modelMs) || math.IsInf(modelMs, 0) || modelMs <= 0 {
		return FidelityBounded, 0, false
	}
	if math.IsNaN(gammaMax) || math.IsInf(gammaMax, 0) || gammaMax <= 0 || gammaMax > FidelityGammaMax ||
		math.IsNaN(epsilonFrac) || math.IsInf(epsilonFrac, 0) || epsilonFrac <= 0 || epsilonFrac > FidelityEpsilonFracMax {
		return FidelityDiscard, 0, false // seuil invalide ou élargi : refus net
	}
	if len(samples) == 0 {
		return FidelityBounded, 0, false
	}

	dev := make([]float64, len(samples))
	bound := epsilonFrac * modelMs
	outside := 0
	for i, v := range samples {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return FidelityDiscard, 0, false
		}
		d := v - modelMs
		if d < 0 {
			d = -d
		}
		dev[i] = d
		if d > bound {
			outside++
		}
	}

	epsilonMs = quantileSorted(sortedCopy(dev), 1-gammaMax)
	gamma := float64(outside) / float64(len(samples))
	switch {
	case gamma > gammaMax || epsilonMs > bound:
		// Preuve contraire : le manque de paires ne masque jamais un écart.
		return FidelityDiscard, epsilonMs, false
	case len(samples) < FidelityMinSamples:
		return FidelityBounded, epsilonMs, false
	default:
		return FidelityTrust, epsilonMs, true
	}
}
