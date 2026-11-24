# 2026-11-24 — new organic eBPF trace (bounded 3 h capture)

Builder pass. Host contact only through `C:\WEBSERVER` guarded tools, under
`APPROVED-2026-11-23-organic-capture-2.md` and
`APPROVED-2026-11-24-organic-retrieval.md`. No direct SSH, no frozen corpus,
gate, ledger or permit left changed. Operator approved the one narrow sudoers
addition and the three exact readback commands; both were rolled back.

## Capture

Existing tagged host binary `/home/sadc/labo-cgo/field-trace/cgo-observe-x`
(object sha `8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`,
no runtime clang), launched with
`--duration 10800 --interval 5 --exclude-ports 3333,4444,14444,4555 --out <new out2>`.
Clean stop at `2026-11-24T02:13:52Z` (host-local time): **109,550 rows, 0 excluded**.

| Artifact | Value |
|---|---|
| `out2/manifest.json` | sha256 `9cc06f0a9d12882ad62f6b5a3a9d907c19ffa6162020263ea130f685715c2d07`; duration_s 10800, interval_s 5, object sha as above |
| `out2/flows.csv` | sha256 `8e04ecd7c8c929fa7e155e3bc634723fdf1726a18f566dfe9dcc45ea80c14983` (11,244,861 bytes, stays on host) |
| stop proof | `20261123-231435` evidence `785e4e6c…` |
| retrieval | `20261124-060301` evidence `202b1e1f…` |
| ownership rollback | `20261124-060634` evidence `3aa607cd…` |

Repo copies: `manifest.json` (byte-exact), plus bounded extracts of `flows.csv`
(`flows-sample.csv` 1,199 rows, `flows-band0/1.csv` 599 rows each,
`flows-p3.csv`) retrieved through R0 read windows; they are SAMPLES, not the
full trace. Extract hashes: `19b566a1…` (sample), `42436cfa…` (bands),
`a794fa12…` (p3 probe).

## Result: no calibratable profile band

Banded by the ratified floors (P1 20-99 excluding 40-59, P2 100-199, P3 >=600,
P4 40-59 ms) over all retrieved rows:

| Band | Samples | >=30 floor |
|---|---:|---|
| P1 | 0 | no |
| P2 | 0 | no |
| P3 | 3 | no |
| P4 | 11 | no |

Most observations are either host-local (~0.013 ms, below every floor) or in
the 200-600 ms gap that no profile defines. **No `.dist` was produced and no
calibration was claimed.** The 30-sample rule and P4 precedence were applied as
ratified; lowering the floor or inventing samples would break the evidence
chain. This mirrors the 2026-11-19 result where the passive trace had no P2 band.

## Disposition

Task 1 of the plan is complete (capture + retrieval + provenance). Task 2's code
is complete and reviewed (`cced336`: `source_ref` provenance, P4 precedence).
Task 3 (live A/B replay) is **BLOCKED** on the missing calibrations. Named unblock
condition: a longer authorized capture whose traffic actually reaches a profile
band with >=30 samples, or a new operator ruling on the sample floor. No live
bench, no fabricated value, no VERIFY-row promotion.
