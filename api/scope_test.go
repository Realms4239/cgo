package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestBoundaryRateLimitScopedToAPIPaths — le limiteur ne couvre que /api/* :
// un chemin SPA/statique, même répété au-delà de la rafale, ne doit jamais
// consommer le seau de lecture ni renvoyer 429. La 61e lecture /api/health
// reste 429 avec Retry-After, exactement comme avant le ciblage.
func TestBoundaryRateLimitScopedToAPIPaths(t *testing.T) {
	withDevNoAuth(t)
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	// (a) plus que la rafale de lecture sur un chemin non-API : jamais 429.
	for i := 0; i <= readPerMin; i++ {
		resp, err := http.Get(srv.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			t.Fatalf("GET / #%d : 429 sur un chemin non-API — le limiteur n'est pas ciblé", i+1)
		}
		resp.Body.Close()
	}

	// (b) la 61e lecture /api/health reste un 429 avec Retry-After.
	var last int
	var saw429 bool
	for i := 0; i <= readPerMin; i++ {
		resp, err := http.Get(srv.URL + "/api/health")
		if err != nil {
			t.Fatal(err)
		}
		last = resp.StatusCode
		if i == readPerMin && resp.StatusCode == http.StatusTooManyRequests {
			assertRetryAfter(t, resp, "", "/api/health")
			saw429 = true
		}
		resp.Body.Close()
	}
	if last != http.StatusTooManyRequests || !saw429 {
		t.Fatalf("61e lecture /api/health : %d (429 vu=%v), want 429 avec Retry-After", last, saw429)
	}
}
