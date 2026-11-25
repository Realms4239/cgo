# Host-side Go verification — 2026-11-25

Scope: run the CGO Go module suites on the sadcnwp host to determine whether the local
Windows `TempDir` failures were environmental. Toolchain: Go 1.26.5 linux-amd64, uploaded
to `/home/sadc/labo-cgo/cgo-gate-test/` from `git archive HEAD` (`3ff8fef` at the time of
the archive). No repository `go.mod`, dependency version, gate, ledger or frozen corpus
was changed. No live bench cell ran.

All transcripts here were produced by `safe_read.py` / `safe_write.py`. They keep the
tool's redactions; each `.txt.sha256` sidecar was re-verified against the copied bytes.

## Commands and exits (host)

| Pass | Command (in `C:\WEBSERVER`) | Exit | Result |
|---|---|---|---|
| R0 facts | `python -X utf8 safe_read.py --run --file commands/read-cgo-host-gate-check.txt` | 0 | no Go toolchain, no CGO source on host, `/home` 121 G free |
| Upload | `python -X utf8 safe_write.py --run --file commands/write-cgo-host-gate.txt` | 1 | six writes succeeded; detached launch returned `transport: exit=255` |
| Root module | `python -X utf8 safe_read.py --run --file commands/read-cgo-host-gate-poll.txt` | 0 | `gate.exit=1` (private-target test) |
| api module | `safe_write.py --run … write-cgo-host-api-test.txt`, then `safe_read.py --run … read-cgo-host-api-poll.txt` | 1 / 0 | `api-gate.exit=0`, `ok github.com/Realms4239/cgo/api 41.039s` |
| 15 modules | `safe_write.py --run … write-cgo-modules-test.txt`, then `safe_read.py --run … read-cgo-modules-poll.txt` | 1 / 0 | `modules-gate.exit=1` (audit setup, probe loopback) |
| Corrected flags | `safe_write.py --run … write-cgo-env-test.txt`, then `safe_read.py --run … read-cgo-env-poll.txt` | 1 / 0 | `env-gate.exit=1` (audit only); root and probe `ok` |
| audit via go.work | `safe_write.py --run … write-cgo-audit-work.txt`, then `safe_read.py --run … read-cgo-audit-work-poll.txt` | 1 / 0 | `audit-work.exit=0`, `ok …/audit 39.154s`, `ok …/typesgen 0.008s` |

The repeated `exit=1` on the write side is the same detached-transport behaviour: the
uploaded payload ran, and the result files were read back with R0 passes. No transport
retry was made and no permit was widened.

## Flags and host-only workspace

- `CGO_ENABLED=1` and `CGO_ALLOW_PRIVATE_TARGETS=1` are test-only process settings: the
  first allows the published `probe` package's C source, the second opts in the loopback
  targets used by `TestTestbedDownloadProtocol` and `TestBulkStatsMultiFlowLocal`.
- `src/go.work` is a host-only workspace that points `audit` at the repository's local
  `probe` module instead of the published `probe v0.1.0`. No repository file was edited.

## Verdict

All Go module suites pass on the Linux host with those settings. The local Windows
`TempDir` failures did not reproduce, so they are recorded as environmental, not as a
code defect. The full local six-gate suite is **not** claimed green by this artifact.

## Evidence

- `20261124-182308-read-window.txt` — R0 host facts, sha `a9375277120c8a546269b922158d3aa36ed1a3b0e97d8747e8157064c88bdb44`
- `20261124-192441-read-window.txt` — api green, sha `cf4d76ab2221583fed445f53e6967a5eb75ba513dc7de06087affcaa4d512ee0`
- `20261124-192853-read-window.txt` — module summary, sha `f80b4d2ee59099e911376166080616c0c1d033c02db67ba485750b4cd9de16df`
- `20261124-192913-read-window.txt` — full module log, sha `2144ada39d8e4b4c450d9c5dd06e5dea228d84afc6ce84753c5b5586513ddd9c`
- `20261124-193334-read-window.txt` — corrected-flags summary, sha `349791ad7b91f15ce28bde99b22a1c9208e60eded2b81125db0cb5a1c5531c62`
- `20261124-193354-read-window.txt` — corrected-flags full log, sha `2398ba8fcf7e06e1e0825cfa26deb65fb8b71b81ed352d731ec4f2c016b2acac`
- `20261124-193744-read-window.txt` — audit green, sha `c89b35e9376f5ba8a4cd660dd50fd958c736bc7f341c4f29bba95b590646f8f9`

The remaining files in this directory are the launch-window transcripts and their
sidecars. All sidecars were re-verified byte-for-byte after copying.
