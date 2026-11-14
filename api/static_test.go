package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// TestStaticFSServesSPAAndNilIsAnHonest404 — la couture SPA : le FS injecté
// sert le vrai bundle ; nil répond 404 en français, jamais un stub muet.
func TestStaticFSServesSPAAndNilIsAnHonest404(t *testing.T) {
	h := New(Deps{StaticFS: fstest.MapFS{
		"index.html":    {Data: []byte(`<!doctype html><html lang="fr"><div id="root"></div></html>`)},
		"assets/app.js": {Data: []byte("console.log('spa')")},
	}})
	defer h.CloseHub()
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `id="root"`) {
		t.Fatalf("/: status %d, body %q", resp.StatusCode, body)
	}

	resp, err = srv.Client().Get(srv.URL + "/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/assets/app.js: status %d", resp.StatusCode)
	}

	nilFS := New(Deps{})
	defer nilFS.CloseHub()
	srvNil := httptest.NewServer(nilFS)
	defer srvNil.Close()
	resp, err = srvNil.Client().Get(srvNil.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound || !strings.Contains(string(body), "SPA non embarqué") {
		t.Fatalf("nil StaticFS: status %d, body %q", resp.StatusCode, body)
	}
}
