# API timing flake diagnosis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the two flaky `api` timing tests deterministic so the native `tests` gate stops failing under load, without weakening their oracles.

**Architecture:** Both tests currently sleep a fixed 250 ms and assert a 10 Hz ticker produced ≥2 calls. Under suite load that wall-clock budget is not guaranteed. Replace the sleeps with condition-based waiting (poll the observed call count until the assertion holds or a generous deadline), preserving the exact behavioural assertions. Prefer the stdlib in-process virtual clock only if it can drive the real hub without a test-only seam.

**Tech Stack:** Go 10 Hz SSE hub, `testing`, condition-based polling.

## Global Constraints

- Root cause first (systematic-debugging): reproduce, trace, form one hypothesis, test minimally.
- No weakening: the assertions `NextSnap calls >= 2` and "ticks stop after Close" must remain meaningful and fail if the behaviour regresses.
- No `t.Parallel()` added, no blanket sleeps increased, no skips, no gate edits.
- Native gates still show the pre-existing `links` false positive and `book/` SKIPs — record, never edit.
- No `git add -A`; stage explicit paths only. No merge. Protected ledgers/frozen corpus untouched.
- French comments where comments are needed; English identifiers and gate messages.

---

### Task 1: Replace wall-clock sleeps in the two flaky tests

**Files:**
- Modify: `api/next_snap_test.go:43` (sleep before `NextSnap calls >= 2` assertion)
- Modify: `api/sse_hub_test.go:95` and `:102` (sleeps around tick/close assertions)
- Test: the two test files themselves

**Interfaces:**
- Consumes: real hub behaviour — `NextSnap` call counter and post-`Close` tick counter, both already instrumented in the tests.
- Produces: the same assertions, driven by condition-based waiting with a bounded deadline.

- [ ] **Step 1: Reproduce and confirm the hypothesis**

Run: `cd api && go test -count=1 -run 'TestNextSnapDrivesHubAndNotState|TestHubCloseStopsTicks' .` three times; then run the same two tests under load with the full `go test ./...` once. Record exact failures. Confirm the failure is "fewer ticks than the fixed sleep assumes", not a logic error.

- [ ] **Step 2: Read the working pattern**

Read `docs/..` skill reference `condition-based-waiting.md` from the systematic-debugging skill directory and the two test files completely. List every difference between the fixed-sleep approach and condition-based waiting.

- [ ] **Step 3: Write the failing oracle first**

Replace only the fixed sleeps with a small local helper that polls the observed counter at a short interval until the existing expected condition holds, with a deadline of 2 seconds (well above 10 Hz × 2 ticks). The helper must call `t.Fatalf` with the counter value and elapsed time on timeout. Run the two tests in isolation: expect PASS (the helper observes the real condition). To prove the oracle is genuine, temporarily invert the expected condition and confirm it FAILS, then restore it.

- [ ] **Step 4: Apply the helper to all three sleeps**

Update `api/next_snap_test.go:43`, `api/sse_hub_test.go:95` and `:102`. For the post-Close assertion, wait for the ticker to reach its expected count before closing, then assert no further ticks after Close with a short bounded observation window. No assertion is removed or loosened.

- [ ] **Step 5: Verify under load**

Run `cd api && go test -count=1 ./...` five times serially; all five must pass. Then run `sh .githooks/run-gates.sh --only tests` once. Record exit codes. If it still fails, STOP and return to Phase 1 with the new evidence — do not add a sixth run or a larger timeout blindly.

- [ ] **Step 6: Commit**

`git add api/next_snap_test.go api/sse_hub_test.go && git commit -m "test(api): make hub timing tests deterministic under load"`. If a pre-existing gate blocks the hook, use the documented no-hooks override and record why.

## Self-Review

- Spec coverage: root cause → Step 1, pattern → Step 2, failing oracle → Step 3, fix → Step 4, load verification → Step 5, commit → Step 6.
- No placeholders, no `t.Parallel`, no skips, no removed assertions.
- Type consistency: one helper name used in both files; counters are the existing instrumented values.
