# Two-plane kernel truth and richer netem physics — design (2026-11-23)

Authority: operator pivot 2026-11-23. The kernel-truth seal is real but was judged
"lackluster" as a *physics* result; the operator chose tc as the declared
accounting truth source and explicitly required eBPF to stay as the near-wire-speed,
near-true observation plane ("we extremely need epbf for near speed of the bytes and
near true observation"). eBPF is therefore **not demoted and not deleted**.

Research: `last30days` engine v3.27.2 (55.3s, saved raw file under
`C:/Users/ASUS/Documents/Last30Days/`) plus 20 citable web sources appended to that
raw file. Social clusters were empty or off-topic (the only ranked HN item was an
LLM-worm thread matching on the word "realistic"; Reddit 403; YouTube 0), so no
community claim is used. Primary anchors: man7.org `tc-netem(8)` and `tc-cake(8)`,
kernel `sch_netem.c` / `sch_cake.c`, iproute2 `q_netem.c`, Netdev 0x17 slides,
arXiv 2511.19213v1 and 2510.27027v1, Excentis, OpenWrt, Bufferbloat.net.

## 0. Reality check: what already exists (do not rebuild)

The `qdisc` module already implements, with tests:
- `ApplyNetemGE` / `GEModel` / `GEMeanLoss` — the documented 4-argument
  `loss gemodel p r 1-h 1-k` (ge.go);
- `ApplyNetemPhysics` — uniform loss, byte-identical to the frozen-corpus path;
- `ApplyNetemDistribution` — empirical `.dist` table with path-traversal refusal;
- `BuildCalibration` / `LoadCalibration` — trace-driven RTT-to-table calibration,
  >=30 samples, sha256-sealed sidecar (dist.go);
- `ApplyShaper` — cake / tbf / fq_codel stacking with a documented fallback.

`pkg/campagne/dualns.go` already wires GE and `.dist` into the bench, both
directions, with fail-closed topology and calibration-validity checks.

Therefore this programme adds only what is genuinely missing. It does **not**
re-implement GE, distributions, calibration, or the shaper.

## 1. Two-plane truth model (architecture, no schema bump)

The seal and confrontation stay at `cgo.kernel-truth-seal/1` and
`cgo.ebpf-vs-tc/1` — archived evidence must remain readable. The model is made
explicit in provenance fields and operator copy, never by ranking one plane:

| Plane | Authoritative for | Source | Never claim |
|---|---|---|---|
| tc accounting | drops, overlimits, backlog, BLUE state, delay counters | `tc -s qdisc` kernel counters (`sch_netem.c`, `sch_cake.c`) | per-packet timing at wire speed |
| eBPF temporal | per-packet retrans events, srtt, byte counts at near-wire speed | sealed audit `kernel_telemetry` source=ebpf | that a retrans *reason* is a proven root cause |

The confrontation is a **consistency proof between the two planes**, not a
primary/cross-check hierarchy. "Primary" from the earlier spec is re-labelled as
"declared accounting source"; eBPF retains its first-class temporal role. No
gate, permit, ledger, schema version or archived seal is edited.

## 2. Gap 1 — jitter correlation (executable)

man7: `delay TIME [ JITTER [ CORRELATION ]]`; correlation is explicitly an
approximation, not true statistical correlation. OpenWrt shows 25% as normal
operator practice. The bench currently emits `delay <d> <j>` with no correlation.

Add `Profile.JitterCorrelation float64` (`json:"jitter_correlation,omitempty"`,
range 0..1, default 0). Emit the third delay token only when > 0, in
`qdisc.applyNetemArgs`, before any `distribution` token (grammar order), and
**as a percentage with an explicit `%` suffix**: a profile fraction of `0.25`
must be emitted as `25%`, not `0.25`. iproute2 `q_netem.c` parses delay
correlation with `get_percent()` -> `parse_percent()`, which divides by 100
unconditionally; a bare `0.25` would be 0.25 %, making the lever ~100x too weak.
This is the identical encoding rule already applied to `loss gemodel` in
`qdisc/ge.go`. Wire it from `pkg/campagne.applyPhysics` and from the CSV import
as optional column 10.
Never silently coerce an out-of-range value: refuse before any `tc` mutation.

Byte-identical invariant: with correlation 0, every emitted tc argument vector
is exactly today's — the frozen corpus stays reproducible and existing hostile
oracles keep passing unchanged.

## 3. Gap 2 — truth-plane provenance on physics artifacts (executable)

`qdisc.Calibration` gains `PhysicsPlane string` (`json:"physics_plane,omitempty"`)
with the closed vocabulary `ebpf_observation`, `rtt_baseline`, `literature`.
New calibrations must declare a plane; loading an archived sidecar without the
field is allowed but reports `rtt_baseline (legacy)` — never promoted to eBPF
provenance it cannot prove. Unknown values fail closed before a run.

The confrontation report gains an additive `truth_planes` object naming tc as
the accounting surface and eBPF as the temporal surface. No numeric field,
schema version, PASS state or existing seal is changed.

## 4. Gap 3 — serialization rate / packet overhead (named human-gated park)

`sch_netem.c:packet_time_ns()` gives deterministic per-byte serialization
(`len * NSEC_PER_SEC / rate`), with cell/overhead quantization. Today the bench
models bandwidth with a flat `cake bandwidth` / `tbf rate` cap; there is no
explicit per-packet serialization or link overhead on the propagation hop.

This is intentionally **not** switched on by default. Adding netem `rate` makes
the propagation hop a bottleneck and can suppress the AQM queueing the campaign
exists to measure. It needs a live A/B run (same cell, serialization on/off,
tc accounting + eBPF temporal both recorded) before becoming a declared profile
parameter. Recorded as a park with named unblock condition, not silently shipped.

## 5. Acceptance

- Correlation: RED then GREEN hostile oracles (`c > 1` and `c < 0` refused before
  tc; `c = 0` vector byte-identical; `c = 0.25` emits the third token in the
  right position, with and without `distribution`); campaign and CSV paths
  covered; Go suites and build green.
- Provenance: RED then GREEN oracles for closed vocabulary, legacy label and
  refusal; confrontation copy names both planes without a hierarchy claim.
- No gate relaxation, no ledger/frozen-corpus edit, no fabricated value, no new
  dependency, no live bench in these two tasks.
- Each task: spec -> plan -> fresh implementer -> fresh reviewer -> gates ->
  explicit-path commit. Sequential, one worker, no parallel subagents.
