package results

import (
	"os"
	"path/filepath"
	"testing"
)

// En-tête réel des runs gelés (18 colonnes : qdi_ms inséré en 8, pas de
// voip_r) — chaque colonne porte une valeur distinctive pour que toute
// lecture positionnelle (schéma pré-qdi) échoue bruyamment.
const header18 = "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"

func writeRun(t *testing.T, dir, run, header string, rows ...string) {
	t.Helper()
	runDir := filepath.Join(dir, run)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	out := header + "\n"
	for _, r := range rows {
		out += r + "\n"
	}
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(out), 0644); err != nil {
		t.Fatal(err)
	}
}

func findGroup(groups []Group, profile, qdisc, cc string) *Group {
	for i := range groups {
		if groups[i].Profile == profile && groups[i].Qdisc == qdisc && groups[i].CC == cc {
			return &groups[i]
		}
	}
	return nil
}

// TestScanReadsByName — non-régression du décalage qdi_ms : Scan doit agréger
// par NOM de colonne, pas par position. Avec l'ancien code positionnel :
// Smallp95Median recevait qdi (57.4), GoodputMedian recevait deadline_ok_pct
// (60.2), WastedMedian recevait retransmissions, et Quarantined restait 0
// (cpu_pct testé comme gate_status).
func TestScanReadsByName(t *testing.T) {
	dir := t.TempDir()
	writeRun(t, dir, "run-t", header18,
		"run-t,1,P2,pfifo_fast,bbr,1,111.0,168.4,57.4,328.2,60.2,18.2,443,5,641464,66.40,0.0,valid",
		"run-t,2,P2,cake,bbr,1,100.0,114.0,14.0,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
		"run-t,3,P2,pfifo_fast,bbr,2,112.0,170.0,58.0,330.0,59.0,18.0,450,6,650000,67.00,0.0,invalid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d (%+v)", len(groups), groups)
	}
	pfifo := findGroup(groups, "P2", "pfifo_fast", "bbr")
	if pfifo == nil {
		t.Fatal("missing P2|pfifo_fast|bbr")
	}
	if pfifo.Count != 2 {
		t.Fatalf("pfifo count = %d, want 2", pfifo.Count)
	}
	if pfifo.Quarantined != 1 {
		t.Fatalf("pfifo quarantined = %d, want 1", pfifo.Quarantined)
	}
	// médianes par nom, lignes invalid exclues (comme extract-stats.js) :
	// small=328.2 (ligne valid seule), pas qdi 57.4
	if pfifo.Smallp95Median != 328.2 {
		t.Fatalf("pfifo Smallp95Median = %v, want 328.2 (qdi leak = 57.4)", pfifo.Smallp95Median)
	}
	if pfifo.GoodputMedian != 18.2 {
		t.Fatalf("pfifo GoodputMedian = %v, want 18.2 (deadline leak = 60.2)", pfifo.GoodputMedian)
	}
	if pfifo.DeadlineMedian != 60.2 {
		t.Fatalf("pfifo DeadlineMedian = %v, want 60.2 (small leak = 328.2)", pfifo.DeadlineMedian)
	}
	if pfifo.WastedMedian != 641464 {
		t.Fatalf("pfifo WastedMedian = %v, want 641464 (retrans leak = 5)", pfifo.WastedMedian)
	}
	cake := findGroup(groups, "P2", "cake", "bbr")
	if cake == nil {
		t.Fatal("missing P2|cake|bbr")
	}
	if !cake.Best || pfifo.Best {
		t.Fatalf("best must be cake (small 209.2 < 329.1): cake=%v pfifo=%v", cake.Best, pfifo.Best)
	}
	if cake.Smallp95Median != 209.2 || cake.DeadlineMedian != 98.1 || cake.GoodputMedian != 18.1 {
		t.Fatalf("cake medians wrong: %+v", cake)
	}
}

// TestScanMixedLayouts — l'historique mélange 17 colonnes (sans qdi_ms),
// 18 (sans voip_r) et 19 (actuel) : l'agrégation par nom doit les fusionner
// sans décalage.
func TestScanMixedLayouts(t *testing.T) {
	dir := t.TempDir()
	header17 := "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"
	writeRun(t, dir, "run-old", header17,
		"run-old,1,P2,cake,cubic,1,90,100,200.0,95.0,15.0,10,0,50000,5.0,0,valid",
	)
	writeRun(t, dir, "run-mid", header18,
		"run-mid,1,P2,cake,cubic,1,90,100,12.0,210.0,96.0,15.5,11,0,51000,5.2,0,valid",
	)
	header19 := "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,voip_r,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"
	writeRun(t, dir, "run-new", header19,
		"run-new,1,P2,cake,cubic,1,90,100,12.0,88.5,220.0,97.0,16.0,12,0,52000,5.4,0,valid",
	)
	groups, err := Scan(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	g := findGroup(groups, "P2", "cake", "cubic")
	if g == nil {
		t.Fatalf("missing P2|cake|cubic in %+v", groups)
	}
	if g.Count != 3 {
		t.Fatalf("count = %d, want 3 (3 layouts merged)", g.Count)
	}
	// small = med(200,210,220) = 210 — pas de décalage d'une colonne
	if g.Smallp95Median != 210 {
		t.Fatalf("Smallp95Median = %v, want 210", g.Smallp95Median)
	}
	if g.GoodputMedian != 15.5 {
		t.Fatalf("GoodputMedian = %v, want med(15,15.5,16)=15.5", g.GoodputMedian)
	}
	if g.Quarantined != 0 {
		t.Fatalf("Quarantined = %d, want 0", g.Quarantined)
	}
}
