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

## 2026-10-28 20:01 UTC · slice 12 approved, not implemented

- Branch/HEAD: `rebuild/plan1-economics-audit` / `f2a6f07`.
- Changed: `docs/superpowers/specs/2026-10-28-cgo-book-truth-up-design.md`,
  `docs/superpowers/plans/2026-10-28-cgo-book-truth-up.md`, and plan index
  row 12. No `book/` file touched.
- Proven: the spec and plan are committed; the commit's pre-commit hook ran
  `hygiene` and `links`, `2 passed, 0 failed, 0 not run`. No book gate was
  run this session.
- Unproven/blocked: the book still carries the stale title, the three-site
  field-audit claim, the retail cost default, the English "eight gates"
  sentence, and no coverage of the corpus seal, the live interface chart, or
  IFB ingress. All are scoped in the plan. Production measurement, field
  audit, live Linux acceptance, eBPF, BBRv3, and `.dist` remain blocked or
  deferred as in `docs/VERIFY.md`.
- Next: execute the plan inline, Task 1 first (baseline word counts, numeric
  snapshot, audit), then Tasks 2–12. The user asked to stop before any book
  edit.

## 2026-10-28 20:24 UTC · slice 12 book truth-up and prose restructure

- Branch/HEAD: `rebuild/plan1-economics-audit` / `36bfc14` (book working tree;
  `book/` is gitignored and was not staged). The IDEA memos stayed out of the
  book as the spec requires: contradictory numbers, forbidden hostname/IP, and
  unperformed production claims.
- Changed: `book/content/00-front.js`, `01-intro.js`, `015-organisme.js`,
  `02-partie1.js`, `03-partie2.js`, `04-pratique.js`, `06-conclusion.js`,
  `08-glossaire.js`, `09-annexes.js`, and the cover `00-couverture.xml`
  (working tree, never committed); `docs/CHECKPOINTS.md`, `docs/VERIFY.md`,
  `docs/superpowers/plans/README.md`.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  `python book/tools/audit-archive.py` → exit 0, both `OK` lines
  (`159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid`; `22 cellules,
  159 manifestes, inventaire, texte et tableau 15 cohérents`);
  all ten content modules load under node;
  `node book/build-docx.js` → `book/memoire-rebuild.docx` (3054 Ko);
  `memoire.docx` SHA-256 unchanged (`bd171227…6a2e030`);
  `python book/tools/audit-biblio.py` → 47/47 references cited;
  `python book/tools/audit-renvois.py` → 40 literal references, one
  pre-existing Figure 2 caption warning unchanged;
  word counts body 14,761 and total 16,462 (baseline 24,373), within the plan's
  ~14,500 body and ~16,400 total targets;
  new title present in `word/document.xml`; old title, `eight quality gates`,
  `160 runs`, `trois sites`, and `conditions réelles de l’institution` absent.
- Numeric-token diff: 31 tokens disappeared, each justified here; no result
  value was lost. `01-intro` `2016`, `2026`: compressed intro; the BBR date and
  the tariff year remain covered in chapters II and I. `02-partie1` `0,32 1,12
  106 237 1,25 7 5,6 120 166 680`: descriptive operator-catalogue and
  public-debate figures deleted by plan Task 5 Step 4; none feeds a result or a
  cost row. `03-partie2` `0,15 0,3 0,36 13,9`: components of the R-score worked
  example; the four scores (`92,9`, `91,4`, `71,5`, `92,1`) remain. `2011`:
  duplicate of the introduction. `72`: now spelled « soixante-douze ».
  `04-pratique` `300`: UI Campagne detail compressed into Annex E; `4.4`: old
  section reference removed; `71,5 91,4 92,1 92,9`: theoretical R scores removed
  from the RQ3 answer, retained in chapter II. `06-conclusion` `159 208 528`:
  de-duplicated, retained in chapter III and the annexes; `27`: citation moved
  to the chapter III limits and restored there; `5.8`: old section reference
  removed. `09-annexes` `79`: Figure 12 commentary compressed; `8.8`: hardcoded
  `8.8.8.8` target removed from the preset prose, the command takes a site
  parameter. Required values (`324,25`, `209,25`, `210,3`, `51,75`, `45,75`,
  `46,15`) and the six Table 15 rows remain.
- Claims table (spec §5): all ten resolved — title, field-audit claim,
  bench-versus-field separation, enterprise-primary cost framing, seven measured
  gates, interface chart in Annex E labelled `twin`, IFB ingress labelled with
  its blocked live acceptance, corpus seal, regime label, and blocked eBPF /
  BBRv3 / `.dist` kept out of the feature prose. Table 13 stays in chapter III
  to preserve Word `SEQ` numbering (moving it would have renumbered Table 15);
  its walkthrough moved to Annex E.
- Unproven/blocked: production measurement (read-only access pending); field
  audit (not executed); Linux 6.8 live acceptance, eBPF, BBRv3, and `.dist`
  (blocked or deferred as in `docs/VERIFY.md`). `book/` is a working tree only;
  supervisor, jury, date, and Master-vs-Licence placeholders remain the
  human's. No `book/` path was staged.
- Next: human review of the rendered book; fill the front-matter placeholders;
  merge to `main` remains the human's decision.

## 2026-10-28 20:52 UTC · slice 13 production-facing reframing

- Branch/HEAD: `rebuild/plan1-economics-audit` / `31ad4b1` (book working tree;
  `book/` is gitignored and was not staged).
- Changed: `PRODUCT.md` (description and thesis title; gitignored, working tree
  only); `book/content/00-front.js`
  and `book/content/00-couverture.xml` (title page/cover); `01-intro.js`,
  `03-partie2.js`, `04-pratique.js`, `06-conclusion.js`; the stale extraction
  headers in `015-organisme`, `03-partie2`, `04-pratique`, `07-refs`,
  `08-glossaire`, `09-annexes`; new spec and plan; plan index; this checkpoint.
- Title decision: the latest IDEA draft supplies the main title,
  « Architecture d’ordonnancement et de régulation de trafic pour la
  sécurisation des flux temps réel du système d’information de Météo
  Madagascar ». The subtitle is written for the evidence actually produced:
  « audit client-side, banc reproductible et comparaison des politiques AQM/BBR
  sur les liens d’accès contraints ». The memos’ production numbers, hostname/IP,
  and alternate titles stay excluded; the chapter I limitation paragraph still
  says the production server was not measured.
- Added to the book: one edge-architecture sentence in the introduction and one
  in the conclusion; one chapter II paragraph for Model A (capacity
  depreciation, 1 388,89 Ar/h at GOLD), Model B (metered volume, ×20 for 180 s),
  Model C (WIS 2.0 SLA, 220 ms / 95 % as the DGM operator objective, not a WMO
  mandate), the primary-scenario mapping (fiber→GOLD, 4G/5G→APN 150 Ar/Go,
  VSAT→wholesale 81,38 Ar/Go) and the contractual-scenario disclaimer; one
  limits paragraph naming eBPF, BBRv3 and trace-driven `.dist` as perspectives
  whose prerequisites (object/Linux 6.8, kernel 7.1, ≥30-sample trace) are
  absent. No production claim was added.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  `python book/tools/audit-archive.py` → exit 0, both `OK` lines;
  all ten modules load under node;
  `node book/build-docx.js` → `book/memoire-rebuild.docx` (3055 Ko);
  `memoire.docx` SHA-256 unchanged (`bd171227…6a2e030`);
  `python book/tools/audit-biblio.py` → 47/47 cited;
  `python book/tools/audit-renvois.py` → 40 literal references, one
  pre-existing Figure 2 warning;
  body 15,042 words, total 16,743, within the slice ceilings 15,300/17,000;
  new title present in `word/document.xml`, old title absent;
  six stale headers removed; no `extrait automatiquement` remains.
- Numeric-token diff against the slice-12 snapshot: 0 lost; 8 added, all
  contractual-scenario or blocked-directive values: `180` (Model B window),
  `2.0` (the 2³⁰ token), `388,89` (from 1 388,89), `600` (from 3 600), `720`
  (monthly hours), `81,38` (wholesale Ar/Go), `6.8` (Linux prerequisite), `7.1`
  (BBRv3 kernel prerequisite). Every slice-12 required string and Table 15 row
  is unchanged.
- Unproven/blocked: production measurement (read-only access pending); field
  audit (not executed); Linux 6.8 live acceptance, eBPF, BBRv3, `.dist`
  (blocked or deferred as in `docs/VERIFY.md`); the tariff values are
  contractual scenarios, not measurements; independent verifier pending for
  this slice. No `book/` path was staged.
- Next: human review of the rendered book; production access remains the
  unblock condition for the live claims; merge to `main` is the human's
  decision.

## 2026-10-28 21:22 UTC · slice 8 audit surface and citation (builder pass)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `5293799` (pre-commit).
- Changed: `web/frontend/src/lib/types.ts` (ingress fields);
  `web/frontend/src/components/AuditReportCard.tsx` (ingress block, phases,
  shared `Provenance`, citation action); new
  `web/frontend/src/lib/citation.ts` and `citation.test.ts`; new
  `AuditReportCard.dom.test.tsx`; `docs/api.md` (`/api/remediate`,
  `/api/audit/presets`, CSV provenance correction); spec, plan, index,
  `docs/ERRORS.md`, `docs/VERIFY.md`, this checkpoint.
- Behaviour: download remediation renders only when `linux_tc_ingress` or
  `routeros_ingress` is non-empty; `phases` renders only the present windows,
  never zero-filled; the report card shows exactly one `data-testid="provenance"`
  and keeps both seal buttons plus the Markdown link; `CITER CE RAPPORT` builds
  the citation from report fields only (id, site, link/provider, timestamp,
  direction, full `source_sha256`), omits missing fields, and returns empty for
  an empty id. `docs/api.md` now matches the code for the two endpoints and the
  CSV/Markdown provenance distinction.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64):
  `cd web/frontend && npm test` → 40 files, 156 tests passed (6 new: 3 citation,
  3 SSR DOM);
  `npm run build` → `tsc --noEmit` + Vite + bundle check, exit 0,
  1038.7 KB raw / 339.9 KB gz;
  `go test ./pkg/api/... ./cmd/cgo/...` → `ok`, 67.2 s / 11.9 s;
  `python tools/verify.py` → first run FAIL `tests` on
  `TestMatrixFilteredSubmatrix` with the known E5-family Windows
  `TempDir RemoveAll` cleanup shape (this slice changed no Go file); targeted
  `go test ./pkg/campagne -run TestMatrixFilteredSubmatrix -count=1` → `PASS`,
  4.2 s; `python tools/verify.py --only tests` → `PASS`, 29.8 s; full re-run
  → `6 passed, 0 failed, 0 not run`.
- Unproven: no live audit, bench, namespace, qdisc, or listener; the card is
  exercised with fixture reports only. `web/frontend/dist/index.html` stays a
  known working-tree modification from the build, not a commit. No `book/`
  edit. The independent verifier has not re-run the commands; the `VERIFY` row
  stays `PENDING`.
- Next: independent verifier re-runs the four commands and owns the `VERIFIED`
  row; the human pastes the citation into the thesis.

## 2026-10-29 15:41 UTC · deferred engines Windows pass, integrity truth, Synthèse UI

- Branch/HEAD: `rebuild/plan1-economics-audit` / `9a59276` before the ledger
  commit (code commits `53bd6a2`, `41ef618`, `ca09893`, `9a59276`). `book/` and
  `PRODUCT.md` stay working-tree only.
- Changed: `pkg/results`, `pkg/api`, `web/frontend` integrity counts; `pkg/model`,
  `pkg/doctor`, `pkg/campagne`, `cmd/cgo`, `kit/install-bbr3.sh` for BBRv3;
  `pkg/audit`, `pkg/qdisc`, `pkg/campagne`, `cmd/cgo` for trace-driven `.dist`;
  `web/frontend` Synthèse, SVG nav, deferred mounting and last-audit strip;
  `book/content` and `book/figures` (topology, current UI captures); `PRODUCT.md`
  rewritten (gitignored); docs, plan index, ledgers.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64), 2026-10-29:
  `python tools/verify.py` → `6 passed, 0 failed, 0 not run` (after the
  environment fixes below);
  `cd web/frontend && npm test` → 41 files, 161 tests passed;
  `npm run build` → typecheck + Vite + bundle, exit 0, 342.3 KB gz;
  `go test ./... -count=1` → every package `ok`, including the new model,
  doctor, campagne, audit, qdisc and cmd tests;
  `bash -n kit/install-bbr3.sh` → exit 0;
  `GOOS=linux GOARCH=amd64 go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...`
  → exit 0;
  `python book/tools/audit-archive.py` → exit 0, both `OK` lines;
  `python book/tools/audit-biblio.py` → 47/47 cited;
  `node book/build-docx.js` → `book/memoire-rebuild.docx` (2793 Ko);
  `memoire.docx` SHA-256 unchanged (`bd171227…6a2e030`);
  book body 15,042 words, total 16,752 (ceilings 15,300/17,000);
  `book/figures/figR1-topologie.png` now shows the four-namespace dual-ns
  topology with the management veth and IFB ingress; Annex E uses fresh
  1920×1080 captures of Campagne, Live replay, Résultats, Comparaison and
  Provenance, and the provenance capture shows 159 runs / 208 valid / 74
  degraded / 246 quarantined.
- Environment failures, all recorded: C: at 99 % plus a cold Go cache made the
  `exit` gate time out (E7); IPv4 ICMP loopback is blocked on this host while
  external IPv4 and `::1` answer, so `TestPingLiveLoopback` now uses `::1`
  (E6); the E5-family Windows temp-cleanup flake recurred in
  `TestMatrixFilteredSubmatrix`, and that test plus the new bbr3 test now
  `defer stopWait(t, m)` — 10× stress pass (E5 update).
- Live test: **not possible**. `ssh Airis@100.72.251.32` timed out;
  `tailscale status` could not reach `tailscaled`; the PC2 runbook is a
  Windows/WSL Graylog lab, not the Linux 6.8 CGO bench. No live audit,
  campaign, IFB or BBRv3 run was attempted. The `PC2 / nemo live bench access`
  row is `BLOCKED`.
- Unproven: Linux 6.8 live acceptance for the bench, IFB ingress, interface
  sampler, `.dist` and BBRv3; production access and field audit; eBPF ground
  truth. The engines are Windows-green, not live-verified. Independent
  verifier pending for slices 8 and 14 and both engine rows.
- Next: restore the tailnet and PC2 or use the Linux 6.8 bench; then run the
  `.dist`/BBRv3 Task 8 acceptance. Human review of the rendered book and the
  Synthèse screenshots remains the decision point.

## 2026-10-29 16:06 UTC · Linux 6.8 live attempt and eBPF build preparation

- Branch/HEAD: `rebuild/plan1-economics-audit` / `0bebb9f` plus the uncommitted
  eBPF source/build-target change.
- Live attempt: Tailscale recovered and PC2 `nemo` (`100.72.251.32`) answered
  ping and SSH. The `CGO-DEV` VirtualBox VM (Ubuntu, 3 GB, 1 vCPU, SSH
  2222→22) started headless; its SSH port opened, but every recorded user/key
  pair was refused (`altfloat`, `cgo`, `ubuntu`, `meteolink`, `airis`, `root`
  × `id_ed25519`, `nemo_lab`, `nemo_ed25519`). No password is recorded. The VM
  was powered off again and verified `poweroff`. PC2's WSL is Ubuntu 26.04 on
  kernel `6.18.54.1-microsoft-standard-WSL2`, exposes only `reno cubic`, and
  lacks `tc`, `iperf3`, `clang`, `bpftool` and `jq`; it cannot certify the
  Linux 6.8 path. No live audit, campaign, IFB, `.dist` or BBRv3 run occurred.
- eBPF preparation: `pkg/probe/bpf/bpf_probe.c` (four programs, two maps),
  `pkg/probe/bpf/BUILD.md`, the `vmlinux.h` ignore rule and the
  `make bpf`/`make bpf-check` maintainer targets are landed. `make` is absent
  on this Windows host, so the target was not run. No object, hash, loader or
  test ships; the slice stays **Blocked** until `make bpf` runs on a stable
  Linux 6.8 host. No object or digest was fabricated.
- Proven: `python tools/verify.py` → `6 passed, 0 failed, 0 not run` at
  `0bebb9f`; frontend 161 tests and build green; eBPF C source is a maintainer
  artefact only, not compiled by the Go build.
- Unproven: the eBPF object and attach, Linux 6.8 live acceptance for the bench,
  IFB, `.dist` and BBRv3, production access and field audit. The live blocker is
  now the missing VM credential/key, not the tailnet.
- Next: install a known key in CGO-DEV through its console (or supply the
  `altfloat` password), then `make bpf`, `make bpf-check`, the opt-in eBPF
  integration test and the Task 8 acceptance. Alternatively provide a stable
  Linux 6.8 host.

## 2026-10-30 09:52 UTC · Interface v2 shipped — four modes, root audit session, linear bench, Option A title

- Branch/HEAD: `rebuild/plan1-economics-audit` / `7303ede`. Slice 15 commits:
  `66e0877` (pure libs: audit summary, operator language, ribbon, bench
  stages), `8f9af74` (audit session at the app root + Auditer view),
  `7b79a30` (four-mode shell, dashboard, linear bench, evidence reports,
  operator language), `7303ede` (polish, lazy bench chunk, shared
  headers/summaries, accessibility).
- What changed: `PanelChooser`, `Rail`, `WebGLMesh`, `SyntheseView` and
  `CampagneView` are gone; the navigation is one `ModeRail` with keys 1–4 and a
  `StatusRibbon` fed once by `/api/health`, `/api/integrity` and the audit
  store. The audit session lives in `store/audit.ts`, so switching destination
  keeps polling and never prints “suivi interrompu” unless the backend stops.
  The bench is one page (Configurer → Observer → Résultats) with shared
  campaign state in `store/bench.ts`. The dashboard separates “Mesuré sur le
  lien réel, sans modification” from “Simulation de remédiation sur profil P2
  — pas une mesure du lien”, adds cost, a local previous-audit comparison, the
  archived proof and a first-run path. Statistics moved to
  `Détails scientifiques` drawers. The Option A title and subtitle were applied
  verbatim to `PRODUCT.md`, the title page, the cover XML and the document
  metadata, and Annex E plus Table 13 now describe the four modes.
- Proven on DESKTOP-B6K9UCM: `npm run typecheck` clean; `npm test` → 46 files,
  202 tests; `npm run build` → 345.4 KB gz total (111.2 KB gz initial + 234.3 KB
  gz lazy bench chunk); `node perf-probe.mjs` → shell 1,422 ms, DCL 612 ms,
  LCP 1,536 ms, bench reveal 890 ms, 0 console errors (baseline before the
  split: 2,171 / 1,264 / 2,208 ms, one 344.2 KB gz chunk); `go test ./...
  -count=1` PASS; `python tools/verify.py` → 6 passed, 0 failed, 0 not run;
  book `audit-archive`, `audit-biblio`, `audit-renvois`, `audit-nombres` and
  `build-docx` green with `memoire.docx` unchanged; impeccable detector clean
  after two inline waivers for values `DESIGN.md` pins.
- Unproven: no live audit, campaign, IFB, `.dist`, BBRv3 or eBPF run; the
  screenshots and the four rendered modes still need the human’s review; an
  independent verifier has not re-run these commands; `book/` and `PRODUCT.md`
  remain uncommitted by design.
- Next: the verifier re-runs the four commands above; the human reviews the
  modes; the long-standing blockers (Linux 6.8 bench, VM credentials,
  production read-only access) are unchanged.

## 2026-10-30 10:49 UTC · slice 15 Task 11 — quality of life and runtime QA on the served build

- Branch/HEAD: `rebuild/plan1-economics-audit` / `43f2066`.
- Changed: `pkg/api/server.go` gains `auditReportExists` and a `has_report`
  field on `GET /api/audit/list`, so `AuditReportsPanel` marks a row
  `non archivé` from the server's word instead of learning it from a failed
  click; `BenchView` opens the results section inside `startTransition`;
  `web/frontend/qa-walk.mjs` is committed for the first time and now attributes
  every long task to the gesture that produced it; `useModeActive` replaces four
  duplicated mode selectors; `BenchConfig` disarms DÉMARRER with no profile and
  says why; `AuditDeclare` states the phase plan of the chosen duration.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64): `npm run typecheck` exit 0;
  `npm test` → 46 files, 209 tests passed; `npm run build` → 111.5 KB gz initial
  + 234.2 KB gz lazy bench = 345.6 KB gz total; `go build ./... && go test
  ./... -count=1` exit 0 with two new API tests
  (`TestAuditReportExists`, `TestAuditListMarksArchivableReports`); `python
  tools/verify.py` → 6 passed, 0 failed, 0 not run; the pre-commit hook re-ran
  three gates at `43f2066` → 3 passed. Runtime against the served build
  (`--serve --addr 127.0.0.1:9091 --mode observe`, embedded fresh bundle):
  `node qa-walk.mjs` → 19 steps, 0 failures, 0 console errors, 0 resource
  errors, 10 long tasks, peak 276 ms (boot, before the first mark), bench entry
  87 ms, results mount 184–246 ms across runs; `node perf-probe.mjs` → shell
  1,050 ms, DCL 437 ms, LCP 1,296 ms, bench reveal 551 ms, 0 console errors,
  heap 8.3 MB (previous pass: 1,422 / 612 / 1,536 / 890 ms).
  `GET /api/audit/list` → 9 rows, 6 `has_report=true`, exactly the six
  `data/audits/<id>/audit_report.json` folders on disk; the three rows that
  precede folder archiving now read `non archivé` before any click.
  Screenshots recaptured: four modes and three bench sections at 1920, two at
  1366, two at 390 in `web/frontend/shots-fresh/`.
- Unproven: no live audit or campaign run — archived corpus and fixtures only;
  the results mount is still a single unrecoverable React unit (splitting the
  ECharts subtree out needs its init effect to re-run on a deferred mount, a
  separate change); long-task figures vary run to run (peak 184–276 ms), so the
  QA gate is a 400 ms budget, not a point estimate; the independent verifier has
  not re-run these commands; `book/`, `PRODUCT.md`, `DESIGN.md` and
  `web/frontend/dist/index.html` stay uncommitted by design.
- Next: the verifier re-runs `npm test`, `npm run build`, `go test ./...`,
  `python tools/verify.py`, `node qa-walk.mjs` and `node perf-probe.mjs`; the
  human reviews the four rendered modes; the Linux 6.8 bench, VM credentials and
  production read-only access remain blocked as listed in `docs/VERIFY.md`.

## 2026-10-30 11:20 UTC · slice 15 Task 12 — first chunk and the cost of opening results

- Branch/HEAD: `rebuild/plan1-economics-audit` / `0c9ccbf`.
- Changed: `AuditView` and `EvidenceView` join `BenchView` behind `lazy()`, each
  section keeping its own `Suspense` boundary; `lib/preload.ts` becomes a mode
  table that `ModeRail` (hover/focus) and `store/ui.setMode` (keyboard, palette,
  click) both use; `ResultatsView` mounts the TradeSpace one paint after the
  ranking behind a height-reserving placeholder and lists the mount flag in the
  chart effect; `qa-walk.mjs` gains a canvas assertion and `benchLayout.test.ts`
  a lazy-destination gate.
- Proven on DESKTOP-B6K9UCM (Windows 10, MINGW64): `npm run typecheck` exit 0;
  `npm test` → 46 files, 209 tests; `npm run build` → initial chunk 99.1 KB gz
  (306.5 KB raw, was 111.5 KB gz / 350.1 KB raw), `AuditView` 4.7 KB gz,
  `EvidenceView` 6.4 KB gz, four shared UI chunks, total 349.9 KB gz against the
  600 KB budget; `go test ./pkg/api/ -count=1` ok; `python tools/verify.py` →
  6 passed, 0 failed, 0 not run. Runtime on the served build: `node qa-walk.mjs`
  → 20 steps, 0 failures, 0 console errors, 0 resource errors, TradeSpace canvas
  880×320 px, results-anchor task 163–232 ms over six runs (was 184–463 ms),
  bench entry 52–97 ms; `node perf-probe.mjs` → heap 7.6 MB (was 8.3 MB), five
  split runs median shell 962 / DCL 378 / LCP 1,180 ms against the earlier
  single-sample pass 1,050 / 437 / 1,296 ms.
- Unproven: this host spreads ±60% run to run — one split run measured shell
  1,515 ms and bench reveal 1,705 ms — so the timings are reported as medians
  with their spread, and only the chunk and heap figures are claimed; the boot
  task (210–310 ms, React init and first render) is the new peak and is not
  addressed here; the four extra chunk requests are unmeasured on a real link,
  only on localhost; no live audit or campaign; independent verifier pending.
- Next: the verifier re-runs the six commands above; the human reviews the four
  rendered modes; the Linux 6.8 bench, VM credentials and production read-only
  access remain blocked as listed in `docs/VERIFY.md`.

## 2026-10-30 12:10 UTC · slice 15 Task 13 — no control that can only dead-end

- Branch/HEAD: `rebuild/plan1-economics-audit` / `69f8f37`.
- Changed: a run with zero archived lines no longer offers `Rejouer` (23 of 179
  were in that state) and says `non rejouable` instead; `.data-missing` is now
  one shared style for both the reports table (`non archivé`) and the runs
  table, replacing the reports-only class; `benchLayout.test.ts` pins both
  gates and `qa-walk.mjs` checks the arithmetic on the visible page.
- Proven on DESKTOP-B6K9UCM: `npm run typecheck` exit 0; `npm test` → 46 files,
  210 tests; `python tools/verify.py` → 6 passed, 0 failed, 0 not run;
  `node qa-walk.mjs` on the served build → 21 steps, 0 failures, 0 console
  errors, 0 resource errors — 7 replayable + 5 marked = 12 visible run rows,
  6 openable + 3 marked = 9 report rows.
- Unproven: `/api/replay/list` currently covers every run row, so the row-count
  gate is the operative guard rather than file existence; no live run; the
  independent verifier has not re-run these commands.
- Next: impeccable audit pass on the four modes, then the verifier pass and the
  human's visual review.

## 2026-10-30 19:10 UTC · slice 15 Task 14 — impeccable audit and the mobile pass

- Branch/HEAD: `rebuild/plan1-economics-audit` / `1668941`.
- Changed: impeccable audit of the four modes (detector 0 findings on sources;
  URL scan unavailable — puppeteer absent, no dependency added). Two fixes from
  the audit: at 390 px the ribbon drops the corpus counts, the seal and its
  separator (still live in Preuves) and `.btn`/`.c-input` rise from 40 to 44 px;
  `qa-walk.mjs` asserts both at 390 px after the desktop long-task collection.
- Proven on DESKTOP-B6K9UCM: `npm test` → 46 files, 210 tests;
  `python tools/verify.py` → 6 passed, 0 failed, 0 not run; `node qa-walk.mjs`
  → 23 steps, 0 failures, 0 console errors, 0 resource errors (44 px measured,
  corpus hidden at 390 px); screenshots recaptured at 1920/1366/390 and the
  390 dashboard checked by eye — header two rows instead of three.
  Audit scores: a11y 3/4, performance 4/4, theming 4/4, responsive 3/4,
  integrity 4/4 → 18/20.
- Unproven: the audit is the builder's own pass, not an independent reviewer;
  the ribbon's `role="status"` still wraps its buttons (P3, recorded with the
  reason the `display: contents` fix is worse); no live run; verifier pending.
- Next: the verifier pass and the human's visual review; the Linux 6.8 bench,
  VM credentials and production read-only access remain blocked.

## 2026-10-30 18:38 UTC · slice 16 — CHR proof built, waiting on QEMU

- Branch/HEAD: `rebuild/plan1-economics-audit` / `e14753c`.
- Changed: `pkg/audit` gains the RouterOS parser (both dialects), the
  field-level diff, the sealed CHR proof writer and the report block plus its
  Markdown line; `cmd/cgo-chr` builds the QEMU line for TCG and KVM, drives the
  serial console through an injectable transport and applies the failure
  ladder; `kit/chr/` holds the portable wrapper, the pinned-image manifest, the
  export fixture and the human README; the report card states the validation
  level and never claims the configuration was applied.
- Proven on DESKTOP-B6K9UCM: `go build ./... && go test ./... -count=1` → exit
  0, `cmd/cgo-chr` included; `npm test` → 46 files, 213 tests; `npm run build`
  → 350.1 KB gz; `python tools/verify.py` → 6 passed, 0 failed, 0 not run.
  Replay evidence: a labelled fixture report with `remediation.routeros` and
  the committed export fixture give 12/12 checks, `accepted: true`, sealed
  `proof.json` at level `windows-tcg`; the same export with
  `cake-ack-filter=none` exits 1 and names exactly
  `queue type cgo-cake.cake-ack-filter`.
- Interrupted once: the pre-commit gate failed with "Espace insuffisant sur le
  disque" at 12 MB free. Cleared only reproducible caches (`go clean -cache`,
  `go clean -testcache`, `npm cache clean --force`), re-ran the gate, commit
  landed. No user file was touched; no backup archive, no `data/`, no `book/`.
- Unproven: no CHR boot — `qemu-system-x86_64` is absent from this host and no
  CHR image is cached; the Linux namespace attach is a runbook step, not a
  result; the fixture report is synthetic and is labelled as such; no
  production contact.
- Next: install QEMU and the CHR image (human), run
  `bash kit/chr/setup-mikrotik-chr.sh` against a report that carries a
  RouterOS script, then record the `windows-tcg` proof and move the VERIFY row.

## 2026-10-30 19:10 UTC · slice 17 — moitié Go du slice 4, readiness, runbook, badge

- Branch/HEAD: `rebuild/plan1-economics-audit` / `ec0cf68`.
- Changed: `pkg/probe/bpf` gains the typed contract, typed errors, the object
  behind `-tags ebpf_object` and a pure fallback ladder; `probe.KernelSource`
  injects the host facts; the report carries an additive `kernel_telemetry`
  with its reason; `/api/health` exposes the source and the ribbon shows
  `Télémétrie : …`; `kit/host-readiness.sh` and the acceptance runbook for
  slices 4–7 exist. The runbook's commands were corrected against the real
  CLI (`run` has no sub-command, `--profiles`, ingress via `CGO_SHAPE_MODE`,
  `calibrate --audit/--profile`, installer without `--dry-run`).
- Proven on DESKTOP-B6K9UCM: `go test` on the four touched packages → ok;
  `npm test` → 46 files, 214 tests; `npm run build` → 350.2 KB gz;
  `python tools/verify.py` → 6 passed; runtime `/api/health` →
  `{"source":"tcpinfo","reason":"hôte non-Linux"}` and the page renders
  `Télémétrie : TCP_INFO` with that title; readiness prints n/a rows (exit 0)
  and exits 2 with `--require-linux`.
- Unproven: no object built, nothing attached, no audit ran with the probe —
  `source: ebpf` is untestable here and the ladder says so; the runbook was
  not executed on Linux; `root` approximates CAP_BPF in `kernelFacts`.
- Next: Linux 6.8 host — `make bpf`, tagged build, one audit with the probe,
  then the slice-4 row opens. Slice 18 (verifier packets + methodology
  chapter) is next on this host.

## 2026-10-30 19:40 UTC · slice 18 — paquet vérificateur et chapitre méthodologie

- Branch/HEAD: `rebuild/plan1-economics-audit` / `7d4e747` (docs commit).
- Changed: `docs/verification/2026-10-30-slices-8-14-15.md` (seven commands,
  expected outputs, per-slice evidence, empty signature block) and
  `docs/thesis_trace_driven_methodology.md` (trace-driven `.dist`,
  Gilbert–Elliott, the 6.8 `kfree_skb` detour, the BBRv1 table with the CUBIC
  limitation, the control-point paragraph, and a sources table).
- Proven on DESKTOP-B6K9UCM: `python tools/verify.py --only links` → 1 passed;
  every path named in the index resolves. The chapter cites the frozen corpus
  through `audit-nombres` and the `Tableau 15` limitation row, the tariffs as
  `RECORDED`, the eBPF and CHR rows as `BLOCKED`, and explicitly refuses the
  deployment, BBRv3 and production claims.
- Unproven: the verifier has not run the packet; the chapter is a draft for
  the author's review and has not been pasted into the book; no number was
  re-measured for it.
- Next: slice 19 contract stays blocked on the operator's grant; the human
  reviews the four modes; a Linux 6.8 host unlocks slices 4–7.

## 2026-10-30 20:05 UTC · slice 19 contract — rien ne s'exécute avant le grant

- Branch/HEAD: `rebuild/plan1-economics-audit` / `50e9179` (code) + this docs commit.
- Changed: `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md` fixes
  what the case study would read (operator-supplied dumps only), what is never
  done (no capture, load, write, root, login, scan, query), the five
  deliverables, the acceptance rules (declared/observed/inferred labels,
  pseudonym by default, one unreproducible cell named) and the exact request
  to put to the operator.
- Proven on DESKTOP-B6K9UCM: docs only; `python tools/verify.py --only links`
  → 1 passed. No connection was made to any operator system; no artifact was
  produced.
- Unproven: everything the slice would produce — the slice is BLOCKED on the
  written grant; no hostname, address or topology is recorded anywhere.
- Also in this wave: `cmd/cgo-chr` refuses `--level linux-namespace` off Linux
  with exit 2 and a named reason (test `TestLevelAllowed`), instead of letting
  QEMU fail obscurely and stamp a wrong level.
- Next: the CHR boot (QEMU + image), a Linux 6.8 host for the runbook, the
  verifier's signature on the packet, and the operator's grant.

## 2026-10-30 20:40 UTC · livre — Annexe F et champs REF réparés

- Branch/HEAD: `rebuild/plan1-economics-audit` / `dcce673` (docs) — book changes
  stay in the working tree, never committed.
- Changed: `book/content/09-annexes.js` gains **Annexe F — Télémétrie noyau et
  validation de la configuration RouterOS** (three tiers and the non-silent
  fallback, the 6.8 hook facts — `qdisc_drop` absent, `__qdisc_drop` inlined,
  `kfree_skb` + reason, fentry for exact bytes — the blocked status, and the
  CHR proof at two levels with 12/12 replay and `accepted : false` on refusal).
  `book/build-docx.js` now caches each Word `REF` field with the real number
  read from the content (`REF_NUMS`) instead of a literal `0`.
- Proven on DESKTOP-B6K9UCM: `node book/build-docx.js` → 2888 Ko;
  `memoire.docx` SHA-256 unchanged (`bd171227…`); `audit-archive` both `OK`;
  `audit-nombres` OK; `audit-biblio` 47/47; `audit-sigles` OK;
  `audit-typo` exit 0 (2 cadratins de couverture, reprise verbatim);
  `audit-renvois` → OK, 40 renvois littéraux, one pre-existing “Figure 2”
  warning; zero `Tableau 0` / `Figure 0` left in the rendered text (five
  before), and `Tableau 13` now resolves in Annex E.
- Unproven: Annexe F is in the rebuild only, not in the reference
  `memoire.docx`; no number was added by the annex; the human has not read it;
  the REF cache assumes content numbers stay the source of truth.
- Next: human review of Annexe F; the verifier's signature on the packet; QEMU
  and a Linux 6.8 host for the remaining rows; merge to `main` is the human's.

## 2026-10-30 20:00 UTC · deepwork — API, GUI, livre (trois phases, trois revues)

- Branch/HEAD: `rebuild/plan1-economics-audit` / `91b6ecb`.
- Phase A (`1536f09`): 9 audit routes moved to `pkg/api/server_audit.go`;
  `server.go` 1,899 → 1,495. Oracle gate PASS: route lists identical, 377 body
  lines byte-identical, package state declared once, `handleAuditStart`
  locking/goroutine untouched, `go vet` clean, targeted tests ok.
- Phase B (`17ab85c`): `cmd/cgo-gui` diagnostic composes
  `version · télémétrie eBPF|TCP_INFO|non instrumenté` from `/api/health`.
  Both shells render that string, so the Fyne view gained it without the
  Linux-only file being edited blind. Gate PASS; remediation added
  `version inconnue` for an empty version and two test cases.
- Phase C: Annexe E gained the ribbon paragraph and the validation sentence;
  book rebuilt (2,889 Ko), `memoire.docx` hash unchanged, five audits green.
  The oracle found a blocking overclaim (a concrete `validée (7.15.3, CHR)`
  implying an acceptance): reworded as a format plus the replay caveat. The
  follow-up gate found the same empty-`checks` hole in the Markdown renderer;
  fixed in `pkg/audit/report.go` and the card, each with a test (`91b6ecb`).
- Proven: `python tools/verify.py` → 6 passed; frontend 46 files / 215 tests;
  `go test ./cmd/cgo-gui` ok; `go vet ./pkg/api ./cmd/cgo-gui` clean.
- Non-blocking, recorded for later: “non instrumenté” for a pre-telemetry
  server that may run eBPF (under-claims; vocabulary is prescribed); the Fyne
  open-dashboard line puts two facts in one “(version …)” parenthesis; qa-walk
  asserts only the corpus hidden at 390 px, not the seal; `core.go` carries
  pre-existing gofmt noise left untouched to keep diffs reviewable.
- Book working tree and `web/frontend/dist/index.html` uncommitted by rule;
  the deepwork state lives at `.slim/deepwork/book-arch-gui.md` (gitignored).
- Next: human read of Annexe E/F; verifier signature on the packet; QEMU and a
  Linux 6.8 host for the blocked rows; merge to `main` is the human's.

## 2026-10-30 20:35 UTC · polish — région live du bandeau, et l'assertion qui manquait

- Branch/HEAD: `rebuild/plan1-economics-audit` / `7f99dba` + this pass.
- Changed: `StatusRibbon` no longer carries `role="status"` on the container;
  the running-audit item alone carries `role="status" aria-live="polite"`, so a
  health reload no longer re-announces corpus counts, the seal or the two
  buttons. `qa-walk.mjs` now asserts what its step name claimed: corpus AND
  seal hidden at 390 px (before, only the corpus was checked).
- Proven on DESKTOP-B6K9UCM: `npm run typecheck` clean; `npm test` 46 files /
  215 tests; `npm run build` 350.2 KB gz; served page DOM check —
  container role `null`, `aria-live` count in the ribbon `0` at rest,
  `Télémétrie : TCP_INFO` still rendered; `node qa-walk.mjs` → 23 steps,
  0 failures, 0 console/resource errors, “corpus masqué, sceau masqué”,
  44 px targets, TradeSpace canvas 880×320, peak task 254 ms.
- Unproven: the live region's actual announcement was not heard on a screen
  reader; only the DOM contract was checked. No other a11y surface re-audited.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant; the
  remaining deferred items are recorded (Fyne parenthesis wording,
  pre-telemetry server wording).

## 2026-10-30 20:16 UTC · tests et débogage — le niveau inconnu et les surfaces non parcourues

- Branch/HEAD: `rebuild/plan1-economics-audit` / `7f9550d`.
- Debugging: a new integration test of the CHR replay path found that
  `--level kvm` travelled to `WriteCHRProof` and exited 1 (seal failure)
  instead of 2 (usage). `levelAllowed` now rejects any level outside
  windows-tcg/linux-namespace before the report is read. This is the second
  defect the test-first discipline caught in this surface, after the empty
  `checks` array rendering as a pass.
- Test coverage added (`cmd/cgo-chr/run_test.go`): replay end-to-end (exit 0,
  level, 12/12 checks, sealed hashes, export and transcript files), mutated
  export (exit 1), report without a RouterOS script (exit 2), unknown level
  (exit 2); `TestLevelAllowed` extended with the unknown-level rung.
- Debugging sweep (one-off, then made permanent): settings drawer and command
  palette open and close on Escape; zero horizontal overflow at 1366, 1024,
  1920 and 390 px; evidence report open/back; three bench anchors — 0 console
  errors, 0 page errors. `qa-walk.mjs` now always asserts the six chrome and
  overflow steps: 29 steps, 0 failures, 0 console/resource errors, peak task
  267 ms.
- Proven: `go vet ./cmd/cgo-chr` clean; `go test ./cmd/cgo-chr` ok; frontend
  suite 46 files / 215 tests (previous run, unchanged by this pass);
  `python tools/verify.py` was green at `7f99dba` and the only Go change here
  is the guard plus tests.
- Unproven: the sweep exercises one browser (Edge) on one host; no screen
  reader; the Fyne view still has no runtime test here.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant, and
  the human's visual review.

## 2026-10-30 20:30 UTC · débogage systématique — la palette morte et le tiroir sans focus

- Branch/HEAD: `rebuild/plan1-economics-audit` / `936af6c`.
- Phase 1 (preuves, avant toute correction) : sous `reducedMotion: reduce`,
  tout est visible (opacité 1, 0 erreur) — hypothèse écartée. Sur les deux
  surfaces modales : les réglages laissaient le focus sur la goupille et Tab
  fuyait vers le rail ; la palette laissait `activeElement = BODY`, Tab
  n'entrait jamais. Mesure décisive : `#palette` est un descendant de `#shell`
  (App.tsx:119) et la palette posait `#shell.inert = true` — elle s'inertait
  elle-même. Contre-preuve : `shell.inert = false` puis `input.focus()` →
  `activeElement = INPUT`.
- Phase 2–3 : le tiroir n'avait aucun code de focus ; la palette en avait un,
  faux par construction. Une seule cause racine par surface, chacune testée en
  échec avant correction (`qa-walk` : « la palette prend le focus » False/BODY,
  « le focus reste dans les réglages » False).
- Phase 4 : `useModalFocus` partagé — portail hors de `#shell`, inertie du
  fond, focus dans le panneau, Tab/Shift-Tab piégés, Échap, focus rendu au
  déclencheur ; palette et tiroir l'utilisent. Preuves : parcours 34 étapes,
  0 échec, 0 erreur console, focus INPUT des deux côtés, focus rendu à la
  goupille ; `npm test` 46 fichiers / 215 tests ; build 350.3 KB gz.
- Non couvert : un lecteur d'écran réel ; le tiroir n'a pas de test unitaire
  (le contrat est prouvé au runtime par le parcours).
- Next: règles du repo inchangées — QEMU/Linux, signature du vérificateur,
  grant opérateur, revue humaine.

## 2026-10-30 20:51 UTC · débogage systématique — arrêt d'audit, identifiant de run, et un incident

- Branch/HEAD: `rebuild/plan1-economics-audit` / `5d2c6dd`.
- **Incident, disclosed**: while probing POST validation with supposed
  reject-only bodies, `POST /api/audit/start` with `duration: 0` was accepted —
  0 means “default 30 s”, not “invalid” — so a live audit started against the
  default target on this host. It was stopped immediately (`stop`), the
  cancelled run wrote **no row** (`data/link_audit.csv` ends at
  `audit-1793015258`, 2026-10-26), and state returned to `running:false`.
  Lesson recorded: a reject-only battery may only use values the API provably
  rejects (`direction=sideways`, `flows=99`), never a zero that the schema
  defaults. The frozen corpus (`book/data/runs`) was never involved.
- Defect 1 — `POST /api/audit/stop` answered `{"stopped":true}` twice for two
  rapid requests while `running` still said true: `handleAuditStop` delivered
  `cancel()` without consuming `auditCancel`, which only the goroutine's defer
  clears. Fixed by consuming it under the same lock; test
  `TestAuditStopDeliversCancelOnce` reproduces the state without network.
- Defect 2 — `GET /api/results?run=…` passed the id into a glob
  (`data/runs/<run>/aqm_eval.csv`) with no traversal guard, while replay/stream,
  run/rows and quarantine all rejected the same input with 400. One
  `validRunID` rule now guards all four; test `TestResultsRejectsUnsafeRun`
  requires 400 on four hostile ids and 200 on a safe unknown run.
- Proven on DESKTOP-B6K9UCM: `go test ./pkg/api -count=1` → ok 63.6 s;
  `python tools/verify.py` → 6 passed, 0 failed, 0 not run; live server after
  restart — hostile run 400, plain run 200, idle stop `{"stopped":false}`,
  status `running:false`.
- Unproven: the stop fix was proven by unit state, not by a second real live
  audit (none will be run without approval); the quarantine empty-id case
  keeps its old behaviour (`!validRunID("")` is false, so empty is allowed).
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant,
  human visual review.

## 2026-10-30 21:04 UTC · UX — l'absence ne pèse plus comme une mesure

- Branch/HEAD: `rebuild/plan1-economics-audit` / `34090b6`.
- Changed: `OperatorStats` tags each cell without a value (`data-empty`) and
  the stylesheet gives those cells a compact row under 900 px: on a phone the
  two unmeasured facts (`alertes à l'heure`, `qualité du lien`) took ~90 px
  each with a large dash — more space than the one measured fact. Measured
  values keep the numeral; absences read as one line. Desktop grid unchanged
  apart from a lighter padding on empty cells.
- Proven on DESKTOP-B6K9UCM: measured heights in the browser — 2 empty cells
  at 24 px on 390 px, 0 at 1920; SSR test (3 empty without a report, 0 with a
  measured one, the first fixture was wrong because it omitted the CAP SLA and
  was fixed, not the component); `npm test` 47 files / 217 tests;
  `npm run build` 350.3 KB gz; screenshots recaptured at 1920/1366/390.
- Long-task gate: one walk reported 520 ms (budget 400). Two immediate re-runs
  measured 194 ms and 212 ms with the same per-step profile, right after a
  90 s test suite — recorded as host contention, not a code regression, and
  the threshold was NOT relaxed. A single breach must be re-run before being
  believed; this occurrence is the evidence for that rule.
- Unproven: no human visual review of the new mobile fold; no screen reader.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-30 21:14 UTC · UX + performance — la bande des absences, et le CLS mesuré

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Changed: on the dashboard, when neither the cost nor the change comparison
  carries a value, the two full-height cards collapse into one band that names
  both absences with their reasons (`data-testid="dash-quiet"`); as soon as one
  of the two exists, the grid returns. `bothAbsent` is a pure predicate with a
  unit test. `perf-probe.mjs` now observes `layout-shift` (excluding recent
  input) and publishes `cls` — layout stability was the one vital the project
  never measured.
- Proven on DESKTOP-B6K9UCM: served build — quiet band present, empty two-card
  grid absent (1920 screenshot reviewed); `perf-probe` → CLS 0.004, LCP 1152 ms,
  DCL 427 ms, shell 888 ms, bench 465 ms, 0 console errors; `qa-walk` 34 steps,
  0 failures, max long task 198 ms; `npm test` 47 files / 218 tests;
  `npm run build` 350.4 KB gz.
- Unproven: no Chrome DevTools full trace (MCP not connected) — LCP/CLS come
  from PerformanceObserver in this probe, not from a trace; INP is proxied by
  long tasks only; no screen reader; no human review of the desktop fold yet.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 00:26 UTC · débogage — une hypothèse réfutée, et l'invariant qui la remplace

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Lead: selecting one of the 23 zero-line archived runs looked like a blank
  results panel (my probe reported `rows: 0`, `emptyState: 0`).
- Phase 1: the probe searched `.empty-state`/`data-testid="empty-state"`, but
  `EmptyState` renders `.c-empty` (`web/frontend/src/components/ui/EmptyState.tsx:9`).
  Re-probe with the real class: one `.c-empty` with
  “run vide ou supprimé — passez à tous runs”, and the reset control visible.
  **No product defect** — the defect was in the measurement.
- Action: the walk now asserts the real invariant, dataset-independent — after
  choosing a run, the panel shows either rows (`.lb-card`) or an explicit empty
  state (`.c-empty`), never a mute zone. 35 steps, 0 failures, 0 console or
  resource errors, peak long task 204 ms.
- Lesson recorded: a selector written from memory is a hypothesis; read the
  component before believing an absence it reports.
- Unproven: the invariant walk happened to pick a run with data; the empty
  branch is proven by the corrected probe today, not continuously.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 00:57 UTC · interface entière relue — mesure du formulaire, copie des phases, et la sonde de bruit

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Read of every destination (screenshots at 1920 and 390 plus source): the
  audit form rendered a short site name and a three-value select across
  ~1,050 px while duration and direction stayed compact. It is now a 760 px
  grid of two 374 px columns (one 332 px column at 390, measured in the DOM).
  First attempt silently failed because a duplicate `.audit-form { display:
  flex }` sat later in the stylesheet — the stale rule was removed, not
  overridden with `!important`; the tokens test then caught that the new
  media query had been inserted before the phone block and was folded into it.
- Native checkboxes took the browser blue in a monochrome world: one
  `accent-color` rule pins them to the living cyan token.
- The live banner said “Idle” and “Baseline” in an otherwise French operator
  surface: now “Au repos” and “Repos (référence)”, and the header/footer phase
  fallbacks say “au repos”. The colour comparison follows the new string.
- Proven: suite 47 files / 218 tests; build 350.4 KB gz; walk 35 steps, 0
  failures in the settled runs (211/221/230 ms), 0 console/resource errors.
- Long-task flake, measured rather than excused: two walks right after the
  90 s suite breached at 509 and 489 ms (same step “ancre résultats”), then
  five runs measured 215/211/221/230 ms with no breach. The 400 ms gate is
  kept because it is what caught the 463 ms results-mount regression; the
  recorded rule stands — re-run a single breach before believing it.
- Unproven: no human review of the new form grid at 1366; no screen reader;
  the Fyne view still cannot be compiled here.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 02:20 UTC · deepwork — stress des 19 slices, et une correction réfutée

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Inventory: `tools/slice_matrix.json` (19 slices; 4–7 and 19 unrunnable with the
  ledger's blocker text; 16/17 partial). Review 1 blocked the first version:
  marker-only checks could pass a failing package, and slice 1 was narrower
  than its plan. Remediation added the runner contract (exit 0 AND every
  marker), widened slice 1 to `./pkg/probe/...`, removed `audit-biblio` (prints
  OK before errors) and moved verify/build to `global_checks`.
- Stress run (`tools/slice_audit.py`, sequential, per-check timeout):
  **26 passed, 0 failed, 5 skipped, 1644 s**. A first background dispatch lost
  its output to a dead pipe and was killed; the harness later reported exit 0
  after its buffered log flushed, which the runner reconciled before trusting.
- Probe defects, found by reading, not by tests (no compiler here):
  1. Map full: retransmitted bytes vanished from every total. Fixed — the
     branch now adds payload/1 to the per-CPU globals; the exactly-once
     accounting was reviewed clean; `Globals` gained the retrans fields so
     those bytes can land in Go.
  2. A “fix” subtracting the TCP header from `skb->len` was **refuted by the
     oracle with kernel-source evidence**: at the tracepoint the queued skb is
     headerless and `skb->len` is already payload (`tp->bytes_retrans +=
     skb->len`), while `transport_header` still holds the 0xFFFF sentinel — the
     subtraction would read out of bounds. Reverted; the comment now carries
     the evidence; the vacuous source-contract test was deleted rather than
     kept as false comfort.
- Proven: `go test ./pkg/probe/...` ok; six gates green in the audit run
  (global check) and at this commit's hook.
- Unproven: no `make bpf` — the object still requires Linux 6.8; the map-full
  branch and the restored payload line are reviewed, not compiled.
- Next: run `make bpf` on the host when it exists; P4 ledger update lives in
  the slice audit's own record above.

## 2026-10-31 02:40 UTC · interface distillée — profils sur une ligne, lecture bornée

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Research input: NN/g on minimizing cognitive load and form clarity
  (structure, transparency, clarity, support) — applied as cuts, not a
  redesign.
- Changed: `.bench-profile` renders id and description on one baseline row
  (small 13 px per line, list 50 px tall — it stacked three lines before in a
  narrow column); prose in cards (`.syn-read`, `.dash-disclaimer`,
  `.audit-verdict`, `.dash-steps span`) is clamped to 68ch (550 px at 1920),
  matching the craft floor's 65–75ch and NN/g's clutter guidance.
- Proven: measurements in the served DOM; `npm test` 47 files / 218 tests;
  build 350.4 KB gz; walk 35 steps — 4 runs: 175 and 222 ms clean, 495 and
  529 ms breaching with no step attribution; the 400 ms gate stays and the
  breaches are recorded as host variance, not waived.
- Unproven: no human visual review of the configurer after the change; the
  long-task variance remains unexplained beyond host load.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 05:48 UTC · coupe plus dure — la preuve d'abord, l'attente sans écho

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Changed: the provenance preamble on the evidence destination is now a
  collapsed `<details>` (summary keeps the heading, ▸/▾ marker), so the page
  opens on the proof band — measured in the served DOM: card closed, proof
  band at the top. The bench idle state said “EN ATTENTE — en attente — …”;
  the hint now names only the gesture (“démarrez une campagne dans Configurer”).
- Proven: walk 35 steps, 0 failures, 0 console/resource errors, peak 217 ms;
  suite 47 files / 218 tests; build 350.5 KB gz.
- Unproven: the collapsed summary's discoverability was not tested with a
  human; no screen-reader pass.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 06:05 UTC · coupe plus dure — quatre raccourcis, quatre mots

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Changed: the dashboard shortcut cards drop their sub-lines (“sans droits
  admin, rapport scellé”, “matrice P1–P4, 4 flux”, …), which restated what the
  destination already says; icon + title remain, the four shortcuts of slice 14
  are intact, and the row halves in height. The `sub` field left the data type
  so no dead string survives.
- Proven: walk 35 steps, 0 failures, 0 console/resource errors, peak 277 ms;
  47 files / 218 tests; 350.4 KB gz.
- Hard stops recorded for the next harsh pass: the shortcut row itself is
  spec-pinned (removing it needs the human), and the archived-comparison card
  cannot go inside a closed `<details>` — a canvas mounted at zero size does
  not resize when the disclosure opens (known ECharts failure), so that cut
  needs a mount-aware fix first, not a CSS trick.
- Unproven: no human read of the shortened dashboard.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 06:18 UTC · coupe plus dure — les sous-titres cessent de réciter la page

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Changed: destination subtitles shortened to one line each and reduced to
  their unique fact — Auditer keeps “sans droits administrateur · aucun CSV
  partiel” (the report tag already states “mesuré, sans modification”),
  Banc “une page : configurer, observer, lire” (the three section titles say
  it), Preuves lists its three proof objects. No claim removed.
- Proven: 47 files / 218 tests; build 350.3 KB gz.
- Stopped here on purpose: the remaining harsh cuts are the spec-pinned
  shortcut row (human scope) and the archived-comparison card behind a
  mount-aware disclosure (engineer task: canvas must not mount at zero size).
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 06:30 UTC · a11y — les retours d'action sont annoncés

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Checked before cutting: the big Observer card is the edge-control panel
  (Surveillance, Façonnage, Essai), not a pure archived comparison — hiding it
  behind a disclosure would hide controls. The cut was dropped, and the reason
  is recorded instead of forced.
- Changed: the audit declaration message and the shaping result line carry
  role=status / aria-live=polite; the ribbon keeps no live region (only the
  running-audit item does). 47 files / 218 tests; build 350.3 KB gz; typecheck
  clean.
- Unproven: no screen-reader pass; the announcement wording was not heard.
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant, and a
  human scanning pass before further layout cuts.

## 2026-10-31 08:00 UTC · typographie +1 et graphes bornés

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Changed: the type scale moves up one step (micro 11, tag 12, cap 13, body 14,
  head 22) — measured at 390 px: the headline value renders at 22 px; the walk
  shows zero horizontal overflow at 390/1024/1366 and intact 44 px targets.
  ECharts: `devicePixelRatio` capped at 2 in ResultatsView and CompareView
  (LiveView already had it), and a new source gate pins canvas + dirty-rect +
  DPR cap + `animation: false` for every chart.
- Proven: walk 35 steps, 0 failures, 0 console/resource errors, peak 233 ms;
  48 files / 220 tests; build 350.3 KB gz.
- Unproven: no human review of the new scale at 1366; ECharts memory win is
  reasoned from DPR scaling, not profiled (no heap comparison per chart).
- Next: unchanged — QEMU/Linux host, verifier signature, operator grant.

## 2026-10-31 08:25 UTC · phase propre — l'ancre atterrit et emmène le focus

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- New stance, first increment: navigation must land. Anchor jumps put section
  headings under the ~71 px top bar (only `.bench-section` had a 56 px margin)
  and left keyboard focus in the rail. Targets now carry `scroll-margin-top:
  72px`; `scrollToId` gives the target `tabindex="-1"` once and focuses it
  without a second scroll.
- Proven: walk 36 steps (new gate: focus inside `#bench-results` after the
  anchor click), 0 failures, 0 console/resource errors, peak 220 ms; 48 files
  pass in the suite (run before the commit, unchanged files); build 350.4 KB gz.
- Unproven: focus ring on the section container was not visually reviewed; no
  screen reader.
- Next: continue the clean phase on the two remaining structural candidates —
  the spec-pinned shortcut row (human scope) and chart disclosures (mount-aware
  work) — otherwise unchanged blockers.

## 2026-10-31 09:05 UTC · deepwork — anneau eBPF et ingress × CHR

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Lane A: `KernelTelemetry` gains additive flows fields; `KernelRing` renders
  ≤12 proportional slots with an aria label, only for `source=ebpf` with flows;
  mounted in the report card. Review found the blocking **producer gap**: no
  code populates Flows until slice 4's loader lands — the ring is
  fixture-testable only, and this is recorded rather than papered over. Fixed
  from review: slot overflow (flex-grow, not percentage widths) and the missing
  role/label.
- Lane B: ingress export fixture + replay tests — 11/11 checks including the
  LAN parent; the mutated parent fails with the faulty property named in the
  proof (assertion moved to the right test after a mis-anchor was caught by the
  package run).
- Proven: `go test ./cmd/cgo-chr` ok, `go vet` clean, gofmt clean; KernelRing
  suite 3/3; typecheck clean; build 350.7 KB gz.
- Unproven: no producer for flows; no CHR boot; no eBPF attach — live rows stay
  BLOCKED. `make bpf` remains the first Linux step.

## 2026-10-31 09:25 UTC · fiche terrain, et le livre remis à jour

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Field sheet: `@media print` hides the shell (rail, ribbon, footer, buttons,
  quick actions), forces black on white, keeps cards unbroken, sets a page
  margin — the operator carries the report to the antenna. Source gate
  `src/lib/print.test.ts` because no screenshot shows print output.
- Book (working tree, never committed): Annexe E gained the type scale,
  absent-fact band, anchor-focus and field-sheet paragraphs; Annexe F gained
  the ingress-through-CHR result (11/11, faulty parent named), the kernel ring
  with its honest no-producer gap, and the two probe corrections (map-full
  accounting; the refuted TCP-header subtraction). Rebuild 2,889 Ko;
  `memoire.docx` hash unchanged; archive, nombres, renvois, sigles, biblio all
  green. The number audit refused “1024/1366” and demanded “1 024 / 1 366” —
  fixed at the source, not waived.
- Rule adopted: the book is touched in the same pass as any interface or
  measurement change, and rebuilt before the commit that records it.
- Unproven: no printed page was produced on a real printer/PDF; no human read.

## 2026-10-31 10:05 UTC · nouvelle phase — l'identité de la thèse

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit.
- Critical finding: the title claimed three undefendable things — “garantie
  d’échéances” (corpus: 60.2 % deadlines met under the simple queue),
  “application au système d’information” (case study BLOCKED, nothing read),
  and BBR adjacent to a guarantee (corpus holds BBRv1 only). IDEA1–3 and
  Untitled 37 overclaim further (“sécurisation”, “priorisation”, one-click).
- Decision (spec `2026-10-31-cgo-thesis-identity-design.md`): main title takes
  the master prompt’s own wording — « Conception et évaluation d’un banc
  d’essai reproductible pour la comparaison de politiques AQM et BBR sur des
  liens d’accès contraints »; subtitle demotes the institution to case — « Le
  cas des liens d’accès du service national de prévision météorologique de
  Madagascar ».
- Applied (working tree): PRODUCT.md, title page, cover XML (2+2 occurrences),
  abstract numbers corrected from a transcription drift (324,3/210,3) to the
  corpus-rendered values (328,2/210,5 ms). Rebuild 2,889 Ko; `memoire.docx`
  SHA-256 unchanged (`bd171227…`); archive, nombres, renvois, sigles, biblio
  green; the new title is in `word/document.xml` and the old one is absent.
- Unproven: the introduction's one-sentence project definition was NOT added —
  placing it needs a careful read of the intro module, and half-placed prose is
  worse than none; it is named in the spec as the remaining step. No human read
  of the new title page.
- Next: add the project-definition sentence, then the human reviews the title
  page; verifier and merge decisions unchanged.

## 2026-10-31 10:20 UTC · « no AI » au sens AIR/DIS — la traçabilité écrite dans le mémoire

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit (book working tree).
- Rule applied, from the DIS guidance (`C:/DISanalyzer-logs/research/ai-assisted-log-analysis.md`): every AI claim traces to a line of evidence or it is labelled unverified. The project already enforces it in the ledgers; the thesis now states it: a “Traçabilité” paragraph in the annexes says each number carries a VERIFY row or « non mesuré », no assisted phrase introduced a figure, an unopenable source is called inaccessible rather than paraphrased, and blocked passages name their unblock condition.
- Lexicon scan (révolutionnaire, sans effort, en un clic, puissant, robuste, seamless, powered by, intelligent, unlock, solution miracle) over `book/content`, `PRODUCT.md`, frontend sources: 5 hits, all legitimate technical uses (“robustesse” in a requirements sentence; three code comments). No marketing or AI-tic pattern to remove.
- Proven: rebuild 2 890 Ko; `memoire.docx` SHA-256 unchanged (`bd171227…`); archive, nombres, renvois, sigles, biblio green; note present in `word/document.xml`.
- Unproven: no external hallucination evaluator was run (the DIS note mentions HEM as an optional gate); the check is rule + scan, not a model audit.
- Next: the intro project-definition sentence (named in the identity spec), the human read, then the verifier.

## 2026-10-31 10:40 UTC · coquille distillée — le bandeau ne décode plus rien

- Branch/HEAD: `rebuild/plan1-economics-audit` / this pair of commits.
- Changed: `StatusRibbon` drops the corpus counts, the seal hash and the
  telemetry tier (all remain in Preuves and the report), drops the integrity
  fetch entirely — one health request instead of two — and keeps mode,
  connection, running audit, last audit and VÉRIFIER. Walk asserts zero
  corpus/seal nodes in the ribbon.
- Test realignment: `title.test.ts` asserted the old “garantie” title and failed
  after the identity decision; it now pins the bench title and forbids the old
  phrase in both PRODUCT.md and the title page.
- Proven: walk 37 steps, 0 failures, 0 console/resource errors, peak 235 ms;
  50 files / 224 tests; build 350.4 KB gz.
- Unproven: no human read of the simplified chrome.

## 2026-10-31 11:00 UTC · anti-slop selon AIR — trois paragraphes remis en rythme

- Branch/HEAD: `rebuild/plan1-economics-audit` / this commit (book working tree).
- Standard read: `C:/AIR/40 - Workflow/Agent Guidelines.md` §2 and §4 — corrective
  framing (“not X but Y”, “rather than”) runs 100× the human rate; a paragraph
  with no short sentence is a defect; the banned decoration list; the “only the
  writer could know” test. Skill `no-ai-slop` applied as detect-then-edit.
- Detect: French corrective framing absent (0 hits of plutôt que / non pas /
  au lieu de across book and docs); one English “rather than”; 2 em-dashes in
  the front module. Rhythm: 6 paragraphs over 55 words with no short sentence in
  the annexes, worst 173 words (mine, on ingress and the ring).
- Edit: the three paragraphs I authored (ingress/ring, type scale, traçabilité)
  are split into short sentences with at least one line under eight words, every
  figure and every non-claim kept verbatim. Three long paragraphs from earlier
  authorship remain and are recorded rather than rewritten blind.
- Proven: rebuild 2 889 Ko; `memoire.docx` SHA-256 unchanged; archive, nombres,
  renvois, sigles, biblio green.
- Unproven: the rhythm check is a heuristic (sentence split on punctuation);
  no human read of the rewritten annex.
