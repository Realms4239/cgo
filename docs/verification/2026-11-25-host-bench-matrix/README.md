# Host-server bench — full dualns matrix, 2026-11-25

Host: `sadcnwp` (`sadc@41.207.49.27:9922`), already-running `labo-cgo` qemu guest reached over hostfwd `127.0.0.1:10022` with the pre-existing `bench-key-600`. Guest: the deployed `cgo` binary, `kit/run-dualns.sh`, `/home/bench/bench-data`.

Authorization: operator chat on 2026-11-25 — "window is this entire evening till night and I approve all", then the exact-scope answer "Full matrix campaign". Task-scoped token `C:\WEBSERVER\APPROVED-2026-11-25-cgo-host-bench-matrix.md` (R2, reversible, expires 23:59Z). No guard, permit, rate limit, host service, host tc/netns, bridge, VM config, field trace, frozen corpus or `book/data/runs` was changed. The ftp2 compromise and the remote PC were untouched.

Same-day rehearsal (before the socket opened): `safe_write.py --selftest` exit 0 — 954 cases, 542 allowed, 412 refused; `--check` of the exact command file ALLOW of both lines, nothing sent; near-variant `rm -rf` of the bench data path REFUSE 1 of 1 by the always-deny list, exit 1.

## Commands, exits, evidence

| Evidence file | What it was | Exit |
|---|---|---|
| `20261125-155138-read-window.txt` | R0 preflight (host load, containers, key auth) | 0 |
| `20261125-160940-write-window.txt` | First write attempt — transport failed: `paramiko` missing in the harness Python. Nothing was sent | 1 |
| `20261125-161104-write-window.txt` | `run-dualns.sh up --ingress off` — `[dualns] ok — link physics and queue policy are isolated` | 0 |
| `20261125-161148-write-window.txt` | Guest status — no campaign running, only `testbedsrv`; stale `..lock` from 2026-11-16 present | 0 |
| `20261125-172803-write-window.txt` | Full 24-cell campaign, run-id `run-20261125-matrix-1` | 0 (stream truncated by the output cap at 15/24; authoritative result below) |
| `20261125-173036-write-window.txt` | Guest result inspection — CSV, manifest and valuation read; one guessed file (`quarantine.json`) did not exist | 1 (that one read; all others exit 0) |

Each `.txt` has its original binary-mode `.sha256` sidecar from the tool, copied unchanged. `guest-result-read.log` is the local capture of the result inspection.

## First launch refused, then fixed without widening anything

The first campaign command omitted `--run-id`; `cmd/cgo/runcli.go:83-99` builds the lock path as `dataDir + "/." + runID + ".lock"`, so the empty run-id collided with the orphan `..lock` from 2026-11-16 and the binary exited 3 (`déjà en cours`). Guest `ps` proved no campaign was running. The relaunch used an explicit fresh `--run-id run-20261125-matrix-1` through a second exact, pre-classified command file. No lock was deleted and no gate was changed.

## Sealed result — 24 cells, all `valid`

Command: `cgo run --bench dualns --run-id run-20261125-matrix-1 --profiles P1,P2,P3,P4 --reps 3 --qdiscs cake --cc cubic,bbr --direction down --target 10.30.0.2 --deadline 220 --flows 4 --data /home/bench/bench-data`.

Data dir contains `aqm_eval.csv` (24 rows + header), `manifest.json`, `valuation.json`. No `quarantine.json` was present. `gate_status` is `valid` in all 24 rows. No campaign process remained after the run.

| Profile | RTT p50 / p95 (ms) | small p95 (ms) | deadline ok % | bulk (Mbit/s) | gate status |
|---|---|---|---|---|---|
| P1 (6 cells) | 20.1–20.3 / 22.8–23.0 | 30.9–32.1 | 100.0 | 72.5–75.2 | valid ×6 |
| P2 (6 cells) | 99.8–101.0 / 120.0–122.0 | 205.4–287.6 | 84.0–96.4 | 2.8–17.8 | valid ×6 |
| P3 (6 cells) | 597.0–603.0 / 637.1–644.0 | 1273.2–2068.7 | 0.0 | 0.1–0.6 | valid ×6 |
| P4 (6 cells) | 39.8–42.1 / 66.4–69.2 | 96.4–137.5 | 62.5–100.0 | 12.7–25.7 | valid ×6 |

P3's `deadline_ok_pct = 0.0` is recorded verbatim; the sealed per-cell gate still reads `valid` for all six P3 cells. That is a fact about this run, not a claim that the profile is broadly acceptable.

## What this does not prove

- The path is emulated: netem delay/loss + CAKE shaping inside namespaces, not a real Malagasy trace. No `.dist` field-trace execution.
- One matrix, one direction (`down`), one repetition count (3). No up-direction claim, no BBRv3 claim, no nsperf claim.
- Nothing was copied into the frozen corpus, `book/data/runs`, or any gate ledger. No branch, tag or remote was changed by this bench.
- The dualns topology was left up; the rollback (`run-dualns.sh down`) was written and ready but not executed because the campaign finished normally and no intervention was needed.
