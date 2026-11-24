package metrics

import (
	"math"
	"testing"
)

// fidelityModelMs — δ du profil dans les oracles : la borne epsilon vaut 5 ms.
const fidelityModelMs = 100.0

func fidelityRepeated(n int, v float64) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = v
	}
	return s
}

// fidelityMixed — « good » paires à goodV (écart 4 ms, dans la borne) et
// « bad » paires à badV (écart 30 ms, hors borne).
func fidelityMixed(good, bad int, goodV, badV float64) []float64 {
	s := make([]float64, 0, good+bad)
	for i := 0; i < good; i++ {
		s = append(s, goodV)
	}
	for i := 0; i < bad; i++ {
		s = append(s, badV)
	}
	return s
}

// TestFidelityThresholdsAreFixed — les seuils sont déclarés dans le code, pas
// dans une configuration : aucun appelant ne peut les élargir.
func TestFidelityThresholdsAreFixed(t *testing.T) {
	if FidelityEMDMax != 0.10 || FidelityGammaMax != 0.05 || FidelityEpsilonFracMax != 0.05 || FidelityMinSamples != 30 {
		t.Fatalf("seuils figés altérés : EMD=%v gamma=%v epsilon=%v min=%d",
			FidelityEMDMax, FidelityGammaMax, FidelityEpsilonFracMax, FidelityMinSamples)
	}
}

// TestDistributionEMDIdentical — deux CDF identiques donnent zero, même si
// l'ordre des échantillons diffère (l'EMD trie les deux séries).
func TestDistributionEMDIdentical(t *testing.T) {
	live := []float64{10, 20, 30, 40, 50}
	emulated := []float64{50, 10, 40, 20, 30}
	got, err := DistributionEMD(live, emulated)
	if err != nil {
		t.Fatalf("DistributionEMD: %v", err)
	}
	if got != 0 {
		t.Fatalf("EMD de distributions identiques = %.12f, voulu 0", got)
	}
}

// TestDistributionEMDShiftedUniform — un décalage constant produit un EMD
// positif égal au décalage divisé par la moyenne live (sans unité).
func TestDistributionEMDShiftedUniform(t *testing.T) {
	live := make([]float64, 100)
	emulated := make([]float64, 100)
	mean := 0.0
	for i := range live {
		live[i] = float64(i + 1)
		emulated[i] = live[i] + 5
		mean += live[i]
	}
	mean /= float64(len(live))

	got, err := DistributionEMD(live, emulated)
	if err != nil {
		t.Fatalf("DistributionEMD: %v", err)
	}
	want := 5 / mean
	if got <= 0 || math.Abs(got-want) > 1e-9 {
		t.Fatalf("EMD décalé = %.12f, voulu %.12f (> 0)", got, want)
	}

	// Un décalage plus grand franchit le plafond déclaré de 10 %.
	far := make([]float64, 100)
	for i := range far {
		far[i] = live[i] + 20
	}
	gotFar, err := DistributionEMD(live, far)
	if err != nil {
		t.Fatalf("DistributionEMD (grand décalage): %v", err)
	}
	if gotFar <= FidelityEMDMax {
		t.Fatalf("EMD grand décalage = %.12f, doit dépasser %v", gotFar, FidelityEMDMax)
	}
}

// TestDistributionEMDRefusesInvalid — un jeu vide, une moyenne non positive
// ou un échantillon non fini ne produisent jamais un nombre « mesuré ».
func TestDistributionEMDRefusesInvalid(t *testing.T) {
	cases := [][2][]float64{
		{nil, []float64{1}},
		{[]float64{1}, nil},
		{[]float64{0, 0}, []float64{0, 0}},
		{[]float64{math.NaN(), 10}, []float64{10, 10}},
		{[]float64{10, 10}, []float64{10, math.Inf(1)}},
	}
	for i, c := range cases {
		got, err := DistributionEMD(c[0], c[1])
		if err == nil || got != 0 {
			t.Fatalf("cas %d : EMD = %v, err = %v ; refus et 0 attendus", i, got, err)
		}
	}
}

// TestDelayAccuracyTrust — assez de paires, écart dans la borne : trust.
func TestDelayAccuracyTrust(t *testing.T) {
	state, eps, ok := DelayAccuracy(fidelityRepeated(30, 104), fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state != FidelityTrust || !ok {
		t.Fatalf("30 paires conformes : state=%q ok=%v, voulu trust/true", state, ok)
	}
	if math.Abs(eps-4) > 1e-9 {
		t.Fatalf("epsilon observé = %.12f, voulu 4 ms", eps)
	}
}

// TestDelayAccuracyGammaRefusesTrust — gamma = 6/100 = 0.06 > 0.05 : le
// verdict ne peut pas être trust, même avec 100 paires.
func TestDelayAccuracyGammaRefusesTrust(t *testing.T) {
	state, _, ok := DelayAccuracy(fidelityMixed(94, 6, 104, 130), fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state == FidelityTrust || ok {
		t.Fatalf("gamma 0.06 : state=%q ok=%v, refus de trust attendu", state, ok)
	}
}

// TestDelayAccuracyEpsilonRefusesTrust — gamma = 0.05 (non strictement au-dessus)
// mais le quantile 95 des écarts dépasse 5 % du modèle : pas de trust.
func TestDelayAccuracyEpsilonRefusesTrust(t *testing.T) {
	state, eps, ok := DelayAccuracy(fidelityMixed(95, 5, 104, 130), fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state == FidelityTrust || ok {
		t.Fatalf("epsilon > 5 %% du modèle : state=%q ok=%v eps=%.4f, refus attendu", state, ok, eps)
	}
	if eps <= FidelityEpsilonFracMax*fidelityModelMs {
		t.Fatalf("epsilon observé = %.4f ms, doit dépasser la borne %.4f ms", eps, FidelityEpsilonFracMax*fidelityModelMs)
	}
}

// TestDelayAccuracyTooFewPairs — 29 paires conformes ne peuvent pas donner
// trust : le plancher de 30 paires financières est maintenu.
func TestDelayAccuracyTooFewPairs(t *testing.T) {
	state, _, ok := DelayAccuracy(fidelityRepeated(29, 104), fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state == FidelityTrust || ok {
		t.Fatalf("29 paires : state=%q ok=%v, bounded attendu", state, ok)
	}
}

// TestDelayAccuracyNaNSampleIsDiscard — un écart non calculable n'est pas
// remplacé par une borne : le verdict est discard.
func TestDelayAccuracyNaNSampleIsDiscard(t *testing.T) {
	samples := append(fidelityRepeated(29, 104), math.NaN())
	state, _, ok := DelayAccuracy(samples, fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state != FidelityDiscard || ok {
		t.Fatalf("échantillon NaN : state=%q ok=%v, discard/false attendu", state, ok)
	}
	inf := append(fidelityRepeated(29, 104), math.Inf(1))
	state, _, ok = DelayAccuracy(inf, fidelityModelMs, FidelityGammaMax, FidelityEpsilonFracMax)
	if state != FidelityDiscard || ok {
		t.Fatalf("échantillon Inf : state=%q ok=%v, discard/false attendu", state, ok)
	}
}

// TestDelayAccuracyRefusesWidenedThresholds — même si l'appelant passe un
// seuil plus large, le plafond figé interdit le trust.
func TestDelayAccuracyRefusesWidenedThresholds(t *testing.T) {
	samples := fidelityRepeated(30, 104)
	cases := [][3]float64{
		{0.06, FidelityEpsilonFracMax, fidelityModelMs},
		{FidelityGammaMax, 0.06, fidelityModelMs},
		{FidelityGammaMax, FidelityEpsilonFracMax, 0},
	}
	for i, c := range cases {
		state, _, ok := DelayAccuracy(samples, c[2], c[0], c[1])
		if state == FidelityTrust || ok {
			t.Fatalf("cas %d (gamma=%v epsilonFrac=%v model=%v) : state=%q ok=%v, refus attendu",
				i, c[0], c[1], c[2], state, ok)
		}
	}
}
