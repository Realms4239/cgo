# CGO / Meteolink Chunk 3B — Guaranteed Milestone Events Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a milestone-event frame path to the SSE hub that is never dropped ahead of ticks, and raise subscriber channel capacity from 16 to 32.

**Architecture:** `api/sse.go` gains `PublishEvent(kind string, detail map[string]any)`, which stores the event in the same 4096 ring and delivers it to every subscriber. A subscriber's channel has `cap 32`; when it is full, the sender evicts stale tick payloads (never a milestone event) to make room. Backpressure notifications remain the fallback only when no evictable tick exists.

**Tech Stack:** Go 1.25 `encoding/json`, `strings`, `sync`; module `api/`.

**Source:** `docs/MUSTREAD/13-chunk-3-realtime-pipeline.md` §2 (verbatim corpus `CGO5.md` 23–27). Chunk 3 is decomposed; this is sub-plan B.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, frozen corpus/schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, live bench, remote hosts.
- Real file is `api/sse.go` (corpus `pkg/api/sse.go` is stale).
- Preserve everything from 3A: 20 Hz ticker, 4096 ring, `Last-Event-ID` resync, 64-subscriber cap, delta compression, fresh-client `lastFull` sync, existing backpressure semantics.
- Milestone kinds are exactly: `CALIBRATION_BASELINE`, `BULK_SATURATION_START`, `SLA_BREACH_220MS`, `QDISC_DROP_BURST`, `AQM_STABILIZED`, `DUAL_SEAL_COMMITTED`.
- This plan delivers the **mechanism**; wiring real audit/campaign state to emit these kinds is 3C scope and must be recorded as deferred, not claimed here.
- French comments in Go; English plan/gate text.
- One commit, exact message: `feat(realtime): deliver milestone SSE events ahead of ticks`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Event publication with eviction-based priority

**Files:**
- Modify: `api/sse.go`
- Test: `api/sse_events_test.go` (create)

**Interfaces:**
- Consumes: `Hub.Publish`, `frame{id, raw, name}`, `sub{ch, dropped}` from 3A.
- Produces: `func (h *Hub) PublishEvent(kind string, detail map[string]any)`; `const subCap = 32`; `func isMilestonePayload(s string) bool`.

- [ ] **Step 1: Write the failing test**

Create `api/sse_events_test.go`:

```go
package api

import (
	"strings"
	"testing"
)

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

func TestEventFrameIsInTheRing(t *testing.T) {
	h := NewHub()
	h.PublishEvent("AQM_STABILIZED", map[string]any{"qdisc": "cake"})
	frames := h.replayFrom(0)
	if len(frames) != 1 || !strings.Contains(frames[0], "event: AQM_STABILIZED") {
		t.Fatalf("event replay = %v, want one milestone frame", frames)
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run 'TestMilestoneEventSurvivesSaturatedSubscriber|TestEventFrameIsInTheRing' ./...` → FAIL (`undefined: subCap`, `PublishEvent`).

- [ ] **Step 3: Implement**

In `api/sse.go`:
- Add `const subCap = 32` and `var milestoneKinds = map[string]bool{...}` with the six exact names.
- Add:

```go
func isMilestonePayload(payload string) bool {
	for k := range milestoneKinds {
		if strings.Contains(payload, "event: "+k) {
			return true
		}
	}
	return false
}
```

- Add `PublishEvent`, which locks `h.mu`, increments `lastID`, writes a `frame{id, raw, name: kind}` into the ring, builds the payload `event: <kind>\ndata: <json>\n\n`, and calls a shared `deliver(payload)` helper.

- Factor the delivery loop out of `Publish` into `deliver(payload string)` (still under `h.mu`):

```go
func (h *Hub) deliver(payload string) {
	priority := isMilestonePayload(payload)
	for s := range h.subs {
		select {
		case s.ch <- payload:
			continue
		default:
		}
		if !priority {
			s.dropped++
			select {
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
			continue
		}
		// Milestone: evict one stale tick, never another milestone.
		for i := 0; i < cap(s.ch); i++ {
			old, ok := <-s.ch
			if !ok {
				break
			}
			if isMilestonePayload(old) {
				s.ch <- old
				continue
			}
			s.dropped++
			break
		}
		select {
		case s.ch <- payload:
		default: // no evictable tick: last resort, report backpressure
			s.dropped++
			select {
			case s.ch <- fmt.Sprintf("event: backpressure\ndata: {\"type\":\"backpressure\",\"dropped\":%d}\n\n", s.dropped):
			default:
			}
		}
	}
}
```

- Change `s := &sub{ch: make(chan string, 16)}` at `:208` to `make(chan string, subCap)`.
- Keep `Publish`'s tick payload delivery calling `h.deliver(payload)` so behavior is unchanged except capacity.

- [ ] **Step 4: Verify** — targeted tests PASS; full `go -C api test -count=1 ./...` ok.

---

### Task 2: Prove the priority is live with a mutant

**Files:** none new (test-only verification).

**Interfaces:** consumes Task 1.

- [ ] **Step 1: Mutant proof** — Temporarily change `deliver` to treat milestones as normal ticks (`priority := false`), run `go -C api test -count=1 -run TestMilestoneEventSurvivesSaturatedSubscriber ./...` → must FAIL, then restore byte-for-byte and confirm SHA-256 unchanged and the test PASSes again. Record both outputs.
- [ ] **Step 2: Static proof** — `git grep -n 'subCap' api/sse.go` ≥2 hits; `git grep -n 'PublishEvent' api/` ≥1; `git grep -n 'make(chan string, 16)' api/sse.go` → EMPTY.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suite** — `go -C api test -count=1 ./...` → exit 0.
- [ ] **Step 2: Commit once** — stage only `api/sse.go` and `api/sse_events_test.go` by explicit path; message `feat(realtime): deliver milestone SSE events ahead of ticks`. If a gate fails, STOP and report.
- [ ] **Step 3: Record** — verbatim outputs, mutant proof and limitations to the SDD report; explicitly record that real audit/campaign emission of the six kinds is 3C scope, not delivered here. No `docs/CHECKPOINTS.md` append.
