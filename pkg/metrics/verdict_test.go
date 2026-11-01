package metrics

import (
	"strings"
	"testing"
)

func sla(pct float64, measured bool) CAPSLA {
	return CAPSLA{Measured: measured, OKPct: pct, Breach: measured && pct < CAPBreachThresholdPct}
}

func TestAssessVerdictCascade(t *testing.T) {
	cases := []struct {
		name  string
		in    VerdictInput
		level VerdictLevel
		word  string
	}{
		{"jamais testé", VerdictInput{}, VerdictNeutral, "Pas encore mesuré"},
		{"charge non établie", VerdictInput{LoadAttempted: true}, VerdictNeutral, "Charge non établie"},
		{"alertes en retard", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "B", SLA: sla(60.2, true)}, VerdictRed, "Alertes en retard"},
		{"failure prime sur note absente", VerdictInput{LoadAttempted: true, LoadEstablished: true, SLA: sla(80, true)}, VerdictRed, "Alertes en retard"},
		{"note absente", VerdictInput{LoadAttempted: true, LoadEstablished: true}, VerdictNeutral, "Note indisponible"},
		{"saturation sans sonde alerte", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "F"}, VerdictRed, "Saturation critique"},
		{"fragile", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "C"}, VerdictAmber, "Lien fragile"},
		{"sain", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "A", SLA: sla(97.5, true)}, VerdictGreen, "Lien sain"},
		{"sain sans sonde alerte", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "B"}, VerdictGreen, "Lien sain"},
		{"note hors référentiel", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "b"}, VerdictNeutral, "Note indisponible"},
		{"note hors vocab avec sonde", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "b", SLA: sla(99.9, true)}, VerdictNeutral, "Note indisponible"},
		{"charge tentée prime sur manquement", VerdictInput{LoadAttempted: true, LoadEstablished: false, SLA: sla(60.2, true)}, VerdictNeutral, "Charge non établie"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := AssessVerdict(c.in)
			if got.Level != c.level || got.Word != c.word {
				t.Fatalf("verdict = %q/%q, voulu %q/%q", got.Level, got.Word, c.level, c.word)
			}
			if got.Sentence == "" {
				t.Fatal("phrase vide")
			}
		})
	}
}

// Le piège signalé en revue : une sonde d'alerte qui passe alors qu'aucune
// charge n'a été établie ne doit jamais produire de vert.
func TestVerdictNeverGreenWithoutGrade(t *testing.T) {
	got := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, SLA: sla(100, true)})
	if got.Level == VerdictGreen {
		t.Fatalf("vert sans note : %+v", got)
	}
	if got.Level != VerdictNeutral {
		t.Fatalf("niveau = %q, voulu neutral", got.Level)
	}
}

func TestVerdictSentencesCarryTheirNumbers(t *testing.T) {
	red := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "D", SLA: sla(60.2, true)})
	if !strings.Contains(red.Sentence, "60.2 %") {
		t.Fatalf("phrase rouge %q doit citer le pourcentage mesuré", red.Sentence)
	}
	crit := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "F"})
	if !strings.Contains(crit.Sentence, "note F") {
		t.Fatalf("phrase saturation %q doit nommer la note", crit.Sentence)
	}
}
