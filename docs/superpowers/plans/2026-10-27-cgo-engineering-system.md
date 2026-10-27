# CGO Engineering System — Implementation Plan

*Date: 2026-10-27 · Status: pending user review · Branch:
`rebuild/plan1-economics-audit` · HEAD at writing: `05715f3` · Spec:
`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md`*

---

## 0. Scope guard

This plan builds the process layer only. It MUST NOT:

- modify `book/data/runs`, `book/data/stats.json`, `book/content/`, or any
  frozen run;
- change `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, or a Go CSV
  header;
- touch eBPF, IFB, BBRv3, `.dist`, the UI, or the thesis chapters;
- run a campaign, the live bench, Graylog, or the remote PC
  `100.114.101.13`;
- commit `epbfBBR.md`, `backup/`, `data/`, `book/`, `testbed-bootstrap/`,
  `unp*/`, `*.docx`, `*.pdf`, `cgo.exe`, or any user artifact;
- use `git add -A`;
- merge to `main`.

Every number quoted in the new documents is labelled `MEASURED`, `RECORDED`,
`PENDING` or `BLOCKED`. No live claim becomes `VERIFIED` in this plan.

The six gates, the three ledgers, `AGENTS.md`, the two hooks and the Makefile
targets are the whole deliverable. If an implementation step wants a seventh
gate, a registry file, or a generator, stop and apply the admission test in the
spec §4.12.

---

## 1. File map

| Path | Kind | Phase |
|---|---|---|
| `AGENTS.md` | new, root entry | A |
| `docs/VERIFY.md` | new, claim ledger | A |
| `docs/CHECKPOINTS.md` | new, session ledger | A |
| `docs/ERRORS.md` | new, defect ledger | A |
| `docs/superpowers/plans/README.md` | edit, slice index | A |
| `tools/verify.py` | new, gate runner | B |
| `tools/verify_test.py` | new, unit tests | B |
| `Makefile` | edit, three targets | B |
| `.githooks/pre-commit` | new | C |
| `.githooks/pre-push` | new | C |
| `README.md` | edit, one pointer | D |
| `docs/architecture.md` | edit, one pointer | D |

No other file is in scope. A file that appears in the diff without a line in
this table stops the task.

---

## 2. Task 0 — Preflight (read-only)

**Goal:** record the starting state and the first dated audit output, without
changing anything.

1. Record the branch and HEAD:
   ```bash
   git log --oneline -1
   git status --short
   ```
   Expected HEAD `05715f3` or a later commit on the same branch. If the branch
   differs, stop and report before acting.

2. Confirm the prerequisites:
   ```bash
   test -d book/data/runs && echo book-ok
   test -f book/tools/audit-archive.py && echo audit-ok
   test -d web/frontend/dist && echo dist-ok || echo dist-missing
   python --version
   go version
   ```

3. Run the authoritative frozen-corpus audit and capture the exact output:
   ```bash
   python book/tools/audit-archive.py
   ```
   Expected: exit 0 and two lines beginning `OK —`. Quote both lines verbatim
   in the Phase A checkpoint and in the first `docs/VERIFY.md` row. This is the
   first date on which the frozen counts are verified in this system. If the
   audit fails, stop: the frozen corpus is the one thing this plan must not
   touch, and a failing audit is a finding, not a gate to bypass.

4. Record the CGO code size for the spec's inventory:
   ```bash
   find . -name '*.go' -not -path './book/*' -not -path './.worktrees/*' | wc -l
   find . -name '*_test.go' -not -path './book/*' -not -path './.worktrees/*' | wc -l
   ```

**Acceptance:** HEAD, branch, audit output, Python version and Go version are
recorded in the first checkpoint. No file changed.

**Commit:** none.

---

## 3. Phase A — Entry and ledgers (docs only)

### Task A1 — `AGENTS.md`

**Files:** `AGENTS.md` (new, at most 120 lines).

**Content, in this order:**

1. **What this repository is.** One paragraph: CGO/Meteolink, the reproducible
   constrained-link bench, the frozen corpus, the thesis. No marketing.
2. **Read first.** `AGENTS.md` → latest `docs/CHECKPOINTS.md` entry →
   `docs/VERIFY.md` → the current row of
   `docs/superpowers/plans/README.md` → the artifact the entry names.
3. **The one command.** `make verify`, `make verify-fast`,
   `make verify-selftest`. State: 0 passed, 1 failed, 2 could not run; 2 is not
   green; a skipped gate is named.
4. **Never touch.** `book/data/runs`; the three CSV schemas; `.dist`
   execution; Graylog; the remote PC `100.114.101.13`; the live bench without
   approval. State that no runtime `clang`, `gcc`, `llvm` or `linux-headers`
   goes on the VM.
5. **Never commit.** The list from §0 of this plan, plus "never `git add -A`".
6. **Workflow.** Scope line → spec/plan for a new slice → build → verify →
   checkpoint → commit per logical group with explicit paths.
7. **Roles.** Builder, verifier, narrator, human; one role per pass; the
   verifier does not read the builder's summary; a number leaving the
   repository gets a separate pass.
8. **Language.** French comments and user-facing text; English identifiers,
   specs, plans and gate messages. A file keeps its existing language.
9. **Stop and ask.** Destructive actions, live bench, merge to `main`, secret
   exposure, frozen-data change, new dependency, silent scope growth.
10. **Pointers.** The six gates, the three ledgers, the plan index, the
    umbrella spec, `PRODUCT.md`, `DESIGN.md`.

Do not restate the frozen counts in `AGENTS.md`. Point at `docs/VERIFY.md`.
Do not restate the whole spec. `AGENTS.md` is the operating card.

**Acceptance:** every repository path named in `AGENTS.md` exists; the file is
under 120 lines; it contains no unlabelled number and no promise of an
outcome.

### Task A2 — `docs/VERIFY.md`

**Files:** `docs/VERIFY.md` (new).

**Content:**

1. One paragraph: what the ledger is, and the rule that a number in a document
   cites a row or is labelled `RECORDED` or hypothesis.
2. The status definitions from the spec §4.6, copied exactly.
3. The table. Seed rows, in this order:
   - Frozen corpus — `RECORDED` (umbrella spec 2026-10-25) until Task 0's audit
     output replaces it with `VERIFIED`, date and host.
   - Frozen corpus unchanged since the seal — `PENDING` until the `frozen` gate
     exists.
   - Windows engine (`go test ./...`, `go vet ./...`, cross-build) —
     `RECORDED` from the plan index; re-run in Phase B.
   - Linux live acceptance (G0–G7, campaign) — `BLOCKED`, needs the stable
     Linux 6.8 bench.
   - eBPF retransmit/RTT/qdisc drops — `BLOCKED`, no `bpf_probe.o` and no
     Linux 6.8.
   - IFB ingress shaping — `BLOCKED`, live acceptance; Windows only generates
     commands.
   - BBRv3 3×3 matrix — `BLOCKED`, kernel provisioning and live acceptance.
   - `.dist` netem — `BLOCKED`, no field trace at or above the 30-sample floor.
   - Tableau 15 CUBIC cells — `RECORDED` limitation: P2/up/BBRv1/1-flow legacy
     bench; CUBIC cells are invalid on G4 and MUST NOT be cited as a valid
     comparison.
   - Thesis book numbers — `RECORDED`, pending the Phase A audit.
   - Second copy of `book/` and the frozen corpus — `PENDING`; a dated off-tree
     copy and a restore drill are required before submission.
4. A "How a row moves" section: only with a command, a date and a host; a
   `RECORDED` row is not promoted by confidence; a `REFUTED` row stays.

**Acceptance:** no row lacks a date, an evidence path or a named blocker; the
first row quotes the Task 0 audit output; no row says "verified" without a
command that ran.

### Task A3 — `docs/CHECKPOINTS.md`

**Files:** `docs/CHECKPOINTS.md` (new).

**Content:**

1. One paragraph: append-only; a session without an entry loses its detail to
   compaction.
2. The resume protocol:
   - Announce the entry id and `git log --oneline -1`. If HEAD differs from the
     entry, say so before acting.
   - Read the latest entry, then `docs/VERIFY.md`, then the plan-index row,
     then the artifact the entry names.
3. The entry format from the spec §4.8. Time is UTC from a command, never the
   taskbar.
4. The first entry: Task 0's preflight, with branch, HEAD, audit output,
   Python/Go versions, and "no file changed".

**Acceptance:** the first entry is present and dated; the protocol names the
exact files to read; a reader can resume from it without the chat.

### Task A4 — `docs/ERRORS.md`

**Files:** `docs/ERRORS.md` (new).

**Content:**

1. One paragraph: append-only; an error without a cause is a rumor; a fix
   without a proof is a wish.
2. The entry format from the spec §4.7.
3. Two seed entries, both labelled `recorded` because they were found in
   earlier sessions:
   - **E1 — the dualns target overwrite.** `runcli.go` overwrites an explicit
     dualns `--target` with the CLI default `1.1.1.1`; a 3×3 campaign must run
     inside `ns_client` with `--target 10.30.0.2`. Symptom, cause, the current
     workaround, and the lesson: a default that replaces an explicit flag is a
     false configuration.
   - **E2 — the Tableau 15 regime mix.** The table is P2, `direction=up`,
     BBRv1, 1-flow legacy bench; its CUBIC cells are invalid on G4. Symptom,
     cause, the labelling rule, and the lesson: a table that mixes bench
     regimes is not a comparison.
4. Do not invent a proof for either entry. "Proof: recorded in the umbrella
   spec and the harmonization matrix; no live re-run in this session."

**Acceptance:** each entry has symptom, cause, fix/workaround, proof and
lesson; no fabricated date or measurement.

### Task A5 — Plan index

**Files:** `docs/superpowers/plans/README.md` (edit).

**Steps:**

1. Fix slice 7 (BBRv3): link the existing spec and plan, set status
   **Proposed**, and state that it is gated on kernel provisioning and live
   acceptance. The files already exist:
   - `docs/superpowers/specs/2026-10-27-cgo-bbrv3-matrix-design.md`
   - `docs/superpowers/plans/2026-10-27-cgo-bbrv3-matrix.md`
2. Add slice 9: **Engineering system (AIR × DIS synthesis)**, spec and plan
   paths, status **Proposed**, depends on slices 1–3 (done), next action
   "user approval; Phase A".
3. Add slice 9 to the dependency graph as independent of 4–8, with a note:
   run it before any live slice so the frozen guard exists first.
4. Replace the duplicated frozen counts in "Standing constraints" with a
   pointer to `docs/VERIFY.md`. Keep the operational guards (no `.dist`, no
   Graylog, no remote PC, no `git add -A`, no main merge, no fabricated
   measurement) because they are slice guards, not counts.
5. Add the plan-hygiene rule: a new governing document is added to
   `AGENTS.md` in the same commit; a new slice gets a spec and plan first.

**Acceptance:** the index points at existing files; slice 7 and 9 statuses are
accurate; no count is duplicated from `VERIFY`; the dependency graph renders.

### Phase A commit

```bash
git add AGENTS.md docs/VERIFY.md docs/CHECKPOINTS.md docs/ERRORS.md \
  docs/superpowers/plans/README.md
git commit -m "docs: add agent entry and evidence ledgers"
```

**Phase A acceptance:** all named paths exist; `python book/tools/audit-archive.py`
output is quoted in `VERIFY` and the checkpoint; every number is labelled; no
code, hook or product file changed.

---

## 4. Phase B — Gate runner and six gates

### Task B1 — `tools/verify.py` skeleton

**Files:** `tools/verify.py` (new).

**Design:**

- Python 3.11+, standard library only. No YAML, no third-party package.
- Root: `Path(__file__).resolve().parents[1]`.
- CLI:
  - no arguments → all six gates;
  - `--fast` → all except `tests`;
  - `--staged` → select by staged paths, always `hygiene`;
  - `--only NAME` (repeatable) → named gates only;
  - `--list` → name, what it proves, fast/full membership, proof status;
  - `--selftest` → run every gate's fault proof plus its control.
- Result model:
  ```python
  PASS, FAIL, SKIP = "pass", "fail", "skip"
  @dataclass
  class Result:
      name: str
      status: str
      detail: str
  ```
- `verdict(results)`:
  - any `FAIL` → 1;
  - else any `SKIP` → 2;
  - else 0.
- Registry: one list of `Gate` records:
  ```python
  Gate(name, proves, paths, run, prove, fast)
  ```
  `paths` is the prefix list used by `--staged`; `run` returns a `Result`;
  `prove` returns `(caught: bool, detail: str)` and MUST exist.
- `--selftest` refuses a gate with no `prove` (exit 2) and prints one line per
  gate: `caught` or `SURVIVED`.
- Startup line-ceiling check: count the lines of `tools/verify.py`,
  `tools/verify_test.py`, `.githooks/pre-commit`, `.githooks/pre-push`; if the
  total exceeds `MAX_GATE_LINES`, print the count and exit 2. The constant is
  set in Task B8.
- Every message is English; comments are French, per the repository rule.
- No `shell=True`. Subprocess calls use argument lists. No network.

**Acceptance:** `python tools/verify.py --list` runs; `--only nope` exits 2
with a named error; a gate with no proof is refused; the file is under 250
lines at this stage.

### Task B2 — `frozen` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. If `book/tools/audit-archive.py` or `book/data/runs` is absent, return
   `SKIP` with `book/ absent: frozen corpus not checked`.
2. Run `sys.executable book/tools/audit-archive.py` with `cwd=root`, capture
   stdout/stderr, timeout 120 s.
3. Non-zero exit → `FAIL`, with the last 20 lines of output.
4. Zero exit but a missing `OK —` line → `FAIL` (the script did not attest).
5. Zero exit and both `OK —` lines → `PASS`, with the two lines in `detail`.
6. Do not parse or re-implement the counts. The audit is the only source.

**Fault proof (`prove`):**

1. If `book/` is absent, return `False, "book/ absent: the frozen proof cannot run"`.
2. Copy `book/` to a temp directory, excluding `node_modules`, `.git` and
   `dist` if present.
3. Find the first run directory with a `manifest.json` whose `files` list
   names an existing file. Mutate one byte in that file (append a newline).
4. Run the copied `book/tools/audit-archive.py` with `cwd=temp`, expect a
   non-zero exit. If it exits 0, the proof failed: the audit cannot see a
   changed corpus file, and that is a finding about the audit, not about the
   gate.
5. Control: run the copied audit on the unmutated copy first and require exit
   0. If the control fails, the copy is incomplete; say so.
6. Delete the temp tree in `finally`.

**Acceptance:** on the real tree, `frozen` is `PASS`; `--selftest` reports
`caught` for `frozen`; mutating a copied run file is observed.

### Task B3 — `provenance` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. Reference header: read the `aqm_eval.csv` header from the first frozen run
   under `book/data/runs` (sorted). If the corpus is absent, return `SKIP` with
   the reason. Do not hardcode the 23 columns; the frozen corpus is the
   reference.
2. For every `data/runs/*/aqm_eval.csv`, the header must equal the reference
   exactly. Count the files checked. A mismatch is `FAIL` with the run name
   and the first differing column.
3. For every `valuation.json` under `data/runs/` and `book/data/runs/`:
   - `basis` must be exactly `measured` or `estimated`;
   - `source`, if present, must be exactly `ebpf`, `tcpinfo` or `""`;
   - `source` must not be non-empty when `basis` is `estimated`.
   Count the sidecars checked.
4. If no `valuation.json` exists, the detail says
   `0 valuation sidecars checked; writer covered by tests` and the gate still
   `PASS` on the headers. This is a named limit, not a hidden pass.
5. Do not inspect or rewrite the CSV rows. The gate seals headers and
   provenance enums only.

**Fault proof:**

1. Build a temp `book/data/runs/ref/aqm_eval.csv` with the reference header
   and a temp `data/runs/x/aqm_eval.csv` with an extra column; the header check
   must fail.
2. Build a temp `valuation.json` with `basis=guess`; must fail.
3. Build `basis=estimated, source=ebpf`; must fail.
4. Control: `basis=measured, source=ebpf`, `basis=estimated, source=""`, and an
   exact header must pass.
5. Run the validator functions directly against the temp paths; do not copy the
   real corpus.

**Acceptance:** the real tree reports the number of headers checked and the
number of sidecars checked; `--selftest` reports `caught`; a header mismatch is
named.

### Task B4 — `tests` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. If `web/frontend/dist` is absent, return `SKIP` with
   `web/frontend/dist absent: build the frontend first`.
2. Run `go vet ./...` with a 300 s timeout; non-zero → `FAIL`.
3. Run `go test ./...` with a 600 s timeout; non-zero → `FAIL`.
4. Both pass → `PASS`, with the elapsed time.
5. Use `shutil.which("go")`; if absent, `SKIP`.
6. Do not run `-race`, do not run the live E2E, do not build the frontend.

**Fault proof:**

1. Create a temp module with `go.mod` (`module fixture`, `go 1.25`), one
   passing file and one failing test; run `go test ./...` in it and require a
   non-zero exit.
2. Create a second temp module with a vet error (`fmt.Printf("%d", "x")`) and
   require `go vet ./...` to fail.
3. Control: a module with one passing test must pass both commands.
4. If `go` is absent, return `False` with the reason.

**Acceptance:** the real gate is green on Windows with `dist/` present;
`--selftest` catches the failing module; a missing `dist/` reports 2, not 0.

### Task B5 — `exit` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. Inspect `cmd/cgo/main.go` and choose five invocations that cannot touch the
   network, the bench, or the frozen corpus. Record them in a
   `SAFE_INVOCATIONS` table with a one-line reason each. Candidates:
   - `[]` (usage);
   - `["--nope"]` (unknown flag);
   - `["version"]`;
   - `["report", "--run", "no-such-run"]` if the command validates offline;
   - `["figures", "--help"]` if it exits before reading data.
   Reject any candidate that writes to `data/`, starts a server, or runs `tc`.
2. Build the CLI once to a temp path:
   `go build -o <temp>/cgo[.exe] ./cmd/cgo`. A build failure is `FAIL`.
3. Run each invocation with a 30 s timeout and an environment without
   credentials. Accept only exit codes 0, 1, 2. Exit 3, a signal, or a
   traceback/panic in stderr is `FAIL`.
4. A command that exits 0 while printing a refusal (for example "no such
   file") is `FAIL`, following DISanalyzer's `exit_code_sweep.py` rule.
5. Do not run `serve`, `run`, `audit`, `kit`, or anything that opens a socket.
6. Also assert `verdict()`: `PASS` → 0, `PASS+SKIP` → 2, `FAIL+SKIP` → 1.
   This is the gate's own skip contract.

**Fault proof:**

1. A fixture script that exits 3 must be rejected.
2. A fixture script that raises an uncaught exception must be rejected.
3. A fixture script that exits 0 while printing "no such file" must be
   rejected.
4. A fixture script that exits 2 must be accepted.
5. A skipped result must yield 2, not 0.

**Acceptance:** the safe invocations are listed with reasons; the real CLI
never returns an unlisted code; `--selftest` catches each fixture.

### Task B6 — `hygiene` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. File list: `git ls-files -z` from the root. If git fails, `SKIP`.
2. For each tracked text file up to 2 MB, scan for:
   - private key blocks: `-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`;
   - token shapes: `ghp_[A-Za-z0-9]{30,}`, the GitHub PAT prefix, `sk-[A-Za-z0-9]{20,}`,
     `AKIA[0-9A-Z]{16}`, `xox[baprs]-`;
   - tailnet addresses: `\b100\.(6[4-9]|[7-9][0-9]|1[01][0-9]|12[0-7])\.\d{1,3}\.\d{1,3}\b`.
   Do not ban RFC1918 addresses; `10.30.0.2` is legitimate.
3. Forbidden tracked paths:
   - `epbfBBR.md`, `session-*.md`, `handoff-*.md`, `critical-review-*.md`;
   - `backup/`, `data/`, `book/`, `testbed-bootstrap/`, `unp*/`,
     `node_modules/`, `.worktrees/`;
   - `*.docx`, `*.pdf`, `*.tar.gz`, `*.zip`, `*.exe`.
4. Oversized files: measure the largest tracked file first. Set the ceiling
   above it with a named allowlist (`docs/img/` is allowed). Do not hardcode a
   number in the plan; the implementation records the measured ceiling in a
   comment.
5. A finding names the path and the rule, never the matched secret value.
6. Allow placeholder forms: `KEY=<your-key>`, an empty assignment, and a prose
   mention. The scan skips comments only where the format makes a comment
   unambiguous (`.md` prose is scanned; a false positive is fixed, not
   bypassed).

**Fault proof:**

1. A temp tree with a fake private key and a fake `ghp_` token must fail.
2. A temp tree with a `100.72.251.32` address must fail.
3. A temp tree with a tracked `note.docx` must fail.
4. Control: `KEY=<your-key>`, `10.30.0.2`, and `password = ""` must pass.
5. Build a temp git index or a pure file-list function; do not stage anything
   in the real repository.

**Acceptance:** the real tree is clean; the finding count is zero or the
finding names a real file; `--selftest` catches each fixture and accepts the
control.

### Task B7 — `links` gate

**Files:** `tools/verify.py` (edit).

**Behavior:**

1. Files scanned: `AGENTS.md`, `docs/VERIFY.md`, `docs/CHECKPOINTS.md`,
   `docs/ERRORS.md`, `docs/superpowers/plans/README.md`.
2. Extract backticked tokens. A token is a path candidate when it contains
   `/`, has a file extension or ends with `/`, and does not start with
   `http`, `git`, `go`, `make`, `python`, `-`, and contains no space, `*`,
   `<`, `>` or `{`.
3. Resolve against the repo root. A missing path is `FAIL` with the source
   file and the token. A path inside a fenced code block is skipped only if it
   is clearly a command; prefer scanning and fixing false positives over
   skipping whole blocks.
4. Also parse the plan index's spec/plan columns: every backticked `*.md`
   filename must exist under `docs/superpowers/specs/` or
   `docs/superpowers/plans/`.
5. If a scanned file is absent, `SKIP` with its name.

**Fault proof:**

1. A fixture document naming `docs/gone.md` must fail.
2. A fixture document naming an existing file must pass.
3. A fixture plan index naming a missing spec must fail.
4. Control: a document with a command like `make verify` and a URL must pass.

**Acceptance:** the real tree has zero missing paths; a renamed file is caught;
`--selftest` reports `caught`.

### Task B8 — `tools/verify_test.py`, Makefile, ceiling

**Files:** `tools/verify_test.py` (new), `Makefile` (edit),
`tools/verify.py` (edit the constant).

**Unit tests (`unittest`, no network, no bench):**

1. `verdict`: `[PASS]` → 0; `[PASS, SKIP]` → 2; `[FAIL, SKIP]` → 1.
2. `--staged` selection:
   - `book/data/runs/x` → `frozen`, `hygiene`;
   - `pkg/api/x.go` → `provenance`, `exit`, `hygiene`;
   - `docs/x.md` → `links`, `hygiene`;
   - `tools/verify.py` → all six;
   - an unknown path → `hygiene`.
3. `provenance` header mismatch and bad `basis` (reuse the gate's pure
   functions).
4. `hygiene` fake key accepted/failed.
5. `links` missing path.
6. `exit` classifier: 3 and panic fail; 0/1/2 pass; 0 with a refusal fails.
7. A gate with no `prove` is refused.

**Makefile:**

```make
verify:
	python tools/verify.py

verify-fast:
	python tools/verify.py --fast

verify-selftest:
	python tools/verify.py --selftest

hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/pre-push 2>/dev/null || true
	python tools/verify.py --list
```

Add these to `.PHONY`. Do not change `test`, `vet`, `build`, `frontend` or
`figures`.

**Ceiling:** measure the four files. Target ≤ 600 lines total. If the
implementation is ≤ 600, set `MAX_GATE_LINES` to the measured count plus 10
percent, rounded up to 25. If it is > 600, delete features first; do not set a
ceiling above 600 in this slice. Record the measured count in the Phase B
checkpoint.

**Acceptance:** `python tools/verify_test.py` passes; `--list` shows six gates;
the ceiling check fails when a temporary 999-line copy is measured (a unit
test can call the pure `line_ceiling_problem` function).

### Task B9 — Full verification and commit

**Files:** none new.

**Steps:**

1. Run:
   ```bash
   python tools/verify.py --list
   python tools/verify.py --selftest
   python tools/verify.py
   ```
   Expected: six gates listed; six `caught` proofs plus six controls;
   `6 passed, 0 failed, 0 not run` with `book/` and `dist/` present, or the
   named `SKIP` and exit 2.
2. Run the existing Windows gates:
   ```bash
   go test ./...
   go vet ./...
   GOOS=linux GOARCH=amd64 go build ./...
   python book/tools/audit-archive.py
   ```
   Record the output. Do not claim a live result.
3. Deliberate mutation check, in a disposable copy or with `--only`:
   - append a newline to a copied frozen CSV → `frozen` fails;
   - add `basis=guess` to a fixture `valuation.json` → `provenance` fails;
   - break a fixture test → `tests` fails;
   - add a fake token to a fixture tracked file → `hygiene` fails;
   - rename a doc named in `AGENTS.md` in a copy → `links` fails;
   - a fixture CLI exit 3 → `exit` fails.
   Do not mutate the real tree.
4. Record the measured gate-line count and the `MAX_GATE_LINES` value in the
   checkpoint.

**Commit:**

```bash
git add tools/verify.py tools/verify_test.py Makefile
git commit -m "tools: add six-gate verify runner with fault proofs"
```

**Phase B acceptance:** all six gates have a proof and a control; `--selftest`
exits 0; `make verify` is green on Windows or exits 2 with the named missing
prerequisite; no gate runs a live command; the ceiling is measured and honest.

---

## 5. Phase C — Hooks and branch guard

### Task C1 — `.githooks/pre-commit`

**Files:** `.githooks/pre-commit` (new, executable).

**Content:**

```sh
#!/bin/sh
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"
if ! command -v python >/dev/null 2>&1; then
  echo "pre-commit: python not found; commit refused (2)" >&2
  exit 2
fi
python tools/verify.py --staged
```

- The hook runs the fast set, not the Go suite.
- A `SKIP` returns 2, so the commit is blocked. This is fail-closed: a commit
  that cannot verify its selected gates is not allowed.
- The message names the gate that could not run; the user may bypass with
  `--no-verify` as a recorded decision.

**Acceptance:** the hook runs from any subdirectory; a missing Python blocks
with 2; a normal staged change passes.

### Task C2 — `.githooks/pre-push`

**Files:** `.githooks/pre-push` (new, executable).

**Content:**

```sh
#!/bin/sh
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"

main=0
while read -r local_ref local_sha remote_ref remote_sha; do
  [ "$remote_ref" = "refs/heads/main" ] && main=1
done

if [ "$main" = 1 ] && [ "${CGO_ALLOW_MAIN:-}" != "1" ]; then
  echo "pre-push: push to main refused; merge is a user decision (set CGO_ALLOW_MAIN=1 to override)" >&2
  exit 1
fi

python tools/verify.py
```

- Git always supplies the refspecs on stdin. A manual run with no stdin does
  not trigger the main guard; it still runs the full verification.
- The full set includes `tests`; a push is slower than a commit, which is the
  point.
- `CGO_ALLOW_MAIN=1` is an explicit override and must be printed in the
  checkpoint when used.

**Acceptance:** a refspec naming `refs/heads/main` exits 1; the same refspec
with `CGO_ALLOW_MAIN=1` exits 0 (or the verification's code); a feature branch
runs the full set.

### Task C3 — `make hooks` and proof

**Files:** `Makefile` (already edited in B8).

**Steps:**

1. Run `make hooks`. It sets `core.hooksPath`, makes the scripts executable,
   and prints `--list`.
2. Prove in a disposable copy, under the approved temp directory:
   ```bash
   git clone --no-hardlinks . "$TMP/cgo-hooks"
   cd "$TMP/cgo-hooks"
   printf 'FAKE_TOKEN=ghp_%s\n' "$(python -c 'print("a"*36)')" >> note.txt
   git add note.txt
   sh .githooks/pre-commit   # expect non-zero
   git reset
   printf 'refs/heads/x %s refs/heads/main %s\n' "$(git rev-parse HEAD)" \
     "$(git rev-parse HEAD)" | sh .githooks/pre-push   # expect 1
   printf 'refs/heads/x %s refs/heads/main %s\n' "$(git rev-parse HEAD)" \
     "$(git rev-parse HEAD)" | CGO_ALLOW_MAIN=1 sh .githooks/pre-push
   ```
   A clone of a repository with an ignored `book/` will not have `book/`; the
   full verification will exit 2. That is the correct fail-closed result for a
   clone without the corpus. Record it; do not claim green.
3. Do not stage the fake token in the real repository. Delete the temp clone.

**Acceptance:** the hook blocks the fake token; the main push is refused; the
override is honoured; the temp clone is removed; the real index is untouched.

### Phase C commit

```bash
git add .githooks/pre-commit .githooks/pre-push Makefile
git commit -m "hooks: enforce fast and full gates at commit and push"
```

If the Makefile was already committed in Phase B, commit only the two hook
files here and say so in the checkpoint.

**Phase C acceptance:** `make hooks` is idempotent; the pre-commit blocks a
secret and a book change; the pre-push refuses `main`; a normal feature push
runs the full set; the proof is recorded.

---

## 6. Phase D — Cadence and retirement

### Task D1 — Pointers

**Files:** `README.md`, `docs/architecture.md`.

**Steps:**

1. In `README.md`, add one short section near the top or the development
   section: "Travailler sur ce dépôt" → `AGENTS.md`, `make verify`,
   `docs/VERIFY.md`. One paragraph, no duplicate rules.
2. In `docs/architecture.md`, add one line in the introduction: the operating
   rules and current claim status live in `AGENTS.md` and `docs/VERIFY.md`.
3. Do not restate the gates or the frozen counts in either file.

**Acceptance:** both pointers resolve; no duplicated rule or count.

### Task D2 — Cadence

**Files:** `AGENTS.md` (edit), `docs/CHECKPOINTS.md` (edit).

**Steps:**

1. Add the WIP rule: three active slices; blocked slices do not count; a fourth
   active slice is refused until one closes or is deferred.
2. Add the Friday routine: full `make verify`, update the `VERIFY` rows touched
   that week, checkpoint, name any blocker older than 48 hours with its owner
   and unblock condition.
3. Add the closure rule: a slice closes only when its acceptance runs and its
   rows move to `VERIFIED` or `BLOCKED` with a named condition. "Windows-green"
   is not "done".
4. Do not create `docs/WEEKLY.md`. The plan index is the backlog; the
   checkpoint is the history.

**Acceptance:** the rules are in one place; no new weekly file exists.

### Task D3 — First weekly checkpoint

**Files:** `docs/CHECKPOINTS.md` (edit).

**Steps:**

1. Run the full verification.
2. Append a checkpoint with the date, branch, HEAD, gates run, results,
   `VERIFY` rows changed, blockers older than 48 hours, and the next slice.
3. If a gate was skipped, say which and why; do not call the week green.

**Acceptance:** the checkpoint is dated, the results are quoted, and the next
session can resume from it.

### Task D4 — Stale artifacts and second copy

**Files:** none.

**Steps:**

1. List the untracked handoff/session/user artifacts (`git status --short`).
   Do not delete or move them; record them as out of scope in the checkpoint.
2. Check the second copy: does a dated archive of `book/` and `book/data/runs`
   exist outside `C:\cgo`? Record the date and the path in `VERIFY` if yes;
   leave the row `PENDING` if no. Do not create the copy in this plan.
3. Record the largest tracked file and the hygiene ceiling in the checkpoint.

**Acceptance:** no user artifact was touched; the second-copy row is honest;
the ceiling is recorded.

### Phase D commit

```bash
git add README.md docs/architecture.md AGENTS.md docs/CHECKPOINTS.md \
  docs/VERIFY.md
git commit -m "docs: adopt cadence and point the front door at the system"
```

**Phase D acceptance:** the pointers resolve; the Friday checkpoint exists; the
plan index is current; the next session can resume from `AGENTS.md` alone.

---

## 7. Whole-slice acceptance

Run, in this order, and paste the output into the final checkpoint:

```bash
python tools/verify.py --list
python tools/verify.py --selftest
python tools/verify.py
go test ./...
go vet ./...
GOOS=linux GOARCH=amd64 go build ./...
python book/tools/audit-archive.py
git status --short
```

Pass conditions:

- `--list` shows six gates, each with a proof.
- `--selftest` shows six `caught` and six controls; exit 0.
- `verify` shows `6 passed, 0 failed, 0 not run` with `book/` and `dist/`
  present, or the named `SKIP` and exit 2.
- The Go suite, vet, cross-build and audit are green on Windows with `book/`
  present.
- Each of the six deliberate mutations fails the right gate.
- The hooks block a fake secret and a main push in a disposable clone.
- `book/data/runs`, the CSV schemas, `.dist`, eBPF, IFB, BBRv3 and the thesis
  are unchanged.
- `git status --short` contains only the intended files. No `git add -A`.
- No document claims a live result that did not run.

---

## 8. Self-review

- **Spec/plan agreement:** file names, six gate names, statuses, phases and
  acceptance match the spec. The spec's §4.4 table and this plan's tasks B2–B7
  name the same six gates.
- **Scope:** every file in the file map has a task; every task names its files;
  no task touches `book/data/runs`, the CSV schemas, or product code.
- **Proof discipline:** each gate has a `prove` and a control; `--selftest`
  refuses a gate without one; no gate is trusted because it is green.
- **Exit contract:** 0/1/2 everywhere; a skip is 2; the summary never says
  "all checks passed" with a skip.
- **Source of truth:** the frozen counts live in the audit and `VERIFY`; the
  plan index points at them; `AGENTS.md` does not restate them; no second copy.
- **Hooks:** fail-closed on a missing prerequisite; the main guard reads the
  refspec; a bypass is explicit and recorded.
- **Ceiling:** measured, not guessed; target ≤ 600 lines; a gate that pushes it
  over is cut, not blessed.
- **No fabrication:** the frozen counts are `RECORDED` until Task 0; the live
  slices stay `BLOCKED`; the seed error entries are labelled `recorded`.
- **No ceremony:** no weekly file, no doc register, no YAML registry, no
  generated docs, no CI, no full mutation harness. Each refusal has a reason in
  the spec.
- **Language:** French comments and user-facing text; English identifiers,
  specs, plans and gate messages.
- **Git:** explicit `git add` paths; per-phase commits; no `main` merge; user
  artifacts stay untracked.

---

## 9. Rollback

If Phase B exceeds the ceiling or a gate cannot be proven:

1. Stop. Do not raise `MAX_GATE_LINES` to fit the implementation.
2. Delete the unproven gate and its tests; keep the five that are proven, or
   cut to `frozen` + `hygiene` + `links` if needed.
3. Record the cut in the checkpoint and in the spec's open questions.
4. Do not leave a gate that prints green without a proof.

If the hooks block normal work because `book/` is absent on a fresh clone:

1. Confirm the change does not touch `book/` or `data/`.
2. Run `python tools/verify.py --fast --only hygiene --only links` to verify
   what can be verified.
3. Use `--no-verify` only as a recorded decision, and record the missing
   `frozen` proof in the checkpoint.
4. Do not change the gate to return 0 when the corpus is absent.

If a frozen-data change is discovered:

1. Stop all gate work.
2. Record the exact changed path and the audit output.
3. Do not restore from an untrusted copy; the user decides which copy is the
   frozen reference.
4. Do not commit anything until the reference is settled.

---

## 10. What this plan does not do

- It does not implement eBPF, IFB, BBRv3 or `.dist`.
- It does not run a campaign, a live bench or Graylog.
- It does not move `book/` into version control or create the second copy.
- It does not add CI, a document register, a weekly file, or a full mutation
  harness.
- It does not certify any live claim as verified.

The next action after approval is Task 0, not code.
