package api

import (
	"crypto/subtle"
	"net"
	"net/http"
	"os"
	"strings"
)

// requireAuth — mutations require a bearer token when CGO_API_TOKEN is set.
// Empty token is dev mode: loopback only, still 401 for non-loopback callers.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		want := os.Getenv("CGO_API_TOKEN")
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if want != "" && ok && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
			next(w, r)
			return
		}
		if want == "" && isLoopback(r.RemoteAddr) {
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
