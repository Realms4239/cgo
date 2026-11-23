# Kernel-truth honest-absence seal plan (2026-11-23)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Steps use `- [ ]` checkboxes.

**Goal:** Let `cgo kerneltruth seal` emit a real seal from the existing sealed eBPF audit when `qdisc_drop_packets` is absent, representing that absence explicitly (documented 6.8-cake reason-filter blind spot), then produce the confrontation report with `match` from retrans vs tc delta.

**Root cause (Phase 1, 3 guest windows):** no `bpf_probe.o` / no clang on guest; all 5 sealed audits have `qdisc_drop_packets` ABSENT; real 226-drop truth exists only as `tc -s` before/after. `seal.go` `requireJSONFields(telRaw, ..., "qdisc_drop_packets", ...)` and `confront.go` `loadSeal` required-key loop both reject absence. Same command cannot succeed — hence explicit-absence representation, not a third repeat.

## Global Constraints
- Never coerce absent→0; never fabricate a drop counter. `qdisc_drop_packets` absent must be an explicit boolean + note, never a measured 0.
- `match` computed ONLY from retrans_events vs tc drops delta (unchanged).
- No frozen corpus/book/C:\END/ledger/gate edit, no new dependency, no `git add -A`, explicit paths only, no merge, no host contact in code tasks.
- Native gates once; pre-existing `links` FAIL + `book/` SKIPs recorded, never edited.

## Task 1: producer accepts explicit absence
**Files:** `cmd/cgo/seal.go`, `cmd/cgo/seal_test.go`, `cmd/cgo/confront.go`, `cmd/cgo/confront_test.go`.

- [ ] **Step 1:** Failing Go tests: (a) `buildSeal` with a real audit fixture whose `kernel_telemetry` omits `qdisc_drop_packets` succeeds and emits `qdisc_drop_packets_absent:true` + note, numeric field omitted/not 0; (b) an audit with `qdisc_drop_packets:null` still FAILS CLOSED (null is a disguised zero, never absence); (c) `loadSeal` accepts a seal with the explicit absence flag and no numeric field, rejects one with neither; (d) `match` still true when retrans==tc delta. Run `go -C cmd/cgo test -count=1 ./...` → RED.
- [ ] **Step 2:** Minimal change: in `seal.go`, drop `qdisc_drop_packets` from `requireJSONFields`; if absent, set `QdiscDropPacketsAbsent=true` and a named note (reuse `sealZeroDropNote` style), leaving the numeric zero only when the field is genuinely present and zero (which its own validation already notes). In `confront.go`, add `QdiscDropPacketsAbsent bool \`json:"qdisc_drop_packets_absent,omitempty"\`` to `kernelTruthSeal`; in `loadSeal`, require either a non-null numeric `qdisc_drop_packets` OR `qdisc_drop_packets_absent:true` (never neither, never null). Copy the absence flag+note verbatim into the confrontation report's kernel truth block. Run tests → GREEN.
- [ ] **Step 3:** `go build ./...`, gates once; commit explicit paths `feat(cgo): seal kernel truth with explicit qdisc-drop absence`.

## Task 2: final review + 4th live window
- [ ] Reviewer over the diff; one fix wave + scoped re-review if Critical/Important.
- [ ] Build linux binary, deploy as new guest file under a fresh token + rehearsal (persisted 2026-11-22 ruling), run `kerneltruth seal` from the REAL `audit-1794860919` report + REAL `tc -s` captures, then `cgo run --confront-out`. Preserve sha256 sidecars. Any refusal/ambiguity → rollback + BLOCKED. No VERIFY promotion.

## Self-Review
- Spec coverage: root cause → Task 1; live unblock → Task 2. No placeholders. `kernelTruthSeal` reused; absence never numeric 0.
