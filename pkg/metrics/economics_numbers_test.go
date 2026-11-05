package metrics

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildEconomicsNumbersPerModel(t *testing.T) {
	e := BuildEconomics("fiber", RetransRatio{Ratio: 0.0187, Basis: "measured"}, 1e9, 1.87e7, 40)
	if e.Primary == nil {
		t.Fatal("aucun scénario primaire")
	}
	var a, b, em *EconomicsScenario
	for i := range e.Scenarios {
		switch e.Scenarios[i].Model {
		case "A":
			if a == nil {
				a = &e.Scenarios[i]
			}
		case "B":
			if b == nil {
				b = &e.Scenarios[i]
			}
		case "emergency":
			em = &e.Scenarios[i]
		}
	}
	if a == nil || b == nil || em == nil {
		t.Fatalf("scénarios manquants : %+v", e.Scenarios)
	}
	if a.Numbers.MonthlyOpExMGA <= 0 || a.Numbers.CIRMbps <= 0 {
		t.Fatalf("modèle A sans charge mensuelle ni CIR : %+v", a.Numbers)
	}
	if a.Numbers.ARPerGB != 0 {
		t.Fatalf("modèle A ne doit pas passer par l'arithmétique du volume : %+v", a.Numbers)
	}
	if b.Numbers.ARPerGB <= 0 || b.Numbers.MonthlyOpExMGA != 0 {
		t.Fatalf("modèle B : %+v", b.Numbers)
	}
	if em.Numbers.ARPerGB != RetailEmergencyARPerGB {
		t.Fatalf("pénalité de secours : %+v", em.Numbers)
	}
	if a.Numbers.RetransRatio != 0.0187 || a.Numbers.WastedBytes != 1.87e7 {
		t.Fatalf("base de gaspillage absente : %+v", a.Numbers)
	}
}

// PlanPLATINE est un plan à volume plafonné (FUP), pas à capacité garantie : son
// scénario A porte sa charge mensuelle et AUCUN CIR. Le FUP est exposé par le
// catalogue d'entreprise (GET /api/cost/enterprise → fup_bytes, var source), pas
// par EconomicsNumbers — ici on vérifie l'absence de zéro inventé, et non un CIR
// que le contrat ne prévoit pas.
func TestBuildEconomicsNumbersModelAWithoutCIR(t *testing.T) {
	e := BuildEconomics("fiber", RetransRatio{Ratio: 0.02, Basis: "measured"}, 1e9, 2e7, 40)
	var plat *EconomicsScenario
	for i := range e.Scenarios {
		if e.Scenarios[i].Name == PlanPLATINE.Name {
			plat = &e.Scenarios[i]
		}
	}
	if plat == nil {
		t.Fatalf("scénario PLATINE absent : %+v", e.Scenarios)
	}
	if plat.Numbers.Model != "A" {
		t.Fatalf("modèle PLATINE : %q, voulu A", plat.Numbers.Model)
	}
	if plat.Numbers.MonthlyOpExMGA != PlanPLATINE.MonthlyOpExMGA {
		t.Fatalf("charge mensuelle PLATINE : %v, voulu %v",
			plat.Numbers.MonthlyOpExMGA, PlanPLATINE.MonthlyOpExMGA)
	}
	if plat.Numbers.CIRMbps != 0 || plat.Numbers.ARPerGB != 0 {
		t.Fatalf("PLATINE invente un CIR ou un prix au volume : %+v", plat.Numbers)
	}
}

// La voie de sortie anticipée reste intacte : sans base de réémission, ou sans
// octet transporté, aucun scénario — donc aucun Numbers fabriqué (un zéro se
// lirait « aucun gaspillage »).
func TestBuildEconomicsNumbersNotInventedWhenUnmeasured(t *testing.T) {
	cases := []struct {
		name        string
		ratio       RetransRatio
		transported float64
	}{
		{"base de réémission vide", RetransRatio{Ratio: 0.0187, Basis: ""}, 1e9},
		{"rien de transporté", RetransRatio{Ratio: 0.0187, Basis: "measured"}, 0},
	}
	for _, tc := range cases {
		e := BuildEconomics("fiber", tc.ratio, tc.transported, 1.87e7, 40)
		if len(e.Scenarios) != 0 || e.Primary != nil {
			t.Fatalf("%s : ni scénario ni primaire attendus, obtenu %+v / %+v",
				tc.name, e.Scenarios, e.Primary)
		}
	}
}

// La base est réellement machine-readable : les clés JSON sortent, et un
// scénario n'emporte pas les clés d'un autre modèle (omitempty).
func TestEconomicsNumbersJSONKeys(t *testing.T) {
	e := BuildEconomics("vsat", RetransRatio{Ratio: 0.01, Basis: "measured"}, 1e9, 1e7, 40)
	var wholesale, gold *EconomicsScenario
	for i := range e.Scenarios {
		switch e.Scenarios[i].Name {
		case TariffWholesale.Name:
			wholesale = &e.Scenarios[i]
		case PlanGOLD.Name:
			gold = &e.Scenarios[i]
		}
	}
	if wholesale == nil || gold == nil {
		t.Fatalf("scénarios attendus absents : %+v", e.Scenarios)
	}
	wb, err := json.Marshal(wholesale)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := json.Marshal(gold)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"numbers"`, `"ar_per_gb"`, `"retrans_ratio"`, `"wasted_bytes"`, `"transport_bytes"`, `"load_seconds"`} {
		if !strings.Contains(string(wb), want) {
			t.Fatalf("scénario volume sans clé %s : %s", want, wb)
		}
	}
	for _, absent := range []string{`"monthly_opex_mga"`, `"cir_mbps"`} {
		if strings.Contains(string(wb), absent) {
			t.Fatalf("un scénario au volume ne doit pas porter %s : %s", absent, wb)
		}
	}
	if strings.Contains(string(gs), `"ar_per_gb"`) || !strings.Contains(string(gs), `"cir_mbps"`) {
		t.Fatalf("un scénario CIR ne doit pas porter le prix au volume : %s", gs)
	}
}
