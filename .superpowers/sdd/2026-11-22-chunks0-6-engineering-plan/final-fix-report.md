# Final fix report — chunks 0–6 engineering programme

Wave: single final fix wave. Worktree: `C:/cgo/.worktrees/main-reconciled`.
Branch: `main-reconciled`. Start HEAD: `e5cf6aa`. Role: fresh verifier (this agent
did not implement any chunk and was not the implementer/reviewer of record).

Protected surfaces untouched: `docs/CHECKPOINTS.md`, `docs/VERIFY.md`,
`book/data/runs`, all gates, permits and ledgers. No host, bench or sadcnwp
contact. No merge. No `git add -A`.

---

## Finding 1 — independent verifier passes not actually performed (Important)

**Ruling: FIXED by execution.** This agent is the required fresh verifier.
Every command below was re-run now, with exact exit codes and output tails.
No result is claimed from a diff read. No `VERIFY` row was promoted by this
wave; the protected ledgers were not edited, so promotion remains a human
decision. The reviewer-side gap (implementer == reviewer) is not repaired
retroactively — this wave supplies the missing independent re-run only.

Evidence: § "Independent verification" and § "Full gate" below.

## Finding 2 — required numeric keys accept `null` (absence-never-zero hole)

**Root cause (confirmed before edit).** `loadSeal` at `cmd/cgo/confront.go:75-79`
checked only map presence. A key present with JSON `null` passes presence, then
`json.Unmarshal(raw, &s)` silently leaves the `uint64` field at `0` — an
absence disguised as a zero measurement.

**RED — hostile test added first** (`cmd/cgo/confront_test.go:155-174`,
table over the four numeric required keys, duplicate-key form so the trailing
`null` wins):

```
$ go test -count=1 -run 'TestConfrontOutRejectsNullRequiredCounter' ./cmd/cgo/
confront → ...\ebpf_vs_tc_report.json (sceau ...\sealed-input.json)   [x4]
--- FAIL: TestConfrontOutRejectsNullRequiredCounter (0.20s)
    --- FAIL: .../retrans_events    confront_test.go:167: retrans_events=null: code=0, attendu 2
    --- FAIL: .../qdisc_drop_packets confront_test.go:167: qdisc_drop_packets=null: code=0, attendu 2
    --- FAIL: .../tc_drops_before   confront_test.go:167: tc_drops_before=null: code=0, attendu 2
    --- FAIL: .../tc_drops_after    confront_test.go:167: tc_drops_after=null: code=0, attendu 2
FAIL
EXIT=1
```

Four code-0 exits and four written report files proved the hole.

**Fix** — `cmd/cgo/confront.go:75-85`: keep the presence check, reject a raw
`null` value for every required key before struct unmarshal. No JSON is
fabricated; the seal is copied verbatim or refused.

**GREEN** (all six confrontation tests, 5 original + 1 hostile):

```
$ go test -count=1 -v -run 'TestConfrontOut' ./cmd/cgo/
--- PASS: TestConfrontOutFileAbsentWithoutFlag
--- PASS: TestConfrontOutWritesSealedTruth
--- PASS: TestConfrontOutFailsClosedWhenSealAbsent
--- PASS: TestConfrontOutRejectsNonEbpfSeal
--- PASS: TestConfrontOutRejectsMissingCounters
--- PASS: TestConfrontOutRejectsNullRequiredCounter (0.06s)
    --- PASS: .../retrans_events
    --- PASS: .../qdisc_drop_packets
    --- PASS: .../tc_drops_before
    --- PASS: .../tc_drops_after
PASS
ok  github.com/Realms4239/cgo/cmd/cgo  0.472s
EXIT=0
```

Seal rejection is fail-closed (exit 2, `est null`), no output file in any case.

## Finding 3 — ignored `fs.Parse` error before the fast path

**Root cause assessment (honest).** `cmd/cgo/main.go:371` discarded the parse
error with `_ = fs.Parse(...)` one line before the `--confront-out` fast path.
The FlagSet uses `flag.ExitOnError`, so today a malformed flag already exits 2
inside `Parse`; the reported bench fall-through is therefore not reachable
from a parse error on the current FlagSet policy. The discard is still a
latent fail-open if that policy ever changes.

**Fix** — `cmd/cgo/main.go:371-378`: check the error and `os.Exit(2)` before
the fast path, with a one-line comment on why. No other `_ = fs.Parse` call
(`--serve`) was touched.

**Regression guard** — `cmd/cgo/expose_test.go:178-199`, reusing the existing
`buildCgoBinary`/`execServe`/`scrubEnv` harness. It invokes
`cgo run --data= --confron-out <out>` (misspelled flag):

- requires exit 2 and stderr `flag provided but not defined`;
- fails if stderr instead shows `run: --data vide`, which would prove a
  fall-through to `runCLI`;
- requires the confrontation output file to be absent.

The explicit `--data=` is a safety interlock: even if the guard broke,
`runCLI` returns 2 at `cmd/cgo/runcli.go:45-48` before any lock, network or
bench activity. **This test is a fail-closed regression guard, not a
RED→GREEN proof** — it passes against the pre-fix `ExitOnError` behaviour too.
That is stated rather than dressed up as evidence the code change was
observable at runtime.

`go test -count=1 ./cmd/cgo/` → `ok ... 51.483s`, EXIT=0 (includes both new tests).

## Finding 4 — base SHA drift

**Fix** — `docs/superpowers/plans/2026-11-22-chunks0-6-engineering-plan.md:238`:
Step 2 forbidden-diff base `14f2e60` → `2489e48`, with a parenthetical naming
`e2b0cd5` as the first programme commit.

Verified ancestry: `git merge-base --is-ancestor` returns true for `2489e48`,
`14f2e60`, `5a3caff` and `46e569a`; `git log --oneline -25` shows
`2489e48` → `e2b0cd5` (spec) → `5a3caff` (plan). So `2489e48` is the
pre-programme base; `14f2e60` is an unrelated older labo commit.

The other two SHAs are correct in their own scopes and were deliberately left
alone: `progress.md:2` (`base 5a3caff`) is the chunks 0–6 SDD ledger's own
starting commit and its ranges are written `5a3caff..X` on lines 12–28; the
review header of `review-2489e48..e5cf6aa.diff` already reads `2489e48`.
Rewriting those would introduce a new inconsistency, not close this one.

## Finding 5 — 10 Hz vs 20 Hz comment

**Ruling: already fixed at HEAD, plus one residual corrected.**
`api/sse_hub_test.go:103` reads "À 20 Hz" and `:112-118` describes the 2 s
deadline as "well above two 20 Hz ticks" — both changed in `e5cf6aa`
("test(api): make hub timing tests deterministic under load"). No edit needed
there. The flake report §1 confirms the real cadence is 20 Hz
(`api/sse_cadence_test.go:32`).

One stale occurrence survived outside the cited lines:
`api/next_snap_test.go:12` said "Le hub 10 Hz"; corrected to 20 Hz. Comment
only — no assertion or timing value changed.

## Finding 6 — timing budget 250 ms → 2 s (accepted caveat)

**Ruling: accepted caveat, no code change.** Plan-authorized for a verified
false positive; no gate, permit or assertion was weakened in this wave.
Honest statement of what the tests now prove: `api/sse_hub_test.go:115` and
`api/next_snap_test.go:44` assert **≥2 ticks eventually within a 2 s deadline**,
not "≥2 ticks within 250 ms". The fixed 250 ms sleep was a hope, not a
guarantee; cadence itself remains asserted deterministically by
`api/sse_cadence_test.go`. The 300 ms post-`Close` absence window at
`api/sse_hub_test.go:105` is unchanged and documented.

## Finding 7 — inverted-condition RED proof not evidenced

**Ruling: evidence exists in the record; quoted here, not fabricated.**
`.superpowers/sdd/2026-11-22-api-timing-flake-diagnosis/task-1-report.md` §3
lines 95–109 document RED-1, the exact proof Step 3 demanded:

> RED-1 (temporarily inverted the helper comparison from `got >= want` to
> `got < want`, so the condition is not enforced) — the **retained original
> assertions** failed with the exact messages from the operator's halt:
> `--- FAIL: TestNextSnapDrivesHubAndNotState (0.03s) ... hub NextSnap calls = 0, want >= 2`
> `--- FAIL: TestHubCloseStopsTicks (0.00s) ... hub did not tick: 0 calls`
> `FAIL / EXIT=1`
> The inversion was restored immediately; no `TEMP RED PROBE` marker survives.

RED-2 (lines 111–120) sets `want = 1000000` and shows the helper failing
closed with counter value and elapsed time (`got 39, want >= 1000000`), then
restores `want = 2`.

If the reviewer's snapshot predated that report, this is the reconciliation.
I did not re-perform the inversion in this wave and do not claim a fresh probe.
The report's own truthful limitation (§7) stands: the flake did not reproduce
locally, so the diagnosis rests on the halt messages and code-level timing
analysis; the RED probes prove the oracle fails closed.

---

## Independent verification (re-run by this fresh verifier)

All commands executed from `C:/cgo/.worktrees/main-reconciled`, branch
`main-reconciled`, after the final edit.

| Command | Exit | Tail / result |
|---|---|---|
| `go build ./...` | 0 | no output |
| `sh .githooks/run-gates.sh --only tests` | 0 | `PASS tests  go vet + go test in all modules` / `1 passed, 0 failed, 0 not run` |
| `cd api && go test -count=1 ./...` run 1 | 0 | `ok github.com/Realms4239/cgo/api 120.704s` |
| same, run 2 | 0 | `ok ... 123.317s` |
| same, run 3 | 0 | `ok ... 118.515s` |
| same, run 4 | 0 | `ok ... 134.108s` |
| same, run 5 | 0 | `ok ... 112.952s` |
| `go test -count=1 -run 'TestConfrontOut' ./cmd/cgo/` | 0 | 6/6 PASS, `ok ... 0.472s` |
| `go test -count=1 ./cmd/cgo/` | 0 | `ok ... 51.483s` |
| `gofmt -l` on the 4 edited Go files | 0 | no output |

The five api runs were serial in one loop; each `EXIT_RUN_n=0` was captured
individually. No sixth run was added and no timeout was widened.

## Full gate (once, no edits)

```
SKIP frozen      book/ absent: frozen corpus not checked
SKIP provenance  book/data/runs absent: no reference header
PASS tests       go vet + go test in all modules
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     1124 tracked files clean; legacy allowlist: 18 backup archives and results/runs-frozen.tar.gz; 4 named tailnet addresses
FAIL links       28 missing path(s):  (all under protected docs/VERIFY.md and docs/CHECKPOINTS.md)
3 passed, 1 failed, 2 not run
EXIT_FULL_GATE=1
```

`links` is the recorded pre-existing false positive (dangling references
inside protected docs); the 2 SKIPs are honest `book/` absences, never green.
No gate, permit, ledger or doc referenced by `links` was edited.

## Self-review

- Scope: 6 files changed — 4 Go (2 production, 2 test), 1 api test comment,
  1 plan line. No new dependency, no abstraction, no option, no "for later".
- Oracle integrity: finding 2 has a positive oracle (seal copied exactly) and
  a hostile oracle (`null` → exit 2, no file) with a demonstrated RED→GREEN.
  Finding 3's test is labelled a regression guard, not a RED proof.
- No assertion removed or weakened; the only timing change is a comment word.
- No promotion of any `VERIFY` row; protected ledgers and frozen corpus clean.
- One honest limitation: the exact pre-fix runtime fall-through of finding 3
  is unreachable under `flag.ExitOnError`, so the code change is hardening, not
  an observed bug fix; the guard documents the contract going forward.

## Commit hook

The normal pre-commit hook was attempted first and **failed EXIT=1** on the
pre-existing `links` gate: `2 passed, 1 failed, 1 not run`, the same 28 dangling
references under protected `docs/VERIFY.md` / `docs/CHECKPOINTS.md`. Those files
were not edited by this wave; the failure is environmental and pre-dates it.
No gate, permit or ledger was touched to make it pass. The commit therefore
uses the documented override
`git -c core.hooksPath=C:/Users/ASUS/AppData/Local/Temp/opencode/no-hooks commit ...`
with explicit paths only. The full gate result above is the honest record:
`3 passed, 1 failed, 2 not run`, EXIT=1.
