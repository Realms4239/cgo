package audit

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendLinkAuditCAPColumns(t *testing.T) {
	dir := t.TempDir()
	res := &Result{
		AuditID: "audit-x", Timestamp: "2026-10-25T10:00:00Z",
		Site: "S", LinkType: "4g", Provider: "yas",
		CAPSLAMeasured: true, CAPDeadlineOKPct: 88.5, CAPSLABreach: true,
		CAPSLALabel: "ALERTE OPÉRATIONNELLE",
	}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "link_audit.csv"))
	hdr := strings.Split(strings.Split(string(raw), "\n")[0], ",")
	found := 0
	for _, h := range hdr {
		if h == "cap_deadline_ok_pct" || h == "cap_sla_breach" {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("colonnes CAP absentes : %v", hdr)
	}
	if !strings.Contains(string(raw), "88.5") || !strings.Contains(string(raw), "true") {
		t.Fatal("valeurs CAP non écrites")
	}
}

func TestAppendLinkAuditMigratesOldHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "link_audit.csv")
	// historique court : 3 colonnes — la migration doit rembourrer, pas casser.
	os.WriteFile(path, []byte("audit_id,timestamp,site\nold,2026,t\n"), 0644)
	res := &Result{AuditID: "new", Timestamp: "now", Site: "S", LinkType: "4g"}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	cols := strings.Split(lines[0], ",")
	for _, row := range lines[1:] {
		if n := len(strings.Split(row, ",")); n != len(cols) {
			t.Fatalf("ligne %q a %d champs, en-tête en a %d", row, n, len(cols))
		}
	}
}

func TestAppendLinkAuditCAPUnmeasuredBlank(t *testing.T) {
	// CAP non mesurée : colonnes VIDES, jamais «0.0,false» qui se lirait
	// « 0 % sous échéance, pas de manquement » — une absence, pas un zéro.
	// Par POSITION : d'autres colonnes légitimes valent 0.0 sur ce résultat vide.
	dir := t.TempDir()
	res := &Result{AuditID: "audit-y", Timestamp: "t", Site: "S", LinkType: "4g"}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(dir, "link_audit.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	hdr := rows[0]
	iPct, iBreach := -1, -1
	for i, h := range hdr {
		switch h {
		case "cap_deadline_ok_pct":
			iPct = i
		case "cap_sla_breach":
			iBreach = i
		}
	}
	if iPct < 0 || iBreach < 0 {
		t.Fatalf("colonnes CAP absentes : %v", hdr)
	}
	last := rows[len(rows)-1]
	if last[iPct] != "" || last[iBreach] != "" {
		t.Fatalf("CAP non mesurée doit être vide, pas «%s,%s»", last[iPct], last[iBreach])
	}
}
