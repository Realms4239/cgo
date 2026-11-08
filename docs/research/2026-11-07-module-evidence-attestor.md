# Frozen corpus + provenance as standalone tool — primary-source research

*Date: 2026-11-07 · Scope: frozen corpus + provenance seal / header / valuation system (`docs/VERIFY.md`, `docs/frozen-manifests.sha256`, `pkg/results/*.go`, `pkg/figures/*.go`, `tools/verify.py` frozen+provenance gates, `book/tools/audit-archive.py` spec) read as if released tomorrow as independent CLI `cgo-attest`. Research only — no code changed.*

Primary sources fetched this pass: (R1) SLSA v1.0 Security levels (<https://slsa.dev/spec/v1.0/levels>); (R2) SLSA v1.0 Provenance predicate (<https://slsa.dev/spec/v1.0/provenance>); (R3) SLSA v1.0 Producing artifacts / Requirements (<https://slsa.dev/spec/v1.0/requirements>); (R4) SLSA v1.0 Distributing provenance (<https://slsa.dev/spec/v1.0/distributing-provenance>); (R5) SLSA v1.0 Verifying artifacts (<https://slsa.dev/spec/v1.0/verifying-artifacts>); (R6) in-toto Attestation Framework spec v1.2 (<https://github.com/in-toto/attestation/blob/main/spec/v1/README.md>) and repository overview (<https://github.com/in-toto/attestation>); (R7) CycloneDX SBOM capability (<https://cyclonedx.org/capabilities/sbom/>); (R8) Go `go.mod` file reference (<https://go.dev/doc/modules/gomod-ref>); (R9) Git Book “Signing Your Work” (<https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work>). Local corpus seal design: `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md`. Existing sibling research (containers, not duplicated): `docs/research/2026-11-07-module-results-figures.md` (R1–R6 there = RFC 8290 / Prometheus / Go modules / semver / `encoding/csv` / `embed`), `docs/research/2026-11-07-module-deploy-supplychain.md` (Nginx / Prometheus / OCI / semver / GoReleaser SBOM / SLSA / CycloneDX).

Method: read `docs/VERIFY.md:24-26,42` (frozen/seal rows), `docs/frozen-manifests.sha256` (159 lines), `pkg/results/csvrows.go:9-65`, `pkg/results/results.go:13-236`, `pkg/results/seed.go:15-79`, `pkg/results/hardware.go:7-12`, `pkg/figures/figures.go:15-156`, `pkg/campagne/valuation.go:12-134`, `tools/verify.py:46-191` (`seal_findings`, `gate_frozen`, `prove_frozen`, `gate_provenance`, `prove_provenance`), `docs/architecture.md:76-83` (freeze/manifest/provenance), `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-95` (seal design), `go.mod:1-3`, `.githooks/pre-commit:9`. `book/` is gitignored and absent in this worktree, so `book/tools/audit-archive.py` was not read directly — its contract is quoted from `tools/verify.py:63-81` and `docs/VERIFY.md:24-26` only. Every technical claim below cites `file:line` or `(R#)`; anything else is `unverified — primary source not found`.

## 1. Responsibility

Today the frozen-corpus + provenance system does three distinct jobs that a standalone tool must keep separate (conflating them is the main design risk):

1. **Content integrity — manifest → file.** `book/tools/audit-archive.py` verifies every run’s `manifest.json` against the files it names (file hashes, config JSON, `config_sha256`). The contract quoted in `tools/verify.py:63-81` is: exit 0 only with two `OK` lines; the verified shape is **159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells** (`docs/VERIFY.md:24`). The per-run manifests are the evidence chain (`docs/architecture.md:76-79`: writer freezes the run, CSVs become immutable, a SHA-256 manifest is written).

2. **Corpus seal — anchor → manifest.** `docs/frozen-manifests.sha256` holds one SHA-256 per run manifest (159 lines, sorted, `sha256sum`-compatible, `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`). `seal_findings` (`tools/verify.py:48-62`) fails closed on absent anchor, malformed line, changed manifest, missing run, or extra run. The seal is explicitly *a seal, not a signature*: “it detects any rewrite of a manifest, including one made together with the file that manifest covers. It does not defend against an attacker who can rewrite the anchor too” (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`; repeated `docs/VERIFY.md:26`). The corpus (`book/`) is gitignored (`.gitignore:68`, `docs/VERIFY.md:25`), so the tracked anchor outside `book/` is the only versioned root — chain anchor → manifest → file (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:20-23`).

3. **Provenance honesty — what a number came from.** Three sub-systems with different authority:
   - `provenance` gate (`tools/verify.py:147-192`): checks that live `aqm_eval.csv` headers belong to the frozen eras and that `valuation.json` sidecars carry allowed `basis`/`source` enums (`tools/verify.py:132-146`: `basis ∈ {measured, estimated}`, `source ∈ {ebpf, tcpinfo, ""}`, estimated+non-empty-source refused). The sidecar itself is written by `pkg/campagne/valuation.go:50-95` with a disclaimer that scenarios are contractual, not measurements (`pkg/campagne/valuation.go:36-44`, `docs/VERIFY.md:48`).
   - `pkg/results` cell provenance: `Group.Runs` lists only runs with ≥1 non-`invalid` row (`pkg/results/results.go:43-46,146-148`); `Scan` tolerates 17/18/19-column historic schemas by header name, never position (`pkg/results/csvrows.go:9-18`).
   - `pkg/figures` figure provenance: `ProvenanceSHA` / `ProvenanceHash8` (`pkg/figures/figures.go:41-61`) hashed only the lexically-last `aqm_eval.csv` and were injected into SVGs, the exported report, and `/api/integrity` (`docs/architecture.md:80-83`, `pkg/api/server.go:896,975,1022,1079`). This is the weakest link of the three (see §5).

What the system deliberately does **not** do: no signing key, no remote timestamp, no transparency log, no git LFS, no new dependency (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:90-95` “Out of scope”). `AGENTS.md` also forbids touching `book/data/runs`, the CSV schemas, and live bench runs without human approval. The standalone tool inherits those constraints: it verifies evidence, it never re-measures and never re-seals silently (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:38-40`: “The gate never writes the anchor”).

## 2. Exported surface (file:line)

Complete inventory of the evidence/provenance-relevant exported identifiers. Unexported helpers (`seal_findings` `tools/verify.py:48-62` is module-level but not a library API; `provenanceMeta` `pkg/figures/figures.go:63-79`; `runIDs` `pkg/results/results.go:53-63`) are not API.

### 2.1 `tools/verify.py` (the gate layer that becomes `cgo-attest`)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `seal_findings(root)` | func | `tools/verify.py:48-62` | Anchor parser + three-way diff (changed / missing / extra). Returns `[]string`, never writes. Malformed line → immediate finding (`tools/verify.py:55-56`). |
| `gate_frozen()` | func | `tools/verify.py:63-81` | Runs `book/tools/audit-archive.py`, requires exit 0 + ≥2 `OK` lines, then `seal_findings`. SKIP (exit 2 path) when `book/` or audit absent (`tools/verify.py:65-66`); FAIL on timeout (`tools/verify.py:69-70`). |
| `prove_frozen()` | func | `tools/verify.py:82-128` | Fault proof: control copy passes, whitespace-rewritten manifest caught by seal, mutated data file caught by audit. Proves both layers independently (`tools/verify.py:99-124`). |
| `gate_provenance()` | func | `tools/verify.py:147-169` | Frozen-era header set vs live headers (`tools/verify.py:151-156`), plus `valuation.json` enum check (`tools/verify.py:157-163`). SKIP when frozen corpus absent (`tools/verify.py:149-153`). |
| `prove_provenance()` | func | `tools/verify.py:170-191` | Fault proof: extra column, `basis=guess`, `estimated+ebpf` caught; `measured+ebpf` / `estimated+empty` / exact header are controls (`tools/verify.py:178-185`). |
| `valuation_problem(obj)` | func | `tools/verify.py:138-146` | Pure enum validator — the only piece already shaped as a reusable predicate. |
| `header_problems(...)` / `csv_header(...)` | funcs | `tools/verify.py:133-137` | First-line header reader; exact-match against frozen era set. |

### 2.2 `docs/frozen-manifests.sha256` (the anchor itself)

Tracked file, 159 lines, format `<sha256>  book/data/runs/<run>/manifest.json` (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`). Verified properties (`docs/VERIFY.md:26`): 159 lines, sorted, unique, every path under `book/data/runs/` and present, every hash matches; `git ls-files --eol` → `i/lf w/lf`. Not a signature — a static anchor (`docs/VERIFY.md:26` “Proves a static anchor only: not a signature”).

### 2.3 `pkg/results` (aggregation provenance)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ReadAQM(path)` | func | `pkg/results/csvrows.go:19-45` | Header-name CSV reader, `FieldsPerRecord=-1`, short rows kept. |
| `ColIndex(header, name)` | func | `pkg/results/csvrows.go:49-56` | Trimmed exact-match lookup, `-1` if absent. |
| `Field(header, row, name)` | func | `pkg/results/csvrows.go:59-65` | Named field of one row, `""` if absent. |
| `Group` | struct | `pkg/results/results.go:13-47` | Cell `profile\|qdisc\|cc\|direction`; `Count/Quarantined/Degraded`; medians+IQRs; valid-only sub-struct; `Runs` sorted provenance (`pkg/results/results.go:46`). |
| `Scan(dataDir, runFilter)` | func | `pkg/results/results.go:65-236` | Glob `*/aqm_eval.csv`; skips `run-smoke*`; silent-skip on unreadable/missing-header (`pkg/results/results.go:88-100`). |
| `SeedRuns(root)` | func | `pkg/results/seed.go:24-79` | Extracts embedded `runs-frozen.tar.gz` only when no CSV exists; anti-traversal (`pkg/results/seed.go:43-54`). **Not** part of `cgo-attest` (see §5.5). |
| `HardwareRecommendation(bq, profile)` | func | `pkg/results/hardware.go:7-12` | Hardcoded FR vendor string. **Excluded** from `cgo-attest` (see §5.5). |

### 2.4 `pkg/figures` (figure provenance)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ProvenanceSHA(dataDir)` | func | `pkg/figures/figures.go:41-53` | SHA-256 of lexically-last `aqm_eval.csv` only. **Deficient** (§5.3) — to be replaced by corpus manifest hash. |
| `ProvenanceHash8(dataDir)` | func | `pkg/figures/figures.go:56-61` | First 8 hex of `ProvenanceSHA`. Short label, not a proof. |
| `Generate(dataDir, outDir)` | func | `pkg/figures/figures.go:17-37` | Calls `results.Scan`, writes `small_p95.svg` + `scatter.svg` with non-reproducible `provenanceMeta` (`pkg/figures/figures.go:63-79`). Not part of `cgo-attest` core. |
| `provenanceMeta` (unexported) | — | `pkg/figures/figures.go:63-79` | Embeds `time.Now().UTC()` + local path + hardcoded creator — the reproducibility bug (§5.3). |

### 2.5 `pkg/campagne/valuation.go` (sidecar provenance)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ValuationSidecar` / `ValuationLabels` / `EventValuation` | structs | `pkg/campagne/valuation.go:14-44` | JSON sidecar schema; `Disclaimer` is load-bearing honesty text (`pkg/campagne/valuation.go:64-68`). |
| `WriteValuation(dir, ev, labels)` | func | `pkg/campagne/valuation.go:99-111` | Idempotent per-cell writer; not part of the verifier, but the verifier’s `gate_provenance` validates its output. |

### 2.6 `book/tools/audit-archive.py` (opaque, contract only)

Not present in this worktree (`book/` gitignored). Contract from `tools/verify.py:63-81` + `docs/VERIFY.md:24-26`: two `OK` lines, exit 0 ⇒ `OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid` and `OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents`. Its internal manifest→file checks are `unverified — primary source not found` in this pass (file not read).

## 3. Dependencies

### 3.1 Outgoing (what `cgo-attest` needs)

- `tools/verify.py:1-6` — Python 3 stdlib only (`argparse`, `hashlib`, `json`, `os`, `re`, `shutil`, `subprocess`, `sys`, `tempfile`, `time`, `dataclasses`, `pathlib`). Docstring: “Standard library only; no network; no shell” (`tools/verify.py:2`). Zero third-party imports — the extraction blocker of `pkg/results`→`pkg/metrics` (§5.1 of sibling results-figures research) does **not** exist here, because the seal/seal-findings/header/valuation layer never imports `pkg/results` or `pkg/metrics`.
- `pkg/results/csvrows.go:3-7` stdlib `encoding/csv`, `os`, `strings`. `pkg/results/results.go:3-10` stdlib + internal `pkg/metrics`. `pkg/figures/figures.go:3-13` stdlib + internal `pkg/results`. `pkg/campagne/valuation.go:3-10` stdlib + internal `pkg/metrics`, `pkg/model`.
- `go.mod:1-3`: module `github.com/Realms4239/cgo`, `go 1.25`. Per (R8), a module path is the unique identifier combined with the version, and the prefix of every package path in it; the `go` directive is a mandatory minimum since Go 1.21 (R8, `go` directive notes). A standalone `cgo-attest` needs its own module path chosen before first publish (R8, module notes). Deps `fyne`/`bubbletea`/`lipgloss`/`golang.org/x/sys` (`go.mod:5-10`) are not needed by the frozen/provenance layer.
- **No `syscall`, no network, no clock dependence** in the seal/header/valuation layer: `seal_findings` is pure file hashing (`tools/verify.py:48-62`); `valuation_problem` is pure JSON validation (`tools/verify.py:138-146`). The only clock (`time.Now` in `pkg/figures/figures.go:77`) and the only subprocess (`audit-archive.py` in `tools/verify.py:68`) are both removable/testable seam points (§6).

### 3.2 Incoming (who needs the frozen/provenance layer)

- `tools/verify.py` itself: `gate_frozen` is called by `--only frozen` and by the full six-gate run (`tools/verify.py:455-462` registry, `tools/verify.py:517-523` verdict 0/1/2). The pre-commit hook runs `python tools/verify.py --staged`, and any change to `tools/verify.py` forces all six gates (`tools/verify.py:464-471` `select_staged`).
- `pkg/api/server.go:975` (`/api/integrity`: `sha256`+`hash8` via `figures.ProvenanceSHA/Hash8`), `:1022` (report header), `:1079` (evidence bundle), `:896` (compare-cache key).
- `pkg/results` callers: `pkg/api/server.go:770,942,982,1004`, `pkg/api/duel.go:77`, `pkg/api/translate.go:25` (all `results.Scan`).
- In-repo verification docs: `docs/VERIFY.md:24-26` is the claim ledger these gates keep true; `.githooks/pre-commit:9` is the commit-time enforcement.

### 3.3 Licence

Repo root `LICENSE` is MIT (per sibling research `docs/research/2026-11-07-module-deploy-supplychain.md` §3/§6; `LICENSE:3` “Copyright (c) 2026 Meteolink contributors”). A standalone CLI repo inherits this cleanly only with the same licence file kept — anti-toy criterion §8.

## 4. Primary-source requirements

What each primary source **legitimately** requires of a standalone evidence attestor — and what it does not require:

### 4.1 SLSA (R1–R5) — the level vocabulary

- **Build L0:** “No guarantees” (R1, Build L0). **Build L1:** “Provenance exists” — producer follows a consistent process, provenance exists describing build platform/process/top-level inputs, and is distributed to consumers; provenance “may be incomplete and/or unsigned at L1” (R1, Build L1). **Build L2:** “Signed provenance, generated by a hosted build platform” — provenance tied to dedicated infrastructure through a digital signature, verification includes validating authenticity (R1, Build L2). **Build L3:** “Hardened builds” — prevents runs influencing one another and keeps signing secrets inaccessible to user-defined build steps (R1, Build L3).
- **Producer requirements (all of L1–L3):** choose an appropriate build platform, follow a consistent build process so verifiers can form expectations, and distribute provenance (R3, Producer). **Build-platform requirements:** provenance MUST unambiguously identify the output package by cryptographic digest and describe how it was produced; at L2 authenticity MUST be validatable (signature + trust definition); at L3 provenance MUST be strongly resistant to forgery by tenants with secrets inaccessible to user build steps (R3, Provenance generation table). Isolation: L2 hosted (not an individual workstation), L3 isolated/ephemeral environments (R3, Isolation strength).
- **Provenance predicate (R2):** `predicateType` MUST be `"https://slsa.dev/provenance/v1"`; REQUIRED for L1: `buildDefinition` (`buildType`, `externalParameters`) + `runDetails` (`builder.id`); `resolvedDependencies` are `[ResourceDescriptor]` with `uri`+`digest` (R2, Schema). `builder.id` is the sole determiner of SLSA Build level and MUST reflect the trust base (R2, Builder). Extension fields SHOULD be `<vendor>_<field>` and MUST NOT alter meaning of standard fields (R2, Extension fields).
- **Distributing provenance (R4):** attestations SHOULD be bound to artifacts, not releases; provenance SHOULD accompany the artifact at publish time; attestations MUST be published in at least one place and SHOULD be in more than one (source-repo release, package registry sidecar, transparency log); attestations SHOULD be immutable — a new attestation for the same artifact MUST NOT overwrite the old; in-toto recommends filename `<filename>.intoto.jsonl` (R4). Source-based ecosystems (Go via module proxy) SHOULD provide build provenance for source archives since the proxy transforms the repository (R4, Considerations for source-based ecosystems).
- **Verifying artifacts (R5):** Step 1 — check SLSA level: verify envelope signature from roots of trust, verify `subject` matches artifact digest, verify `predicateType`, look up level from (`public key`, `builder.id`); Step 2 — check expectations (`builder.id`, canonical source repo, `buildType`, `externalParameters`, reject unrecognized fields); Step 3 (optional) — check `resolvedDependencies` recursively (R5). Expectations may be trust-on-first-use, defined by producer, or defined in source (R5, Forming Expectations).

**Consequence for `cgo-attest`:** today the frozen system is honestly **L0+ (static anchor, no signature, no hosted builder, no transparency log)** — stronger than L0 because the tracked anchor + fault proofs exist (`docs/VERIFY.md:26`, `tools/verify.py:82-128`), but explicitly not a signature (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`). A standalone tool must (a) declare its achieved level per artifact the way R1/R3 demand, (b) emit SLSA-shaped provenance (R2 schema) so that R5 verification is possible, and (c) never claim L2/L3 without a signed hosted builder + isolated platform (R1/R3). Anything beyond that is citation theatre.

### 4.2 in-toto (R6) — the attestation envelope

The in-toto Attestation Framework “provides a specification for generating verifiable claims about any aspect of how a piece of software is produced” (R6, README). The spec decomposes into Envelope (authentication + serialization), Statement (binds to subject, identifies predicate type), Predicate (arbitrary typed metadata), Bundle (groups attestations) (R6, spec v1 README index). Parsing rules: consumers MUST ignore unrecognized fields; versioning is SemVer, TypeURI reflects major version; extension fields MUST follow the monotonic principle — ignoring a field MUST never turn a DENY into an ALLOW (R6, Parsing rules; R2 Extension fields states the same principle for SLSA provenance).

**Consequence for `cgo-attest`:** the CSS/JSON sidecar pattern of `valuation.json` (`pkg/campagne/valuation.go:36-44`) is a *predicate-shaped* annex, not a full attestation — it has no envelope signature, no `subject` binding to the CSV digest, no `predicateType`. A conformant `cgo-attest` wraps each published evidence bundle in an in-toto Statement (`subject` = sha256 of the artifact) so that R5 Step 1 verification can run. Today nothing does this.

### 4.3 CycloneDX SBOM (R7) — the dependency graph

CycloneDX SBOM “inventory software components and services and the dependency relationships between them”, including versions and hierarchical interconnections, in machine-readable form enabling vulnerability management and supply-chain risk identification (R7, SBOM capability). It “goes beyond simple inventories” via dependency hierarchy, not a flat package list (R7, Introduction).

**Consequence for `cgo-attest`:** the frozen corpus is *evidence data*, not a software dependency graph. An SBOM is the wrong primary for corpus integrity — the honest requirements are the seal (content) + SLSA/in-toto attestation (provenance). Where CycloneDX *is* legitimate: the `cgo-attest` CLI itself is shipped software — its own Go dependencies (`go.mod:5-10`) and release artifacts should publish a CycloneDX SBOM per artifact so consumers can inventory *the tool* (R7). Conflating the two (SBOM of the tool presented as proof of the corpus) would be a toy. The corpus SBOM-equivalent is the seal + per-run manifests + attestation bundle, not a CycloneDX document.

### 4.4 Git ledger (R9) — the commit anchor

Git is cryptographically secure but not foolproof; signing tags (`git tag -s`) and commits (`git commit -S`) with GPG lets consumers verify work comes from a trusted source and detect tampering (R9, Signing Your Work). `git tag -v` / `git log --show-signature` / `git merge --verify-signatures` verify signatures (R9). Everyone on the team must sign (`commit.gpgsign true`) or the ledger degrades to unsigned history (R9, Everyone Must Sign).

**Consequence for `cgo-attest`:** the tracked anchor `docs/frozen-manifests.sha256` is currently unsigned — anyone who can push a rewrite of the anchor defeats the seal (the design acknowledges this: `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`). The minimal git-ledger upgrade is to sign the anchor commit/tag (R9 signing) and have `cgo-attest` verify the signature before trusting the anchor. Without that, the tool proves *internal consistency* only, which is exactly the current VERIFY row (“static anchor only”, `docs/VERIFY.md:26`). The tool must say so.

### 4.5 Go modules (R8) — the distribution mechanics

Module path + version uniquely identify a module and prefix all package paths; v2+ paths must end `/v2`; the `go` directive is a mandatory minimum Go version (R8, `module` + `go`). `require` records a module path + version (`v1.2.3` or pseudo-version) (R8, `require`). Retracted versions remain available but are not selected by `go get` (R8, `retract`). Source-based distribution via the module proxy is explicitly called out in R4 as an ecosystem that SHOULD provide build provenance for source archives because the proxy transforms the repository (R4, Considerations for source-based ecosystems).

**Consequence for `cgo-attest`:** as an independent CLI repo it gets: its own `go.mod` + module path (R8), `go.sum` checksums via R8 `require`, immutable tags (R8 retract semantics), and the R4 obligation to ship SLSA provenance alongside release artifacts. The frozen corpus seal is best distributed the same way R4 prescribes: alongside the source-repo release.

## 5. Gap analysis (current → releasable `cgo-attest`)

Ordered by severity. Each gap cites local evidence; primary-source side cites R# or is `unverified — primary source not found`.

1. **No signature, no hosted builder — the tool is L0 by design, and must declare it.** Anchor is unsigned (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:90-95` out of scope: “No signing key, remote timestamp, transparency log”). Per R1, L1 requires provenance *exists* (satisfied by the seal + manifests); L2 requires signed provenance from a hosted builder (not satisfied); L3 requires hardened isolation (not satisfied). Per R3, producer MUST distribute provenance (satisfied in-repo by tracked anchor, not yet by release artifact). Gap: no signed anchor, no transparency-log entry (R4 SHOULD publish in >1 place). Fix: sign the anchor (R9), publish attestation alongside release (R4), declare L1 until L2 infrastructure exists (R1). Do not claim “verified” beyond what `docs/VERIFY.md:26` already narrows.
2. **Anchor → manifest → file chain has a known SKIP hole.** `gate_frozen` SKIPs when `book/` exists but audit or `data/runs` missing (`tools/verify.py:65-66`), and the VERIFY row notes the resulting exit-2 detail “says ‘book/ absent’ inaccurately” (`docs/VERIFY.md:26`). Per R5 verification, a verifier that answers “not run” instead of DENY on a partial tree is a soundness bug: the step-1 check (R5) cannot complete. Fix: distinguish `absent` (SKIP) from `partial` (FAIL, naming the missing piece); `cgo-attest` must never return exit 0 on a partial tree. This is the cheapest correctness fix and the one most likely to be skipped.
3. **`ProvenanceSHA` is not a corpus provenance.** Hashes only lexically-last `aqm_eval.csv` (`pkg/figures/figures.go:41-53`); mutating any of the other 158 runs does not move it; renaming a run changes which file is “last”. `docs/architecture.md:80-83` honestly says “sha256 du dernier `aqm_eval.csv`”, but the “triple provenance” framing (`pkg/figures/figures.go:39-40`) lets readers believe it covers the corpus. Per R2, `subject` MUST identify output artifacts by digest — a last-file digest cannot be the subject of a corpus attestation. Fix: `cgo-attest` computes a corpus manifest hash (sorted `path + sha256` over all runs, same shape as the seal) and binds both `ProvenanceSHA` (kept, renamed) and the corpus hash into the attestation subject.
4. **No SLSA/in-toto envelope at all.** `valuation.json` (`pkg/campagne/valuation.go:36-44`) is predicate-shaped but has no `_type`, no `subject`, no `predicateType`, no signature. Per R2/R6, without a Statement binding `subject` to a digest and an Envelope signature, R5 verification cannot run. Fix: `cgo-attest emit` produces an in-toto Statement (`_type https://in-toto.io/Statement/v1`, `subject` = corpus digest, `predicateType` = `https://slsa.dev/provenance/v1` per R2) with `buildDefinition` (`buildType`, `externalParameters` = expectations inputs, `resolvedDependencies` = per-run manifest descriptors per R2) and `runDetails.builder.id` restricted to an honest identifier (today: `cgo-attest/0.x`, **not** a hosted-builder URI).
5. **`SeedRuns` + `HardwareRecommendation` leak into the evidence layer.** `SeedRuns` ships an unversioned, unhashed `runs-frozen.tar.gz` with no relationship to the seal (`pkg/results/seed.go:15-16`, `docs/VERIFY.md:26`); `HardwareRecommendation` ships a vendor claim with **unverified — primary source not found** provenance (CAKE/RFC 8290 informative only, MikroTik not fetched — see sibling research §4.1). Neither belongs in a verifier. Fix: `cgo-attest` has no embed, no vendor string, no shaping/measurement code; it reads a corpus path and answers. (`SeedRuns` moves to app `cmd`; hardware advice stays CGO-app-only, labelled unverified.)
6. **Silent skip makes “empty” indistinguishable from “corrupt”.** `Scan` returns `nil,nil` on no files and `continue`s on unreadable/missing-header (`pkg/results/results.go:71-100`); HTTP callers ignore the error (`groups, _ :=`). Per R5 Step 2, a verifier MUST fail on unrecognized/unexpected input rather than silently producing an ALLOW — silent skip violates the spirit of that requirement and of the in-toto monotonic principle (R6: ignoring MUST NOT turn DENY into ALLOW). Fix: `cgo-attest` reports skipped files by name with reason; empty corpus is an explicit error, never a zero-count PASS.
7. **The seed’s anti-traversal is good; the corpus path handling is not.** `SeedRuns` correctly rejects `..`/absolute (`pkg/results/seed.go:43-54`), but `Scan`’s glob (`pkg/results/results.go:66-70`) and `generate`’s path writes (`pkg/figures/figures.go:26-35`) have no such guard, and `runFilter` is interpolated into a `filepath.Join` pattern (`pkg/results/results.go:67-69`). A standalone verifier must treat the corpus path as untrusted input (cf. `docs/VERIFY.md:69` `validRunID` on `POST /api/run/start`). Fix: canonicalise + prefix-check every corpus path before open; reject `..`/absolute in `--corpus`/`--anchor` arguments.
8. **No SBOM of the tool, no version identity.** Per R7, shipped software should have a machine-readable component + hierarchy inventory; per R8, `go.mod` path+version is the module identity; per R4, attestations accompany artifacts. Today there is no `.goreleaser` config, no SBOM, no signed tag in the evidence layer — the provider is `cdnjs.cloudflare.com` in the deploy sibling for the same reason. Fix: `cgo-attest` releases carry CycloneDX SBOM of its own deps, `cgo-attest version` == tag (R8 `require` version form), and every release carries the in-toto bundle (R4 `<filename>.intoto.jsonl` pattern, R4; exact filename convention **unverified — primary source not found** beyond R4’s `<filename>.intoto.jsonl` recommendation, which *is* verified in R4).
9. **`book/tools/audit-archive.py` is opaque in this worktree.** Its internals (file-hash checks, `config_sha256`, the `Tableau 15` coherence line) were not read because `book/` is gitignored and absent. Per R3, provenance completeness is best-effort at L1 but the *digest identification* is MUST — the tool must not re-implement what it cannot see. Fix: `cgo-attest` shells the audit (as `gate_frozen` does today, `tools/verify.py:68`) and records its exit code + OK lines verbatim into the attestation `byproducts`/`externalParameters`, rather than re-deriving the counts in Go.

## 6. Target tool — `cgo-attest` (independent CLI repo)

Single-purpose, zero-spine design. It reads evidence, it never measures, never shapes, never re-seals. Per R4 it ships an attestation alongside every artifact; per R5 it verifies in three steps; per R6 it emits in-toto Statements.

```
cgo-attest verify  --corpus <runs-dir> --anchor <seal-file> [--expect-counts 159/528/208/74/246]
                   [--audit book/tools/audit-archive.py] [--signed-anchor <sig>]
                   → 0 iff audit exit 0 with 2 OK lines AND seal_findings empty AND counts match
                   → 1 on any mismatch (names the run/file); 2 on absent corpus (SKIP, never green)
cgo-attest counts  --corpus <runs-dir>  → JSON {runs, lines, valid, degraded, quarantined, cells} (via Scan semantics, skipped files named)
cgo-attest header  --frozen <runs-dir> --live <runs-dir> → frozen-era header set vs live (gate_provenance)
cgo-attest sidecar --sidecar <valuation.json> → valuation_problem verdict (basis/source enums)
cgo-attest emit    --corpus <runs-dir> --anchor <seal-file> → in-toto Statement bundle (subject=corpus digest,
                   predicateType https://slsa.dev/provenance/v1, buildDefinition+resolvedDependencies per run,
                   builder.id=cgo-attest/<version>), written to <artifact>.intoto.jsonl (R4) + printed digest
cgo-attest version → prints tag version (R8 form vX.Y.Z), exit 0
```

Design invariants (each kills one §5 gap):

1. **Pure functions, filesystem as input.** `seal_findings`, `valuation_problem`, `header_problems` are already pure (`tools/verify.py:48-62,138-146,133-137`) — port them nearly verbatim into Go internals. Clock, subprocess, and network stay in `cmd/` adapters only (gap §5.2 SKIP distinction enforced in the CLI exit mapping).
2. **Corpus digest = sorted manifest hash**, same shape as `docs/frozen-manifests.sha256` (R2 `subject`). `ProvenanceSHA` is kept for compat but renamed in docs to `last-file SHA-256`, never presented as corpus proof (gap §5.3).
3. **Signed anchor is optional-verified, never assumed.** `--signed-anchor` accepts an R9-style GPG detached sig; when absent the tool prints “static anchor only (docs/VERIFY.md:26)” and refuses `--require-signed` success without it (gap §5.1). SLSA level is declared in `emit` output (`builder.id` never claims a hosted builder until one exists, R1/R2).
4. **Path canonicalisation** on every argument before any open (gap §5.7); skipped files are reported, never silent (gap §5.6); partial tree is FAIL, absent tree is exit 2 (gap §5.2).
5. **No `embed`, no vendor strings, no RFC 8290 claims, no CycloneDX of the corpus.** The tool’s own CycloneDX SBOM is a release asset (R7), not a corpus artifact (gap §5.5, §5.8).
6. **Audit is shelled, not re-implemented** (gap §5.9); its stdout/exit are quoted into the bundle so the verdict is replayable (R4 immutable attestation, R5 reproducible verification).

Module shape: independent Go module (`go.mod` with its own path, `go 1.21+` minimum taking R8 into account; today's monorepo uses `go 1.25` `go.mod:3` — the standalone may keep `1.25` or lower, stated in its `go.mod`). No dependencies beyond stdlib (`archive/…`, `crypto/sha256`, `encoding/json`, `encoding/csv`) — the verify layer is already stdlib-only (`tools/verify.py:2`), so the Go port has no reason to add one. Tag `vX.Y.Z` (R8); never move a published tag (R8 retract semantics).

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it. Slices 1–3 are the honesty floor; 4–6 are the attestation/distribution layer.

1. **S1 — Freeze + declare.** Snapshot §2 exported surface into `cgo-attest` README as the `0.1.0` declared public API (per R4-style producer requirement: MUST form expectations, R3 Producer; per R8 module path declared before publish). Port `seal_findings`/`gate_frozen`/`prove_frozen`/`gate_provenance`/`prove_provenance` unchanged (they carry the fault proofs, `tools/verify.py:82-128,170-191`). Acceptance: `go test ./...` green on the extracted tree; README lists every §2 identifier with its `file:line`; `--selftest` equivalents pass with both controls + fault cases.
2. **S2 — Fix SKIP honesty + path hygiene.** Distinguish absent (exit 2) vs partial/corrupt (exit 1, named) (§5.2); canonicalise corpus/anchor paths, reject traversal (§5.7); report skipped files (§5.6). Acceptance: temp-copy fault proof — partial tree → FAIL (not SKIP), one corrupt + 158 good runs → skipped name reported, counts still `159/528/208/74/246` on the intact copy.
3. **S3 — Corpus digest.** Implement sorted-manifest hash; keep `ProvenanceSHA` as renamed last-file helper; emit both into the bundle subject (`subject` = corpus digest per R2). Acceptance: mutating any non-last file changes the corpus digest; two `emit` runs over unchanged data are byte-identical (no `time.Now`, no local path in the bundle except caller-supplied).
4. **S4 — Signed anchor + level declaration.** Accept a GPG detached signature on the anchor (R9), verify via `git tag -v` / `gpg --verify` semantics (R9 Verifying Tags); add `--require-signed`; declare SLSA level in every `emit` (`L1` until a hosted builder exists, R1/R2). Acceptance: unsigned anchor + `--require-signed` → exit ≠ 0; signed anchor with bad key → named failure; `emit` output contains `builder.id` + level that matches the actual signing infrastructure.
5. **S5 — in-toto bundle + SBOM.** `emit` writes `<artifact>.intoto.jsonl` (R4) with Statement + SLSA predicate (R2 schema: `buildDefinition`, `resolvedDependencies`, `runDetails.builder`); each release ships CycloneDX SBOM of `cgo-attest` deps (R7) + `go.mod`/`go.sum` (R8). Acceptance: bundle parses against an in-toto validator (R6 spec); SBOM parses as CycloneDX and lists direct deps from `go.mod`; `subject` digest matches the corpus.
6. **S6 — Recertify + release.** Re-run `cgo-attest verify` against the live frozen corpus, compare to `docs/VERIFY.md:24-26` (`159/528/208/74/246`, seal 159 OK, six gates 6 passed); publish the attestation alongside the repo release (R4 multi-place SHOULD); write the VERIFY row with command, host, date. Acceptance: independent verifier re-runs the command and owns the `VERIFIED` row (per repo Roles: builder may not certify its own number); `cgo-attest version` == tag == release assets.

Each slice is independently reviewable. Slices 4–5 are the only ones that change the trust claim (unsigned → signed, JSON → attestation) and each must be verifier-re-run before its row is claimed.

## 8. Anti-toy criteria (what would make this a toy — and the bar for “not a toy”)

A `cgo-attest` submission is a toy if it re-describes its own outputs instead of attesting against an external check. It is **not** a toy only if all of the following hold:

1. **Independent reproduction of the frozen corpus.** `cgo-attest verify` over the sealed corpus reproduces `159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells` (`docs/VERIFY.md:24`) and the seal 159-OK (`docs/VERIFY.md:26`) — not just its own fixture tests. If a counting-rule change moves any of those numbers without a named `VERIFIED` row, it rewrites history.
2. **Fault-proof parity.** Every witness in `prove_frozen`/`prove_provenance` (`tools/verify.py:82-128,170-191`) is ported and still catches: rewritten manifest, mutated data file, extra column, `basis=guess`, `estimated+ebpf` — with the controls (`measured+ebpf`, `estimated+empty`) passing. A verifier without a fault proof is an opinion.
3. **Provenance honesty, SLSA-typed.** This is **Build L1 at best** until a hosted builder signs (R1/R3): the tool declares its level, binds `subject` to the corpus digest (R2), emits the predicate schema (R2), and never presents the last-file hash (`pkg/figures/figures.go:41-53`) or the Fenwick vendor string (`pkg/results/hardware.go:7-12`) as corpus proof. CAKE/MikroTik/statistics claims stay **unverified — primary source not found**.
4. **No silent skip, no silent zero, no green SKIP.** Partial tree, corrupt file, skipped run, and absent `book/` are four distinct answers (`FAIl` / named skip / exit-2), never a zero count read as a measurement (R5 Step 2 reject-unrecognized, R6 monotonic principle). If two corpora can yield the same verdict with no way to tell which files were skipped, it is a helper, not an attestor.
5. **Verify-from-a-clean-clone.** A fresh clone (no `book/`) answers exit 2 naming the absence — never exit 0 — and a restored copy (`docs/VERIFY.md:37` second-copy drill, exit 0 with identical bytes/counts) verifies green. The tool must survive the R4 distribution story: artifact + attestation `subject` digest matches a corpus the verifier assembled independently.
6. **Version discipline.** Declared API, immutable `vX.Y.Z` tags (R8), changelog in version terms, and a release that carries bundle + SBOM + `go.sum` together (R4/R7/R8). A repo that silently rewrites the anchor between tags, or ships no declared API, is a toy regardless of code quality.

---
*Research only. No source was modified in this pass; implementation of `cgo-attest` waits for human approval per AGENTS.md (“Stop and ask” — new dependency, scope growth, frozen-data adjacent). `book/` was never touched.*

*Skipped / not checked: `book/tools/audit-archive.py` internals (gitignored, absent — contract only, §5.9), live bench re-measurement (no Linux 6.8 host — `docs/VERIFY.md:28` BLOCKED), cosign/Sigstore signing-key choice (**unverified — primary source not found** for this org), GoReleaser SBOM command details beyond R8 (fetched in sibling deploy research, not re-fetched here). Risk the user must know: the current `ProvenanceSHA` covers only the last file, so any figure/report/integrity claim citing it as corpus proof is weaker than it reads (`pkg/figures/figures.go:41-53` vs `docs/VERIFY.md:26`); and the anchor is unsigned, so a writer who can rewrite the anchor too is out of scope by design (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`).*
