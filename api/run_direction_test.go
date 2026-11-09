package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRunStartDirection — le sens de charge voyage dans POST /api/run/start :
// up (défaut/vide), down, both (RRUL séquentiel). Valeur inconnue = 400
// (validé à l'entrée même sans moteur câblé — prévention §6).
func TestRunStartDirection(t *testing.T) {
	h := New(Deps{StartFn: func(o RunOpts) error { return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(body string) int {
		resp, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	for _, body := range []string{
		`{"profiles":["P2"],"reps":1}`,
		`{"profiles":["P2"],"reps":1,"direction":"up"}`,
		`{"profiles":["P2"],"reps":1,"direction":"down"}`,
		`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["bbr"],"reps":1,"direction":"both"}`,
	} {
		if code := post(body); code != http.StatusOK {
			t.Fatalf("%s: %d, want 200", body, code)
		}
	}
	if code := post(`{"profiles":["P2"],"reps":1,"direction":"sideways"}`); code != http.StatusBadRequest {
		t.Fatalf("direction inconnue: %d, want 400", code)
	}
}
