# Verification gates as a standalone tool — primary-source research

*Date: 2026-11-07 · Scope: `tools/verify.py` (six gates) + `.githooks/pre-commit` (+`pre-push` for context) + `docs/VERIFY.md` gate ledger rows, read as if shipped tomorrow as an independent tool `cgo-verify`. Research only — no code changed, no commit.*

Primary sources fetched this pass:

- **(R1)** Git `githooks` — `git-scm.com/docs/githooks` (last updated in 2.54.0): synopsis `$GIT_DIR/hooks/* (or git config core.hooksPath/*)`; DESCRIPTION on executable-bit, `core.hooksPath` override, working-directory change, and per-hook STDIN/exit semantics for `pre-commit` and `pre-push`.
- **(R2)** Python `subprocess` — `docs.python.org/3/library/subprocess.html` (Python 3.15.2 docs): `subprocess.run(args, …, timeout=None, check=False, capture_output=False, …)` returns a `CompletedProcess` with `returncode`; `check=True` raises `CalledProcessError` on non-zero exit; `TimeoutExpired` is raised on timeout; sequence-args avoid shell quoting; `returncode 0` indicates success.
- **(R3)** Python `unittest` — `docs.python.org/3/library/unittest.html` (Python 3.15.2 docs): `TestCase` subclasses with `test_*` methods, `assert*` helpers, `setUp`/`tearDown` fixtures, and `unittest.main()` runner; the framework is the idiomatic stdlib unit-test runner.
- **(R4)** Semantic Versioning 2.0.0 — `semver.org`: MUST declare a public API (§1); normal version is `X.Y.Z` with no leading zeroes (§2); released version contents MUST NOT be modified (§3); PATCH/MINOR/MAJOR bump rules (§§6–8); FAQ “v1.2.3 is not a semantic version” (tag ≠ version).
- **(R5)** SLSA v1.0 *Producing artifacts* — `slsa.dev/spec/v1.0/requirements`: producer MUST choose a build platform capable of the desired level, follow a consistent build process, and distribute provenance; build platform MUST generate provenance that unambiguously identifies the output package by cryptographic digest; L2 requires signed (authentic) provenance; L3 requires unforgeable provenance and isolated, ephemeral builds; the page status is “Retired” in favour of v1.2 (fetched banner) — level semantics below are R5-as-fetched, not current-spec advice.
- **(R6)** Git `git-config` (`core.hooksPath` anchor) — `git-scm.com/docs/git-config` (last updated in 2.56.0): this pass could not isolate the `core.hooksPath` paragraph in the fetched text dump, so any claim that depends on the *config-file syntax/scope* (not the `githooks` DESCRIPTION statement of the override) is marked **`unverified — primary source not found`**.

Any claim without a primary source is marked **`unverified — primary source not found`**. Codebase claims cite `file:line`; external-normative claims cite R1–R6.

---

## 1. Responsibility (what owns what)

| Artifact | Responsibility | Evidence |
|---|---|---|
| `tools/verify.py` | The six-gate runner. Owns the gate registry, the 0/1/2 verdict contract, staged/fast/only/selftest selection, the line-ceiling guard, and all six `gate_*` + `prove_*` implementations. | `tools/verify.py:455-462` registry; `:36` `verdict()`; `:478-527` `main()`; `:11-13` ceiling constants |
| `.githooks/pre-commit` | Thin shim: refuse when `python` is absent (exit 2), else delegate to the runner in staged mode. No gate logic of its own. | `.githooks/pre-commit:1-9` |
| `.githooks/pre-push` (context, not in the stated scope) | Refuse any push whose remote ref is `refs/heads/main` unless `CGO_ALLOW_MAIN=1`, then delegate to the full runner. | `.githooks/pre-push:1-13` |
| `docs/VERIFY.md` | Claim ledger, not a gate implementation. Defines statuses (VERIFIED/RECORDED/PENDING/BLOCKED/REFUTED/SUPERSEDED), the “row is source of truth” rule, and the promotion rule. The two rows that certify this subsystem are the six-gates row and the hooks row. | `docs/VERIFY.md:1-18` (intro + statuses); `:83-91` (How a row moves); `:42` (six gates VERIFIED); `:43` (hooks VERIFIED) |
| `tools/verify_test.py` | Unit tests for the runner’s pure helpers (verdict, staged selection, provenance, hygiene, links, exit classifier, seal, line ceiling) via stdlib `unittest`. | `tools/verify_test.py:1-100`, e.g. `:4-6` stdlib-only imports, `:13-96` test methods |

The hooks’ division of labour matches R1: R1 states a `pre-commit` hook takes no parameters, is invoked by `git-commit` before the commit message is obtained and the commit is made, can be bypassed with `--no-verify`, and aborts the commit on non-zero exit (R1, `pre-commit` section) — so a shim that only forwards to a staged gate check is exactly R1’s intended role. R1 likewise states `pre-push` is called by `git-push` to prevent a push, receives `<local-ref> SP <local-object-name> SP <remote-ref> SP <remote-object-name> LF` lines on STDIN, and aborts without pushing on non-zero exit (R1, `pre-push` section) — the `refs/heads/main` scan in `.githooks/pre-push:6-8` is a consumer of that protocol.

---

## 2. Exported gate surface (`file:line`)

`tools/verify.py:2` declares “Standard library only; no network; no shell.” Imports are the stdlib set `argparse, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time` plus `dataclasses`, `pathlib` (`tools/verify.py:4-6`).

### 2.1 Shared core

| Symbol | Surface | Notes |
|---|---|---|
| `Result(name, status, detail)` | `tools/verify.py:14-18` | `PASS/FAIL/SKIP = "pass","fail","skip"` (`:8`) |
| `Gate(name, proves, paths, run, prove, fast=True)` | `tools/verify.py:19-26` | `paths` drives staged selection; `fast=False` excludes the gate from `--fast` |
| `ROOT` | `tools/verify.py:7` | `Path(__file__).resolve().parents[1]` — hard-coded repo-relative root; no `--root` override |
| `verdict(results)` | `tools/verify.py:36` | `1` if any FAIL else `2` if any SKIP else `0` |
| `_proc(args, timeout, cwd=ROOT, text=True, env=None)` | `tools/verify.py:27-30` | Wrapper over `subprocess.run(..., check=False, capture_output=True, timeout=timeout)` (R2: `run()` returns `CompletedProcess`, `returncode 0` = success, `TimeoutExpired` on timeout) |
| `line_ceiling_problem(root, files, maximum)` | `tools/verify.py:37-41` | Sums line counts over `CEILING_FILES` |
| `missing_proofs(gates)` | `tools/verify.py:42-43` | Gate without callable `prove` fails startup |
| `MAX_GATE_LINES=650`, `CEILING_FILES`, `SIZE_CEILING=384*1024` | `tools/verify.py:11-13` | Ceiling check runs before any gate (`:486-489`); breach → exit 2 |

### 2.2 The six gates (registry order)

Registry `GATES` at `tools/verify.py:455-462`:

| # | Name (`Gate.name`) | `proves` string | `paths` (staged trigger) | `fast` | `run` | `prove` |
|---|---|---|---|---|---|---|
| 1 | `frozen` | “the frozen corpus audit still attests” | `("book/", "data/")` | True | `gate_frozen` `:63-81` | `prove_frozen` `:82-128` |
| 2 | `provenance` | “CSV headers and valuation enums cannot lie” | `("pkg/", "cmd/")` | True | `gate_provenance` `:147-169` | `prove_provenance` `:170-191` |
| 3 | `tests` | “go vet and the Go suite pass” | `()` (never staged-selected; see §2.3) | **False** | `gate_tests` `:195-209` | `prove_tests` `:215-234` |
| 4 | `exit` | “the CLI keeps the 0/1/2 contract” | `("pkg/", "cmd/")` | True | `gate_exit` `:262-291` | `prove_exit` `:292-303` |
| 5 | `hygiene` | “no tracked secret, artifact or oversized file” | `("tools/",)` | True | `gate_hygiene` `:362-373` | `prove_hygiene` `:374-388` |
| 6 | `links` | “AGENTS.md and the ledgers name existing paths” | `("docs/", "AGENTS.md")` | True | `gate_links` `:420-429` | `prove_links` `:430-451` |

Per-gate detail:

**frozen** — `SEAL="docs/frozen-manifests.sha256"` (`:47`). `seal_findings(root)` (`:48-62`) parses the seal (requires `digest  path` with path under `book/data/runs/` ending `/manifest.json`), SHA-256-hashes every `*/manifest.json` under `book/data/runs`, and reports changed/missing/extra manifests. `gate_frozen` (`:63-81`) SKIPs when `book/` or its audit script is absent, else runs `book/tools/audit-archive.py` via `_proc([sys.executable, audit], 120)` (R2 sequence-args form), requires exit 0 *and* ≥2 `OK` lines, then requires `seal_findings()==[]`, and PASSes with the OK lines + seal count. `prove_frozen` (`:82-128`) copies `book/` to a temp dir, re-derives a seal, proves the seal catches a rewritten manifest, proves the audit catches a mutated data file, and restores — a control-then-mutate fault proof. Python `hashlib`/`shutil`/`tempfile` semantics beyond stdlib existence: **`unverified — primary source not found`** (no `hashlib` doc fetched; only the call-site `file:line` is evidence).

**provenance** — `BASES={"measured","estimated"}`, `SOURCES={"ebpf","tcpinfo",""}` (`:132`). `csv_header` (`:133-135`), `header_problems` (`:136-137`: every `aqm_eval.csv` header must be in the frozen-era set), `valuation_problem` (`:138-146`: basis/source enum check + “estimated must not carry a source”). `gate_provenance` (`:147-169`) builds the reference header set from frozen `book/data/runs/*/aqm_eval.csv`, checks `data/runs` headers against it, and validates every `valuation.json` sidecar. `prove_provenance` (`:170-191`) is a pure temp-dir control/attack pair (extra column, `basis=guess`, `estimated+ebpf` must be caught; `measured+ebpf`, `estimated+""`, exact header must pass).

**tests** — `gate_tests` (`:195-209`) SKIPs when `web/frontend/dist/` is absent or `go` is not on PATH (`shutil.which("go")`), else runs `go vet ./...` (300 s) then `go test ./...` (600 s) via `_proc` and PASSes with wall time. `prove_tests` (`:215-234`) builds three throwaway Go modules (`_module`, `:210-214`) — a passing module as control, a `t.Fatal` module, and a `fmt.Printf("%d","x")` vet violation — and requires the first green and the other two red. Go toolchain semantics (`go vet`/`go test` flags, module format `go 1.25` at `:212`): **`unverified — primary source not found`** (no Go docs fetched; only `file:line` is evidence).

**exit** — `SAFE_INVOCATIONS` (`:238-244`): `()`, `("--nope",)`, `("version",)`, `("shape",)`, `("service",)` probed against a freshly built `cmd/cgo` binary (`gate_exit` `:266-291`) with an env scrubbed of `key|token|secret|password|credential` (`:274-275`). `classify_exit` (`:246-254`) fails exit codes outside `{0,1,2}`, panic/traceback output, or exit 0 with a refusal regex (`REFUSAL`, `:245`). `verdict_contract` (`:255-261`) pins `verdict()` to the 0/1/2 table. `prove_exit` (`:292-303`) feeds synthetic (code, out, err) triples to the classifier plus the verdict contract — no binary build. The CLI’s own 0/1/2 behaviour being tested (what `cgo` prints for `shape`/`service`/usage) is a product claim evidenced only by the ledger row at `docs/VERIFY.md:42-43`: **`unverified — primary source not found`** beyond `file:line` (product source not read in this pass).

**hygiene** — Selectors `KEY_RX/TOKEN_RX/NET_RX` (`:307-309`), allowlists `ALLOWED_ADDRESSES` (`:312`), `ALLOWED_ARTIFACTS` (`:313`), `LEGACY_ARCHIVES` (`:316-323`), `USER_DOC_RX` (`:324`), `FORBIDDEN_PREFIXES` (`:325`). `path_problem` (`:326-333`) rejects `backup/|data/|book/|testbed-bootstrap/|node_modules/|.worktrees/`, `unp*/`, `epbfBBR.md`, user docs, and `*.docx/*.pdf/*.tar.gz/*.zip/*.exe`. `scan_text` (`:334-343`) flags private-key blocks, token shapes, and non-allowlisted `100.x` tailnet addresses. `hygiene_findings` (`:344-361`) enforces `SIZE_CEILING` (except `docs/img/`), reads up to 2 MB per file, and scans. `gate_hygiene` (`:362-373`) enumerates tracked files via `git ls-files -z` (R2 sequence-args; `git` availability → SKIP on `OSError`). `prove_hygiene` (`:374-388`) is a pure control/attack set. Which addresses/tokens historically existed (the two allowlisted IPs, the 18 legacy archives) is ledger/history evidence only: **`unverified — primary source not found`** beyond `file:line` + `docs/VERIFY.md:43`.

**links** — `LINK_FILES` (`:392-393`): `AGENTS.md`, `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, `docs/superpowers/plans/README.md`. `link_candidates` (`:394-403`) extracts backticked tokens containing `/` whose first segment is an existing top-level entry; `missing_links` (`:404-414`) checks existence; `missing_plan_links` (`:415-419`) checks `2026-*-*.md` names against `docs/superpowers/{specs,plans}`. `gate_links` (`:420-429`) SKIPs if a scanned file is absent. `prove_links` (`:430-451`) is a temp-dir control/attack set. Whether a backticked token *should* be treated as a path (vs. a command) follows the local heuristic at `:397-403`; any broader markdown-link standard: **`unverified — primary source not found`** (no CommonMark doc fetched).

### 2.3 Selection and CLI surface

- `--fast` (`:480`, `:506-507`): all gates except `tests` (`fast=False` at `:458`).
- `--staged` (`:481`, `:503-505`): `select_staged(staged_paths())` (`:464-477`). `staged_paths()` runs `git diff --cached --name-only -z` (`:474`); `select_staged` (`:464-471`) returns **all** gates if `tools/verify.py` or `tools/verify_test.py` changed, else `hygiene` plus any gate whose `paths` prefix-matches a staged path. Consequence visible at `tools/verify.py:458`: `tests` has `paths=()` so it is *never* staged-selected — staged pre-commit never runs the Go suite. `hygiene` is always staged-selected (`:467`).
- `--only NAME` (repeatable) (`:482`, `:497-502`): unknown name → stderr + exit 2.
- `--list` (`:483`, `:493-496`): prints `name fast|full proof <proves>` for all gates, exit 0.
- `--selftest` (`:484`, `:510-516`): runs each selected gate’s `prove()` inside the runner, prints `caught`/`SURVIVED`, exit 1 if any survive.
- Default (no flags): all six gates (`:508-509`).
- Startup guards (`:486-492`): line-ceiling breach → stderr + exit **2**; any gate missing a fault proof → stderr + exit **2**. I.e. “2 is not green” is enforced *before* gate execution, mirroring `.githooks/pre-commit:5-8`.
- Output contract: one line per gate `STATUS name detail` plus a `N passed, N failed, N not run` summary (`:517-522`), then `verdict(results)` as the process exit code (`:523`). R2: `subprocess.run`’s `returncode` convention (0 = success) is the mechanism the runner *exploits* for child processes; the runner’s *own* 0/1/2 exit contract is a local design choice pinned by `verdict`/`verdict_contract` (`:36`, `:255-261`) — it is not an R2 requirement.

### 2.4 Unit-test surface

`tools/verify_test.py` targets `verify` via `sys.path.insert` (`:7-10`) using R3 `unittest.TestCase` (`:13`). Tests: `test_verdict` (`:14-18`), `test_staged_selection` (`:20-29`), `test_provenance` (`:31-39`), `test_hygiene` (`:41-50`), `test_links` (`:52-59`), `test_exit_classifier` (`:61-66`), `test_missing_proof` (`:68-72`), `test_seal` (`:74-88`), `test_line_ceiling` (`:90-96`). R3 states tests are `TestCase` methods named `test*` checked with `assert*` helpers and run via `unittest.main()` (`:99-100` here uses `unittest.main()`) — this file is a conformant R3 consumer. The `--staged` truth table at `:21-26` is the normative mapping for §2.3 (`book/data/runs/x → [frozen, hygiene]`, `pkg/api/x.go → [provenance, exit, hygiene]`, `docs/x.md → [hygiene, links]`, `tools/verify.py → all`, `README.md → [hygiene]`).

### 2.5 Hook surfaces

- `.githooks/pre-commit:1-9`: `#!/bin/sh`, `set -eu` (`:2`), `root=$(git rev-parse --show-toplevel); cd "$root"` (`:3-4`), `command -v python` guard → stderr + `exit 2` (`:5-8`), `python tools/verify.py --staged` (`:9`). R1 requires hooks to be executable and notes Git changes the working directory to the work-tree root for non-bare repos (R1, DESCRIPTION) — the `cd "$root"` is redundant under that rule but harmless (see §5). R1: non-zero exit aborts the commit; `--no-verify` bypasses (R1, `pre-commit` section) — the shim’s exit code *is* the runner’s verdict (0/1/2), so exit 2 (skip) also aborts the commit, which is the intended “2 is not green” semantics (`docs/VERIFY.md:26` narrowing note; `docs/VERIFY.md:43` “no `book/` → exit 2” evidence).
- `.githooks/pre-push:1-13` (context): parses STDIN ref lines per R1’s `pre-push` protocol (`:6-8`), refuses `refs/heads/main` with exit 1 unless `CGO_ALLOW_MAIN=1` (`:9-12`), then runs the full runner (`:13`). Push-to-main refusal (exit 1) vs. gate skip (exit 2) preserves the runner’s FAIL vs. not-green distinction.

### 2.6 `docs/VERIFY.md` gate-ledger surface

- Statuses: VERIFIED = “named command ran on named host and date, output saved/quoted in checkpoint”; RECORDED/PENDING/BLOCKED/REFUTED/SUPERSEDED with explicit non-promotion rules (`docs/VERIFY.md:7-18`). Promotion: only a named command on a named host + date + saved output; BLOCKED → PENDING only when the named condition clears; never by repetition (`:83-91`).
- Six-gates row (`:42`, VERIFIED 2026-10-28): `python tools/verify.py` → exit 0, six PASS, `6 passed, 0 failed, 0 not run` (wall 82.9 s); `--selftest` → six `caught`; `--list`/`--fast`/`--only links`; `tools/verify_test.py` → 8 tests OK; gate total 596 lines at `MAX_GATE_LINES` 600 (historical values; current constants are 650/527 lines — `tools/verify.py:11` + file length).
- Hooks row (`:43`, VERIFIED 2026-10-28): `core.hooksPath=.githooks`, both hooks mode 100755; disposable-clone evidence: staged `ghp_` token → exit 1 (`token shape`), staged `book/data/` probe → exit 1 (`forbidden tracked path` + `frozen` SKIP), `refs/heads/main` → exit 1, `CGO_ALLOW_MAIN=1` → exit 2 with 2 SKIPs, staged `docs/VERIFY.md` → exit 0, missing `python` → exit 2.
- Seal row (`:26`, VERIFIED 2026-10-28): `sha256sum -c docs/frozen-manifests.sha256` → 159 OK; notes the anchor is a static hash list, not a signature (“a writer who can rewrite the anchor too is out of scope”).

---

## 3. Dependencies

### 3.1 Runtime (declared by the code)

- **Python 3 stdlib only** (`tools/verify.py:2-6`). No third-party packages, no network calls in the runner itself (`:2`). External processes are spawned only via R2 `subprocess.run` sequence-args (`:27-30`), never `shell=True` — consistent with R2’s Security Considerations (“This library will not implicitly choose to call a system shell … can safely be passed to child processes” without shell quoting).
- **`python` on PATH** for the hook shim (`.githooks/pre-commit:5-7`); `sys.executable` is used inside the runner for the book audit (`tools/verify.py:68`) and R2 documents `sys.executable` as the recommended way to re-launch the current interpreter.
- **`git`** for `git ls-files -z` (`:364`) and `git diff --cached --name-only -z` (`:474`); `git rev-parse --show-toplevel` for the shim (`.githooks/pre-commit:3`). R1: hooks live under `core.hooksPath` (R1 DESCRIPTION) — which directory Git searches is Git-configured, not repo-guaranteed: config-file syntax/scope **`unverified — primary source not found`** (R6, §Preamble).

### 3.2 Toolchain (per-gate, all optional-with-SKIP)

| Gate | External tool | SKIP condition |
|---|---|---|
| frozen | `sys.executable` + `book/tools/audit-archive.py` | `book/` or audit script/corpus absent (`:64-66`) |
| provenance | none (pure file reads) | frozen reference absent (`:149-153`) |
| tests | `go` on PATH **plus** `web/frontend/dist/` present | either absent (`:196-200`) |
| exit | `go` on PATH (to build `./cmd/cgo`) | absent (`:264-265`) |
| hygiene | `git` | `OSError`/`SubprocessError` (`:365-366`) |
| links | none (pure file reads) | any `LINK_FILES` entry absent (`:421-423`) |

Go toolchain versions/flags, `git ls-files -z`/`git diff --cached -z` wire formats, `web/frontend/dist` production rules, and SHA-256 seal file format beyond their `file:line` occurrences: **`unverified — primary source not found`** (Go/Git porcelain docs not fetched; R1 covers hook exit/STDIN semantics only).

---

## 4. Primary-source requirements (what a standalone tool must satisfy)

| # | Requirement | Primary source | Consequence for `cgo-verify` |
|---|---|---|---|
| H1 | `pre-commit` takes no parameters; non-zero exit aborts the commit before it is created; `--no-verify` bypasses it | R1, `pre-commit` section | Shim must translate the runner verdict to 0/1/2 without extra args; document `--no-verify` as a deliberate bypass, never as an error path |
| H2 | `pre-push` receives `<local-ref> SP <local-object-name> SP <remote-ref> SP <remote-object-name> LF` on STDIN; non-zero exit aborts the push without pushing; reasons go to stderr | R1, `pre-push` section | Main-ref guard must parse exactly that line format; denial message on stderr (current `:10` does this) |
| H3 | Hook scripts without the executable bit are ignored; the hooks directory defaults to `$GIT_DIR/hooks` and is overridable via `core.hooksPath` | R1, DESCRIPTION + `githooks - Hooks used by Git` SYNOPSIS | Repo must ship an install step that sets `core.hooksPath=.githooks` + `chmod +x` (the ledger records this as already done: `docs/VERIFY.md:43`); a fresh clone without that step has no gates (see §5 gap G1). Config-file details: **`unverified — primary source not found`** (R6) |
| H4 | Git changes the hook working directory to the work-tree root (non-bare) / `$GIT_DIR` (bare), and exports `GIT_DIR` etc. | R1, DESCRIPTION | Hooks must not assume the invocation cwd; `cd "$(git rev-parse --show-toplevel)"` (`.githooks/pre-commit:3-4`) is correct defensive practice even if redundant |
| P1 | `subprocess.run` sequence-args; `CompletedProcess.returncode` is the exit status; 0 = success | R2, `run()` + `CompletedProcess.returncode` | All child gates must keep `shell=False` sequence-args and branch on `returncode`, with `TimeoutExpired` → FAIL (current pattern at `:64-73`, `:201-208`, `:266-271`) |
| P2 | `check=True` raises `CalledProcessError` on non-zero exit | R2, `run()` “check” + `CalledProcessError` | The runner deliberately uses `check=False` (`:30`) and inspects `returncode` itself — a standalone fork into a library API should preserve that (error mapping is gate logic, not subprocess boilerplate) |
| P3 | Unit tests are `TestCase`/`test_*`/`assert*` + a `main()` runner; setup/teardown share fixtures | R3, overview + `TestCase` | The extracted package ships its gate-helper tests in the same R3 shape as `tools/verify_test.py:13-100`; pytest migration is out of scope (stdlib-only is the product constraint at `:2`) |
| V1 | A released version’s contents MUST NOT be modified; changes ship as a new version | R4, §3 | `cgo-verify` releases are immutable: no moving tags, no re-cut binaries under the same version; gate threshold changes bump per §4 below |
| V2 | `X.Y.Z` with PATCH = compatible bug fixes, MINOR = compatible new functionality, MAJOR = incompatible API changes; `1.0.0` defines the public API | R4, Summary + §§1–8 | Declare the public API before 1.0.0: `Gate`/`Result` protocol, CLI flags, 0/1/2 exit contract, staged-selection table, hook shims. R4 FAQ: tag `v1.2.3` ≠ version `1.2.3` — the release command must state which one it prints |
| S1 | Producer MUST choose a build platform capable of the desired SLSA level | R5, “Choose an appropriate build platform” | Name the builder per artifact (local Python project vs. hosted CI); claim only the level the platform supports |
| S2 | Producer MUST follow a consistent build process; if the ecosystem requires build metadata, the config file MUST be complete and up to date | R5, “Follow a consistent build process” | Version/flags/gate thresholds live in one checked-in config, not in ad-hoc shell history |
| S3 | Provenance MUST identify the output package by cryptographic digest and describe how it was produced; the producer MUST distribute provenance to consumers | R5, “Provenance Exists” + “Distribute provenance” | Every `cgo-verify` release ships a digest-identified provenance record (what built what, from which commit, with which gate thresholds), not just a tarball |
| S4 | L2: provenance is signed by the build platform (authentic); L3: provenance is unforgeable and builds are isolated + ephemeral | R5, “Provenance is Authentic” / “Provenance is Unforgeable” + “Isolated” | State the achieved level per release (L1 at minimum). Signing-key custody and hosted-builder choice: **`unverified — primary source not found`** for this org (no policy read this pass) |

---

## 5. Gap analysis (ranked, as a standalone tool)

**G1 — Hooks are opt-in and silently absent without setup.** R1 requires the executable bit and a `core.hooksPath`/`$GIT_DIR/hooks` location (R1 DESCRIPTION); nothing in the repo enforces prereqs beyond `docs/VERIFY.md:43`’s one-time manual setup. A fresh clone that forgets `git config core.hooksPath .githooks` or the `+x` bit has *no* gate protection and no diagnostic. Standalone fix: an `install`/`doctor` step that asserts both and fails loudly (exit 2 semantics already exist at `.githooks/pre-commit:5-8`).

**G2 — `--staged` cannot select the slowest gate by construction.** `tests` has `paths=()` (`tools/verify.py:458`) so `select_staged` (`:464-471`) never returns it — pre-commit therefore *by design* never runs `go vet`/`go test` (gates take 75–280 s per `docs/VERIFY.md:42,81`). That is a defensible latency policy but is undocumented as policy; a standalone tool must document which gates are staged-eligible and why, or add an opt-in `CGO_VERIFY_STAGED_TESTS=1` path.

**G3 — Hard-coded CGO layout (`ROOT`, `book/`, `data/`, `pkg/`, `cmd/`, `docs/`, `web/frontend/dist`, `cmd/cgo`).** `ROOT` is `parents[1]` (`:7`); every gate bakes relative paths (`:47`, `:64`, `:148`, `:196`, `:269`, `:392`). There is no `--root`, no project config, and no gate-plugin API (`Gate` is an internal tuple at `:455-462`). Export as-is and the tool only runs inside this repo. Standalone fix: a project-descriptor (paths + audit command + frontend dist + CLI build target) with repo defaults.

**G4 — Seal is a hash list, not a signature.** The ledger already narrows this: “Proves a static anchor only: not a signature; a writer who can rewrite the anchor too is out of scope” (`docs/VERIFY.md:26`). R5 L2/L3 require signed/unforgeable provenance (R5, “Provenance is Authentic”/“Unforgeable”) — `docs/frozen-manifests.sha256` + `seal_findings` (`:48-62`) can at best claim SLSA **L1**. A standalone verifier must either sign the seal (key custody: **`unverified — primary source not found`** for this org) or label its frozen/provenance attestations L1-only.

**G5 — Exit 2 conflates “could not run” with “passed with skips”.** `verdict()` (`:36`) returns 2 if any gate SKIPs, and startup guards also return 2 (`:486-492`). That is the house rule (“2 is not green”) and hooks correctly block on it (R1 non-zero aborts), but a library consumer cannot distinguish “prerequisite missing (e.g. no `go`)” from “gate skipped for a known reason”. The per-gate `detail` strings do carry the reason (`:66`, `:150`, `:197`, `:265`, `:366`, `:423`) — the gap is only the absence of a machine-readable reason code.

**G6 — No machine-readable output.** `main()` prints human lines and returns an int (`:517-523`); there is no `--json`/`--junit`, so CI and the `docs/VERIFY.md` ledger both parse free text (ledger evidence is quoted stdout, `docs/VERIFY.md:42-43`). R5 S3 (distribute provenance that identifies the artifact by digest) cannot be satisfied by stdout alone.

**G7 — Gate thresholds versioned only by editing source.** `MAX_GATE_LINES`/`CEILING_FILES`/`SIZE_CEILING` (`:11-13`), hygiene allowlists (`:312-325`), `LINK_FILES` (`:392-393`), `SAFE_INVOCATIONS` (`:238-244`) all live in `tools/verify.py`. R4 V1/V2 require the *released* tool’s contents to be immutable per version and increments to convey API meaning — so every threshold tweak is today a new unversioned source edit. A standalone repo needs the thresholds in versioned config, with MINOR for new gates/flags and MAJOR for contract breaks (R4 §§7–8).

**G8 — `prove_*` duplication between runner and unit tests.** Every pure helper is covered twice: once in `prove_*` (`:170-191`, `:292-303`, `:374-388`, `:430-451`) and once in `tools/verify_test.py` (`:20-96`), with `test_seal`/`test_line_ceiling` having *no* `--selftest` counterpart beyond `prove_frozen`/the ceiling guard. That duplication is intentional (fault proofs vs. unit tests) but drift-prone; a standalone package should generate or co-locate them.

**G9 — `subprocess` fragility surface.** Five distinct timeout values (120/300/600/30 s across `:68`, `:202`, `:278`) with no jitter/backoff; `myth` of determinism is already REFUTED in the ledger (E5 flake, `docs/VERIFY.md:44-46`). R2 guarantees `TimeoutExpired` on expiry but gives no guidance on retry/cancellation policy: retry semantics **`unverified — primary source not found`** in this pass.

**G10 — `verify.py` vs. `verify_test.py` ceiling coupling.** `CEILING_FILES` (`:12`) counts `tools/verify.py + tools/verify_test.py + .githooks/pre-commit + .githooks/pre-push` against `MAX_GATE_LINES=650` (`:11`, `:486-489`). The ceiling guards the *runner*, not the *product*; extracting `cgo-verify` dissolves the coupling — the new repo needs its own gate-budget story or it inherits a 650-line limit for a different codebase.

---

## 6. Target tool: `cgo-verify` as an independent gate-runner repo

Design principle: extract the *protocol* (Gate/Result/verdict/0-1-2/selftest/staged-selection), keep the *CGO binding* (book audit, CSV eras, Go suite, CLI build) as a thin adapter.

### 6.1 Repository layout (proposed)

```
cgo-verify/
  pyproject.toml            # name, version (SemVer), python>=3.x, entry point cgo-verify
  src/cgo_verify/
    core.py                 # Result, Gate, verdict, select_staged, main plumbing (no CGO paths)
    gates.py                # generic gates: seal, headers, hygiene, links, exit-classifier, line-ceiling
    selftest.py             # prove_* registry + --selftest runner
    hooks/pre-commit        # shim: python check + cgo-verify --staged (R1 H1)
    hooks/pre-push          # shim: main-ref guard + cgo-verify (R1 H2)
  adapters/cgo/
    gates.py                # frozen (book audit), provenance eras, tests (go+dist), exit (cmd/cgo)
    project.toml            # paths: book/, data/, pkg/, cmd/, docs/, web/frontend/dist, cmd/cgo build target
  tests/                    # R3 TestCase suites (successor to tools/verify_test.py)
  docs/VERIFY.md            # same ledger discipline (statuses + promotion rule)
  .githooks -> hooks install script (core.hooksPath + chmod +x, R1 H3)
```

Why this split: R4 §1 requires a *declared* public API — `core.Gate/Result/verdict` + the CLI contract + the project-descriptor schema *is* that API; R5 S2 requires the build config (here `project.toml` + `pyproject.toml`) to be complete and checked in; R5 S3 requires per-release provenance identifying the artifact by digest.

### 6.2 Public API to declare before 1.0.0 (R4 V2)

- `Gate(name, proves, paths, run, prove, fast)` + `Result(name, status, detail)` (`tools/verify.py:14-26`).
- `verdict(results) -> {0,1,2}` with FAIL > SKIP > PASS precedence (`:36`, pinned by `:255-261`).
- CLI: `--fast | --staged | --only NAME | --list | --selftest | (default: all)` (`:479-484`); exit codes 0/1/2 preserved; add `--json FILE` for machine-readable provenance (closes G6) without changing the human default.
- `project.toml` schema: `root`, `seal`, `frozen_audit`, `frozen_corpus`, `data_corpus`, `go_root`, `frontend_dist`, `cli_build_target`, `link_files`, `allowed_*` allowlists, `size_ceiling`, `max_gate_lines` — every current literal in §2 becomes a field with the repo value as default.

### 6.3 Provenance & versioning (R4 + R5)

- **SemVer**: `0.x` until the API in §6.2 is declared stable; MINOR for new gates/flags/config fields, MAJOR for verdict/CLI/selection breaks, PATCH for bug fixes (per R4 Summary + §§6–8). Never move a released tag or binary (R4 §3). The `--version` output prints `X.Y.Z`, and release notes state the tag (`vX.Y.Z`) separately (R4 FAQ).
- **SLSA**: L1 from the first release (provenance file naming the artifact digest, the commit, and the effective thresholds/config hash — R5 “Provenance Exists”); L2 once a hosted builder signs the provenance (R5 “Provenance is Authentic”); L3 explicitly a non-goal until builds are isolated/ephemeral (R5 “Isolated”). The frozen-corpus seal in the CGO adapter remains L1-only until signed (G4).

---

## 7. Slices (each with acceptance)

- **S1 — Extract `core` + CLI with no CGO paths.** Move `Result/Gate/verdict/_proc/line_ceiling/missing_proofs` + `main()` selection (`tools/verify.py:7-43,478-527`) into `core.py`; keep flags byte-identical. Acceptance: `cgo-verify --list` prints the six gate names with identical `proves` strings (`:455-462`); `tools/verify_test.py` passes unchanged against `core` (R3 shape preserved); no CGO literal remains in `core.py`.
- **S2 — Generic gates.** Port `seal_findings`/`gate_frozen` plumbing, `header_problems`/`valuation_problem`, `path_problem`/`scan_text`/`hygiene_findings`, `missing_links`/`missing_plan_links`, `classify_exit`/`verdict_contract` into `gates.py` parameterised by a descriptor. Acceptance: every `prove_*` (`:82-128,170-191,292-303,374-388,430-451`) passes against synthetic fixtures; `tools/verify_test.py` cases (`:20-96`) all ported.
- **S3 — CGO adapter + `project.toml`.** Repo literals (`:47,64,148,196,269,392`) become descriptor fields; `gate_frozen`/`gate_provenance`/`gate_tests`/`gate_exit` become adapter gates. Acceptance: `cgo-verify --only frozen` + `--only provenance` + `--only exit` reproduce `docs/VERIFY.md:42`’s six-PASS run from the CGO repo root; `--staged` table matches `tools/verify_test.py:21-26`.
- **S4 — Hooks + installer.** Ship `hooks/pre-commit`/`pre-push` (R1 H1/H2) and an `install` command that sets `core.hooksPath` and the executable bit (R1 H3), plus a `--check` that fails with exit 2 when either is missing (closes G1). Acceptance: disposable-clone redo of `docs/VERIFY.md:43` (secret → 1, main push → 1, missing `python` → 2, `CGO_ALLOW_MAIN=1` → 2-with-SKIPs).
- **S5 — Machine-readable provenance + SemVer release.** Add `--json`, publish L1 provenance per release (digest + commit + effective config), declare the public API in a `PROTOCOL.md`, cut `0.1.0` with `VERSION` == tag-without-`v` (R4). Acceptance: a consumer can parse the JSON into the same verdict without scraping stdout (closes G6); `sha256sum` of the release artifact matches the provenance digest (R5 S3).
- **S6 — Fault-proof co-location + ceiling split.** Merge `prove_*` and `tools/verify_test.py` coverage so each helper has one canonical control/attack pair; give the new repo its own line ceiling independent of CGO’s 650 (G10). Acceptance: `--selftest` still yields six `caught` (ledger `docs/VERIFY.md:42`); line ceiling no longer references `web/frontend/dist` or `cmd/cgo`.
- **S7 (conditional) — Signed seal (SLSA L2).** Sign `docs/frozen-manifests.sha256` and the `cgo-verify` release provenance; state the achieved SLSA level per artifact (R5 S4). Key custody / transparency-log choice: **`unverified — primary source not found`** for this org (no policy read; document before implementing).

Out of scope: the CGO product itself (`pkg/`, `cmd/`, `web/frontend/`, `book/`, `kit/` measurement code) and the SLSA L3 platform — both named as non-goals per R5’s isolation requirements.

---

## 8. Anti-toy criteria (what fails review as a standalone tool)

A `cgo-verify` submission fails review if any of these holds:

1. **It only runs inside CGO.** Any hard-coded `book/|data/|pkg/|cmd/|web/frontend/dist` path with no descriptor override (`tools/verify.py:47,64,148,196,269,392` pattern) is not a tool, it is a script. Fails R4 §1 (no declared, reusable public API).
2. **Exit 2 is treated as green.** Any path where a SKIP (or missing prerequisite) yields process exit 0, or where `--selftest` SURVIVED yields 0, breaks the contract (`:36`, `:355-261`, `:510-516`) and R1’s non-zero-aborts rule (R1 `pre-commit`/`pre-push`). Ledger phrasing is “2 is not green” (`docs/VERIFY.md:26,43`) — keep it.
3. **Hooks without install.** A repo whose `.githooks/` ships without an installer asserting `core.hooksPath` + executable bit fails R1 DESCRIPTION (hooks without the bit are ignored) and leaves fresh clones unprotected (G1).
4. **Unversioned mutation.** A released `cgo-verify` binary/tag whose contents change in place fails R4 §3; a MINOR that breaks the CLI or verdict contract fails R4 §§7–8. Threshold edits in source (not config) fail R5 S2 (incomplete build metadata).
5. **No provenance.** A release without a digest-identified provenance record fails R5 S3 (“MUST identify the output package by cryptographic digest” + “MUST distribute provenance”). The seal stays L1-only until signed (G4) — claiming L2/L3 without a signed/isolated builder fails R5 S4.
6. **Shell-spawned gates.** Any `shell=True` or string-command subprocess breaks R2’s Security Considerations and reintroduces quoting/injection the sequence-args design (`:27-30`) exists to avoid.
7. **Staged-select lies.** Adding a gate to staged selection without registering its `paths` (or silently including `tests` at pre-commit latency) breaks the `tools/verify_test.py:21-26` truth table — the one table a commit author relies on.
8. **Ledger laundering.** Promoting a claim from RECORDED to VERIFIED without a named command on a named host with saved output fails `docs/VERIFY.md:83-91` (“RECORDED is not promoted by confidence, repetition, or a later document agreeing”). The tool’s own release claims obey the same rule.

---

*Teams checklist: every codebase claim traces to `file:line`; every normative claim traces to R1–R5; R6 config-file syntax, Go toolchain, `hashlib`/`argparse`, Git porcelain formats, GoReleaser/cosign/Prometheus-Nginx specifics, retry policy, and signing-key custody are **`unverified — primary source not found`** in this pass and were not used as requirements.*
