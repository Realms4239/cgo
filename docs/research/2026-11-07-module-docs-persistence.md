# Docs / Persistence / Ledger layer as a standalone modular tool — primary-source research

*Date: 2026-11-07 · Scope: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, `tools/verify.py`, `docs/superpowers/*`, `docs/api.md`, `docs/architecture.md`, `.gitignore`, `Makefile`, `.githooks/*` read as if extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | SLSA Provenance v1.1 | https://slsa.dev/spec/v1.1/provenance | fetched | What provenance must contain (buildDefinition, runDetails, builder.id, resolvedDependencies, subject) to be verifiable |
| R2 | SLSA Requirements v1.1 — Producing artifacts | https://slsa.dev/spec/v1.1/requirements | fetched | Provenance Exists / Authentic / Unforgeable levels; producer vs build-platform duties |
| R3 | CycloneDX SBOM capability + CycloneDX v1.5 JSON Reference | https://cyclonedx.org/capabilities/sbom/ , https://cyclonedx.org/docs/1.5/json/ | fetched | SBOM inventory contract (bomFormat, specVersion, serialNumber, components with hashes / licenses / supplier) a persistence module must parallel |
| R4 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | Public-API declaration, MAJOR/MINOR/PATCH rules, immutability of released versions |
| R5 | gitignore manual (git-scm) | https://git-scm.com/docs/gitignore | fetched | Why tracked-vs-ignored decides what persists; "Files already tracked are not affected" |
| R6 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | *unverified – primary source not found* in this pass (cited only from sibling research `docs/research/2026-11-07-module-metrics.md:11`; not re-fetched here) | Module path / versioning mechanics for a standalone extraction |
| — | AIR / DISanalyzer rule system (as synthesised in `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:19-36`) | — | **unverified — primary source not found** (not fetched; only the local synthesis cites them) | Admission test, fault-proof rule, three-state exit contract |
| — | verifiable-ledger / transparency-log prior art (e.g. in-toto, Rekor, Sigstore) | — | **unverified — primary source not found** (not fetched in this pass) | Whether an append-only Markdown ledger can claim tamper-evidence without signatures |

Method: read `docs/VERIFY.md:1-115`, `docs/CHECKPOINTS.md:1-809` (sampled beyond), `docs/ERRORS.md:1-233`, `docs/superpowers/plans/README.md:1-376`, `docs/api.md:1-116`, `docs/architecture.md:1-103`, `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:1-666`, `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:1-378`, `docs/superpowers/specs/2026-10-31-cgo-operator-console-design.md:1-505`, `docs/superpowers/plans/2026-10-27-cgo-engineering-system.md:1-912`, `tools/verify.py:1-527`, `.gitignore:1-82`, `Makefile:1-59`, plus grep of importers / callers of the ledger formats (none — Markdown is the carrier, `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:509`). Every technical claim below cites `file:line` or `(R#)`; anything else is labelled **unverified — primary source not found**.

---

## 1. Responsibility — what the docs/ledger/verify layer does today

Today the **Markdown ledgers + `tools/verify.py` are the entire persistence contract** of CGO. There is no database, no SQLite/WAL, no object store for claims: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:378` explicitly records "SQLite/WAL journaling — CSV append + manifest remains" as out of scope. The layer does three jobs:

1. **Claim persistence (the ledger).** `docs/VERIFY.md:1-5` states its contract in one paragraph: "One row per claim a reader could rely on. A number in a document cites a row here, or is labelled `RECORDED` or a hypothesis. The row is the source of truth; prose is not. Every row carries a date and an evidence path or command." Statuses are a closed set — `VERIFIED`, `RECORDED`, `PENDING`, `BLOCKED`, `REFUTED`, `SUPERSEDED` (`docs/VERIFY.md:7-18`) — updated from the five-status set in the founding spec (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:349-358`) by adding `SUPERSEDED` later. A row moves only "with a command, a date and the output saved or quoted" (`docs/VERIFY.md:83-91`); `RECORDED` "is not promoted by confidence, repetition, or a later document agreeing" (`docs/VERIFY.md:87-88`).

2. **Session persistence (the checkpoints).** `docs/CHECKPOINTS.md:1-4` is "Append-only, newest at the bottom. A session that ends without an entry loses its detail to compaction" and defines a resume protocol (`docs/CHECKPOINTS.md:6-13`) and a fixed entry format (`docs/CHECKPOINTS.md:15-28`: Branch/HEAD, Changed, Proven, Unproven/blocked, Next). The founding spec justifies the file the same way (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:390-414`). `docs/ERRORS.md:1-18` is the third ledger — append-only, "an error without a cause is a rumor; a fix without a proof is a wish" — with entries E1–E12 in `docs/ERRORS.md:20-233`.

3. **Machine enforcement (the gates).** `tools/verify.py:1-2` is a "Six-gate verifier for CGO. Standard library only; no network; no shell." The six gates are `frozen`, `provenance`, `tests`, `exit`, `hygiene`, `links` (`tools/verify.py:454-462`), each a `Gate(name, proves, paths, run, prove)` record with a mandatory fault proof (`tools/verify.py:19-26`, `tools/verify.py:490-492`). The verdict is three-state: any FAIL → 1, else any SKIP → 2, else 0 (`tools/verify.py:35-36`, `tools/verify.py:523`). The plan index is the backlog that points at the ledgers: `docs/superpowers/plans/README.md:333-350` ("The plan index is the backlog; the checkpoint is the history" — verbatim at `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:470-472`) and its "Standing constraints" say "Cite that row; never restate a count from memory" (`docs/superpowers/plans/README.md:335-337`).

What the layer does **not** do: it stores no measurement data (that is `book/data/runs`, `data/runs`, CSVs — `docs/architecture.md:78-83`), it exposes no HTTP surface (that is `pkg/api` — `docs/architecture.md:10-21`, `docs/api.md:1-116`), and it does not sign anything. That negative scoping is what makes extraction plausible: the gate runner imports only Python stdlib (`tools/verify.py:3-5`: `argparse, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time`) and is invoked only via `Makefile:24-36` (`verify`, `verify-fast`, `verify-selftest`, `hooks`) and `.githooks/pre-commit|pre-push` (see §2.5).

`docs/architecture.md:6` is the pointer that binds the layer to the product: "Les règles d'exploitation et l'état courant des affirmations vivent dans `AGENTS.md` et `docs/VERIFY.md`." The system was designed that way deliberately: "Six gates, each proven to fail; one ledger per kind of truth; one written scope per commit" (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:32-36`).

---

## 2. Exported contracts (file:line)

Complete inventory of the contracts a modular tool would export. "Exported" for Markdown means a format a writer must obey; for Python it means a callable / CLI flag.

### 2.1 `docs/VERIFY.md` — claim-ledger contract

| Export | Kind | Location | Notes |
|---|---|---|---|
| Row schema `| Claim \| Status \| Evidence \| Date \| Host |` | table | `docs/VERIFY.md:22-24` | One row per claim "a reader could rely on" (`docs/VERIFY.md:3`); evidence is a command or a path. |
| Status enum (6) | closed set | `docs/VERIFY.md:7-18` | `VERIFIED` (command ran on host/date, output saved/quoted), `RECORDED` (prior doc, not re-run — "It is not 'verified'" `docs/VERIFY.md:11-12`), `PENDING` (check exists, not run), `BLOCKED` (named condition), `REFUTED` (check showed false, row stays), `SUPERSEDED` (surface gone, row stays as history). |
| Move rule | prose | `docs/VERIFY.md:83-91` | Only a named command on a named host moves a row; `REFUTED` and `SUPERSEDED` never get deleted. Example: the frozen-corpus seal row `docs/VERIFY.md:26` carries the full verifier command output + narrowing note inline. |
| Verifier signature block | table + prose | `docs/VERIFY.md:93-115` | Session id, host, branch/HEAD, per-gate table (Go suite / six gates / book archive / frontend / build / vet / formatting), "Residuals accepted: Y", "Signature: verified …" (`docs/VERIFY.md:99-113`). This is the closest thing to a SLSA `runDetails` (R1) — see §4.1. |

### 2.2 `docs/CHECKPOINTS.md` — session-ledger contract

| Export | Kind | Location | Notes |
|---|---|---|---|
| Entry schema | heading + 5 bullets | `docs/CHECKPOINTS.md:15-28` | `## YYYY-MM-DD HH:MM UTC · <focus>` then Branch/HEAD, Changed, Proven (`<command → result, host>`), Unproven/blocked, Next. "UTC comes from a command, never the taskbar" (`docs/CHECKPOINTS.md:27`). |
| Resume protocol | 3 steps | `docs/CHECKPOINTS.md:6-13` | `AGENTS.md` → latest entry → `docs/VERIFY.md` → plan-index row → named artifact; compare branch/HEAD with `git log --oneline -1`; "The artifact wins over the entry's prose." |
| Append-only rule | prose | `docs/CHECKPOINTS.md:3` | Newest at the bottom; written "at session end, at a settled decision, and at a handoff" (`docs/CHECKPOINTS.md:28`). |

### 2.3 `docs/ERRORS.md` — defect-ledger contract

| Export | Kind | Location | Notes |
|---|---|---|---|
| Entry schema | 5-step | `docs/ERRORS.md:7-18` | `## E<n>. <title> (YYYY-MM-DD)` then Symptom / Cause / Fix / Proof / Lesson. "A fixed defect stays because it is the evidence that it was understood" (`docs/ERRORS.md:18`). Entries E1–E12 at `docs/ERRORS.md:20-233` (E5 at `docs/ERRORS.md:78-129` is the Windows temp-cleanup flake carried through multiple updates). |

### 2.4 `docs/superpowers/plans/README.md` — plan-index contract

| Export | Kind | Location | Notes |
|---|---|---|---|
| Slice map | table | `docs/superpowers/plans/README.md:27-58` | `# \| Slice \| Spec \| Plan \| Status \| Depends on \| Next action / blocker`; statuses defined at `docs/superpowers/plans/README.md:10-23` (`Done`, `Windows-green`, `In progress`, `Approved`, `Proposed`, `Deferred`, `Blocked`, `Superseded`, `Not planned`, `Historical`). |
| Standing constraints | list | `docs/superpowers/plans/README.md:333-350` | Frozen-corpus citation rule, CSV freezes, `basis: measured\|estimated` preservation, "Never fabricate a measurement", "Do not merge to `main`", explicit-path commits. |
| Windows gates | code block | `docs/superpowers/plans/README.md:352-365` | `go test ./...`, `go vet ./...`, cross-build, `python book/tools/audit-archive.py`, `git status --short`; "If `book/` or `web/frontend/dist` is absent … record the gate as 'not runnable here'; do not claim it passed" (`docs/superpowers/plans/README.md:364-365`). |
| Plan hygiene | numbered | `docs/superpowers/plans/README.md:367-376` | Spec+plan before code; index updated in the same commit; "Every number in a plan is labelled: measured, recorded, pending or blocked" (`docs/superpowers/plans/README.md:376`). |

### 2.5 `tools/verify.py` — gate-runner contract

| Export | Kind | Location | Notes |
|---|---|---|---|
| `PASS, FAIL, SKIP` + `Result(name, status, detail)` + `Gate(name, proves, paths, run, prove, fast)` | dataclasses | `tools/verify.py:8-26` | `Result` at `tools/verify.py:14-18`, `Gate` at `tools/verify.py:19-26`; every gate MUST have a callable `prove` or startup refuses it (`tools/verify.py:490-491`). |
| `verdict(results)` | func | `tools/verify.py:35-36` | 1 if any FAIL, else 2 if any SKIP, else 0. Tested as a contract at `tools/verify.py:255-261`. |
| `MAX_GATE_LINES = 650`, `CEILING_FILES`, `SIZE_CEILING = 384*1024` | consts | `tools/verify.py:11-13` | Ceiling enforced at startup: `line_ceiling_problem` at `tools/verify.py:37-41`, checked in `main` at `tools/verify.py:486-489`. |
| `gate_frozen` / `prove_frozen` | funcs | `tools/verify.py:63-128` | Delegates to `book/tools/audit-archive.py` (the "only source", `tools/verify.py:46`), additionally checks the SHA-256 seal `docs/frozen-manifests.sha256` via `seal_findings` (`tools/verify.py:47-62`); SKIP when `book/` absent (`tools/verify.py:65-66`); fault proof copies `book/` to a temp dir, rewrites a manifest, asserts seal catches it while audit passes, then mutates a corpus file and asserts audit catches it (`tools/verify.py:82-128`). |
| `gate_provenance` / `prove_provenance` | funcs | `tools/verify.py:147-191` | Header set = frozen eras (`tools/verify.py:151`); `valuation_problem` checks `basis ∈ {measured, estimated}`, `source ∈ {ebpf, tcpinfo, ""}`, estimated+source refused (`tools/verify.py:138-146`); fault proof: extra column, `basis=guess`, `estimated+ebpf` caught (`tools/verify.py:170-191`). |
| `gate_tests` / `prove_tests` | funcs | `tools/verify.py:195-234` | `go vet` + `go test` with timeouts 300/600 s (`tools/verify.py:202`); SKIP when `web/frontend/dist` absent (`tools/verify.py:196-197`); proof uses temp Go modules (passing, failing, vet-error) (`tools/verify.py:215-234`). |
| `gate_exit` / `prove_exit`, `SAFE_INVOCATIONS`, `classify_exit`, `verdict_contract` | funcs/consts | `tools/verify.py:238-303` | Five safe invocations (`tools/verify.py:238-244`); `classify_exit` rejects codes ∉ {0,1,2}, panics/tracebacks, exit-0-with-refusal (`tools/verify.py:246-254`); proof fixtures exit-3, traceback, refusal-vs-exit-0 (`tools/verify.py:292-303`). |
| `gate_hygiene` / `prove_hygiene`, `scan_text`, `path_problem`, `ALLOWED_ADDRESSES`, `LEGACY_ARCHIVES` | funcs/consts | `tools/verify.py:306-388` | Scans `git ls-files -z` for private keys, token shapes, tailnet addresses (`tools/verify.py:307-309`), forbidden paths incl. `backup/ data/ book/ *.docx *.pdf *.tar.gz *.zip *.exe` (`tools/verify.py:325-333`); 2 allowlisted tailnet addresses (`tools/verify.py:312`); 18 legacy `backup/*.tar.gz` allowlisted as named debt (`tools/verify.py:316-323`). |
| `gate_links` / `prove_links`, `LINK_FILES`, `link_candidates`, `missing_links`, `missing_plan_links` | funcs/consts | `tools/verify.py:391-451` | Scans `AGENTS.md, docs/VERIFY.md, docs/CHECKPOINTS.md, docs/ERRORS.md, docs/superpowers/plans/README.md` (`tools/verify.py:392-393`) for backticked path tokens; plus plan/spec filename resolution (`tools/verify.py:415-419`). |
| `GATES` registry (6), `select_staged`, `staged_paths` | registry | `tools/verify.py:454-477` | `tests` is the only non-fast gate (`tools/verify.py:458`); `select_staged` always includes `hygiene` and expands to all six when the gate layer itself changed (`tools/verify.py:464-471`). |
| CLI: `--fast --staged --only NAME --list --selftest` | argparse | `tools/verify.py:478-523` | `--only` with unknown name → exit 2 with named error (`tools/verify.py:497-501`); `--selftest` prints `caught`/`SURVIVED` per gate (`tools/verify.py:510-516`); final line prints `N passed, M failed, K not run` (`tools/verify.py:520-522`). |
| `Makefile` targets + hooks | make / sh | `Makefile:24-36`, `.githooks/pre-commit`, `.githooks/pre-push` | `verify`/`verify-fast`/`verify-selftest` thin wrappers; `hooks` sets `core.hooksPath` and lists gates; pre-commit runs `--staged`, pre-push refuses `refs/heads/main` unless `CGO_ALLOW_MAIN=1` (proven in checkpoints at `docs/CHECKPOINTS.md:159-203`). |

### 2.6 `docs/api.md` + `docs/architecture.md` — the product side the ledger points at (not part of the ledger, but part of its contract surface)

| Export | Location |
|---|---|
| `GET /api/integrity`, `/api/report/export`, `/api/audit/report`, `/api/audit/list`, `/api/results`, `/api/duel` (§2.5 above) | `docs/api.md:66-104` |
| `pkg/api` = HTTP+SSE surface, `pkg/results` = `Scan` of `data/runs`, `pkg/figures` = `ProvenanceSHA/Hash8` injected "into the SVG, the exported report and `/api/integrity`" | `docs/architecture.md:10-21`, `docs/architecture.md:76-83` |

The ledger never calls the API; the API never writes the ledger directly. The binding is prose: `docs/architecture.md:6` + the verifier-signature block at `docs/VERIFY.md:93-115`.

### 2.7 `.gitignore` + seal anchor — the persistence boundary

| Export | Location | Notes |
|---|---|---|
| `/book/` ignored | `.gitignore:69` | The frozen corpus is gitignored — therefore git is **not** its persistence layer (see §5). |
| `data/runs/`, `data/logs/` ignored | `.gitignore:21-22` | Live runs are not persisted either. |
| `docs/frozen-manifests.sha256` (159 lines `<sha256>  book/data/runs/<run>/manifest.json`, LF) | referenced at `docs/VERIFY.md:26`, `tools/verify.py:47-62` | The only committed anchor over the ignored corpus. |

---

## 3. Dependencies

### 3.1 Outgoing (what the layer needs) — Python stdlib + git + Markdown

* **`tools/verify.py:3-5` imports only `argparse, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time`** (plus `dataclasses`, `pathlib` at `tools/verify.py:5-6`). No PyYAML, no third party — the founding spec says "No YAML registry" (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:281-283`) and the plan says "Python 3.11+, standard library only" (`docs/superpowers/plans/2026-10-27-cgo-engineering-system.md:274`). R1/R2 (SLSA) require provenance formats to be interoperable and unambiguous — a stdlib-only Python producer satisfies the "format acceptable to the ecosystem" bar only if it emits the exact schema; today's gate emits free-text `Result.detail` strings (`tools/verify.py:14-18`), not structured attestations (see §4.1).
* **Git plumbing:** `git ls-files -z` (`tools/verify.py:364`), `git diff --cached --name-only -z` (`tools/verify.py:473-477`), `git rev-parse --show-toplevel` in the hooks. R5 (gitignore) says plumbing tools "read `gitignore` patterns specified by command-line options, or from files specified by command-line options" while porcelain (`git status`, `git add`) uses the full precedence chain (R5, DESCRIPTION). The hygiene gate reads patterns from the working-tree `.gitignore` plus the index — it does **not** read `$GIT_COMMON_DIR/info/exclude` or `core.excludesFile` (R5, DESCRIPTION), so a pattern only in a user's global ignore file is invisible to the gate. That is a stated boundary, not a bug: the gate is a *tracked-file* scanner (`tools/verify.py:344-361`).
* **Subprocess targets (not imports):** `go vet` / `go test` (`tools/verify.py:202`), `go build ./cmd/cgo` (`tools/verify.py:269`), `book/tools/audit-archive.py` via `sys.executable` (`tools/verify.py:68`). The gate never shells out to `tc`, the bench, or the network — "no shell" (`tools/verify.py:2`) and `subprocess.run` with argument lists, `check=False` (`tools/verify.py:27-30`).
* **Markdown:** no parser dependency. Link extraction is regex over backticked tokens (`tools/verify.py:394-403`); there is no Markdown AST, no frontmatter, no YAML anywhere in the ledger layer (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:508-509`).

### 3.2 Incoming (who needs the layer)

* **Human/agent sessions** must read `AGENTS.md → CHECKPOINTS → VERIFY → plan-index row → artifact` before acting (`docs/CHECKPOINTS.md:6-13`). This is the single largest consumer.
* **`Makefile` + `.githooks/*`** invoke the gate runner (`Makefile:24-36`); a commit must pass `--staged`, a push must pass the full set (`docs/superpowers/plans/2026-10-27-cgo-engineering-system.md:624-687`).
* **Product code does NOT import the ledger.** No Go file imports `tools/verify.py` or parses `docs/VERIFY.md`; the binding is citation (a number in prose cites a VERIFY row — `docs/VERIFY.md:3-5`), not a function call. `docs/api.md` and `docs/architecture.md` describe the product's Go packages (`pkg/api`, `pkg/campagne`, `pkg/results`, … at `docs/architecture.md:10-21`) but are never invoked by the gate except through the `links` path-existence check (`tools/verify.py:391-429`).
* **The book build** is the reverse direction: `book/tools/audit-archive.py` is invoked *by* the `frozen` gate (`tools/verify.py:63-81`), and book audits (`audit-biblio`, `audit-renvois`, `audit-nombres`) are run *beside* it in checkpoints (e.g. `docs/CHECKPOINTS.md:713-721`). The book never reads the gate.

### 3.3 Licence

Repo root is MIT ("Copyright (c) 2026 Meteolink contributors", per `docs/research/2026-11-07-module-metrics.md:123` citing `LICENSE:3` — **unverified — primary source not found** in this pass; the licence text itself was not read here). Any standalone module keeps MIT (see §7).

---

## 4. Primary-source requirements — what SLSA / CycloneDX / semver / gitignore *legitimately* require of a docs/persistence module

### 4.1 R1/R2 — SLSA Provenance: the ledger's `VERIFY` rows are a hand-rolled provenance, and SLSA says what is missing

R1 defines provenance as "the verifiable information about software artifacts describing where, when, and how something was produced" (R1, Purpose) and requires, for SLSA Build L1, two things: `buildDefinition` and `runDetails` (R1, Provenance table). Fields:

* `buildDefinition.buildType` — a URI identifying the build template (R1, BuildDefinition table). `externalParameters` (MUST be complete at L3 — R1, BuildDefinition: "They MUST be complete at SLSA Build L3") are the untrusted inputs and MUST be included in provenance and verified downstream. `resolvedDependencies` are `[ResourceDescriptor]` with `uri` + `digest` (`sha256`, `gitCommit`, …) (R1, ResourceDescriptor schema).
* `runDetails.builder.id` — a URI identifying "the transitive closure of the trusted build platform" and is "the sole determiner of the SLSA Build level" (R1, Builder table). `metadata.invocationId` / `startedOn` / `finishedOn` identify the specific invocation (R1, BuildMetadata table).

R2 grades three requirements upward: **Provenance Exists** (L1+: MUST generate provenance that "unambiguously identifies the output package by cryptographic digest and describes how that package was produced"; RECOMMENDED to use SLSA Provenance format — R2, Provenance Exists), **Provenance is Authentic** (L2+: signature validation, control-plane generation, `subject` excepted — R2, Provenance is Authentic), **Provenance is Unforgeable** (L3+: signing secrets in a secure management system, not accessible to the build steps, every field generated/verified by the platform — R2, Provenance is Unforgeable).

Mapping to CGO today:

* `docs/VERIFY.md` rows **already carry the R1-shape fields informally**: Claim ≈ `subject`, Evidence (command/path) ≈ `buildDefinition.externalParameters`, Date/Host ≈ `runDetails.metadata.startedOn` + a `builder.id` proxy, and the verifier signature block (`docs/VERIFY.md:93-115`) ≈ `runDetails.builder.id` + per-gate `runDetails.metadata`. The `frozen` gate's seal (`docs/frozen-manifests.sha256`, `tools/verify.py:47-62`) is a `resolvedDependencies` analogue — 159 `<sha256> <path>` lines, exactly the ResourceDescriptor `(uri, digest)` pair (R1, ResourceDescriptor), and `docs/architecture.md:76-83` says `ProvenanceSHA` / `ProvenanceHash8` are "injected in the SVG, the exported report and `/api/integrity`" — a `subject`-digest carrier (R1, Purpose).
* What CGO lacks against R1/R2: **(a)** no `predicateType` — there is no `"https://slsa.dev/provenance/v1"`-equivalent URI declaring which predicate a VERIFY row satisfies (R1, `predicateType` "MUST" use the URI, parse rules: consumers MUST ignore unrecognized fields, predicateType includes major version); **(b)** no authenticity — rows are unsigned Markdown, so R2-L2 "Provenance is Authentic" (signature validation, tamper evidence after the build) is not met; the slice-11 checkpoint says this plainly: "this is a seal, not a signature; a writer who can rewrite the anchor too is out of scope" (`docs/CHECKPOINTS.md:533-534`); **(c)** the `builder.id` is a hostname string (`DESKTOP-B6K9UCM`, `docs/VERIFY.md:24`), not a URI that "SHOULD resolve to documentation explaining scope, claimed level, and accuracy guarantees" (R1, Builder `id`), and the same host string is reused for builder and verifier passes, so "Consumers MUST accept only specific signer-builder pairs" (R1, Builder) cannot be enforced.

A standalone ledger tool therefore MUST, per R1 parsing rules, tolerate unrecognized fields (forward-compat), MUST version its predicate by major version in its type URI, and MUST NOT claim L2/L3 (authentic/unforgeable, R2) until it has signature + control-plane generation. Until then, claiming "SLSA L1" is itself overclaiming unless the row carries subject-digest + how-produced fields at R1-`buildDefinition` completeness ("best effort" at L1 — R2, Provenance Exists, Completeness note).

### 4.2 R3 — CycloneDX SBOM: the corpus + ledger as an inventory, and where CGO already is one

R3's SBOM capability page says CycloneDX "represents comprehensive inventories of software components, dependencies, and relationships… in a machine-readable format" to identify "outdated components, licensing conflicts, or vulnerabilities" (R3, Introduction), and the v1.5 JSON reference makes the contract concrete: top-level `bomFormat: "CycloneDX"` (MUST be exactly that — R3, bomFormat) + `specVersion` (R3, specVersion), a `serialNumber` that "SHOULD have a unique serial number, even if the contents … have not changed" (R3, serialNumber, RFC-4122 UUID), a monotonically-`version: 1` bump on modification ("the version of the BOM SHOULD be incremented by 1" — R3, version), `metadata.timestamp` (R3, timestamp date-time), and `components[]` each with `type` (application/framework/library/container/platform/operating-system/device/device-driver/firmware/file/machine-learning-model/data — R3, component type), `name` (REQUIRED), `version`, `hashes[]` (`alg` ∈ MD5/SHA-1/SHA-256/…/BLAKE3 + `content` — R3, hashes), `licenses[]` (SPDX id or named — R3, licenses), `supplier` with name/url/contact (R3, supplier), `scope` ∈ required/optional/excluded (R3, scope), and unique `bom-ref` per component (R3, bom-ref).

CGO mapping:

* The frozen corpus (159 runs / 528 lines / 22 cells — `docs/VERIFY.md:24`) **is** an inventory in all but format: `docs/frozen-manifests.sha256` gives `(path, sha256)` per manifest (R3, hashes); per-run `manifest.json` files give component-like records; `pkg/figures` `ProvenanceSHA/Hash8` (`docs/architecture.md:78-83`) is a digest carrier. A standalone persistence module that emits a CycloneDX BOM over `{Go module, bench scripts, book corpus, .dist artefacts}` would satisfy R3's `bomFormat`/`specVersion`/`components[].hashes|licenses|supplier|scope` fields — today CGO has the data but emits none of the envelope fields (no `serialNumber` UUID, no `metadata.timestamp`, no `specVersion`).
* `tools/verify.py` hygiene (forbidden paths, token/private-key scanning — `tools/verify.py:344-388`) is an SBOM-adjacent *license/exposure* check (R3, Expected Outcomes: "rapid identification of outdated components, licensing conflicts, or vulnerabilities"). But it is a denylist, not an inventory: it proves absence of known-bad shapes (`ghp_…`, `AKIA…`, `xox[baprs]-` — `tools/verify.py:308`), never enumerates components. R3 requires the positive form: a `components[]` list with names + versions + hashes + licenses (R3, components).
* The plan index `docs/superpowers/plans/README.md:27-58` is a *work* inventory (slices with spec/plan/status/depends-on), not a *component* inventory; conflating the two is the kind of second-source-of-truth the founding spec forbids ("A second copy is a defect waiting to drift", `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:199-200`).

### 4.3 R4 — Semantic Versioning: the ledger formats are a public API that has no version

R4 §1: "Software using Semantic Versioning MUST declare a public API. This API could be declared in the code itself or exist strictly in documentation. However it is done, it SHOULD be precise and comprehensive." §§6–8: PATCH for backward-compatible bug fixes, MINOR for backward-compatible functionality (+ deprecations), MAJOR for backward-incompatible API changes, with minor/patch reset on bumps (R4, §§6–8). §3: "Once a versioned package has been released, the contents of that version MUST NOT be modified" (R4, §3). §4: 0.y.z MAY change and is not stable (R4, §4). FAQ: "v1.2.3 is not a semantic version" — `v`-prefixed strings are tag names, the SemVer is `1.2.3` (R4, FAQ).

CGO mapping:

* The VERIFY row schema, CHECKPOINT entry schema, and the six-gate CLI (`--fast/--staged/--only/--list/--selftest` — `tools/verify.py:479-484`) **are** a public API in R4's sense: writers (builder sessions) and readers (verifier sessions) interoperate through them, and a breaking change (renaming a status, changing the verdict contract) would silently corrupt in-flight claims. Yet neither the schemas nor the gate layer declares a version: `tools/verify.py` has `MAX_GATE_LINES = 650` (`tools/verify.py:11`) but no `__version__` / `ledger_version`; `docs/VERIFY.md` grew `SUPERSEDED` from the founding five statuses (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:349-358` → `docs/VERIFY.md:7-18`) without a version bump or migration note recorded anywhere in `docs/CHECKPOINTS.md` or `docs/ERRORS.md` (grep: **unverified — primary source not found** for any "ledger v2" note).
* Per R4 §1/§4, a standalone tool must declare the API preciously: start at `0.y.z` (MAY change), move to `1.0.0` only when the row schemas + CLI + verdict contract are pinned as compatibility promises, and treat released `vX.Y.Z` tags as immutable (R4 §3). Go consumers get this from the module path + `vX.Y.Z` tag scheme (R6 — **unverified — primary source not found** in this pass; see `docs/research/2026-11-07-module-metrics.md:145-147` for the fetched version of the same point).

### 4.4 R5 — gitignore: the persistence boundary is real, and "tracked" is the wrong word for half the system

R5 DESCRIPTION: "A `gitignore` file specifies intentionally untracked files that Git should ignore. **Files already tracked by Git are not affected**" (R5, DESCRIPTION, emphasis added), and NOTES: "The purpose of gitignore files is to ensure that certain files not tracked by Git remain untracked. To stop tracking a file that is currently tracked, use `git rm --cached`" (R5, NOTES). Patterns are read from (high→low): command line, `.gitignore` files in the same/parent directories, `$GIT_COMMON_DIR/info/exclude`, `core.excludesFile` (R5, DESCRIPTION). Only version-controlled, shared patterns belong in a committed `.gitignore` (R5, DESCRIPTION, "Which file to place a pattern in").

CGO mapping:

* `.gitignore:69` (`/book/`) plus `data/runs/`, `data/logs/` (`.gitignore:21-22`) and `*.docx / *.pdf` (`.gitignore:66-68`) mean **the measurement corpus, the thesis book, and all live-run output are intentionally untracked** — correctly, per R5's purpose. The persistence layer therefore cannot be "git history": git literally does not carry the frozen corpus, and a fresh clone has no `book/` so `frozen`/`provenance` SKIP (exit 2 — `tools/verify.py:65-66`, `docs/CHECKPOINTS.md:176-178`). R5's "Files already tracked are not affected" cuts the other way for the 18 legacy `backup/*.tar.gz` archives: they were tracked *before* the `backup/` rule, so ignoring `backup/` does not untrack them — which is why the hygiene gate needs an exact-path `LEGACY_ARCHIVES` allowlist (`tools/verify.py:316-323`) rather than a pattern fix. To stop tracking them you would need `git rm --cached` (R5, NOTES) — a human decision still open (`docs/CHECKPOINTS.md:153-155`).
* The `.gitignore` in this repo also carries product decisions disguised as ignore rules: `DIST` (`web/frontend/dist/*` with `!web/frontend/dist/index.html` — `.gitignore:11,35`), `vmlinux.h` (`.gitignore:17`), `kit/chr/cache/` (`.gitignore:82`). A standalone tool must distinguish *persistence policy* ("never commit `book/`", `AGENTS.md` Never-commit list — `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:227-229`) from *build-artifact exclusion* (no need to version `vmlinux.h`). R5 helps: only patterns that "all developers will want to ignore" belong in the committed file (R5, DESCRIPTION); everything else is `info/exclude` / `core.excludesFile` territory.

### 4.5 What the primaries do NOT require (avoid citation theatre)

* Neither R1 nor R2 requires a Markdown ledger, a `VERIFIED` status, or a checkpoint format — those are CGO-local invention (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:32-36`). SLSA requires subject+how-produced fields *however they are carried*; Markdown can carry them, but the carry does not inherit SLSA's authenticity (R2-L2) without signatures (see §4.1).
* R3 does not require the inventory to be gitignored or tracked — `scope: excluded` (R3, scope) is the mechanism for "present but not required"; CGO uses gitignore instead, which is a different mechanism (R5 tracking vs R3 scope) and a standalone tool must not conflate them.
* R4 does not require semver on *documents*; it requires it on *the code/tool's public API*. Versioning `docs/VERIFY.md` itself is a design choice the module may make (ledger format version), not an R4 mandate — but once the tool declares a ledger format, R4 §1 requires that declaration to be precise.
* The "ledger conventions" named in the brief (append-only + no-delete + checkpoint-at-handoff) are **CGO-local** — the founding spec attributes them to the AIR/DIS synthesis (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:19-36`), and the AIR/DIS primaries were not fetched in this pass (**unverified — primary source not found**). Do not cite R1–R5 for the append-only rule itself.

---

## 5. Gap analysis — why docs *are* the persistence contract (and where that breaks)

Ordered by severity. Local file:line is the evidence; the primary side cites R# or is marked unverified.

1. **Docs are the only persistence layer because git deliberately does not carry the evidence.** `.gitignore:69` ignores `/book/`; `.gitignore:21-22` ignores `data/runs/` and `data/logs/`. R5 says gitignore makes files "remain untracked" (R5, NOTES) — so `git log` cannot reconstruct the 159-run corpus, the thesis, or any live run. The `docs/VERIFY.md:1-5` rule ("The row is the source of truth; prose is not") exists *because* there is no other store: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:378` records "SQLite/WAL journaling — CSV append + manifest remains" as refused. Every number in `docs/api.md` (`docs/api.md:79` integrity counts) and `docs/architecture.md` (`docs/architecture.md:76-83` seal) resolves to the same ledger rows, not to a query. If the ledgers were deleted, the *interpretation* of the CSVs would be lost even though the CSVs survived — which is why `docs/VERIFY.md:83-91` forbids deleting `REFUTED`/`SUPERSEDED` rows: they are history, and history has no other home.
2. **The corpus seal is self-referential without the ledger.** The `frozen` gate checks manifests against per-run SHA-256s (`tools/verify.py:60-62`) **and** against the committed anchor `docs/frozen-manifests.sha256` (`tools/verify.py:47-62`). That anchor was added in slice 11 precisely because the pre-slice-11 row read "no hash above the 159 manifests" and the verifier narrowed the claim ("unchanged since the seal" is not provable — `docs/VERIFY.md:25`). R1 calls the `resolvedDependencies` digest list the completeness vehicle (R1, BuildDefinition); CGO's anchor is that list. But the anchor itself is a plain file: "a writer who can rewrite the anchor too is out of scope" (`docs/CHECKPOINTS.md:533-534`), and `docs/VERIFY.md:26` says so in the row: "Proves a static anchor only: not a signature". Against R2, this is L1-at-best ("best effort" completeness, R2/Provenance Exists): it catches a *stale* manifest, not a *forged* one. The ledger is the only thing that names the limit — the seal format does not.
3. **The gate layer is capped, not complete — by design, and it is CGO's own admission test that says so.** `MAX_GATE_LINES = 650` (`tools/verify.py:11`) with `CEILING_FILES` counting gate code + hooks (`tools/verify.py:12`). The founding spec's §4.12 says "Adding a gate requires deleting something or raising the ceiling in a commit that names what was deleted" (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:483-484`). The ceiling is the AIR admission test made executable ("no artifact without a reader-facing failure" — `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:205-208`). Gap for a standalone tool: the ceiling measures *lines*, not *coverage* — a gate can stay under 650 lines while missing a whole class (there is deliberately no `counts` gate, no docs linter, no full mutation harness — `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:309-314`, with reasons). Whether that refusal list is correct at standalone scale is **unverified — primary source not found** (no external mutation-floor evidence was fetched; the DISanalyzer "~300 floor" is cited second-hand in the local spec at `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:25-30`).
4. **The verdict contract is unenforced for readers that never run the gate.** `verdict()` (`tools/verify.py:35-36`) is asserted by `verdict_contract` (`tools/verify.py:255-261`) and by the `exit` gate (`tools/verify.py:285-286`), but a human reading `docs/VERIFY.md` in a browser never invokes it. The ledger's protection against SKIP-being-read-as-PASS is purely prose: git history shows rows where "2 is not green" (`docs/VERIFY.md` and `AGENTS.md` pointer at `docs/architecture.md:6`) was load-bearing — the slice-11 verifier recorded that a partial-`book/` SKIP "says 'book/ absent' inaccurately" while exiting 2 (`docs/VERIFY.md:26`). R2 would call this an authenticity gap (R2, Provenance is Authentic: consumers MUST validate the attestation) — CGO has no validation surface, only a string.
5. **No SBOM, no component inventory, no licence surface.** `hygiene` is a *absence* check (no keys/tokens/tailnet/forbidden paths — `tools/verify.py:334-388`); there is no positive `components[]` enumeration with `name`/`version`/`hashes`/`licenses`/`supplier`/`scope` (R3, components). The Go dependencies (`go.mod` `require` lines: `fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys` — per `docs/research/2026-11-07-module-metrics.md:113` citing `go.mod:5-10`) and the Node side (`web/frontend/package.json`) are inventoried nowhere in the repo. Any standalone release that claims supply-chain posture must emit the envelope (R3: `bomFormat`, `specVersion`, `serialNumber`, `version`, `metadata.timestamp`, `components[]`) — today the ledger can name a *corpus* digest but cannot name a *dependency* digest.
6. **The citation surface is the API surface, and it is unversioned.** `GET /api/audit/report` (`docs/api.md:100`), `ProvenanceSHA/Hash8` (`docs/architecture.md:78-83`), and the verifier-signature block (`docs/VERIFY.md:93-115`) are how a number leaves the repo. The slice-9 spec says "For any number leaving the repository, a separate pass is mandatory" (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:192-193`) — but there is no `citation: <row-id, digest, date>` format: citations are free-text inside checkpoint prose (e.g. "slice 10 closed" at `docs/CHECKPOINTS.md:93-112`). R4 §1 requires the public API to be declared precisely; today the citation format is declared *nowhere*. The frontend `citation.ts` surface cited in checkpoints (`docs/CHECKPOINTS.md:712-721`: "CITER CE RAPPORT builds the citation from report fields only (id, site, link/provider, timestamp, direction, full `source_sha256`)") is the closest thing — but it is a UI copy-button, not a verifiable attestation.
7. **Versioning of the ledger itself is silent.** `SUPERSEDED` was added to the status enum (`docs/VERIFY.md:17-18`) after the founding spec's five statuses (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:349-358`), and `MAX_GATE_LINES` went 600 → 650 (`docs/CHECKPOINTS.md:511-521`) when the seal was added. Both are API changes with no major/minor marker (cf. R4 §§7–8: MAJOR for incompatible changes, MINOR for backward-compatible additions). A reader joining mid-stream cannot tell which rows predate which contract. There is no ledger-format version field in any of the three ledgers.
8. **Markdown is the carrier; there is no machine-readable ledger.** The spec refused "Generated ledgers from YAML" and "Generated docs" (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:508-509`, §5 refusal table) — "Markdown is the carrier". Consequence: the *only* machine checks over the ledgers are `links` (do the named paths exist — `tools/verify.py:391-429`) and `hygiene` (no secrets — `tools/verify.py:334-388`). No schema validation over the VERIFY table, no "every VERIFIED row has a command+host+date" check, no "every BLOCKED row names a condition" check. R1 parsing rules ("Consumers MUST ignore unrecognized fields" — R1, Parsing rules) are meaningless when there is no field structure to ignore.

---

## 6. Target tool decomposition — three tools, not one

The requirement names three tools. The decomposition is justified by the three different readers and the three different write patterns:

* **Writer of claims** (builder session) — appends rows; never rewrites history.
* **Reader that runs gates** (verifier session / hook) — executes; never edits.
* **Reader that quotes a number out of the repo** (thesis / report / UI) — cites; never writes.

Collapsing them into one binary loses the SLSA separation R1 requires between `buildDefinition` (inputs), `runDetails.builder.id` (who ran), and the consumer's verification (R1, Model; R2, split of duties between producer and build platform). Three small tools with a thin shared schema is the minimal complete decomposition; one tool with three subcommands is the same thing with a marketing name. The schemas are the API (R4 §1) — declare them first.

### 6.1 Ledger tool (`ledger`) — owns the three Markdown ledgers as data

* **Owns:** parsing, validating, and appending `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, plus the plan-index slice map (`docs/superpowers/plans/README.md:27-58`) as a read-only view. It is the only writer.
* **Contract (illustrative; the shape is the deliverable, per R4 §1):**
  ```text
  ledger add-row --claim "..." --status VERIFIED|RECORDED|PENDING|BLOCKED|REFUTED|SUPERSEDED --evidence "<command or path>" --date YYYY-MM-DD --host ...
  ledger move-row <id> --to VERIFIED|BLOCKED ...  # BLOCKED→PENDING requires the named condition text; VERIFIED requires non-empty evidence
  ledger checkpoint --focus "..." --branch ... --head ... --changed ... --proven ... --unproven ... --next ...
  ledger error --title ... --symptom ... --cause ... --fix ... --proof ... --lesson ...
  ledger validate          # schema + status-closed-set + "every BLOCKED names a condition" + "every VERIFIED names command+host+date" (gaps §5.2, §5.7)
  ledger signature ...     # writes the verifier-signature block shape (docs/VERIFY.md:93-115)
  ```
* **Why it is a tool and not a template:** today the rule "only a named command on a named host moves a row" (`docs/VERIFY.md:83-91`) is prose enforced by the verifier's memory. The ledger tool makes the move *structurally* impossible otherwise (an edit that changes a status without evidence fails validation) — that is the difference between a persistence contract and a style guide. R1's `externalParameters` MUST-be-verified-downstream (R1, BuildDefinition) maps to `Evidence` being a required field, not a convention.

### 6.2 Gate tool (`gate` / `verify`) — owns `tools/verify.py` + hooks, and nothing else

* **Owns:** the six gates + fault proofs + ceiling + verdict + hook selection (`tools/verify.py:454-523`, `tools/verify.py:478-523`). Real (non-fixture) verification belongs here.
* **Contract:** keep the current CLI shape (`verify [--fast|--staged|--only NAME]`, `--list`, `--selftest` — `tools/verify.py:479-484`, `Makefile:24-36`) because hooks and checkpoints already depend on it (`docs/CHECKPOINTS.md:159-203`). Add `--json` emitting the CycloneDX-shaped inventory for gate-relevant inputs (R3: `components[]` with `hashes`, `supplier`, `scope`) so that "which Go module version did `tests` run" is attestable next to the verdict — closing gap §5.5 without a new gate.
* **What it does NOT own:** the ledgers. Today `hygiene` + `links` read the ledgers as *paths to scan* (`tools/verify.py:392-393`) — that stays. But the gate never edits them; the slice-11 precedent is correct: the gate *enforces* ("the gate only reads the anchor" — `docs/CHECKPOINTS.md:568-571`), the ledger *records*. Merging the two re-introduces the self-authored-verification failure the spec's Law 6 exists to prevent (`docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:190-193`).

### 6.3 Citation tool (`cite`) — owns "how a number leaves the repo"

* **Owns:** turning a `VERIFIED` ledger row + its artefact digest into a portable citation record a downstream reader can re-verify without the repo.
* **Contract (illustrative):**
  ```text
  cite <row-id>            # emits s<a>L<S>A>-shaped JSON (R1 predicate type) with:
                           #   subject: {name, digest: sha256}   (the artefact, R1 subject)
                           #   buildDefinition: {buildType, externalParameters: {claim,status,evidence,date,host}, resolvedDependencies: [{uri, digest}]}
                           #   runDetails: {builder: {id: ledger://host/...}, metadata: {invocationId, startedOn, finishedOn}}
  cite --verify <record>   # checks the digest + the status + the "2 is not green" rule (tools/verify.py:35-36)
  ```
* **Why it is a tool and not a doc paragraph:** the R1 requirement is that consumers "can verify that the artifact was built according to expectations" and "can rebuild the artifact, if desired" (R1, Purpose). A copy-pasted paragraph in a thesis cannot be verified; a JSON record with `subject.digest` + `resolvedDependencies` can (R1, ResourceDescriptor). The frontend "CITER CE RAPPORT" action (`docs/CHECKPOINTS.md:712-721`) already builds this shape from report fields — the tool generalises it from one report to *any* VERIFY row, and the UI then copies the tool's output instead of hand-assembling strings.
* **Honesty note (R2):** until the citation record is signed by a builder identity (R2-L2 Authentic, R2-L2 Authentic), `cite --verify` validates *shape + digest consistency*, not authenticity — the record must say so in an `authenticity: unsigned-local-seal` field rather than implying L2. The slice-11 precedent ("seal, not signature" — `docs/CHECKPOINTS.md:533-534`) is the model.

### 6.4 Interaction diagram (text)

```text
builder session ──append──▶ ledger (VERIFY/CHECKPOINTS/ERRORS)
   │                           ▲            │
   │                           │validate    │cite row
   ▼                           │            ▼
gate (verify.py) ──reads──▶ ledgers   cite ──emits──▶ downstream reader
   │                                   (SLSA JSON)
   ├──frozen──▶ docs/frozen-manifests.sha256 + book audit
   ├──provenance/hygiene/links──▶ tracked files + .gitignore patterns
   └──tests/exit──▶ go toolchain + CLI binary
```

The ledgers sit between the gate and the citation tool so that neither gate output nor citations are self-authored (Law 6, `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:190-193`).

---

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it. Order is ledger-first because the ledger is the contract; gate and citation follow.

1. **Slice 1 — Freeze + declare the ledger API.** Snapshot the three schemas (§2.1–2.3) + the plan-index shape (§2.4) into a versioned `ledger.schema` (initial `0.1.0` per R4 §4 — unstable until `1.0.0` per R4 §5); declare the CLI surface (`ledger validate`, `gate verify`, `cite`) as the public API (R4 §1 — "MUST declare a public API"). Port every existing VERIFY row, CHECKPOINT entry, and ERROR entry through `ledger validate` unchanged. Acceptance: `ledger validate` exits 0 on the current tree with zero suppressions; README lists every §2 contract; `gate verify --list` still prints six gates with proofs (`tools/verify.py:493-496`).
2. **Slice 2 — Ledger tool (validate + append).** Implement `ledger add-row / move-row / checkpoint / error / validate` (§6.1) in Python stdlib only (matching `tools/verify.py:3-5`), with the move rules from `docs/VERIFY.md:83-91` enforced structurally (VERIFIED requires evidence+date+host; BLOCKED requires a named condition; REFUTED/SUPERSEDED rows are append-only history). Acceptance: seeded unit tests mutate statuses without evidence and require failure; a real checkpoint written through the tool parses and round-trips byte-identically (modulo the new entry).
3. **Slice 3 — Gate tool (decouple + JSON inventory).** Split `tools/verify.py` into importable gate functions (the `gate_*` / `prove_*` pairs already have this shape — `tools/verify.py:63-451`) behind the CLI; add `--json` emitting the R3-shaped inventory (at minimum: `bomFormat`, `specVersion`, `serialNumber` UUID, `version`, `metadata.timestamp`, `components[]` for `go.mod` + `web/frontend/package.json` deps + the frozen anchor). Keep `MAX_GATE_LINES` + hooks behaviour unchanged (`tools/verify.py:11-12`, `Makefile:24-36`). Acceptance: `--selftest` prints six `caught` (`tools/verify.py:510-516`) and the JSON validates against the CycloneDX v1.5 reference fields (R3: `bomFormat` MUST be `"CycloneDX"`, hashes `alg` ∈ enumerated set).
4. **Slice 4 — Citation tool (SLSA-shaped records).** Implement `cite <row-id>` / `cite --verify` (§6.3) emitting `predicateType`-versioned JSON (R1: type URI carries major version; R1 parsing rules: ignore unrecognized fields) for every `VERIFIED` row, starting with the frozen-corpus row and a sealed audit report (`docs/api.md:100` report shape is the subject). Migrate the frontend "CITER CE RAPPORT" copy to the tool's output. Acceptance: a citation generated for `docs/VERIFY.md:24`'s frozen row re-verifies against `docs/frozen-manifests.sha256` by digest (R1 `subject` + `resolvedDependencies`); the record carries an explicit `authenticity: unsigned-local-seal` marker (R2 — no L2 claim until signed).
5. **Slice 5 — First tag + in-repo migration.** Resolve module path, tag `v0.1.0`, publish; migrate `Makefile` targets and `.githooks/*` to call the tools (no `replace` directives, no Markdown edits by hand — ledgers are written only via `ledger`). Acceptance: `gate verify` → `6 passed, 0 failed, 0 not run` on the real tree (or named SKIP → 2, per `tools/verify.py:520-523`); `ledger validate` green; one out-of-tree consumer has imported the module (or the slice stays unclosed — see §8).

Each slice is independently reviewable; Slices 2–4 change behaviour visible to writers and each should be verifier-re-run before its `VERIFIED` row is claimed (per `docs/superpowers/specs/2026-10-27-cgo-engineering-system-design.md:190-193`).

---

## 8. Anti-toy criteria (what would make this a toy — and the bar for "not a toy")

A docs/persistence module is a toy if it re-describes its own files instead of enforcing an external check. This extraction is **not** a toy only if all of the following hold:

1. **Independent verification per contract family.** Row-schema validation against a *second* implementation or property test (not the same regex used to write the rows); gate verdicts against deliberate mutations of each target (the existing `--selftest` six `caught` — `tools/verify.py:510-516` — plus the slice-11 seal proof that a consistent rewrite passes the audit but fails the seal, `docs/CHECKPOINTS.md:560-567`); citations against independently recomputed digests (e.g. `sha256sum -c docs/frozen-manifests.sha256` → 159 `OK`, `docs/VERIFY.md:26`). A module whose tests only re-read its own output is a toy.
2. **Provenance honesty (R1/R2).** Every citation record carries `subject.digest` + `buildDefinition` + `resolvedDependencies` (R1, Schema) and an explicit authenticity marker; no record claims R2-L2 "Authentic" or R2-L3 "Unforgeable" until the `builder.id` is a resolvable URI bound to a signer (R1, Builder; R2, Provenance is Authentic/Unforgeable). The slice-11 wording is the template: "static anchor only: not a signature" (`docs/VERIFY.md:26`). A tool that lets a Markdown edit mint a green w/o a command is a toy.
3. **No silent promotion.** `RECORDED` never becomes `VERIFIED` by repetition or confidence (`docs/VERIFY.md:87-88`); SKIP (exit 2) never renders as pass in any surface reading the gate (`tools/verify.py:520-523` prints all three counts). If the citation tool can emit a record for a `BLOCKED`/`PENDING` row without the row's condition inline, it is a toy.
4. **Append-only history is structural, not editorial.** `REFUTED`/`SUPERSEDED` rows and corrected checkpoints cannot be deleted or rewritten — the tool rejects the edit (gap §5.7), and `.gitignore`-adjacent policy (R5: "Files already tracked are not affected", R5 DESCRIPTION) is modelled as *tracked-vs-ignored* honestly: the tool reports what git carries vs what the corpus carries instead of pretending git is the store (`.gitignore:69`; §5.1).
5. **A positive inventory exists (R3).** The module enumerates components with `bomFormat`, `specVersion`, `serialNumber`, `version`, `metadata.timestamp`, and per-component `name`/`version`/`hashes`/`licenses`/`supplier`/`scope` (R3, bomFormat→components) for at least the Go module and the frozen corpus. A hygiene denylist alone (today's state, §5.5) is a scanner, not an SBOM — claiming SBOM without the envelope is a toy.
6. **Version discipline (R4).** A declared public API (ledger schemas + gate CLI + citation JSON) (R4 §1), tags treated as immutable (R4 §3), and releases phrased MAJOR/MINOR/PATCH (R4 §§7–9). A repo that edits a released tag, silently adds a status (as happened with `SUPERSEDED`, §5.7), or ships no declared API is a toy regardless of code quality.
7. **A real consumer + a real citation.** At least one out-of-tree importer/reader (not only `Makefile` + `.githooks/*` in this repo) has parsed a ledger / run a gate / verified a citation — otherwise the "modular tool" claim is decoration. (Cf. the metrics analogue at `docs/research/2026-11-07-module-metrics.md:232-233`.)

---
*Teams note: slices 1–5 are proposals in this research file only. No `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, `tools/verify.py`, or any product source was modified in this pass; implementation waits for human approval per `AGENTS.md` ("Stop and ask" — new dependency, scope growth, destructive actions). Ponytail note: what was skipped or not checked — the AIR/DIS primaries, the Go-modules reference, the in-toto/Rekor/Sigstore prior art, and the `LICENSE` text were not fetched this pass and are marked **unverified — primary source not found** where they matter (§0, §4.5); the CHECKPOINTS tail beyond `docs/CHECKPOINTS.md:809` was sampled, not line-by-line read; risk the user must know — the SLSA L2/L3 and SBOM claims above are requirements mapping, not compliance assertions, and no citation record yet carries a signature.*
