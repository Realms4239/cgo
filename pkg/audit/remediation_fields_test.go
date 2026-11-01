package audit

import (
	"math"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
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
}
