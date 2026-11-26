# Spec — deadline semantics for heterogeneous links (2026-11-26)

Operator approval: 2026-11-26 chat "I approve the deadline semantic". This is the implementation of the architectural ruling recorded in `docs/research/2026-11-25-latency-objective-design-research.md` and `docs/research/2026-11-25-trace-driven-emulation-research.md`.

## Problem (evidence, not hypothesis)

- `metrics/cost.go:81-82` fixes `CAPDeadlineMs = 220` and `CAPBreachThresholdPct = 95` for the operator CAP objective. `audit/report.go:488-489` correctly keeps the campaign's `deadline_ok_pct` separate from the link audit's `cap_sla`.
- `pkg/campagne/campagne.go:649` computes the campaign metric as `metrics.DeadlineOKPct(chgSmall, d.DeadlineMs)`; `d.DeadlineMs` is a configurable campaign value (`200-5000 ms`) but the deployed matrix used 220 ms for all four profiles.
- P3 is a 5 Mbit/s / 600 ms RTT / 1 % loss VSAT profile (`audit/profile/profile_test.go:10-17`). With a 220 ms due value, all six P3 cells in `run-20261125-matrix-1` recorded `deadline_ok_pct = 0.0` while `gate_status = valid`. The eight gates (`campagne.go:698-702`) contain no deadline gate, so the pair is not contradictory, but a 0 % is a deterministic consequence of propagation, not a measurement of queue behaviour.
- `metrics.DeadlineOKPct([]float64, due)` returns `0` when the completion slice is empty (`metrics/metrics.go:48-50`), conflating absence with a measured failure. The audit path guards `len == 0`; the campaign path does not.

## Scope line

Change the campaign deadline computation and the absence encoding. The fixed 220 ms CAP SLA stays exactly as it is, labelled as operator policy. No frozen CSV header, gate, corpus, book, ledger, permit, host or dependency changes.

## Required behaviour

1. **Two explicitly named bases.**
   - `operator-fixed`: the existing 220 ms objective, used only by `metrics.AssessCAPSLA` and its reports. Unchanged numerically and in wording.
   - `profile-relative`: the campaign basis. Effective due value = profile propagation floor + operator service allowance. The arithmetic must use `metrics.ModelDelayMs` (`propagation + packetBytes*8/(rateMbps*1e6)*1000 + queueMs`) with the profile's own `DelayMs`, capacity and a declared, bounded queue allowance. No invented number; invalid rate fails closed.
2. **Absence is not zero.** `metrics.DeadlineOKPct` returns `math.NaN()` for an empty completion set. The campaign event carries a boolean `deadline_measured` in live JSON; the UI renders `metrics.AbsentValue = "—"` with a word when it is false. The CSV keeps its frozen 23-column header; a non-measured cell is written as an empty field, never `0.0` and never `NaN`.
3. **Provenance.** Each sealed campaign manifest records `deadline_basis` and the effective due value per profile. A reader can tell which basis produced a percentage without guessing.
4. **No gate change.** Deadline remains a reported metric, not one of the eight gates. A bad cell is still invalid only on the existing gates.

## Non-goals

- No change to WIS 2.0 prose or the 220 ms policy itself.
- No retroactive rewrite of frozen runs, `book/`, or prior evidence.
- No new dependency and no new endpoint unless a test forces one.
- No live run, no `.dist`, no host contact.

## Verification

- Red first for every behaviour: empty completions -> NaN; profile-relative due arithmetic with real profile values; manifest fields; empty CSV cell; UI absence word.
- `go test -count=1 ./...` in each touched Go module; `npm -C web/frontend run typecheck`, `npm -C web/frontend test -- --run`, `sh .githooks/run-gates.sh` printing `6 passed, 0 failed, 0 not run`.
- A frozen-corpus audit must still pass unchanged (`PASS frozen`, same counts).
