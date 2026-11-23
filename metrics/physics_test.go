package metrics

import (
	"math"
	"strings"
	"testing"
)

// TestModelDelayMsArithmetic — l'oracle exact du brief : 100 + |P|/B + Q.
// 1500 octets à 20 Mbit/s = 12000 bits / 20e6 = 0.6 ms.
func TestModelDelayMsArithmetic(t *testing.T) {
	got, err := ModelDelayMs(100, 1500, 20, 0)
	if err != nil {
		t.Fatalf("ModelDelayMs: %v", err)
	}
	want := 100 + (1500 * 8 / 20e6 * 1000) + 0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("ModelDelayMs = %.12f, voulu %.12f", got, want)
	}
}

// TestModelDelayMsIncludesSerializationAndQueue — la sérialisation dépend de
// la taille du paquet et la file s'ajoute, sans être confondue avec la mesure.
func TestModelDelayMsIncludesSerializationAndQueue(t *testing.T) {
	got, err := ModelDelayMs(40, 1500, 10, 5)
	if err != nil {
		t.Fatalf("ModelDelayMs: %v", err)
	}
	want := 40.0 + 1.2 + 5.0 // 1500*8/10e6*1000 = 1.2 ms
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("ModelDelayMs = %.12f, voulu %.12f", got, want)
	}
}

// Oracle hostile : un taux nul, négatif ou NaN ne doit jamais produire un
// délai (division par zéro silencieuse → +Inf ou NaN transformé en valeur).
func TestModelDelayMsRefusesNonPositiveRate(t *testing.T) {
	for _, rate := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		got, err := ModelDelayMs(100, 1500, rate, 0)
		if err == nil {
			t.Fatalf("taux %v : refus attendu, obtenu %.12f", rate, got)
		}
		if got != 0 {
			t.Fatalf("taux %v : appel refusé a renvoyé %v, voulu 0", rate, got)
		}
	}
}

// TestConfiguredVsRealisedAbsentIsDashWithNote — une absence n'est jamais un
// zéro : le tiret porte une note explicite, aucun nombre n'est fabriqué.
func TestConfiguredVsRealisedAbsentIsDashWithNote(t *testing.T) {
	row := ConfiguredVsRealised("delay", "", "")
	if row.Metric != "delay" {
		t.Fatalf("metric = %q, voulu delay", row.Metric)
	}
	if !strings.Contains(row.Configured, AbsentValue) || !strings.Contains(row.Configured, "configur") {
		t.Fatalf("configured %q doit contenir %q et une note d'absence", row.Configured, AbsentValue)
	}
	if !strings.Contains(row.Realised, AbsentValue) || !strings.Contains(row.Realised, "mesur") {
		t.Fatalf("realised %q doit contenir %q et une note d'absence", row.Realised, AbsentValue)
	}
}

// TestConfiguredVsRealisedKeepsActualValues — le sous-typage NetEm documenté
// s'affiche tel que mesuré, jamais en rescaleant la valeur configurée.
func TestConfiguredVsRealisedKeepsActualValues(t *testing.T) {
	row := ConfiguredVsRealised("delay", "100 ms", "84 ms")
	if row.Configured != "100 ms" || row.Realised != "84 ms" {
		t.Fatalf("ligne = %+v, valeurs configurées/réalisées altérées", row)
	}
}
