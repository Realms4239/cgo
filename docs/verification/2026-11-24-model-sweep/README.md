# 2026-11-24 — four-profile model sweep: sealed `.dist` vs scalar

Builder pass under `C:/WEBSERVER/APPROVED-2026-11-24-huge-model-sweep.md`.
All host contact through `safe_write.py`; no direct SSH, no host/service/config
change, no frozen corpus, gate, permit or ledger edit, no VERIFY-row promotion.
Topology restored after both arms (evidence `4a0a43ed…`).

## Design

One binary (`/home/bench/cgo-physval-20261123`), one topology, one rep,
cake, down, 4 flows, deadline 220 ms. Two sequential 8-cell campaigns:
`run-1795500982` with `CGO_DELAY_MODEL=dist` and the four sealed tables in
`/home/bench/dist`; `run-1795502736` with scalar declared delays. Same
P1/P2/P3/P4 × cubic/bbr grid in each arm. Dist launch `6bbd1888…`, scalar
launch `9cee5eb1…`.

## rtt_p95 (ms): calibrated table vs scalar

| Profile | Calibration source p95 | dist/cubic | dist/bbr | scalar/cubic | scalar/bbr |
|---|---:|---:|---:|---:|---:|
| P1 | 23.028 | 23.3 | 23.4 | 23.0 | 22.9 |
| P2 | 175.282 | 176.0 | 176.0 | 119.0 | 121.1 |
| P3 | 839.66 (field-derived) | 841.0 | 842.0 | 637.5 | 639.5 |
| P4 | 43.025 | 43.5 | 43.3 | 69.5 | 67.6 |

All 16 rows have `gate_status: valid` (measurement integrity, not an SLA). The
dist arm reproduces each sealed table within ~1 ms at p95 for both CCs; the
scalar arm sits near the declared profile delays (P1 20, P2 100, P3 600,
P4 40 ms), deliberately far from the measured distributions for P2/P3/P4.

## Congestion-control divergence in the dist arm

| Profile | cubic drops / goodput | bbr drops / goodput |
|---|---:|---:|
| P1 | 1,093 / 76.7 Mbps | 46,469 / 80.2 Mbps |
| P2 | 8 / 1.7 Mbps | 4,604 / 19.2 Mbps |
| P3 | 5 / 0.2 Mbps | 178 / 0.7 Mbps |
| P4 | 4 / 24.1 Mbps | 25,111 / 98.8 Mbps |

The same emulated delay/loss distribution produces radically different loss and
goodput under BBR versus cubic. This is a result about control-loop interaction,
not a defect in the table: matching delay does not imply matching throughput.
The scalar arm shows much smaller CC divergence for P4 (7 vs 10 drops), so the
difference is driven by the loss pattern carried in the calibrated table.

## Honest limits

- One repetition per cell and one direction (down): these are point readings,
  not distributions. No p99.9 claim.
- Provenance: P3 is field-derived (the 48 h passive trace, `.dist` sha
  `998904af…`); P1/P2/P4 are bench-derived observatory samples (shas
  `fb99def6…`, `f9c23c39…`, `619c7a88…`). Bench-derived is not field-valid.
- The new 3 h organic trace (`../2026-11-23-organic-trace/`) has no profile band
  with >=30 samples, so it contributed NO calibration here. It was not silently
  folded in, and no sample floor was relaxed.
- `gate_status: valid` is a measurement-integrity verdict, never an SLA pass.
