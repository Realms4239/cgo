# Passive field trace out3 — 2026-11-26

Operator approval: 2026-11-26 "I approve the passive capture", "I aprove the task token write itup", "Add one exact sudo verb", "Install and restore". Task-scoped token `C:\WEBSERVER\APPROVED-2026-11-26-cgo-field-capture.md`. No guard or sudoers change retained: guard restored byte-exact from `safe_write.py.pre-cgo-field-20261126` (sha256 `c4282af7…`, selftest back to 954 cases), and the original two-line sudoers fragment restored (`20261126-032158`, sha `9641423a…`). Password used only through the private card and `safe_write.py`'s `sudo -S` stdin path.

## Capture facts

- Binary: already-tagged `/home/sadc/labo-cgo/field-trace/cgo-observe-x`, `--duration 10800 --interval 5`, output `/home/sadc/labo-cgo/field-trace/out3`.
- Launched under PID `3373373` (sudo wrapper `3373364`); clean finish with `manifest.json` written and no process left.
- `flows.csv`: **109,489 lines** (header + 109,488 data rows in the manifest), 11,140,704 bytes, owner `sadc`, mode 600 at stop. `manifest.json`: 457 bytes, owner `root`, retrieved through a separate approved R2 pass (`APPROVED-2026-11-26-cgo-field-retrieve.md`) into a sadc-readable copy whose ownership was restored immediately after the read.
- Manifest contents (read-window evidence `20261126-064926`, sha `57aede62…`): `kind: cgo-observe/manifest`, `duration_s: 10800`, `interval_s: 5`, `exclude_ports: ""`, `excluded_rows: 0`, `rows: 109488`, `started_unix: 1795674017`, `object_sha256: 8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`, `flows_csv_sha256: 48cafc1235b197693858c075f0ae014d33e9dcce2eb3462a8e8ccd71b5082283`.
- Independent `sha256sum` of the retrieved `flows.csv` (evidence `20261126-064953`, sha `d327ec25…`) returned exactly `48cafc12…`, matching the manifest byte-for-byte.
- Retrieval write evidence: `20261126-064731` sha `0c7c3011…` (copy), `20261126-064848` sha `0e33c75e…` (brief sadc ownership), `20261126-065024` sha `254d55fc…` (final root ownership restore). Guard restored byte-exact from `safe_write.py.pre-cgo-field-retrieve-20261126` sha `c4282af7…`; selftest back to 954 cases.
- Stop/liveness evidence: `20261126-032239` sha `5c614c5c…`; final state `20261126-063025` sha `cf3d86b6…`.

## Banding (honest, bounded reads)

The read window cannot emit the whole 11 MB file in one pass, so two disjoint bounded windows were used and banded locally by the ratified floors in `cmd/tracemodel/main.go:17-21` (P1 20-40 ms excluding P4's band, P4 40-60 ms, P2 100-200 ms, P3 >=600 ms):

| Window | Evidence | Positive SRTT values | P1 | P4 | P2 | P3 | 200-600 gap | <20 ms |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| first 2,000 rows | `20261126-063206`, sha `b6b0c3de…` | 1,953 | 4 | 1,237 | 0 | 0 | 9 | 1,934 |
| last 50,000 rows | `20261126-063510` | 47,059 | 0 | 981 | 0 | 0 | 0 | 46,078 |

The windows are disjoint (head vs tail of 109,489 rows), giving **2,218 P4 samples** in the observed portions, above the 30-sample floor. P1 (4), P2 (0) and P3 (0) are below the floor. No `.dist` was produced and no calibration is claimed: the manifest hash and a full-file band count still need a separate authorized retrieval pass.

## Defect found in the tracer output

Many rows carry a literal `%!(EXTRA uint32=1)` prefix before the `kind` field (visible in `20261126-063056`, sha `e6138558…`). The SRTT column remains numeric on those rows, so banding used column 11 only. This is a formatting defect in the tracer, not a reason to discard the samples, but it must be fixed before a sealed calibration consumes the file.

## Disposition

- **Unblocked for P4 only**: a real field band with >=30 samples now exists. Next step is a separately approved retrieval of the bounded sample + manifest, a tracer formatting fix, then `cgo tracemodel` and `cgo calibrate --profile P4` with a sealed sidecar.
- **Still BLOCKED**: P1, P2, P3 have no calibratable band. Named condition: a longer authorized capture whose traffic reaches those RTT bands with >=30 samples, or a new operator ruling on the floor. Lowering the floor or inventing samples is forbidden.
- No frozen corpus, gate, ledger, `.dist`, host service, field-trace execution at the sample floor, or VERIFY-row promotion occurred in this pass.
