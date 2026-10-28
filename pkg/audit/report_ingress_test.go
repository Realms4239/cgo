package audit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
)

func testIngressLink() remediate.Link {
	l := testLink()
	l.RouterOSLanInterface = "ether2-lan"
	return l
}

func TestReportDownEmitsOnlyIngress(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	// The down window is the headline, so the measured capacity is
	// ThroughputMbps; ThroughputDownMbps is the `both` mirror and stays 0.
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("no remediation for a measured down link")
	}
	if len(rep.Remediation.LinuxTC) != 0 {
		t.Fatalf("down report emitted an upload script: %v", rep.Remediation.LinuxTC)
	}
	if len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("down report emitted no ingress script")
	}
	joined := strings.Join(rep.Remediation.LinuxTCIngress, "\n")
	if !strings.Contains(joined, "ifb0") || !strings.Contains(joined, " ingress") {
		t.Fatalf("ingress script is not an IFB ingress config:\n%s", joined)
	}
	if !strings.Contains(joined, "4.66mbit") {
		t.Fatalf("ingress script is not sized from the measured down rate:\n%s", joined)
	}
	if rep.Remediation.RouterOS != "" {
		t.Fatalf("down report emitted an upload RouterOS script:\n%s", rep.Remediation.RouterOS)
	}
	if !strings.Contains(rep.Remediation.RouterOSIngress, "parent=ether2-lan") {
		t.Fatalf("download queue tree is not on the LAN interface:\n%s", rep.Remediation.RouterOSIngress)
	}
	if !strings.Contains(rep.Remediation.Reason, "download path") {
		t.Fatalf("down reason still justifies upload ack-filter: %q", rep.Remediation.Reason)
	}
}

func TestReportDownWithoutLANStillEmitsLinuxIngress(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("down report lost Linux ingress when no LAN interface was named")
	}
	if rep.Remediation.RouterOSIngress != "" {
		t.Fatal("emitted RouterOS ingress without a LAN interface")
	}
}

func TestReportDownHasEmptyLinuxTCArray(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	js, err := json.Marshal(rep.Remediation)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(js), `"linux_tc":[]`) {
		t.Fatalf("down remediation must carry an empty linux_tc array, not null:\n%s", js)
	}
	if strings.Contains(string(js), `"linux_tc":null`) {
		t.Fatalf("linux_tc is null:\n%s", js)
	}
}

func TestReportBothEmitsBothDirections(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputDownMbps = 10
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("no remediation for a measured both link")
	}
	up := strings.Join(rep.Remediation.LinuxTC, "\n")
	in := strings.Join(rep.Remediation.LinuxTCIngress, "\n")
	if !strings.Contains(up, "4.66mbit") {
		t.Fatalf("upload script is not sized from ThroughputMbps:\n%s", up)
	}
	if !strings.Contains(in, "9.50mbit") {
		t.Fatalf("download script is not sized from ThroughputDownMbps:\n%s", in)
	}
	if !strings.Contains(rep.Remediation.RouterOS, "cgo-") {
		t.Fatal("both report has no upload RouterOS script")
	}
	if !strings.Contains(rep.Remediation.RouterOSIngress, "cgoingress-") {
		t.Fatal("both report has no download RouterOS script")
	}
	md := rep.RenderMarkdown()
	for _, want := range []string{"upload (egress)", "download (ingress IFB)", "download (LAN queue tree)"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown heading %q missing:\n%s", want, md)
		}
	}
}

func TestReportBothMissingDownCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputDownMbps = 0
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTC) == 0 {
		t.Fatal("the measured upload remediation was lost")
	}
	if len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatal("a download config was sized from the upload number")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation (download)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing down capacity is silent: %+v", rep.Unavailable)
	}
}

func TestReportBothMissingUpCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputMbps = 0
	res.ThroughputDownMbps = 10
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("the measured download remediation was lost")
	}
	if len(rep.Remediation.LinuxTC) != 0 {
		t.Fatal("an upload config was sized from the download number")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation (upload)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing up capacity is silent: %+v", rep.Unavailable)
	}
}

func TestReportDownMissingCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	res.ThroughputMbps = 0
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation != nil {
		t.Fatal("emitted a remediation for an unmeasured down link")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unmeasured down remediation is silent: %+v", rep.Unavailable)
	}
}

func TestReportDownWithoutLinuxInterfaceStillNamesRouterOS(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	l := testIngressLink()
	l.Interface = ""
	rep, err := BuildReport(res, l, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("RouterOS ingress was lost with no Linux interface")
	}
	if len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatal("invented a Linux interface")
	}
	if rep.Remediation.RouterOSIngress == "" {
		t.Fatal("RouterOS ingress not generated from its own interface")
	}
}
