# Kernel Truth Seal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `cgo kerneltruth seal` that emits `cgo.kernel-truth-seal/1` from a real sealed audit and two real `tc -s` captures, fail-closed on any missing or ambiguous value.

**Architecture:** New `cmd/cgo/seal.go` with a pure `buildSeal(auditJSON, tcBefore, tcAfter, iface, qdisc)` function, a new top-level command in `cmd/cgo/main.go` reusing `confront.go`'s `kernelTruthSeal` struct and atomic-write pattern. No new dependency.

**Tech Stack:** Go, `encoding/json`, `regexp` for the `dropped N` capture, `testing`.

## Global Constraints

- No fabricated counter/seal/PASS; absence-never-zero; no default input path, no embedded fixture.
- Source must be `ebpf`, telemetry available, flows present; `tc_drops_after >= tc_drops_before`; exactly one unambiguous `dropped` line per capture.
- No corpus, ledger, gate, frozen data, permit or dependency change. No `git add -A`; explicit paths only. No merge.
- `qdisc_drop_packets=0` must carry the named reason-filter blind-spot note, never mean zero loss.
- Native gates once; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Seal producer

**Files:** Create `cmd/cgo/seal.go`, `cmd/cgo/seal_test.go`; Modify `cmd/cgo/main.go`.

**Interfaces:** consumes `audit.Report` JSON and `tc -s` text; produces `kernelTruthSeal` (`cmd/cgo/confront.go:14-25`) exactly as `loadSeal` expects.

- [ ] **Step 1: Write failing hostile tests** in `cmd/cgo/seal_test.go` against `buildSeal`: positive fixture with `source=ebpf`, telemetry available, two flows 1+225, object sha, payload bytes, `qdisc_drop_packets=0`, tc captures `dropped 1` / `dropped 227` → seal with all exact values, `match`-relevant fields copied, note about the 0 blind spot. Hostile: missing audit path, source `tcpinfo`, telemetry unavailable, no flows, duplicate `dropped` lines, `after < before`, a `null` counter, and `schema` wrong. Run → FAIL (undefined).
- [ ] **Step 2: Implement `buildSeal`** in `cmd/cgo/seal.go`: parse audit JSON into the needed shape (reuse `audit.Report` if importable, else a local struct with the exact JSON tags), sum flow `retrans_events`, parse `dropped N` with a strict regex requiring exactly one match per file, fail closed with a named reason on every hostile case, write atomically via the `confront.go` pattern. Run tests → GREEN.
- [ ] **Step 3: Register `kerneltruth seal`** in `cmd/cgo/main.go` with `--audit`, `--tc-before`, `--tc-after`, `--iface`, `--qdisc`, `--out`; exit 2 on any error. Add a CLI-level test that a missing audit exits 2 with no output file.
- [ ] **Step 4: Verify** `go -C cmd/cgo test -count=1 ./...`, `go build ./...`, `sh .githooks/run-gates.sh` once; record exit codes. Commit explicit paths with `feat(cgo): seal kernel truth from real audit and tc captures`.

---

### Task 2: Live continuation (no further operator approval — authorized)

**Files:** new evidence under `C:/WEBSERVER/evidence/`; new token/commands under `C:/WEBSERVER`.

- [ ] **Step 1:** Cross-compile the linux binary, deploy as a NEW guest file under a fresh token + classify-only rehearsal (persisted 2026-11-22 ruling; no permit widening).
- [ ] **Step 2:** Run a fresh eBPF audit on the bench with the real prebuilt object through `CGO_ALLOW_PRIVATE_TARGETS=1`, capturing real `tc -s` before/after into guest files with sha256 sidecars.
- [ ] **Step 3:** Run `kerneltruth seal` on those real files, then `cgo run --confront-out` on the produced seal; capture report + sha256.
- [ ] **Step 4:** If any guard refuses, any value is ambiguous, or `qdisc_drop_packets=0`, STOP, apply the token's rollback, report BLOCKED with the exact evidence sha256. Never hand-write a seal, never promote a VERIFY row from this pass.

## Self-Review

- Spec coverage: §3 mapping → Task 1 Step 2; §4 command → Step 3; §5 verification → Step 4; §6 live continuation → Task 2.
- No placeholders: every step names a command, file or exact condition.
- Type consistency: `kernelTruthSeal` reused from `confront.go`; `schema`/`source` constants unchanged.
