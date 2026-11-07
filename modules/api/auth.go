package api

import (
	"net/http"
	"os"
	"strings"
)

// requireAuth — mutations require a bearer token when CGO_API_TOKEN is set.
// Empty token is dev mode: loopback only, still 401 for non-loopback callers.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		want := os.Getenv("CGO_API_TOKEN")
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if want != "" && got == want {
			next(w, r)
			return
		}
		if want == "" && strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeErr(w, r, "authentification requise", http.StatusUnauthorized)
	}
}
