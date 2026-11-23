package model

import (
	"math"
	"testing"
)

// TestValidateSerial — sérialisation opt-in : 0/0 = défaut historique ;
// overhead sans taux, taux négatif/NaN et overhead négatif refusés avant
// toute mutation tc (jamais une valeur silencieusement coercée).
func TestValidateSerial(t *testing.T) {
	for _, ok := range []struct {
		rate float64
		ov   int
	}{{0, 0}, {40, 0}, {40, 38}, {100, 1}} {
		if err := ValidateSerial(ok.rate, ok.ov); err != nil {
			t.Fatalf("rate %v overhead %v doit être accepté : %v", ok.rate, ok.ov, err)
		}
	}
	for _, bad := range []struct {
		rate float64
		ov   int
	}{{0, 38}, {-1, 0}, {40, -1}, {math.NaN(), 0}} {
		if err := ValidateSerial(bad.rate, bad.ov); err == nil {
			t.Fatalf("rate %v overhead %v doit être refusé", bad.rate, bad.ov)
		}
	}
}

// TestLegacyProfilesSerialZero — P1-P4/L1/V1 figés : aucune sérialisation
// par défaut, aucun jeton tc ajouté au corpus gelé.
func TestLegacyProfilesSerialZero(t *testing.T) {
	for id := range BuiltinProfileIDs {
		p := Profiles[id]
		if p.SerialRateMbps != 0 || p.PacketOverheadBytes != 0 {
			t.Fatalf("%s : sérialisation non nulle par défaut : %+v", id, p)
		}
	}
}
