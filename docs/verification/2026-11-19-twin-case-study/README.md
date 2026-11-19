# Twin-derived case study

This case study is complete on the CGO side. Its subject is the reproducible
constrained-link twin and the passive observatory on the bench host, not the
national forecasting server. No operator data was supplied, so VERIFY row 74
stays BLOCKED on the written grant (a) to (d). The service map, before and after
cost figures, and DSI signature remain explicit gaps rather than invented facts.

The five deliverables of the case study contract are filled below with sealed
CGO evidence, each claim labelled `measured`, `configured`, `generated-only` or
`inferred`. The owner of the operator-facing request is
`docs/verification/2026-11-19-service-map-packet/README.md`.

## 1. Service map

No service class is declared by an operator. The classes below are defined by
observed RTT regime, so every identity claim is `inferred` and unverified.

| Class | Observed regime | Evidence | Confidence |
|---|---|---|---|
| A, low latency | Host-local flows, srtt 52 to 92 us | `docs/verification/2026-11-19-field-trace/README.md` head slice | `inferred`, service identity unknown |
| B, middle latency | 68,496 samples in 200 to 599 ms | `docs/verification/2026-11-19-field-trace-p3/evidence/p3-band-counts-20261119.txt` | `inferred`, service identity unknown |
| C, high latency | 113 samples at or above 600 ms, p50 839.657 ms | `docs/verification/2026-11-19-field-trace-p3/README.md` | `inferred`, service identity unknown |

`[input wanted: operator-declared service list, termination points and direction]`

## 2. Workload model

The workload is configured by the bench, not observed from a production
service. Profile values are recorded in `docs/verification/2026-11-17-final-closure.md`
and cross checked against live `tc -s` output there.

| Profile | Delay | Jitter | Loss | Evidence level |
|---|---|---|---|---|
| P1 | 20 ms | configured | configured | `configured` |
| P2 | 100 ms | 15 ms | 0.5 % | `configured` |
| P3 | 600 ms | configured | configured | `configured` |
| P4 | 40 ms | configured | configured | `configured` |

Audit shape: 60 s split 10/40/10, 4 flows, direction down, deadline 220 ms.
The shape is a locked constant in `docs/MUSTREAD/INDEX.md`.

## 3. SLA candidate

| Field | Value | Source |
|---|---|---|
| CAP objective | 220 ms small-object p95 | `docs/MUSTREAD/INDEX.md` |
| Operator target | `[input wanted]` | none supplied |
| Target source | `[input wanted]` | none supplied |

## 4. Experiment matrix

The 24-cell corpus covers P1 to P4, cake, cubic and bbr, 3 repetitions,
direction down, 4 flows and the 220 ms deadline. Evidence is
`docs/verification/2026-11-16-v2-corpus/run-1794853778/`; all 24 cells sealed
`gate_status=valid` under the verifier pass recorded in `docs/VERIFY.md`.
The exhaustive campaign adds 14 runs and 146 cells, recorded in
`docs/verification/2026-11-16-exhaustive/CONSOLIDATION.md`.

| Class | Reproducing cell | Qdisc | Congestion control | Reproducible | Reason if not |
|---|---|---|---|---|---|
| A, inferred low latency | none | not applicable | not applicable | no | The bench minimum is P1 at 20 ms. Sub-100 us host-local flows in the passive trace cannot be instantiated by this bench. |
| B, inferred middle latency | P2, cake | cake | cubic, bbr | yes | 24-cell corpus, down, 4 flows |
| C, inferred high latency | P3, cake | cake | cubic, bbr | configured only | P3 `.dist` calibrated from 113 field samples, no replay cell sealed for P3 |

The sub-100 us case is the required non-reproducible row. A matrix claiming
total fidelity is refused by the contract.

## 5. Remediation plan

| Field | Value | Evidence level |
|---|---|---|
| Control point | CAKE on the egress queue of the constrained link | `measured` in the twin |
| Generated configuration | RouterOS queue tree and queue type script | `generated-only` for production |
| RouterOS acceptance | 11 of 11 checks, `accepted: true`, pinned image `c6abf876` | `measured` on the CHR guest, see `docs/verification/2026-11-19-chr-verifier-replay/README.md` |
| Applier | `[input wanted: named operator role]` | none supplied |
| Rollback | `[input wanted: operator-defined]` | none supplied |

No generated configuration was applied to a production host.

## Before and after, twin measured

P2, cake, down, 4 flows, 220 ms deadline. The 2026-11-16 reduced matrix in
`docs/VERIFY.md` row 42 sealed three repetitions per congestion control.
Goodput and small p95 are measured twin values. Cost is a contractual scenario,
not a measurement, so its cells stay empty.

| Metric | Before, cubic | After, bbr | Evidence level |
|---|---|---|---|
| Small-object p95, ms | 205.8 / 219.8 / 244.2 | 257.0 / 246.1 / 279.0 | `measured`, 3 repetitions |
| Bulk goodput, Mbit/s | 2.8 / 6.2 / 3.3 | 16.5 / 17.1 / 17.2 | `measured`, 3 repetitions |
| Deadline compliance, % | 95.5 / 94.5 / 90.7 | 91.7 / 88.9 / 91.7 | `measured`, 3 repetitions |
| Cost, Ariary per hour | `[input wanted]` | `[input wanted]` | contractual scenario only |

The table is a comparison of two congestion controls on the same emulated
link. It is not an operator-applied remediation and supports no production
cost claim.

## Countersign

| Field | Value |
|---|---|
| Service map reviewed by | `[input wanted: DSI name and role]` |
| Figures approved for the thesis | `[input wanted]` |
| Pseudonym required | `[input wanted]` |
| Signature | `[input wanted: DSI signature]` |
| Date | `[input wanted]` |

## Acceptance check

1. Every service map row has a confidence label: yes, all `inferred`.
2. No hostname, address, credential or topology element appears: yes.
3. At least one non-reproducible cell is named: yes, the sub-100 us case.
4. `docs/VERIFY.md` gains no row for unread or unmeasured material: yes, row 74
   stays BLOCKED on the named grant.
5. DSI signature is present before any figure leaves the machines: not met, the
   signature is `[input wanted]`.

This document is not countersigned and is not an operator case study. It is a
complete twin-side draft that the DSI can fill and sign when the grant exists.
