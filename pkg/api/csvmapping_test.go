package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/model"
)

// Fixtures au schéma réel gelé (18 colonnes, qdi_ms en 8) — valeurs
// distinctives par colonne pour détecter toute lecture positionnelle.

const header18 = "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"

func chdirTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	return dir
}

func writeCSVRun(t *testing.T, dir, run, header string, rows ...string) {
	t.Helper()
	runDir := filepath.Join(dir, "data", "runs", run)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	out := header + "\n" + strings.Join(rows, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(out), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestReplayStreamReadsByName — le payload SSE doit porter small_p95_ms depuis
// la colonne small_p95_ms (328.2), pas qdi_ms (57.4) ; drops depuis drops
// (443), pas deadline_ok_pct (60.2).
func TestReplayStreamReadsByName(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-t", header18,
		"run-t,1,P2,pfifo_fast,bbr,1,111.0,168.4,57.4,328.2,60.2,18.2,443,5,641464,66.40,0.0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srv.URL + "/api/replay/stream?run=run-t")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{
		`"small_p95_ms":"328.2"`, `"bulk_goodput_mbps":"18.2"`,
		`"drops":"443"`, `"gate_status":"valid"`,
		`"rtt_p50_ms":"111.0"`, `"rtt_p95_ms":"168.4"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("replay payload missing %s in %s", want, s)
		}
	}
}

// TestReplayStreamRejectsTraversal — comme /api/run/rows : 400, pas de sortie
// de data/runs.
func TestReplayStreamRejectsTraversal(t *testing.T) {
	chdirTemp(t)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, bad := range []string{"../x", "..\\x", "run.x", ""} {
		resp, err := http.Get(srv.URL + "/api/replay/stream?run=" + bad)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("run=%q: status = %d, want 400", bad, resp.StatusCode)
		}
	}
}

// TestProfilesBuiltinNotImported — P4 est natif du binaire, pas importé.
// Un profil importé (PWT) porte imported=true ; nettoyé après le test pour
// ne pas fuir dans la carte globale.
func TestProfilesBuiltinNotImported(t *testing.T) {
	chdirTemp(t) // profile.Import persiste data/profiles.json relatif au CWD
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	get := func() string {
		resp, err := http.Get(srv.URL + "/api/profiles")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return string(body)
	}
	s := get()
	if !strings.Contains(s, `"id":"P4"`) {
		t.Fatalf("P4 missing from profiles: %s", s)
	}
	// aucun natif ne doit porter imported:true
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		marker := `"id":"` + id + `"`
		at := strings.Index(s, marker)
		if at < 0 {
			t.Fatalf("builtin %s missing from profiles", id)
		}
		seg := s[at:]
		if i := strings.Index(seg, "}"); i >= 0 && i < 400 {
			seg = seg[:i]
		}
		if strings.Contains(seg, `"imported":true`) {
			t.Fatalf("builtin %s flagged imported: %s", id, seg)
		}
	}

	req, _ := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader(`{"id":"PWT","capacity_mbps":9}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	s = get()
	if !strings.Contains(s, `"id":"PWT"`) || !strings.Contains(s, `"imported":true`) {
		t.Fatalf("imported PWT must carry imported:true: %s", s)
	}
	model.ProfilesMu.Lock()
	delete(model.Profiles, "PWT")
	model.ProfilesMu.Unlock()
}

// TestIntegrityBreakdown — détail par run + updated, triés du plus récent.
func TestIntegrityBreakdown(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-a", header18,
		"run-a,1,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
		"run-a,2,P2,cake,bbr,2,100,115,15,210.0,97.0,18.0,82,0,117300,12.2,0.0,invalid",
	)
	writeCSVRun(t, dir, "run-b", header18,
		"run-b,1,P1,pfifo_fast,cubic,1,20,25,5,30.0,100,70.0,0,0,0,0,0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/integrity")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{`"breakdown"`, `"run":"run-b"`, `"run":"run-a"`, `"rows":2`, `"quarantined":1`, `"updated":"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("integrity missing %s in %s", want, s)
		}
	}
	if strings.Index(s, "run-b") > strings.Index(s, "run-a") {
		t.Fatalf("breakdown must list newest run first: %s", s)
	}
}

// (300→330 = +10 %), pas sur qdi_ms (50→40 = −20 %).
func TestResultsDeltaReadsByName(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-a", header18,
		"run-a,1,P2,cake,bbr,1,140.0,150.0,50.0,300.0,90.0,10.0,20,0,100000,10.0,0.0,valid",
	)
	writeCSVRun(t, dir, "run-b", header18,
		"run-b,1,P2,cake,bbr,1,150.0,160.0,40.0,330.0,91.0,11.0,22,0,110000,11.0,0.0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/results/delta?cell=P2|cake|bbr")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{
		`"available":true`, `"small_p95_pct":10`, `"goodput_pct":10`, `"rtt_p95_pct":7`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("delta missing %s in %s", want, s)
		}
	}
}
