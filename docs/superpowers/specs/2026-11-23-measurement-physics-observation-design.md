# Measurement physics, observation truth and fidelity — design (2026-11-23)

Authority: operator chat 2026-11-23 asked to fix the tagged/untagged eBPF system
and make profile parameters, physics and metrics a real model of reality, chose
approach "A + all three layers" and approved the design on 2026-11-23. SDD
required: one slice at a time, fresh implementer then fresh reviewer.

Research: `last30days` engine v3.27.2 (45.8s, saved under
`C:/Users/ASUS/Documents/Last30Days/`; social clusters empty or off-topic) plus
20 citable primary web sources appended to that raw file. Key anchors:
[kernel ringbuf semantics](https://docs.kernel.org/bpf/ringbuf.html),
[eBPF event-loss asymmetry](https://github.com/eunomia-bpf/eunomia.dev/blob/main/docs/ebpf-qa/2026-09-10-process-behavior-reconstruction-event-loss.md),
[INRIA fidelity thesis](https://www-sop.inria.fr/members/Chadi.Barakat/HoussamElBouananiPhD.pdf),
[Hifinet](https://www.sciencedirect.com/science/article/pii/S1389128624003633),
[NetEm empirical defect](https://folk.idi.ntnu.no/alfw/publications/icccn2011-netem.pdf),
[ITU-T G.1051 + Amd.1](https://www.itu.int/rec/T-REC-G.1051/en),
[Hypatia](https://arxiv.org/html/2510.27027v1).

No frozen corpus, book, thesis, gate, permit, ledger or protected config edit.
No fabricated value, no widened threshold, no self-certified PASS.

## 1. Observation truth (slice 1)

The `ebpf_object` build tag stays, but it must never silently determine a claim.

1. `probe.KernelSource` is extended with a closed `Observation` block embedded in
   every audit report and every kernel-truth seal: `source` (`ebpf` |
   `tcpinfo` | `absent`), `object_sha256` (present ONLY when `source=ebpf`;
   null or omitted otherwise, never a placeholder hash), `reason`, and
   `reserve_failures`.
2. A `cgo audit --preflight` verb prints the resolved source and object hash
   BEFORE any load or run. It is a read-only proof, not a gate.
3. `kerneltruth seal` refuses when `source != "ebpf"` or when the observation
   block is missing. Untagged binaries may still audit at the TCP_INFO tier,
   but no report from one may be described as eBPF observation.
4. A ring-buffer lane is added alongside the existing hash-map aggregates:
   `BPF_MAP_TYPE_RINGBUF` for streaming events plus an **out-of-band**
   `BPF_MAP_TYPE_PERCPU_ARRAY` reserve-failure counter; userspace drains with
   epoll. The kernel emits no loss record on a full ring, so any
   `reserve_failures > 0` forces the explicit note "timeline is a lower bound"
   in the report. The existing per-flow hash map and globals stay as the
   aggregate source; the ring is for event-level observation.
5. CI builds and publishes tagged and untagged variants with provenance.
   Vendoring `bpf_probe.o` into the repo, or building it in CI with clang, is a
   NAMED HUMAN-GATED PARK: this worktree has no object and the card forbids a
   runtime clang build.

## 2. Profile parameters (slice 2)

All new fields are optional and `omitempty`; absent or zero reproduces today's
tc vectors byte-for-byte, so P1-P4, L1 and V1 stay unchanged.

- `DelayDistribution` (`uniform|normal|pareto|paretonormal`) and `JitterShape`
  (milliseconds of stddev for `normal`; dimensionless tail index for `pareto`
  and `paretonormal`; unused for `uniform`). The unit is determined by the
  declared distribution, never guessed from the value. `JitterShape` must be
  > 0 whenever the distribution requires it, and must be absent for `uniform`.
- `LossModel` (`uniform|gemodel`) with per-profile `LossP`, `LossR`, `LossBad`,
  `LossGood` in the documented 0..1 fraction space; emit the 4-argument
  `loss gemodel` with percent suffixes at the tc boundary, as ge.go already does.
- `Basis` (`measured|standard|declared`) and `SourceRef` (a citation string or
  trace id). A profile whose basis is `measured` must carry a non-empty
  `SourceRef`; the engine refuses a field-validity claim without provenance.
- Validation of every numeric range happens before any tc mutation; NaN and
  infinities fail closed. CSV import gains optional trailing columns; old rows
  are unchanged. Campaign `channel_model` remains an explicit override of the
  per-profile loss model, never a silent replacement.

## 3. Physics and metrics (slice 3)

- Model delay for a packet of size `P`: `l/v + |P|/B + |Q(P)|/B`
  (propagation, size-dependent serialization, queueing). Expose it as a
  computed metric with the profile inputs that produced it, not as a measured
  value.
- Emit a configured-vs-realised table for delay, jitter, loss and rate. The
  documented NetEm jitter undershoot is a finding to display, never to hide by
  silently rescaling the configured value.
- Fidelity verdict with three states: `trust`, `bounded`, `discard`. A sample
  set with delay deviation beyond an explicit `epsilon` (INRIA (gamma, epsilon)
  definition) yields `discard`; absence of enough samples yields `bounded`,
  never `trust`. Thresholds are declared in the report and never widened to
  make a red result green.
- Observer-loss budget: `observer_events`, `observer_reserve_failures`,
  `observer_loss_pct` with explicit denominators, exported outside the
  saturated stream. `gate_status` remains a measurement-integrity verdict,
  never an SLA.

## Slices and acceptance

1. Observation truth: RED-first oracles for source mismatch, missing
   observation block, preflight output, and ring reserve-failure propagation.
2. Profile parameters: RED-first oracles for byte-identical defaults, range
   refusal before tc, per-profile GE emission and provenance enforcement.
3. Physics + fidelity: RED-first oracles for the delay model, configured-vs-realised
   rows, three-state verdict, and observer-loss budget.

Each slice: spec -> plan -> fresh implementer -> fresh reviewer -> gates ->
explicit-path commit. No live bench in these slices without a separate
same-day token; no VERIFY-row promotion from builder passes.
