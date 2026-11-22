# Parked Clearance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close five locally-engineerable parks with real fixes and do the full local preparation for the sixth (live kernel trace), which stays human-gated.

**Architecture:** Six sequential slices. Each follows systematic-debugging's Iron Law where a defect is involved: root cause first, failing oracle, minimal fix, full verification, explicit-path commit, fresh reviewer.

**Tech Stack:** Go (`api`, `capabilities`, `cmd/cgo`), React/TypeScript (`web/frontend`), Vitest + happy-dom, Playwright on Edge, `sh .githooks/run-gates.sh`.

## Global Constraints

- No fabricated epoch/frame/detection/measurement/PASS/seal. Absence stays `—` + French word or `NaN` where numeric.
- No new dependency except the explicit local `replace` in Task 2 (approved by this plan). No new hex, font-size, `transition: all`, `@keyframes` or component duplication.
- No host, bench, sadcnwp, SSH key, live kernel trace, corpus edit, frozen data, thesis, tag/freeze. C:\END and protected ledgers read-only.
- No `git add -A`; explicit paths only. No merge. Native gates once per slice; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.
- Three-state exits; a skipped check is never green. Override the pre-commit hook only after a real pre-existing gate block, never precautionarily.

---

### Task 1: SSE epoch marker

**Files:** `api/sse.go`, new `api/sse_epoch_test.go`, `web/frontend/src/lib/sse.ts`, `web/frontend/src/lib/live.ts`, new DOM test.

- [ ] **Step 1:** Add a failing Go wire-level test: fresh hub with epoch B, client reconnects with `Last-Event-ID` landing mid-ring on non-empty replay; assert a full structural frame is emitted (epoch B). Watch RED.
- [ ] **Step 2:** Implement: `Hub.epoch` generated once; include `epoch` in `lastFull` and every published frame; on reconnect, when the client's last frame epoch differs from the current hub epoch, emit `lastFull` before replay. Minimal, no fabricated frame. Watch GREEN.
- [ ] **Step 3:** Client: store `live.epoch`; on frame with a different epoch, clear rings/live state and apply the full snapshot before deltas. Add a DOM oracle for the reset; watch RED→GREEN.
- [ ] **Step 4:** Verify: `go -C api test -count=1 ./...`, targeted DOM, full vitest, build + Go rebuild (rebuild BEFORE restoring tracked `dist/index.html`), unmocked Playwright, gates once. Commit explicit paths.

---

### Task 2: BBRv3 dependency

**Files:** `api/go.mod`, new `api/bbr3_local_test.go`.

- [ ] **Step 1:** Record the proven root cause (`api/go.mod` lacks `replace`, cached v0.1.0 uses `CCPresent("bbr3")`). Write a failing test using `capabilities.StubProvisioned` that expects `BBRv3 == true` through the `api` module's resolved capabilities. Watch RED (undefined symbol).
- [ ] **Step 2:** Add `replace github.com/Realms4239/cgo/capabilities => ../capabilities` to `api/go.mod` (matching the eight sibling replaces). Watch GREEN.
- [ ] **Step 3:** Prove `/api/schema` exposes `bbr3` in the detected state via the existing schema contract test path; no fabricated detection on a stock host.
- [ ] **Step 4:** Verify `go -C api test -count=1 ./...`, `go build ./...`, gates once. Commit explicit paths. Record human gate: real provisioned-kernel verifier re-run.

---

### Task 3: E2E DOM failures

**Files:** determined by reproduction (spec or source).

- [ ] **Step 1:** Rebuild SPA + Go binary in correct order; run `fullstack-realtime.spec.ts` unmocked. Capture exact failures.
- [ ] **Step 2:** If green, record stale-dist false positive with the command evidence. If red, trace why `[data-slot="connection"] .trust-word` is absent despite `App.tsx` mounting `TrustStrip`; fix at source with a genuine oracle. Do not weaken the assertion.
- [ ] **Step 3:** Verify targeted spec + green-bar suite; commit explicit paths or record the false-positive ruling.

---

### Task 4: BenchConfig load flake

**Files:** `web/frontend/src/components/BenchConfig.dom.test.tsx` (or the cause found).

- [ ] **Step 1:** Gather timing evidence: run the suite 5× serially and under load; record durations and the exact failing test.
- [ ] **Step 2:** Root-cause the 5 s exceedance; form one hypothesis; test minimally.
- [ ] **Step 3:** Fix with condition-based determinism (single import, awaited setup) — not a timeout bump. Prove with 5 clean runs.
- [ ] **Step 4:** Full vitest at 2 workers; commit explicit paths.

---

### Task 5: Replay counters and identifiers

**Files:** `web/frontend/src/lib/replay.ts`, `web/frontend/src/lib/replay.test.ts`.

- [ ] **Step 1:** Add `archivedInt(v)` returning `NaN` for absent/blank/malformed input; failing oracle for `drops`, `repetition`, `event_id` → NaN and real `0` stays `0`. Watch RED.
- [ ] **Step 2:** Replace the three `parseInt(...) || 0` sites. Watch GREEN. No type cascade — `NaN` is a valid `number`.
- [ ] **Step 3:** Verify targeted test, full vitest, build + Go rebuild (correct order), unmocked Playwright, gates once. Commit explicit paths.

---

### Task 6: Kernel-truth local preparation

**Files:** new `docs/verification/2026-11-22-ebpf-confrontation-runbook.md`, optional deterministic `cmd/cgo` oracle test.

- [ ] **Step 1:** Write the runbook: exact rebuilt-eBPF-object + real-kernel-trace steps, the sealed 226/242 evidence references, human approvals required, and what must NOT be done (no zero-as-loss, no clang on VM, no host contact without token/window/rollback/rehearsal).
- [ ] **Step 2:** Add a deterministic local oracle for the `--confront-out` JSON shape if not already covered; no live data.
- [ ] **Step 3:** Verify local tests/build; commit explicit paths. The live trace step remains PARKED with the named unblock condition — that is the honest DONE for this task.

---

### Task 7: Programme close

- [ ] Final whole-programme reviewer over the combined diff with the remaining park list; one fix wave + scoped re-review if Critical/Important; append final state; local fast-forward only.

## Self-Review

- Spec coverage: six parks map 1:1 to Tasks 1–6; close in Task 7.
- No placeholders: each step names a command, file or exact record.
- Type consistency: `Hub.epoch`, `LiveFrame.epoch`, `archivedInt` used identically across tasks; no new invented type before proof.
