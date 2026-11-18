package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
)

// postConfirmed — POST avec la confirmation explicite exigée depuis le plan K2
// sur les cinq routes destructives. Partagé par les tests api du paquet.
func postConfirmed(t *testing.T, url string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Confirm-Action", "true")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// TestCostTierSelect — le palier actif se choisit et s'applique : inconnu =
// 400, connu = 200 + reflété par GET /api/cost/tiers.
func TestCostTierSelect(t *testing.T) {
	withDevNoAuth(t)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(body string) (int, string) {
		resp := postConfirmed(t, srv.URL+"/api/cost/tier", bytes.NewBufferString(body))
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	if code, _ := post(`{"tier":"nope"}`); code != http.StatusBadRequest {
		t.Fatalf("palier inconnu: %d, want 400", code)
	}
	if code, s := post(`{"tier":"yas-ftth-100gb"}`); code != http.StatusOK || !strings.Contains(s, `"ok":true`) {
		t.Fatalf("palier connu: %d %s", code, s)
	}
	resp, err := http.Get(srv.URL + "/api/cost/tiers")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), `"default":"yas-ftth-100gb"`) {
		t.Fatalf("tiers ne reflète pas le choix: %s", b)
	}
	// restaurer le défaut du contexte DSI pour les autres tests
	if code, _ := post(`{"tier":"yas-month-4.5gb"}`); code != http.StatusOK {
		t.Fatal("restauration défaut")
	}
}

// TestCostTierRefusedWhileCampaignRuns — le palier est gelé avec le run :
// le changer en cours de campagne donnerait deux bases de coût dans le même gel.
func TestCostTierRefusedWhileCampaignRuns(t *testing.T) {
	withDevNoAuth(t)
	before := metrics.ActiveTierName()
	h := New(Deps{RunningFn: func() bool { return true }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := postConfirmed(t, srv.URL+"/api/cost/tier",
		bytes.NewBufferString(`{"tier":"yas-day-1gb"}`))
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, voulu 409 (%s)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "gelé") {
		t.Fatalf("le refus ne dit pas que le palier est gelé : %s", body)
	}
	if after := metrics.ActiveTierName(); after != before {
		t.Fatalf("palier global modifié malgré le refus : %s → %s", before, after)
	}
}
