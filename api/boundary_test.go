package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// mutatingRoutes — routes de mutation protégées par le boundary. /api/login
// et /api/logout sont volontairement anonymes (entrée/sortie de session) et
// donc hors de cette liste de conformance.
var mutatingRoutes = []string{
	"/api/auth/logout", "/api/cost/tier", "/api/burst", "/api/watch", "/api/shape",
	"/api/run/start", "/api/run/stop", "/api/run/skip", "/api/figures/regen",
	"/api/audit/start", "/api/audit/stop", "/api/audit/toprofile",
	"/api/profile/import", "/api/profile/delete",
}

// assertNoLeak — oracle hostile : un corps d'erreur ne contient ni le secret,
// ni le chemin demandé, ni le préfixe de l'API.
func assertNoLeak(t *testing.T, body []byte, secret, path string) {
	t.Helper()
	s := string(body)
	for _, needle := range []string{secret, "/api/", path, "CGO_API_TOKEN", "RemoteAddr"} {
		if needle == "" {
			continue
		}
		if strings.Contains(s, needle) {
			t.Fatalf("fuite de %q dans le corps d'erreur : %s", needle, s)
		}
	}
}

// assertRetryAfter — un 429 porte un Retry-After entier positif (jamais 0,
// jamais vide) et un corps sans fuite.
func assertRetryAfter(t *testing.T, resp *http.Response, secret, path string) {
	t.Helper()
	ra := resp.Header.Get("Retry-After")
	n, err := strconv.Atoi(ra)
	if err != nil || n < 1 {
		t.Fatalf("Retry-After=%q, want entier >= 1", ra)
	}
	body, _ := io.ReadAll(resp.Body)
	assertNoLeak(t, body, secret, path)
}

// TestBoundaryUnauthenticatedMutateIs401BeforeHandler — conformance : chaque
// route de mutation refuse un appel anonyme AVANT le handler (aucune dépendance
// injectée n'est appelée), avec un corps structuré sans fuite.
func TestBoundaryUnauthenticatedMutateIs401BeforeHandler(t *testing.T) {
	secret := "boundary-secret-do-not-leak"
	t.Setenv("CGO_API_TOKEN", secret)
	SetDevNoAuth(false)
	var handlerCalls int
	srv := httptest.NewServer(New(Deps{
		StartFn: func(RunOpts) error { handlerCalls++; return nil },
		StopFn:  func() { handlerCalls++ },
		SkipFn:  func() { handlerCalls++ },
		ShapeFn: func(ShapeReq) error { handlerCalls++; return nil },
		BurstFn: func(string, int) error { handlerCalls++; return nil },
		WatchFn: func(bool) error { handlerCalls++; return nil },
	}))
	defer srv.Close()

	for _, path := range mutatingRoutes {
		req, err := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("POST %s anonyme : %d, want 401", path, resp.StatusCode)
		}
		if got := resp.Header.Get("WWW-Authenticate"); got != "Bearer" {
			t.Fatalf("POST %s : WWW-Authenticate=%q, want Bearer", path, got)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assertNoLeak(t, body, secret, path)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil || payload["error"] == "" {
			t.Fatalf("corps 401 non structuré : %s", body)
		}
	}
	if handlerCalls != 0 {
		t.Fatalf("%d handlers appelés malgré le refus 401", handlerCalls)
	}
}

// TestBoundaryCrossOwnerIs403BeforeHandler — le garde d'appartenance traduit
// la décision en 403 sans appeler le handler, et le corps ne fuit rien. Sans
// Bearer, principalFrom rend un Principal vide (non-admin) : chemin hostile
// réel, aucun faux rôle injecté dans l'application.
func TestBoundaryCrossOwnerIs403BeforeHandler(t *testing.T) {
	secret := "owner-secret-do-not-leak"
	t.Setenv("CGO_API_TOKEN", secret)
	called := false
	ownerOf := func(*http.Request) string { return "un-autre-proprietaire" }
	next := func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}
	h := (&Server{}).requireOwnerMutate(next, ownerOf)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/run/start", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("mutation cross-owner : %d, want 403", rec.Code)
	}
	if called {
		t.Fatal("handler appelé malgré le refus 403")
	}
	assertNoLeak(t, rec.Body.Bytes(), secret, "/api/run/start")
}

// TestBoundaryMutateRateLimit11thIs429BeforeHandler — oracle des limites
// déclarées : 10 mutations/minute, rafale 5. La 11e mutation rapide est 429
// avec Retry-After, et le handler n'est appelé que pendant la rafale.
func TestBoundaryMutateRateLimit11thIs429BeforeHandler(t *testing.T) {
	withDevNoAuth(t)
	var handlerCalls int
	srv := httptest.NewServer(New(Deps{StopFn: func() { handlerCalls++ }}))
	defer srv.Close()

	const limit = 10
	var eleventh int
	var saw429 bool
	for i := 0; i <= limit; i++ {
		resp, err := http.Post(srv.URL+"/api/run/stop", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		if i == limit {
			eleventh = resp.StatusCode
			if resp.StatusCode == http.StatusTooManyRequests {
				assertRetryAfter(t, resp, "", "/api/run/stop")
				saw429 = true
			}
		}
		resp.Body.Close()
	}
	if eleventh != http.StatusTooManyRequests || !saw429 {
		t.Fatalf("11e mutation : %d (429 vu=%v), want 429 avec Retry-After", eleventh, saw429)
	}
	// rafale 5 : au plus 5 appels au handler, jamais sur les requêtes refusées.
	if handlerCalls > 5 {
		t.Fatalf("%d appels handler, want <= 5 (rafale déclarée)", handlerCalls)
	}
}

// TestBoundaryReadRateLimit61stIs429 — oracle des limites déclarées : 60
// lectures/minute, rafale 20. La 61e lecture rapide est 429 avec Retry-After.
func TestBoundaryReadRateLimit61stIs429(t *testing.T) {
	withDevNoAuth(t)
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	var last int
	var saw429 bool
	for i := 0; i <= 60; i++ {
		resp, err := http.Get(srv.URL + "/api/cost/tiers")
		if err != nil {
			t.Fatal(err)
		}
		last = resp.StatusCode
		if i == 60 {
			if resp.StatusCode == http.StatusTooManyRequests {
				assertRetryAfter(t, resp, "", "/api/cost/tiers")
				saw429 = true
			}
		}
		resp.Body.Close()
	}
	if last != http.StatusTooManyRequests || !saw429 {
		t.Fatalf("61e lecture : %d (429 vu=%v), want 429 avec Retry-After", last, saw429)
	}
}
