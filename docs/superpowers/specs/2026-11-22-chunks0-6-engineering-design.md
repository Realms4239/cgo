# Chunks 0–6 engineering programme — design (2026-11-22)

## Authority and precedence

Operator authorization (2026-11-22): engineer chunks 0–6, maximal locally-engineerable
state first, then live-bench debugging. The standing operator constitution and the
CGO operating card bind throughout; the MUSTREAD pack governs the chunk definitions.
Where they conflict, MUSTREAD governs the chunk scope but an operator-level ban cannot
be waived by a project document. No chat secrets, no ad-hoc ssh, no frozen-corpus or
thesis writes, no ledger edits, no gate weakening, no parallel subagents.

## 1. Context and audit finding

- The seven-chunk MUSTREAD route is `docs/MUSTREAD/10-…` through `16-…`. Its SDD ledger
  in `INDEX.md` still marks all seven `pending`; there are no chunk SDD workspaces and
  no document claims `Chunk N verified`.
- Read-only audit 2026-11-22: chunks 1–5 already carry most acceptance markers in the
  current tree — `probe.IsBlockedAddr` + `guardedDialContext`, API integrity statuses,
  `Float64Array` rings, synctest, `@axe-core/playwright`, CI and release workflows.
  This is therefore a **gap-audit + SDD-fix** programme, not a greenfield rebuild.
- Chunk 0 check 1 is BLOCKED on the human Cap-40 purge (checkpoint
  `2026-11-18T06:51Z`; "no agent may rewrite history").
- Chunk 6 is partial: `docs/verification/2026-11-19-chunk6-readiness/README.md` records
  6.1.1 and 6.1.4 done, 6.2 done with caveats, 6.3 done (VERIFY row 71); gaps 6.1.2
  (real-tag lane not run) and 6.1.3 (`--confront-out` CLI absent); 6.4 thesis and 6.5
  v2.0.0 tag/freeze human-gated. That record promotes no VERIFY row.

## 2. Goal and definition of done

Maximal locally-engineerable state across chunks 0–6:

- every real, locally-fixable gap is closed and independently re-verified;
- evidence is assembled with exact commands, exit codes and named oracles;
- human-gated steps are parked with named unblock conditions;
- no `VERIFY.md` row is promoted without a separate verifier pass.

Out of scope until explicitly unblocked: the Cap-40 purge, frozen-corpus writes, thesis
build, v2.0.0 tag/freeze/force-push, sadcnwp writes, and any unapproved bench window.

## 3. Decomposition — phases

- **P0 — read-only gap audit.** Produce a per-chunk acceptance table with columns
  `already-met / gap / blocked`, each cell carrying file:line or command evidence.
  No writes to source or ledgers in this phase.
- **P1 — chunks 1→2→3→4→5 local gaps.** One SDD slice per chunk, in dependency order.
  Fix only gaps proven by P0; do not churn already-implemented markers.
- **P2 — chunk 0 preparation.** Build the Cap-40 verification harness/check command
  and an honest BLOCKED oracle. The chunk stays unverified until the human purge.
- **P3 — chunk 6 local preparation.** Close locally-fixable gaps (notably the
  `--confront-out` CLI for 6.1.3 and real-tag lane readiness for 6.1.2); assemble
  evidence. Bench, thesis and freeze remain gated.
- **P4 — gated live-bench debugging.** Systematic-debugging over real bench failures:
  reproduce, isolate root cause, prove with an independent oracle, fix at source,
  re-verify. Starts only after the named artifacts in §6 are verified.

## 4. SDD cadence

Each slice: its own spec → implementation plan → SDD workspace → fresh implementer →
fresh task reviewer → gates → one commit with explicit paths → local fast-forward.
Strictly sequential, one subagent at a time. Skills invoked as appropriate:
`writing-plans`, `subagent-driven-development`, `systematic-debugging`,
`verification-before-completion`. Work stays in the `main-reconciled` worktree.

## 5. Evidence and ledgers

- Each slice writes its report under its SDD workspace and updates only that workspace's
  progress ledger. `docs/CHECKPOINTS.md` and `docs/VERIFY.md` are never edited by slice work.
- Three-state exits are preserved: 0 = all checks ran and passed, 1 = a check ran and
  failed, 2 = a check did not run. A skipped or mocked check is never reported green.
- Every claim carries the exact command and numeric exit code. No mocks of the module
  under test, no self-certification; a verifier re-runs the command that owns a row.

## 6. Gates and stop conditions

- Native gates run serially. The pre-existing `links` false positive is recorded and
  never edited away. Two failures of the same command trigger `HALTED AFTER 2 STRIKES`
  with the last stderr lines, permissions/disk state and a root suspicion, awaiting the operator.
- P4 does not start until all four artifacts are verified: a same-day `APPROVED-*`
  token scoped to this task, a **named quiet bench window**, a rollback line, and a
  same-day rehearsal. sadcnwp access, if ever needed, goes through the `C:\WEBSERVER`
  toolbox only; secrets are supplied through the private secret card, never pasted in chat.
- Stop and ask before any destructive action, frozen-data change, live run, scope growth
  or merge to `main`.

## 7. Testing contract

- Per slice: typecheck + vitest/Go tests + build + unmocked Edge Playwright where
  relevant; the verifier re-runs. Absence-never-zero, no fabricated PASS/seal/campaign,
  no `/api` stubbing, no gate weakening.
- UI slices capture screenshots at 890/1280/1920 and validate them.
- A fix that touches production code carries a focused test that fails before the fix.

## 8. Parks and unblock conditions

| Item | State | Unblock condition |
|---|---|---|
| Chunk 0 Cap-40 purge | BLOCKED | Human completes the file purge and reports the empty-history oracle |
| Chunk 6 bench cells and matrix | GATED | Named quiet window + APPROVED token + rollback + rehearsal |
| Chunk 6 `--confront-out` JSON | GAP | Local CLI slice in P3, then verifier pass |
| Chunk 6 real-tag suite | GAP | Local lane slice in P3, then bench host |
| Thesis build, frozen corpus | HUMAN-OWNED | Human decision; no agent write |
| v2.0.0 tag / freeze / force-push | HUMAN-OWNED | Human decision |
| sadcnwp writes | GATED | Scoped token + toolbox rung + rollback; no ad-hoc ssh |

## 9. Risks

- The MUSTREAD SDD ledger is stale, so "pending" understates partial chunk-6 evidence;
  P0 must record actual state rather than inherit the ledger's wording.
- Most of chunks 1–5 may already satisfy their plans; the main risk is needless churn.
  P0's already-met cells are a no-change contract for P1.
- Live-bench debugging may expose defects that need hardware or kernel changes beyond a
  local fix; those are recorded as blockers, not papered over.
