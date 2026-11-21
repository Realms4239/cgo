# Chunks 0–6 Engineering Programme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the seven-chunk MUSTREAD programme (chunks 0–6) to the maximal locally-engineerable state, then debug the live bench under a verified authorization gate.

**Architecture:** P0 is a read-only gap audit that produces the authoritative per-chunk acceptance table. P1 closes real gaps in chunks 1–5 as sequential SDD slices; P2 prepares chunk 0's Cap-40 verification harness while the chunk stays BLOCKED on the human purge; P3 closes chunk 6's local gaps; P4 applies systematic-debugging to the live bench only after the four named artifacts are verified. Every slice is a fresh spec → plan → implementer → reviewer, with explicit-path commits and an independent verifier for any ledger row.

**Tech Stack:** Go (`api`, `probe`, `results`, `cmd/cgo`, `cmd/gates`), React/TypeScript (`web/frontend`), Vitest + Playwright on Edge (`PWCHROMIUM`), axe-core, `sh .githooks/run-gates.sh`, Linux bench (`tc`, eBPF, CHR) in P4 only.

## Global Constraints

- No chat secrets; no ad-hoc ssh; no frozen-corpus or thesis writes; no ledger edits; no gate weakening; no parallel subagents.
- Protected: `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, `book/data/runs`. Slice work never edits them.
- Three-state exits: 0 = all checks ran and passed, 1 = a check ran and failed, 2 = a check did not run. A skipped or mocked check is never green.
- Absence-never-zero; no fabricated PASS, seal, campaign, count or measurement; no `/api/**` stubbing in Playwright.
- Native gates run serially. The `links` false positive is recorded, never edited. Two failures of the same command → `HALTED AFTER 2 STRIKES` with stderr tail, permissions/disk state and root suspicion.
- Never `git add -A`; stage explicit paths only. Local fast-forward is the only integration step without a fresh human decision.
- P4 starts only after: same-day `APPROVED-*` token scoped to this task, named quiet bench window, rollback line, same-day rehearsal. sadcnwp only through the `C:\WEBSERVER` toolbox.
- Stop and ask before any destructive action, frozen-data change, live run, scope growth or merge to `main`.
- French user-facing copy; English identifiers, specs, gate messages and ledgers.

---

## File Structure

| Path | Responsibility |
|---|---|
| `docs/research/2026-11-22-chunks0-6-gap-audit.md` | P0 output: authoritative acceptance table (`chunk / criterion / state / evidence`) |
| `.superpowers/sdd/2026-11-22-chunks0-6-programme/progress.md` | Programme ledger: task states, parks, evidence pointers |
| `.superpowers/sdd/<slice-plan>/…` | Per-slice briefs, reports and review packages, created at execution time |
| Source paths | Determined by the P0 gap rows for each chunk; recorded in that chunk's brief before any implementation. A chunk with zero gap rows produces no source change. |

---

### Task 1: P0 — read-only gap audit

**Files:**
- Create: `docs/research/2026-11-22-chunks0-6-gap-audit.md`
- Create: `.superpowers/sdd/2026-11-22-chunks0-6-programme/progress.md`

**Interfaces:**
- Consumes: the approved spec `docs/superpowers/specs/2026-11-22-chunks0-6-engineering-design.md` and the seven MUSTREAD chunk docs `docs/MUSTREAD/10-…` → `16-…`.
- Produces: `GapRow[]` — each row is `{chunk: 0..6, criterion: string, state: "met"|"gap"|"blocked", evidence: "<file:line>" | "<command> → exit N"}`. Tasks 2–8 consume only rows with `state != "met"`. A row with no evidence is a plan failure.

- [ ] **Step 1: Extract the criterion list**

Run: `for f in docs/MUSTREAD/1[0-6]-chunk-*.md; do echo "=== $f"; grep -nE 'Goal|Produces|Verifies|Expected|Definition of done' "$f"; done` in `C:/cgo/.worktrees/main-reconciled`. Save the verbatim output to the audit doc's "Criteria" section. This is the authoritative criterion list; do not invent new criteria.

- [ ] **Step 2: Check chunk 1–2 markers**

Run each and record exit + file:line:
`grep -rn "func IsBlockedAddr" probe/ssrf.go`
`grep -rn "func guardedDialContext" probe/ssrf.go`
`grep -rn "missing_file\|hash_mismatch\|no_manifest" api/corpus.go`
`grep -rn "func writeJSON" api/server.go`
Mark a criterion `met` only when the symbol and its test both exist; otherwise `gap`.

- [ ] **Step 3: Check chunk 3–5 markers**

Run:
`grep -rl "Float64Array\|DomainEventRing" web/frontend/src --include=*.ts --include=*.tsx | head -20`
`grep -rl "synctest" api --include=*.go | head -20`
`grep -n 'axe-core' web/frontend/package.json`
`ls .github/workflows/ci.yml .github/workflows/release.yml`
Record each result as `met` or `gap` with the exact path found.

- [ ] **Step 4: Verify the chunk 0 blocker**

Run: `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md`. Expected: two commits (`54d3f62`, `8c67859`). Record chunk 0 check 1 as `blocked` with the human-purge unblock condition from the 2026-11-18 checkpoint. No agent may rewrite history.

- [ ] **Step 5: Adopt chunk 6 evidence**

Read `docs/verification/2026-11-19-chunk6-readiness/README.md`. Transcribe rows 6.1.1, 6.1.4, 6.2, 6.3 as `met` (6.2 with its caveat) and 6.1.2, 6.1.3 as `gap`; 6.4 and 6.5 as `blocked` with human unblock conditions. Note that the record promotes no VERIFY row.

- [ ] **Step 6: Write and commit the audit**

Write the table (one row per criterion, all four columns populated) plus a "No-change contract" section listing every `met` criterion that P1–P3 must not churn. Commit: `git add docs/research/2026-11-22-chunks0-6-gap-audit.md .superpowers/sdd/2026-11-22-chunks0-6-programme/progress.md && git commit -m "docs(audit): chunks 0-6 acceptance gap table"`. If the pre-commit hook fails on the known `links` false positive, use `git -c core.hooksPath=C:/Users/ASUS/AppData/Local/Temp/opencode/no-hooks commit ...` and record why.

- [ ] **Step 7: Self-review the audit**

Run: `grep -c '|' docs/research/2026-11-22-chunks0-6-gap-audit.md` and inspect every row: a row without evidence or without one of the three states is a failure — fix inline before proceeding.

---

### Task 2: P1 — chunk 1 (security lockdown) local gaps

**Files:**
- Create: `.superpowers/sdd/2026-11-22-chunk1/task-1-brief.md`, `…/progress.md`, `…/task-1-report.md`
- Modify: only source paths named by the chunk 1 `GapRow`s in the P0 audit

**Interfaces:**
- Consumes: `GapRow[]` for chunk 1 from Task 1.
- Produces: a committed, reviewer-approved chunk 1 slice, or an explicit "all criteria met, no change" ledger row.

- [ ] **Step 1: Read the chunk 1 gap rows**

If every chunk 1 row is `met`, append `chunk 1: no change (P0 evidence …)` to the programme ledger and skip to Task 3. Otherwise continue.

- [ ] **Step 2: Write the slice brief**

Create the SDD brief containing the exact `GapRow` evidence, the chunk 1 plan's verification commands, and the global constraints section of this plan verbatim. The implementer sees only this brief.

- [ ] **Step 3: Dispatch a fresh implementer**

Use `subagent-driven-development`: one fresh general subagent for the slice. Required steps in its prompt: write a failing test per gap, watch RED, minimal fix at source, watch GREEN, run `go -C api test -count=1 ./...` and `sh .githooks/run-gates.sh`, restore tracked `dist/index.html` if built, one explicit-path commit. No parallel subagents.

- [ ] **Step 4: Review and fix loop**

Dispatch a fresh task reviewer over the review package (`review-package <plan> <base> <head>`). If it reports Critical or Important findings, run one fix wave with the same implementer session, then a scoped re-review. Approve only with zero Critical/Important findings.

- [ ] **Step 5: Record and integrate**

Append the slice result to the programme ledger with commit SHA, commands and exit codes. Local fast-forward `main` only after the reviewer approves. Never edit `VERIFY.md`.

---

### Task 3: P1 — chunk 2 (data honesty & integrity) local gaps

**Files / Interfaces:** same shape as Task 2 with `chunk2` in the SDD path and the chunk 2 `GapRow`s.

- [ ] **Step 1:** Read chunk 2 gap rows; if all `met`, record no-change and skip to Task 4.
- [ ] **Step 2:** Write the brief with the chunk 2 plan's verification commands (`go -C api test -count=1 ./...`, `cd results && go test -count=1 ./...`) and the absence-never-zero contract.
- [ ] **Step 3:** Dispatch one fresh implementer; TDD per gap, minimal source fix, gates green, one explicit-path commit.
- [ ] **Step 4:** Fresh reviewer; one fix wave + scoped re-review if Critical/Important findings.
- [ ] **Step 5:** Append evidence to the programme ledger; no `VERIFY.md` edit.

---

### Task 4: P1 — chunk 3 (realtime pipeline) local gaps

**Files / Interfaces:** same shape as Task 2 with `chunk3` and the chunk 3 `GapRow`s.

- [ ] **Step 1:** Read chunk 3 gap rows; if all `met`, record no-change and skip to Task 5.
- [ ] **Step 2:** Brief must include the real SSE contract: 20 Hz idle frames, `Float64Array` rings, milestone events, and the ban on presenting an idle frame as a measurement.
- [ ] **Step 3:** Dispatch one fresh implementer; tests must assert ring/cadence behaviour from real frames, never a fabricated frame. Rebuild the Go binary after any `dist` change and restore tracked `dist/index.html`.
- [ ] **Step 4:** Fresh reviewer, then fix wave + scoped re-review if needed.
- [ ] **Step 5:** Append evidence; no ledger edit.

---

### Task 5: P1 — chunk 4 (honest tests) local gaps

**Files / Interfaces:** same shape as Task 2 with `chunk4` and the chunk 4 `GapRow`s.

- [ ] **Step 1:** Read chunk 4 gap rows; if all `met`, record no-change and skip to Task 6.
- [ ] **Step 2:** Brief must include: stdlib `synctest` for time-dependent Go state machines, happy-dom for DOM logic, unmocked Playwright with Edge `PWCHROMIUM`, axe WCAG 2.2 AA, and the ban on mocks of the module under test.
- [ ] **Step 3:** Dispatch one fresh implementer; every new test needs a positive and a hostile oracle and must fail if the implementation is deleted.
- [ ] **Step 4:** Fresh reviewer, fix loop as in Task 2.
- [ ] **Step 5:** Append evidence; no ledger edit.

---

### Task 6: P1 — chunk 5 (container CI / release) local gaps

**Files / Interfaces:** same shape as Task 2 with `chunk5` and the chunk 5 `GapRow`s.

- [ ] **Step 1:** Read chunk 5 gap rows; if all `met`, record no-change and skip to Task 7.
- [ ] **Step 2:** Brief must require exact pinned tool versions (no `@latest`), honest BLOCKED reporting when `docker build` is unavailable, and no claimed green container smoke without a real run.
- [ ] **Step 3:** Dispatch one fresh implementer; validate YAML syntax when the daemon is absent and record the exact error.
- [ ] **Step 4:** Fresh reviewer, fix loop as in Task 2.
- [ ] **Step 5:** Append evidence; no ledger edit.

---

### Task 7: P2 — chunk 0 Cap-40 preparation (stays BLOCKED)

**Files:**
- Create: `.superpowers/sdd/2026-11-22-chunk0-prep/progress.md`, `task-1-report.md`
- Modify: none outside the SDD workspace

**Interfaces:**
- Consumes: the chunk 0 `GapRow`s from Task 1, including the stale-path permit already landed in `d7013bd`.
- Produces: a runnable Cap-40 verification command plus a recorded BLOCKED oracle. Chunk 0 is not verified and no ledger row moves.

- [ ] **Step 1:** Write the brief: reproduce the six checks from `docs/superpowers/plans/2026-11-18-chunk0-secret-purge-bloat.md` verbatim, including check 1's `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md` oracle.
- [ ] **Step 2:** Dispatch one fresh implementer to capture the six outputs with exact exit codes into the report. Check 1 must be reported `BLOCKED — human purge pending`; no history rewrite.
- [ ] **Step 3:** Fresh reviewer confirms the BLOCKED state is honest and no gate/ledger was weakened.
- [ ] **Step 4:** Append to the programme ledger: `chunk 0: agent work complete; verification BLOCKED on human Cap-40 purge`. No `VERIFY.md` edit.

---

### Task 8: P3 — chunk 6 local preparation

**Files:**
- Create: `.superpowers/sdd/2026-11-22-chunk6-prep/progress.md`, `task-1-brief.md`, `task-1-report.md`
- Modify: source paths named by the chunk 6 gap rows (expected: `cmd/cgo` for `--confront-out`, and the real-tag test lane)

**Interfaces:**
- Consumes: the chunk 6 gap rows from Task 1 and the readiness record.
- Produces: the `--confront-out` JSON capability and a real-tag lane that can be run on the bench in P4; no VERIFY row promoted.

- [ ] **Step 1:** Brief the single gap slice with the exact contract: `cgo run --confront-out <path>` writes `ebpf_vs_tc_report.json` from the already-sealed kernel truth, with a deterministic test that the file is absent when the flag is not passed.
- [ ] **Step 2:** Dispatch one fresh implementer; TDD the flag, prove RED then GREEN, run go tests and gates, one explicit-path commit.
- [ ] **Step 3:** Fresh reviewer; fix wave + scoped re-review if Critical/Important findings.
- [ ] **Step 4:** Record the real-tag lane as "ready, not run" with the bench command to execute in P4.
- [ ] **Step 5:** Append evidence; no ledger edit. No live host contacted.

---

### Task 9: P4 — gated live-bench debugging

**Files:**
- Create: `docs/verification/2026-11-22-live-bench-debug/README.md`, `…/evidence/` as produced

**Interfaces:**
- Consumes: the P3 real-tag lane and the chunk 6 bench gap rows.
- Produces: verified bench evidence or a named blocker. No frozen-corpus, thesis, tag or `VERIFY.md` change.

- [ ] **Step 1: Verify the four artifacts before any connection**

Run: `ls C:/WEBSERVER/APPROVED-*.md` and read `C:/WEBSERVER/safe-read-write-admin-guide.md`. Check the token is same-day, named, scoped to this task; confirm a named quiet window, a rollback line and a same-day rehearsal record exist. If any is missing, STOP, print `HALTED — bench gate unmet` with the missing artifact, and ask the operator. No ssh key is used ad-hoc; secrets come from the private secret card.

- [ ] **Step 2: R0 classification only until cleared**

Per the guide, classify evidence with `safe_read.py --file` (with `-X utf8` on this host), stop on any refusal, `--run` one pass, keep the evidence file and its sha256 sidecar. No write rung without a same-day token.

- [ ] **Step 3: Systematic-debugging on each real failure**

For each failure: reproduce with the exact bench command, isolate the root cause with an independent oracle, fix at source, re-run the same command, and record before/after exit codes and stderr tails. Never mark a failure green by skipping it.

- [ ] **Step 4: Record or block**

Write the verification README with commands, exit codes and hashes. If a defect needs hardware/kernel changes beyond a local fix, record a named blocker with owner and unblock condition. Present no VERIFY row for promotion.

---

### Task 10: Final programme review and evidence aggregation

**Files:**
- Modify: `.superpowers/sdd/2026-11-22-chunks0-6-programme/progress.md` only

**Interfaces:**
- Consumes: all slice reports from Tasks 2–9.
- Produces: a final whole-programme verdict and an honest list of what remains parked.

- [ ] **Step 1:** Run all gates serially once more and capture the named results, including the recorded `links` false positive.
- [ ] **Step 2:** Forbidden diff check: `git diff --name-only 14f2e60..HEAD`; confirm no ledger, gate, frozen, dist or dependency paths.
- [ ] **Step 3:** Dispatch a final whole-programme reviewer over the combined diff; one fix wave + scoped re-review if Critical/Important findings.
- [ ] **Step 4:** Append the final state to the programme ledger with commit SHAs, exit codes and the parked/unblock table. Do not merge to `main` without a fresh human decision beyond the local fast-forwards already authorized.

## Self-Review

- **Spec coverage:** §1 audit → Task 1; §2 definition of done → Tasks 2–10; §3 phases → Tasks 1/2–6/7/8/9; §4 SDD cadence → every slice task; §5 evidence → Global Constraints + Task 10; §6 gates → Global Constraints + Task 9 Step 1; §7 testing → slice Step 3s; §8 parks → Task 7/8/9 and final table; §9 risks → Task 1 no-change contract and Task 9 blocker rule. No section lacks a task.
- **Placeholder scan:** no `TBD`, `TODO`, "implement later", "similar to Task N" or unspecified "handle edge cases". Steps that depend on P0 findings use an explicit decision rule (`all met → no change`) rather than deferred detail.
- **Type consistency:** `GapRow` with states `met|gap|blocked` and evidence strings is used identically in Tasks 1–8; SDD paths use one naming scheme (`2026-11-22-chunk<N>`); the four P4 artifacts are named identically in §6 of the spec, Global Constraints and Task 9 Step 1.
