# Two-plane truth and physics implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development.
> One task at a time, fresh implementer then fresh reviewer, fix loops, sequential.
> Checkbox (`- [ ]`) syntax for tracking.

**Goal:** make the two truth planes explicit in provenance, and add the one
missing netem physics lever that is safe to ship locally (jitter correlation),
without touching the frozen corpus, gates, ledgers or archived seals.

**Architecture:** `model.Profile` carries the new optional field;
`qdisc.applyNetemArgs` is the single tc boundary that emits it;
`pkg/campagne.applyPhysics` passes it to both directions; `api` CSV import gains
one optional trailing column. Calibration/confrontation gain additive provenance.
No new dependency, no schema version bump.

**Tech Stack:** Go modules `model`, `qdisc`, `api`, `pkg/campagne`, `cmd/cgo`.

## Global constraints

- Spec: `docs/superpowers/specs/2026-11-23-two-plane-truth-physics-design.md`.
- P1-P4, frozen corpus, book, C:\END, protected ledgers, gates, permits and
  config are never edited. No `git add -A`; stage explicit paths; no merge.
- eBPF stays a first-class temporal observation plane. "tc primary" means
  declared accounting source, not removal or demotion of eBPF.
- Correlation 0 => byte-identical tc argument vectors (frozen-corpus oracle).
- Absence is explicit (`—` + French word), never a fabricated zero or green.
- No live bench, no host contact, no new dependency in these two tasks.
- Native gates once per task; pre-existing `links` FAIL and `book/` SKIPs are
  recorded, never repaired by editing ledgers/gates.

---

### Task 1 — Jitter correlation end to end

**Files:** `model/types.go`, `qdisc/ge.go`, `qdisc/ge_test.go`,
`pkg/campagne/dualns.go`, `api/server.go` (CSV import only), plus a focused
orchestration test in `pkg/campagne/dualns_test.go`.

- [ ] **Step 1:** Write failing Go oracles first. In `qdisc/ge_test.go`:
  - `c < 0` and `c > 1` each return an error and produce **zero** `FakeRunner`
    calls (fail closed before tc);
  - `c = 0` yields the exact existing vector
    `qdisc replace dev <i> root handle 1: netem delay 100ms 15ms`;
  - `c = 0.25` yields `... delay 100ms 15ms 0.25`;
  - with `distribution`, the token order is `delay 100ms 15ms 0.25 distribution P2`.
  In `model`: a profile with out-of-range `JitterCorrelation` is rejected by a
  shared `ValidateJitterCorrelation` before import; P1-P4 remain unchanged.
  Run `go -C qdisc test -count=1 ./...` and `go -C model test -count=1 ./...` -> RED.
- [ ] **Step 2:** Add `JitterCorrelation float64 \`json:"jitter_correlation,omitempty"\``
  to `model.Profile`, plus `ValidateJitterCorrelation(c float64) error` (0..1,
  reject NaN). Extend `applyNetemArgs` with a correlation parameter emitted
  between jitter and distribution only when > 0. Extend `ApplyNetemPhysics`,
  `ApplyNetemGE`, `ApplyNetemDistribution` additively (new variadic option or new
  wrapper functions — keep existing call sites byte-identical). Run tests -> GREEN.
- [ ] **Step 3:** Wire `prof.JitterCorrelation` through
  `pkg/campagne.applyPhysics` / `applyPhysicsDist` (both directions, unchanged
  reset-before-mutation ordering) and through CSV import as optional column 10.
  Old rows (<=10 columns) keep the historical zero-correlation behaviour. Update
  only the assertions that enumerate the CSV shape; never weaken a hostile check.
- [ ] **Step 4:** Run `go -C qdisc test -count=1 ./...`,
  `go -C model test -count=1 ./...`, `go -C api test -count=1 ./...`,
  `go -C pkg/campagne test -count=1 ./...`, `go build ./...`. Run gates once,
  record exact exit codes and the pre-existing links FAIL / book SKIPs.
  Commit explicit paths: `feat(physics): netem jitter correlation with fail-closed validation`.

### Task 2 — Truth-plane provenance (executable, additive)

**Files:** `qdisc/dist.go`, `qdisc/dist_test.go`, `qdisc/stats.go`,
`cmd/cgo/confront.go`, `cmd/cgo/confront_test.go`.

- [ ] **Step 1:** Write failing oracles. `BuildCalibration` with an unrecognised
  plane fails before writing a sidecar; `LoadCalibration` of an archived sidecar
  without `physics_plane` returns a value whose plane reads
  `rtt_baseline (legacy)` and is never promoted to `ebpf_observation`;
  a confrontation report names tc as accounting surface and eBPF as temporal
  surface in an additive `truth_planes` object, and contains no hierarchy claim.
  Run targeted tests -> RED.
- [ ] **Step 2:** Add `PhysicsPlane string \`json:"physics_plane,omitempty"\``
  to `qdisc.Calibration`; a closed validator accepts only `ebpf_observation`,
  `rtt_baseline`, `literature`; legacy absence maps to the explicit legacy label.
  Add `truth_planes` to the confrontation struct with fixed role strings. No
  numeric field, no schema version, no archived seal is rewritten. Run tests -> GREEN.
- [ ] **Step 3:** Full Go suites for the four modules and `cmd/cgo`; gates once;
  commit explicit paths: `feat(truth): explicit tc accounting and ebpf temporal planes`.

### Named park — serialization rate / packet overhead (NOT in this plan's executable tasks)

`netem rate RATE [PACKETOVERHEAD [CELLSIZE [CELLOVERHEAD]]]` gives deterministic
per-byte serialization (`sch_netem.c:packet_time_ns`). Enabling it changes which
hop is the bottleneck and can suppress the AQM queueing under study. Unblock
condition: one human-approved live A/B cell (serialization on/off, same profile,
tc accounting and eBPF temporal both captured), with measured before/after and
no gate relaxation. Record as BLOCKED with that condition; do not implement a
silent default.

## Self-review

- Spec §1 -> Task 2; §2 -> Task 1; §3 -> park; §5 -> acceptance in both tasks.
- Every task names exact files, RED->GREEN oracles and commands. No TBD.
- Type consistency: `JitterCorrelation` and `PhysicsPlane` used identically in
  model, qdisc, campagne and confrontation.
- eBPF is preserved as the near-wire-speed observation plane throughout.
