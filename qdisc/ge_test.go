package qdisc

import (
	"math"
	"strings"
	"testing"
)

func TestGEMeanLossMatchesPaperModel(t *testing.T) {
	// Exemple travaillé de l'article des implémenteurs (Carry/Briscoe,
	// netem) : p=0.01, r=0.19, 1-h=0.81, 1-k=0.01 → 5 % de perte moyenne
	// en salves. La formule : π_mauvais = p/(p+r) ; perte = π·(1-h) + (1-π)·(1-k).
	got := GEMeanLoss(GEModel{GoodToBad: 0.01, BadToGood: 0.19, LossBad: 0.81, LossGood: 0.01})
	if got < 4.9 || got > 5.1 {
		t.Fatalf("exemple article : %v %% (attendu ≈5)", got)
	}
	// Paramètres P2 du banc durci : ≈0.544 %.
	p2 := GEMeanLoss(GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005})
	if p2 < 0.5 || p2 > 0.6 {
		t.Fatalf("P2 : %v %% (attendu ≈0.544)", p2)
	}
	// P3 : ≈1.08 %.
	p3 := GEMeanLoss(GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001})
	if p3 < 1.0 || p3 > 1.15 {
		t.Fatalf("P3 : %v %% (attendu ≈1.08)", p3)
	}
	if GEMeanLoss(GEModel{}) != 0 {
		t.Fatal("modèle vide : 0, pas une division par zéro")
	}
}

func TestApplyNetemGEFourArguments(t *testing.T) {
	// LA correction centrale : la forme documentée est QUATRE arguments
	// « p r 1-h 1-k » — l'ancienne forme à cinq (p r h 1-r k) est invalide
	// sur Linux et disparaît.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	if err := ApplyNetemGE(f, "w-edge", 100, 15, ge); err != nil {
		t.Fatal(err)
	}
	call := f.Calls[0]
	idx := -1
	for i, a := range call {
		if a == "gemodel" {
			idx = i
			break
		}
	}
	if idx < 0 || len(call) != idx+5 {
		t.Fatalf("gemodel doit recevoir EXACTEMENT 4 args : %v", call)
	}
	want := []string{"0.1%", "10%", "50%", "0.05%"}
	for i, w := range want {
		if call[idx+1+i] != w {
			t.Fatalf("arg %d = %q, want %q (ordre p, r, 1-h, 1-k)", i+1, call[idx+1+i], w)
		}
	}
}

func TestApplyNetemPhysicsUniformUnchanged(t *testing.T) {
	// ge nil ⇒ la forme uniforme historique, à l'octet près — la
	// reproduction du corpus gelé en dépend.
	f := &FakeRunner{}
	if err := ApplyNetemPhysics(f, "w-edge", 100, 15, 0.5, nil); err != nil {
		t.Fatal(err)
	}
	if joined := strings.Join(f.Calls[0], " "); !strings.Contains(joined, "loss 0.5%") {
		t.Fatalf("forme uniforme attendue : %v", f.Calls[0])
	}
}

func TestApplyNetemPhysicsBursty(t *testing.T) {
	// ge fourni ⇒ gemodel 4 args, PAS de « loss X% » uniforme en plus.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001}
	if err := ApplyNetemPhysics(f, "c-wan", 600, 30, 1, &ge); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Calls[0], " ")
	if !strings.Contains(joined, "loss gemodel 0.2% 10% 50% 0.1%") {
		t.Fatalf("gemodel 4 args attendu : %v", f.Calls[0])
	}
	if strings.Contains(joined, "loss 1%") {
		t.Fatalf("perte uniforme ne doit pas accompagner gemodel : %v", f.Calls[0])
	}
}

func TestApplyNetemJitterCorrelationFailClosed(t *testing.T) {
	// Corrélation hors plage ou NaN : refus AVANT tout appel tc — la
	// physique d'une cellule ne doit jamais rester à moitié configurée.
	for _, c := range []float64{-0.01, 1.01, math.NaN()} {
		f := &FakeRunner{}
		if err := ApplyNetemPhysics(f, "w-edge", 100, 15, 0, nil, c); err == nil {
			t.Fatalf("corrélation %v doit être refusée", c)
		}
		if len(f.Calls) != 0 {
			t.Fatalf("corrélation %v : tc appelé avant le refus : %v", c, f.Calls)
		}
	}
}

func TestApplyNetemJitterCorrelationVector(t *testing.T) {
	// c=0 : vecteur historique à l'octet près — le corpus gelé se reproduit.
	f0 := &FakeRunner{}
	if err := ApplyNetemPhysics(f0, "w-edge", 100, 15, 0, nil, 0); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f0.Calls[0], " "), "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms"; got != want {
		t.Fatalf("c=0 : got %q, want %q", got, want)
	}
	// c=0.25 : troisième jeton juste après la gigue.
	f25 := &FakeRunner{}
	if err := ApplyNetemPhysics(f25, "w-edge", 100, 15, 0, nil, 0.25); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f25.Calls[0], " "), "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms 25%"; got != want {
		t.Fatalf("c=0.25 : got %q, want %q", got, want)
	}
	// Avec distribution : la corrélation précède « distribution » (ordre tc).
	fd := &FakeRunner{}
	if err := ApplyNetemDistribution(fd, "w-edge", 100, 15, 0, nil, "P2", 0.25); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(fd.Calls[0], " "), "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms 25% distribution P2"; got != want {
		t.Fatalf("c=0.25 + dist : got %q, want %q", got, want)
	}
}

func TestApplyNetemJitterCorrelationWithGEModelOrder(t *testing.T) {
	// Ordre tc verrouillé : delay jitter corrélation loss gemodel distribution.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	if err := ApplyNetemDistribution(f, "c-wan", 600, 30, 0, &ge, "P3", 0.25); err != nil {
		t.Fatal(err)
	}
	got, want := strings.Join(f.Calls[0], " "),
		"qdisc replace dev c-wan root handle 1: netem delay 600ms 30ms 25% loss gemodel 0.1% 10% 50% 0.05% distribution P3"
	if got != want {
		t.Fatalf("ordre des jetons : got %q, want %q", got, want)
	}
}

// TestApplyNetemSerialZeroByteIdentical — taux 0 (défaut) : AUCUN jeton
// « rate », vecteurs historiques à l'octet près, y compris corrélation,
// perte (uniforme et gemodel) et distribution. Le corpus gelé se reproduit.
func TestApplyNetemSerialZeroByteIdentical(t *testing.T) {
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	cases := []struct {
		name string
		run  func(*FakeRunner) error
		want string
	}{
		{"nu", func(f *FakeRunner) error {
			return ApplyNetemPhysics(f, "w-edge", 100, 15, 0, nil, 0, 0, 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms"},
		{"corrélation", func(f *FakeRunner) error {
			return ApplyNetemPhysics(f, "w-edge", 100, 15, 0, nil, 0.25, 0, 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms 25%"},
		{"perte uniforme", func(f *FakeRunner) error {
			return ApplyNetemPhysics(f, "w-edge", 100, 15, 0.5, nil, 0, 0, 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms loss 0.5%"},
		{"gemodel + distribution", func(f *FakeRunner) error {
			return ApplyNetemDistribution(f, "c-wan", 600, 30, 0, &ge, "P3", 0.25, 0, 0)
		}, "qdisc replace dev c-wan root handle 1: netem delay 600ms 30ms 25% loss gemodel 0.1% 10% 50% 0.05% distribution P3"},
	}
	for _, tc := range cases {
		f := &FakeRunner{}
		if err := tc.run(f); err != nil {
			t.Fatalf("%s : %v", tc.name, err)
		}
		if got := strings.Join(f.Calls[0], " "); got != tc.want {
			t.Fatalf("%s : got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestApplyNetemSerialRateVector — opt-in : « rate Rmbit [overhead] » après
// la perte et AVANT la distribution (ordre documenté tc-netem).
func TestApplyNetemSerialRateVector(t *testing.T) {
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	f := &FakeRunner{}
	if err := ApplyNetemDistribution(f, "c-wan", 600, 30, 0, &ge, "P3", 0.25, 40, 38); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f.Calls[0], " "),
		"qdisc replace dev c-wan root handle 1: netem delay 600ms 30ms 25% loss gemodel 0.1% 10% 50% 0.05% rate 40mbit 38 distribution P3"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// Sans distribution ni perte : rate suit la corrélation.
	f2 := &FakeRunner{}
	if err := ApplyNetemPhysics(f2, "w-edge", 100, 15, 0, nil, 0, 40, 38); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f2.Calls[0], " "),
		"qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms rate 40mbit 38"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// Taux sans overhead : aucun jeton d'octets orphelin.
	f3 := &FakeRunner{}
	if err := ApplyNetemPhysics(f3, "w-edge", 100, 15, 0, nil, 0, 40, 0); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f3.Calls[0], " "),
		"qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms rate 40mbit"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestApplyNetemSerialFailClosed — taux négatif/NaN, overhead négatif ou
// overhead sans taux : refus AVANT tout appel tc (jamais de racine netem à
// moitié configurée, jamais de valeur coercée).
func TestApplyNetemSerialFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		rate     float64
		overhead int
	}{
		{"taux négatif", -0.1, 0},
		{"overhead négatif", 40, -1},
		{"taux négatif + overhead", -1, 38},
		{"taux NaN", math.NaN(), 0},
		{"overhead sans taux", 0, 38},
	}
	for _, tc := range cases {
		f := &FakeRunner{}
		if err := ApplyNetemPhysics(f, "w-edge", 100, 15, 0, nil, 0, tc.rate, float64(tc.overhead)); err == nil {
			t.Fatalf("%s doit être refusé", tc.name)
		}
		if len(f.Calls) != 0 {
			t.Fatalf("%s : tc appelé avant le refus : %v", tc.name, f.Calls)
		}
		// Même garde sur la voie distribution.
		fd := &FakeRunner{}
		if err := ApplyNetemDistribution(fd, "w-edge", 100, 15, 0, nil, "P2", 0, tc.rate, float64(tc.overhead)); err == nil {
			t.Fatalf("%s (dist) doit être refusé", tc.name)
		}
		if len(fd.Calls) != 0 {
			t.Fatalf("%s (dist) : tc appelé avant le refus : %v", tc.name, fd.Calls)
		}
	}
}

// TestApplyNetemProfileAbsentByteIdentical — champs de distribution/perte
// absents : le vecteur historique est reproduit à l'octet près, y compris
// la perte uniforme et la sérialisation. Le gel du corpus en dépend.
func TestApplyNetemProfileAbsentByteIdentical(t *testing.T) {
	cases := []struct {
		name string
		run  func(*FakeRunner) error
		want string
	}{
		{"nu", func(f *FakeRunner) error {
			return ApplyNetemProfile(f, "w-edge", 100, 15, 0, nil, "", 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms"},
		{"uniforme", func(f *FakeRunner) error {
			return ApplyNetemProfile(f, "w-edge", 100, 15, 0.5, nil, "", 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms loss 0.5%"},
		{"gemodel", func(f *FakeRunner) error {
			return ApplyNetemProfile(f, "w-edge", 100, 15, 0, &GEModel{GoodToBad: 0.01, BadToGood: 0.19, LossBad: 0.81, LossGood: 0.01}, "", 0)
		}, "qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms loss gemodel 1% 19% 81% 1%"},
	}
	for _, tc := range cases {
		f := &FakeRunner{}
		if err := tc.run(f); err != nil {
			t.Fatalf("%s : %v", tc.name, err)
		}
		if got := strings.Join(f.Calls[0], " "); got != tc.want {
			t.Fatalf("%s : got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestApplyNetemProfileDistributionShape — la distribution n'est émise
// que si le profil la déclare ; la forme suit le type : millisecondes pour
// normal/paretonormal, indice sans unité pour pareto, absente pour uniform.
func TestApplyNetemProfileDistributionShape(t *testing.T) {
	cases := []struct {
		name  string
		dist  string
		shape float64
		want  string
	}{
		{"normal", "normal", 8, "distribution normal 8ms"},
		{"pareto", "pareto", 1.5, "distribution pareto 1.5"},
		{"paretonormal", "paretonormal", 2, "distribution paretonormal 2"},
		{"uniform", "uniform", 0, "distribution uniform"},
	}
	for _, tc := range cases {
		f := &FakeRunner{}
		if err := ApplyNetemProfile(f, "w-edge", 100, 15, 0, nil, tc.dist, tc.shape); err != nil {
			t.Fatalf("%s : %v", tc.name, err)
		}
		got := strings.Join(f.Calls[0], " ")
		if !strings.HasSuffix(got, " "+tc.want) {
			t.Fatalf("%s : got %q, veut suffixe %q", tc.name, got, tc.want)
		}
		if strings.Count(got, "distribution") != 1 {
			t.Fatalf("%s : un seul jeton distribution attendu : %q", tc.name, got)
		}
	}
}

// TestApplyNetemProfileGEModelOrder — ordre tc verrouillé avec perte
// gemodel et distribution déclarée : corrélation, perte, distribution.
func TestApplyNetemProfileGEModelOrder(t *testing.T) {
	f := &FakeRunner{}
	ge := &GEModel{GoodToBad: 0.01, BadToGood: 0.19, LossBad: 0.81, LossGood: 0.01}
	if err := ApplyNetemProfile(f, "w-edge", 100, 15, 0, ge, "normal", 8, 0.25); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(f.Calls[0], " "),
		"qdisc replace dev w-edge root handle 1: netem delay 100ms 15ms 25% loss gemodel 1% 19% 81% 1% distribution normal 8ms"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
