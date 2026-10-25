package audit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestAuditSplit — la fenêtre chargée doit se dérouler PENDANT le bulk :
// le bulk démarre à la fin de l'idle, pas à t=12 s quelle que soit la durée.
// Version courte (1 s idle + 1 s chargé) pour ne pas coûter 30 s par test.
func TestAuditSplit(t *testing.T) {
	var loading int32
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := Run(ctx, Params{Duration: 2, Target: "1.1.1.1", BulkAddr: "fake:1"}, Deps{
		Ping: func(context.Context, string, int) []float64 {
			if atomic.LoadInt32(&loading) == 1 {
				return []float64{120, 121, 122}
			}
			return []float64{20, 21, 22}
		},
		Small: func(context.Context) (float64, error) { return 50, nil },
		Bulk: func(ctx context.Context, _ string) (uint64, error) {
			atomic.StoreInt32(&loading, 1)
			<-ctx.Done()
			return 1_000_000, nil
		},
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
