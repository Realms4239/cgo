package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/figures"
)

// TestIntegrityCachedAndTagged — le bootstrap du corpus est mémoïsé (une seule
// passe de hachage) et expose un ETag stable : le cockpit ne paie pas le scan
// à chaque rendu. hash8 doit dériver du SHA complet, pas d'un second parcours.
func TestIntegrityCachedAndTagged(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-cache", header18,
		"run-cache,1,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	get := func(etag string) (*http.Response, string) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/integrity", nil)
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, string(b)
	}

	r1, body1 := get("")
	if r1.StatusCode != http.StatusOK {
		t.Fatalf("first status %d, body %s", r1.StatusCode, body1)
	}
	etag := r1.Header.Get("ETag")
	if etag == "" {
		t.Fatal("integrity response has no ETag")
	}
	if r1.Header.Get("Cache-Control") == "" {
		t.Fatal("integrity response has no Cache-Control")
	}
	var d map[string]any
	if err := json.Unmarshal([]byte(body1), &d); err != nil {
		t.Fatal(err)
	}
	sha, _ := d["sha256"].(string)
	h8, _ := d["hash8"].(string)
	if len(sha) != 64 || h8 != sha[:8] {
		t.Fatalf("hash8 %q must be first 8 of sha256 %q", h8, sha)
	}
	if want := figures.ProvenanceHash8(filepath.Join(dir, "data", "runs")); want != h8 {
		t.Fatalf("hash8 %q != figures.ProvenanceHash8 %q", h8, want)
	}

	r2, body2 := get("")
	if r2.StatusCode != http.StatusOK || body2 != body1 {
		t.Fatalf("cached payload drifted: %d / %s", r2.StatusCode, body2)
	}

	r3, _ := get(etag)
	if r3.StatusCode != http.StatusNotModified {
		t.Fatalf("If-None-Match status %d, want 304", r3.StatusCode)
	}

	if !strings.Contains(body1, `"available":true`) {
		t.Fatalf("integrity payload unavailable: %s", body1)
	}
}
