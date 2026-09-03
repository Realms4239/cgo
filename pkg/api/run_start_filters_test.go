package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// TestRunStartSubmatrixFilters — axes qdiscs/ccs validés à l'entrée (même
// sans moteur câblé : prévention §6), paire filtrée ensemble, matrice
// pleine quand les deux sont absents.
func TestRunStartSubmatrixFilters(t *testing.T) {
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

	if code := post(`{"profiles":["P2"],"reps":1}`); code != http.StatusOK {
		t.Fatalf("matrice pleine: %d, want 200", code)
	}
	if code := post(`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["bbr"],"reps":1}`); code != http.StatusOK {
		t.Fatalf("sous-matrice: %d, want 200", code)
	}
	if code := post(`{"profiles":["P2"],"qdiscs":["hfsc"],"ccs":["bbr"],"reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("qdisc inconnu: %d, want 400", code)
	}
	if code := post(`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["westwood"],"reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("cc inconnue: %d, want 400", code)
	}
	if code := post(`{"profiles":["P2"],"qdiscs":["cake"],"reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("qdisc sans cc: %d, want 400", code)
	}
	if code := post(`{"profiles":["P2"],"ccs":["bbr"],"reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("cc sans qdisc: %d, want 400", code)
	}
}

// TestRunOptsReachStartFn — les filtres arrivent intacts au moteur (le
// câblage serveur les matérialise en matrice pleine quand vides).
func TestRunOptsReachStartFn(t *testing.T) {
	var got RunOpts
	h := New(Deps{StartFn: func(o RunOpts) error { got = o; return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/run/start", "application/json",
		strings.NewReader(`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["cubic"],"reps":2}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if len(got.Qdiscs) != 1 || got.Qdiscs[0] != "cake" || len(got.CCs) != 1 || got.CCs[0] != "cubic" {
		t.Fatalf("filtres perdus: %+v", got)
	}
	_ = model.AllQdiscs // garde l'import si le test évolue
}
