package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Realms4239/cgo/results"
)

// La racine épinglée sert réellement : le corpus gelé seedé sous
// CGO_DATA_DIR/runs est lu par les MÊMES handlers que les opérateurs — sans
// aucun réseau et sans audit sous la main. Un trou de câblage (une seule
// route qui retournerait encore « ./data » relatif) ferait échouer ce test.
func TestPinnedCorpusServesDashboard(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CGO_DATA_DIR", dir)
	seeded, _, err := results.SeedRunsFor(results.RunsRoot())
	if err != nil {
		t.Fatalf("seed : %v", err)
	}
	if !seeded {
		t.Fatal("seed : nothing was frozen under the pinned root — corpus empty")
	}

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	get := func(path string) (int, map[string]any) {
		t.Helper()
		resp, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var out map[string]any
		_ = json.Unmarshal(b, &out)
		return resp.StatusCode, out
	}

	// integralité : le corpus gelé est vu comme disponible, avec les runs
	code, j := get("/api/integrity")
	if code != http.StatusOK || j["available"] != true {
		t.Fatalf("/api/integrity = %d %v, veut 200 available:true", code, j)
	}
	if runs, ok := j["runs"].(float64); !ok || runs < 1 {
		t.Fatalf("/api/integrity = %v — aucun run gelé vu par le dashboard sous %q", j, dir)
	}

	// résultats : le dashboard rend au moins une cellule
	code, j = get("/api/results")
	if code != http.StatusOK {
		t.Fatalf("/api/results = %d %v", code, j)
	}
	groups, ok := j["groups"].([]any)
	if !ok || len(groups) == 0 {
		t.Fatalf("/api/results = %v — aucun groupe sous %q", j, dir)
	}

	// quarantine : le handler accepte le corpus épinglé sans écrire sous ./data
	code, _ = get("/api/quarantine")
	if code != http.StatusOK {
		t.Fatalf("/api/quarantine = %d", code)
	}
}
