package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompressResponses(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", "13") // volontairement faux après compression
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	h := compressResponses(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	if got := rec.Header().Get("Vary"); !strings.Contains(got, "Accept-Encoding") {
		t.Fatalf("Vary = %q", got)
	}
	if got := rec.Header().Get("Content-Length"); got != "" {
		t.Fatalf("Content-Length must be stripped, got %q", got)
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(zr)
	if string(body) != `{"ok":true}` {
		t.Fatalf("round-trip body = %q", body)
	}
}

func TestCompressResponsesRespectsSSEAndClient(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("event: snapshot\ndata: {}\n\n"))
	})
	h := compressResponses(inner)
	for _, tc := range []struct {
		name, path, enc string
	}{
		{"sse stream never gzipped", "/api/stream", "gzip"},
		{"replay stream never gzipped", "/api/replay/stream", "gzip"},
		{"no client support", "/api/health", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.enc != "" {
				req.Header.Set("Accept-Encoding", tc.enc)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if got := rec.Header().Get("Content-Encoding"); got != "" {
				t.Fatalf("Content-Encoding = %q, want none", got)
			}
		})
	}
}

func TestCompressResponsesImmutableAssets(t *testing.T) {
	h := compressResponses(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("console.log(1)"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/assets/index-abc123.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "immutable") || !strings.Contains(got, "31536000") {
		t.Fatalf("Cache-Control = %q, want immutable 1-year", got)
	}
}

func TestCompressibleMatrix(t *testing.T) {
	yes := []string{"/", "/api/integrity", "/api/results", "/assets/index-a.js"}
	no := []string{"/api/stream", "/api/replay/stream", "/fonts/inter-var.woff2", "/textures/noise.png", "/favicon.svg"}
	for _, p := range yes {
		if !compressible(p) {
			t.Errorf("compressible(%q) = false, want true", p)
		}
	}
	for _, p := range no {
		if compressible(p) {
			t.Errorf("compressible(%q) = true, want false", p)
		}
	}
}
