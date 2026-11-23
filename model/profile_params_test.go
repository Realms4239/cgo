package model

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// profileParams — construit un profil par JSON : les champs optionnels encore
// absents sont ignorés par l'import, l'oracle RED reste comportemental autant
// que possible (aucune valeur fabriquée, aucune dérive du domaine).
func profileParams(t *testing.T, raw string) Profile {
	t.Helper()
	var p Profile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestValidateProfileParamsDefaultsByteIdentical — P1-P4/L1/V1 portent
// tous les nouveaux champs absents/zéro : le validateur doit rester muet
// (aucun changement de vecteur tc ni de JSON historique).
func TestValidateProfileParamsDefaultsByteIdentical(t *testing.T) {
	for _, p := range Profiles {
		if err := ValidateProfileParams(p); err != nil {
			t.Fatalf("%s : profil historique refusé : %v", p.ID, err)
		}
	}
}

func TestValidateProfileParamsDistributionClosedSet(t *testing.T) {
	for _, d := range []string{"", "uniform", "normal", "pareto", "paretonormal"} {
		p := profileParams(t, `{"id":"PD","delay_distribution":"`+d+`"}`)
		if d == "normal" || d == "pareto" || d == "paretonormal" {
			p.JitterShape = 2.5 // requise par le type déclaré
		}
		if err := ValidateProfileParams(p); err != nil {
			t.Fatalf("distribution %q doit être acceptée : %v", d, err)
		}
	}
	for _, d := range []string{"exponential", "weibull", "NORMAL", "normal "} {
		p := profileParams(t, `{"id":"PD","delay_distribution":"`+d+`"}`)
		if err := ValidateProfileParams(p); err == nil {
			t.Fatalf("distribution %q doit être refusée", d)
		}
	}
}

func TestValidateProfileParamsJitterShape(t *testing.T) {
	// uniform : la forme est inutilisée, elle doit être ABSENTE (hostile).
	p := profileParams(t, `{"id":"PU","delay_distribution":"uniform","jitter_shape":5}`)
	if err := ValidateProfileParams(p); err == nil {
		t.Fatal("jitter_shape avec uniform doit être refusé")
	}
	// Distributions à forme : la shape est requise > 0.
	for _, d := range []string{"normal", "pareto", "paretonormal"} {
		for _, shape := range []float64{0, -1, math.NaN(), math.Inf(1)} {
			p = profileParams(t, `{"id":"PS","delay_distribution":"`+d+`"}`)
			p.JitterShape = shape
			if err := ValidateProfileParams(p); err == nil {
				t.Fatalf("%s avec shape %v doit être refusé", d, shape)
			}
		}
		p = profileParams(t, `{"id":"PS","delay_distribution":"`+d+`","jitter_shape":2.5}`)
		if err := ValidateProfileParams(p); err != nil {
			t.Fatalf("%s avec shape 2.5 doit être accepté : %v", d, err)
		}
	}
	// Absence de distribution : shape zéro reste historique.
	p = profileParams(t, `{"id":"PN"}`)
	if err := ValidateProfileParams(p); err != nil {
		t.Fatalf("profil sans distribution ni shape : %v", err)
	}
}

func TestValidateProfileParamsGEModelFractions(t *testing.T) {
	good := profileParams(t, `{"id":"PG","loss_model":"gemodel","loss_p":0.01,"loss_r":0.19,"loss_bad":0.81,"loss_good":0.01}`)
	if err := ValidateProfileParams(good); err != nil {
		t.Fatalf("fractions valides : %v", err)
	}
	// Hostile : chaque fraction hors [0,1] est refusée, avant tout tc.
	for _, raw := range []string{
		`{"id":"PG","loss_model":"gemodel","loss_p":-0.1,"loss_r":0.19,"loss_bad":0.81,"loss_good":0.01}`,
		`{"id":"PG","loss_model":"gemodel","loss_p":0.01,"loss_r":1.1,"loss_bad":0.81,"loss_good":0.01}`,
		`{"id":"PG","loss_model":"gemodel","loss_p":0.01,"loss_r":0.19,"loss_bad":2,"loss_good":0.01}`,
		`{"id":"PG","loss_model":"gemodel","loss_p":0.01,"loss_r":0.19,"loss_bad":0.81,"loss_good":-0.01}`,
	} {
		if err := ValidateProfileParams(profileParams(t, raw)); err == nil {
			t.Fatalf("fractions hors plage doivent être refusées : %s", raw)
		}
	}
	// NaN/Inf échouent fermés, jamais coercés.
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		p := good
		p.LossP = bad
		if err := ValidateProfileParams(p); err == nil {
			t.Fatalf("fraction %v doit être refusée", bad)
		}
	}
	// Sans gemodel, les fractions sont ignorées (champs de format externe).
	p := profileParams(t, `{"id":"PG","loss_model":"uniform","loss_p":9}`)
	if err := ValidateProfileParams(p); err != nil {
		t.Fatalf("loss_model uniform ignore les fractions : %v", err)
	}
	p = profileParams(t, `{"id":"PG","loss_model":"gaussian"}`)
	if err := ValidateProfileParams(p); err == nil {
		t.Fatal("loss_model hors vocabulaire doit être refusé")
	}
}

func TestValidateProfileParamsBasisAndProvenance(t *testing.T) {
	for _, b := range []string{"", "measured", "standard", "declared"} {
		p := profileParams(t, `{"id":"PB","basis":"`+b+`","source_ref":"trace-42"}`)
		if err := ValidateProfileParams(p); err != nil {
			t.Fatalf("basis %q doit être accepté : %v", b, err)
		}
	}
	// Hostile : measured sans provenance = revendication de vérité non étayée.
	p := profileParams(t, `{"id":"PB","basis":"measured"}`)
	p.SourceRef = ""
	if err := ValidateProfileParams(p); err == nil {
		t.Fatal("basis=measured sans source_ref doit être refusé")
	}
	if !strings.Contains(ValidateProfileParams(p).Error(), "source_ref") {
		t.Fatalf("l'erreur doit nommer source_ref : %v", ValidateProfileParams(p))
	}
	p.SourceRef = "audit-2026-11-23#42"
	if err := ValidateProfileParams(p); err != nil {
		t.Fatalf("measured avec source_ref doit être accepté : %v", err)
	}
	if err := ValidateProfileParams(profileParams(t, `{"id":"PB","basis":"folk"}`)); err == nil {
		t.Fatal("basis hors vocabulaire doit être refusé")
	}
}
