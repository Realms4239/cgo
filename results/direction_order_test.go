package results

import (
	"testing"
)

// TestScanDirectionOrderStable — deux groupes de la MÊME cellule ne différant
// que par la direction sortent TOUJOURS dans le même ordre (up d'abord) :
// le constat UI prend le premier pfifo via find — l'itération aléatoire des
// maps Go + sort.Slice sans tiebreak mélangeait les lignes liées à chaque
// appel /api/results, et le verdict basculait (-37 % / -65 %) entre deux
// chargements de la même vue (prouvé : 6 appels, ordres mélangés).
func TestScanDirectionOrderStable(t *testing.T) {
	dir := t.TempDir()
	header := "run_id,event_id,profile,qdisc,cc,direction,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,voip_r,jfi_pct,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,rtt_base_p50_ms,rtt_base_p95_ms,gate_status"
	writeRun(t, dir, "run-d", header,
		"run-d,1,P2,pfifo_fast,bbr,down,1,101,127,26,91.4,0,582.1,33.1,0.8,22,0,31856,3.30,0.0,110.0,126.0,valid",
		"run-d,2,P2,pfifo_fast,bbr,up,1,100,168,57,91.4,0,328.2,60.2,18.2,443,0,641464,66.40,0.0,98.3,113.0,valid",
	)
	for i := 0; i < 30; i++ {
		groups, err := Scan(dir, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 2 {
			t.Fatalf("itération %d: want 2 groupes, got %d", i, len(groups))
		}
		if groups[0].Direction != "up" || groups[1].Direction != "down" {
			t.Fatalf("itération %d: ordre instable, want up,down got %s,%s",
				i, groups[0].Direction, groups[1].Direction)
		}
	}
}
