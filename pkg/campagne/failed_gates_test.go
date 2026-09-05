package campagne

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

func TestQuarantineLogsFailedGates(t *testing.T) {
	dir := t.TempDir()
	w, err := OpenRun(filepath.Join(dir, "run-x"))
	if err != nil {
		t.Fatal(err)
	}
	ev := model.Event{RunID: "run-x", EventID: 1, Profile: "P2", Qdisc: "cake", CC: "bbr", GateStatus: model.GateInvalid}
	ev.FailedGates = []string{"G4ThroughputCoherent"}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	if err := w.Freeze("cfg"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "run-x", "quarantine.json"))
	var q []map[string]any
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatal(err)
	}
	if q[0]["failed_gates"] == nil {
		t.Fatalf("failed_gates absent: %s", string(b))
	}
}
