# 2026-11-23 — field-trace P3 `.dist` replay and physics validation (builder pass)

Guarded R2 window under `C:/WEBSERVER/APPROVED-2026-11-23-field-trace-replay.md`.
All host contact went through `safe_write.py`; no direct SSH, no host service or
config change, no frozen corpus, no field-trace mutation, no gate/permit/ledger
edit, no VERIFY-row promotion.

## Input: a real passive field trace, not a bench audit

Source: `../2026-11-19-field-trace-p3/` — 113 samples >=600 ms extracted from the
48 h passive trace (1,752,062 rows), `basis: measured`, sealed sidecar:

| Artifact | Value |
|---|---|
| `P3.dist` | 6144 bytes, sha256 `998904af7412cca245c60d5ecf76ba935ddea25d4cb8fead0e7d9069059b9499` (matched locally and on the guest) |
| `P3.json` | sha256 `429eac8ccf3066417522889a8d7a3354c01facfa6c2f2e72b4fa54f79127ae73`, 113 samples, floor 600 ms, scale 90.663 ms |
| Trace RTT | p50 839.657 ms, p95 839.657 ms, p99.9 962.642 ms (roundtrip p50/p95 839.662 ms) |

## Live A/B on the bench

Same binary (`/home/bench/cgo-physval-20261123`), same topology, same
P3/cake/cubic/down/4-flows/1-rep cell. Only the delay model differs.

| Run | Delay model | rtt_p50 | rtt_p95 | small_p95 | drops | gate_status |
|---|---|---:|---:|---:|---:|---|
| `run-1795454014` | `CGO_DELAY_MODEL=dist`, sealed P3 table | 840.0 ms | 840.0 ms | 1703.8 ms | 5 | valid |
| `run-1795454385` | scalar P3 (600 ms declared) | 601.0 ms | 645.0 ms | 1446.5 ms | 14 | valid |

Deployment evidence `326f74c3…`, replay launch `ef4fd7bb…`, replay artifacts
`b0e936cc…`, control launch `a31094c5…`, control result and topology restore
`5e17a0d1…`, all under `C:/WEBSERVER/evidence/`.

## Result

The field-derived `.dist` replay reproduced the measured RTT distribution at
p50/p95: **840.0 ms vs 839.66 ms (about +0.34 ms)**, while the scalar control
missed it by roughly **239 ms** at p50 and **195 ms** at p95. The A/B separation
proves the trace table reached the kernel on the live bench. This is the first
time this programme has shown an emulator reproducing a *measured* link's delay
distribution rather than a literature profile.

## Honest limits

- One repetition each: p99.9 (962.64 ms) is NOT validated by this run; only the
  p50/p95 region is evidenced. The replay's small_p95 (1703.8 ms) is worse than
  the control's (1446.5 ms) and drops differ (5 vs 14) — a distribution can match
  delay while other metrics diverge.
- This binary's manifest omits `delay_model` from its config string, so both
  manifests carry the same `config_sha256`; the A/B RTT difference, not the
  manifest, is the proof. That recording gap should be fixed in a later slice.
- No fresh eBPF audit was produced in these cells; no kernel-truth seal claim is
  made from them.
- No field-derived P2 exists (the passive trace has zero 100-199 ms samples), and
  no P1/P4 replay was run. This validates P3 only.
- `gate_status: valid` is a measurement-integrity verdict, never an SLA pass.

## Extended A/B: three repetitions per arm (same day)

Under `APPROVED-2026-11-23-field-replay-extended.md`, the same two launchers were
re-run with `--reps 3` (no other change):

| Arm | Run | rtt_p95 by repetition | Mean | Spread | vs trace p95 839.66 ms |
|---|---|---|---:|---:|---:|
| P3.dist replay | `run-1795454981` | 840.0 / 840.5 / 841.0 | 840.5 ms | 1.0 ms | +0.84 ms |
| scalar P3 | `run-1795455597` | 643.0 / 645.5 / 640.6 | 643.0 ms | 4.9 ms | -196.6 ms |

All six rows have `gate_status: valid`. Launch evidence `00176456…` (dist) and
`c9a3a2cf…` (scalar); topology restored with `39bc077a…`. The between-arm
separation (~197.5 ms) is roughly 40x the within-arm spread, so the match is not
a single-cell coincidence. The still-open caveats are unchanged: p99.9 is not
measured by CSV p95 rows, small_p95 and drops diverge, no fresh eBPF audit, no
P2 field-derived trace, and the manifest still omits `delay_model`.

## Remaining profiles: bench-derived `.dist` A/B (NOT organic field trace)

Under `APPROVED-2026-11-23-more-profiles-replay.md`, the sealed P1/P2/P4
calibrations from `../2026-11-19-multiprofile-trace/calibrations/` were staged
(hash-verified on the guest) and replayed as one multi-profile campaign, with a
scalar control campaign of the same profiles. **Provenance:** those sidecars say
`basis: measured`, but the samples came from the eBPF observatory watching
bench-generated shaped traffic, not a passive field capture. This validates the
`.dist` mechanism, not P1/P2/P4 field validity.

| Profile | Calibration source p95 | Replay rtt_p95 | Scalar control rtt_p95 | Replay error | Scalar error |
|---|---:|---:|---:|---:|---:|
| P1 | 23.028 ms | 23.4 ms | 22.8 ms | +0.37 ms | -0.23 ms |
| P2 | 175.282 ms | 176.0 ms | 121.8 ms | +0.72 ms | -53.5 ms |
| P4 | 43.025 ms | 43.5 ms | 68.2 ms | +0.48 ms | +25.2 ms |

Six rows, all `gate_status: valid`. Dist run `run-1795456726` (rows P1/P2/P4),
scalar control `run-1795457310`. Evidence: deploy `89dda4b1…`, dist launch
`ce21a036…`, scalar launch `30a2fc29…`, topology restore `b75bbf76…`.

This shows the emulator reproduces a sealed distribution table for every profile
tested, with the largest scalar error (P2, -53.5 ms) far exceeding replay error
(+0.72 ms). It does NOT promote bench-derived samples to field measurements: the
only organic field-derived result in this file remains the P3 replay above.
