# CGO / Meteolink Chunk 3A — 20 Hz O(1) SSE Ring Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Raise the SSE hub from 10 Hz with a shifting 2048-slice to 20 Hz with a fixed 4096-slot O(1) circular ring, keeping Last-Event-ID resync, the 64-subscriber cap, delta compression and backpressure named events.

**Architecture:** `api/sse.go` replaces `ring []frame` with `ring [4096]frame` plus `head`/`count` modulo pointers. `Publish` writes at `head`, advances modulo 4096, and replays from the oldest surviving frame. Tick frames are the normal snapshots; milestone event frames are carried in the same ring and are never dropped ahead of ticks on a slow subscriber channel.

**Tech Stack:** Go 1.25 `testing/synctest` for cadence, `sync`, `time`, `net/http/httptest`.

**Source:** `docs/MUSTREAD/13-chunk-3-realtime-pipeline.md` §1–2 (verbatim corpus `CGO5.md` 19–38). Chunk 3 is decomposed; this is sub-plan A.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, frozen corpus/schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, live bench, remote hosts.
- Corpus paths `pkg/api/sse.go` are stale: real file is `api/sse.go`.
- Existing `maxSubs` (64), delta compression, `lastFull` replay for fresh clients and backpressure notification semantics must be preserved, not replaced.
- French comments in Go; English plan/gate text.
- One commit, exact message: `feat(realtime): raise SSE hub to 20hz o(1) 4096-slot ring`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Fixed 4096-slot ring with deterministic wrap

**Files:**
- Modify: `api/sse.go`
- Test: `api/sse_ring_test.go` (create)

**Interfaces:**
- Consumes: existing `frame{id, raw, name}` and `Hub.Publish(snap any)`.
- Produces: `h.ring` as a fixed array with `head`/`count`; replay yields the oldest surviving frame after wrap.

- [ ] **Step 1: Write the failing test**

Create `api/sse_ring_test.go`:

```go
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
```

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestRingWrapKeepsLast4096Deterministically ./...` → FAIL (`undefined: ringCap`, `h.count`, `h.replayFrom`).

- [ ] **Step 3: Implement the ring**

In `api/sse.go`:
- Change `const frameHz = 10` to `const frameHz = 20`.
- Add `const ringCap = 4096`.
- Replace `ring []frame // cap 2048` in `Hub` with `ring [ringCap]frame`, `head int`, `count int`.
- In `NewHub()`, remove any slice initialization of `ring` (a zero-value array is correct).
- In `Publish`, replace the append/shift block with:

```go
	h.lastID++
	f := frame{id: h.lastID, raw: string(raw)}
	h.ring[h.head] = f
	h.head = (h.head + 1) % ringCap
	if h.count < ringCap {
		h.count++
	}
```

- Add a helper:

```go
// replayFrom — frames still present in the ring, oldest first, starting at id >= from.
func (h *Hub) replayFrom(from uint64) []string {
	out := make([]string, 0, h.count)
	start := (h.head - h.count + ringCap) % ringCap
	for i := 0; i < h.count; i++ {
		f := h.ring[(start+i)%ringCap]
		if f.id < from {
			continue
		}
		out = append(out, "id: "+strconv.FormatUint(f.id, 10)+"\ndata: "+f.raw+"\n\n")
	}
	return out
}
```

- Replace the whole replay-index block in `SSE` (from `var from int` through the `replay := ...` loop) with this exact code, preserving the stale-ID `lastFull` resync below it:

```go
	var replay []string
	stale := false
	if lastEventID != "" {
		if id, err := strconv.ParseUint(lastEventID, 10, 64); err == nil {
			if h.count > 0 {
				start := (h.head - h.count + ringCap) % ringCap
				if id < h.ring[start].id {
					stale = true // ID antérieur à l'anneau : resync sur le frame complet
				}
			}
			replay = h.replayFrom(id + 1)
		} else {
			lastEventID = "" // ID illisible ⇒ client frais
		}
	} else {
		// Sans Last-Event-ID : 5 derniers frames seulement (0,25 s à 20 Hz).
		if h.count > 5 {
			start := (h.head - h.count + ringCap) % ringCap
			replay = h.replayFrom(h.ring[(start+h.count-5)%ringCap].id)
		} else {
			replay = h.replayFrom(0)
		}
	}
```

Remove the old `from` variable and the `for _, f := range h.ring[from:]` loop; `replay` is now ready for the existing write loop.

- [ ] **Step 4: Verify** — targeted PASS; full `go -C api test -count=1 ./...` ok.

---

### Task 2: 20 Hz cadence with a virtual clock

**Files:**
- Modify: `api/sse.go` (cadence constant already changed in Task 1)
- Test: `api/sse_cadence_test.go` (create)

**Interfaces:**
- Consumes: `Hub.Serve(next func() any)` and `Hub.Close()`.
- Produces: a 50 ms ticker, verified without wall-clock sleeps.

- [ ] **Step 1: Write the failing test**

Create `api/sse_cadence_test.go`:

```go
package api

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestHubServesAt20Hz(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := NewHub()
		published := make(chan struct{}, 16)
		h.Serve(func() any {
			published <- struct{}{}
			return map[string]any{"tick": true}
		})
		defer h.Close()
		for i := 0; i < 3; i++ {
			select {
			case <-published:
			case <-time.After(time.Second):
				t.Fatal("no tick within 1 virtual second")
			}
		}
		if frameHz != 20 {
			t.Fatalf("frameHz = %d, want 20", frameHz)
		}
	})
}
```

- [ ] **Step 2: Note the oracle** — this test fails if `frameHz` is reverted to 10 (the constant assertion) or if ticks stop. It is not a timing-flake test because `synctest` advances a virtual clock.

- [ ] **Step 3: Implement** — no further product change expected (Task 1 set `frameHz = 20`). If the ticker still uses `time.Second / frameHz`, it now fires at 20 Hz.

- [ ] **Step 4: Verify** — targeted PASS; full api suite ok.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `go -C api test -count=1 ./...` → exit 0.
- [ ] **Step 2: Static proofs** — `git grep -n 'frameHz = 20\|ringCap = 4096' api/sse.go` → 2 hits; `git grep -n 'h.ring = append' api/sse.go` → EMPTY.
- [ ] **Step 3: Commit once** — stage only `api/sse.go`, `api/sse_ring_test.go`, `api/sse_cadence_test.go` by explicit path; message `feat(realtime): raise SSE hub to 20hz o(1) 4096-slot ring`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
