# Plan — deadline semantics (2026-11-26)

Spec: `docs/superpowers/specs/2026-11-26-deadline-semantics-design.md`. Operator approved the semantic change in chat on 2026-11-26.

## Global Constraints

- No frozen CSV header change (23 columns stay in the same order), no gate change, no frozen corpus / `book/data/runs` / ledger / permit / host / dependency change.
- The fixed 220 ms CAP SLA keeps its number, its `metrics.AssessCAPSLA` path and its operator-policy wording. Do not rename it as an WMO requirement.
- No fabricated value, no `0` standing in for absence, no `NaN` written into CSV or JSON. Invalid rate fails closed through `metrics.ModelDelayMs`.
- TDD: red test before each implementation; one logical commit per task with explicit paths; never `git add -A`, no merge, no force push, no branch deletion.
- French comments and user-facing text; English identifiers and commit messages.
- Verification per task: touched-module `go test -count=1 ./...`, `npm -C web/frontend run typecheck`, `npm -C web/frontend test -- --run`, `sh .githooks/run-gates.sh` printing `6 passed, 0 failed, 0 not run`. Exit 2 is never green.

## Task 1 — metrics primitives (Go module `metrics`)

Red tests first:
- `DeadlineOKPct(nil, 220)` returns `math.NaN()`; a non-empty slice keeps the existing strict `< due` behaviour (75 % golden unchanged).
- `EffectiveDeadlineMs(profileDelayMs, packetBytes, rateMbps, allowanceMs)` returns `ModelDelayMs(profileDelayMs, packetBytes, rateMbps, allowanceMs)` and refuses an invalid/non-positive rate with an error.

Implement only those two functions. Do not change `AssessCAPSLA` or its 220 ms/95 % constants.

## Task 2 — campaign uses the profile-relative basis and honest absence (Go modules `model`, `pkg/campagne`, `api`)

Red tests first:
- Snapshot/event carries `deadline_measured bool` (false when the completion set was empty) and the effective due value used, in JSON/manifest provenance only.
- `campagne.go` computes `DeadlineOKPct` with `EffectiveDeadlineMs` from the profile's own `DelayMs`, capacity and the campaign's operator allowance; the existing `DeadlineMs` remains the allowance, not an absolute cross-profile due value.
- `pkg/campagne/writer.go` writes an **empty CSV cell** for a non-measured deadline and keeps the 23-column header and every other cell byte-format unchanged.
- `api` live/SSE JSON never contains `NaN`; absence is a boolean plus a word, not a number.

## Task 3 — frontend renders absence (TypeScript/DOM)

Red DOM test first: an audit/campaign card with `deadline_measured: false` shows `—` plus the French word `absent`, never `0,0 %`. Keep the measured percentage rendering unchanged.

## Task 4 — full verification

Run all gates, the frozen-corpus audit, frontend typecheck/tests and six native gates; report exact commands, exits and counts. No commit unless a real defect appears, in which case stop and report.
