# CGO — plan and specification index

*Last updated: 2026-10-29 · Branch: `rebuild/plan1-economics-audit`*

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
| 5 | Trace-driven netem `.dist` | `2026-10-27-cgo-trace-driven-dist-design.md` | `2026-10-27-cgo-trace-driven-dist.md` | **Windows-green** (engine Tasks 1–7) | real field trace; Linux 6.8 bench for acceptance | Produce a field trace of at least 30 baseline samples, run `cgo calibrate`, then Task 8 on Linux; never cite a `SMOKE-TEST` `.dist` |
| 6 | Ingress IFB shaping | `2026-10-27-cgo-ifb-ingress-design.md` | `2026-10-27-cgo-ifb-ingress.md` | **Windows-green** (verifier pass 2026-10-28) | Linux 6.8 bench for live acceptance; Windows for engine | Engine row `VERIFIED`; Task 9 not executed, live acceptance blocked on a stable Linux 6.8 bench. |
| 7 | BBRv3 evaluation matrix | `2026-10-27-cgo-bbrv3-matrix-design.md` | `2026-10-27-cgo-bbrv3-matrix.md` | **Windows-green** (engine Tasks 1–7) | 7.1 BBRv3 kernel provisioning; live Linux acceptance | Build and boot the 7.1 BBRv3 kernel on the bench VM, then run Task 8; `bbr` still means BBRv1 |
| 8 | Audit report surface & thesis citation | `2026-10-28-cgo-audit-surface-citation-design.md` | `2026-10-28-cgo-audit-surface-citation.md` | **In progress** (builder pass; verifier pending) | 1, 3, 10 (done); not 4–7 | Independent verifier re-runs `npm test`, `npm run build`, the Go tests and `python tools/verify.py`; the thesis paste stays the human's |
| 9 | Engineering system (AIR × DIS synthesis) | `2026-10-27-cgo-engineering-system-design.md` | `2026-10-27-cgo-engineering-system.md` | **Done** (verifier pass 2026-10-28) | — (process layer; runs before any live slice) | — (merge to `main` is the human's decision) |
| 10 | Live interface observation | `2026-10-27-cgo-live-interface-observation-design.md` | `2026-10-27-cgo-live-interface-observation.md` | **Done** | 9 Phase A; twin bench | — (Phase 2 pending operator access) |
| 11 | Frozen corpus seal anchor | `2026-10-28-cgo-corpus-seal-design.md` | `2026-10-28-cgo-corpus-seal.md` | **Done** (verifier pass 2026-10-28) | — | — (follow-up: FAIL instead of SKIP on a partial `book/`; recorded) |
| 12 | Book truth-up & prose restructure | `2026-10-28-cgo-book-truth-up-design.md` | `2026-10-28-cgo-book-truth-up.md` | **Done** (book working tree, never committed) | 1–3, 6, 10, 11 | — (human review of the rendered book; front-matter placeholders remain) |
| 13 | Production-facing reframing & economic reconciliation | `2026-10-28-cgo-production-reframing-design.md` | `2026-10-28-cgo-production-reframing.md` | **Done** (book working tree, never committed) | 12 | — (human review of the rendered book; production access and independent verification remain open) |
| 14 | Operator overview & navigation | `2026-10-29-cgo-operator-overview-design.md` | `2026-10-29-cgo-operator-overview.md` | **Done** (builder pass; verifier pending) | 8, 13 | Independent verifier re-runs `npm test`, `npm run build` and `python tools/verify.py` |
| 15 | Interface v2 + academic title | `2026-10-29-cgo-interface-v2-design.md` | — (plan after spec review) | **Approved** (revised after critique) | 8, 10, 14 | Review the revised spec, then write the implementation plan; no code yet |

## What each live plan covers

### Slices 1–3 — enterprise rebuild (done)

- **Plan 1:** enterprise capacity-depreciation and pooled-APN economics, audit
  core, sealed reports, direction-aware `down`/`both`.
- **Plan 2:** hardened `dualns` bench, campaign matrix, multi-flow JFI,
  `clean`/`bursty` channels, provenance labels.
- **Plan 3:** additive API surface, audit live/stop, dashboard hardening,
  accessibility and copy. The frozen corpus was not touched.

### Slice 4 — eBPF ground truth (blocked)

- Scope: eBPF probe under `pkg/probe`: retransmit bytes and socket RTT, `tcp_retransmit_skb`
  raw tracepoint, `tcp_rcv_established` socket sample, `kfree_skb` qdisc-drop
  counter, all-or-nothing per window, `source: ebpf` provenance.
- Scope guard: no BBRv3, `.dist`, IFB, UI, or thesis work in this slice.
- Blocker: no committed `bpf_probe.o`; Windows has no clang/bpftool/BTF. The
  CGO-DEV VM is reachable through PC2 but its SSH rejects every recorded
  key/user, and PC2's WSL is 6.18 without clang/bpftool/tc. Do not call eBPF
  "verified" before the live test.
- Evidence 2026-10-29: the CO-RE source (`pkg/probe/bpf/bpf_probe.c`),
  `BUILD.md`, `vmlinux.h` ignore rule and `make bpf`/`make bpf-check` targets
  are committed. No object, loader, or hash exists; the next step is exactly
  `make bpf` on a stable Linux 6.8 host, then the Go loader/tests.

### Slice 5 — trace-driven `.dist` (Windows-green)

- Scope: strict telemetry reader, baseline-only quantile table, sealed native
  netem `.dist` + sidecar, opt-in `CGO_DELAY_MODEL=dist`, fail-closed loading.
- Scope guard: no IFB, BBRv3, eBPF, UI, or thesis work here.
- Evidence: 2026-10-29 on DESKTOP-B6K9UCM: `go test ./pkg/audit ./pkg/qdisc
  ./cmd/cgo ./pkg/campagne` → `ok`; `TestReadTelemetry*`, `TestBuildCalibration*`,
  `TestWriteLoadCalibration*`, `TestApplyPhysicsDist*` and
  `TestRunCalibrateWritesSealedArtifacts` pass; `bash -n kit/install-bbr3.sh`
  clean; full `python tools/verify.py` → 6 passed. No `.dist` artifact is
  committed; tests use `t.TempDir()`. Task 8 remains blocked.
- Next: produce a real field trace of at least 30 baseline samples, run
  `cgo calibrate`, then the Linux 6.8 netem acceptance. Never cite a
  `SMOKE-TEST` `.dist`.

### Slice 6 — ingress IFB shaping (Windows-green)

- Scope: Linux IFB ingress commands in `pkg/remediate`, direction-aware
  `BuildReport`, RouterOS LAN queue tree, opt-in `CGO_SHAPE_MODE=ingress`
  campaign on `dualns`/`down`, IFB preflight, bench `--ingress on`.
- Scope guard: no BBRv3, eBPF, `.dist`, UI, thesis, or `both`-direction
  ingress campaign.
- Blocker: live acceptance needs the Linux 6.8 bench. Windows proves command
  generation and orchestration only.
- Evidence: independent verifier pass 2026-10-28 at `ad203f9` on
  DESKTOP-B6K9UCM: all seven re-runs exit 0 except the pre-existing
  `cmd/cgo-gui` cross-build, four fault proofs red-then-green, six gates
  `6 passed`. Engine row `VERIFIED`; Task 9 row stays `BLOCKED`.

### Slice 7 — BBRv3 evaluation matrix (Windows-green)

- Scope: additive `bbr3` label gated on an exact token in
  `/proc/sys/net/ipv4/tcp_available_congestion_control`; P2 ×
  `{pfifo_fast,fq_codel,cake}` × `{cubic,bbr,bbr3}` × 3 reps, `direction=up`,
  `flows=4`; documented command, not a new preset.
- Scope guard: `bbr` keeps meaning BBRv1; no UI, ECN/L4S, IFB, `.dist`, eBPF,
  or thesis chapter. The frozen corpus and Tableau 15 stay untouched.
- Evidence: 2026-10-29 on DESKTOP-B6K9UCM: `go test ./pkg/model ./pkg/doctor
  ./pkg/campagne ./cmd/cgo` → `ok`; exact-token gating refuses explicit `bbr3`
  on a stock host and accepts it when the stub kernel exposes it; the kernel
  release is frozen in the matrix config; `bash -n kit/install-bbr3.sh` clean;
  `GOOS=linux go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` → exit 0.
  Full `python tools/verify.py` → 6 passed. Task 8 remains blocked.
- Next: build and boot the 7.1 BBRv3 kernel on the bench VM, then run Task 8;
  `bbr` still means BBRv1. No live BBRv3 claim is made.

### Slice 8 — audit report surface & thesis citation (builder pass)

- Scope: type and render the ingress remediation the report already carries;
  render `phases`; use the shared `Provenance` line; add a "cite this report"
  Markdown copy with the report id and full `source_sha256`; correct
  `docs/api.md` for `/api/remediate`, `/api/audit/presets` and the CSV export.
- Scope guard: no `book/` edit (the thesis paste is the human's), no Go or CSV
  shape change, no new audit form field, chart, or dependency, no eBPF/IFB
  live badge.
- Evidence: 2026-10-29 on DESKTOP-B6K9UCM: `npm test` → 40 files, 156 tests
  passed (6 new: citation and SSR DOM); `npm run build` → typecheck + Vite +
  bundle, exit 0 (339.9 KB gz); `go test ./pkg/api/... ./cmd/cgo/...` → `ok`;
  full `python tools/verify.py` → first run hit the E5-family Windows
  `TempDir` cleanup flake in `TestMatrixFilteredSubmatrix` (no Go change in
  this slice), targeted re-run `PASS`, `--only tests` `PASS`, full re-run
  `6 passed, 0 failed, 0 not run`. No live audit; fixture reports only.
- Next: independent verifier re-runs the commands and owns the `VERIFIED` row;
  the human pastes the citation into the thesis.

### Slice 9 — engineering system (done)

- Scope: root `AGENTS.md`, three ledgers (`docs/VERIFY.md`,
  `docs/CHECKPOINTS.md`, `docs/ERRORS.md`), six gates with one fault proof each,
  pre-commit/pre-push hooks, WIP and Friday cadence.
- Scope guard: no product code, no CSV schema change, no frozen run, no eBPF,
  IFB, BBRv3, `.dist`, UI, or thesis work; no CI, no document register, no
  weekly file, no full mutation harness.
- Evidence: independent verifier pass 2026-10-28 on DESKTOP-B6K9UCM, at
  `04fad4d`: six gates `VERIFIED`, hooks `VERIFIED`, `--selftest` six `caught`,
  gate total 596 lines, frozen audit and scoped Linux build green. One
  hermeticity defect was found and fixed (`04fad4d`), then re-proved in a fresh
  clone. `make` is absent on this host, so the targets were run as
  `python tools/verify.py`; the full Linux cross-build still fails only in the
  pre-existing `cmd/cgo-gui` Fyne package. Merge to `main` is the human's
  decision.

### Slice 10 — live interface observation (done)

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
  delta). The browser gate passes on the plan's chart-advance criterion:
  259 real samples in 25.23 s (10.27/s), 0 long tasks. Repaints measured
  7.97/s, bounded by this host's 45–52 fps compositor; both numbers are in
  `docs/VERIFY.md` and 10 repaints/s is not claimed.
- Plan: `2026-10-27-cgo-live-interface-observation.md` (Tasks 1–10).
- Next action: —. The Phase 2 production adapter waits for operator
  read-only access.

### Slice 11 — frozen corpus seal anchor (done)

- Scope: a committed SHA-256 anchor over the 159 run manifests, checked by the
  `frozen` gate; fail closed when `book/` is present and the anchor does not
  match. Closes the self-referential seal the slice 9 verifier named.
- Scope guard: no change to `book/`, its audit, CSV schemas, or frozen runs;
  no signing key, new gate, dependency, UI, or thesis work.
- Evidence: independent verifier pass 2026-10-28 at `06bc310` on
  DESKTOP-B6K9UCM: 159 `OK` from `sha256sum -c`, 9 unit tests, six gates
  `6 passed`, six `caught`; a temp copy with `quarantine.json` and its manifest
  rewritten consistently passed the audit and was caught by the seal. Gate
  layer 649 lines ≤ 650. Recorded follow-up: the pre-existing SKIP on a partial
  `book/` still says "book/ absent"; it exits 2, so it cannot certify.

### Slice 12 — book truth-up and prose restructure (done)

- Scope: reconcile `book/content/*.js` with `PRODUCT.md`, the umbrella spec §9,
  and `docs/VERIFY.md` — title, RQs, field-audit claim, economic framing, corpus
  seal, live interface chart, IFB ingress, regime label — then restructure
  Chapter III (open with the campaign, move the UI walkthrough to Annex E and
  the installer detail to Annex C) and cut the old body prose.
- Scope guard: no new measurement, no corpus change, no production claim, no
  hostname/IP, no title beyond `PRODUCT.md`, no UI redesign, no `book/` commit.
- Evidence: 2026-10-28 on DESKTOP-B6K9UCM: `python book/tools/audit-archive.py`
  → exit 0, both `OK` lines; ten modules load; `node book/build-docx.js` →
  `book/memoire-rebuild.docx` (3054 Ko); `memoire.docx` hash unchanged;
  `audit-biblio` 47/47 cited; body 14,761 words, total 16,462 (baseline 24,373);
  31 numeric tokens lost, all justified in the checkpoint; old title, the
  eight-gate claim, the three-sites claim and the field-data framing absent from
  the rendered document. `book/` is working-tree only, never staged.
- Next: human review of the rendered book; fill supervisor/jury/date
  placeholders; Master vs Licence wording remains the human's decision.

### Slice 13 — production-facing reframing and economic reconciliation

- Scope: adopt the latest IDEA title (« Architecture d’ordonnancement et de
  régulation de trafic… ») with an honest subtitle, in `PRODUCT.md`, the book
  title page and the cover; add the edge-architecture framing, the enterprise
  Model A/B/C and WIS 2.0 SLA paragraph, the blocked-directive status, and the
  source-of-truth headers.
- Scope guard: no production measurement or control, no hostname/IP, no live
  bench, no UI redesign, no new table/annex, no `book/` commit. The memos supply
  the title direction only; their numbers and production claims stay excluded.
- Evidence: 2026-10-28 on DESKTOP-B6K9UCM: `python book/tools/audit-archive.py`
  → exit 0, both `OK` lines; all ten modules load; `node book/build-docx.js` →
  `book/memoire-rebuild.docx` (3055 Ko); `memoire.docx` hash unchanged;
  `audit-biblio` 47/47 cited; body 15,042 words, total 16,743 (within the
  slice's 15,300/17,000 ceilings); numeric diff against the slice-12 snapshot
  shows 0 lost and 8 added tokens, all contractual-scenario or blocked-
  directive values; new title present in `word/document.xml`, old title absent.
  `book/` is working-tree only, never staged.
- Unproven: no production measurement, no field audit, no Linux live
  acceptance; independent verifier pending.

### Slice 14 — operator overview and navigation (builder pass)

- Scope: add the Synthèse landing (frozen-proof counts, aggregated P2 constat,
  next action, last audit, four shortcuts), authored SVG nav icons, contextual
  chart toggles, and first-visit mounting of the heavy views.
- Scope guard: no redesign of the four existing views, no new data claim, no
  new backend endpoint or dependency, no production access.
- Evidence: 2026-10-29 on DESKTOP-B6K9UCM: `cd web/frontend && npm test` → 41
  files, 161 tests passed; `npm run build` → typecheck + Vite + bundle, 342.3 KB
  gz; screenshots at 1920×1080 and 1366×1080 show 159 runs / 208 valid / 74
  degraded / 246 quarantined and the aggregate P2 constat; full
  `python tools/verify.py` → 6 passed.
- Next: independent verifier re-runs the frontend gates; visual review stays
  the human's.

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
