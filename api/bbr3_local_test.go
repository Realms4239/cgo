package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/Realms4239/cgo/capabilities"
)

// Pin épinglé, verbatim de kit/install-bbr3.sh (miroir du test du module
// capabilities) : deux preuves concordantes exigées — marqueur + suffixe -g.
const pinnedBBR3Marker = `{"branch":"v3","tag":"bbrv3-2025-03-18","commit":"90210de4b779d40496dee0b89081780eeddf2a60","kernel":"6.13.7","built":"2026-11-12T00:00:00+00:00"}`

// TestBBRv3ProvisionedHostDetected — faux négatif prouvé : sur l'hôte
// google/bbr épinglé, le token v3 est « bbr » (pas « bbr3 »). Le module local
// le reconnaît via la provenance ; le v0.1.0 en cache (CCPresent("bbr3")) le
// manquait. Aucun hôte réel n'est touché : tout passe par les stubs du module.
func TestBBRv3ProvisionedHostDetected(t *testing.T) {
	// CC « bbr » seul, aucun token « bbr3 » nu : seule la provenance prouve la v3.
	defer capabilities.StubHost([]string{"cubic", "bbr"}, "6.13.7")()
	defer capabilities.StubProvisioned(pinnedBBR3Marker, "Linux version 6.13.7-g90210de4b779 #1 SMP")()

	if got := capabilities.BBRv3Token(); got != "bbr" {
		t.Fatalf("BBRv3Token = %q, want « bbr » (token v3 sur arbre google/bbr)", got)
	}
	if !capabilities.CapabilitiesNow().BBRv3 {
		t.Fatal("CapabilitiesNow().BBRv3 = false sur hôte provisionné épinglé — faux négatif")
	}
}

// TestBBRv3StockHostNotDetected — oracle hostile : sans provenance, aucun
// repli ne fabrique la v3, même si « bbr » est enregistré.
func TestBBRv3StockHostNotDetected(t *testing.T) {
	defer capabilities.StubHost([]string{"cubic", "bbr"}, "6.8.0")()
	defer capabilities.StubProvisioned("", "Linux version 6.8.0-generic #1 SMP")()

	if got := capabilities.BBRv3Token(); got != "" {
		t.Fatalf("BBRv3Token = %q sur hôte stock, want « »", got)
	}
	if capabilities.CapabilitiesNow().BBRv3 {
		t.Fatal("BBRv3 = true sur hôte stock — détection fabriquée")
	}
}

// schemaParam sert GET /api/schema et rend le paramètre demandé (contrat
// existant : map[string]any{"params": schemaParamsNow()}).
func schemaParam(t *testing.T, srv *httptest.Server, key string) struct {
	Key     string   `json:"key"`
	Enum    []string `json:"enum"`
	Default any      `json:"default"`
} {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + "/api/schema")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/schema = %d", resp.StatusCode)
	}
	var body struct {
		Params []struct {
			Key     string   `json:"key"`
			Enum    []string `json:"enum"`
			Default any      `json:"default"`
		} `json:"params"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	for _, p := range body.Params {
		if p.Key == key {
			return p
		}
	}
	t.Fatalf("paramètre %q absent de /api/schema", key)
	return struct {
		Key     string   `json:"key"`
		Enum    []string `json:"enum"`
		Default any      `json:"default"`
	}{}
}

// TestSchemaExposesDetectedBBRv3 — la voie de contrat : avec l'hôte
// provisionné (token « bbr »), /api/schema annonce bbr3 dans burst_cc et porte
// la phrase « détecté ». Même source que la validation (paramByKey).
func TestSchemaExposesDetectedBBRv3(t *testing.T) {
	withDevNoAuth(t)
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	defer capabilities.StubHost([]string{"cubic", "bbr"}, "6.13.7")()
	defer capabilities.StubProvisioned(pinnedBBR3Marker, "Linux version 6.13.7-g90210de4b779 #1 SMP")()

	burst := schemaParam(t, srv, "burst_cc")
	if !slices.Contains(burst.Enum, "bbr3") {
		t.Fatalf("burst_cc.enum n'expose pas bbr3 sur hôte détecté : %v", burst.Enum)
	}
	state := schemaParam(t, srv, "bbr3_state")
	if got, _ := state.Default.(string); got != "BBRv3 détecté — token noyau actif" {
		t.Fatalf("bbr3_state = %v, want phrase « détecté »", state.Default)
	}
}

// TestSchemaStockHostHidesBBRv3 — oracle hostile du contrat : sur hôte stock,
// /api/schema ne fabrique aucun bbr3 et garde la phrase du corpus.
func TestSchemaStockHostHidesBBRv3(t *testing.T) {
	withDevNoAuth(t)
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	defer capabilities.StubHost([]string{"cubic", "bbr"}, "6.8.0")()
	defer capabilities.StubProvisioned("", "Linux version 6.8.0-generic #1 SMP")()

	burst := schemaParam(t, srv, "burst_cc")
	if slices.Contains(burst.Enum, "bbr3") {
		t.Fatalf("burst_cc.enum expose bbr3 sur hôte stock : %v", burst.Enum)
	}
	state := schemaParam(t, srv, "bbr3_state")
	const sentence = "BBRv3 non détecté — CUBIC + BBRv1 sur noyau 6.8, détection bbr3 prête"
	if got, _ := state.Default.(string); got != sentence {
		t.Fatalf("bbr3_state = %v, want %q", state.Default, sentence)
	}
}
