# CGO — plan and specification index

*Last updated: 2026-10-27 · Branch: `rebuild/plan1-economics-audit`*

This is the single index for every CGO rebuild plan. A **spec** is the design
contract for one slice; a **plan** is the implementation contract that follows
it. The source prompt is `epbfBBR.md` (untracked, not a plan). The umbrella
design is `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`.

## Status legend

| Status | Meaning |
|---|---|
| **Done** | Implemented, committed, and the slice's gates passed. |
| **Windows-green** | Engine/tests pass on Windows; live Linux acceptance is still pending. Never call this "verified". |
| **In progress** | Being executed now. |
| **Approved** | Design approved; implementation not started. |
| **Proposed** | Spec and/or plan written; waiting for user approval. |
| **Deferred** | Deliberately not being executed now. |
| **Blocked** | Cannot proceed until the named blocker clears. |
| **Not planned** | Named in `epbfBBR.md`, but no spec or plan exists yet. |
| **Historical** | Pre-rebuild plan; kept for provenance, not to be re-executed. |

## Slice map

| # | Slice | Spec | Plan | Status | Depends on | Next action / blocker |
|---|---|---|---|---|---|---|
| 1 | Enterprise economics & audit | `2026-10-25-cgo-enterprise-rebuild-design.md` | `2026-10-25-cgo-plan1-economics-audit.md` | **Done** (committed; checkboxes not maintained) | — | — |
| 2 | Hardened bench & campaign | same umbrella spec | `2026-10-25-cgo-plan2-bench-campaign.md` | **Done** (committed; checkboxes not maintained) | 1 | — |
| 3 | UI/API surface | same umbrella spec | `2026-10-25-cgo-plan3-ui-api-surface.md` | **Done** (29/29 checked) | 1–2 | — |
| 4 | eBPF ground truth | `2026-10-27-cgo-ebpf-ground-truth-design.md` | `2026-10-27-cgo-ebpf-ground-truth.md` | **Blocked** | stable Linux 6.8 host; prebuilt `bpf_probe.o` | Build `bpf_probe.o` on Linux 6.8, then execute the plan; Windows cannot run Task 1 Step 4 or the live gates |
| 5 | Trace-driven netem `.dist` | `2026-10-27-cgo-trace-driven-dist-design.md` | `2026-10-27-cgo-trace-driven-dist.md` | **Deferred** (user stopped execution) | real field trace; Linux 6.8 bench for acceptance | Resume only on user request; do not commit a `.dist` built from the two local `SMOKE-TEST` traces |
| 6 | Ingress IFB shaping | `2026-10-27-cgo-ifb-ingress-design.md` | `2026-10-27-cgo-ifb-ingress.md` | **Proposed** | Linux 6.8 bench for live acceptance; Windows for engine | User approval, then execute Tasks 1–8; Task 9 stays blocked until the bench is stable |
| 7 | BBRv3 evaluation matrix | `2026-10-27-cgo-bbrv3-matrix-design.md` | `2026-10-27-cgo-bbrv3-matrix.md` | **Deferred** | 7.1 BBRv3 kernel provisioning; live Linux acceptance | Deferred by the user; revisit only when a 7.1 BBRv3 kernel is built and booted. Do not execute now |
| 8 | UI/thesis integration | not written | not written | **Not planned** | measurements from slices 4–7 | UI badge, waste/SLA surfaces, thesis chapter |
| 9 | Engineering system (AIR × DIS synthesis) | `2026-10-27-cgo-engineering-system-design.md` | `2026-10-27-cgo-engineering-system.md` | **In progress** (Phase A done) | — (process layer; runs before any live slice) | Phase B: gate runner and six gates |
| 10 | Live interface observation | `2026-10-27-cgo-live-interface-observation-design.md` | `2026-10-27-cgo-live-interface-observation.md` | **In progress** | 9 Phase A; twin bench | Paint gate: this host's compositor is 45–52 fps with frame skips → 7.1–8.0 paints/s |

## What each live plan covers

### Slices 1–3 — enterprise rebuild (done)

- **Plan 1:** enterprise capacity-depreciation and pooled-APN economics, audit
  core, sealed reports, direction-aware `down`/`both`.
- **Plan 2:** hardened `dualns` bench, campaign matrix, multi-flow JFI,
  `clean`/`bursty` channels, provenance labels.
- **Plan 3:** additive API surface, audit live/stop, dashboard hardening,
  accessibility and copy. The frozen corpus was not touched.

### Slice 4 — eBPF ground truth (blocked)

- Scope: `pkg/probe/bpf` retransmit bytes and socket RTT, `tcp_retransmit_skb`
  raw tracepoint, `tcp_rcv_established` socket sample, `kfree_skb` qdisc-drop
  counter, all-or-nothing per window, `source: ebpf` provenance.
- Scope guard: no BBRv3, `.dist`, IFB, UI, or thesis work in this slice.
- Blocker: no committed `bpf_probe.o`; Windows has no clang/bpftool/BTF; the
  CGO-DEV VM is unstable. Do not call eBPF "verified" before the live test.

### Slice 5 — trace-driven `.dist` (deferred)

- Scope: strict telemetry reader, baseline-only quantile table, sealed native
  netem `.dist` + sidecar, opt-in `CGO_DELAY_MODEL=dist`, fail-closed loading.
- Scope guard: no IFB, BBRv3, eBPF, UI, or thesis work here.
- Blocker: a real field trace. The two `SMOKE-TEST` traces have 19 and 25
  rows, below the 30-sample floor, and are not citable. Live netem acceptance
  needs the Linux 6.8 bench.

### Slice 6 — ingress IFB shaping (proposed)

- Scope: Linux IFB ingress commands in `pkg/remediate`, direction-aware
  `BuildReport`, RouterOS LAN queue tree, opt-in `CGO_SHAPE_MODE=ingress`
  campaign on `dualns`/`down`, IFB preflight, bench `--ingress on`.
- Scope guard: no BBRv3, eBPF, `.dist`, UI, thesis, or `both`-direction
  ingress campaign.
- Blocker: live acceptance needs the Linux 6.8 bench. Windows proves command
  generation and orchestration only.

### Slice 7 — BBRv3 evaluation matrix (deferred)

- Scope: additive `bbr3` label gated on an exact token in
  `/proc/sys/net/ipv4/tcp_available_congestion_control`; P2 ×
  `{pfifo_fast,fq_codel,cake}` × `{cubic,bbr,bbr3}` × 3 reps, `direction=up`,
  `flows=4`; documented command, not a new preset.
- Scope guard: `bbr` keeps meaning BBRv1; no UI, ECN/L4S, IFB, `.dist`, eBPF,
  or thesis chapter. The frozen corpus and Tableau 15 stay untouched.
- Blocker: the 7.1 BBRv3 kernel must be built and booted on the bench VM.
  Windows can only prove the engine; `--cc bbr3` fails closed on a stock host.
  **Deferred by the user**; do not execute now. Revisit only on request, after
  the kernel is available.

### Slice 8 — UI/thesis integration (not planned yet)

- Waits for the measurement slices. The IFB report JSON and Markdown are
  available to the future UI, but no frontend change is in the IFB slice.

### Slice 9 — engineering system (proposed)

- Scope: root `AGENTS.md`, three ledgers (`docs/VERIFY.md`,
  `docs/CHECKPOINTS.md`, `docs/ERRORS.md`), six gates with one fault proof each,
  pre-commit/pre-push hooks, WIP and Friday cadence.
- Scope guard: no product code, no CSV schema change, no frozen run, no eBPF,
  IFB, BBRv3, `.dist`, UI, or thesis work; no CI, no document register, no
  weekly file, no full mutation harness.
- Next action: user approval, then Phase A only. The design deliberately
  contains no live claim; Phase A re-runs the frozen audit and records the
  first dated verification.

### Slice 10 — live interface observation (in progress)

- Scope: a Linux `/proc/net/dev` sampler driven by the existing 10 Hz hub
  clock, additive `iface_*` frame fields, `CGO_IFACE`/`CGO_IFACE_SOURCE`, a
  raw RX/TX interface chart with real timestamps and gap breaks, twin live
  acceptance, and a ready-to-send read-only production request.
- Scope guard: no production writes, control, capture, or root; no SNMP/
  RouterOS adapter yet; no eBPF, IFB, BBRv3, `.dist`, thesis, CSV, or frozen
  corpus change; no new dependency or SSE stream; no replay, interpolation,
  or fake 10 Hz for a slower source.
- Evidence: the pre-slice 10 Hz probe on the twin measured a 100.4–101.8 ms
  window and a 0.2 % byte agreement with the qdisc for the same 25 s upload.
  That is a source probe, not the slice's acceptance.
- Acceptance: the twin SSE, byte-accounting, gap and race gates pass on
  CGO-DEV (325 unique samples in 32 s, window p95 101.3 ms, 0.032 % byte
  delta). The browser paint gate fails as measured: 7.1–8.0 paints/s on a
  46–52 Hz surface, below the 8/s bar; the foreground run had 0 long tasks
  and 10.3 chart-data advances/s. See `docs/VERIFY.md`.
- Plan: `2026-10-27-cgo-live-interface-observation.md` (Tasks 1–10).
- Next action: a stable ≥60 Hz surface for the 25 s paint trace (this host
  skips frames at 45–52 fps); otherwise the human decides between accepting
  the data-advance reading and changing the approved gate. No code change is
  pending. Production numbers wait for operator read-only access and are
  Phase 2.

## Historical plans

| Plan | Status | Why it is here |
|---|---|---|
| `2026-09-03-refonte-leaderboard.md` | **Historical** | Pre-rebuild UI plan; superseded by the 2026-10-25 rebuild. Checkboxes were never maintained. |
| `2026-09-05-remontee-16-20.md` | **Historical** | Pre-rebuild plan; superseded. Do not re-execute. |

## Dependency graph

```
Plans 1–3 (done)
      │
      ├──► engineering system (slice 9) ──► Phase A ──► Phase B ──► C ──► D
      │
      ├──► IFB ingress (slice 6) ──► live acceptance on Linux 6.8
      │
      ├──► trace-driven .dist (slice 5) ──► real field trace ──► live netem acceptance
      │
      └──► eBPF ground truth (slice 4) ──► stable Linux 6.8 + bpf_probe.o

BBRv3 (slice 7) ◄── provisioning + live acceptance        [deferred]
UI/thesis (slice 8) ◄── measurements from 4–7
Live interface observation (slice 10) ◄── slice 9 Phase A + twin bench
```

Slices 4–7 are independent of each other, but they share the same Linux 6.8
bench. Do not run their acceptance tasks concurrently on one host: each
mutates namespaces and qdiscs.

Slice 9 is independent of 4–8 and should run first: it installs the frozen
guard and the ledgers that the live slices record into. It does not itself
provide a Linux bench.

Slice 10 depends on slice 9 Phase A for the ledgers and the frozen guard, and
on the twin bench for acceptance. It is independent of slices 4–7; it shares
the twin with slice 6, so do not run both acceptance tasks at the same time.

## Standing constraints (every slice)

- Frozen corpus guard: the counts and their status live in `docs/VERIFY.md`.
  Cite that row; never restate a count from memory. Say "valid lines", never
  "runs".
- `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, and `data/runs/` are
  frozen. New provenance is additive only.
- Preserve `basis: measured|estimated`; an eBPF measurement adds
  `source: ebpf`, it does not rewrite `basis`.
- No runtime `clang`, `gcc`, `llvm`, or `linux-headers` on the VM. The eBPF
  object is prebuilt and committed.
- French comments and user-facing text; English identifiers.
- Commit per logical group with explicit paths. Never `git add -A`.
- Leave `epbfBBR.md` and user artifacts untracked.
- Do not start Graylog. The remote PC `100.114.101.13` is off-limits.
- Never fabricate a measurement. Do not call a path "verified" until its live
  Linux acceptance task has actually run.
- Do not merge to `main` unless the user decides.

## Windows gates

Each plan ends with the same host gates, unless the plan states otherwise:

```bash
go test ./...
go vet ./...          # needs web/frontend/dist; build the frontend first if absent
GOOS=linux GOARCH=amd64 go build ./...
python book/tools/audit-archive.py   # expect its two OK lines; counts in docs/VERIFY.md
git status --short
```

If `book/` or `web/frontend/dist` is absent in the checkout, record the gate
as "not runnable here"; do not claim it passed.

## Plan hygiene

1. A new slice gets a spec and a plan **before** code.
2. Update this index in the same commit as the spec/plan.
3. Tick plan steps as they finish; record a blocker instead of silently
   dropping a step.
4. A plan that changes a shared invariant updates the umbrella spec's
   decision log, not only its own.
5. A new governing document is added to `AGENTS.md` in the same commit.
6. Every number in a plan is labelled: measured, recorded, pending or blocked.
