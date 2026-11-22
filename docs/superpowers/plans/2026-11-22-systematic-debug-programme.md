# Systematic Debug Programme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply systematic-debugging's four phases to the four named surfaces — backend detection, kernel truth, real-time SSE, measurement chain — finding and fixing root causes with independent oracles.

**Architecture:** Four sequential SDD slices. Each slice runs Phase 1 (root-cause evidence), Phase 2 (pattern analysis), Phase 3 (single hypothesis, minimal test), Phase 4 (one fix with a failing test) before any code change. A slice that finds no defensible root cause records the investigation and parks with a named condition rather than inventing a fix.

**Tech Stack:** Go (`api`, `capabilities`, `probe/bpf`, `cmd/cgo`), React/TypeScript (`web/frontend`), `synctest`, Vitest, unmocked Playwright on Edge.

## Global Constraints

- Iron Law: NO FIXES WITHOUT ROOT-CAUSE INVESTIGATION FIRST. Three failed fix attempts → stop and question architecture.
- No fabricated evidence, measurement, seal or PASS. Absence stays `—` + French word. No gate weakening, no skip, no timeout raised to turn red green.
- No live bench, sadcnwp, frozen corpus, thesis, tag/freeze, host or hardware contact. All work local and read-only against real local artifacts.
- Protected `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, frozen corpus never edited. No `git add -A`; explicit paths only. No merge.
- Native gates once per slice; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Backend detection skew

**Files:** `api/go.mod`, `api/server.go`, `api/*_test.go` (paths determined by root-cause evidence)

**Interfaces:** consumes `capabilities.CapabilitiesNow()`; produces either a corrected real BBRv3 detection or an honest parked record.

- [ ] **Step 1 (Phase 1):** Trace which capabilities module the `api` module actually compiles against. Run `go -C api list -m all | grep capabilities`, `go -C api list -f '{{.Dir}}' github.com/Realms4239/cgo/capabilities`, and read `api/go.mod`. Record the exact version, the detected semantics (`CCPresent("bbr3")` vs `BBRv3Token()`), and the observable consequence.
- [ ] **Step 2 (Phase 2):** Compare with the local `capabilities` package and its tests (`capabilities/bbr3.go`, `bbr3_test.go`). List every behavioural difference, however small.
- [ ] **Step 3 (Phase 3):** Write one hypothesis in the report (e.g. "the cached v0.1.0 module under-detects v3 on a provisioned host because..."). Test it with the smallest possible Go test in `api/` that captures the real detection path. Do NOT change dependencies yet.
- [ ] **Step 4 (Phase 4):** If the hypothesis is confirmed and the fix is a local code change with no new dependency, write a failing test, apply the minimal fix, watch GREEN. If the fix requires a dependency change or a real kernel host, STOP and park with the named unblock condition — do not fabricate a detection.
- [ ] **Step 5:** Run `go -C api test -count=1 ./...`, `go build ./...`, `sh .githooks/run-gates.sh --only tests`; record exit codes. Commit explicit paths with `fix(api): ...` or record the parked ruling in the SDD ledger only.

---

### Task 2: Kernel-truth blind spot

**Files:** `cmd/cgo/confront.go`, `cmd/cgo/confront_test.go`, `probe/bpf/BUILD.md` (read-only), `docs/VERIFY.md` row 43 (read-only)

**Interfaces:** consumes the sealed 226-event confrontation evidence and the `--confront-out` JSON contract; produces an honest root-cause record and, only if proven locally fixable, a source fix.

- [ ] **Step 1 (Phase 1):** Gather evidence: read VERIFY row 43, `probe/bpf/BUILD.md` § « Limitation mesurée », `docs/verification/2026-11-15-bench-vm/README.md:80`, and `cmd/cgo/confront.go`. Record the three named limitations and the 226/242 event relationship verbatim.
- [ ] **Step 2 (Phase 2):** Compare the reason-filter semantics (`SKB_DROP_REASON_QDISC_DROP`, `kfree_skb`) with the `qdisc_drop_packets=0` observation. Identify whether zero is a measurement or a filter artifact.
- [ ] **Step 3 (Phase 3):** State one hypothesis in the report. Test it with the smallest local oracle (a unit test on the JSON/field semantics, or a code-path trace) — no bench, no host.
- [ ] **Step 4 (Phase 4):** If a local source defect is proven (e.g. a field silently coerced to zero), write a failing test and fix it. If the truth requires a live kernel trace, park with the named unblock condition; do NOT reinterpret `qdisc_drop_packets=0` as loss.
- [ ] **Step 5:** Run targeted Go tests and build; record exit codes; commit explicit paths or park in the ledger.

---

### Task 3: Real-time SSE under load

**Files:** `api/sse_hub.go`, `api/sse_cadence_test.go`, `web/frontend/src/lib/live.ts` (paths by evidence)

**Interfaces:** consumes the 20 Hz hub, `synctest` cadence oracle, and the real idle-frame contract; produces a fix only for a proven defect.

- [ ] **Step 1 (Phase 1):** Reproduce with evidence: run `go -C api test -count=5 -run 'TestHub|TestNextSnap|TestSSE' ./...` and the unmocked realtime E2E spec; capture any non-green run with exact test name and stderr. If all green, record that the symptom is not reproducible locally and state what evidence is missing.
- [ ] **Step 2 (Phase 2):** Compare the virtual-clock cadence proof with the wall-clock ticker and the E2E cadence envelope. List differences that could matter under load.
- [ ] **Step 3 (Phase 3):** Form one hypothesis (e.g. "idle-frame field X is presented as a measurement under load because..."). Test minimally with a DOM or Go oracle.
- [ ] **Step 4 (Phase 4):** Fix only a proven defect with a failing test. Do not relax a cadence assertion or widen a tolerance to turn red green.
- [ ] **Step 5:** Run targeted tests, full suite, build, unmocked Playwright; record exit codes; commit explicit paths or park.

---

### Task 4: Measurement chain

**Files:** `web/frontend/src/lib/ringBuffer.ts`, `web/frontend/src/lib/trace.ts`, `web/frontend/src/components/labo/ProvenanceStrip.tsx` (paths by evidence)

**Interfaces:** consumes the `NaN = absence` rule, integrity status field, and real calibration sidecars; produces a fix only for a proven defect.

- [ ] **Step 1 (Phase 1):** Trace end-to-end: from a real SSE frame into `push`/`pushIface`, through `ringTrace`/`buildTrace`, into the rendered absence words and integrity status. Run the existing DOM/unit oracles and record every boundary where a value could become a fabricated zero or a green without a real source.
- [ ] **Step 2 (Phase 2):** Compare with the ratified honesty contract (corpus `CGO4`, ratified spatial spec §1.2) and list every divergence.
- [ ] **Step 3 (Phase 3):** State one hypothesis about a concrete divergence and test it with a failing oracle.
- [ ] **Step 4 (Phase 4):** Fix at the source boundary with one minimal change; watch RED→GREEN. If no divergence is found, record the clean result and park with the evidence.
- [ ] **Step 5:** Run targeted DOM tests, full vitest, build, unmocked Playwright, gates once; record exit codes; commit explicit paths or park.

---

### Task 5: Programme close

- [ ] Run the final whole-programme reviewer over the combined diff with the parked list; one fix wave + scoped re-review if Critical/Important findings appear; append final state to the SDD ledger; no merge beyond the local fast-forward.

## Self-Review

- Spec coverage: the four named surfaces map one-to-one to Tasks 1–4; close in Task 5.
- No placeholders: each step names a command, a file or an exact record to gather. No `TBD`, no "similar to Task N".
- Type consistency: `Capabilities`, `Calibration`, `LiveFrame` field names are used as defined in the existing code; no new type is invented before a root cause is proven.
