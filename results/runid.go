package results

import (
	"errors"
	"unicode/utf8"
)

// ErrInvalidRunID — un identifiant de run n'est pas un élément de chemin
// simple : séparateur de chemin, point, caractère de contrôle, ou trop long.
var ErrInvalidRunID = errors.New("identifiant de run invalide : séparateur, point, contrôle ou longueur")

const maxRunIDLen = 128

// ValidRunID — la seule règle qui définit « un identifiant de run » dans
// tout le système (HTTP, CLI, banc, reprises).
//
// Un identifiant de run est un nom de dossier unique directement sous le
// corpus : aucun séparateur de chemin (/ ni \), aucun point (pas de
// « .. », pas de suffixe qui traverse le glob), aucun caractère de contrôle,
// au plus 128 octets. La chaîne vide est invalide ici — les appelants qui
// acceptent « vide = tous les runs » ou « vide = génération automatique »
// le gèrent eux-mêmes.
func ValidRunID(id string) bool {
	return validateRunID(id) == nil
}

func validateRunID(id string) error {
	if id == "" {
		return ErrInvalidRunID
	}
	if len(id) > maxRunIDLen {
		return ErrInvalidRunID
	}
	if !utf8.ValidString(id) {
		return ErrInvalidRunID
	}
	for _, r := range id {
		if r == '/' || r == '\\' || r == '.' {
			return ErrInvalidRunID
		}
		// C0 controls rejetés; 0x7f-0x9f aussi (SPOF visuel dans les journaux)
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return ErrInvalidRunID
		}
	}
	return nil
}

// ValidateRunID — la même règle en forme d'erreur, pour les appelants qui
// doivent refuser la saisie (CLI) plutôt que raboter en silence.
func ValidateRunID(id string) error {
	return validateRunID(id)
}
