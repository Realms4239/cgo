package model

import "testing"

// TestProfileAsymmetricCapacity — CapacityDown/Up : défaut up=down
// (rétrocompatible : un profil importé historique sans capacity_up_mbps
// garde sa capacité symétrique), valeurs explicites quand fournies.
func TestProfileAsymmetricCapacity(t *testing.T) {
	// symétrique : CapacityMbps seul → up = down (tous les profils natifs)
	p2 := Profiles["P2"]
	if p2.CapUp() != p2.CapDown() {
		t.Fatalf("P2 symétrique attendu: down=%v up=%v", p2.CapDown(), p2.CapUp())
	}
	// asymétrique explicite (4G réelle : 20 down / 5 up)
	p := Profile{ID: "PX", CapacityMbps: 20, CapacityUpMbps: 5}
	if p.CapDown() != 20 || p.CapUp() != 5 {
		t.Fatalf("asymétrique: down=%v up=%v", p.CapDown(), p.CapUp())
	}
	// CapacityUpMbps > CapacityMbps : incohérent, la borne est le max
	q := Profile{ID: "PY", CapacityMbps: 10, CapacityUpMbps: 50}
	if q.CapUp() != 10 {
		t.Fatalf("up plafonné par capacity_mbps: %v", q.CapUp())
	}
}
