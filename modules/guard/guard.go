// Package guard — bibliothèque d'authentification par jeton Bearer.
// Aucune mutation n'est autorisée sans jeton sur un bind non-loopback.
package guard

import (
	"net/http"
	"os"
	"strings"
)

// RequireAuth protège un handler : le jeton attendu vient de CGO_API_TOKEN.
// Jeton vide = mode dev loopback uniquement ; les appelants non-loopback
// reçoivent 401, comme le serveur historique.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuthToken(os.Getenv("CGO_API_TOKEN"), next)
}

// RequireAuthToken protège un handler avec un jeton explicite.
// Même règle : jeton vide = loopback seulement.
func RequireAuthToken(want string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if want != "" && got == want {
			next(w, r)
			return
		}
		if want == "" && isLoopback(r.RemoteAddr) {
			next(w, r)
			return
		}
		// 401 avec défi Bearer, message en français comme le serveur.
		w.Header().Set("WWW-Authenticate", "Bearer")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"authentification requise"}`))
	}
}

// isLoopback teste l'adresse distante sans résoudre de DNS.
func isLoopback(addr string) bool {
	return strings.HasPrefix(addr, "127.0.0.1") || strings.HasPrefix(addr, "[::1]")
}
