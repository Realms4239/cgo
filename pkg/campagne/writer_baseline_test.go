package campagne

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// TestWriterGelsBaselineRTT — l'écriture gelée porte rtt_base_p50/p95_ms
// (colonnes après cpu_pct, avant gate_status) : la note Waveform d'une
// cellule exige la latence au repos, aujourd'hui perdue au gel.
func TestWriterGelsBaselineRTT(t *testing.T) {
	dir := t.TempDir()
	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	ev := model.Event{RunID: "r9", EventID: 1, Profile: "P2", Qdisc: model.Cake, CC: model.BBR, Repetition: 1,
		RTTp50Ms: 100, RTTp95Ms: 114, RTTBaseP50Ms: 21, RTTBaseP95Ms: 23, GateStatus: model.GatePass}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	if err := w.Freeze("cfg"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "aqm_eval.csv"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "rtt_base_p50_ms") {
		t.Fatalf("header missing baseline columns: %s", s)
	}
	// ligne : ...,100,114 (rtt),...,21.0,23.0 (base),...,valid — mêmes décimales que les autres colonnes
	line := strings.Split(s, "\n")[1]
	cols := strings.Split(line, ",")
	if len(cols) < 2 || cols[len(cols)-2] != "23.0" || cols[len(cols)-3] != "21.0" {
		t.Fatalf("baseline RTT not gelled last-before-gate: %v (cols=%d)", cols, len(cols))
	}
}
