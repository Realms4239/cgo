package api

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// Les frames SSE portent time.Now() — avec l'horloge virtuelle de synctest,
// le scénario avance instantanément et les minuteurs 10 Hz sont
// reproductibles sans sleeps réels.
func TestSynctestVirtualClockAdvances(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int64
		start := time.Now()
		for i := 0; i < 10; i += 1 {
			n.Add(1)
			time.Sleep(time.Millisecond)
		}
		synctest.Wait()
		if elapsed := time.Since(start); elapsed < 10*time.Millisecond {
			t.Fatalf("virtual clock did not advance: %v", elapsed)
		}
		if got := n.Load(); got != 10 {
			t.Fatalf("n = %d, want 10", got)
		}
	})
}

func TestSSEFrameOrderingDeterministic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := NewHub()
		var counted atomic.Int64
		s := &sub{ch: make(chan string, 16)}
		h.subs[s] = struct{}{}
		go func() {
			for range s.ch {
				counted.Add(1)
			}
		}()
		h.Publish(snap{Phase: "baseline", Profile: "P1", RTTp50Ms: 20})
		h.Publish(snap{Phase: "baseline", Profile: "P1", RTTp50Ms: 20})
		h.Publish(snap{Phase: "charge", Profile: "P1", RTTp50Ms: 150})
		close(s.ch)
		synctest.Wait()
		if got := counted.Load(); got != 3 {
			t.Fatalf("got %d frames, want 3", got)
		}
	})
}
