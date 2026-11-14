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
- The targets exist since slice 9 Phase B. On a host without `make`, run
  `sh .githooks/run-gates.sh` with the same flags and say so; never report a
  missing prerequisite as passed.

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

## Cadence

- At most three active slices; blocked slices do not count. A fourth active
  slice is refused until one closes or is deferred.
- Friday routine: run the full `make verify`, update the `VERIFY` rows touched
  that week, append a checkpoint, and name any blocker older than 48 hours with
  its owner and unblock condition.
- A slice closes only when its acceptance runs and its rows move to `VERIFIED`
  or `BLOCKED` with a named condition. "Windows-green" is not "done".

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

## WEBSERVER (sadcnwp) access

Every interaction with the sadcnwp production host goes through the operator's
`C:\WEBSERVER` toolbox — never through ad-hoc ssh from an agent session:

- Read first: `C:\WEBSERVER\safe-read-write-admin-guide.md` (the harness, the
  rungs R0–R3, the token rules, measured host facts). Password source:
  `C:\cgo\mdp.txt` — read it into the command's environment, never print it,
  never copy it into this repo, never type it into a file.
- R0 only by default: classify with `safe_read.py --file`, stop on any refusal,
  `--run` one pass, keep the evidence file and its sha256 sidecar. On this host
  `safe_read.py` needs `-X utf8` (console ANSI page mangles UTF-8 otherwise);
  patching the tool is a human decision, not an agent fix.
- `ftp2` compromise: out of scope until the operator rewrites the standing
  2026-10-27 direction. Any change to that direction lands as a new checkpoint
  entry, never as an edit in place.
- R2/R3 (writes, LXD starts, bridges, tc, sudo): only with a same-day named
  token in `C:\WEBSERVER\APPROVED-*.md`, a rollback line, and a same-day
  rehearsal. No token → no connection that can write.
- Production load is real: schedule bench windows only in a quiet window the
  operator names. Evidence copied into this repo keeps the tool's redactions
  and sidecar.

## FAST TOOLS — standing arsenal (ingested 2026-11-14)

Source: `FINAL PLAN/FAST TOOLs.md` (ingested 2026-11-14). Read it before
touching tests, gates, CI, probing, or data scripts. Not a reference appendix:
when a number in it applies here, the rule applies here.

**Build and test at the lightest lawful tier.**
- Pure logic tests run plain; DOM assertions stay in DOM tests; browser e2e is
  the exception. Real timers only under an explicit `real` build tag on the lab
  host; any tag with zero tests behind it is deleted, never shipped.
- Time-dependent Go tests prefer the stdlib in-process virtual clock
  (synctest) over wall-clock sleeps; time-dependent SSE/timeout assertions
  belong in that lane. In-memory transports beat loopback pipes.
- Unit tests never fork `ping`, `tc`, or `sudo`; the lab host carries those.

**Never pay the OS tax in a hot loop.**
- No fork/exec per sample; no fresh http.Transport per sample (one shared
  transport per phase, pooled buffers, idle conns closed); read binary kernel
  or rtnetlink state, not text scraped from `tc`, behind an honest fallback
  that is labelled as fallback, not asserted as measurement.
- `/api` handlers that aggregate the corpus parallelize (all CPU cores,
  provenance-keyed memoization, ETag) — the 360–490 ms sequential bootstraps
  are a defect, not a platform fact.

**Frontend stays cheap to check.**
- Split pure TS tests (plain runner) from DOM tests (happy-dom class, never
  jsdom for logic); split echarts into its own manualChunk; budgets are
  measured: TOTAL 400 KB gz, echarts 350 KB gz — never recalibrated upward
  without a named justification.

**Python is for analysis, not for the critical path.**
- Corpus/thesis scans: uv + Polars (or DuckDB) + orjson + Ruff. Never
  pip + venv + pandas for routine aggregation; never Python where a gate
  needs Go's compiled guarantees.

**Honesty clause.**
- Faster never means weaker: a swap ships with before/after timings and zero
  gate relaxation, and nothing here is self-certified — the verifier re-runs
  the command. The gate runner is a compiled Go command (`cmd/gates`, landed
  2026-11-14 by direct mandate); Python survives only inside the frozen-corpus
  audit (`book/tools/audit-archive.py`) and for analysis. The exit 0/1/2
  contract, named gates, `-count=1`, and the frozen corpus are outside this
  arsenal's authority.
- Perf lanes apply directly under the standing 2026-11-14 mandate (they are
  not plan-gated): gate-runner port done; `synctest` for pkg/audit and pkg/api
  time tests; happy-dom split for DOM tests; native ICMP probing; rtnetlink-backed
  qdisc reads; parallel provenance-keyed corpus aggregation. Each lands as its
  own measured commit with the suite green after it. Any lane may be declined by
  the human; none may be landed without before/after timings in the report.
- Live bench, cap-40 purge, force-push, and any eBPF/live measurement stay
  human-gated regardless of what the arsenal would make possible. Speed
  applies to iteration, never to release decisions.

## Pointers

- Six gates and their fault proofs:
  `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md` §4.4.
- Ledgers: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`.
- Plan index: `docs/superpowers/plans/README.md`.
- Umbrella spec:
  `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`.
- Product and design: `PRODUCT.md`, `DESIGN.md`.
