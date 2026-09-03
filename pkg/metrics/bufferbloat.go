package metrics

import "math"

// BufferbloatGrade — note A+..F d'un écart de latence sous charge (ms).
// Bandes Waveform verbatim (waveform.com/tools/bufferbloat, consulté
// 2026-09-03) : A+ <5 · A <30 · B <60 · C <200 · D <400 · F ≥400. La note
// porte sur l'écart, pas la latence absolue — agnostique au délai de base,
// donc applicable du VSAT à la fibre. Un écart négatif (lien plus rapide
// chargé, artefact de mesure) est clampé à A+.
func BufferbloatGrade(deltaMs float64) string {
	if math.IsNaN(deltaMs) {
		return "" // indisponible honnête
	}
	switch {
	case deltaMs < 5:
		return "A+"
	case deltaMs < 30:
		return "A"
	case deltaMs < 60:
		return "B"
	case deltaMs < 200:
		return "C"
	case deltaMs < 400:
		return "D"
	default:
		return "F"
	}
}

// BufferbloatWorst — le pire écart mesuré, download et upload confondus
// (méthode Waveform : whichever test creates the larger increase).
// NaN = sens non mesuré (pas encore de campagne download) : l'autre décide
// seul ; les deux NaN rendent NaN (note indisponible, jamais fabriquée).
func BufferbloatWorst(downDelta, upDelta float64) float64 {
	if math.IsNaN(downDelta) {
		return upDelta
	}
	if math.IsNaN(upDelta) {
		return downDelta
	}
	return math.Max(downDelta, upDelta)
}

// BufferbloatVerdict — chaque note porte son remède en langage opérateur.
// Vide pour une note inconnue.
func BufferbloatVerdict(grade string) string {
	switch grade {
	case "A+":
		return "Latence excellente sous charge : l'AQM tient les files, appel et pilotage restent nets."
	case "A":
		return "Latence bonne sous charge : tous usages passent, y compris voix interactive et jeu."
	case "B":
		return "Bufferbloat léger : gênant en jeu rapide et appels pendant un gros transfert."
	case "C":
		return "Bufferbloat modéré : appels et jeu se dégradent dès que le lien travaille."
	case "D":
		return "Bufferbloat sévère : tout temps réel devient difficile pendant la charge."
	case "F":
		return "Bufferbloat critique : le lien cesse de répondre dès qu'on l'utilise."
	}
	return ""
}
