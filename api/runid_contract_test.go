package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Contrat des identifiants de run côté HTTP — les sémantiques « vide = tous
// les runs » et « non-vide = élément de chemin » sont épinglées ici pour
// qu'aucun refactor ne les rabote silencieusement.

func TestRunIDEndpointContract(t *testing.T) {
	dir := chdirTemp(t)
	header := "run_id,event_id,profile,qdisc,cc,direction,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,voip_r,jfi_pct,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,rtt_base_p50_ms,rtt_base_p95_ms,gate_status"
	writeCSVRun(t, dir, "run-a", header,
		"run-a,1,P2,cake,bbr,up,1,100,114,14,91.4,0,200.0,98.0,18.0,81,0,117288,12.1,0.0,98.3,113.0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	get := func(path string) (int, string) {
		t.Helper()
		resp, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}

	// ?run= vide reste un cas légitime : tous les runs (200), jamais 400.
	if code, _ := get("/api/quarantine"); code != 200 {
		t.Fatalf("quarantine tous les runs : %d, veut 200", code)
	}
	// un identifiant bien formé est accepté
	if code, _ := get("/api/quarantine?run=run-a"); code != 200 {
		t.Fatalf("quarantine run-a : %d, veut 200", code)
	}
	// traversal, séparateur Windows, point, contrôle : refusés en 400
	for _, q := range []string{"..", "a/b", `a\b`, "a.b", ".", "..%2F"} {
		if code, _ := get("/api/quarantine?run=" + q); code != 400 {
			t.Fatalf("quarantine %q : %d, veut 400", q, code)
		}
	}
	// un identifiant bien formé mais inexistant : 404 (pas 400, pas 200)
	if code, _ := get("/api/quarantine?run=nope"); code != 404 {
		t.Fatalf("quarantine run inconnu : %d, veut 404", code)
	}

	// POST /api/run/start avec un id hostile : refus avant tout lancement
	if code, _ := get("/api/run/rows?run=.."); code != 400 {
		t.Fatalf("run/rows traversal : %d, veut 400", code)
	}
}

func TestRunStartRejectsHostileRunID(t *testing.T) {
	dir := chdirTemp(t)
	_ = dir
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, hostile := range []string{"..", "../freeze", `..\freeze`, "a/b", `a\b`, ".", "a.b"} {
		body, _ := json.Marshal(map[string]any{
			"profiles": []string{"P2"}, "reps": 1, "channel_model": "clean",
			"deadline_ms": 1000, "target": "127.0.0.1", "run_id": hostile,
		})
		req, err := http.NewRequest("POST", srv.URL+"/api/run/start", strings.NewReader(string(body)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 400 {
			t.Fatalf("run/start id %q : %d, veut 400", hostile, resp.StatusCode)
		}
		// le refus nomme l'identifiant — jamais un 400 d'une autre cause
		if !strings.Contains(string(b), "id de run invalide") {
			t.Fatalf("run/start id %q : %q — le refus doit nommer l'identifiant", hostile, string(b))
		}
	}
}
