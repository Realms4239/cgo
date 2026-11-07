package audit

import "encoding/json"

// RemediationState — l'absence comme état nommé, pas une absence de champ.
// Un rapport scellé avant ce champ décode comme « observation_only », jamais
// « available » par zéro Go.
type RemediationState string

const (
	RemediationAvailableState   RemediationState = "available"
	RemediationObservationState RemediationState = "observation_only"
)

// UnmarshalJSON rend les rapports scellés avant la slice 22 « observation_only »
// au lieu d'un état vide. Un état inconnu est pass-through : le moteur n'a pas
// à rejeter un rapport qu'il ne comprend pas. L'absence du champ ne déclenche
// pas UnmarshalJSON (un champ string pointeur reste null) : Report.UnmarshalJSON
// applique le défaut quand le champ manque.
func (s *RemediationState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch RemediationState(str) {
	case RemediationAvailableState, RemediationObservationState:
		*s = RemediationState(str)
	case "":
		*s = RemediationObservationState
	default:
		*s = RemediationState(str)
	}
	return nil
}

// IsAvailable — l'alias booléen de l'état, pour l'UI qui veut un discriminant.
func (r *Report) IsAvailable() bool {
	return r != nil && r.RemediationState == RemediationAvailableState && r.Remediation != nil
}
