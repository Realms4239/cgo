package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const duelHeader = "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"

func writeDuelRun(t *testing.T, dir, run, qdisc, cc string, reps int, small, deadline float64, gate string) {
	t.Helper()
	runDir := filepath.Join(dir, run)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := duelHeader + "\n"
	for i := 1; i <= reps; i++ {
		body += run + "," + itoa(i) + ",P2," + qdisc + "," + cc + "," + itoa(i) +
			",20,25,3," + ftoa(small) + "," + ftoa(deadline) + ",18,0,0,1000,10,30," + gate + "\n"
	}
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func itoa(i int) string     { return strconv.Itoa(i) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

func TestDuelPicksBestAndBaseWithGuardrail(t *testing.T) {
	dir := t.TempDir()
	writeDuelRun(t, dir, "run-a", "pfifo_fast", "bbr", 3, 328.2, 60.2, "valid")
	writeDuelRun(t, dir, "run-b", "fq_codel", "bbr", 3, 205.4, 96.2, "valid")
	writeDuelRun(t, dir, "run-c", "cake", "bbr", 1, 100, 100, "valid") // n=1 : exclu

	old := duelDir
	duelDir = dir
	t.Cleanup(func() { duelDir = old })

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/duel?profile=P2")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got duelResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Base == nil || got.Base.Qdisc != "pfifo_fast" {
		t.Fatalf("base = %+v", got.Base)
	}
	if got.Best == nil || got.Best.Qdisc != "fq_codel" {
		t.Fatalf("la cellule n=1 doit rester inéligible : %+v", got.Best)
	}
	if got.Best.N < 3 || got.MinReps != 3 {
		t.Fatalf("garde d'échantillonnage : %+v", got)
	}
	if len(got.Best.Runs) == 0 {
		t.Fatal("le duel doit citer ses runs")
	}
}

func TestDuelRefusesUnknownProfile(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/duel?profile=../../etc")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, voulu 400", resp.StatusCode)
	}
}
