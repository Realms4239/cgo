package api

import (
	"strings"
	"testing"
)

// A guaranteed milestone frame must survive a saturated subscriber: it evicts
// a stale tick rather than being dropped in favour of one.
func TestMilestoneEventSurvivesSaturatedSubscriber(t *testing.T) {
	h := NewHub()
	s := &sub{ch: make(chan string, subCap)}
	h.subs[s] = struct{}{}
	// Fill the channel with tick payloads.
	for i := 0; i < subCap; i++ {
		s.ch <- "id: 1\ndata: {\"tick\":true}\n\n"
	}
	h.PublishEvent("SLA_BREACH_220MS", map[string]any{"rtt_ms": 260})

	found := false
	for len(s.ch) > 0 {
		p := <-s.ch
		if strings.Contains(p, "event: SLA_BREACH_220MS") {
			found = true
		}
	}
	if !found {
		t.Fatal("milestone event was dropped in favour of ticks")
	}
}

// A published milestone is written to the ring and replayed with its event name.
func TestEventFrameIsInTheRing(t *testing.T) {
	h := NewHub()
	h.PublishEvent("AQM_STABILIZED", map[string]any{"qdisc": "cake"})
	frames := h.replayFrom(0)
	if len(frames) != 1 || !strings.Contains(frames[0], "event: AQM_STABILIZED") {
		t.Fatalf("event replay = %v, want one milestone frame", frames)
	}
}

// Hostile oracle: a saturated tick is only dropped (it never evicts a frame),
// and eviction never discards a previously delivered milestone.
func TestDeliverBackpressureAndMilestonePreservation(t *testing.T) {
	h := NewHub()

	// A tick arriving on a tick-only full channel is counted as dropped and
	// leaves the buffered frames untouched. The backpressure notification is
	// best-effort: with no consumer draining, it cannot land on a full channel.
	saturated := &sub{ch: make(chan string, subCap)}
	h.subs[saturated] = struct{}{}
	for i := 0; i < subCap; i++ {
		saturated.ch <- "id: 1\ndata: {\"tick\":true}\n\n"
	}
	h.deliver("id: 99\ndata: {\"tick\":true}\n\n")
	if saturated.dropped != 1 {
		t.Fatalf("saturated tick dropped = %d, want 1", saturated.dropped)
	}
	if len(saturated.ch) != subCap {
		t.Fatalf("tick evicted a frame: len = %d, want %d", len(saturated.ch), subCap)
	}
	delete(h.subs, saturated)

	// A new milestone on a full channel evicts one tick, never the old milestone.
	mixed := &sub{ch: make(chan string, subCap)}
	h.subs[mixed] = struct{}{}
	mixed.ch <- "event: AQM_STABILIZED\ndata: {}\n\n" // oldest frame is a milestone
	for i := 0; i < subCap-1; i++ {
		mixed.ch <- "id: 1\ndata: {\"tick\":true}\n\n"
	}
	h.deliver("event: SLA_BREACH_220MS\ndata: {}\n\n")

	old, fresh := false, false
	for len(mixed.ch) > 0 {
		p := <-mixed.ch
		if strings.Contains(p, "event: AQM_STABILIZED") {
			old = true
		}
		if strings.Contains(p, "event: SLA_BREACH_220MS") {
			fresh = true
		}
	}
	if !old || !fresh {
		t.Fatalf("eviction lost a milestone: old=%v fresh=%v", old, fresh)
	}
}
