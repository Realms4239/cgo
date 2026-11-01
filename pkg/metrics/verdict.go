package metrics

import "fmt"

// VerdictLevel — les quatre niveaux du langage opérateur. Le rouge garde un
// seul niveau et deux mots, selon ce qu'il a mesuré : une mission manquée et
// une saturation sont deux urgences différentes, pas deux couleurs.
type VerdictLevel string

const (
	VerdictNeutral VerdictLevel = "neutral"
	VerdictRed     VerdictLevel = "red"
	VerdictAmber   VerdictLevel = "amber"
	VerdictGreen   VerdictLevel = "green"
)

// VerdictInput — les seuls faits que la cascade lit. Rien n'est recalculé ici :
// la note vient de BufferbloatGrade, la conformité de AssessCAPSLA.
type VerdictInput struct {
	LoadAttempted   bool
	LoadEstablished bool
	Grade           string
	SLA             CAPSLA
}

type Verdict struct {
	Level    VerdictLevel `json:"level"`
	Word     string       `json:"word"`
	Sentence string       `json:"sentence"`
}

// AssessVerdict — cascade à premier gagnant, bornée par les seuils déjà
// défendus par le moteur. Un vert sans note est impossible : la charge peut
// avoir fait passer des sondes sans jamais être établie.
func AssessVerdict(in VerdictInput) Verdict {
	switch {
	case !in.LoadAttempted:
		return Verdict{VerdictNeutral, "Pas encore mesuré",
			"Aucun test sous charge n'a été réalisé sur ce lien."}
	case !in.LoadEstablished:
		return Verdict{VerdictNeutral, "Charge non établie",
			"Le flux saturant n'a pas pu démarrer ; la tenue sous congestion n'est pas mesurée."}
	case in.SLA.Measured && in.SLA.Breach:
		return Verdict{VerdictRed, "Alertes en retard",
			fmt.Sprintf("%.1f %% des alertes passent sous %d ms (objectif %.0f %%).",
				in.SLA.OKPct, CAPDeadlineMs, CAPBreachThresholdPct)}
	case in.Grade == "":
		return Verdict{VerdictNeutral, "Note indisponible",
			"La charge a tourné mais l'écart de latence n'a pas pu être calculé."}
	case in.Grade == "D" || in.Grade == "F":
		return Verdict{VerdictRed, "Saturation critique",
			fmt.Sprintf("Le lien sature sous charge (note %s) ; les télétransmissions risquent d'échouer.", in.Grade)}
	case in.Grade == "C":
		return Verdict{VerdictAmber, "Lien fragile",
			"La latence monte nettement sous charge ; les alertes peuvent partir en retard."}
	case (in.Grade == "A+" || in.Grade == "A" || in.Grade == "B") && in.SLA.Measured:
		return Verdict{VerdictGreen, "Lien sain",
			fmt.Sprintf("Sous charge, la latence reste sous l'objectif et %.1f %% des alertes passent à l'heure.", in.SLA.OKPct)}
	case in.Grade == "A+" || in.Grade == "A" || in.Grade == "B":
		return Verdict{VerdictGreen, "Lien sain",
			"Sous charge, la latence reste sous l'objectif."}
	default:
		// Vocabulaire de note inconnu : rien à classer. Jamais un vert par
		// défaut — la porte 7 de la cascade n'accepte que A+, A et B.
		return Verdict{VerdictNeutral, "Note indisponible",
			"La charge a tourné mais l'écart de latence n'a pas pu être calculé."}
	}
}
