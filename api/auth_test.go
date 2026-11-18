package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// withDevNoAuth — opt-in dev explicite, avec retour à l'état strict en fin de
// test. Les tests fonctionnels qui appellent le handler en anonyme sur
// loopback dépendent de ce chemin ; les tests du jeton, eux, posent leur propre
// CGO_API_TOKEN et leurs credentials.
func withDevNoAuth(t *testing.T) {
	t.Helper()
	SetDevNoAuth(true)
	t.Cleanup(func() { SetDevNoAuth(false) })
}

// TestLoopbackWithoutTokenOrDevFlagIs401 — sans jeton ni opt-in dev, le
// loopback ne passe plus : le contournement implicite a été supprimé.
func TestLoopbackWithoutTokenOrDevFlagIs401(t *testing.T) {
	SetDevNoAuth(false)
	defer SetDevNoAuth(false)
	os.Unsetenv("CGO_API_TOKEN")
	h := New(Deps{})
	req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("loopback sans jeton ni dev : %d, want 401", rec.Code)
	}
}

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

// TestSessionKeyIsRandomAndRevocable — clé par démarrage : deux sessions
// émises par le même processus diffèrent, une session révoquée ne repasse
// jamais, et révoquer l'une laisse les autres valides.
func TestSessionKeyIsRandomAndRevocable(t *testing.T) {
	a := newSessionToken()
	b := newSessionToken()
	if a == b {
		t.Fatal("two sessions must differ")
	}
	if !validSession(a) {
		t.Fatal("fresh session must validate")
	}
	revokeSession(a)
	if validSession(a) {
		t.Fatal("revoked session must not validate")
	}
	if !validSession(b) {
		t.Fatal("revoking one session must not revoke another")
	}
}

// TestSessionKeyIsPerBoot — simuler un restart (changement de clé) invalide
// les jetons émis par l'instance précédente.
func TestSessionKeyIsPerBoot(t *testing.T) {
	token := newSessionToken()
	old := bootKey
	defer func() { bootKey = old }()
	bootKey = [32]byte{0xFF}
	if validSession(token) {
		t.Fatal("token from a previous boot must not validate")
	}
}

// TestValidSessionRejectsTamperedAndMalformed — oracle hostile : signature
// altérée, jeton tronqué, entrée vide ou non-hex sont refusés.
func TestValidSessionRejectsTamperedAndMalformed(t *testing.T) {
	good := newSessionToken()
	jti, sig, ok := strings.Cut(good, ".")
	if !ok {
		t.Fatal("session token must contain a separator")
	}
	badSig := "0" + sig[1:]
	if badSig == sig {
		badSig = "1" + sig[1:]
	}
	if validSession(jti + "." + badSig) {
		t.Fatal("tampered signature must not validate")
	}
	if validSession(jti) {
		t.Fatal("token without signature must not validate")
	}
	if validSession("") {
		t.Fatal("empty token must not validate")
	}
	if validSession("zz") {
		t.Fatal("non-hex token must not validate")
	}
	if validSession(strings.Repeat("a", 64)) {
		t.Fatal("forged token must not validate")
	}
}

// TestLogoutRevokesSessionCookie — la route /api/auth/logout révoque le cookie
// côté serveur ; Bearer reste valable après la révocation.
func TestLogoutRevokesSessionCookie(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "logout-secret")
	h := New(Deps{})
	token := newSessionToken()

	req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	req.RemoteAddr = "10.9.8.7:1234"
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cookie before logout: %d, want 200", rec.Code)
	}

	out := httptest.NewRequest("POST", "http://example.com/api/auth/logout", nil)
	out.RemoteAddr = "10.9.8.7:1234"
	out.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	outRec := httptest.NewRecorder()
	h.ServeHTTP(outRec, out)
	if outRec.Code != http.StatusOK {
		t.Fatalf("logout: %d, want 200", outRec.Code)
	}

	after := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	after.RemoteAddr = "10.9.8.7:1234"
	after.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	afterRec := httptest.NewRecorder()
	h.ServeHTTP(afterRec, after)
	if afterRec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked cookie after logout: %d, want 401", afterRec.Code)
	}

	bearer := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	bearer.RemoteAddr = "10.9.8.7:1234"
	bearer.Header.Set("Authorization", "Bearer logout-secret")
	bearerRec := httptest.NewRecorder()
	h.ServeHTTP(bearerRec, bearer)
	if bearerRec.Code != http.StatusOK {
		t.Fatalf("bearer after logout: %d, want 200", bearerRec.Code)
	}
}
