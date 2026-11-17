package api

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// gzipResponseWriter — compresse le corps à la volée. Content-Length est
// supprimé au moment du WriteHeader : http.FileServer le calcule sur le brut,
// ce qui serait faux après compression.
type gzipResponseWriter struct {
	http.ResponseWriter
	gw          *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.gw.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(code)
}

// compressResponses — le serveur embarque la SPA et sert ses JSON sur LE LIEN
// DE L'OPÉRATEUR, souvent contraint. Sans ça, 380 Ko de JS et 25 Ko de JSON
// traversent le lien en clair (mesuré 2026-11-17, aucun Content-Encoding).
// Les flux SSE sont exclus : gzip les mettrait en tampon et détruirait le
// 10 Hz temps réel. Les polices woff2 et figures sont déjà compressées.
func compressResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasPrefix(p, "/assets/") {
			// Noms de fichiers hachés par Vite : cache immuable sans risque.
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		if !compressible(p) || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gw: gw}, r)
	})
}

func compressible(path string) bool {
	// Flux temps réel : jamais de tampon.
	if strings.Contains(path, "/stream") {
		return false
	}
	if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/api/") || path == "/" {
		return true
	}
	// woff2/woff, png/jpg/svg déjà compressés ou minuscules.
	return false
}
