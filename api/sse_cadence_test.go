package api

import (
	"testing"
	"testing/synctest"
	"time"
)

// TestHubServesAt20Hz proves the 50 ms SSE cadence under a virtual clock:
// no wall-clock sleeps. The oracle is threefold — frameHz must still be 20,
// the ticker must actually deliver ticks (a stopped ticker fails closed), and
// consecutive ticks must be exactly 50 ms apart on the virtual clock (a wrong
// ticker interval fails closed even if the constant is left untouched).
func TestHubServesAt20Hz(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := NewHub()
		ticks := make(chan time.Time, 16)
		h.Serve(func() any {
			ticks <- time.Now()
			return map[string]any{"tick": true}
		})
		defer h.Close()

		var stamp [3]time.Time
		for i := range stamp {
			select {
			case stamp[i] = <-ticks:
			case <-time.After(time.Second):
				t.Fatal("no tick within 1 virtual second")
			}
		}
		if frameHz != 20 {
			t.Fatalf("frameHz = %d, want 20", frameHz)
		}
		for i := 1; i < len(stamp); i++ {
			if d := stamp[i].Sub(stamp[i-1]); d != 50*time.Millisecond {
				t.Fatalf("tick %d interval = %v, want 50ms (frameHz=%d)", i, d, frameHz)
			}
		}
	})
}
