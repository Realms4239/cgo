package cite

import (
	"strings"
	"testing"
)

const testSeal = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func testReport() AuditReport {
	return AuditReport{
		AuditID:   "audit-1",
		Timestamp: "2026-10-28T10:00:00Z",
		Site:      "site-a",
		LinkType:  "4g",
		Provider:  "yas",
		Measured:  AuditMeasured{Direction: "down"},
		Provenance: &AuditProvenance{SourceSHA256: testSeal},
	}
}

// TestAuditCitationSealed — chemin scellé : id, site, lien, direction, scellé.
func TestAuditCitationSealed(t *testing.T) {
	s := AuditCitation(testReport())
	for _, want := range []string{"audit-1", "site-a", "4g/yas", "direction descendant (down)", testSeal} {
		if !strings.Contains(s, want) {
			t.Errorf("citation %q : %q manquant", s, want)
		}
	}
}

// TestAuditCitationOmitsMissing — champs absents omis, phrase de scellé présente.
func TestAuditCitationOmitsMissing(t *testing.T) {
	r := AuditReport{
		AuditID:    "audit-1",
		Measured:   AuditMeasured{},
		Provenance: &AuditProvenance{SourceSHA256: testSeal},
	}
	s := AuditCitation(r)
	if !strings.Contains(s, "Rapport scellé") {
		t.Errorf("citation %q : phrase de scellé manquante", s)
	}
}

// TestAuditCitationMissingSeal — jamais de citation partielle sans scellé.
func TestAuditCitationMissingSeal(t *testing.T) {
	r := testReport()
	r.Provenance = nil
	if s := AuditCitation(r); s != "" {
		t.Errorf("sans provenance : %q, attendu \"\"", s)
	}
	r.Provenance = &AuditProvenance{SourceSHA256: "  "}
	if s := AuditCitation(r); s != "" {
		t.Errorf("scellé vide : %q, attendu \"\"", s)
	}
}

// TestAuditCitationEmptyID — identifiant vide = "".
func TestAuditCitationEmptyID(t *testing.T) {
	r := testReport()
	r.AuditID = ""
	if s := AuditCitation(r); s != "" {
		t.Errorf("id vide : %q, attendu \"\"", s)
	}
}
