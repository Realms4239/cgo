package metrics

import "testing"

func TestAmortizedARPerSecond(t *testing.T) {
	// 1 000 000 Ar / (720 h × 3600 s) = 0.3858 Ar/s (3 décimales utiles)
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
	if s := AssessCAPSLA(nil); s.Measured || s.Breach || s.TargetPct != 0 || s.DeadlineMs != 0 {
		t.Fatalf("sans échantillons : %+v", s)
	}
	// 3/5 sous 220 ms = 60 % < 95 % : manquement (100,150,210 passent ; 700,300 échouent).
	s := AssessCAPSLA([]float64{100, 150, 210, 700, 300})
	if !s.Measured || !s.Breach || s.OKPct != 60 {
		t.Fatalf("60%% attendu en échec : %+v", s)
	}
	// Les deux objectifs voyagent avec le verdict : sans eux l'interface les retape.
	if s.TargetPct != CAPBreachThresholdPct || s.DeadlineMs != CAPDeadlineMs {
		t.Fatalf("objectifs non scellés : %+v", s)
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

func TestBuildEconomicsScenarios(t *testing.T) {
	e := BuildEconomics("4g", RetransRatio{Ratio: 0.05, Basis: "measured"}, 100_000_000, 5_000_000, 40)
	if len(e.Scenarios) != 5 {
		t.Fatalf("5 scénarios (GOLD, PLATINE, 81,38, APN 150, détail) : %d", len(e.Scenarios))
	}
	if e.Primary == nil || e.Primary.Model != "B" || e.Primary.Name != TariffCorporateAPN.Name {
		t.Fatalf("primaire 4g = APN corporate : %+v", e.Primary)
	}
	for _, s := range e.Scenarios {
		if s.ARPerHour <= 0 || s.Basis == "" {
			t.Fatalf("scénario incohérent : %+v", s)
		}
	}
	if e.Disclaimer == "" {
		t.Fatal("le rapport doit dire que ce sont des scénarios contractuels, pas des mesures")
	}
}

func TestPrimaryScenarioMapping(t *testing.T) {
	cases := map[string]string{
		"fiber": PlanGOLD.Name, "other": PlanGOLD.Name,
		"4g": TariffCorporateAPN.Name, "5g": TariffCorporateAPN.Name,
		"vsat": TariffWholesale.Name,
	}
	for lt, want := range cases {
		e := BuildEconomics(lt, RetransRatio{Ratio: 0.01, Basis: "estimated"}, 1e9, 1e7, 40)
		if e.Primary == nil || e.Primary.Name != want {
			t.Fatalf("link %q : primaire %+v, want %s", lt, e.Primary, want)
		}
	}
}

func TestBuildEconomicsNoRatio(t *testing.T) {
	// Sans ratio mesuré ni estimé : aucun scénario chiffré — jamais un zéro
	// qui se lirait « aucun gaspillage ».
	e := BuildEconomics("fiber", RetransRatio{}, 1e9, 0, 40)
	if len(e.Scenarios) != 0 {
		t.Fatalf("sans ratio : zéro scénario, pas des zéros : %+v", e.Scenarios)
	}
}

func TestBuildEconomicsNoRatioNoPrimary(t *testing.T) {
	// Sans ratio : AUCUN scénario ET aucun primaire — les deux moitiés de
	// l'absence, pas seulement la première.
	e := BuildEconomics("fiber", RetransRatio{}, 1e9, 0, 40)
	if len(e.Scenarios) != 0 || e.Primary != nil {
		t.Fatalf("sans ratio : ni scénarios ni primaire : %+v %+v", e.Scenarios, e.Primary)
	}
}

func TestHumanMGA(t *testing.T) {
	cases := map[float64]string{
		1000000: "1 000 000", 2500000: "2 500 000", 150: "150", 999: "999",
	}
	for v, want := range cases {
		if got := humanMGA(v); got != want {
			t.Fatalf("humanMGA(%v) = %q, want %q", v, got, want)
		}
	}
}
