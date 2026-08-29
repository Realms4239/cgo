package campagne

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/model"
)

func fastDeps() Deps {
	return Deps{
		TC: &fakeTC{}, CliIf: "veth-c", ShaperIf: "veth-s",
		Target: "10.0.0.1", SmallURL: "http://127.0.0.1/obj", BulkAddr: "127.0.0.1:5201",
		Ping:        func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 24} },
		Small:       func(context.Context) (float64, error) { return 25, nil },
		Bulk:        func(ctx context.Context, _ string) (uint64, error) { return 2_500_000, nil }, // ≈20 Mbit/s over the ≥1 s window
		CPU:         func() float64 { return 30 },
		BaselineSec: -1, ChargeSec: -1, RecupSec: -1, // instant windows in tests
	}
}

type fakeTC struct{}

func (fakeTC) Run(args ...string) ([]byte, error) { return nil, nil }

func TestRunEventHappyPath(t *testing.T) {
	ev := model.Event{RunID: "r1", EventID: 1, Profile: "P2", Qdisc: model.FqCodel, CC: model.Cubic, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], fastDeps())
	if err != nil {
		t.Fatal(err)
	}
	if got.GateStatus != model.GatePass {
		t.Fatalf("status=%s want valid (gates=%v)", got.GateStatus, got.RTTp95Ms)
	}
	if got.BulkGoodputMbps <= 0 || got.RTTp95Ms == 0 {
		t.Fatalf("metrics not filled: %+v", got)
	}
}

func TestRunEventQuarantinesWhenBulkFails(t *testing.T) {
	d := fastDeps()
	d.Bulk = func(ctx context.Context, _ string) (uint64, error) { return 0, nil } // G1 fails → invalid
	ev := model.Event{RunID: "r1", EventID: 2, Profile: "P2", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d)
	if err != nil {
		t.Fatal(err)
	}
	if got.GateStatus != model.GateInvalid {
		t.Fatalf("status=%s want invalid", got.GateStatus)
	}
}

func TestWriterAppendAndFreeze(t *testing.T) {
	dir := t.TempDir()
	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	ev := model.Event{RunID: "run9", EventID: 3, Profile: "P1", Qdisc: model.Cake, CC: model.Cubic, Repetition: 2,
		RTTp50Ms: 20, RTTp95Ms: 24, Smallp95Ms: 26, DeadlineOKPct: 99, BulkGoodputMbps: 79.5,
		WastedBytes: 1024, CostARPerH: 0, CPUPct: 31, GateStatus: model.GatePass}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(ev); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("G5 not enforced: %v", err)
	}
	if err := w.Freeze("cfgsha"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		t.Fatal("manifest missing")
	}
}

// §5 live contract: OnSnap carries running measurements DURING the charge
// window — boundary-only publishing starved the Wall with zeros for the
// whole charge (charts flatlined at 0 while SSE said running).
func TestRunEventLivePublish(t *testing.T) {
	d := fastDeps()
	d.BaselineSec = 1
	d.ChargeSec = 2 // probe loop rounds at ~300ms
	d.RecupSec = -1 // instant — defaults() would replace 0 with the real recup window

	release := make(chan struct{})
	var once sync.Once
	releaseAll := func() { once.Do(func() { close(release) }) }
	defer releaseAll()
	d.Bulk = func(ctx context.Context, _ string) (uint64, error) {
		<-release // hold bulk mid-charge — collect loop must publish live regardless
		return 2_500_000, nil
	}

	var mu sync.Mutex
	liveNonZero := 0
	d.OnSnap = func(s Snapshot) {
		if s.Phase == model.PhaseCharge && s.RTTp50Ms > 0 && s.Smallp95Ms > 0 {
			mu.Lock()
			liveNonZero++
			mu.Unlock()
		}
	}

	ev := model.Event{RunID: "r1", EventID: 3, Profile: "P2", Qdisc: model.FqCodel, CC: model.Cubic, Repetition: 1}
	done := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = RunEvent(context.Background(), ev, model.Profiles["P2"], d)
		close(done)
	}()

	deadline := time.Now().Add(10 * time.Second)
	for {
		mu.Lock()
		n := liveNonZero
		mu.Unlock()
		if n > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no live snapshot with running measurements mid-charge")
		}
		select {
		case <-done:
			t.Fatal("RunEvent finished without any live mid-charge publish")
		case <-time.After(50 * time.Millisecond):
		}
	}
	releaseAll()
	<-done
	if runErr != nil {
		t.Fatal(runErr)
	}
}

// The shape lever (and any stale state) can leave a foreign qdisc at root —
// RunEvent resets the hop before applying netem so cells self-heal instead of
// failing wholesale.
func TestRunEventResetsStaleQdisc(t *testing.T) {
	ops := [][]string{}
	ftc := &recTC{sink: &ops}
	d := fastDeps()
	d.TC = ftc
	ev := model.Event{RunID: "r1", EventID: 9, Profile: "P2", Qdisc: model.FqCodel, CC: model.Cubic, Repetition: 1}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err != nil {
		t.Fatal(err)
	}
	foundDel := false
	netemIdx, delIdx := -1, -1
	for i, op := range ops {
		if len(op) >= 5 && op[0] == "qdisc" && op[1] == "del" && op[4] == "root" {
			foundDel = true
			if delIdx < 0 {
				delIdx = i
			}
		}
		if len(op) >= 6 && op[0] == "qdisc" && op[1] == "replace" && op[4] == "root" && op[5] == "netem" && netemIdx < 0 {
			netemIdx = i
		}
	}
	if !foundDel {
		t.Fatal("no qdisc del before apply — stale handles would break the cell")
	}
	if netemIdx >= 0 && delIdx > netemIdx {
		t.Fatalf("del must precede netem replace: del@%d netem@%d", delIdx, netemIdx)
	}
}

type recTC struct{ sink *[][]string }

func (r *recTC) Run(args ...string) ([]byte, error) {
	cp := make([]string, len(args))
	copy(cp, args)
	*r.sink = append(*r.sink, cp)
	return nil, nil
}

// Surveillance continue — ping+small sans bulk (ARG.md: non-intrusif) : le
// mur reste vivant hors campagne, le levier de façonnage devient visible.
func TestStartWatchPublishes(t *testing.T) {
	d := fastDeps()
	var mu sync.Mutex
	n := 0
	d.OnSnap = func(s Snapshot) {
		if s.Phase == "surveil" && s.RTTp50Ms > 0 {
			mu.Lock()
			if s.Running {
				t.Error("watch snapshot must not claim Running — it flaps the panel")
			}
			n++
			mu.Unlock()
		}
	}
	d.BaselineSec = -1
	stop := StartWatch(context.Background(), d)
	deadline := time.Now().Add(5 * time.Second)
	for {
		mu.Lock()
		c := n
		mu.Unlock()
		if c >= 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no surveillance snapshots with measurements")
		}
		time.Sleep(50 * time.Millisecond)
	}
	stop()
	mu.Lock()
	after := n
	mu.Unlock()
	time.Sleep(600 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if n > after {
		t.Fatalf("stop() did not stop the loop: %d → %d", after, n)
	}
}
