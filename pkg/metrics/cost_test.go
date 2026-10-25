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

func TestMeteredWasteARPerHour(t *testing.T) {
	// 1 GiB gaspillé à 81,38 Ar/Go sur une fenêtre de 180 s :
	// 81,38 × 20 = 1627,6 Ar/h — le ×20 de la directive est LE cas 180 s.
	got := MeteredWasteARPerHour(1024*1024*1024, 81.38, 180)
	want := 81.38 * 20
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("×20 : got %v want %v", got, want)
	}
	// Fenêtre mesurée 40 s : facteur 90, pas 20.
	got40 := MeteredWasteARPerHour(1024*1024*1024, 150, 40)
	if diff := got40 - 150*90; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("fenêtre 40 s : got %v want %v", got40, 150*90)
	}
	if MeteredWasteARPerHour(1e9, 81.38, 0) != 0 {
		t.Fatal("windowSeconds=0 doit rendre 0")
	}
}

func TestB2BTariffs(t *testing.T) {
	if TariffWholesale.ARPerGB != 81.38 || TariffCorporateAPN.ARPerGB != 150 {
		t.Fatalf("tarifs B2B = %+v %+v", TariffWholesale, TariffCorporateAPN)
	}
	if RetailEmergencyARPerGB != 5556 {
		t.Fatalf("secours détail = %v", RetailEmergencyARPerGB)
	}
}

func TestAssessCAPSLA(t *testing.T) {
	// Aucun échantillon chargé : non mesuré, jamais un succès de complaisance.
	if s := AssessCAPSLA(nil); s.Measured || s.Breach {
		t.Fatalf("sans échantillons : %+v", s)
	}
	// 3/5 sous 220 ms = 60 % < 95 % : manquement (100,150,210 passent ; 700,300 échouent).
	s := AssessCAPSLA([]float64{100, 150, 210, 700, 300})
	if !s.Measured || !s.Breach || s.OKPct != 60 {
		t.Fatalf("60%% attendu en échec : %+v", s)
	}
	if s.Label == "" {
		t.Fatal("un manquement doit porter son étiquette opérationnelle")
	}
	// Tous sous 220 ms : conforme, pas de label d'alerte.
	ok := AssessCAPSLA([]float64{100, 150, 210, 219})
	if !ok.Measured || ok.Breach || ok.Label != "" {
		t.Fatalf("conforme : %+v", ok)
	}
}
