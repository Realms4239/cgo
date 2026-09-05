package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCompareEndpoint — MW bilatéral + Cliff valid-only, additif (anciens
// champs /api/results intacts : testé par les suites existantes).
func TestCompareEndpoint(t *testing.T) {
	dir := chdirTemp(t)
	// pfifo : latences hautes ; cake : basses, distribution disjointe.
	pfifo := []string{
		"run-c,1,P2,pfifo_fast,bbr,1,150.0,300.0,60.0,320.0,55.0,19.0,400,0,600000,60.0,0.0,valid",
		"run-c,2,P2,pfifo_fast,bbr,1,151.0,301.0,61.0,321.0,56.0,19.1,401,0,600001,60.1,0.0,valid",
		"run-c,3,P2,pfifo_fast,bbr,1,152.0,302.0,62.0,322.0,57.0,19.2,402,0,600002,60.2,0.0,valid",
		"run-c,4,P2,pfifo_fast,bbr,1,153.0,303.0,63.0,323.0,58.0,19.3,403,0,600003,60.3,0.0,valid",
		"run-c,5,P2,pfifo_fast,bbr,1,154.0,304.0,64.0,900.0,5.0,2.0,404,0,600004,60.4,0.0,invalid",
	}
	cake := []string{
		"run-c,6,P2,cake,bbr,1,100.0,114.0,14.0,209.0,98.0,18.5,200,0,300000,30.0,0.0,valid",
		"run-c,7,P2,cake,bbr,1,101.0,115.0,15.0,210.0,98.1,18.6,201,0,300001,30.1,0.0,valid",
		"run-c,8,P2,cake,bbr,1,102.0,116.0,16.0,211.0,98.2,18.7,202,0,300002,30.2,0.0,valid",
		"run-c,9,P2,cake,bbr,1,103.0,117.0,17.0,212.0,98.3,18.8,203,0,300003,30.3,0.0,valid",
	}
	writeCSVRun(t, dir, "run-c", header18, append(pfifo, cake...)...)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srv.URL + "/api/results/compare?a=P2|pfifo_fast|bbr&b=P2|cake|bbr&metric=small_p95_ms")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["available"] != true {
		t.Fatalf("available=%v (%v)", got["available"], got)
	}
	if got["n_a"] != float64(4) || got["n_b"] != float64(4) {
		t.Fatalf("n_a=%v n_b=%v want 4/4 (invalid exclu)", got["n_a"], got["n_b"])
	}
	if p, _ := got["mw_p_two_sided"].(float64); p >= 0.05 {
		t.Fatalf("mw p=%v want <0.05 (distributions disjointes)", p)
	}
	if got["cliff_interp"] != "grand" {
		t.Fatalf("cliff_interp=%v want grand", got["cliff_interp"])
	}
	// deadline : flag mixed-D présent.
	resp2, err := client.Get(srv.URL + "/api/results/compare?a=P2|pfifo_fast|bbr&b=P2|cake|bbr&metric=deadline_ok_pct")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var got2 map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&got2); err != nil {
		t.Fatal(err)
	}
	if got2["deadline_mixed_D"] != true {
		t.Fatalf("deadline_mixed_D=%v want true", got2["deadline_mixed_D"])
	}
}

// TestQuarantineFailedGates — passthrough additif + summary live.
func TestQuarantineFailedGates(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-q", header18,
		"run-q,1,P2,pfifo_fast,cubic,1,100.0,124.0,24.0,213.9,98.0,2.0,81,0,117288,12.10,0.0,invalid",
	)
	q := `[{"event_id":1,"profile":"P2","qdisc":"pfifo_fast","cc":"cubic","gate_status":"invalid","failed_gates":["G4ThroughputCoherent"]}]`
	if err := os.WriteFile(filepath.Join(dir, "data", "runs", "run-q", "quarantine.json"), []byte(q), 0644); err != nil {
		t.Fatal(err)
	}
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srv.URL + "/api/quarantine?run=run-q")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	qs, _ := got["quarantines"].([]any)
	if len(qs) != 1 {
		t.Fatalf("quarantines=%v want 1", got)
	}
	fg, _ := qs[0].(map[string]any)["failed_gates"].([]any)
	if len(fg) != 1 || fg[0] != "G4ThroughputCoherent" {
		t.Fatalf("failed_gates=%v want [G4ThroughputCoherent]", fg)
	}

	resp2, err := client.Get(srv.URL + "/api/quarantine/summary")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var sum map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	// P2 cap 20 : goodput 2.0 < 10 → g4_low=1.
	if sum["total_invalid"] != float64(1) || sum["g4_low"] != float64(1) {
		t.Fatalf("summary=%v want total=1 low=1", sum)
	}
}
