package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/remediate"
)

func loadedResult() *Result {
	return &Result{
		AuditID:        "audit-test-1",
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
}

func testLink() remediate.Link {
	return remediate.Link{
		Interface:         "wan0",
		RouterOSInterface: "ether1-wan",
	}
}

// A report from a real load must carry the measured numbers and a remediation
// sized from them — not from anything the caller declared.
func TestReportFromMeasuredLoad(t *testing.T) {
	rep, err := BuildReport(loadedResult(), testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Measured.LoadMeasured {
		t.Fatal("load was measured but report says otherwise")
	}
	if rep.Grade != "C" {
		t.Fatalf("grade lost: %q", rep.Grade)
	}
	// QDI is loaded p95 minus loaded p50.
	if got, want := rep.Measured.QDIMs, 242.8-47.0; got < want-0.01 || got > want+0.01 {
		t.Fatalf("QDI = %.2f, want %.2f", got, want)
	}
	if rep.Remediation == nil {
		t.Fatal("no remediation from a fully measured link")
	}
	// The remediation must be sized from the measured throughput, shaped at 95%.
	joined := strings.Join(rep.Remediation.LinuxTC, "\n")
	if !strings.Contains(joined, "4.66mbit") {
		t.Fatalf("remediation not sized from measured throughput (4.9 * 0.95 = 4.66):\n%s", joined)
	}
	// Asymmetric link -> cake.
	if rep.Remediation.Qdisc != remediate.QdiscCake {
		t.Fatalf("asymmetric link got %s, want cake", rep.Remediation.Qdisc)
	}
	if rep.Remediation.RouterOS == "" {
		t.Fatal("RouterOS interface was named but no script emitted")
	}
}

// The central honesty property: no load means no grade, no throughput, no
// remediation — and every absence is named.
func TestReportWithoutLoadIsHonest(t *testing.T) {
	res := loadedResult()
	res.ThroughputMbps = 0
	res.BloatGrade = ""
	res.BloatVerdict = "Bufferbloat non mesuré"
	res.BloatDeltaMs = 0
	res.HTTPSmallP95 = 0

	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Measured.LoadMeasured {
		t.Fatal("report claims a load that never ran")
	}
	if rep.Grade != "" {
		t.Fatalf("graded an unloaded link: %q", rep.Grade)
	}
	if rep.Remediation != nil {
		t.Fatal("emitted remediation from an unmeasured link")
	}
	// Every missing metric must be named, not zero-filled silently.
	want := map[string]bool{
		"bufferbloat grade": false,
		"throughput":        false,
		"voip_r":            false,
		"remediation":       false,
		"http_small_p95":    false,
	}
	for _, u := range rep.Unavailable {
		if _, ok := want[u.Metric]; ok {
			want[u.Metric] = true
		}
		if u.Reason == "" {
			t.Fatalf("unavailable metric %q has no reason", u.Metric)
		}
	}
	for k, seen := range want {
		if !seen {
			t.Fatalf("%q missing from unavailable list: %+v", k, rep.Unavailable)
		}
	}
}

// deadline_ok_pct is never available from a link audit, because the audit keeps
// percentiles and not per-probe completions. It must always be declared absent.
func TestDeadlineAlwaysDeclaredUnavailable(t *testing.T) {
	rep, _ := BuildReport(loadedResult(), testLink(), "", true, false)
	for _, u := range rep.Unavailable {
		if u.Metric == "deadline_ok_pct" {
			return
		}
	}
	t.Fatal("deadline_ok_pct was not declared unavailable; it cannot be computed from a link audit")
}

func TestVoIPRComputedOnlyWithLoad(t *testing.T) {
	rep, _ := BuildReport(loadedResult(), testLink(), "", true, false)
	if rep.Measured.VoIPR <= 0 {
		t.Fatal("VoIP R not computed despite a measured load")
	}
	res := loadedResult()
	res.ThroughputMbps = 0
	res.BloatGrade = ""
	rep, _ = BuildReport(res, testLink(), "", true, false)
	if rep.Measured.VoIPR != 0 {
		t.Fatalf("computed VoIP R without a load: %.2f", rep.Measured.VoIPR)
	}
}

func TestWriteArtifactsCreatesBothFormats(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "link_audit.csv")
	if err := os.WriteFile(csvPath, []byte("audit_id\nx\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rep, err := BuildReport(loadedResult(), testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := rep.SealSource(csvPath); err != nil {
		t.Fatal(err)
	}
	jp, mp, err := WriteArtifacts(rep, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{jp, mp, filepath.Join(dir, "audit_report.json"), filepath.Join(dir, "audit_report.md")} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing artifact %s: %v", p, err)
		}
	}
	var got Report
	b, _ := os.ReadFile(jp)
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Provenance.SourceSHA256 == "" {
		t.Fatal("artifact lost the source seal")
	}
	if got.Phases.LoadSeconds != 0 && got.Phases.LoadSeconds != rep.Phases.LoadSeconds {
		t.Fatalf("phase metadata drifted: %+v", got.Phases)
	}
}

func TestSealSourceHashesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "link_audit.csv")
	content := "audit_id,timestamp\nA,2026-10-23T10:00:00Z\nB,2026-10-23T11:00:00Z\n"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	rep, _ := BuildReport(loadedResult(), testLink(), "", true, false)
	if err := rep.SealSource(p); err != nil {
		t.Fatal(err)
	}
	if len(rep.Provenance.SourceSHA256) != 64 {
		t.Fatalf("not a SHA-256: %q", rep.Provenance.SourceSHA256)
	}
	if rep.Provenance.RowCount != 2 {
		t.Fatalf("row count = %d, want 2", rep.Provenance.RowCount)
	}
}

func TestRenderJSONRoundTrips(t *testing.T) {
	rep, _ := BuildReport(loadedResult(), testLink(), "", true, false)
	data, err := rep.RenderJSON()
	if err != nil {
		t.Fatal(err)
	}
	var back Report
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("report JSON does not round-trip: %v", err)
	}
	if back.AuditID != rep.AuditID || back.Grade != rep.Grade {
		t.Fatal("round-trip lost fields")
	}
}

func TestRenderMarkdownStatesAbsences(t *testing.T) {
	res := loadedResult()
	res.ThroughputMbps = 0
	res.BloatGrade = ""
	rep, _ := BuildReport(res, testLink(), "", true, false)
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "not measured") {
		t.Fatalf("markdown hides the missing grade:\n%s", md)
	}
	if !strings.Contains(md, "Not measured") {
		t.Fatalf("markdown has no absence section:\n%s", md)
	}
	if strings.Contains(md, "## Remediation") {
		t.Fatalf("markdown offers remediation for an unmeasured link:\n%s", md)
	}
}

func TestRenderMarkdownIncludesRemediationWhenMeasured(t *testing.T) {
	rep, _ := BuildReport(loadedResult(), testLink(), "", true, false)
	// Provenance only exists once the source is sealed; an unsealed report must
	// not print a provenance section with an empty hash.
	dir := t.TempDir()
	p := filepath.Join(dir, "link_audit.csv")
	if err := os.WriteFile(p, []byte("audit_id\nA\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := rep.SealSource(p); err != nil {
		t.Fatal(err)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "## Remediation") {
		t.Fatalf("no remediation section:\n%s", md)
	}
	if !strings.Contains(md, "RouterOS") {
		t.Fatalf("no RouterOS section:\n%s", md)
	}
	if !strings.Contains(md, "SHA-256") {
		t.Fatalf("no provenance section:\n%s", md)
	}
}

func TestBuildReportRejectsNil(t *testing.T) {
	if _, err := BuildReport(nil, testLink(), "", true, false); err == nil {
		t.Fatal("accepted a nil result")
	}
}

// The caller cannot smuggle a capacity in through the Link: the measurement wins.
func TestCallerCannotDeclareCapacity(t *testing.T) {
	l := testLink()
	l.CapacityMbps = 1000 // a lie
	l.RTTms = 1
	rep, err := BuildReport(loadedResult(), l, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(rep.Remediation.LinuxTC, "\n")
	if strings.Contains(joined, "1000mbit") {
		t.Fatalf("caller-declared capacity reached the config:\n%s", joined)
	}
}

func TestBuildReportEconomicsAllScenarios(t *testing.T) {
	res := &Result{
		AuditID: "a", Timestamp: "t", LinkType: "4g",
		ThroughputMbps: 10, BloatGrade: "A", LoadSeconds: 40,
		DataUsedMB: 50, LossPct: 4,
		RTTIdleP50: 20, RTTIdleP95: 25, RTTLoadedP50: 25, RTTLoadedP95: 30,
		RetransRatio: metrics.RetransRatio{Ratio: 0.04, Basis: "measured"},
		Direction:    "up", FlowCount: 4, JFIPct: 97.5,
		CAPSLAMeasured: true, CAPDeadlineOKPct: 97, CAPSLABreach: false,
	}
	rep, err := BuildReport(res, remediate.Link{}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Economics == nil || len(rep.Economics.Scenarios) != 5 {
		t.Fatalf("Economics : %+v", rep.Economics)
	}
	if rep.Economics.Primary == nil || rep.Economics.Primary.Name != metrics.TariffCorporateAPN.Name {
		t.Fatalf("primaire 4g : %+v", rep.Economics.Primary)
	}
	if !rep.CAPSLA.Measured || rep.CAPSLA.Breach {
		t.Fatalf("CAP : %+v", rep.CAPSLA)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "97.5") {
		t.Fatal("JFI absent du rendu")
	}
	if !strings.Contains(md, metrics.TariffWholesale.Name) || !strings.Contains(md, "retail-outofbundle") {
		t.Fatal("tous les scénarios doivent être visibles, y compris secours")
	}
	if !strings.Contains(md, "objectif DGM 220 ms") && rep.CAPSLA.Measured {
		t.Fatal("le cadre du 220 ms doit être dit (objectif opérateur, pas un seuil OMM)")
	}
	if !strings.Contains(md, "scénarios contractuels") {
		t.Fatal("la section économique doit se dire contractuelle, pas mesurée")
	}
}

func TestBuildReportNoRatioNoEconomics(t *testing.T) {
	res := &Result{
		AuditID: "a", Timestamp: "t", LinkType: "fiber",
		ThroughputMbps: 10, BloatGrade: "A", LoadSeconds: 40, DataUsedMB: 50, LossPct: 0,
		RTTIdleP50: 20, RTTIdleP95: 25, RTTLoadedP50: 25, RTTLoadedP95: 30,
	}
	rep, err := BuildReport(res, remediate.Link{}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Economics != nil {
		t.Fatal("sans ratio : pas de valorisation chiffrée (jamais de zéro fabriqué)")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "economics" {
			found = true
		}
	}
	if !found {
		t.Fatal("l'absence de valorisation doit être déclarée indisponible")
	}
}
