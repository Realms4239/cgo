# Exhaustive live-bench plan — 2026-11-16/17 (operator-authorized)

Scope: live bench + kernel truth + RouterOS validation on sadcnwp, sequential
micro-tasks, one campaign at a time (single-writer lock), same-day scoped
`APPROVED-2026-11-16-exhaustive-bench.md` token, all traffic via `C:\WEBSERVER`.
Never: edit `safe_write.py`, add sudo permits, touch field trace, write
`book/data/runs`, `git add -A`, self-certify a VERIFIED row, push/tag/merge.

Oracle convention: each task names a command whose exit 0 = pass, 1 = failed
check, 2 = did not run. A task that cannot run is BLOCKED with a named
condition, never silently skipped.

Reading rule: `gate_status=valid|degraded|invalid` is a MEASUREMENT-INTEGRITY
verdict (G0 target reachable … G7 CPU not saturated, `model/types.go:98`), not
an SLA verdict. A `valid` pfifo_fast cell with `deadline_ok_pct=20.9` is an
honest measurement of bad performance — do not "fix" it into a gate failure.

## Layer 0 — Preflight and instrumentation (T01–T08)
- T01 host capacity + field-trace clock (R0) — DONE 2026-11-16T20:18Z.
- T02 bench topology up + `cgo-bench-dualns.sh check` — DONE (exit 0).
- T03 confirm committed kit scripts deployed in guest (`pkill` in down, `/small` check) — DONE.
- T04 confirm CHR image on host (sha256 pinned) — DONE.
- T05 capture `tc -s` baseline for four transit interfaces — DONE (pre-ktruth).
- T06 verify gates 6/6 + `kit/test-kit.sh` before any live run — DONE.
- T07 verify frozen corpus unchanged (159/528/208) — DONE via gates.
- T08 confirm no second campaign writer before each launch — enforced by kit script.

## Layer 1 — Full qdisc/CC matrix (T09–T16)
- T09 launch 72-cell sweep — first two attempts (`run-1794860405`, `run-1794861247`) INVALIDATED:
  concurrent-audit contamination, then the `run` CLI default target 1.1.1.1 (rtt 0.0, gates failed).
  Both stopped by exact PID; no data fabricated. Relaunched with explicit `--target 10.30.0.2`.
- T10 poll CSV until 73 lines (header + 72) — pending on the valid relaunch.
- T11 assert all 72 rows have explicit gate_status — pending.
- T12 classify valid/degraded/invalid per qdisc × CC — pending.
- T13 investigate any invalid cell root cause before touching data — pending.
- T14 dump `tc -s` after matrix for drop/overlimit deltas — pending.
- T15 seal matrix evidence with own hash8 under docs/verification — pending.
- T16 checkpoint + builder-pass row (verifier pending) — pending.

## Layer 2 — Kernel truth differential (T17–T24)
- T17 fresh audit with eBPF attach + `tc -s` before/after — DONE clean (alone, tagged binary,
  canonical 100ms/0% + cake 20mbit). The first loaded-bench attempt was contaminated and invalidated.
- T18 read audit kernel_telemetry — DONE: source=ebpf, object `8612f578…`, flows 4/4,
  retrans 327248 B / 226 events, grade A+.
- T19 compare tc drops with eBPF qdisc_drops — CLOSED by kernel-source root cause: Linux 6.8
  `cake_drop()` uses plain `__qdisc_drop()` and sets no drop reason; `dev.c` frees the deferred
  list with `tcf_get_drop_reason()` = 0, so the probe's QDISC_DROP-only filter misses all 226.
  Documented in `probe/bpf/BUILD.md` (`ad86f15`); no object rebuild (clang banned on VM).
- T20 interpret retrans-at-0%-loss — RESULT: real transport retransmits caused by cake overlimit
  drops. Independent counters agree (tc delta 226 = eBPF retrans 1+225=226); not 0%-loss
  invention. `qdisc_drop_packets=0` is a known blind spot, never to be reported as zero loss.
- T21-T22 clean vs loaded comparison — the loaded-bench audit was invalidated (concurrent
  writers); only the clean audit is citable. No loaded comparison was fabricated.
- T23 probe fix — DEFERRED with named condition: needs an object rebuild (clang on a build host,
  not the VM) or a separate `tc -s` qstats field in the report. Not shippable from this VM.
- T21 repeat audit in clean window after Layer 1 ends (no campaign) — pending.
- T22 compare loaded vs clean telemetry; name confounder — pending.
- T23 if mismatch, instrument root cause in code with a regression test — pending.
- T24 seal kernel-truth evidence + checkpoint — pending.

## Layer 3 — Direction and flow sweeps (T25–T34)
- T25 up-direction P2 × cake × {cubic,bbr} × reps 3 × flows 4 (18 cells).
- T26 both-direction P2 same grid (18 cells).
- T27 one-flow P2 × cake × {cubic,bbr} × reps 3 (6 cells) — historical parity.
- T28 two-flow variant (6 cells).
- T29 eight-flow saturation (6 cells).
- T30 compare flow-count impact on goodput and small p95.
- T31 compare up vs down RTT asymmetry against profile spec.
- T32 capture `tc -s` per sweep for drops.
- T33 seal direction/flow evidence.
- T34 checkpoint.

## Layer 4 — Profile edge cases (T35–T42)
- T35 P3 (VSAT 600 ms) qdisc comparison with deadline 1000 ms (not 220).
- T36 P3 at default 220 ms — expected invalid, quarantine taxonomy.
- T37 P4 (LEO) jitter-orbital sweep.
- T38 P1 (fibre 80 Mb/s) saturation above shaper rate.
- T39 deadline sweep 200/220/500/1000 ms on P2/cake/cubic.
- T40 loss injection 0/0.5/1% on P2 — BLOCKED as a CLEAN sweep: `cgo run` has no
  standalone `--loss` flag (`cmd/cgo/main.go:333-345`). CORRECTION: profile `LossPct`
  IS applied to netem by the campaign (live P2 cell showed `loss 0.5%` on w-edge), so
  P1/P2/P3 give 0/0.5/1% — but they also differ in capacity/delay, so they cannot
  isolate loss. Named condition: add a netem-loss axis to the CLI with a fail-closed
  test, or shape loss in a dedicated window. Not faked by relabelling another profile.
- T41 bursty vs clean channel (P2/P3 only per canal gate).
- T42 seal edge-case evidence + quarantine taxonomy.

## Layer 5 — RouterOS validation (T43–T50)
- T43 CHR 7.24.6 boot under QEMU on sadcnwp — ALREADY DONE 2026-11-16 (row 71 VERIFIED,
  23/23, image `c6abf876…`, export `61e51573…`). No new boot while the campaign is the
  single writer; a real-audit-driven boot still needs a separate named window.
- T44 import WAN remediation; real export — PROVEN (sealed export replayed locally 23/23).
- T45 ingress (LAN) remediation — covered by the same 23 checks on the real export; no
  separate ingress-only import was performed this session.
- T46 pfifo_fast baseline script — NOT DONE; no import artifact exists.
- T47 negative control — DONE no-QEMU: `parent=ether1`→`ether9` in a copy of the real
  export → exit 1, 22/23, exact failed check `queue tree cgo-wan-out.parent`. Evidence in
  `docs/verification/2026-11-17-chr-replay/`. Reading rule: `import.accepted=true` is NOT
  a compliance verdict; require exit 0 AND all checks passing.
- T48 repeat-import idempotence — NOT DONE (needs live CHR).
- T49 serial transcripts — the prior seal references an ANSI transcript on the host
  (sha `88b492d1…`) but the transfer through the safe channel is not byte-exact.
- T50 seal matrix — partial: replay positive/negative proofs sealed under
  `docs/verification/2026-11-17-chr-replay/`.
- Named blocker for a REAL-audit CHR proof: `audit-1794798576` has
  `remediation_state=observation_only` ("no interface was named"). A 45 s audit with
  `-iface`/`-routeros-iface`/`-ros-lan` is staged in the guest (`/home/bench/eb-realrem.sh`)
  and must run in a single-writer gap; then a live CHR boot under a separate token.

## Layer 6 — Economics, honesty, consolidation (T51–T56)
- T51 recompute cost model from sealed Layer 1 cells (Ariary/h, no invented tariffs).
- T52 verify no frozen corpus file changed (`git status` + audit-archive).
- T53 verify no claim promoted to VERIFIED by builder — verifier owns rows.
- T54 write consolidated evidence README with all seal IDs and named confounders.
- T55 full gates + kit tests after all live work.
- T56 human handoff: push/freeze/tag remain your decision; field trace seal 2026-11-17T16:22:27Z.

Stop conditions: two consecutive failures of the same command → systematic-debugging
root-cause loop; any frozen-data write → abort; any `safe_write.py` edit → abort;
any field-trace change → abort; host load/critical memory → pause live campaigns.
