package model

import (
	"math"
	"testing"
)

// TestValidateJitterCorrelation — plage 0..1, NaN refusé, 0 = défaut
// historique (aucun jeton de corrélation émis en aval).
func TestValidateJitterCorrelation(t *testing.T) {
	for _, c := range []float64{0, 0.25, 1} {
		if err := ValidateJitterCorrelation(c); err != nil {
			t.Fatalf("%v doit être accepté : %v", c, err)
		}
	}
	for _, c := range []float64{-0.01, 1.01, math.NaN()} {
		if err := ValidateJitterCorrelation(c); err == nil {
			t.Fatalf("%v doit être refusé", c)
		}
	}
}

// TestLegacyProfilesJitterCorrelationZero — P1-P4 figés au bit près : la
// corrélation n'ajoute aucun jeton par défaut, les valeurs historiques bougent
// aussi peu que le corpus gelé l'exige.
func TestLegacyProfilesJitterCorrelationZero(t *testing.T) {
	want := map[string]Profile{
		"P1": {ID: "P1", CapacityMbps: 80, DelayMs: 20, JitterMs: 2, LossPct: 0},
		"P2": {ID: "P2", CapacityMbps: 20, DelayMs: 100, JitterMs: 15, LossPct: 0.5},
		"P3": {ID: "P3", CapacityMbps: 5, DelayMs: 600, JitterMs: 30, LossPct: 1},
		"P4": {ID: "P4", CapacityMbps: 100, DelayMs: 40, JitterMs: 20, LossPct: 0.3},
	}
	for id, exp := range want {
		got, ok := Profiles[id]
		if !ok {
			t.Fatalf("profil figé %s absent", id)
		}
		if got.JitterCorrelation != 0 {
			t.Fatalf("%s : JitterCorrelation %v, want 0", id, got.JitterCorrelation)
		}
		if got.ID != exp.ID || got.CapacityMbps != exp.CapacityMbps ||
			got.DelayMs != exp.DelayMs || got.JitterMs != exp.JitterMs ||
			got.LossPct != exp.LossPct {
			t.Fatalf("%s modifié : got %+v, want %+v", id, got, exp)
		}
	}
}
