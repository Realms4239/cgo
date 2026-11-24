# Rigorous emulation fidelity: tails, ingress, AQM and trace-driven realism — design (2026-11-24)

Authority: operator chat 2026-11-24 asked to close the p99.9, up-direction,
multi-qdisc and new-trace-calibration gaps, to make profile parameters and
measured values realistic for the meteo constrained link replayed locally, and
to use `last30days` plus the best-engineered execution with SDD. The combined
four-slice design was approved on 2026-11-24.

Research: three `last30days` deep passes (45.8 s, 100.2 s, 195.0 s; social
clusters empty or near-empty; saved raw files under
`C:/Users/ASUS/Documents/Last30Days/`) plus 15 citable primary sources appended
to the primary raw file. Key anchors: NetObserv's move from per-packet ring
export to in-kernel aggregation, Red Hat netstacklat (<2 % RPS loss, <6 % p99
inflation), [tc-cake(8)](https://www.man7.org/linux/man-pages/man8/tc-cake.8.html)
ingress semantics, [IFB/mirred recipes](https://ubuntu.com/server/docs/how-to/networking/traffic-shaping-tc-cake/),
the peer-reviewed (γ, ε) fidelity definition with packet-pair RTD, and
[CellReplay's EMD](https://www.usenix.org/system/files/nsdi25-sentosa.pdf).

No frozen corpus, book, thesis, gate, permit, ledger or protected config edit.
No fabricated sample, percentile, seal, PASS state or field provenance.

## 1. Slice 1 — defensible tail measurement

1. The eBPF probe already has a ring-buffer lane for events and a per-flow hash
   map. Add an in-kernel power-of-two histogram for per-packet latency samples
   (log2 bucketing, bounded memory), read through the existing snapshot path.
2. Expose `p50`, `p95` and `p99.9` from histogram bins with the bin bounds and
   sample count. A p99.9 claim requires at least 1000 samples; fewer yields an
   explicit absence word, never an extrapolated percentile.
3. Carry observer overhead as declared evidence: ring reserve failures, map-full
   count, histogram sample count and the measurement start/stop window. If
   observer loss is nonzero, the report states that tails are a lower bound.
4. Keep per-packet ring events for diagnosis, not for percentile computation.
   RED-first oracles for bin edges, small-sample refusal and loss annotation.

## 2. Slice 2 — ingress and up-direction realism

1. The bench supports `--ingress on` with IFB/mirred plumbing and a fail-closed
   `CheckIngress` preflight; all prior sweeps used `--ingress off`. Add a live
   validation of ingress mode with cake and fq_codel, checking the qdisc stats
   on the IFB and the named interface.
2. Document and enforce the semantic difference from egress: cake ingress counts
   dropped packets as transferred and keeps at least two packets queued per
   flow. The report must name that difference, not present ingress drops as
   egress drops.
3. Guard the shaping order: netem delay before the rate/AQM shaper; refuse a
   topology where the shaper would consume the injected delay.
4. Run paired up/down cells for the same profile and compare RTT, loss and
   goodput, with asymmetry explicitly declared (capacity up/down already in
   `model.Profile`).

## 3. Slice 3 — multi-qdisc fidelity gates

1. Sweep cake, fq_codel and pfifo_fast with the same sealed profile tables and
   the same traffic. Report rtt p50/p95/p99.9, small-object p95, goodput,
   drops and JFI per cell.
2. Add two independent fidelity verdicts:
   - distribution error: Earth Mover's Distance between live and emulated CDFs,
     normalized by the live mean; `trust` requires EMD <= 0.10 (10 %);
   - delay accuracy: (γ, ε) criterion from packet-pair RTD and the model
     δ = l/v + |P|/B + |Q|/B, using quantiles rather than MAE; `trust` requires
     gamma <= 0.05 and epsilon <= 0.05 x (model delay ms), evaluated over at
     least 30 packet pairs. These thresholds are declared in every report and
     are never widened to turn a red verdict green.
3. Verdict states are `trust`, `bounded`, `discard`. Insufficient samples or
   missing validation data yields `bounded` or `discard`, never `trust`.
   Calibration and validation datasets must be separate; a profile fitted on a
   trace cannot be validated on the same trace.

## 4. Slice 4 — trace-driven realism for the meteo replay

1. Correct the CloudEmu double-counting trap: when measured delay is injected
   into `tc` alongside a rate limit, apply the declared delay/rate correction
   rather than summing a measured delay on top of an already-queued shaper.
2. The new 3 h organic trace has no band with >=30 samples, so it cannot yield a
   calibration under the ratified floor. This slice must state that honestly.
   A new targeted capture is allowed only with a fresh same-day token, a named
   window of at least 6 h, and an explicit traffic source likely to reach a
   profile band; the 30-sample floor is not relaxed without a new operator
   ruling. The existing 3 h trace stays as evidence of honest absence.
3. Profile parameters must carry realistic detail already supported by the
   project: capacity up/down, delay, jitter with distribution family and shape,
   per-profile GE loss, basis and source_ref, serialization rate/overhead and
   jitter correlation. Each reported measurement names the profile fields that
   produced it.
4. Final report: what was measured on the bench, what the sealed trace says,
   where they agree, and where the local replay cannot claim field reality.

## 5. Execution and acceptance

Sequential slices; each with spec -> plan -> fresh implementer -> fresh reviewer
-> fix loops -> native gates -> explicit-path commit. Host work always under a
same-day token with reversible rollback and rehearsal. RED-first hostile oracles
for histogram bins and sample floors, ingress preflight and drop semantics,
EMD/(γ, ε) verdicts, and double-counting correction. No VERIFY-row promotion
from builder passes.
