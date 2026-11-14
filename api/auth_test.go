package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRequireAuth_UnauthenticatedMutationIs401(t *testing.T) {
	os.Setenv("CGO_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CGO_API_TOKEN")
	h := requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/shape", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

// TestSessionToken_RoundTrip — le cookie de session est le HMAC-SHA256 du
// secret de CGO_API_TOKEN sur le message figé, hex, vérifié en temps constant ;
// un octet modifié échoue, un secret vide rend le cookie inerte.
func TestSessionToken_RoundTrip(t *testing.T) {
	got := sessionToken("test-token-123")
	if len(got) != 64 {
		t.Fatalf("longueur du MAC hex = %d, want 64", len(got))
	}
	if sessionToken("test-token-123") != got {
		t.Fatal("sessionToken non déterministe")
	}
	if sessionToken("autre-token") == got {
		t.Fatal("sessionToken identique pour deux secrets différents")
	}
	if !validSession(got, "test-token-123") {
		t.Fatal("cookie légitime refusé")
	}
	bad := "0" + got[1:]
	if bad == got {
		bad = "1" + got[1:]
	}
	if validSession(bad, "test-token-123") {
		t.Fatal("cookie altéré accepté")
	}
	if validSession("", "test-token-123") {
		t.Fatal("cookie vide accepté")
	}
	if validSession(got, "") {
		t.Fatal("cookie vérifié avec secret vide — l'adaptateur doit rester inerte")
	}
	if validSession("zz", "test-token-123") {
		t.Fatal("cookie non-hex accepté")
	}
}
