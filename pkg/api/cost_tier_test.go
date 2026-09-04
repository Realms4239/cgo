package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCostTierSelect — le palier actif se choisit et s'applique : inconnu =
// 400, connu = 200 + reflété par GET /api/cost/tiers.
func TestCostTierSelect(t *testing.T) {
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(body string) (int, string) {
		resp, err := http.Post(srv.URL+"/api/cost/tier", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
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
