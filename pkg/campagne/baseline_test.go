package campagne

import (
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

func prof(id string) model.Profile { return model.Profiles[id] }

// A wired link declaring no jitter must behave exactly as the old rule did:
// floor = max(5, 20% of median), no extra allowance.
func TestG6WiredUnchanged(t *testing.T) {
	p := model.Profile{ID: "P1", CapacityMbps: 80, DelayMs: 20, JitterMs: 0}
	thr := baselineThreshold(20, p)
	if thr != 5 {
		t.Fatalf("wired threshold = %v, want 5 (max(5, 0.2*20))", thr)
	}
	thr = baselineThreshold(100, p)
	if thr != 20 {
		t.Fatalf("wired threshold = %v, want 20 (0.2*100)", thr)
	}
}

// The P4 case that motivated this: a Starlink baseline spread of 17-19.5 ms
// must now pass, because the profile declares 20 ms of jitter.
func TestG6StarlinkBaselineNowPasses(t *testing.T) {
	p := prof("P4")
	if p.JitterMs != 20 {
		t.Fatalf("P4 jitter changed: %v", p.JitterMs)
	}
	// Reproduce the observed frozen-corpus shape: median ~41 ms, p95 ~58 ms.
	base := make([]float64, 0, 20)
	for i := 0; i < 20; i++ {
		base = append(base, 41.0)
	}
	// Make p95 land near 58 by adding high samples.
	for i := 0; i < 3; i++ {
		base = append(base, 58.0)
	}
	ok, thr := BaselineStable(base, p)
	if !ok {
		t.Fatalf("Starlink baseline still fails G6 (threshold %.2f ms)", thr)
	}
	// The threshold must be floor + 2*20 where floor = max(5, 0.2*41) = 8.2,
	// so 48.2 ms. The observed spread of 17-19.5 ms sits well inside it.
	if thr < 47 || thr > 49 {
		t.Fatalf("P4 threshold = %.2f, want ~48.2", thr)
	}
}

// A genuinely unstable host must still fail, even on a jittery link: the
// allowance is bounded, not unlimited.
func TestG6StillFailsAThrashingHost(t *testing.T) {
	p := prof("P4")
	base := make([]float64, 0, 20)
	for i := 0; i < 20; i++ {
		base = append(base, 41.0)
	}
	// A host under contention: p95 far above anything the link explains.
	for i := 0; i < 3; i++ {
		base = append(base, 400.0)
	}
	ok, thr := BaselineStable(base, p)
	if ok {
		t.Fatalf("a 400 ms baseline spike passed G6 (threshold %.2f ms)", thr)
	}
}

// P2 cellular: 15 ms declared jitter, 100 ms delay. A spread within the link's
// jitter must pass; the old rule allowed only max(5, 20) = 20 ms.
func TestG6CellularAllowsDeclaredJitter(t *testing.T) {
	p := prof("P2")
	thr := baselineThreshold(100, p)
	// floor 20 + 2*15 = 50
	if thr != 50 {
		t.Fatalf("P2 threshold = %v, want 50", thr)
	}
}

// P3 VSAT: 30 ms jitter at 600 ms delay.
func TestG6VSATAllowsDeclaredJitter(t *testing.T) {
	p := prof("P3")
	thr := baselineThreshold(600, p)
	// floor max(5, 120) = 120 + 2*30 = 180
	if thr != 180 {
		t.Fatalf("P3 threshold = %v, want 180", thr)
	}
}

// Too few samples cannot pass, regardless of how flat they look.
func TestG6NeedsEnoughSamples(t *testing.T) {
	p := prof("P1")
	ok, _ := BaselineStable([]float64{20, 20, 20}, p)
	if ok {
		t.Fatal("G6 passed on 3 samples; a p95 needs more")
	}
}

// The threshold must never be negative or zero for a real profile.
func TestG6ThresholdAlwaysPositive(t *testing.T) {
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		if thr := baselineThreshold(0, prof(id)); thr <= 0 {
			t.Fatalf("%s threshold = %v", id, thr)
		}
	}
}
