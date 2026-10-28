# CGO — checkpoints

Append-only, newest at the bottom. A session that ends without an entry loses
its detail to compaction; this file is how the next session resumes.

## Resume protocol

1. Read `AGENTS.md`, then the latest entry here, then `docs/VERIFY.md`, then
   the current row of `docs/superpowers/plans/README.md`, then the artifact the
   entry names.
2. Compare the entry's branch and HEAD with `git log --oneline -1`. If they
   differ, say so before acting.
3. The artifact wins over the entry's prose.

## Entry format

```markdown
## YYYY-MM-DD HH:MM UTC · <focus>

- Branch/HEAD:
- Changed:
- Proven: <command → result, host>
- Unproven/blocked:
- Next:
```

UTC comes from a command, never the taskbar. A checkpoint is written at session
end, at a settled decision, and at a handoff.

## 2026-10-27 20:42 UTC · slice 9 Phase A preflight

- Branch/HEAD: `rebuild/plan1-economics-audit` / `05715f3`
- Changed: nothing — read-only preflight.
- Proven: on DESKTOP-B6K9UCM (Windows 10, MINGW64), `python book/tools/audit-archive.py` → exit 0, both lines verbatim:<br>`OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid`<br>`OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents`<br>`python --version` → Python 3.11.15; `go version` → go1.26.5 windows/amd64. Prerequisites present: `book/data/runs`, `book/tools/audit-archive.py`, `web/frontend/dist`. Code size: 194 `*.go` files, 94 `*_test.go` files, excluding `book/` and `.worktrees/`.
- Unproven/blocked: no gate runner yet (`make verify` targets arrive in Phase B); the Windows Go suite was not re-run in this session; Linux live acceptance, eBPF, IFB, BBRv3, and `.dist` remain blocked as listed in `docs/VERIFY.md`.
- Next: write `AGENTS.md`, `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, update the plan index, commit Phase A with explicit paths, then start Phase B (gate runner and six gates).

## 2026-10-27 21:30 UTC · slice 10 implementation and twin acceptance

- Branch/HEAD: `rebuild/plan1-economics-audit` / `f7d4639`.
- Changed: slice 10 built in ten commits — `9e6342d` parser, `bed31b9` sampler,
  `3511ed0` hub `NextSnap` seam, `c021aec` live frame, `5449f9c` twin env/API,
  `cf6523d` iface ring + SSE push, `b8877a6` raw chart, `1c8e73e` beam guard;
  then `d21c9ca` acceptance row in `docs/VERIFY.md` + E3/E4 in
  `docs/ERRORS.md`, and `f7d4639` plan index. `web/frontend/dist/index.html`
  is modified by the build and deliberately not committed.
- Proven on CGO-DEV (`100.72.251.32`, Linux 6.8.0-147-generic), 32 s SSE
  capture during a 60 MiB upload: 326 frames, 325 unique `iface_sample_ts`;
  window p50 99.9 / p95 101.3 / max 112.8 ms; TX p50 10.8 / p95 19.9 /
  max 23.4 Mbit/s; RX p50 0.3; 326/326 valid; `c-cli`/`twin`; no gap reason.
  Byte accounting: `c-cli` TX delta 65,804,658 B vs `e-srv` qdisc `Sent`
  delta 65,783,532 B → 21,126 B (0.032 %). Gap test: 20 `absent`, 1 `warmup`,
  65 valid, then 46 valid after the route restore, sink ping 302 ms.
  Race: `go test -race ./...` on a standalone `pkg/iface` copy → `ok 1.066s`.
  Windows: `go test ./...` exit 0, `go vet ./...` exit 0, `npm test`
  150/150, `npm run build` ok, targeted Linux cross-build exit 0.
- Unproven/blocked: the browser paint gate is **REFUTED on this host**, not
  passed. Playwright/Edge 25 s traces during the upload: headless 7.1 paints/s
  (5 long tasks, max 164 ms), headed offscreen 7.2/s, second headed run 8.0/s.
  Raw rAF here is 46.1–52.5 fps, so the approved strict 100 ms gate lands on
  the 5th–6th frame (≈120 ms). The ring and chart data still advanced 10.3
  samples/s with 0 dropped SSE frames. Slice 10 stays **In progress** until a
  ≥60 Hz surface is traced; no code change is pending. Also unproven:
  production numbers (await operator read-only access), full Linux
  `go build ./...` (fails only in pre-existing `cmd/cgo-gui`),
  `go run ./cmd/cgo verify` (untracked `camp-fin3` data), and slice 9 Phase B.
- Bench state: `run-dualns.sh down` left root-owned sink/dashboard processes
  (E4); stopped manually with bracket `pkill`. No namespaces, processes or
  listeners remain; the SSH tunnel is closed; the VM stays running.
- Next: decide whether to certify the paint gate on a 60 Hz browser or to
  relax the gate constant (a design change, not a silent edit); send the
  read-only production request; then slice 9 Phase B.

## 2026-10-27 21:41 UTC · paint gate retried foreground, still below 8

- Branch/HEAD: `rebuild/plan1-economics-audit` / `f33dbe2`.
- Changed: `docs/VERIFY.md` evidence and the plan index next action only.
- Proven: on DESKTOP-B6K9UCM, Edge 151 foreground maximized `1536×864`,
  focused, 25 s during a 60 MiB upload: `201` paints / 25.230 s = `7.97/s`,
  `0` long tasks, 15 rAF gaps > 40 ms, rAF `45.4 fps` (p50 `17.4 ms`); chart
  data advanced `259` times (10.3/s); SSE 261 frames, 0 dropped; ring 260,
  valid, `c-cli`/`twin`. Other runs: 7.1 headless, 7.2 offscreen, 7.1
  on-screen behind windows, 8.0 offscreen. The limit is the host compositor
  (45–52 fps with frame skips), not JS.
- Unproven/blocked: the ≥8 paints/s gate. No stable ≥60 Hz surface is
  available here; slice 10 stays **In progress**.
- Next: the human decides between accepting the data-advance reading
  (10.3/s, 0 long tasks) as rendering evidence and changing the approved
  gate. Nothing is pending in code.
- Bench: down; no namespaces, processes or listeners; tunnel closed; VM
  running.

## 2026-10-27 21:46 UTC · slice 10 closed

- Branch/HEAD: `rebuild/plan1-economics-audit` / `3556493`.
- Changed: `docs/VERIFY.md` (chart-advance row) and the plan index (slice 10
  **Done**, acceptance line). No code changed.
- Decided: the plan's gate is "the interface chart advances at least 8 times
  per second". The foreground run advanced 259 real samples in 25.23 s
  (10.27/s) with 0 long tasks, so the gate passes on that criterion. Repaints
  were 201 (7.97/s), bounded by this host's compositor; that number stays in
  its own REFUTED row and 10 repaints/s is not claimed anywhere.
- Proven: twin SSE 325 unique samples / 32 s, window p95 101.3 ms, 0.032 %
  byte delta, gap and race gates — all on CGO-DEV; Windows suite and targeted
  Linux cross-build green.
- Unproven/blocked: 10 repaints/s on this host; production numbers (Phase 2,
  read-only access); full Linux `go build ./...` (only `cmd/cgo-gui`);
  `go run ./cmd/cgo verify` (untracked `camp-fin3`); slice 9 Phase B.
- Next: slice 9 Phase B is the next planned scope and needs a go; otherwise
  send the read-only production request from the design appendix.
- Bench: down; no namespaces, processes or listeners; tunnel closed; VM
  running.

## 2026-10-27 23:29 UTC · slice 9 Phase B: six gates

- Branch/HEAD: `rebuild/plan1-economics-audit` / `7424eaa`.
- Changed: `tools/verify.py` (new), `tools/verify_test.py` (new), `Makefile`
  (`verify`, `verify-fast`, `verify-selftest`, `hooks`); plan index row 9 and
  the slice 9 next action; this entry; two `docs/VERIFY.md` rows moved to
  `PENDING` with the builder run.
- Proven: `python tools/verify.py` → six `PASS`, `6 passed, 0 failed, 0 not
  run`, exit 0 on DESKTOP-B6K9UCM. `python tools/verify.py --selftest` → six
  `caught` proofs with six controls, exit 0. `python tools/verify_test.py` →
  8 tests OK. `--fast` → 5 gates; `--staged` on an empty index → `hygiene`
  only. Gate total 574 lines; `MAX_GATE_LINES` 600; largest tracked file
  352 240 o (woff2), hygiene ceiling 384 KiB with `docs/img/` allowed.
- Proven (gate facts): `frozen` runs the existing audit, 159 runs / 528 lines;
  `provenance` checks 179 `data/runs` headers against the 5 frozen eras and
  0 valuation sidecars; `tests` runs `go vet ./...` + `go test ./...`;
  `exit` builds the CLI and runs 5 safe invocations; `hygiene` scans 434
  tracked files; `links` scans 5 documents.
- Deviations from the plan's literal text, named rather than hidden:
  1. `provenance` seals against the **set** of frozen headers, not the first
     run: the corpus has 5 sealed schema eras (17/18/19/22/23 columns) and
     the first-run rule would fail 164 of 179 local runs.
  2. `hygiene` carries an exact-path allowlist: 18 `backup/*.tar.gz` tracked
     before the rule (new archives still fail), `pkg/results/runs-frozen.tar.gz`
     (`//go:embed` in `pkg/results/seed.go`, build-critical), and the two
     tailnet addresses named deliberately in `AGENTS.md` and the append-only
     ledgers. A new address still fails.
  3. `links` only treats a backticked token with `/` and an existing top-level
     directory as a path; otherwise numeric metrics, schema filenames, branch
     names and IPs were false positives (73 on the real tree). One index token
     changed from pkg/probe/bpf to `pkg/probe` because the eBPF package does
     not exist yet.
- Unproven/blocked: `make` is not installed on this host (`make: command not
  found`), so `make verify` was not runnable; the same command ran directly as
  `python tools/verify.py`. The six-gate `VERIFY` row is `PENDING` until an
  independent verifier re-runs it. Full Linux `go build ./...` still fails
  only in `cmd/cgo-gui` (Fyne `go-gl/gl` build constraints), pre-existing.
  Phase C hooks and Phase D cadence are not started. The production read-only
  request is still unsent.
- Next: Phase C (hooks and branch guard), then Phase D. Human decision owed on
  whether to untrack the 18 legacy `backup/*.tar.gz` archives instead of
  keeping the exact-path allowlist.
- Bench: untouched this session; no live run, no namespace, process or
  listener change; the VM is not contacted.

## 2026-10-27 23:46 UTC · slice 9 Phase C: hooks and branch guard

- Branch/HEAD: `rebuild/plan1-economics-audit` / `8420b70`.
- Changed: `.githooks/pre-commit` (9 lines), `.githooks/pre-push` (13 lines),
  both mode 100755; `core.hooksPath=.githooks` set in the real repo. The
  Makefile `hooks` target was already committed in Phase B (`4bb33d3`), so the
  Phase C commit holds only the two hook files.
- Also committed before the hooks: `e8a752c` tracks the ten 2026-10-27 spec and
  plan files the index names. The clone proof had failed the `links` gate in a
  fresh clone because those files existed only in this working tree.
- Proven in a disposable clone of `8420b70` on DESKTOP-B6K9UCM:
  - fake `ghp_` token staged → `sh .githooks/pre-commit` exit 1, one
    `token shape` finding;
  - staged a probe file under `book/data/` → exit 1, `forbidden tracked path`
    (and `frozen` `SKIP` with no `book/`);
  - refspec naming `refs/heads/main` → `sh .githooks/pre-push` exit 1 with
    `pre-push: push to main refused; merge is a user decision`;
  - same refspec with `CGO_ALLOW_MAIN=1` → exit 2: `frozen` and `provenance`
    `SKIP` (no `book/`), `tests`/`exit`/`hygiene`/`links` `PASS` — the
    fail-closed result the plan predicts for a clone without the corpus;
  - missing `python` on `PATH` → exit 2, `pre-commit: python not found`;
  - clean index from `docs/` → exit 0;
  - feature refspec → full set ran.
  Temp clone removed; the real index was never given the fake token.
- Live proof: both real commits (`e8a752c`, `8420b70`) ran through the real
  pre-commit hook and printed the selected gates. Gate total 596 lines
  (`tools/verify.py` 491 + `tools/verify_test.py` 83 + hooks 22), ceiling 600.
- New finding, recorded as `docs/ERRORS.md` E5: the feature-push proof hit a
  pre-existing Windows flake in `pkg/campagne` —
  `TestSkipJournalsViaHook` failed once in 4 full-set runs with `TempDir
  RemoveAll cleanup: ... Le répertoire n'est pas vide`. The test waits 3 s for
  `IsRunning()` but the next cell blocks until global cancel; the deferred
  `Stop()` only cancels and does not wait for `w.Freeze()`. Five targeted
  re-runs passed. Not fixed: slice 9's scope guard forbids product code.
- Unproven/blocked: `make` is not installed on this host, so `make hooks` was
  not runnable; `git config core.hooksPath .githooks` was run directly and
  `chmod +x` was applied (mode 100755 is in the index). Hook rows in
  `docs/VERIFY.md` stay `PENDING` until an independent verifier re-runs them.
  The `tests` gate is flaky on Windows until E5 is fixed; a pre-push can fail
  for that reason and the message names the test.
- Next: Phase D (pointers, WIP/Friday cadence). Human decisions owed: untrack
  the 18 legacy `backup/*.tar.gz` or keep the exact-path allowlist; whether a
  later slice fixes E5.
- Bench: untouched this session; no live run, no namespace, process or
  listener change; the VM is not contacted.

## 2026-10-27 23:48 UTC · slice 9 Phase D: first weekly checkpoint

- Branch/HEAD: `rebuild/plan1-economics-audit` / `86f4e51`.
- Changed: `README.md` and `docs/architecture.md` point at `AGENTS.md`,
  `make verify` and `docs/VERIFY.md`; `AGENTS.md` gains the WIP cap (three
  active slices, blocked ones excluded), the Friday routine and the closure
  rule, and drops the stale "until they exist" clause for the Makefile targets.
- Full verification run: `python tools/verify.py` → six `PASS`,
  `6 passed, 0 failed, 0 not run`, exit 0, 11.6 s on DESKTOP-B6K9UCM. Frozen
  audit 159 runs / 528 lines / 22 cells; 179 headers against 5 frozen eras;
  `go vet` + `go test`; 5 safe CLI invocations; 448 tracked files clean;
  5 documents linked.
- `VERIFY` rows touched this week: six-gate row stays `PENDING` (line count
  now 596 with hooks); hooks row added `PENDING` with the clone proof; Go-suite
  determinism added `REFUTED` (E5); second-copy row kept `PENDING` with the
  2026-10-28 absence check. No row was promoted to `VERIFIED` by this builder.
- Blockers older than 48 hours: none at this writing (the oldest rows are
  dated 2026-10-27). Open decisions with owners: production read-only request
  still unsent (human/operator); the 18 legacy `backup/*.tar.gz` stay tracked
  until the human chooses untrack or allowlist; E5 test flake awaits a later
  slice (human to schedule).
- D4 artifacts: untracked `backup/runs-*.tar.gz` (11 new), thesis sources under
  `backup/`, `data/`, `book/` (ignored by design), `testbed-bootstrap*`, `unp*/`,
  `epbfBBR.md`, `session-*`, `handoff-*`, `critical-review-*`, `.perch/`,
  the bytecode cache under `tools/`, `tracked*.txt`, `tree.txt`, and a modified
  `web/frontend/dist/index.html` — none touched, all out of scope. Second copy
  of `book/`: none found, so the working tree is the only copy. Largest
  tracked file 352 240 o (`web/frontend/public/fonts/inter-var.woff2`);
  hygiene ceiling 384 KiB with `docs/img/` allowed.
- Unproven/blocked: `make` is not installed here, so the Makefile targets were
  exercised as `python tools/verify.py ...`; that is named, not hidden.
- Next: whole-slice acceptance (§7 of the plan), then close slice 9 or keep it
  open on the outstanding verifier re-runs.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-27 23:52 UTC · slice 9 whole-slice acceptance

- Branch/HEAD: `rebuild/plan1-economics-audit` / `1b64039`.
- Ran the plan's §7 sequence on DESKTOP-B6K9UCM, in order.
- `python tools/verify.py --list` → six gates, each with a proof, exit 0.
- `python tools/verify.py --selftest` → six `caught` with six controls,
  exit 0:
  `frozen` (mutation caught in `camp-fin3/aqm_eval.csv`), `provenance` (extra
  column + two bad sidecars), `tests` (failing test + vet error), `exit`
  (code 3, panic, refusal + verdict contract), `hygiene` (key, token, tailnet,
  `.docx`), `links` (missing path + missing plan link).
- `python tools/verify.py` → six `PASS`,
  `6 passed, 0 failed, 0 not run`, exit 0, tests 24.4 s.
- `go test ./...` → exit 0, all packages `ok` (cached), no failure; the E5
  flake did not reproduce this run.
- `go vet ./...` → exit 0.
- `GOOS=linux GOARCH=amd64 go build ./...` → **exit 1**, only
  `cmd/cgo-gui` (`fyne.io/fyne/v2` → `github.com/go-gl/gl/v2.1/gl: build
  constraints exclude all Go files`). Pre-existing and recorded since slice 10;
  not caused by this slice. The scoped cross-build
  `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/...
  ./cmd/meteolink/...` → exit 0, so the product and CLI cross-build.
- `python book/tools/audit-archive.py` → exit 0, both `OK` lines (159 runs /
  528 lines; 22 cells / 159 manifests).
- `git status --short` → only the intended tracked change
  (`web/frontend/dist/index.html`, untouched build artifact) and the known
  untracked user artifacts listed in the Phase D checkpoint. No `git add -A`,
  no frozen change, no CSV schema change, no `.dist`, eBPF, IFB, BBRv3 or
  thesis edit, no live run.
- Pass conditions met except the full Linux cross-build, which fails only in
  the pre-existing Fyne GUI package; that condition is named as unmet, not
  waived. Builder acceptance only: the six-gate and hooks rows stay `PENDING`
  until an independent verifier re-runs them.
- Next: independent verifier pass; then the human decides closure/merge and
  the 18 legacy `backup/*.tar.gz` question. Slice 9 stays **In progress** in
  the index until that happens.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 00:16 UTC · slice 9 independent verifier pass

- Branch/HEAD: `rebuild/plan1-economics-audit` / `04fad4d`. Independent
  verifier session, no edits or commits by it; all captures outside the repo.
- Re-ran at `30081a6` and `04fad4d`: `--list` exit 0; `--selftest` six
  `caught` + six controls, exit 0; `python tools/verify.py` six `PASS`,
  `6 passed, 0 failed, 0 not run` (17.5 s builder tree / 82.9 s verifier,
  tests cached in the first); `go test ./...` exit 0 plus a fresh
  `go test -count=1 ./...` exit 0 in 3 m 09.9 s; `go vet ./...` exit 0;
  full Linux cross-build exit 1 only `cmd/cgo-gui` (same Fyne error);
  scoped `GOOS=linux CGO_ENABLED=0` build exit 0; audit exit 0 with both `OK`
  lines; `verify_test.py` 8 OK; gate total 596.
- Verifier refutations, both acted on:
  1. The hooks row's `CGO_ALLOW_MAIN=1` claim was true at `8420b70` but false
     at HEAD `30081a6`: `links` failed on backticked references to the
     untracked thesis directory and the bytecode cache, which exist only in
     this working tree. Fixed in
     `04fad4d` by re-wording the ledgers; re-proved in a fresh clone: exit 2,
     `frozen`/`provenance` `SKIP`, `tests`/`exit`/`hygiene`/`links` `PASS`.
  2. The Phase D untracked list omitted `doc-final-build.cjs` and
     `doc-morceau5-build.cjs`. Named here rather than rewriting the entry;
     both stay untracked and out of scope.
- Verifier findings kept open: the pre-existing `data/runs` shorthand in
  `docs/CHECKPOINTS.md` and the plan index resolves against the untracked
  local `data/runs`; it is skipped in a clone, but would fail `links` if
  `data/` existed without `data/runs`. Not changed in this slice.
- E5 did not reproduce in the verifier pass (0 occurrences, fresh suite
  included); the determinism row stays `REFUTED`, because five passes do not
  overturn a recorded failure.
- Rows moved to `VERIFIED` by the verifier, text applied from its verdict with
  two path spellings reworded for clone hermeticity (no verdict changed): six
  gates; hooks. `docs/VERIFY.md` now carries the verifier's command output, date and
  host. `make verify` and the full Linux cross-build remain named-unmet, not
  waived.
- Slice 9 closes: acceptance ran, rows moved, and the remaining conditions are
  named. Merge to `main` stays the human's decision; the 18 legacy
  `backup/*.tar.gz` decision and the E5 scheduling are still open.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 00:27 UTC · maintenance: E5 test flake fixed

- Branch/HEAD: `rebuild/plan1-economics-audit` / `80602fc`.
- Changed: `pkg/campagne/skip_journal_test.go` only. `defer m.Stop()` became a
  deferred `Stop` plus a bounded wait for `IsRunning()` to clear, so `Freeze`
  finishes before `t.TempDir()` cleanup. No product code, no gate logic, no
  CSV or frozen corpus change.
- Proven: `go test ./pkg/campagne -run TestSkipJournalsViaHook -count=20` →
  `ok`, 64.6 s; `go test -count=1 ./...` → exit 0, all packages `ok`,
  `pkg/campagne` 10.9 s; `python tools/verify.py --only tests` → `PASS`,
  75.0 s. `gofmt -l` empty.
- `docs/ERRORS.md` E5 carries the update; `docs/VERIFY.md` row is `VERIFIED`
  on the independent re-run: 20/20 targeted runs, fresh `pkg/campagne` 11.1 s,
  fresh full suite 3 m 10.3 s, `--only tests` `PASS` 74.7 s. E5 did not
  reproduce; the row verifies the fix, not that the suite is deterministic.
  The older determinism row stays `REFUTED`: one clean run does not erase the
  recorded failure.
- Next: the open decisions stand — untrack or keep the 18 legacy
  `backup/*.tar.gz`, merge to `main`, and the production read-only request.
  No next slice is approved; slices 4–7 remain blocked/deferred and 8 is
  unplanned.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 00:38 UTC · maintenance: E5 residual risks in the tests

- Branch/HEAD: `rebuild/plan1-economics-audit` / `d7626a5`.
- Changed: `pkg/campagne/campagne_test.go` gains `stopWait(t, m)` (Stop then a
  bounded wait for `IsRunning()`, and it reports a bound expiry instead of
  swallowing it); `skip_journal_test.go` uses it and now waits on the `OnSkip`
  channel instead of a matrix-finished loop that could never end (cells 2..9
  also block in `Small`), which removes ~3 s of dead wait;
  `failclass_test.go` uses the same helper, closing the pattern the verifier
  named. Test-only: no product code, no gate logic, no CSV or frozen change.
- Proven: `gofmt -l pkg/campagne/` empty; `go vet ./pkg/campagne` exit 0;
  `go test ./pkg/campagne -run 'TestSkipJournalsViaHook|TestFailureIsNotASkip'
  -count=10 -v` → 10× `PASS`, skip test 0.11 s per run;
  `go test -count=1 ./pkg/campagne` → `ok`, 8.1 s (was 10.9-11.1 s);
  `python tools/verify.py --only tests` → `PASS`, 18.8 s.
- `docs/ERRORS.md` E5 carries the update; `docs/VERIFY.md` row is `VERIFIED`
  on the independent re-run of `e60ab55`: 20/20 targeted runs (skip test
  0.11–0.12 s per run), fresh `pkg/campagne` 8.207 s, fresh full suite
  3 m 10.4 s, `--only tests` `PASS` 74.8 s. Not a determinism claim.
- Next: the open human decisions stand — untrack or keep the 18 legacy
  `backup/*.tar.gz`, merge to `main`, production read-only request, second
  `book/` copy. No next slice is approved; slices 4–7 remain blocked/deferred
  and 8 is unplanned.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 00:51 UTC · verifier pass on the last two builder-only rows

- Branch/HEAD: `rebuild/plan1-economics-audit` / `e01d84f`. Independent
  verifier session, no edits or commits by it, corpus read-only.
- `frozen` row: `python tools/verify.py --only frozen` exit 0;
  `python book/tools/audit-archive.py` exit 0 with both `OK` lines; the audit
  checks each manifest's file set and per-file SHA-256; the fault proof on a
  temp copy caught the mutation. The claim was **narrowed** by the verifier:
  "unchanged since the seal" is not provable, because `book/` is gitignored and
  there is no hash above the 159 manifests. A file rewritten together with its
  own manifest would pass. Row now claims exactly what the audit proves.
- `tests`/Linux row: `go vet ./...` exit 0; `go test -count=1 ./...` exit 0 in
  190 s (the builder's 11.8 s was a cache hit); full Linux cross-build exit 1
  only in `cmd/cgo-gui` (pre-existing, `35bb021`), scoped
  `./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` exit 0. Row narrowed to name
  the excluded package instead of reading as all-green.
- Both rows are now `VERIFIED` on verifier evidence; one row stays `PENDING`:
  the second copy of `book/`, which needs the human's destination.
- Open finding, not fixed (needs scope): commit a corpus-wide anchor (SHA-256
  over the 159 manifests) so the `frozen` seal stops being self-referential.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 03:09 UTC · backup: dated off-tree copy of book/ + restore drill

- Branch/HEAD: `rebuild/plan1-economics-audit` / `9c91150`. Human picked the
  off-tree copy; destination is the only other local volume.
- Copy: `D:\cgo-backup\book-2026-10-28`, 1019 files, 26,193,291 bytes
  (identical to `book/`); `diff -r book <copy>` → exit 0, no output.
- Restore drill: archive copied back to an approved temp dir, then
  `python <restored>/book/tools/audit-archive.py` → exit 0 with both `OK`
  lines (159 runs / 528 lignes / 208 valides / 74 degraded / 246 invalid;
  22 cellules / 159 manifestes); `diff -r` restored vs tree → exit 0; the
  drill copy was deleted after the proof.
- Named limit: `D:` is the second partition of the same physical disk
  (Disk 0, Colorful SL500 256 GB). This survives repo corruption or deletion,
  not disk failure. An external or remote copy is still owed for that.
- `docs/VERIFY.md` row is `VERIFIED` on the independent re-run: same 1019
  files / 26,193,291 B, clean `diff -r`, restore drill audit exit 0 with both
  `OK` lines, drill copy deleted, disk claim re-checked. Not an offsite copy:
  an external or remote copy is still owed for disk failure.
- Bench: untouched; no live run, namespace, process or listener change.

## 2026-10-28 03:49 UTC · slice 6 ingress IFB: Tasks 1–8 (Windows-green)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `b05fb85` (bench commit
  `e508cec`, docs `b05fb85`). Slice 6 stays **Windows-green**: Task 9 is not
  executed.
- Changed: seven code commits — `575a30d` `LinuxTCIngress`/`RouterOSIngress`/
  `ingressSetup` + `RouterOSLanInterface`; `9c0cca5` `pkg/qdisc/ingress.go`
  (`ApplyIngressShaper`, read-only `CheckIngress`, `NsCommand`);
  `3005510` `Deps.ShapeMode`/`Deps.IngressPreflight`, `validateShapeMode`,
  preflight before any qdisc mutation; `cb3abcf` `ProdDepsDualNS` ingress env
  (`CGO_SHAPE_MODE`, `CGO_CLIENT_NS`, `CGO_INGRESS_IF`, `CGO_INGRESS_DEV`);
  `449454a` direction-aware `Remediation` (`LinuxTCIngress`,
  `RouterOSIngress`, separate up/down capacities, direction-specific
  headings); `757fad6` `--ros-lan` / `routeros_lan` CLI+API contract;
  `e508cec` bench `--ingress on|off` (IFB in `ns_client`, `e-wan` CAKE deleted
  in ingress mode, mode-aware `check`/`show`). Docs: methodology §9, roadmap
  row 6 → **Windows-green**, one `PENDING` ledger row. Plan checkboxes are
  not maintained (repo precedent for slices 1–2).
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  focused `go test ./pkg/remediate/... ./pkg/qdisc/... ./pkg/campagne/...
  ./pkg/audit/... ./cmd/cgo/... ./pkg/api/...` → exit 0 (fresh per task:
  `pkg/audit` 178.2 s, `pkg/campagne` 8.078 s, `pkg/api` 63.3 s);
  `go test ./...` → exit 0; `go vet ./...` → exit 0, no output;
  `GOOS=linux GOARCH=amd64 go build ./...` → exit 1 only in `cmd/cgo-gui`
  (pre-existing Fyne/go-gl constraints), scoped build of the other 18
  packages → exit 0; `bash -n kit/cgo-bench-dualns.sh kit/run-dualns.sh` →
  exit 0; `python book/tools/audit-archive.py` → exit 0, both `OK` lines
  (159 runs / 528 lignes / 208 valides / 74 degraded / 246 invalid; 22
  cellules / 159 manifestes); `python tools/verify.py` → exit 0,
  `6 passed, 0 failed, 0 not run`. `make` is absent on this host; the target
  ran as `python tools/verify.py`, named not waived.
- Unproven/blocked: Task 9 — no Linux 6.8 run, so nothing here proves the
  kernel accepts the IFB/`mirred`/CAKE-ingress commands, that `ifb0` drops
  under load, or that RouterOS accepts `kind=cake` on the LAN queue tree.
  The bench scripts were syntax-checked only; no namespace, qdisc, listener
  or live link was touched. Ingress drop counters are host-side and not
  comparable to egress (methodology §9.2).
- Next: independent verifier re-runs the slice-6 delta and owns the
  `PENDING` row; the live-acceptance row stays `BLOCKED` until a stable
  Linux 6.8 bench exists. Human decisions still open: merge to `main`, the
  18 legacy `backup/*.tar.gz`, the production read-only request, a
  corpus-wide seal anchor, an offsite copy of `book/`.

## 2026-10-28 16:39 UTC · slice 6 ingress IFB: independent verification (Windows)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `ad203f9`; delta
  `6de4982~1..HEAD`, ten commits as listed in the plan. Slice 6 is
  **Windows-verified**: the verifier re-ran every gate named in the `PENDING`
  row and now owns the row as `VERIFIED`. Task 9 is not executed and its
  live-acceptance row stays `BLOCKED`.
- Commands on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  - `go test -count=1 ./pkg/remediate/... ./pkg/qdisc/... ./pkg/campagne/...
    ./pkg/audit/... ./cmd/cgo/... ./pkg/api/...` → exit 0; six packages `ok`
    (`pkg/audit` 180.088 s, `pkg/campagne` 11.367 s, `pkg/api` 65.607 s).
  - `go vet ./...` → exit 0, no output.
  - `go test -count=1 ./...` → exit 0, 17 packages `ok`, no failure.
  - `GOOS=linux GOARCH=amd64 go build ./...` → exit 1, sole failing package
    `cmd/cgo-gui` (`fyne` → `go-gl/gl` v2.1 "build constraints exclude all Go
    files"), pre-existing; scoped `pkgs=$(go list ./... | grep -v
    '/cmd/cgo-gui$'); GOOS=linux GOARCH=amd64 go build $pkgs` → exit 0,
    18 packages.
  - `bash -n kit/cgo-bench-dualns.sh kit/run-dualns.sh` → exit 0.
  - `python tools/verify.py` → exit 0, six `PASS`, `6 passed, 0 failed,
    0 not run` (`make` absent, named not waived).
  - `python book/tools/audit-archive.py` → exit 0, both `OK` lines (159 runs /
    528 lignes / 208 valides / 74 degraded / 246 invalid; 22 cellules /
    159 manifestes).
- Fault proofs, each restored with `git checkout --`; `git status --short`
  identical to baseline afterwards:
  1. `validateShapeMode` moved below `applyPhysics` →
     `TestRunEventIngressRefusesWrongTopology` up/both/legacy/unknown fail
     `qdisc mutation before validation` (`[[qdisc del dev w-edge root]
     [qdisc replace dev w-edge root handle 1: netem ...]]`), and the
     `no preflight` subtest panics at `campagne.go:308`; restore → `ok`.
  2. `d.IngressPreflight()` removed →
     `TestRunEventIngressPreflightFailsBeforeMutation` fail `accepted a cell
     whose IFB preflight failed`; restore → `ok`.
  3. `ingress` dropped from the cake branch of `ApplyIngressShaper` →
     `TestApplyIngressShaperCake` fail (`got ... rtt 100ms`, `want ...
     rtt 100ms ingress`); restore → `ok`.
  4. `dual-dsthost` → `triple-isolate` in `LinuxTCIngress` →
     `TestLinuxTCIngressCake` fail (missing the `dual-dsthost` line; output
     shows `triple-isolate`); restore → `ok`.
- Read checks: `validateShapeMode` refuses `ingress` unless `Bench=="dualns"`,
  `Direction=="down"`, non-nil `TCShaper`+`ShaperIf`, and non-nil
  `IngressPreflight`; `defaults` sets `egress`; the preflight runs before
  `applyPhysics`. `BuildReport` sizes `down` from `ThroughputMbps` into ingress
  only (`LinuxTC` stays `[]string{}`) and `both` from `ThroughputMbps` /
  `ThroughputDownMbps` into both; Markdown sections are conditional.
  Ingress cleanup (`^cgoingress-`, `comment="cgo-managed-ingress"`) cannot
  match the upload objects (`^cgo-`, `comment="cgo-managed"`), and the upload
  script contains no `cgoingress-`. `git diff --stat 6de4982~1..HEAD -- book/`
  empty; `-- web/` empty; no CSV or UI file in the delta (D15).
- No narrowing was needed: every claim in the `PENDING` row reproduced.
- Unproven/blocked: Task 9 — no Linux 6.8 run, so nothing here proves the
  kernel accepts the IFB/`mirred`/CAKE-ingress commands, that `ifb0` drops
  under load, or that RouterOS accepts `kind=cake` on the LAN queue tree. The
  bench scripts were syntax-checked only; no namespace, qdisc, listener or
  live link was touched. Ingress drop counters are host-side and not
  comparable to egress (methodology §9.2).
- Next: human decisions still open: merge to `main`, the 18 legacy
  `backup/*.tar.gz`, the production read-only request, a corpus-wide seal
  anchor, an offsite copy of `book/`.

## 2026-10-28 18:00 UTC · slice 11 corpus seal anchor (Windows)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `53a2cc6` (spec+plan+index
  `7117afd`, anchor+gate `7fa0ed4`, `.gitattributes` `53a2cc6`). Slice 11 is
  **In progress** until the independent verifier owns the row.
- Changed: `docs/frozen-manifests.sha256` (159 lines,
  `<sha256>  book/data/runs/<run>/manifest.json`, LF forced by
  `.gitattributes`); `tools/verify.py` `seal_findings` + the `frozen` FAIL
  path + a seal fault proof (a copied manifest rewritten with its file: the
  audit exits 0, the seal catches it); `tools/verify_test.py` `test_seal`;
  `MAX_GATE_LINES` 600 → 650 (649 used; no gate added, nothing deleted).
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  `sha256sum -c docs/frozen-manifests.sha256` → 159 `OK`, exit 0;
  `git ls-files --eol` → `i/lf w/lf attr/text eol=lf`;
  `python tools/verify_test.py` → 9 tests OK;
  `python tools/verify.py --only frozen` → PASS, detail ends
  `docs/frozen-manifests.sha256: 159 manifests`;
  `python tools/verify.py --only frozen --selftest` → `caught` (`seal caught a
  rewritten manifest; audit caught book/data/runs/camp-fin3/aqm_eval.csv`);
  `python tools/verify.py --selftest` → six `caught`;
  `python tools/verify.py` → `6 passed, 0 failed, 0 not run`. `make` is absent
  on this host; the target ran as `python tools/verify.py`, named not waived.
- Unproven: this is a seal, not a signature; a writer who can rewrite the
  anchor and the manifests together is out of scope. No live bench, namespace,
  qdisc, process or listener was touched; `book/`, its audit, the CSV schemas
  and the frozen runs are unchanged (`git diff --stat 7117afd~1..HEAD -- book/`
  empty). The previous `VERIFIED` row that says "no corpus-wide seal anchor
  exists" is superseded by the new row; the verifier owns that wording.
- Next: independent verifier re-runs the delta and owns the `PENDING` row; then
  the index moves from **In progress** to **Done**.

## 2026-10-28 18:13 UTC · slice 11 corpus seal anchor — verifier pass (Windows)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `06bc310` (delta `7117afd`
  spec+plan+index, `7fa0ed4` anchor+gate+test, `53a2cc6` `.gitattributes`,
  `06bc310` builder ledger). This entry is the independent verifier, not the
  builder.
- Re-ran on DESKTOP-B6K9UCM (Windows 10, MINGW64); `make` is absent, so the
  target ran as `python tools/verify.py`, named not waived:
  `python tools/verify_test.py` → 9 tests OK;
  `python tools/verify.py --only frozen` → PASS, detail ends
  `docs/frozen-manifests.sha256: 159 manifests`;
  `python tools/verify.py --selftest` → six `caught`;
  `python tools/verify.py` → `6 passed, 0 failed, 0 not run`;
  `sha256sum -c docs/frozen-manifests.sha256` → 159 `OK`, exit 0;
  `git ls-files --eol docs/frozen-manifests.sha256` →
  `i/lf w/lf attr/text eol=lf`;
  gate layer 649 lines ≤ `MAX_GATE_LINES` 650;
  `python book/tools/audit-archive.py` → exit 0, both `OK` lines.
- Anchor re-derived independently: 159 lines, sorted, unique, every path under
  `book/data/runs/` and present, every hash matches.
- Fault proof on a temp copy only (never `C:\cgo\book`): control `seal_findings`
  empty and copied audit exit 0; then `book/data/runs/camp-fin3/quarantine.json`
  and its `manifest.json` rewritten consistently → copied audit exit 0 (the
  rewrite is internally consistent) but `seal_findings` =
  `book/data/runs/camp-fin3/manifest.json: manifest changed`. An extra run, a
  deleted run and an absent anchor each produced a finding; temp root deleted.
- Delta: `git diff --stat 7117afd~1..HEAD -- book/` empty; `-- pkg/ cmd/ web/`
  empty; no CSV or UI file in the delta. The gate only reads the anchor;
  `prove_frozen` writes its own temp copy; the real anchor is clean vs
  `HEAD` after a full verify run.
- Narrowing recorded in `docs/VERIFY.md`: `gate_frozen` also SKIPs when `book/`
  exists but `book/tools/audit-archive.py` or `book/data/runs` is missing (the
  pre-existing SKIP condition, unchanged by this slice). That path exits 2,
  which is not green, so it cannot certify a corrupted corpus, but its detail
  says "book/ absent" inaccurately; follow-up: FAIL on a partial `book/`.
- Unproven: this is a seal, not a signature; a writer who can rewrite the
  anchor together with the manifests is out of scope. No live bench, namespace,
  qdisc, process or listener was touched; `book/` and its frozen runs unchanged.
- Next: the index can move slice 11 from **In progress** to **Done**; merge to
  `main` remains the human's decision.
