package metrics

import (
	"fmt"
	"math"
)

// ModelDelayMs — délai modèle d'un paquet : propagation + sérialisation + file.
// C'est une valeur CALCULÉE à partir des intrants du profil, jamais une mesure.
// B = rateMbps * 1e6 bits/s ; |P|/B est en secondes, converti en ms (×1000).
// Un taux absent, nul, négatif ou non fini est refusé : le diviser produirait
// un délai infini ou NaN qui se lirait comme une mesure valide.
func ModelDelayMs(propagationMs, packetBytes, rateMbps, queueMs float64) (float64, error) {
	if math.IsNaN(rateMbps) || math.IsInf(rateMbps, 0) || rateMbps <= 0 {
		return 0, fmt.Errorf("rate must be > 0")
	}
	return propagationMs + packetBytes*8/(rateMbps*1e6)*1000 + queueMs, nil
}

// EffectiveDeadlineMs — échéance relative au profil : délai modèle plus marge.
// Délègue à ModelDelayMs pour hériter de son refus des taux invalides ; c'est
// une valeur calculée, jamais une mesure ni le seuil fixe de 220 ms.
func EffectiveDeadlineMs(profileDelayMs, packetBytes, rateMbps, allowanceMs float64) (float64, error) {
	return ModelDelayMs(profileDelayMs, packetBytes, rateMbps, allowanceMs)
}

// AbsentValue — la cellule rendue quand une valeur manque dans le rapport.
// Jamais un zéro : zéro se lirait comme une mesure réelle.
const AbsentValue = "—"

// ConfiguredRealisedRow — une ligne du tableau configuré vs réalisé. Les deux
// colonnes sont des chaînes déjà formatées par l'appelant à partir de valeurs
// RÉELLES du rapport ; cette couche ne fabrique et ne rescale aucun nombre.
type ConfiguredRealisedRow struct {
	Metric     string `json:"metric"`
	Configured string `json:"configured"`
	Realised   string `json:"realised"`
}

// ConfiguredVsRealised — construit une ligne. Une valeur absente (chaîne vide)
// devient AbsentValue accompagné d'une note explicite ; un sous-typage NetEm
// documenté reste affiché tel que mesuré, sans réécriture de la cible.
func ConfiguredVsRealised(metric, configured, realised string) ConfiguredRealisedRow {
	return ConfiguredRealisedRow{
		Metric:     metric,
		Configured: presentOrAbsent(configured, "configuré"),
		Realised:   presentOrAbsent(realised, "mesuré"),
	}
}

func presentOrAbsent(v, kind string) string {
	if v == "" {
		return AbsentValue + " (non " + kind + ")"
	}
	return v
}
