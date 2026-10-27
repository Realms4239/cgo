# CGO — agent operating card

## What this repository is

CGO (Meteolink) is a Go service and React dashboard that measure a constrained
network link: RTT, small-object latency, bulk goodput, losses, and the cost of
the service at a given capacity. `pkg/`, `cmd/`, and `web/frontend/` are the
product; `book/` holds the frozen measurement corpus and the thesis; `kit/`
holds the reproducible dual-namespace bench. The frozen corpus is evidence, not
scratch space. The bench is the only place where live measurements are real.

## Read first

1. This file.
2. The latest entry in `docs/CHECKPOINTS.md`.
3. `docs/VERIFY.md`.
4. The current row in `docs/superpowers/plans/README.md`.
5. The artifact that entry names.

If the entry's branch or HEAD differs from `git log --oneline -1`, say so
before acting. The artifact wins over the entry's prose.

## The one command

- `make verify` — all six gates; `make verify-fast` — no Go suite;
  `make verify-selftest` — the gates' own fault proofs.
- Exit 0 = every gate passed. Exit 1 = a named gate failed. Exit 2 = a gate
  could not run; **2 is not green**, and the skipped gate is named.
- These targets are installed by slice 9 Phase B. Until they exist, run the
  manual gates listed in `docs/superpowers/plans/README.md` and report a
  missing prerequisite as "not runnable here", never as passed.

## Never touch

- `book/data/runs` and the frozen corpus: no new run, rewrite, or reshape.
- The `telemetry.csv`, `link_audit.csv`, and `aqm_eval.csv` schemas.
- `epbfBBR.md`, `backup/`, `data/`, and any user document.
- `.dist` execution (no field trace at the sample floor).
- Graylog, and the remote PC `100.114.101.13`.
- The live bench without explicit human approval.
- Runtime `clang`, `gcc`, `llvm`, or `linux-headers` on the VM; the eBPF object
  is prebuilt.

## Never commit

`epbfBBR.md`, `critical-review-*.md`, `session-*.md`, `handoff-*.md`,
`backup/`, `data/`, `book/` changes, `testbed-bootstrap/`, `unp*/`, `*.docx`,
`*.pdf`, `cgo.exe`, archives. Never `git add -A`; stage explicit paths. Never
merge to `main` without the human's decision.

## Workflow

1. State the scope in one line.
2. For a new slice: write the spec, then the plan, then update the plan index.
3. Build the smallest complete change, then run the gates.
4. Write a checkpoint with the command, date, host, result, and what stayed
   unproven.
5. Commit one logical group with explicit paths.

A slice that needs two scopes is two slices. A blocker is recorded, not hidden.

## Roles

Builder, verifier, narrator, human. State the pass upfront. The builder may not
certify its own number; the verifier re-runs the command and owns the
`VERIFIED` row; the narrator only arranges verified material; the human decides
scope, live runs, and merges. For a number leaving the repository, a separate
verification pass is mandatory.

## Language

French comments and user-facing text; English identifiers, specs, plans, gate
messages, and these ledgers. A file keeps the language it already uses.

## Stop and ask

Destructive actions, a live bench run, a merge to `main`, a secret exposure, a
frozen-data change, a new dependency, or a scope that grew. Ask before acting,
not after.

## Pointers

- Six gates and their fault proofs:
  `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md` §4.4.
- Ledgers: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`.
- Plan index: `docs/superpowers/plans/README.md`.
- Umbrella spec:
  `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`.
- Product and design: `PRODUCT.md`, `DESIGN.md`.
