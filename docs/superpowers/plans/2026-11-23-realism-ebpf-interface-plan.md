# Profile Realism and eBPF Primary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add research-grounded profiles with provenance, make eBPF the primary measurement with tc cross-check, and apply honest empty/loading states plus feel-better polish to the live interface.

**Architecture:** Three slices: (1) `model` profiles + provenance + `/api/profiles`, (2) shared fact computation preferring eBPF kernel telemetry with honest absence, (3) live components/CSS using existing `EmptyState`/`StatusPip` and tokens. No new dependency, no frozen corpus edit.

**Tech Stack:** Go (`model`, `api`, `audit`), React/TypeScript (`web/frontend`), Vitest + Playwright Edge.

## Global Constraints

- P1-P4 byte-identical; frozen corpus, book, C:\END, protected ledgers and gates never edited.
- New profiles are emulation targets with citations, never claimed as field measurements.
- eBPF retrans reason is a grouping label, not a verified root cause; copy must say "retransmissions observées".
- tc/qdisc remains a separate named cross-check, never merged into the eBPF number.
- No fabricated zero/green/seal; absence = `—` + French word. No new dependency, hex, font-size, accent, `transition: all`, `@keyframes`.
- 44px targets, WCAG 2.2 AA, forced-colors, reduced-motion; 890-1920 no overflow.
- No `git add -A`; explicit paths only; no merge; native gates once per slice; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Research-grounded profiles

**Files:** Modify `model/types.go`, `model/profile_asym_test.go` (or new `model/profile_real_test.go`), `api/profiles_test.go` if needed.

- [ ] **Step 1:** Write failing Go tests: `Profiles["L1"]` exists with capacity 200/30, delay 26, jitter 15, loss 0.1 and a non-empty `source`; `Profiles["V1"]` with 10/2, 600, 30, 1 and source; `BuiltinProfileIDs` includes both; P1-P4 remain exactly unchanged. Run `go -C model test -count=1 ./...` -> RED.
- [ ] **Step 2:** Add `Source string \`json:"source,omitempty"\`` to `Profile`; add L1/V1 entries with exact citation strings from design §1. Keep P1-P4 lines byte-identical. Run tests -> GREEN.
- [ ] **Step 3:** Check `/api/profiles` contract: run `go -C api test -count=1 ./...`; if an exact-list assertion fails, update it to include the two new IDs without weakening any other assertion. Verify no CSV import test breaks with the new optional field.
- [ ] **Step 4:** Verify `go build ./...`, gates once; commit explicit paths with `feat(model): add cited LEO and VSAT profiles`.

---

### Task 2: eBPF-primary facts

**Files:** Modify `web/frontend/src/lib/dashboard.ts`, `web/frontend/src/lib/dashboard.test.ts`, `web/frontend/src/components/cockpit/DecisionSpine.tsx` (or the single shared fact consumer found during exploration).

- [ ] **Step 1:** Write failing TS tests: with `kernel_telemetry.available=true, source=ebpf`, two flows of retrans_events 1 and 225, the operator loss/queueing word is derived from 226 eBPF retrans events, not userspace RTT; srtt_us median is used for the latency word; tc drops remain a separate labelled number; `source=tcpinfo` or `available=false` renders `—` + absence word with the real reason; no test asserts a verified root cause. Run targeted vitest -> RED.
- [ ] **Step 2:** Implement in the single shared function: prefer eBPF telemetry when valid; fall back honestly; never merge tc into eBPF. Copy uses "retransmissions observées". Watch GREEN.
- [ ] **Step 3:** Full vitest at 2 workers, tsc, build, unmocked Playwright green-bar; gates once. Commit explicit paths with `feat(cockpit): eBPF primary measurement with tc cross-check`.

---

### Task 3: Interface honesty and polish

**Files:** Modify live chapter/field components + `components.css` (exact paths from exploration).

- [ ] **Step 1:** DOM oracles: each live surface that previously showed blank/absent data now renders `EmptyState`/`StatusPip` with the correct kind and word; no digit or green in absence; tabular-nums on live counters; no `transition: all`. Watch RED->GREEN.
- [ ] **Step 2:** Apply the primitives with existing tokens only; concentric radii where nested; 44px controls; no new component duplication.
- [ ] **Step 3:** Playwright 890/1280/1920 screenshots read and validated; axe 2.2 AA; full gates once. Commit explicit paths with `feat(ui): consistent honest empty and loading states`.

## Self-Review

- Spec coverage: §1 -> Task 1; §2 -> Task 2; §3 -> Task 3.
- No placeholders: commands, files and exact values given. No `TBD`, "similar to Task N".
- Type consistency: `Profile.Source`, `kernel_telemetry`, `EmptyKind` used identically across tasks.
