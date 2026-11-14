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

// TestResultsETagRoundtrip — le corpus gelé ne change pas : l'ETag doit être
// stable entre deux appels, et un If-None-Match joignant ce tag doit passer
// en 304 (le client n'a pas à re-transférer la même réponse).
func TestResultsETagRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CGO_DATA_DIR", dir)
	if _, _, err := results.SeedRunsFor(results.RunsRoot()); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	get := func(ifNoneMatch string) (int, string) {
		t.Helper()
		req, _ := http.NewRequest("GET", srv.URL+"/api/results", nil)
		if ifNoneMatch != "" {
			req.Header.Set("If-None-Match", ifNoneMatch)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
		return resp.StatusCode, resp.Header.Get("ETag")
	}

	code, etag := get("")
	if code != http.StatusOK {
		t.Fatalf("want 200, got %d", code)
	}
	if etag == "" {
		t.Fatal("ETag manquant sur /api/results")
	}
	code304, etag2 := get(etag)
	if code304 != http.StatusNotModified {
		t.Fatalf("If-None-Match %q → 304 attendu, eu %d", etag, code304)
	}
	if etag2 != etag {
		t.Fatalf("ETag instable sur corpus gelé : %q puis %q", etag, etag2)
	}
}
