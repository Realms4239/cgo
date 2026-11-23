package audit

import (
	"fmt"

	"github.com/Realms4239/cgo/model"
)

// ConfiguredRealisedRow — une ligne du tableau configuré vs réalisé. Le type
// vit dans audit pour ne pas ajouter de dépendance inter-modules : metrics est
// importé par audit, l'inverse créerait un cycle.
type ConfiguredRealisedRow struct {
	Metric     string `json:"metric"`
	Configured string `json:"configured"`
	Realised   string `json:"realised"`
}

const cvrAbsent = "—"

func cvrAbsentNote(kind string) string {
	return cvrAbsent + " (non " + kind + ")"
}

// cvrMeasured — un zéro de débit ou de délai signifie « pas de mesure », pas
// « mesure nulle » : on rend le tiret et une note, jamais un zéro fabriqué.
func cvrMeasured(v float64, qualifier string) string {
	if v == 0 {
		return cvrAbsentNote("mesuré")
	}
	return fmt.Sprintf("%g", v) + " (" + qualifier + ")"
}

// configuredVsRealised — quatre lignes seulement quand un profil de référence
// a réellement été apparié (ProfileMatch est gelé dans le résultat). Sans cet
// appariement, aucune valeur configurée n'existe : la tranche reste vide et
// omitempty garde le JSON du rapport octet pour octet identique.
//
// La colonne configurée porte l'identifiant du profil : ce sont les valeurs du
// référentiel P1–P4, pas un réglage tc relu sur l'hôte. La colonne réalisée
// porte la mesure effective avec son origine, y compris le sous-typage de
// gigue NetEm, affiché tel quel plutôt que rescale vers la cible.
func configuredVsRealised(res *Result) []ConfiguredRealisedRow {
	if res == nil || res.ProfileMatch == "" {
		return nil
	}
	model.ProfilesMu.RLock()
	p, ok := model.Profiles[res.ProfileMatch]
	model.ProfilesMu.RUnlock()
	if !ok {
		return nil
	}
	tag := "profil " + res.ProfileMatch
	jitter := cvrAbsentNote("mesuré")
	if res.RTTIdleP95 > 0 && res.RTTIdleP50 > 0 {
		jitter = fmt.Sprintf("%g", res.RTTIdleP95-res.RTTIdleP50) + " (idle p95−p50)"
	}
	return []ConfiguredRealisedRow{
		{
			Metric:     "delay — RTT idle p95",
			Configured: fmt.Sprintf("%g ms (%s)", p.DelayMs, tag),
			Realised:   cvrMeasured(res.RTTIdleP95, "idle p95"),
		},
		{
			Metric:     "jitter — écart idle p95−p50",
			Configured: fmt.Sprintf("%g ms (%s)", p.JitterMs, tag),
			Realised:   jitter,
		},
		{
			Metric:     "loss — sonde",
			Configured: fmt.Sprintf("%g %% (%s)", p.LossPct, tag),
			Realised:   fmt.Sprintf("%g %% (sonde)", res.LossPct),
		},
		{
			Metric:     "rate — goodput",
			Configured: fmt.Sprintf("%g Mbit/s (%s)", p.CapacityMbps, tag),
			Realised:   cvrMeasured(res.ThroughputMbps, "goodput"),
		},
	}
}
