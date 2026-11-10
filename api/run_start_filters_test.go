package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
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
	if code := post(`{"profiles":["P2"],"qdiscs":["none"],"ccs":["bbr"],"reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("qdisc none hors matrice: %d, want 400", code)
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

// TestRunStartAcceptsBBRv3Label — l'API valide des libellés, pas la
// disponibilité : bbr3 atteint le moteur, qui le refuse avec le chemin
// d'installation si le noyau ne l'expose pas (jamais un repli silencieux).
func TestRunStartAcceptsBBRv3Label(t *testing.T) {
	var got RunOpts
	srv := httptest.NewServer(New(Deps{StartFn: func(o RunOpts) error { got = o; return nil }}))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/run/start", "application/json",
		strings.NewReader(`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["bbr3"],"reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bbr3 étiquette connue: %d, want 200", resp.StatusCode)
	}
	if len(got.CCs) != 1 || got.CCs[0] != "bbr3" {
		t.Fatalf("bbr3 perdu avant le moteur: %+v", got.CCs)
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

// TestRunStartProvenanceValidation — bench et channel_model validés à
// l'entrée (prévention §6 : le client est une indication, jamais le
// contrat), bursty refusé hors P2/P3 AVANT la porte moteur.
func TestRunStartProvenanceValidation(t *testing.T) {
	h := New(Deps{StartFn: func(o RunOpts) error { return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(body string) (int, string) {
		resp, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp.StatusCode, string(b)
	}

	// bursty + P1 ⇒ 400 qui nomme le profil fautif et la règle P2/P3
	code, body := post(`{"profiles":["P1"],"channel_model":"bursty","reps":1}`)
	if code != http.StatusBadRequest {
		t.Fatalf("bursty+P1: %d, want 400", code)
	}
	if !strings.Contains(body, "P1") || !strings.Contains(body, "P2/P3") {
		t.Fatalf("bursty+P1: le refus doit nommer P1 et la règle, got %s", body)
	}
	// canal inconnu ⇒ 400 nommant la règle
	if code, _ = post(`{"profiles":["P2"],"channel_model":"tlmn","reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("canal inconnu: %d, want 400", code)
	}
	// banc inconnu ⇒ 400
	if code, _ = post(`{"profiles":["P2"],"bench":"xyz","reps":1}`); code != http.StatusBadRequest {
		t.Fatalf("banc inconnu: %d, want 400", code)
	}
	// bursty + P2 sans moteur câblé ⇒ 503 (la validation provenance passe,
	// la porte moteur arrive après — l'ordre des portes est le contrat)
	noEngine := New(Deps{})
	srv2 := httptest.NewServer(noEngine)
	defer srv2.Close()
	resp, err := http.Post(srv2.URL+"/api/run/start", "application/json",
		bytes.NewBufferString(`{"profiles":["P2"],"channel_model":"bursty","reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("bursty+P2 sans moteur: %d, want 503 (gate order)", resp.StatusCode)
	}
}

// TestRunStartRefusesEscapingRunID — le run_id est un chemin sous data/runs :
// un identifiant qui en sort (../, /, \\) ferait écrire OpenRun hors du dossier.
// Le reste de l'API valide déjà chaque identifiant via results.ValidRunID.
func TestRunStartRefusesEscapingRunID(t *testing.T) {
	called := false
	h := New(Deps{StartFn: func(o RunOpts) error { called = true; return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, id := range []string{"../evade", `..\evade`, "a/b", "run.1", "."} {
		body, err := json.Marshal(map[string]any{"profiles": []string{"P2"}, "reps": 1, "run_id": id})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("run_id %q: %d, want 400", id, resp.StatusCode)
		}
	}
	if called {
		t.Fatal("un run_id invalide a atteint le moteur")
	}
}

// TestRunStartProvenanceReachesStartFn — la provenance déclarée (banc,
// canal, flux) arrive intacte au moteur ; flows absent passe tel quel (0),
// le câblage cmd/cgo appliquant le défaut 4 (décision 6).
func TestRunStartProvenanceReachesStartFn(t *testing.T) {
	var got RunOpts
	h := New(Deps{StartFn: func(o RunOpts) error { got = o; return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/run/start", "application/json",
		strings.NewReader(`{"profiles":["P2"],"channel_model":"bursty","bench":"dualns","flows":3,"reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if got.ChannelModel != "bursty" || got.Bench != "dualns" || got.Flows != 3 {
		t.Fatalf("provenance altérée: %+v", got)
	}

	var gotNoFlows RunOpts
	h2 := New(Deps{StartFn: func(o RunOpts) error { gotNoFlows = o; return nil }})
	srv2 := httptest.NewServer(h2)
	defer srv2.Close()
	resp2, err := http.Post(srv2.URL+"/api/run/start", "application/json",
		strings.NewReader(`{"profiles":["P2"],"reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("flows absent: %d, want 200", resp2.StatusCode)
	}
	if gotNoFlows.Flows != 0 {
		t.Fatalf("flows absent doit atteindre le moteur tel quel (0), got %d", gotNoFlows.Flows)
	}
}
