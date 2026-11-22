# Exhaustive live bench - 2026-11-22

Builder pass. Single-writer campaign on the sadcnwp bench VM under the
`APPROVED-2026-11-22-exhaustive-bench.md` token and the persisted 2026-11-22
executor-card ruling. Every transition was a separate topology reset plus
`cgo-bench-dualns.sh check` (`UP_RC=0`, `CHECK_RC=0`, e-wan cake 20 Mbit).
All guarded traffic went through `C:\WEBSERVER/safe_write.py`; no direct ssh,
no frozen corpus, no field-trace write, no ledger or gate edit, no VERIFY-row
promotion (a builder pass promotes nothing).

Consolidation: `python3 /home/bench/consolidate-guest.py`, exit 0.
Evidence `20261122-213935-write-window.txt`, sha256
`d1f9234114e7ed4cd50aa92167704064205a9c6ca41c872d1e88ba13f90d7f97`.

## Result: 14 runs / 151 cells

| layer | run | cells | valid | rtt_p95 | small_p95 | goodput | drops | cost Ar/h |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| layer1 | run-1795345725 | 72 | 62 | 120.0 | 239.0 | 4.2 | 0.0 | 0.0 |
| l3up | run-1795359261 | 6 | 3 | 121.0 | 267.8 | 19.5 | 4962.5 | 743.6 |
| l3both | run-1795360439 | 12 | 9 | 121.0 | 257.3 | 17.7 | 2320.5 | 347.8 |
| l3f1 | run-1795362990 | 6 | 6 | 120.5 | 239.6 | 9.5 | 1342.5 | 201.2 |
| l3f2 | run-1795366728 | 6 | 6 | 120.5 | 287.8 | 10.7 | 2648.5 | 396.9 |
| l3f8 | run-1795370548 | 6 | 6 | 121.5 | 259.6 | 14.4 | 5876.5 | 880.6 |
| l4p3d1000 | run-1795374294 | 6 | 6 | 640.0 | 1457.5 | 0.3 | 53.0 | 8.0 |
| l4p3d220 | run-1795378024 | 6 | 6 | 641.6 | 1574.5 | 0.5 | 70.0 | 10.5 |
| l4p4 | run-1795380949 | 6 | 6 | 67.7 | 113.1 | 21.9 | 8.5 | 1.2 |
| l4bursty | run-1795382268 | 6 | 6 | 121.5 | 208.6 | 14.9 | 4807.0 | 720.4 |

Held from the 2026-11-16 campaign (not re-run this window): `l4d200`
(`run-1794886441`), `l4d500` (`run-1794895306`), `l4d1000`
(`run-1794896065`), `l4p2rep5` (`run-1794896712`). Their rows appear in the
consolidation output but are prior evidence, not new measurements.

## Reading the numbers honestly

- `gate_status` is a **measurement-integrity** verdict (G0-G7, `model/types.go`),
  never an SLA pass. An `invalid` cell is an honest measurement of bad behaviour
  or a failed precondition, not something to edit into a pass.
- **Layer 1 loss/cost columns are UNUSABLE.** `eb-layer1.stage:14` invoked
  `/home/bench/cgo`, the pre-fix binary whose `drops`, `retransmissions`,
  `wasted_bytes` and `cost_ar_per_h` are structurally zero. The same caveat is
  recorded for the prior V2 corpus. Its latency columns are sound: P1 22.9 ms,
  P2 120.0 ms, P3 639.8 ms, P4 68.1 ms medians match the ratified profile
  physics. All L3/L4 cells used `/home/bench/cgo-fixed` and carry real counters.
- P3 VSAT measures ~640 ms at both the 1000 ms and 220 ms deadlines. This
  confirms the ~600 ms profile, not a failure of the 220 ms objective, which is
  intentionally unattainable on P3.
- Flow-count sweep: 1 -> 8 concurrent flows raises drops 1342.5 -> 5876.5 and
  cost 201.2 -> 880.6 Ar/h. `both` direction is noisier (9/12 valid) than `up`
  or `down`; its 3 invalid cells are recorded, not discarded.

## Validity scope

A `last30days` research pass this session supports a conditional reading: the
bench is a legitimate controlled comparison instrument, and the eBPF kernel
truth proves the emulator obeyed its configured settings. It does not prove the
emulated link models the real path. Active fidelity monitoring (delay deviation,
size-dependent transmission delay, underlay contention) and a production/field
tier remain required for external-validity claims. The `.dist` field-trace
calibration is the bridge already built for that purpose. No such claim is made
from this builder pass.

## What stayed unproven

- Live eBPF-vs-tc confrontation: two guarded windows failed on the same proven
  cause (the pre-fix `eb-clean-report.json` omitted `qdisc_drop_packets`); the
  two-strike breaker halted further attempts. The serialization fix
  (`38d0ee2`) and label verification (`6d707ae`) are in `main`, but no fresh
  sealed kernel-truth report was produced. No seal was hand-written.
- No VERIFY row was promoted. An independent verifier must re-run the gate and
  own any row leaving the repository.
