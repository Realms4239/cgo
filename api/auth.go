package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"strings"
)

// sessionMsg — message figé du MAC de session ; versionné pour qu'une future
// rotation invalide les anciens cookies.
const sessionMsg = "cgo-session-v1"
const sessionCookie = "cgo_session"

// sessionToken — HMAC-SHA256(secret=CGO_API_TOKEN, msg=sessionMsg), hex.
// Déterministe : le même secret produit toujours le même cookie — un restart
// du service n'invalide pas les sessions (Max-Age=86400 côté client).
func sessionToken(secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(sessionMsg))
	return hex.EncodeToString(mac.Sum(nil))
}

// validSession — vérifie le cookie en temps constant ; secret vide ⇒ inerte.
// "Constante-ish" : subtle compare s'arrête sur mismatch de longueur, ce qui
// ne fuite que la longueur, jamais le contenu.
func validSession(cookieVal, secret string) bool {
	if secret == "" {
		return false // adaptateur inerte en mode dev — le skip loopback gère
	}
	want := sessionToken(secret)
	return subtle.ConstantTimeCompare([]byte(cookieVal), []byte(want)) == 1
}

// bearerOK — même comparaison, en-tête Authorization: Bearer.
func bearerOK(r *http.Request, want string) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return ok && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// authenticated — Bearer ou cookie session ; en l'absence de token configuré,
// le loopback seul passe (mode dev, comportement historique).
func authenticated(r *http.Request) bool {
	want := os.Getenv("CGO_API_TOKEN")
	if want == "" {
		return isLoopback(r.RemoteAddr)
	}
	if bearerOK(r, want) {
		return true
	}
	if c, err := r.Cookie(sessionCookie); err == nil && validSession(c.Value, want) {
		return true
	}
	return false
}

// requireAuth — mutations : Bearer ou cookie, sinon 401. Mode dev
// (CGO_API_TOKEN vide) : loopback seul passe, non-loopback reste 401.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authenticated(r) {
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeErr(w, r, "authentification requise", http.StatusUnauthorized)
	}
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
