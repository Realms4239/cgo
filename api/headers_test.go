package api

import (
	"net/http/httptest"
	"testing"
)

// TestSecurityHeadersOn200And404 — nosniff, DENY, Referrer-Policy et la CSP
// self-only sont posés sur la réponse, qu'elle soit 200 ou 404 ; HSTS est
// absent hors TLS.
func TestSecurityHeadersOn200And404(t *testing.T) {
	for _, path := range []string{"/api/health", "/api/chemin-inconnu"} {
		h := New(Deps{})
		req := httptest.NewRequest("GET", path, nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		want := map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":       "DENY",
			"Referrer-Policy":       "strict-origin-when-cross-origin",
			"Content-Security-Policy": "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'",
		}
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Fatalf("GET %s : %s=%q, want %q", path, k, got, v)
			}
		}
		if h := rec.Header().Get("Strict-Transport-Security"); h != "" {
			t.Fatalf("GET %s : HSTS posé hors TLS (%q)", path, h)
		}
	}
}

// TestSecurityHeadersHSTSOnlyOverTLS — HSTS n'est posé que sur une réponse
// servie sous TLS réel.
func TestSecurityHeadersHSTSOnlyOverTLS(t *testing.T) {
	h := New(Deps{})
	req := httptest.NewRequest("GET", "https://example.com/api/health", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=31536000" {
		t.Fatalf("HSTS=%q, want max-age=31536000 sous TLS", got)
	}
	// les autres en-têtes restent présents sous TLS
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("nosniff manquant sous TLS : %q", got)
	}
}
