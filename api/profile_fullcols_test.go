package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// TestProfileImportFullColumns — CSV étendu : up asymétrique + rafales
// Gilbert-Elliott aux positions 5..9 ; réponse GET /api/profiles les
// expose (sinon l'import UI/API écrit des champs que personne ne lit).
func TestProfileImportFullColumns(t *testing.T) {
	withDevNoAuth(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	csv := "PXF,20,100,15,0.5,5,0.01,0.001,200\n"
	req, _ := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader(csv))
	req.Header.Set("Content-Type", "text/csv")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import csv étendu: %d", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b, _ := io.ReadAll(resp2.Body)
	s := string(b)
	for _, want := range []string{`"capacity_up_mbps":5`, `"loss_burst_p":0.01`, `"loss_burst_r":0.001`, `"loss_burst_h":200`} {
		if !strings.Contains(s, want) {
			t.Fatalf("GET /api/profiles sans %s dans %s", want, s)
		}
	}
}

// TestProfileImportJitterCorrelation — colonne 10 optionnelle : 0.25 lue et
// exposée, anciennes lignes (<=10 colonnes) à zéro, hors plage refusée avant
// toute persistance (jamais de valeur silencieusement coercée).
func TestProfileImportJitterCorrelation(t *testing.T) {
	withDevNoAuth(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(csv string) int {
		req, _ := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader(csv))
		req.Header.Set("Content-Type", "text/csv")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	// 11 colonnes : la corrélation est la onzième (indice 10).
	if code := post("PJ-25,20,100,15,0.5,5,0.01,0.001,200,0.0005,0.25\n"); code != http.StatusOK {
		t.Fatalf("import corrélation : %d", code)
	}
	if got := model.Profiles["PJ-25"].JitterCorrelation; got != 0.25 {
		t.Fatalf("corrélation importée %v, want 0.25", got)
	}
	resp, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), `"jitter_correlation":0.25`) {
		t.Fatalf("GET /api/profiles sans jitter_correlation : %s", b)
	}

	// 10 colonnes (ancienne forme, k inclus) : zéro historique.
	if code := post("PJ-10,20,100,15,0.5,5,0.01,0.001,200,0.0005\n"); code != http.StatusOK {
		t.Fatalf("import ancien format : %d", code)
	}
	if got := model.Profiles["PJ-10"].JitterCorrelation; got != 0 {
		t.Fatalf("ancienne ligne : corrélation %v, want 0", got)
	}

	// Hostile : 1.5 est refusée avant persistance.
	if code := post("PJ-BAD,20,100,15,0.5,0,0,0,0,0,1.5\n"); code != http.StatusBadRequest {
		t.Fatalf("corrélation 1.5 doit être refusée : %d", code)
	}
	if _, ok := model.Profiles["PJ-BAD"]; ok {
		t.Fatal("PJ-BAD persisté malgré le refus")
	}
}
