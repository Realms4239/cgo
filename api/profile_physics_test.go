package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// physicsRow — ligne CSV historique (13 champs, indices 0..12) suivie des
// colonnes optionnelles 13..21 (0-based) dans l'ordre figé du contrat.
func physicsRow(id string, extra ...string) string {
	base := []string{id, "20", "100", "15", "0.5", "0", "0", "0", "0", "0", "0", "0", "0"}
	return strings.Join(append(base, extra...), ",") + "\n"
}

func physicsServer(t *testing.T) *httptest.Server {
	t.Helper()
	withDevNoAuth(t)
	chdirTemp(t)
	srv := httptest.NewServer(New(Deps{}))
	t.Cleanup(srv.Close)
	return srv
}

func postPhysicsCSV(t *testing.T, srv *httptest.Server, row string) int {
	t.Helper()
	req, _ := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader(row))
	req.Header.Set("Content-Type", "text/csv")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode
}

// TestProfileImportPhysicsColumns — colonnes 13..21 lues et exposées dans
// /api/profiles avec les noms de champs exacts attendus par Task 4.
func TestProfileImportPhysicsColumns(t *testing.T) {
	srv := physicsServer(t)
	row := physicsRow("PPH", "normal", "8", "gemodel", "0.01", "0.19", "0.81", "0.01", "measured", "trace-42")
	if code := postPhysicsCSV(t, srv, row); code != http.StatusOK {
		t.Fatalf("import physique : %d", code)
	}
	p := model.Profiles["PPH"]
	if p.DelayDistribution != "normal" || p.JitterShape != 8 || p.LossModel != "gemodel" ||
		p.LossP != 0.01 || p.LossR != 0.19 || p.LossBad != 0.81 || p.LossGood != 0.01 ||
		p.Basis != "measured" || p.SourceRef != "trace-42" {
		t.Fatalf("champs importés : %+v", p)
	}
	resp, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	for _, want := range []string{
		`"delay_distribution":"normal"`, `"jitter_shape":8`, `"loss_model":"gemodel"`,
		`"loss_p":0.01`, `"loss_r":0.19`, `"loss_bad":0.81`, `"loss_good":0.01`,
		`"basis":"measured"`, `"source_ref":"trace-42"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("GET /api/profiles sans %s : %s", want, b)
		}
	}
}

// TestProfileImportPhysicsLegacyZero — anciennes lignes (<=12 colonnes au
// sens historique, 13 champs) : tous les nouveaux champs à zéro, aucun jeton
// ajouté au JSON — le format historique reste byte-identique.
func TestProfileImportPhysicsLegacyZero(t *testing.T) {
	srv := physicsServer(t)
	if code := postPhysicsCSV(t, srv, physicsRow("POLD")); code != http.StatusOK {
		t.Fatalf("import ancien format : %d", code)
	}
	p := model.Profiles["POLD"]
	if p.DelayDistribution != "" || p.JitterShape != 0 || p.LossModel != "" ||
		p.LossP != 0 || p.LossR != 0 || p.LossBad != 0 || p.LossGood != 0 ||
		p.Basis != "" || p.SourceRef != "" {
		t.Fatalf("ancienne ligne doit rester à zéro : %+v", p)
	}
	resp, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	// La carte des profils est globale au paquet : d'autres tests ont pu
	// importer des profils avec physique. On n'inspecte que l'entrée POLD.
	var body struct {
		Profiles []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatal(err)
	}
	var pold map[string]any
	for _, pr := range body.Profiles {
		if pr["id"] == "POLD" {
			pold = pr
			break
		}
	}
	if pold == nil {
		t.Fatalf("POLD absent de la réponse : %s", b)
	}
	for _, absent := range []string{"delay_distribution", "jitter_shape", "loss_model", "basis", "source_ref"} {
		if _, ok := pold[absent]; ok {
			t.Fatalf("JSON historique modifié, %s présent dans POLD : %v", absent, pold)
		}
	}
}

// TestProfileImportPhysicsHostile — refus AVANT persistance des valeurs
// invalides, jamais une valeur coercée dans un profil importé.
func TestProfileImportPhysicsHostile(t *testing.T) {
	srv := physicsServer(t)
	cases := []struct {
		id  string
		row string
	}{
		{"PBAD1", physicsRow("PBAD1", "", "0", "", "0", "0", "0", "0", "measured", "")},
		{"PBAD2", physicsRow("PBAD2", "exponential", "0", "", "0", "0", "0", "0", "", "")},
		{"PBAD3", physicsRow("PBAD3", "", "0", "gemodel", "1.5", "0.19", "0.81", "0.01", "", "")},
		{"PBAD4", physicsRow("PBAD4", "uniform", "5", "", "0", "0", "0", "0", "", "")},
	}
	for _, tc := range cases {
		if code := postPhysicsCSV(t, srv, tc.row); code != http.StatusBadRequest {
			t.Fatalf("%s doit être refusé : %d", tc.id, code)
		}
		if _, ok := model.Profiles[tc.id]; ok {
			t.Fatalf("%s persisté malgré le refus", tc.id)
		}
	}
}
