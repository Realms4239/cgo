package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/remediate"
)

func rowFor(t *testing.T, rows []ConfiguredRealisedRow, metric string) ConfiguredRealisedRow {
	t.Helper()
	for _, r := range rows {
		if strings.HasPrefix(r.Metric, metric) {
			return r
		}
	}
	t.Fatalf("ligne %q absente de %+v", metric, rows)
	return ConfiguredRealisedRow{}
}

// Oracle principal : la table est ÉMISE quand un profil de référence a été
// apparié (valeur réelle du rapport) ; les colonnes viennent du profil et des
// mesures, jamais d'un défaut inventé.
func TestConfiguredVsRealisedEmittedFromMatchedProfile(t *testing.T) {
	res := &Result{
		AuditID: "audit-cvr", Timestamp: "2026-11-23T10:00:00Z",
		ProfileMatch:   "P2",
		RTTIdleP50:     202,
		RTTIdleP95:     205,
		ThroughputMbps: 19,
		LossPct:        0.5,
	}
	p := model.Profiles["P2"]
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	rows := rep.ConfiguredVsRealised
	if len(rows) != 4 {
		t.Fatalf("%d lignes, voulu delay/jitter/loss/rate : %+v", len(rows), rows)
	}
	delay := rowFor(t, rows, "delay")
	if !strings.Contains(delay.Configured, fmt.Sprintf("%g", p.DelayMs)) || !strings.Contains(delay.Configured, "P2") || !strings.Contains(delay.Realised, "205") {
		t.Fatalf("delay = %+v, voulu configuré P2=100 ms et réalisé 205 ms", delay)
	}
	jitter := rowFor(t, rows, "jitter")
	if !strings.Contains(jitter.Configured, "15") || !strings.Contains(jitter.Realised, "3") {
		t.Fatalf("jitter = %+v, sous-typage NetEm à afficher tel quel", jitter)
	}
	loss := rowFor(t, rows, "loss")
	if !strings.Contains(loss.Configured, "0.5") || !strings.Contains(loss.Realised, "0.5") {
		t.Fatalf("loss = %+v", loss)
	}
	rate := rowFor(t, rows, "rate")
	if !strings.Contains(rate.Configured, "20") || !strings.Contains(rate.Realised, "19") {
		t.Fatalf("rate = %+v", rate)
	}
	b, _ := json.Marshal(rep)
	if !strings.Contains(string(b), "configured_vs_realised") {
		t.Fatalf("table absente du JSON : %s", b)
	}
	if !strings.Contains(rep.RenderMarkdown(), "Configuré vs réalisé") {
		t.Fatal("table absente du rapport opérateur markdown")
	}
}

// Sans profil apparié, aucune valeur configurée n'existe : la tranche reste
// vide et omitempty garde le JSON de rapport octet pour octet identique.
func TestConfiguredVsRealisedOmittedWithoutMatch(t *testing.T) {
	res := &Result{AuditID: "audit-nomatch", Timestamp: "2026-11-23T10:00:00Z", RTTIdleP95: 205}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.ConfiguredVsRealised) != 0 {
		t.Fatalf("table non vide sans profil : %+v", rep.ConfiguredVsRealised)
	}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "configured_vs_realised") {
		t.Fatalf("champ vide sérialisé, JSON non identique : %s", b)
	}
	if strings.Contains(rep.RenderMarkdown(), "Configuré vs réalisé") {
		t.Fatal("table vide affichée dans le rapport markdown")
	}
}

// Une mesure absente devient « — » avec une note explicite, jamais un zéro.
func TestConfiguredVsRealisedAbsentRealisedIsDashWithNote(t *testing.T) {
	res := &Result{
		AuditID: "audit-absent", Timestamp: "2026-11-23T10:00:00Z",
		ProfileMatch: "P2", RTTIdleP95: 205,
		// ThroughputMbps = 0 : débit non mesuré.
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	rate := rowFor(t, rep.ConfiguredVsRealised, "rate")
	if !strings.Contains(rate.Realised, "—") || !strings.Contains(rate.Realised, "non mesur") {
		t.Fatalf("rate réalisé = %q, voulu tiret + note d'absence", rate.Realised)
	}
	if !strings.Contains(rate.Configured, "20") {
		t.Fatalf("rate configuré perdu : %q", rate.Configured)
	}
}
