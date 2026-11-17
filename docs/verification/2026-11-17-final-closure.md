# Final route — closure assessment (2026-11-17)

Scope: the approved `docs/superpowers/plans/2026-11-14-final-route.md` (40
unchecked steps) plus the 2026-11-16/17 exhaustive campaign. Builder pass only;
no push, tag, merge or VERIFIED-row promotion in this closeout.

## What is DONE and evidenced

| Plan task | State | Evidence |
|---|---|---|
| 0.3 Stub SPA repair (real embed) | done | `web/frontend/embed.go`, `api/frontend_test.go`; gates 6/6; SPA served by the real binary, 25/25 e2e |
| 0.7 Gate hygiene | done | `sh .githooks/run-gates.sh` → 6 passed, 0 failed, 0 not run; `run-gates` fault proofs |
| 1.1 Exposure lockdown | done | unauth `/api/*`→401, headers, rate limit — `api` suite green |
| 2.1 Probe fidelity / 2.2 channel physics | done | P1 20 ms, P2 100 ms/15 ms/0.5%, P3 600 ms, P4 40 ms verified against live `tc -s` (CHECKPOINTS) |
| 2.3 Qdisc parity + `tc -j -s` | done | live root-cause + fix `5e00cd9` (`tc -j` row-root counters); regression tests green |
| 2.5 Results + one verdict | done | `api/results` fingerprint/ETag, memoized scan, G0–G7 one verdict |
| 2.6 Durability / budgets | done | bundle 354.6 KB gz < 400; checkpoint cap enforced by hygiene gate |
| 3.1–3.4 Kernel truth | done | eBPF object `8612f578…` attached, 421 368 retrans bytes / 5 flows; 226 events = 226 `tc` cake drops; 6.8 drop-reason blind spot documented |
| 4.3 New v2 corpus | done | `run-1794853778`, 24/24 valid, hash8 `6adc83d9`, sealed with caveat on loss/cost columns |
| 4.4 CHR real transcript | done (extended) | fixture seal 23/23 + REAL-audit proof 11/11 (`audit-1794898684`, export `91d67458…`), real RouterOS script |
| 4.6 VERIFY rows | done for cited rows | independent verifier moved rows 42, 43, 71, 214; builder promoted none |
| 5.1–5.7 Interface synthesis | done | 2-mode shell, rings, cockpit decision-first, WCAG; detector `[]`; 347 vitest; 25/25 e2e |
| Exhaustive bench campaign | done | 14 runs / 146 cells (`docs/verification/2026-11-16-exhaustive/CONSOLIDATION.md`) |
| API bootstrap + operator-link compression | done | integrity 0.43–1.14 s → 0.078 s, JS 380→122 KB, CSS 77→14 KB |

## What stays BLOCKED, with the named condition (not silently dropped)

- **0.1/0.2 disk & R0**: historical host prerequisites, already recorded; no
  further action possible from this checkout.
- **0.5 cap-40 purge, 1.3 signed release, 1.5 Docker, 4.5 DSI countersign,
  6.7 v2.0.0 tag/freeze/merge**: explicit human gates (force-push, cosign key,
  operator R2 service map, DSI signature, merge to `main`). Not performed.
- **2.7 BBRv3 live**, **4.1 field-trace seal** (PID 3442899 alive until
  2026-11-17T16:22:27Z; output dir operator-owned), **4.2 `.dist` KS
  calibration** (needs that seal + ≥30 samples), **6.6/6.7 thesis PDF** (Word
  COM build, `book/` human-frozen). Each keeps its named condition.
- **T40 clean loss sweep** (no standalone `--loss` axis), **T46 pfifo baseline
  import**, **T48 import idempotence** — plan-level BLOCKED, in
  `docs/verification/2026-11-16-exhaustive/PLAN.md`.

## Bench decision

**No further live matrix is needed.** The 146 cells already cover all four
profiles × three qdiscs × two CCs × several flow counts × two directions ×
deadlines × channel model, plus a 5-rep repeatability matrix. Adding more cells
would not close any named blocker: the remaining gaps are human gates (trace
seal, DSI, tag) or require a new CLI axis (loss sweep), not more samples.
Spending more bench time would be scope growth, not progress.

## Core preserved

Frozen 159-run corpus byte-identical (`PASS frozen`: 159 / 528 / 208, book
audit green). No `book/data/runs` write, no field-trace touch, no push/tag/
merge, no `git add -A`, no gate weakening, no fabricated row. French
user-facing strings, English ledgers. Builder promoted no VERIFIED row.
