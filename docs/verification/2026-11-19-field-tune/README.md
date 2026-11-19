# 2026-11-19 — Field-tuned `.dist` calibration + replay (bounded dualns cell)

Operator approval: "i approved an authorized audit of the actual constrained link whose
baseline RTT matches a defined profile (≥30 baseline samples), calibrate, then replay
with TC_LIB_DIR".

## Source audit — `audit-1795091574-1fd85cca`

- 60 s dualns audit, `--link-type fiber --site bench-calibration --direction down --flows 4`,
  kernel 6.8.0-147-generic, 20 Mbit CAKE.
- 35 baseline samples, p50 110.0 ms / p95 111.0 ms.
- A first audit at exactly 100 ms (`audit-1795091086-252026ff`) was refused by the
  engine's `p50 > plancher` rule. The bench delay was raised to 110 ms — one variable
  at a time — and the audit re-run. Both audits are kept as evidence of that step.
- `audit-1795091574-1fd85cca` exited 2 (no valid probe report in the sink-detection
  fallback) but its telemetry is intact and is the calibration source only; no
  measurement claim is made from its audit verdict.

## Calibration — `calibrations/`

`go run ./cmd/cgo calibrate --audit audit-1795091574-1fd85cca --profile P2 \
  --data docs/verification/2026-11-19-field-tune \
  --out docs/verification/2026-11-19-field-tune/calibrations`

| artifact | bytes | sha256 |
|---|---|---|
| `P2.dist` | 6144 | `495ef5f3b10dff4e2e1c802672115452c565f4371e8e540af3cdbbf6216895bb` |
| `P2.json` | 960 | `0b6c0efb3f1fb6e03d481c6c2e19942f08d64c6d1b52e7f8e3f3e7c57745fc8c` |

Sidecar seals `source_telemetry_sha256=e130dc842261bf06…` (matches the audit telemetry)
and `dist_sha256=495ef5f3…` (matches the computed `P2.dist` hash). 35 samples,
`below_floor_pct=0`, `clipped_pct=0`, `scale_ms=2.750083925901059`, `hop_delay_ms=50`.

## Replay — `replay/run-dist-1795091600`

Both calibration files were copied to the guest at `/home/bench/dist/` (hashes matched
locally ↔ guest). Command:

```
sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 CGO_DELAY_MODEL=dist \
  CGO_CALIBRATION_DIR=/home/bench/dist \
  ip netns exec ns_client /home/bench/cgo-20261119 run --bench dualns \
  --run-id run-dist-1795091600 --profiles P2 --reps 1 --qdiscs cake --cc cubic \
  --direction down --target 10.30.0.2 --deadline 220 --flows 4 \
  --data /home/bench/replay-data
```

Exit 0. `manifest.json` records `"delay_model":"dist"` and `config_sha256`
`74488bf2…`, which recomputes exactly from the embedded config string.

| artifact | bytes | sha256 |
|---|---|---|
| `aqm_eval.csv` | 363 | `69dc65a76c5a9f5cbc0d7445f663c965354f8084f3b80df76278dd3f5e5f7da0` |
| `manifest.json` | 843 | `0fff9f332382965b481c21b80ea8b86c99aec0f4feda1267283e479bae9d572f` |
| `valuation.json` | 616 | `c7e514cfd4c68b86d22f7e76279579b884b56054507f411861157c324221da04` |

Both hashes declared in `manifest.json` match the computed file hashes. The row is
`gate_status=valid`:

- RTT p50/p95 **110.0 / 111.0 ms** (base 110.0 / 111.0) — matches the calibrated profile
- small p95 252.9 ms, deadline_ok 63.6 %, goodput 7.3 Mbit/s
- drops 0, retransmissions 0, wasted bytes 0, cost 0.00 Ar/h, CPU 4.0 %

### Oracle: was the `.dist` really applied?

During the run, the physics iface reported
`qdisc netem 1: root refcnt 5 limit 1000 delay 50ms  2.75ms loss 0.5%`.
The 2.75 ms is the calibrated `scale_ms`; that value is only present when the
distribution is loaded through `TC_LIB_DIR` — a scalar run would show a plain
`delay 50ms`. This is the independent assertion, not the campaign's own claim.

A first replay attempt was interrupted when the SSH client hit a 120 s tool timeout
and SIGTERM'd the binary before it sealed `manifest.json`/`valuation.json` (only a
partial `aqm_eval.csv` remained). The stale lock and partial directory were removed
and the cell was re-run under a background command that stayed attached for the full
duration. The interrupted artifacts were not used.

## Teardown

`sudo -n bash kit/run-dualns.sh down` → device `e-wan` gone, netns list empty, no
`cgo-20261119` process left (a surviving `testbedsrv` sink was killed explicitly).
The script's `/root/cgo/cgo-linux` warning is the same known teardown noise as the
sealed live cell.

## Proves / does not prove

Proves: a profile-matching ≥30-sample audit calibrates a sealed `.dist`, and a bounded
dualns replay actually loads it through `TC_LIB_DIR` and seals `valid` with hashes that
verify independently.

Does not prove: CHR (separate BLOCKED record), BBRv3, IFB ingress, thesis identity,
v2.0.0 freeze. No merge, tag, or ledger change is made by this slice.
