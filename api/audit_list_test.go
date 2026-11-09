package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestAuditListMarksArchivableReports — le dashboard ne propose « OUVRIR » que
// là où le dossier de rapport existe. Une ligne antérieure à l'archivage par
// dossier doit être annoncée comme telle avant le clic, pas après un 404.
func TestAuditListMarksArchivableReports(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, "data", "audits", "audit-2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "link_audit.csv"),
		[]byte("audit_id,site\naudit-1,Site X\naudit-2,Departement X\n,sans id\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "audits", "audit-2", "audit_report.json"),
		[]byte(`{"audit_id":"audit-2"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/audit/list")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var body struct {
		Audits []map[string]string `json:"audits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Audits) != 3 {
		t.Fatalf("%d lignes, want 3", len(body.Audits))
	}
	want := map[string]string{"audit-1": "false", "audit-2": "true", "": "false"}
	for _, row := range body.Audits {
		if got, ok := want[row["audit_id"]]; !ok {
			t.Fatalf("ligne inattendue : %q", row["audit_id"])
		} else if row["has_report"] != got {
			t.Fatalf("ligne %q : has_report=%q, want %q", row["audit_id"], row["has_report"], got)
		}
	}
}
