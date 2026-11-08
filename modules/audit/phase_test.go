package audit

import "testing"

// The 60 s contract is 10 s quiescent / 40 s saturated / 10 s drain. The old
// engine spent 42 s idle and only 8 s under load, which is not a stress audit
// and made a 60 s run look like the thesis's short probe.
func TestPhasePlanSixtySeconds(t *testing.T) {
	idle, loaded, rec := phasePlan(60, true)
	if idle != 10 || loaded != 40 || rec != 10 {
		t.Fatalf("60 s plan = %d/%d/%d, want 10/40/10", idle, loaded, rec)
	}
}

func TestPhasePlanShortKeepsLoad(t *testing.T) {
	idle, loaded, rec := phasePlan(2, true)
	if idle < 1 || loaded < 1 || rec != 0 {
		t.Fatalf("2 s plan = %d/%d/%d, want a short loaded window", idle, loaded, rec)
	}
	idle, loaded, rec = phasePlan(30, true)
	if idle != 7 || loaded != 16 || rec != 7 {
		t.Fatalf("30 s plan = %d/%d/%d, want 7/16/7", idle, loaded, rec)
	}
}

func TestPhasePlanNoLoad(t *testing.T) {
	idle, loaded, rec := phasePlan(10, false)
	if idle != 10 || loaded != 0 || rec != 0 {
		t.Fatalf("no-load plan = %d/%d/%d, want 10/0/0", idle, loaded, rec)
	}
}

func TestLossPercentCountsRecoveryProbes(t *testing.T) {
	if got := lossPercent(20, 20); got != 0 {
		t.Fatalf("loss = %v, want 0", got)
	}
	// 20 attempted, 18 received: 10 %, not a fabricated 25 %.
	if got := lossPercent(20, 18); got != 10 {
		t.Fatalf("loss = %v, want 10", got)
	}
	if got := lossPercent(0, 0); got != 0 {
		t.Fatalf("loss = %v, want 0", got)
	}
}
