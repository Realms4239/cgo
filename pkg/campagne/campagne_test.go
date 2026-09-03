package campagne

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

func fastDeps() Deps {
	return Deps{
		TC: &fakeTC{}, CliIf: "veth-c", ShaperIf: "veth-s",
		Target: "10.0.0.1", SmallURL: "http://127.0.0.1/obj", BulkAddr: "127.0.0.1:5201",
		Ping:        func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 24} },
		Small:       func(context.Context) (float64, error) { return 25, nil },
		Bulk:        func(ctx context.Context, _ string) (uint64, error) { return 2_500_000, nil }, // ≈ 20 Mbit/s sur la fenêtre ≥ 1 s
		CPU:         func() float64 { return 30 },
		BaselineSec: -1, ChargeSec: -1, RecupSec: -1, // fenêtres instantanées dans les tests
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
	d.Bulk = func(ctx context.Context, _ string) (uint64, error) { return 0, nil } // G1 échoue → invalid
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

// §5 contrat live : OnSnap porte les mesures en cours PENDANT la fenêtre
// de charge — la publication aux seules frontières affamait le mur de zéros
// pendant toute la charge (courbes à 0 alors que SSE disait running).
func TestRunEventLivePublish(t *testing.T) {
	d := fastDeps()
	d.BaselineSec = 1
	d.ChargeSec = 2 // tours de boucle de sonde à ~300 ms
	d.RecupSec = -1 // instantané — defaults() remplacerait 0 par la vraie fenêtre recup

	release := make(chan struct{})
	var once sync.Once
	releaseAll := func() { once.Do(func() { close(release) }) }
	defer releaseAll()
	d.Bulk = func(ctx context.Context, _ string) (uint64, error) {
		<-release // retenir le bulk en plein charge — la boucle de collecte doit publier en live malgré tout
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
// RunEvent réinitialise le saut avant d'appliquer netem — les cellules
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

// Stop doit être prompt : une campagne annulée ne doit pas continuer à sonder
// jusqu'à l'échéance de phase (la boucle de collecte ignorait ctx jusqu'à 120 s).
func TestStopPromptOnCancel(t *testing.T) {
	d := fastDeps()
	d.BaselineSec = 30 // phase longue — seule l'annulation peut terminer la collecte tôt
	var calls int32
	d.Ping = func(ctx context.Context, _ string, n int) []float64 {
		atomic.AddInt32(&calls, 1)
		time.Sleep(20 * time.Millisecond)
		return []float64{20, 21}
	}
	d.Small = func(context.Context) (float64, error) { return 25, nil }

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = RunEvent(ctx, model.Event{RunID: "r1", EventID: 4, Profile: "P2", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}, model.Profiles["P2"], d)
		close(done)
	}()
	time.Sleep(300 * time.Millisecond) // laisser quelques tours tourner
	cancel()
	select {
	case <-done:
		// retourné promptement — bon
	case <-time.After(3 * time.Second):
		t.Fatal("RunEvent ignored cancellation for >3 s mid-phase — stop is not prompt")
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		// RunEvent peut remonter l'annulation ou finir dégradé — les deux sont acceptables
		_ = runErr
	}
	if atomic.LoadInt32(&calls) > 40 { // 300 ms à ~20 ms/tour ≈ ≤ 20 ; généreux ×2
		t.Fatalf("probes kept running after cancel: %d rounds", calls)
	}
}

// Surveillance continue — ping + petits objets, sans bulk (non intrusif):
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

// TestRunEventCounterResetClamped — qdisc remplacé en cours de cellule : les
// compteurs tc repartent de zéro. Le delta final négatif ne doit ni faire
// wraper Drops (uint64 → ~1.8e19) ni gonfler le goodput (G4 invalid à tort) —
// même garde int64 que publishLive.
func TestRunEventCounterResetClamped(t *testing.T) {
	d := fastDeps()
	calls := 0
	d.StatsFn = func() []qdisc.Stats {
		calls++
		// Séquence d'appels SANS OnSnap (fastDeps) : 1 = socle initial,
		// 2 = re-socle de charge (publishLive inopérant sans OnSnap),
		// 3 = comptage final. Le qdisc est remplacé entre 2 et 3.
		if calls <= 2 {
			return []qdisc.Stats{{Kind: "cake", Bytes: 10_000_000, Drops: 500}}
		}
		return []qdisc.Stats{{Kind: "cake", Bytes: 100, Drops: 3}}
	}
	ev := model.Event{RunID: "r1", EventID: 9, Profile: "P2", Qdisc: model.Cake, CC: model.Cubic, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d)
	if err != nil {
		t.Fatal(err)
	}
	if got.Drops != 0 {
		t.Fatalf("Drops = %d after counter reset, want 0 (no wrap)", got.Drops)
	}
	if got.GateStatus != model.GatePass {
		t.Fatalf("status=%s want valid (sender-side goodput 20 Mbit/s, rx delta ignored)", got.GateStatus)
	}
	if got.BulkGoodputMbps < 19 || got.BulkGoodputMbps > 21 {
		t.Fatalf("goodput = %v, want ~20 (sender side, rx reset ignored)", got.BulkGoodputMbps)
	}
}
