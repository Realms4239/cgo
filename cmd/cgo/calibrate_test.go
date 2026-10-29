package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

func TestRunCalibrateWritesSealedArtifacts(t *testing.T) {
	dir := t.TempDir()
	res := &audit.Result{AuditID: "audit-cal"}
	for i := 0; i < 64; i++ {
		phase := "baseline"
		rtt := 100 + float64(i%8)
		if i >= 48 {
			phase = "load"
			rtt = 500
		}
		res.Telemetry = append(res.Telemetry, audit.TelemetryRow{Phase: phase, OffsetMs: int64(i) * 400, RTTms: rtt})
	}
	if _, err := audit.WriteTelemetry(dir, res); err != nil {
		t.Fatal(err)
	}
	report := `{"site":"TEST","link_type":"5g"}`
	if err := os.WriteFile(filepath.Join(dir, "audits", "audit-cal", "audit_report.json"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "calibrations")
	if code := runCalibrate([]string{"--audit", "audit-cal", "--profile", "P2", "--data", dir, "--out", out}); code != 0 {
		t.Fatalf("runCalibrate code %d", code)
	}
	cal, err := qdisc.LoadCalibration(out, "P2")
	if err != nil {
		t.Fatal(err)
	}
	if cal.SourceAudit != "audit-cal" || cal.SourceSite != "TEST" || cal.SourceLinkType != "5g" {
		t.Fatalf("provenance : %+v", cal)
	}
	if cal.Basis != "measured" || cal.HopModel != "rtt_floor_split" || cal.SourceTelemetrySHA256 == "" || cal.SourceReportSHA256 == "" {
		t.Fatalf("sceaux/provenance : %+v", cal)
	}
	if cal.Samples != 48 {
		t.Fatalf("échantillons baseline = %d, want 48 (les lignes load sont exclues)", cal.Samples)
	}
	if cal.SourceP999RTTms > 200 {
		t.Fatalf("ligne load comptée dans la queue : p99.9 %.1f", cal.SourceP999RTTms)
	}
}
