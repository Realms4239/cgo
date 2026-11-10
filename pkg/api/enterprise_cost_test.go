package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Realms4239/cgo/metrics"
)

// TestCostEnterpriseCatalog — le catalogue entreprise expose les MÊMES
// constantes que pkg/metrics (source unique : un tarif changé au paquet
// se retrouve ici, jamais de nombre recopié), avec la provenance de
// chaque valeur : devis opérateur marqué quoted, disclaimer verbatim.
func TestCostEnterpriseCatalog(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/cost/enterprise")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}

	var out struct {
		Plans []struct {
			Name           string  `json:"name"`
			MonthlyOpExMGA float64 `json:"monthly_opex_mga"`
			CIRMbps        float64 `json:"cir_mbps"`
			FUPBytes       float64 `json:"fup_bytes"`
		} `json:"plans"`
		Tariffs []struct {
			Name    string  `json:"name"`
			ARPerGB float64 `json:"ar_per_gb"`
			Quoted  bool    `json:"quoted"`
		} `json:"tariffs"`
		EmergencyARPerGB float64           `json:"emergency_ar_per_gb"`
		PrimaryMapping   map[string]string `json:"primary_mapping"`
		Disclaimer       string            `json:"disclaimer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}

	// plans — les deux noms, opex issus des vars du paquet, capacité ou FUP
	if len(out.Plans) != 2 {
		t.Fatalf("plans: %d, want 2", len(out.Plans))
	}
	opex := map[string]float64{}
	for _, p := range out.Plans {
		opex[p.Name] = p.MonthlyOpExMGA
		switch p.Name {
		case metrics.PlanGOLD.Name:
			if p.CIRMbps != metrics.PlanGOLD.CIRMbps {
				t.Fatalf("GOLD cir_mbps %g, want %g", p.CIRMbps, metrics.PlanGOLD.CIRMbps)
			}
			if p.FUPBytes != 0 {
				t.Fatalf("GOLD ne porte pas de FUP, got %g", p.FUPBytes)
			}
		case metrics.PlanPLATINE.Name:
			if p.FUPBytes != metrics.PlanPLATINE.FUPBytes {
				t.Fatalf("PLATINE fup_bytes %g, want %g", p.FUPBytes, metrics.PlanPLATINE.FUPBytes)
			}
			if p.CIRMbps != 0 {
				t.Fatalf("PLATINE n'a pas de CIR, got %g", p.CIRMbps)
			}
		default:
			t.Fatalf("plan inattendu: %s", p.Name)
		}
	}
	if opex[metrics.PlanGOLD.Name] != metrics.PlanGOLD.MonthlyOpExMGA || opex[metrics.PlanPLATINE.Name] != metrics.PlanPLATINE.MonthlyOpExMGA {
		t.Fatalf("opex perdu: %v", opex)
	}

	// tarifs — grossiste sans devis, APN marqué devis opérateur (invariant 4)
	if len(out.Tariffs) != 2 {
		t.Fatalf("tariffs: %d, want 2", len(out.Tariffs))
	}
	for _, tr := range out.Tariffs {
		switch tr.Name {
		case metrics.TariffWholesale.Name:
			if tr.ARPerGB != metrics.TariffWholesale.ARPerGB || tr.Quoted {
				t.Fatalf("wholesale: %+v, want ar_per_gb %g quoted:false", tr, metrics.TariffWholesale.ARPerGB)
			}
		case metrics.TariffCorporateAPN.Name:
			if tr.ARPerGB != metrics.TariffCorporateAPN.ARPerGB || !tr.Quoted {
				t.Fatalf("APN: %+v, want ar_per_gb %g quoted:true (devis opérateur)", tr, metrics.TariffCorporateAPN.ARPerGB)
			}
		default:
			t.Fatalf("tarif inattendu: %s", tr.Name)
		}
	}

	// pénalité de secours — la constante, pas un nombre recopié
	if out.EmergencyARPerGB != float64(metrics.RetailEmergencyARPerGB) {
		t.Fatalf("emergency %g, want %d", out.EmergencyARPerGB, metrics.RetailEmergencyARPerGB)
	}

	// mapping primaire — les cinq types de lien, depuis la même source que
	// le rapport d'audit
	want := map[string]string{
		"fiber": metrics.PlanGOLD.Name,
		"other": metrics.PlanGOLD.Name,
		"4g":    metrics.TariffCorporateAPN.Name,
		"5g":    metrics.TariffCorporateAPN.Name,
		"vsat":  metrics.TariffWholesale.Name,
	}
	if len(out.PrimaryMapping) != 5 {
		t.Fatalf("primary_mapping: %d clés, want 5", len(out.PrimaryMapping))
	}
	for k, v := range want {
		if out.PrimaryMapping[k] != v {
			t.Fatalf("primary_mapping[%s]=%s, want %s", k, out.PrimaryMapping[k], v)
		}
	}

	// disclaimer — verbatim du paquet (source unique BuildEconomics)
	wantD := metrics.BuildEconomics("", metrics.RetransRatio{}, 0, 0, 0).Disclaimer
	if out.Disclaimer != wantD {
		t.Fatalf("disclaimer altéré:\n got %q\nwant %q", out.Disclaimer, wantD)
	}
}
