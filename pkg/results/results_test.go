package results

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAggregates(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, "run-test")
	os.MkdirAll(runDir, 0755)
	csv := "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status\n" +
		"run-test,1,P2,pfifo_fast,cubic,1,20,25,30,100,15,0,0,0,0,10,valid\n" +
		"run-test,2,P2,pfifo_fast,cubic,2,21,26,31,100,16,0,0,0,0,11,valid\n" +
		"run-test,3,P2,fq_codel,cubic,1,15,18,20,100,14,0,0,0,0,10,valid\n"
	os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(csv), 0644)
	groups, err := Scan(dir, "")
	if err != nil || len(groups) != 2 {
		t.Fatalf("groups=%v err=%v", groups, err)
	}
	// fq_codel should be best (small p95 20 < 30)
	foundBest := false
	for _, g := range groups {
		if g.Best && g.Qdisc == "fq_codel" {
			foundBest = true
		}
	}
	if !foundBest {
		t.Fatalf("best not marked: %+v", groups)
	}
}
