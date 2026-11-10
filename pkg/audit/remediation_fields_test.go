package audit

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Realms4239/cgo/remediate"
)

// BuildReport dérive Measured.LoadMeasured de
// « res.ThroughputMbps > 0 && res.BloatGrade != "" » (pkg/audit/report.go:233) :
// les cas ci-dessous jouent sur ces deux champs, jamais sur un booléen
// LoadMeasured qui n'existe pas sur Result.
func TestReportRemediationCarriesRatesAndInterfaces(t *testing.T) {
	res := &Result{
		AuditID: "audit-rem", Site: "Siège", LinkType: "4g",
		RTTIdleP50: 100, RTTIdleP95: 110,
		LoadAttempted: true, Direction: "both",
		BloatDeltaMs: 12, BloatGrade: "A",
		ThroughputMbps: 5.1, ThroughputDownMbps: 18.5, LoadSeconds: 40,
	}
	rep, err := BuildReport(res, remediate.Link{
		Interface: "eth0", RouterOSInterface: "ether1-wan",
	}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	r := rep.Remediation
	if r == nil {
		t.Fatal("remédiation absente alors que les deux sens sont mesurés")
	}
	if math.Abs(r.UpMeasuredMbps-5.1) > 1e-9 || math.Abs(r.DownMeasuredMbps-18.5) > 1e-9 {
		t.Fatalf("débits mesurés = %v / %v", r.UpMeasuredMbps, r.DownMeasuredMbps)
	}
	if math.Abs(r.UpShapedMbps-remediate.ShapedRate(5.1)) > 1e-9 {
		t.Fatalf("débit façonné montant = %v", r.UpShapedMbps)
	}
	if r.LinuxIface != "eth0" || r.RouterOSWanIface != "ether1-wan" {
		t.Fatalf("interfaces = %q / %q", r.LinuxIface, r.RouterOSWanIface)
	}
	if r.RouterOSLanIface != remediate.DefaultRouterOSLan || !r.RouterOSLanAssumed {
		t.Fatalf("interface LAN = %q assumée=%v", r.RouterOSLanIface, r.RouterOSLanAssumed)
	}
}

func TestReportRemediationOmitsUnmeasuredDirection(t *testing.T) {
	res := &Result{
		AuditID: "audit-up", Site: "Siège", LinkType: "4g",
		RTTIdleP50: 100, LoadAttempted: true, Direction: "up",
		BloatDeltaMs: 12, BloatGrade: "A", ThroughputMbps: 5.1, LoadSeconds: 40,
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0", RouterOSLanInterface: "lan"}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation.DownMeasuredMbps != 0 || len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatalf("script descendant inventé : %+v", rep.Remediation)
	}
	// Pas de script descendant, donc aucune hypothèse d'interface à étiqueter :
	// le rapport ne doit rien affirmer sur une remédiation qu'il ne porte pas.
	// Le contraire afficherait, dans la carte 20B, une « interface supposée »
	// sur un poste qui n'a jamais nommé de routeur.
	if rep.Remediation.RouterOSIngress != "" {
		t.Fatalf("script RouterOS descendant inventé : %q", rep.Remediation.RouterOSIngress)
	}
	if rep.Remediation.RouterOSLanIface != "" || rep.Remediation.RouterOSLanAssumed {
		t.Fatalf("interface LAN déclarée sans script descendant : %q assumée=%v",
			rep.Remediation.RouterOSLanIface, rep.Remediation.RouterOSLanAssumed)
	}
}

// Unmeasured → absence nommée, jamais un script.
func TestRemediationAvailableMirrorsPresence(t *testing.T) {
	res := &Result{
		AuditID: "audit-unmeasured", Site: "Siège", LinkType: "4g", Direction: "up",
		RTTIdleP50: 100, LoadAttempted: true,
		// ThroughputMbps = 0 → aucune charge
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.RemediationAvailable || rep.Remediation != nil {
		t.Fatalf("remédiation disponible sans mesure : %+v", rep.Remediation)
	}
	if rep.RemediationState != RemediationObservationState {
		t.Fatalf("état = %q, want observation_only", rep.RemediationState)
	}

	res2 := &Result{
		AuditID: "audit-measured", Site: "Siège", LinkType: "4g", Direction: "up",
		RTTIdleP50: 100, LoadAttempted: true, BloatGrade: "A",
		ThroughputMbps: 5.1, LoadSeconds: 40,
	}
	rep2, err := BuildReport(res2, remediate.Link{Interface: "eth0"}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !rep2.RemediationAvailable || rep2.Remediation == nil {
		t.Fatalf("remédiation absente alors que la charge est mesurée")
	}
	if rep2.RemediationState != RemediationAvailableState {
		t.Fatalf("état = %q, want available", rep2.RemediationState)
	}
}

// Les six champs structurés traversent le JSON.
func TestRemediationStructuredFieldsRoundTrip(t *testing.T) {
	res := &Result{
		AuditID: "audit-rt", Site: "Siège", LinkType: "4g", Direction: "both",
		RTTIdleP50: 100, LoadAttempted: true, BloatGrade: "A",
		ThroughputMbps: 5.1, ThroughputDownMbps: 18.5, LoadSeconds: 40,
	}
	rep, err := BuildReport(res, remediate.Link{
		Interface: "eth0", RouterOSInterface: "ether1-wan",
	}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	js, err := rep.RenderJSON()
	if err != nil {
		t.Fatal(err)
	}
	var back Report
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatal(err)
	}
	r := back.Remediation
	if r == nil {
		t.Fatal("la remédiation a disparu du JSON")
	}
	if math.Abs(r.UpMeasuredMbps-5.1) > 1e-9 || math.Abs(r.DownMeasuredMbps-18.5) > 1e-9 {
		t.Fatalf("débits mesurés = %v / %v", r.UpMeasuredMbps, r.DownMeasuredMbps)
	}
	if math.Abs(r.UpShapedMbps-remediate.ShapedRate(5.1)) > 1e-9 {
		t.Fatalf("débit façonné montant = %v", r.UpShapedMbps)
	}
	if r.LinuxIface != "eth0" || r.RouterOSWanIface != "ether1-wan" {
		t.Fatalf("interfaces = %q / %q", r.LinuxIface, r.RouterOSWanIface)
	}
	if r.RouterOSLanIface != remediate.DefaultRouterOSLan || !r.RouterOSLanAssumed {
		t.Fatalf("interface LAN = %q assumée=%v", r.RouterOSLanIface, r.RouterOSLanAssumed)
	}
	if !back.RemediationAvailable || back.RemediationState != RemediationAvailableState {
		t.Fatalf("état JSON = %q / %v", back.RemediationState, back.RemediationAvailable)
	}
}

// Un rapport scellé avant la slice décode comme observation_only, jamais available.
func TestSealedReportWithoutStateDecodesAsObservationOnly(t *testing.T) {
	legacy := []byte(`{"audit_id":"old","remediation_state":"","remediation_available":false}`)
	var rep Report
	if err := json.Unmarshal(legacy, &rep); err != nil {
		t.Fatal(err)
	}
	if rep.RemediationState != RemediationObservationState || rep.RemediationAvailable {
		t.Fatalf("legacy = %q / %v", rep.RemediationState, rep.RemediationAvailable)
	}
	absent := []byte(`{"audit_id":"older"}`)
	var rep2 Report
	if err := json.Unmarshal(absent, &rep2); err != nil {
		t.Fatal(err)
	}
	if rep2.RemediationState != RemediationObservationState || rep2.RemediationAvailable {
		t.Fatalf("absent = %q / %v", rep2.RemediationState, rep2.RemediationAvailable)
	}
}
