package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLoginSetsHttpOnlyCookieAndCookieGrantsAccess — le flux navigateur
// complet : login avec le token → cookie session HttpOnly → les requêtes
// same-origin suivantes (y compris l'EventSource sur /api/stream) passent
// sans en-tête Bearer.
func TestLoginSetsHttpOnlyCookieAndCookieGrantsAccess(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	body, _ := json.Marshal(map[string]string{"token": "token-secret-abc"})
	resp, err := client.Post(srv.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login : %d", resp.StatusCode)
	}
	if bytes.Contains(respBody, []byte("token-secret-abc")) {
		t.Fatal("le token est échoé dans le corps de la réponse de login")
	}
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "cgo_session" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("aucun cookie cgo_session posé au login")
	}
	if !cookie.HttpOnly {
		t.Fatal("cookie non HttpOnly")
	}
	if cookie.Path != "/" {
		t.Fatalf("cookie Path=%q, want /", cookie.Path)
	}
	if cookie.MaxAge != 86400 {
		t.Fatalf("cookie Max-Age=%d, want 86400", cookie.MaxAge)
	}
	for _, line := range strings.Split(resp.Header.Get("Set-Cookie"), ";") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SameSite=") && line != "SameSite=Strict" {
			t.Fatalf("SameSite=%q, want Strict", line)
		}
	}
	if !strings.Contains(resp.Header.Get("Set-Cookie"), "SameSite=Strict") {
		t.Fatalf("SameSite=Strict absent du Set-Cookie : %q", resp.Header.Get("Set-Cookie"))
	}
	// http de test : le drapeau Secure ne pose que sous TLS réel
	if strings.Contains(resp.Header.Get("Set-Cookie"), "Secure") {
		t.Fatal("Secure posé hors TLS")
	}

	// lecture avec le cookie, sans Bearer
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/cost/tiers", nil)
	req.AddCookie(cookie)
	r2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	r2.Body.Close()
	if r2.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/cost/tiers avec cookie : %d, want 200", r2.StatusCode)
	}

	// EventSource sur /api/stream avec le cookie, sans Bearer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream", nil)
	sreq.AddCookie(cookie)
	sresp, err := client.Do(sreq)
	if err != nil {
		t.Fatal(err)
	}
	if sresp.StatusCode != http.StatusOK {
		sresp.Body.Close()
		t.Fatalf("GET /api/stream avec cookie : %d, want 200", sresp.StatusCode)
	}
	if ct := sresp.Header.Get("Content-Type"); ct != "text/event-stream" {
		sresp.Body.Close()
		t.Fatalf("Content-Type=%q, want text/event-stream", ct)
	}
	sresp.Body.Close()

	// sans cookie : 401 (token configuré)
	r3, err := client.Get(srv.URL + "/api/cost/tiers")
	if err != nil {
		t.Fatal(err)
	}
	r3.Body.Close()
	if r3.StatusCode != http.StatusUnauthorized {
		t.Fatalf("GET /api/cost/tiers sans credential : %d, want 401", r3.StatusCode)
	}
}

// TestLoginWrongTokenIs401 — mauvais token : 401, jamais de cookie, jamais
// le token dans le corps.
func TestLoginWrongTokenIs401(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	body, _ := json.Marshal(map[string]string{"token": "mauvais-token"})
	resp, err := http.Post(srv.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("login mauvais token : %d, want 401", resp.StatusCode)
	}
	if bytes.Contains(respBody, []byte("token-secret-abc")) || bytes.Contains(respBody, []byte("mauvais-token")) {
		t.Fatal("un token est échoé dans le corps d'erreur")
	}
	if strings.Contains(resp.Header.Get("Set-Cookie"), "cgo_session") {
		t.Fatal("cookie posé sur un login refusé")
	}
}

// TestLogoutClearsCookie — la déconnexion repose le cookie avec Max-Age=0
// et les mêmes drapeaux.
func TestLogoutClearsCookie(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/logout", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logout : %d, want 200", resp.StatusCode)
	}
	sc := resp.Header.Get("Set-Cookie")
	if !strings.Contains(sc, "cgo_session=") || !strings.Contains(sc, "Max-Age=0") {
		t.Fatalf("logout ne efface pas le cookie : %q", sc)
	}
	if !strings.Contains(sc, "HttpOnly") || !strings.Contains(sc, "SameSite=Strict") {
		t.Fatalf("logout : drapeaux perdus : %q", sc)
	}
}

// TestHealthStaysAnonymous — la sonde de vie répond sans credential, même
// avec un token configuré et un appelant non-loopback.
func TestHealthStaysAnonymous(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	h := New(Deps{Version: "test"})
	req := httptest.NewRequest("GET", "http://example.com/api/health", nil)
	req.RemoteAddr = "10.9.8.7:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/health non-loopback : %d, want 200", rec.Code)
	}
}

// TestReadRequiresAuthWhenTokenSet — avec un token configuré, une lecture
// /api/* non authentifiée est refusée (401) pour un appelant non-loopback.
func TestReadRequiresAuthWhenTokenSet(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	h := New(Deps{})
	for _, path := range []string{"/api/cost/tiers", "/api/schema", "/api/stream", "/api/events", "/api/state"} {
		req := httptest.NewRequest("GET", "http://example.com"+path, nil)
		req.RemoteAddr = "10.9.8.7:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s non-loopback sans credential : %d, want 401", path, rec.Code)
		}
	}
}

// TestMutationRequiresAuthWhenTokenSet — un POST non authentifié est refusé
// (401) pour un appelant non-loopback, quel que soit le handler ciblé.
func TestMutationRequiresAuthWhenTokenSet(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	h := New(Deps{})
	for _, path := range []string{"/api/burst", "/api/shape", "/api/run/stop"} {
		req := httptest.NewRequest("POST", "http://example.com"+path, nil)
		req.RemoteAddr = "10.9.8.7:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("POST %s non-loopback sans credential : %d, want 401", path, rec.Code)
		}
	}
}

// TestDevModeStaysAnonymousOnLoopback — mode dev (pas de token) : loopback
// passe, non-loopback est refusé (les lectures comme les mutations).
func TestDevModeStaysAnonymousOnLoopback(t *testing.T) {
	os.Unsetenv("CGO_API_TOKEN")
	h := New(Deps{})
	getLoop := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	getLoop.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, getLoop)
	if rec.Code != http.StatusOK {
		t.Fatalf("dev loopback GET : %d, want 200", rec.Code)
	}
	getRemote := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	getRemote.RemoteAddr = "10.9.8.7:1234"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, getRemote)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("dev non-loopback GET : %d, want 401", rec2.Code)
	}
	postRemote := httptest.NewRequest("POST", "http://example.com/api/burst", nil)
	postRemote.RemoteAddr = "10.9.8.7:1234"
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, postRemote)
	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("dev non-loopback POST : %d, want 401", rec3.Code)
	}
}

// TestCookieGrantsAccessNonLoopback — le cookie de session authentifie même
// hors loopback (cas proxy / alias).
func TestCookieGrantsAccessNonLoopback(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	h := New(Deps{})
	cookie := sessionToken("token-secret-abc")
	req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	req.RemoteAddr = "10.9.8.7:1234"
	req.AddCookie(&http.Cookie{Name: "cgo_session", Value: cookie})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET avec cookie non-loopback : %d, want 200", rec.Code)
	}
}

// TestCookieAdapterInertInDevMode — sans token configuré, le cookie est
// ignoré : le non-loopback reste refusé même avec un cookie forgé.
func TestCookieAdapterInertInDevMode(t *testing.T) {
	os.Unsetenv("CGO_API_TOKEN")
	h := New(Deps{})
	req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
	req.RemoteAddr = "10.9.8.7:1234"
	req.AddCookie(&http.Cookie{Name: "cgo_session", Value: strings.Repeat("ab", 32)})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("cookie forgé en mode dev non-loopback : %d, want 401", rec.Code)
	}
}

// TestAuthAdaptersRejectWeirdInput — les chemins de comparaison ne panissent
// pas sur entrée bizarre (longueur, encodage, vide).
func TestAuthAdaptersRejectWeirdInput(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	h := New(Deps{})
	weird := []string{"", "x", strings.Repeat("zz", 64), "cgo-session-v1", "\x00\x01\x02", strings.Repeat("ab", 32)}
	for _, v := range weird {
		req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
		req.RemoteAddr = "10.9.8.7:1234"
		req.AddCookie(&http.Cookie{Name: "cgo_session", Value: v})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("cookie %q : %d, want 401", v, rec.Code)
		}
	}
	for _, v := range weird {
		req := httptest.NewRequest("GET", "http://example.com/api/cost/tiers", nil)
		req.RemoteAddr = "10.9.8.7:1234"
		req.Header.Set("Authorization", "Bearer "+v)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("bearer %q : %d, want 401", v, rec.Code)
		}
	}
}

// TestLoginWithoutTokenConfiguredIs401 — en mode dev (aucun token configuré),
// le login est refusé : il n'y a pas de secret à vérifier.
func TestLoginWithoutTokenConfiguredIs401(t *testing.T) {
	os.Unsetenv("CGO_API_TOKEN")
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	body, _ := json.Marshal(map[string]string{"token": "nimportequoi"})
	resp, err := http.Post(srv.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("login en mode dev : %d, want 401", resp.StatusCode)
	}
}
