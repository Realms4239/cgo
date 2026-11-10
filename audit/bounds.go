package audit

// Les bornes d'un audit : UNE règle, UN propriétaire.
//
// Avant, la borne de durée (10–600 s) n'existait que dans le handler HTTP ;
// le CLI passait sa valeur telle quelle au moteur — `cgo audit --duration
// 999999` lancait un audit de 27 heures. Le nombre de flux (1–8) était
// borné chez trois appelants, jamais dans le moteur.
//
// Le moteur applique toujours les bornes (sécurité : une mesure ne dure pas
// une journée) ; l'appelant peut les interroger pour refuser une saisie
// hors bornes au lieu de la raboter en silence.

const (
	// MinDurationSec / MaxDurationSec — la fenêtre d'audit documentée.
	MinDurationSec = 10
	MaxDurationSec = 600
	// MinFlows / MaxFlows — flux de charge concurrents (régime du banc durci).
	MinFlows = 1
	MaxFlows = 8
	// DefaultFlows — le régime par défaut quand l'appelant ne dit rien.
	DefaultFlows = 4
)

// ClampDuration — le garde-fou du MOTEUR : jamais plus de MaxDurationSec,
// et une durée nulle prend le défaut. Les audits courts restent possibles
// (mesures de contrôle, tests) : la fenêtre 10–600 s est une règle de
// surface opérateur, pas du moteur.
func ClampDuration(sec int) int {
	switch {
	case sec <= 0:
		return MinDurationSec
	case sec > MaxDurationSec:
		return MaxDurationSec
	default:
		return sec
	}
}

// ClampWindow — la fenêtre documentée d'une SAISIE D'OPÉRATEUR (10–600 s).
// L'API rabote ; le CLI refuse (DurationInBounds) plutôt que de raboter en
// silence. Le moteur, lui, s'en tient à ClampDuration.
func ClampWindow(sec int) int {
	switch {
	case sec < MinDurationSec:
		return MinDurationSec
	case sec > MaxDurationSec:
		return MaxDurationSec
	default:
		return sec
	}
}

// ClampFlows — ramène un nombre de flux dans les bornes documentées.
// 0 ou moins ⇒ le régime par défaut.
func ClampFlows(n int) int {
	switch {
	case n <= 0:
		return DefaultFlows
	case n < MinFlows:
		return MinFlows
	case n > MaxFlows:
		return MaxFlows
	default:
		return n
	}
}

// DurationInBounds — la durée demandée est-elle acceptable telle quelle ?
// L'appelant qui veut REFUSER une saisie (CLI) pose la question ici.
func DurationInBounds(sec int) bool {
	return sec >= MinDurationSec && sec <= MaxDurationSec
}

// FlowsInBounds — le nombre de flux demandé est-il acceptable tel quel ?
func FlowsInBounds(n int) bool {
	return n >= MinFlows && n <= MaxFlows
}
