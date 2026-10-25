package metrics

import "testing"

func TestAmortizedARPerSecond(t *testing.T) {
	// 1 000 000 Ar / (720 h × 3600 s) = 0.3858 Ar/s (3 déciales utiles)
	got := AmortizedARPerSecond(1_000_000)
	want := 1_000_000.0 / (720 * 3600)
	if diff := got - want; diff > 1e-12 || diff < -1e-12 {
		t.Fatalf("AmortizedARPerSecond = %v, want %v", got, want)
	}
}

func TestCapacityWasteARPerHour(t *testing.T) {
	// 10 % de gaspillage sur un GOLD : 0.1 × (1 000 000/720) = 138,89 Ar/h
	got := CapacityWasteARPerHour(1_000_000, 100, 1000)
	want := 0.1 * (1_000_000.0 / 720)
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("CapacityWasteARPerHour = %v, want %v", got, want)
	}
	if CapacityWasteARPerHour(1_000_000, 100, 0) != 0 {
		t.Fatal("transported=0 doit rendre 0, pas une division par zéro")
	}
}

func TestEnterprisePlans(t *testing.T) {
	if PlanGOLD.MonthlyOpExMGA != 1_000_000 || PlanGOLD.CIRMbps != 200 {
		t.Fatalf("GOLD = %+v", PlanGOLD)
	}
	if PlanPLATINE.MonthlyOpExMGA != 2_500_000 || PlanPLATINE.FUPBytes != 30*1024*1024*1024*1024 {
		t.Fatalf("PLATINE = %+v", PlanPLATINE)
	}
}
