# 2026-11-19 — Passive field trace → P3 `.dist` (workaround)

Operator direction: "proceed to field trace workaround and CHR and operator R service map
and merge". This slice closes the field-trace-derived calibration gap identified in
`../2026-11-17-final-closure.md` (task 4.2) using the real 48 h passive trace — not a
bench audit.

## Why the earlier attempt failed and this one works

`../2026-11-19-field-trace/README.md` sampled only 80 rows per slice. The low slice was
host-local (srtt 52–92 µs) and the high slice had just 3 samples ≥600 ms, so P3 calibration
was refused. The workaround is to count and extract the profile-matching population across
**all 1,752,062 rows** of the trace, not a sample.

Full-trace `srtt_us` bands (guarded R0 readback, `evidence/p3-band-counts-20261119.txt`):

| Band | Count |
|---|---|
| 100–199 ms (`^1[0-9]{5}$`) | 0 |
| ≥600 ms (`^[6-9][0-9]{5}$`) | **113** |
| ≥1,000,000 µs | 0 |
| 100–999 ms | 68,609 |
| 200–599 ms | 68,496 |

The 113 samples ≥600 ms are a real P3 population (p50 839.657 ms, p95 839.657 ms,
p99.9 962.642 ms). Exact values are in `evidence/p3-values-readback-20261119.txt`.

R2 re-open/read/restore cycle under `APPROVED-2026-11-19-field-trace-readback.md`:
`evidence/20261119-174646-write-window.txt` (chown sadc), R0 band reads, then
`evidence/20261119-175127-write-window.txt` (chown root — rollback). The original
`out/` directory was never touched.

## Calibration

`telemetry.csv` was built from the 113 exact `srtt_us` values (µs → ms, `baseline`,
5 s offsets) under `audits/field-trace-p3-20261119/`, then:

```
go run ./cmd/cgo calibrate --audit field-trace-p3-20261119 --profile P3 \
  --data docs/verification/2026-11-19-field-trace-p3 \
  --out docs/verification/2026-11-19-field-trace-p3/calibrations
```

Exit 0. Sealed artifacts:

| artifact | bytes | sha256 |
|---|---|---|
| `P3.dist` | 6144 | `998904af7412cca245c60d5ecf76ba935ddea25d4cb8fead0e7d9069059b9499` |
| `P3.json` | 830 | (see sidecar; `source_telemetry_sha256=1fc185d9…`) |

The sidecar records `samples: 113`, `basis: measured`, `floor_ms: 600`,
`below_floor_pct: 0`, `clipped_pct: 0`, `scale_ms: 90.663`. This is field-derived
evidence, distinct from the bench-derived `P2.dist` in `../2026-11-19-field-tune/`.

## Proves / does not prove

Proves: the full passive field trace contains ≥30 P3-profile baseline samples and can
calibrate a sealed P3 `.dist` with the engine's honest floor rule.

Does not prove: a `TC_LIB_DIR` replay of this P3 `.dist` (not run — the live bench is a
human-gated window and the prior replay was P2-only); the P2/passive-trace gap (the trace
has zero 100–199 ms samples, so no field-derived P2 exists); any RouterOS, thesis or
release claim. No VERIFY row is promoted by this builder pass.
