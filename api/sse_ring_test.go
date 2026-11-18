package api

import (
	"strings"
	"testing"
)

func TestRingWrapKeepsLast4096Deterministically(t *testing.T) {
	h := NewHub()
	for i := 1; i <= 4098; i++ {
		h.Publish(map[string]any{"n": i})
	}
	if h.count != ringCap {
		t.Fatalf("count = %d, want %d", h.count, ringCap)
	}
	frames := h.replayFrom(0)
	if len(frames) != ringCap {
		t.Fatalf("replay length = %d, want %d", len(frames), ringCap)
	}
	// After 4098 publishes the surviving IDs are 3..4098.
	if !strings.Contains(frames[0], "id: 3\n") {
		t.Fatalf("oldest replay frame = %q, want id 3", frames[0])
	}
	if !strings.Contains(frames[len(frames)-1], "id: 4098\n") {
		t.Fatalf("newest replay frame = %q, want id 4098", frames[len(frames)-1])
	}
}
