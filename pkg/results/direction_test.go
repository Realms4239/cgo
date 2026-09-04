package results

import (
	"testing"
)

// TestScanSkipsSmokeFixture — run-smoke* exclu du corpus (parité JS) :
// sa ligne valid factice ne doit faire médiane nulle part.
func TestScanSkipsSmokeFixture(t *testing.T) {
	dir := t.TempDir()
	writeRun(t, dir, "run-smoke-x", header18,
		"run-smoke-x,1,P9,pfifo_fast,bbr,1,1,2,0,3,4,5,0,0,0,0,0,valid",
	)
	writeRun(t, dir, "run-real", header18,
		"run-real,1,P9,pfifo_fast,bbr,1,1,2,0,30,4,5,0,0,0,0,0,valid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Smallp95Median != 30 {
		t.Fatalf("fixture smoke contamine: %+v", groups)
	}
}

// TestScanNeverMergesDirections — une ligne download et une ligne upload de
// la MÊME cellule (profil|qdisc|cc) forment DEUX groupes. Sans la colonne
// direction, les 150 runs d'historique (up implicite) fusionneraient avec
// les download et corrompraient les médianes publiées.
func TestScanNeverMergesDirections(t *testing.T) {
	dir := t.TempDir()
	header := "run_id,event_id,profile,qdisc,cc,direction,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,voip_r,jfi_pct,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,rtt_base_p50_ms,rtt_base_p95_ms,gate_status"
	writeRun(t, dir, "run-d", header,
		"run-d,1,P2,cake,bbr,up,1,100,114,14,91.4,0,209.2,98.1,18.1,81,0,117288,12.10,0.0,98.3,113.0,valid",
		"run-d,2,P2,cake,bbr,down,1,101,114,13,91.4,0,469.1,35.1,1.8,22,0,31856,3.30,0.0,110.0,126.0,valid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 groupes (up+down), got %d (%+v)", len(groups), groups)
	}
	var up, down *Group
	for i := range groups {
		if groups[i].Direction == "up" {
			up = &groups[i]
		}
		if groups[i].Direction == "down" {
			down = &groups[i]
		}
	}
	if up == nil || down == nil {
		t.Fatalf("groupes direction manquants: %+v", groups)
	}
	if up.Smallp95Median != 209.2 || down.Smallp95Median != 469.1 {
		t.Fatalf("médianes fusionnées: up=%v down=%v", up.Smallp95Median, down.Smallp95Median)
	}
}

// TestScanLegacyRowsDefaultUp — les 150 runs historiques sans colonne
// direction restent des groupes "up" : l'historique ne bouge pas d'un iota.
func TestScanLegacyRowsDefaultUp(t *testing.T) {
	dir := t.TempDir()
	writeRun(t, dir, "run-old", header18,
		"run-old,1,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Direction != "up" {
		t.Fatalf("legacy doit être up: %+v", groups)
	}
	if groups[0].Smallp95Median != 209.2 {
		t.Fatalf("legacy médiane: %v", groups[0].Smallp95Median)
	}
}
