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
