package metrics

import (
	"math"
	"strings"
	"testing"
)

func pfloat(v float64) *float64 { return &v }

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

// AssessLoadedHealth — même sémantique que l'ancien health.ts, désormais
// dérivée une seule fois ici.
func TestAssessLoadedHealth(t *testing.T) {
	cases := []struct {
		name     string
		linkType string
		idle     *float64
		loaded   *float64
		kind     string
		bloat    float64
		class    string
		red      *float64
	}{
		{"absence d'un côté", "fiber", nil, pfloat(61), "absent", 0, "", nil},
		{"absence de l'autre", "fiber", pfloat(20), nil, "absent", 0, "", nil},
		{"fiber critique", "fiber", pfloat(20), pfloat(61), "critical", 41, "fiber", pfloat(30)},
		{"fiber ok", "fiber", pfloat(20), pfloat(43), "ok", 23, "fiber", pfloat(30)},
		{"fiber attente au seuil 0.8", "fiber", pfloat(20), pfloat(44), "attention", 24, "fiber", pfloat(30)},
		{"fiber juste sous critical", "fiber", pfloat(20), pfloat(50), "attention", 30, "fiber", pfloat(30)},
		{"fiber épineux (30 pile, 0.8×30=24)", "fiber", pfloat(0), pfloat(30), "attention", 30, "fiber", pfloat(30)},
		{"fiber charge plus rapide que repos → clampé", "fiber", pfloat(50), pfloat(40), "ok", 0, "fiber", pfloat(30)},
		{"4g critique", "4g", pfloat(100), pfloat(201), "critical", 101, "cellular", pfloat(100)},
		{"5g même classe", "5g", pfloat(100), pfloat(201), "critical", 101, "cellular", pfloat(100)},
		{"vsat geo", "vsat", pfloat(600), pfloat(720), "attention", 120, "geo", pfloat(150)},
		{"leo sans seuil", "leo", pfloat(40), pfloat(240), "unconfigured", 200, "leo", nil},
		{"p4 = leo", "p4", pfloat(40), pfloat(240), "unconfigured", 200, "leo", nil},
		{"classe inconnue", "xdsl", pfloat(40), pfloat(240), "unconfigured", 200, "other", nil},
		{"casse ignorée", "FIBER", pfloat(20), pfloat(61), "critical", 41, "fiber", pfloat(30)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := AssessLoadedHealth(c.linkType, c.idle, c.loaded)
			if got.Kind != c.kind {
				t.Fatalf("kind = %q, voulu %q", got.Kind, c.kind)
			}
			if c.kind == "absent" {
				return // rien d'autre n'est rendu quand la mesure manque
			}
			if got.BloatMs != c.bloat {
				t.Fatalf("bloat_ms = %v, voulu %v", got.BloatMs, c.bloat)
			}
			if got.LinkClass != c.class {
				t.Fatalf("link_class = %q, voulu %q", got.LinkClass, c.class)
			}
			if c.red == nil {
				if got.RedMs != nil {
					t.Fatalf("red_ms = %v, voulu nil (unconfigured)", *got.RedMs)
				}
			} else if got.RedMs == nil || *got.RedMs != *c.red {
				t.Fatalf("red_ms = %v, voulu %v", got.RedMs, *c.red)
			}
		})
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

// fidelitySet — n paires mesuré/modèle dont l'écart absolu vaut deviation.
func fidelitySet(n int, deviation float64) ([]float64, []float64) {
	measured := make([]float64, n)
	modeled := make([]float64, n)
	for i := range measured {
		modeled[i] = 100
		measured[i] = 100 + deviation
	}
	return measured, modeled
}

// TestAssessFidelityTrust — trust exige au moins 30 samples ET un écart
// explicite <= epsilon. Le dépassement d'epsilon est discard, jamais un vert
// élargi.
func TestAssessFidelityTrust(t *testing.T) {
	measured, modeled := fidelitySet(30, 2)
	got := AssessFidelity(measured, modeled, 5)
	if got.State != FidelityTrust {
		t.Fatalf("state = %q, voulu trust (%+v)", got.State, got)
	}
	if got.Samples != 30 || math.Abs(got.DeviationMs-2) > 1e-9 || got.EpsilonMs != 5 {
		t.Fatalf("verdict = %+v, voulu samples=30 deviation=2 epsilon=5", got)
	}
}

func TestAssessFidelityDiscardOverEpsilon(t *testing.T) {
	measured, modeled := fidelitySet(30, 6)
	got := AssessFidelity(measured, modeled, 5)
	if got.State != FidelityDiscard {
		t.Fatalf("state = %q, voulu discard (%+v)", got.State, got)
	}
	if math.Abs(got.DeviationMs-6) > 1e-9 {
		t.Fatalf("deviation = %v, voulu 6", got.DeviationMs)
	}
}

// Oracle hostile : moins de 30 samples donne bounded, JAMAIS trust — même si
// l'écart est nul. C'est le piège « silence verdirait un manque de preuve ».
func TestAssessFidelityBoundedInsufficientSamples(t *testing.T) {
	measured, modeled := fidelitySet(29, 0)
	got := AssessFidelity(measured, modeled, 5)
	if got.State != FidelityBounded {
		t.Fatalf("state = %q, voulu bounded jamais trust (%+v)", got.State, got)
	}
	if got.Samples != 29 {
		t.Fatalf("samples = %d, voulu 29", got.Samples)
	}
}

// Aucune entrée ne doit produire trust sans borne de validité explicite :
// epsilon nul, négatif, NaN ou infini restent bounded.
func TestAssessFidelityNoTrustWithoutExplicitBound(t *testing.T) {
	measured, modeled := fidelitySet(30, 0)
	for _, eps := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		got := AssessFidelity(measured, modeled, eps)
		if got.State == FidelityTrust {
			t.Fatalf("epsilon %v a produit trust sans borne explicite (%+v)", eps, got)
		}
		if got.State != FidelityBounded {
			t.Fatalf("epsilon %v : state = %q, voulu bounded", eps, got.State)
		}
	}
}

// Oracle hostile : un sample NaN ne doit pas se cacher dans une comparaison
// (NaN > dev est faux) et laisser verdir le verdict.
func TestAssessFidelityNaNSampleDiscards(t *testing.T) {
	measured, modeled := fidelitySet(30, 0)
	measured[7] = math.NaN()
	got := AssessFidelity(measured, modeled, 5)
	if got.State != FidelityDiscard {
		t.Fatalf("state = %q, voulu discard pour sample NaN (%+v)", got.State, got)
	}
}

// Des longueurs divergentes ne peuvent pas être comparées : bounded, jamais
// trust. Une troncature silencieuse fabriquerait une preuve.
func TestAssessFidelityMismatchedLengthsBounded(t *testing.T) {
	measured, modeled := fidelitySet(30, 0)
	got := AssessFidelity(measured, modeled[:29], 5)
	if got.State == FidelityTrust || got.State != FidelityBounded {
		t.Fatalf("state = %q, voulu bounded (%+v)", got.State, got)
	}
}

// TestAssessFidelityStateClosedSet — le state appartient toujours à l'ensemble
// fermé trust|bounded|discard, aucune valeur comparables à « green » n'ajoutée.
func TestAssessFidelityStateClosedSet(t *testing.T) {
	trustM, trustMo := fidelitySet(30, 0)
	boundedM, boundedMo := fidelitySet(29, 0)
	discardM, discardMo := fidelitySet(30, 99)
	for _, v := range []FidelityVerdict{
		AssessFidelity(trustM, trustMo, 5),
		AssessFidelity(boundedM, boundedMo, 5),
		AssessFidelity(discardM, discardMo, 5),
	} {
		switch v.State {
		case FidelityTrust, FidelityBounded, FidelityDiscard:
		default:
			t.Fatalf("state hors ensemble fermé : %q", v.State)
		}
	}
}
