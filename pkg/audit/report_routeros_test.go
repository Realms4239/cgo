package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReportCarriesRouterOSValidation(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.RouterOSValidation = &RouterOSValidation{
		Version: "7.15.3", Level: "windows-tcg", ExportSHA256: "abcdef123456",
		Checks: []RouterOSCheck{{Name: "queue type cgo-cake.kind", Expected: "cake", Found: "cake", Pass: true}},
	}
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"routeros_validation"`) {
		t.Fatal("routeros_validation missing from JSON")
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Configuration RouterOS validée (7.15.3, CHR)") {
		t.Fatalf("markdown lacks the validation line:\n%s", md)
	}
	if strings.Contains(md, "appliquée") {
		t.Fatal("markdown claims the config was applied")
	}
}

func TestReportWithoutValidationOmitsBoth(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "routeros_validation") {
		t.Fatal("absent validation still serialized")
	}
	if strings.Contains(rep.RenderMarkdown(), "Configuration RouterOS") {
		t.Fatal("absent validation still rendered")
	}
}

func TestReportFailedValidationSaysNonValidée(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.RouterOSValidation = &RouterOSValidation{
		Version: "7.15.3", Level: "windows-tcg", ExportSHA256: "abcdef123456",
		Checks: []RouterOSCheck{{Name: "queue type cgo-cake.kind", Expected: "cake", Found: "none", Pass: false}},
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Configuration RouterOS non validée") {
		t.Fatal("failed validation not stated")
	}
	if strings.Contains(md, "Configuration RouterOS validée") {
		t.Fatal("failed validation rendered as a pass")
	}
}

func TestReportEmptyChecksIsNotAValidation(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.RouterOSValidation = &RouterOSValidation{
		Version: "7.15.3", Level: "windows-tcg", ExportSHA256: "abcdef123456",
		Checks: []RouterOSCheck{},
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "non validée") || !strings.Contains(md, "aucune vérification") {
		t.Fatalf("empty checks not stated:\n%s", md)
	}
	if strings.Contains(md, "Configuration RouterOS validée") {
		t.Fatal("empty checks rendered as a pass")
	}
}
