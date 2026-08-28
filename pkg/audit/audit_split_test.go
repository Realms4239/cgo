package audit

import (
	"context"
	"testing"
	"time"
)

// 3-window split: 0-12s idle, 12-22s bulk, 22-30s loaded — loaded RTT must differ from idle
// when the link is actually loaded (fake pings simulate load appearing with the bulk window).
// Real 30s audit by design (B split Q26) — no fake clock seam, honest duration.
func TestAuditSplit(t *testing.T) {
	if testing.Short() {
		t.Skip("30s real audit — skip in -short")
	}
	start := time.Now()
	p := Params{Duration: 30, Target: "1.1.1.1"}
	res, err := Run(context.Background(), p, Deps{
		Ping: func(_ context.Context, _ string, _ int) []float64 {
			// load kicks in at the 22s mark (bulk flood 12-22s congests the link)
			if time.Since(start) > 22*time.Second {
				return []float64{120, 121}
			}
			return []float64{20, 21}
		},
		Small: func(_ context.Context) (float64, error) { return 50, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RTTIdleP50 == res.RTTLoadedP50 {
		t.Fatalf("idle==loaded not split: %v == %v", res.RTTIdleP50, res.RTTLoadedP50)
	}
	if res.RTTLoadedP50 <= res.RTTIdleP50 {
		t.Fatalf("loaded should exceed idle under bulk: idle %v loaded %v", res.RTTIdleP50, res.RTTLoadedP50)
	}
}