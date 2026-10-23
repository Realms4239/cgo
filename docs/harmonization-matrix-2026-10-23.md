# Harmonization matrix — codebase vs. thesis

*Generated 2026-10-23. Every row was verified against the working tree, not against
the thesis prose. Where a claim could not be verified, the row says so.*

Scope: the brief asked for a bidirectional reconciliation between the
implementation and the technical book, with non-destructive resolutions. Nothing
in this pass deletes or rewrites a thesis chapter; where the code was the wrong
side, the code was fixed. Where the thesis was the wrong side, the row records
the correction needed and whether it has been applied.

---

## 1. Audit engine

| Documented claim / specification | Actual implementation status | Resolution |
|---|---|---|
| The audit measures idle vs loaded latency, small-object completion and throughput | **Now true.** `cgo audit` wires `Bulk` and `SmallURL` through CLI flags (`-bulk`, `-small`); before this pass both were empty and the audit never applied load. Verified: `cgo audit -bulk "" -small ""` returns `note ""` and warns `bufferbloat NON MESURÉ`. | Fixed in code. Flags added, defaults point at the testbed sink. |
| A bufferbloat grade is produced | **Conditional.** Grade is emitted only when `bulkBytes > 0` and loaded samples exist. No load ⇒ empty grade, no A+ . | Fixed in code (`pkg/audit/audit.go`). |
| The grade uses p95 | **True.** `bloatDelta = loadedP95 − idleP95`. The earlier median-based grade is gone. | Fixed in code. |
| Profile matching (P1–P4) uses delay and capacity | **True.** `MatchProfile` requires measured throughput; the delay-only path that matched a 4G link to the P4 Starlink profile is gone. | Fixed in code (`pkg/audit/match.go`). |
| Loaded window overlaps the load | **True.** The window is derived from the bulk's actual run, not a fixed `duration − 8` offset. Regression tests assert overlap at 31 s and 60 s. | Fixed in code. |
| An audit produces an operator artifact | **Now true.** `-report out.md` / `out.json` emits measured values, an explicit *not measured* section, a SHA-256 seal of the source CSV, and remediation. | New (`pkg/audit/report.go`). |

## 2. Quality gates

| Documented claim / specification | Actual implementation status | Resolution |
|---|---|---|
| Eight gates G0–G7 | **True as a set**, but the thesis already discloses that G7 was not instrumented on the frozen corpus. | No change needed; thesis is honest. |
| G7 (CPU < 90%) actually guards | **Now true for new runs.** `readCPUIdleGuess()` returning 0 is replaced by a real `/proc/stat` sampler; an unmeasured CPU leaves the gate unevaluated rather than passing. | Fixed in code (`pkg/campagne/cpu_linux.go`). |
| Gate semantics match the brief's G0–G7 list | **Divergence, deliberate.** The brief renumbers the gates (G1 = baseline stability, G2 = bulk initiation, G5 = Mathis bound, G7 = cooldown recovery). The implementation uses G0 reachability, G1 bulk started, G2 probes producing, G3 latency plausible, G4 throughput coherent, G5 no duplicate rows, G6 baseline stable, G7 CPU. | **Not changed.** The thesis documents the implemented numbering in Tables 9 and 23. Renumbering would invalidate 208 frozen rows and the thesis tables for no measurement benefit. The brief's Mathis bound is a *model consistency* check, not a data-quality gate; it belongs in analysis, not in the freeze path. |
| G5 enforces the Mathis throughput model | **Not implemented as a gate.** No Mathis bound exists in the codebase. | **Open.** See §6. |

## 3. Testbed topology

| Documented claim / specification | Actual implementation status | Resolution |
|---|---|---|
| netem and CAKE must not be stacked (netem's internal FIFO would fake the bufferbloat) | **Now true, and verified on Linux.** A new four-namespace bench (`kit/cgo-bench-dualns.sh`) puts netem on one interface in `ns_wan_emulator` and CAKE on separate interfaces in `ns_edge_router`. Verified: `tc qdisc show` reports `netem` on `w-edge` and `cake` on `e-srv`/`e-wan`, with neither containing the other. | New script, validated on the dev VM. |
| CAKE must be a root qdisc, never a child of a shaper | **True in the campaign** (`ApplyShaper` uses `root cake`). **Now also true in remediation output** — a test asserts cake is never emitted under `parent 1:`. | Enforced in `pkg/remediate`. |
| `initcwnd 50` applied on both sides | **Was broken.** `ip route change` ran before addressing, so the connected route did not exist and the change failed silently. | Fixed in both `kit/testbed.sh` and the new bench; verified present in `ip route show` output. |
| BBR available to the bench | **Was broken.** Only `reno cubic` were loaded; BBR cells silently underperformed. | Fixed: `tcp_bbr` preloaded, persisted via `modules-load.d`, and the bench refuses to start without it. |
| Testbed self-check is meaningful | **Was over-strict** (single exact-string match on route output). | Fixed: separate assertions, actual state echoed on failure. |
| The bench shapes the direction under test | **Was false in the first draft of the new bench, and this is the important finding.** Two bugs, both caught by running it: (1) with the server listening inside `ns_edge_router`, an upload flow terminates at the router's local address and never egresses a shaped interface — measured 193–215 Mbit/s through a 20 Mbit/s shaper; (2) replacing root CAKE with root `pfifo` removes the rate limit entirely, so the flow ran at 204 Mbit/s and never queued. | Fixed: the server now has its own namespace (`ns_server`) behind the shaped interface. The bench's `check` asserts CAKE is present on the upload path. **This is why the bench must be run, not just written.** |

## 4. Corpus and provenance

| Documented claim / specification | Actual implementation status | Resolution |
|---|---|---|
| 528 events, 208 valid, reproducible | **True.** `book/tools/audit-archive.py` now exits 0 only on a match: `159 runs, 528 rows, 208 valid, 74 degraded, 246 invalid; 22 cells`. | Fixed. The `or True` vacuity is gone. |
| Run count | **Was 160 including the smoke run.** | Fixed: corpus of record is **159 runs**; smoke excluded; stats, inventory, embedded seed and thesis aligned. |
| `config_sha256` in every manifest | **Was the literal string `"config"`.** | Fixed: manifests store and hash the observed run configuration. |
| Every run has a manifest | **Was 25 of 160 missing.** | Fixed: all 159 manifests rebuilt; `cgo verify` now demands one per run. |
| Checker fails on mismatch | **Was always-OK.** | Fixed: strict comparison, exit 1 on first inconsistency, all errors listed. |

## 5. Statistics

| Documented claim / specification | Actual implementation status | Resolution |
|---|---|---|
| Two-sided Mann-Whitney U | **True.** `metrics.MannWhitneyTwoSided` — rank-based, tie-corrected average ranks, normal approximation with continuity correction. | No change. Already correct and tested. |
| Cliff's delta effect size | **True.** `metrics.CliffDelta` plus `CliffInterp` with Vargha–Delaney thresholds. | No change. Already correct and tested. |
| Bootstrap CI on the median | **True.** `BootstrapMedianCI95`, fixed seed 42, Park–Miller LCG, no external dependency. | No change. |

## 6. Gaps that remain open

These are real and are **not** resolved by this pass. They are listed so the
thesis does not claim them.

| Gap | Status | Why it is open |
|---|---|---|
| **Live OTA mode** (bind to a physical 4G/Starlink interface) | **Not implemented.** No `--iface` binding exists in the probe path; `SO_BINDTODEVICE` is not used. | The audit measures *through* whatever route the host has, which is sufficient for a field audit on a single-homed laptop. True OTA binding matters for multi-homed gateways and would need `Dialer.Control` + `SO_BINDTODEVICE` in `pkg/probe`, plus a raw-socket path for ICMP. That is a real change with real test cost, and it cannot be verified on this Windows host (no Linux, no 4G dongle). **Do not claim it in the thesis.** |
| **Mathis bound as a gate** | **Not implemented.** | The brief specifies `Throughput ≤ 1.22·MSS/(RTT·√p)` as G5. Implementing it as a *freeze gate* would retroactively invalidate rows whose loss estimate is coarse. It is better as an analysis-time consistency check. Recommend adding it to `pkg/metrics` as a diagnostic, not to the gate pipeline. |
| **Deadline compliance in the audit report** | **Declared unavailable, by design.** | The link audit retains percentiles, not per-probe completions, so `deadline_ok_pct` cannot be computed from it. The campaign records it. The report says so instead of inventing a number. |
| **State journal for power-loss resume** | **Partially present.** Resume of an interrupted run works (`resume_test.go`, `matrix.go` skip-journal). There is no SQLite/WAL journal; state is CSV + manifest. | CSV append + manifest is already crash-tolerant for the failure mode that matters (a cell either has a complete row or does not). A WAL would be a rewrite for no measured benefit. **Do not claim SQLite.** |
| **RouterOS `fq_codel`** | **Does not exist on RouterOS.** | The generator emits `sfq` and states in the script that it is *not* fq_codel. Claiming otherwise would be false. |

## 7. What this pass changed in the tree

New:
- `pkg/remediate/` — tc + RouterOS v7 remediation generator, 15 tests.
- `pkg/audit/report.go` — audit report engine (JSON + Markdown, provenance seal).
- `pkg/audit/report_test.go` — 11 tests, including the no-load-is-honest property.
- `pkg/audit/report_fixture_test.go` — opt-in fixture writer.
- `kit/cgo-bench-dualns.sh` — four-namespace bench isolating link physics from
  queue policy; validated on Linux.
- `docs/harmonization-matrix-2026-10-23.md` — this document.

Modified:
- `cmd/cgo/main.go` — `-iface`, `-routeros-iface`, `-report`, `-qdisc`, `-asymmetric`, `-cpu-constrained`.
- `pkg/audit/audit.go`, `match.go` — honest grading, load overlap, p95 delta.
- `pkg/campagne/` — real config hash, non-vacuous G7, CPU sampler.
- `kit/testbed.sh` — BBR, initcwnd ordering, meaningful check.
- `book/tools/audit-archive.py` — strict checker.
- Corpus — 159 runs, manifests rebuilt, smoke excluded.

## 8. Verification status

| Check | Result |
|---|---|
| `go build ./...` | clean |
| `go test ./pkg/remediate/...` | 15/15 pass |
| `go test ./pkg/audit/...` | pass (incl. 11 report tests) |
| `go test ./pkg/campagne/... ./pkg/api/... ./cmd/...` | pass |
| `go test ./pkg/probe/ -run TestPingLiveLoopback` | **fails on the Windows host**: ICMP to 127.0.0.1 is blocked at the OS level (`ping -n 2 127.0.0.1` → 100% loss). Pre-existing, environmental. |
| `python book/tools/audit-archive.py` | `OK — 159 runs, 528 rows, 208 valid` |
| **Linux re-verification (dev VM)** | **Done.** `pkg/remediate`, `pkg/audit`, `pkg/campagne`, `pkg/api` all pass on Ubuntu 24.04. |
| **Dual-namespace bench on Linux** | **Done.** Four namespaces up; `check` reports isolation; `show` confirms netem on `w-edge`, cake on `e-srv` and `e-wan`. |
| **End-to-end audit through the bench** | **Done.** Real saturated audit produced a report with a measured grade, a VoIP R-score, a SHA-256 seal, and both remediation scripts. |

### Measured results from the bench

Same topology, same 20 Mbit/s shaper, same 100 ms delay, 40 s audit, BBR bulk flow.
Only the qdisc on the upload path changed.

| Qdisc on upload path | Throughput | Idle p95 | Loaded p95 | Bloat delta | Grade |
|---|---|---|---|---|---|
| CAKE (root) | 14.55 Mbit/s | 116.3 ms | 125.1 ms | **8.8 ms** | A |
| TBF + pfifo (1000p) | 2.50 Mbit/s | 109.3 ms | 141.4 ms | **32.1 ms** | A |

This is the thesis's central claim reproduced on a bench that cannot fake it: the
AQM holds the queue (8.8 ms of added latency) while the FIFO builds one (32.1 ms),
and the FIFO also collapses goodput to 2.5 Mbit/s. The grade letter is the same
because both deltas fall inside the A band; the delta, the QDI and the goodput are
where the difference lives.

Note on the pfifo row: `pfifo` at the root carries no rate limit, so the first
attempt ran at 204 Mbit/s and never queued at all. The row above is TBF-shaped
pfifo, which is the honest FIFO baseline. A FIFO without a shaper is not a
bufferbloat baseline, it is an unshaped link.
