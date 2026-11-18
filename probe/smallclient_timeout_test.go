package probe

import (
	"testing"
	"time"
)

func TestSmallClientTimeoutUsesCalibratedFloor(t *testing.T) {
	c := SmallClientTimeout(4 * 600 * time.Millisecond)
	if c.Timeout != 2400*time.Millisecond {
		t.Fatalf("P2 VSAT client timeout = %v, want 2.4s", c.Timeout)
	}
	if got := smallClientFloor(10 * time.Millisecond); got != 8*time.Second {
		t.Fatalf("floor(10ms) = %v, want 8s", got)
	}
}
