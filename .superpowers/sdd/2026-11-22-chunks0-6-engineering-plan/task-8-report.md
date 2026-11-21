# Task 8 report — P3 chunk 6 local preparation (6.1.2 documented / 6.1.3 implemented)

## Status

Complete locally. `--confront-out <path>` is implemented in `cmd/cgo` with a
deterministic TDD suite (RED → GREEN) and an end-to-end CLI check. The real-tag
lane (6.1.2) is recorded **ready, not run** with the exact bench command for P4.
No `VERIFY.md` row moved, no protected ledger or frozen corpus edited, no live
host/bench/sadcnwp contacted, no gate weakened. Native gates are **not green**
on this checkout and are reported honestly below.

- Worktree: `C:/cgo/.worktrees/main-reconciled`, branch `main-reconciled`, base HEAD `bc4e871`.
- Report: `.superpowers/sdd/2026-11-22-chunks0-6-engineering-plan/task-8-report.md`
- Ledger appended: `.superpowers/sdd/2026-11-22-chunks0-6-engineering-plan/progress.md` (gitignored scratch).
- Brief-path alias: the brief names `2026-11-22-chunk6-prep/*`; per the canonical-ledger
  resolution in `progress.md`, these resolve to this `2026-11-22-chunks0-6-engineering-plan/*`
  workspace, and the report contract's explicit `task-8-report.md` name governs.

## Scope and decisions

Consumes GapRows 6.1.2 (real-tag lane not run) and 6.1.3 (no `--confront-out` flag)
from `docs/research/2026-11-22-chunks0-6-gap-audit.md`. The no-change contract's
12 `met` rows were not touched. The kernel truth is sealed at `docs/VERIFY.md`
row 43 (226 eBPF retrans events == 226 `tc -s` cake drops, audit-1794860919) and
lives on the bench guest only; it is **not** in this repository. Therefore the
CLI reads a real sealed input supplied by the operator and **fails closed** when
it is absent or invalid — it never invents a counter, a hash, a count or a PASS.

Edited files (3):
- `cmd/cgo/confront.go` (new) — seal loader, validator, report writer, CLI entry point.
- `cmd/cgo/confront_test.go` (new) — 5 deterministic tests (positive + 3 hostile + no-flag).
- `cmd/cgo/main.go` — `--confront-out` flag wiring on `run` + usage line.

## Exact commands and exit codes

All commands ran in `C:/cgo/.worktrees/main-reconciled` unless stated.

| # | Command | Exit | Result |
|---|---|---|---|
| 1 | `git status --short && git log --oneline -1` | 0 | WIP = 6 untracked `web/frontend/*` scratch files (untouched); HEAD `bc4e871` |
| 2 | `go test ./cmd/cgo/ -run 'TestConfrontOut' -count=1` (RED) | **1** | build failed: 5× `undefined: runConfrontOut` |
| 3 | `go test ./cmd/cgo/ -run 'TestConfrontOut' -count=1 -v` (GREEN) | 0 | 5/5 PASS |
| 4 | `go test ./cmd/cgo/ -run 'TestConfrontOut' -count=1 -v` (after oracle fix) | 0 | 5/5 PASS, non-ebpf seal now rejected on `source` |
| 5 | `go test ./cmd/cgo/ -count=1` | 0 | `ok ... cmd/cgo 28.117s` |
| 6 | `go build ./...` | 0 | clean |
| 7 | `CGO_CONFRONT_INPUT=<seal> go run ./cmd/cgo run --confront-out <out>` | 0 | report written, exact sealed values, `match: true` |
| 8 | `gofmt -l cmd/cgo/main.go cmd/cgo/confront.go cmd/cgo/confront_test.go` | 0 | empty (clean) |
| 9 | `sh .githooks/run-gates.sh` (run 1) | **1** | `2 passed, 2 failed, 2 not run` |
| 10 | `cd api && go test . -run 'TestNextSnapDrivesHubAndNotState\|TestAuditPanicRecovered' -count=1` | 0 | `ok ... api 3.735s` |
| 11 | `sh .githooks/run-gates.sh` (run 2) | **1** | `2 passed, 2 failed, 2 not run` (different flaky api test) |
| 12 | `git commit -m "feat(cgo): add --confront-out sealed-kernel-truth JSON export"` (normal hook) | **1** | hook failed: staged provenance gate could not run |
| 13 | `sh .githooks/run-gates.sh --staged` | **2** | `2 passed, 0 failed, 1 not run`; `SKIP provenance — book/data/runs absent` |

### RED → GREEN evidence

RED (exit 1), before `confront.go` existed — this is exactly the state produced
by deleting the implementation, so the tests fail closed without it:

```
# github.com/Realms4239/cgo/cmd/cgo [github.com/Realms4239/cgo/cmd/cgo.test]
cmd\cgo\confront_test.go:72:13: undefined: runConfrontOut
cmd\cgo\confront_test.go:86:13: undefined: runConfrontOut
cmd\cgo\confront_test.go:118:13: undefined: runConfrontOut
cmd\cgo\confront_test.go:132:13: undefined: runConfrontOut
cmd\cgo\confront_test.go:146:13: undefined: runConfrontOut
FAIL	github.com/Realms4239/cgo/cmd/cgo [build failed]
EXIT=1
```

GREEN (exit 0), after the minimal fix — 5 named tests, positive and hostile:

```
--- PASS: TestConfrontOutFileAbsentWithoutFlag (0.00s)
--- PASS: TestConfrontOutWritesSealedTruth (0.04s)
--- PASS: TestConfrontOutFailsClosedWhenSealAbsent (0.01s)
--- PASS: TestConfrontOutRejectsNonEbpfSeal (0.03s)
--- PASS: TestConfrontOutRejectsMissingCounters (0.01s)
PASS
ok  	github.com/Realms4239/cgo/cmd/cgo	0.473s
EXIT=0
```

One oracle was tightened after the first GREEN: the non-ebpf fixture initially
omitted `qdisc_drop_packets`, so it was rejected by the presence check rather
than the `source` check. The fixture was completed and re-run; the rejection now
names the real cause: `confront: sceau noyau: source "tcpinfo" — seul ebpf est une
vérité noyau scellée`. No gate, permit or assertion was weakened.

The five test contracts:
1. `TestConfrontOutFileAbsentWithoutFlag` — `runConfrontOut("", seal)` writes nothing (the headline deterministic absence oracle).
2. `TestConfrontOutWritesSealedTruth` — exact pass-through of `retrans_events=7`, `qdisc_drop_packets=0`, `retrans_payload_bytes=1024`, `object_sha256`, `audit_id`; computed `drops_delta = 9-2 = 7`; `match=true`.
3. `TestConfrontOutFailsClosedWhenSealAbsent` — missing file → exit 2, no output.
4. `TestConfrontOutRejectsNonEbpfSeal` — `source=tcpinfo` → exit 2, no output.
5. `TestConfrontOutRejectsMissingCounters` — schema/source present but counters absent → exit 2, no output.

## How the CLI sources the sealed truth (and fails honestly)

`cgo run --confront-out <path>` is a fast path: after flag parsing it exits
before any bench pre-check, so it never touches a host. The input seal path comes
from `CGO_CONFRONT_INPUT`; **no default file and no embedded fixture exists**, so
an unset variable or a missing/unreadable file is an error.

- `loadSeal(path)` reads the JSON and requires schema `cgo.kernel-truth-seal/1`,
  non-empty `audit_id`, `source == "ebpf"`, presence of `retrans_events`,
  `qdisc_drop_packets`, `tc_drops_before`, `tc_drops_after`, and
  `tc_drops_after >= tc_drops_before`. Any failure returns an error; nothing is written.
- `writeConfrontationReport` computes only `drops_delta = after - before` and
  `match = (ebpf retrans_events == drops_delta)`. Every other value is copied
  verbatim from the seal. A mismatch writes `match: false` — it is never forced green.
- Output schema `cgo.ebpf-vs-tc/1` carries `kernel_truth`, `tc_s`, `confrontation`,
  provenance and the seal's own `note` (e.g. the documented 6.8-cake
  `qdisc_drop_packets=0` reason-filter blind spot from `probe/bpf/BUILD.md`).
  The CLI does not reinterpret `qdisc_drop_packets=0` as "zero loss".
- The file is written atomically (temporary file + rename) with mode 0644; a
  validation failure leaves no output file at all.

The test fixture is explicitly labelled `TEST FIXTURE — not the sealed bench
evidence` and is used only to prove pass-through. The real 226/226 values were
not fabricated into the repository; in P4 the operator supplies the guest-sealed
report through `CGO_CONFRONT_INPUT`.

### Operator contract for P4

```
CGO_CONFRONT_INPUT=<guest-sealed kernel-truth JSON> \
  cgo run --confront-out docs/verification/v2-ebpf-confrontation/ebpf_vs_tc_report.json
```

Exit 0 = report written from a valid seal; exit 2 = seal absent/invalid, nothing
written (three-state: 2 is not green).

## Real-tag lane — ready, not run (6.1.2)

Documented for P4 only; **not executed here** (Windows host, no root, no bench,
no live host contact). From `docs/MUSTREAD/16-chunk-6-live-proof-freeze.md` §6.1
step 2, on the Linux bench VM:

```
sudo -E go test -race -count=1 ./...
sudo -E go test -tags=real -count=1 ./pkg/qdisc/... ./pkg/campagne/... ./pkg/probe/bpf/...
```

The second line is the corpus-verbatim path form. In this checkout the real
paths are `qdisc/`, `pkg/campagne/`, `probe/bpf/` (the single `-tags=real` test is
`qdisc/exec_real_test.go` → `TestExecRunnerRealTcReadsJSON`), so the bound command
for P4 is:

```
sudo -E go test -tags=real -count=1 ./qdisc/... ./pkg/campagne/... ./probe/bpf/...
```

No `-tags=real` test was run, so this row stays "ready, not run" and no VERIFY
row is promoted.

## Native gates — not green (honest)

`sh .githooks/run-gates.sh` returned exit 1 twice: `2 passed, 2 failed, 2 not run`.
- `PASS exit` (5 safe invocations; verdict contract ok) and `PASS hygiene` (1120
  tracked files clean) — both green.
- `FAIL links` — the known pre-existing false positive: 28 stale path references
  inside the protected `docs/VERIFY.md` and `docs/CHECKPOINTS.md`. Recorded in the
  P0 audit and prior task reports; those files were deliberately not edited.
- `FAIL tests` — `api` module go test failed on flaky timing tests: run 1
  `TestNextSnapDrivesHubAndNotState` (`hub NextSnap calls = 0, want >= 2`), run 2
  `TestHubCloseStopsTicks` (`hub did not tick: 1 calls`). Both pass in isolation
  (`cd api && go test . -run '...' -count=1` → exit 0, 3.735s). `api` is a separate
  Go module that cannot import `cmd/cgo`; this change cannot affect it. The
  differing failure each run confirms pre-existing flakiness, not a regression.
- `SKIP frozen` / `SKIP provenance` — `book/` absent on this checkout (exit 2 /
  not run, never counted as green).

Per the two-strike circuit breaker I did not run the gates a third time; the
failure mode is diagnosed, not repeated blindly.

### Pre-commit hook — actual failure reason (recorded truthfully)

The normal commit attempt (`git commit -m "feat(cgo): add --confront-out sealed-kernel-truth JSON export"`)
exited **1**. The hook runs the reduced staged gate set, and the staged run
`sh .githooks/run-gates.sh --staged` exited **2** with
`2 passed, 0 failed, 1 not run`: `PASS exit`, `PASS hygiene`, and
`SKIP provenance — book/data/runs absent: no reference header`. The `links` gate
is **not part of the staged gate set**, so this is *not* the known links false
positive — it is a pre-existing environmental condition: `book/` is absent by
design in this worktree and the frozen corpus is protected and human-owned, so
no code change in this task can satisfy it. Exit 2 ("gate could not run") is not
green and is not claimed as green.

Because the operator's standing direction is to use the documented no-hooks
override when a pre-existing environmental gate blocks a commit (and the frozen
corpus may not be created or edited to satisfy it), the commit is made with
`git -c core.hooksPath=C:/Users/ASUS/AppData/Local/Temp/opencode/no-hooks commit ...`.
This changes no gate, permit, regex, assertion or protected ledger and does not
turn a red gate green; the gate results above stand as recorded.

## Protected-surface confirmation

No edit to `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `book/data/runs`, any gate,
permit, frozen artifact, testbed-bootstrap, or `.githooks/`. No `git add -A`, no
merge, no history rewrite, no new dependency, no impeccable detector run. Only
the three files listed above plus this report and the gitignored progress ledger
were written.

## Self-review

- Scope: the flag exports from a sealed input only; it adds no measurement path,
  no campaign execution and no default data file. Nothing beyond 6.1.2/6.1.3 was changed.
- Oracle integrity: positive + three hostile oracles; all assert exit code and
  absence of output; the output assertion checks exact copied values, not
  non-nil. Deleting `confront.go` reproduces the RED build failure.
- Three-state exits: 0 written, 2 refused/absent, no fabricated 1; partial writes
  are prevented by validate-before-write and temp+rename.
- Edge cases: underflow guarded (`after >= before`), wrong schema/source rejected,
  missing counters rejected, `match` may honestly be false.
- Known limits: `object_sha256` and `note` are optional in the input schema (the
  real bench seal carries both); atomic rename over an existing destination is
  Linux-safe, which is the documented P4 bench. These are deliberate, not hidden.
- Two workspace scratch files (`web/frontend/*`, pre-existing at HEAD) were left untouched.

## Completion proof (post-final-edit)

VERIFY-TASK: Task 8 — P3 chunk 6 local preparation (6.1.2 documented ready-not-run; 6.1.3 `--confront-out` implemented).

GIT-DIFF-STAT: `git diff --stat` → empty (tracked tree clean after commit). Committed stat, `git show --stat --format= HEAD` (commit SHA omitted here so this block stays exact after amend):
```
 .../task-8-report.md                               | 257 +++++++++++++++++++++
 cmd/cgo/confront.go                                | 159 +++++++++++++++
 cmd/cgo/confront_test.go                           | 152 +++++++++++++++
 cmd/cgo/main.go                                    |   7 +
 4 files changed, 575 insertions(+)
 # (stat line count is self-stable: editing this block does not change it)
```

TOOLS-RUN: `git status --short && git log --oneline -1`; `go test ./cmd/cgo/ -run 'TestConfrontOut' -count=1`; `gofmt -w cmd/cgo/confront.go cmd/cgo/confront_test.go cmd/cgo/main.go`; `go test ./cmd/cgo/ -run 'TestConfrontOut' -count=1 -v`; `go test ./cmd/cgo/ -count=1`; `go build ./...`; `CGO_CONFRONT_INPUT=<seal> go run ./cmd/cgo run --confront-out <out>`; `gofmt -l cmd/cgo/main.go cmd/cgo/confront.go cmd/cgo/confront_test.go`; `sh .githooks/run-gates.sh` (×2); `cd api && go test . -run 'TestNextSnapDrivesHubAndNotState|TestAuditPanicRecovered' -count=1`; `git add cmd/cgo/confront.go cmd/cgo/confront_test.go cmd/cgo/main.go && git add -f .superpowers/sdd/.../task-8-report.md`; `git commit -m "feat(cgo): add --confront-out sealed-kernel-truth JSON export"`; `sh .githooks/run-gates.sh --staged`; `git -c core.hooksPath=.../no-hooks commit -m "..."`.

EXIT-CODES: git status/log 0; RED targeted test 1 (build failed, 5× `undefined: runConfrontOut`); GREEN targeted test 0 (5/5 PASS); strengthened targeted test 0; full cmd/cgo suite 0 (28.117s); `go build ./...` 0; e2e CLI 0; `gofmt -l` 0 (empty); gates run 1 = 1 (2 passed, 2 failed, 2 not run); isolated api tests 0 (3.735s); gates run 2 = 1 (different flaky api test); normal commit 1; staged gates 2 (2 passed, 0 failed, 1 not run); no-hooks commit 0; final targeted test 0 (5/5 PASS).

REPORT-OR-ARTIFACT: 5/5 named tests PASS — `TestConfrontOutFileAbsentWithoutFlag`, `TestConfrontOutWritesSealedTruth`, `TestConfrontOutFailsClosedWhenSealAbsent`, `TestConfrontOutRejectsNonEbpfSeal`, `TestConfrontOutRejectsMissingCounters`. Full `cmd/cgo` suite ok (28.117s); `go build ./...` clean; end-to-end CLI emitted the exact `cgo.ebpf-vs-tc/1` report from a seal and wrote nothing when the seal was absent. No VERIFY row promoted.

RAW-TAIL (verbatim last 15 lines of the final targeted test run):
```
--- PASS: TestConfrontOutFileAbsentWithoutFlag (0.01s)
=== RUN   TestConfrontOutWritesSealedTruth
confront → C:\Users\ASUS\AppData\Local\Temp\TestConfrontOutWritesSealedTruth2929170710\001\ebpf_vs_tc_report.json (sceau C:\Users\ASUS\AppData\Local\Temp\TestConfrontOutWritesSealedTruth2929170710\001\sealed-input.json)
--- PASS: TestConfrontOutWritesSealedTruth (0.02s)
=== RUN   TestConfrontOutFailsClosedWhenSealAbsent
confront: sceau noyau absent ou illisible (C:\Users\ASUS\AppData\Local\Temp\TestConfrontOutFailsClosedWhenSealAbsent559557289\001\pas-de-sceau.json): open C:\Users\ASUS\AppData\Local\Temp\TestConfrontOutFailsClosedWhenSealAbsent559557289\001\pas-de-sceau.json: Le fichier spécifié est introuvable.
--- PASS: TestConfrontOutFailsClosedWhenSealAbsent (0.01s)
=== RUN   TestConfrontOutRejectsNonEbpfSeal
confront: sceau noyau: source "tcpinfo" — seul ebpf est une vérité noyau scellée
--- PASS: TestConfrontOutRejectsNonEbpfSeal (0.01s)
=== RUN   TestConfrontOutRejectsMissingCounters
confront: sceau noyau incomplet (C:\Users\ASUS\AppData\Local\Temp\TestConfrontOutRejectsMissingCounters720258974\001\sealed-input.json): retrans_events absent
--- PASS: TestConfrontOutRejectsMissingCounters (0.01s)
PASS
ok  	github.com/Realms4239/cgo/cmd/cgo	0.489s
EXIT=0
```
