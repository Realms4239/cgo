package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
)

// TestRenderMeasuredReportFixture writes a report from a fully measured result
// so the measured path can be inspected by eye, not just asserted on.
func TestRenderMeasuredReportFixture(t *testing.T) {
	if os.Getenv("CGO_WRITE_FIXTURE") == "" {
		t.Skip("set CGO_WRITE_FIXTURE=1 to write the sample report")
	}
	res := &Result{
		AuditID:        "audit-fixture",
		Timestamp:      "2026-10-23T10:00:00Z",
		Site:           "tana-partage-pointe",
		LinkType:       "4g",
		Provider:       "yas",
		RTTIdleP50:     46,
		RTTIdleP95:     49.2,
		RTTLoadedP50:   47,
		RTTLoadedP95:   242.8,
		ThroughputMbps: 4.9,
		LossPct:        0.4,
		HTTPSmallP95:   210,
		DataUsedMB:     18.2,
		BloatDeltaMs:   193.6,
		BloatGrade:     "C",
		BloatVerdict:   "queue sous charge",
	}
	link := remediate.Link{Interface: "wan0", RouterOSInterface: "ether1-wan", Overhead: 0, MTU: 1500}
	rep, err := BuildReport(res, link, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join("..", "..", "data", "audit-report-fixture.md")
	if err := os.WriteFile(out, []byte(rep.RenderMarkdown()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", out)
}
