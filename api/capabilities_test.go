package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/Realms4239/cgo/api/campagne"
)

// TestCapabilitiesEndpoint — le client reçoit les listes de l'hôte, pas une
// constante : bbr3 n'est annoncé que si le noyau l'expose, mais reste un
// libellé connu pour que le refus vienne du moteur avec le chemin d'install.
func TestCapabilitiesEndpoint(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/api/capabilities")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var got campagne.Capabilities
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.Qdiscs, "cake") || !slices.Contains(got.Qdiscs, "pfifo_fast") {
		t.Fatalf("qdiscs = %v", got.Qdiscs)
	}
	if !slices.Contains(got.CCs, "cubic") || !slices.Contains(got.CCs, "bbr") {
		t.Fatalf("ccs = %v", got.CCs)
	}
	if !slices.Contains(got.KnownCCs, "bbr3") {
		t.Fatalf("known_ccs = %v", got.KnownCCs)
	}
	if got.BBRv3 != slices.Contains(got.CCs, "bbr3") {
		t.Fatalf("bbr3=%v incohérent avec ccs=%v", got.BBRv3, got.CCs)
	}
}
