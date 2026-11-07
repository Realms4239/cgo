package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestAuditListCarriesTheSealedVerdict — la liste des preuves affiche un mot par
// ligne, et ce mot vient du rapport scellé. Le CSV porte une note, mais PAS
// l'information « la charge a-t-elle été tentée » : recalculer un verdict depuis
// le CSV rouvrirait exactement le faux vert que la tache 1 de 20A a fermé (une
// note sans charge établie qui devient verte). Donc : verdict publié quand
// l'artefact existe, absence sinon — jamais un verdict dérivé de la note.
func TestAuditListCarriesTheSealedVerdict(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, "data", "audits", "audit-vert"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "data", "audits", "audit-vieux"), 0o755); err != nil {
		t.Fatal(err)
	}
	csv := "audit_id,site,bloat_grade\naudit-vert,Site X,A\naudit-vieux,Departement X,A+\naudit-sans-rapport,Becineman A,\n"
	if err := os.WriteFile(filepath.Join(dir, "data", "link_audit.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "audits", "audit-vert", "audit_report.json"),
		[]byte(`{"audit_id":"audit-vert","operator_verdict":{"level":"green","word":"Lien sain","sentence":"Sous charge, la latence reste sous l'objectif."}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Archive scellée avant le champ : level et word sont vides dans l'artefact.
	if err := os.WriteFile(filepath.Join(dir, "data", "audits", "audit-vieux", "audit_report.json"),
		[]byte(`{"audit_id":"audit-vieux","bloat_grade":"A+"}`), 0o644); err != nil {
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
	rows := map[string]map[string]string{}
	for _, row := range body.Audits {
		rows[row["audit_id"]] = row
	}
	if len(rows) != 3 {
		t.Fatalf("%d lignes distinctes, want 3", len(rows))
	}

	vert := rows["audit-vert"]
	if vert == nil {
		t.Fatal("ligne audit-vert absente")
	}
	if vert["verdict_level"] != "green" || vert["verdict_word"] != "Lien sain" {
		t.Errorf("verdict du rapport scellé non publié : %+v", vert)
	}

	// Les deux absences : un archive sans champ et une ligne sans rapport ne
	// reçoivent AUCUN verdict, pas même « neutre » — le listeur ne classifie pas.
	for _, id := range []string{"audit-vieux", "audit-sans-rapport"} {
		row := rows[id]
		if row == nil {
			t.Fatalf("ligne %s absente", id)
		}
		if _, ok := row["verdict_level"]; ok {
			t.Errorf("%s publie un verdict %q là où aucun rapport ne le scelle", id, row["verdict_level"])
		}
		if _, ok := row["verdict_word"]; ok {
			t.Errorf("%s publie un mot %q sans artefact", id, row["verdict_word"])
		}
	}
}
