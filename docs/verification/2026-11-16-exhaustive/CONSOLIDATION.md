# Exhaustive live-bench consolidation — 2026-11-16/17

Single-writer campaign on the sadcnwp bench VM under same-day scoped authority
(`APPROVED-2026-11-16-exhaustive-bench.md`, `-exhaustive-continuation`, `-late`).
No frozen corpus write, no field-trace touch, no concurrent campaign.

| layer | run | cells | valid | rtt_p95 | small_p95 | goodput | drops | cost Ar/h |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| layer1 | run-1794862046 | 72 | 63 | 120.4 | 245.6 | 3.2 | 0.0 | 0.0 |
| l3up | run-1794875529 | 6 | 3 | 121.0 | 255.1 | 19.5 | 3840.5 | 575.5 |
| l3f1 | run-1794877413 | 6 | 6 | 121.0 | 254.9 | 9.8 | 1343.0 | 201.2 |
| l3f2 | run-1794878770 | 6 | 6 | 121.0 | 232.3 | 10.1 | 2647.0 | 396.6 |
| l3f8 | run-1794880033 | 6 | 6 | 120.0 | 269.9 | 13.9 | 5561.5 | 833.4 |
| l3both | run-1794900710 | 7 | 4 | 120.5 | 271.9 | 19.8 | 12.0 | 1.8 |
| l4p3d1000 | run-1794881292 | 6 | 6 | 641.2 | 1650.0 | 0.4 | 89.5 | 13.4 |
| l4p3d220 | run-1794882626 | 6 | 6 | 641.0 | 1581.5 | 0.4 | 113.0 | 16.9 |
| l4p4 | run-1794883891 | 6 | 6 | 67.8 | 91.5 | 25.9 | 10.0 | 1.5 |
| l4bursty | run-1794885178 | 6 | 6 | 121.5 | 199.3 | 14.3 | 4654.5 | 697.5 |
| l4d200 | run-1794886441 | 3 | 3 | 120.0 | 243.9 | 3.7 | 9.0 | 1.3 |
| l4d500 | run-1794895306 | 3 | 3 | 121.4 | 211.6 | 3.5 | 12.0 | 1.8 |
| l4d1000 | run-1794896065 | 3 | 3 | 121.0 | 264.2 | 3.8 | 22.0 | 3.3 |
| l4p2rep5 | run-1794896712 | 10 | 10 | 121.0 | 275.2 | 5.3 | 464.0 | 69.5 |

**14 runs / 146 cells.** All rows are `valid` or `invalid` by measurement
integrity (G0–G7), never by SLA. Every invalid row is explained in
`docs/CHECKPOINTS.md`: P1/pfifo_fast/cubic G3 straddle at the 400 ms threshold;
P4 no-AQM cubic G4 1%-of-capacity floor; P2 up cubic G4 50%-of-CapUp floor.

## Column validity
- `layer1` used the pre-fix binary: its `drops`, `retransmissions`,
  `wasted_bytes`, `cost_ar_per_h` are **structurally zero and unusable**. All
  other columns are valid. The sealed V2 corpus has the same caveat (in `SEAL.md`).
- Layers `l3*`/`l4*` used `/home/bench/cgo-fixed` (sha `d3219c01…`) with the
  `tc -j` root-counter fix (`5e00cd9`); their loss/cost columns are real,
  first non-zero figures in the campaign.

## Kernel-truth findings
- Physics oracle: declared P2 `{20 Mbit,100 ms,15 ms,0.5%}` matched live kernel
  `netem delay 50ms 15ms loss 0.5%` + `cake bandwidth 20Mbit rtt 100ms`, base
  RTT p50 98.4–102.0 ms. P1 20 ms and P3 600 ms also matched.
- Cross-run reproducibility: P1/cake median RTT p95 22.9 ms in both the sealed
  corpus and this campaign; P3 ~640 ms in both.
- eBPF retrans events (1+225=226) matched `tc -s` e-wan cake drop delta
  (227−1=226) exactly; a real second audit with eBPF telemetry measured
  421 368 retrans bytes / 5 flows.
- `qdisc_drop_packets=0` is a documented 6.8 cake blind spot (dropped with no
  reason), NOT zero loss — see `probe/bpf/BUILD.md`.
- No-AQM CC tradeoff: cubic collapses under injected loss while BBR keeps
  throughput at ~2× latency; cake bounds RTT at ~120 ms across 1/2/8 flows
  while drops/cost scale monotonically (201 → 397 → 833 Ar/h).
- P2 goodput divergence resolved by `l4p2rep5`: cubic 2.9–6.0 Mbps (median 4.2)
  brackets sealed 2.4–4.9; BBR 18.0–19.4 in 4/5 reps with one 1.3 collapse.
  Stochastic loss variance, not a measurement defect.

## RouterOS
Real-audit CHR proof (`docs/verification/2026-11-17-chr-real-audit/`):
`import.accepted=true`, **11/11**, export SHA `91d67458…`, generated from
`audit-1794898684`'s measured script. Fixture seal remains for WAN; local
copies are transcriptions, host SHA is authoritative.

## Not done / human-gated
- `git` push, freeze, tag, merge — your decision.
- VERIFY row moves — independent verifier owns them; builder promoted none in
  this campaign.
- Field trace until 2026-11-17T16:22:27Z; `.dist` KS calibration needs its seal
  plus ≥30 samples; 4.5 needs DSI countersign; 6.7 needs your hands.
- T40 clean loss sweep (no `--loss` axis), T46 pfifo baseline import, T48
  idempotence remain BLOCKED with named conditions in the plan.
