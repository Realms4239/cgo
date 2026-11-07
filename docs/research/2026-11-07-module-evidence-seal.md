# Frozen corpus seal + manifest + hash8 as standalone tool — primary-source research

*Date: 2026-11-07 · Scope: frozen corpus seal + per-run manifest + hash8 (`docs/frozen-manifests.sha256`, `tools/verify.py` `gate_frozen`/`seal_findings`, `pkg/results/*.go`, `pkg/figures/*.go`, `book/tools/audit-archive.py`) read as if released tomorrow as independent CLI `cgo-seal`. Research only — no code changed.*

Primary sources fetched this pass: (R1) SLSA v1.0 Security levels (<https://slsa.dev/spec/v1.0/levels>); (R2) SLSA v1.0 Provenance predicate (<https://slsa.dev/spec/v1.0/provenance>); (R3) SLSA v1.0 Producing artifacts / Requirements (<https://slsa.dev/spec/v1.0/requirements>); (R4) SLSA v1.0 Distributing provenance (<https://slsa.dev/spec/v1.0/distributing-provenance>); (R5) SLSA v1.0 Verifying artifacts (<https://slsa.dev/spec/v1.0/verifying-artifacts>); (R6) in-toto Attestation Framework spec v1.2 (<https://github.com/in-toto/attestation/blob/main/spec/v1/README.md>) and overview (<https://in-toto.io/>); (R7) CycloneDX SBOM capability (<https://cyclonedx.org/capabilities/sbom/>); (R8) GNU Coreutils 9.12 manual, §6.7 sha2 utilities (<https://www.gnu.org/software/coreutils/manual/coreutils.html>). Local seal design: `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md`. Sibling research (containers, not duplicated): `docs/research/2026-11-07-module-evidence-attestor.md` (seal + header/valuation → `cgo-attest`), `docs/research/2026-11-07-module-results-figures.md` (aggregation/rendering → module API), `docs/research/2026-11-07-module-verification-gates.md` (six-gate layer).

Method: read `docs/frozen-manifests.sha256:1-159` (159 lines), `tools/verify.py:46-128` (`seal_findings`, `gate_frozen`, `prove_frozen`), `pkg/results/results.go:1-236`, `pkg/results/csvrows.go:1-65`, `pkg/results/seed.go:1-79`, `pkg/results/hardware.go:1-12`, `pkg/figures/figures.go:1-156`, `book/tools/audit-archive.py:1-352` (this worktree), `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:1-95`, `docs/architecture.md:76-83`, `docs/VERIFY.md:24-26`, `go.mod:1-3`. Every technical claim below cites `file:line` or `(R#)`; anything else is `unverified — primary source not found`.

## 1. Responsibility

Today the evidence seal does three distinct jobs that `cgo-seal` must keep separate — conflating them is the main design risk:

1. **Content integrity — manifest → file.** `book/tools/audit-archive.py` verifies every run's `manifest.json` against the files it names (per-file SHA-256, config JSON, `config_sha256`) (`book/tools/audit-archive.py:167-216`); its header contract is 159 runs / 528 rows with two `OK` lines for corpus counts and for cells/manifests/inventory/text (`book/tools/audit-archive.py:1-17`, `book/tools/audit-archive.py:336-352`). `tools/verify.py` pins the verified shape as 159 runs / 528 rows / 208 valid / 74 degraded / 246 invalid / 22 cells (`docs/VERIFY.md:24`). `docs/architecture.md:76-83` frames this layer: the writer freezes `data/runs/<run_id>/` and writes a SHA-256 manifest.
2. **Corpus seal — anchor → manifest.** `docs/frozen-manifests.sha256` holds one SHA-256 per manifest (159 lines, sorted, `sha256sum`-compatible, `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`). `seal_findings` (`tools/verify.py:48-62`) fails closed on absent anchor, malformed line, changed manifest, missing run, or extra run. The design is explicit: “it detects any rewrite of a manifest, including one made together with the file that manifest covers. It does not defend against an attacker who can rewrite the anchor too” (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`; repeated `docs/VERIFY.md:26`). The corpus (`book/`) is gitignored, so the tracked anchor outside `book/` is the only versioned root — chain anchor → manifest → file (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:20-23`).
3. **Provenance label — figure/report hash8.** `ProvenanceSHA` / `ProvenanceHash8` (`pkg/figures/figures.go:41-61`) hash only the lexically-last `aqm_eval.csv` and are injected into SVGs, the exported report, and `/api/integrity` (`docs/architecture.md:80-83`, `pkg/api/server.go:896,975,1022,1079`). This is the weakest link of the three (see §5.3).

What the system deliberately does **not** do: no signing key, no remote timestamp, no transparency log, no git LFS, no new dependency (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:90-95` “Out of scope”). It never re-seals silently: “The gate never writes the anchor” (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:38-40`). `AGENTS.md` also forbids touching `book/data/runs` and the CSV schemas. The standalone tool inherits those constraints: it verifies evidence, it never re-measures and never re-seals.

## 2. Exported surface (file:line)

Complete inventory of the seal/manifest/hash8-relevant exported identifiers. Unexported helpers (`seal_findings` `tools/verify.py:48-62` is module-level but not a library API; `provenanceMeta` `pkg/figures/figures.go:63-79`; `runIDs` `pkg/results/results.go:53-63`) are not API.

### 2.1 `tools/verify.py` (the gate layer that becomes `cgo-seal`)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `seal_findings(root)` | func | `tools/verify.py:48-62` | Anchor parser + three-way diff (changed / missing / extra). Returns `[]string`, never writes. Malformed line → immediate finding (`tools/verify.py:55-56`). Path must start `book/data/runs/` and end `/manifest.json` (`tools/verify.py:55`). |
| `gate_frozen()` | func | `tools/verify.py:63-81` | Runs `book/tools/audit-archive.py` via `_proc([sys.executable, audit], 120)` (`tools/verify.py:68`), requires exit 0 + ≥2 `OK` lines (`tools/verify.py:74-76`), then `seal_findings`. SKIP (exit 2 path) when `book/` or audit absent (`tools/verify.py:65-66`); FAIL on timeout (`tools/verify.py:69-70`). |
| `prove_frozen()` | func | `tools/verify.py:82-128` | Fault proof: copied corpus passes control, whitespace-rewritten manifest caught by seal (`tools/verify.py:99-102`), mutated data file caught by audit (`tools/verify.py:108-124`), restore verified (`tools/verify.py:105-107`). |
| `SEAL` | const | `tools/verify.py:47` | `"docs/frozen-manifests.sha256"` — the only anchor path the gate knows. |

### 2.2 `docs/frozen-manifests.sha256` (the anchor itself)

Tracked file, 159 lines, format `<sha256>  book/data/runs/<run>/manifest.json` (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`). Verified properties (`docs/VERIFY.md:26`): 159 lines, sorted, unique, every path under `book/data/runs/` and present, every hash matches; `git ls-files --eol` → `i/lf w/lf`; `.gitattributes:5` (`docs/frozen-manifests.sha256 text eol=lf`). Not a signature — a static anchor (`docs/VERIFY.md:26` “Proves a static anchor only: not a signature”).

### 2.3 `pkg/results` (aggregation provenance)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ReadAQM(path)` | func | `pkg/results/csvrows.go:19-45` | Header-name CSV reader, `FieldsPerRecord=-1`, short rows kept (`pkg/results/csvrows.go:26`). |
| `ColIndex(header, name)` | func | `pkg/results/csvrows.go:49-56` | Trimmed exact-match lookup, `-1` if absent. |
| `Field(header, row, name)` | func | `pkg/results/csvrows.go:59-65` | Named field of one row, `""` if absent. |
| `Group` | struct | `pkg/results/results.go:13-47` | Cell `profile\|qdisc\|cc\|direction`; `Count/Quarantined/Degraded`; medians+IQRs; valid-only sub-struct; `Runs` sorted provenance (`pkg/results/results.go:46`). |
| `Scan(dataDir, runFilter)` | func | `pkg/results/results.go:65-236` | Glob `*/aqm_eval.csv`; skips `run-smoke*` (`pkg/results/results.go:88-90`); silent-skip on unreadable/missing-header (`pkg/results/results.go:93-100`); empty → `nil,nil` (`pkg/results/results.go:71-73`). |
| `SeedRuns(root)` | func | `pkg/results/seed.go:24-79` | Extracts embedded `runs-frozen.tar.gz` (`pkg/results/seed.go:15-16`) only when no CSV exists; anti-traversal (`pkg/results/seed.go:43-54`). **Not** part of `cgo-seal` core (see §5.5). |
| `HardwareRecommendation(...)` | func | `pkg/results/hardware.go:7-12` | Hardcoded FR vendor string. **Excluded** from `cgo-seal` core (see §5.5). |

### 2.4 `pkg/figures` (hash8 provenance)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ProvenanceSHA(dataDir)` | func | `pkg/figures/figures.go:41-53` | SHA-256 of lexically-last `aqm_eval.csv` only (`pkg/figures/figures.go:42-47`). **Deficient** (§5.3) — kept as `last-file SHA-256`, never corpus proof. |
| `ProvenanceHash8(dataDir)` | func | `pkg/figures/figures.go:56-61` | First 8 hex of `ProvenanceSHA`. Short label, not a proof. Test pin: `pkg/figures/figures_test.go:55-73`. |
| `Generate(dataDir, outDir)` | func | `pkg/figures/figures.go:17-37` | Calls `results.Scan`, writes `small_p95.svg` + `scatter.svg` with non-reproducible `provenanceMeta` (`pkg/figures/figures.go:63-79`). Not part of `cgo-seal` core. |
| `provenanceMeta` (unexported) | — | `pkg/figures/figures.go:63-79` | Embeds `time.Now().UTC()` (`pkg/figures/figures.go:77`) + local path (`pkg/figures/figures.go:76`) + hardcoded creator (`pkg/figures/figures.go:78`) — reproducibility bug (§5.3). |

### 2.5 `book/tools/audit-archive.py` (opaque contract, read this pass)

`EXPECTED = {"runs": 159, "rows": 528, "valid": 208, "degraded": 74, "invalid": 246}` (`book/tools/audit-archive.py:34`); `sha_file` chunked SHA-256 (`book/tools/audit-archive.py:69-74`); `load_corpus` rejects `run-smoke` dirs (`book/tools/audit-archive.py:77-80`); `check_counts` (`book/tools/audit-archive.py:94-113`); `check_manifests` requires exact file-set equality + per-file SHA-256 match (`book/tools/audit-archive.py:167-216`); `main` prints two `OK` lines on success, `ECHEC` + exit 1 listing all errors on failure (`book/tools/audit-archive.py:336-352`).

## 3. Dependencies

### 3.1 Outgoing (what `cgo-seal` needs)

- `tools/verify.py:1-6` — Python 3 stdlib only (`argparse`, `hashlib`, `json`, `os`, `re`, `shutil`, `subprocess`, `sys`, `tempfile`, `time`, `dataclasses`, `pathlib`). Docstring: “Standard library only; no network; no shell” (`tools/verify.py:2`). Zero third-party imports — the seal/header layer never imports `pkg/results` or `pkg/metrics`.
- `pkg/results/csvrows.go:3-7` stdlib `encoding/csv`, `os`, `strings`. `pkg/results/results.go:3-10` stdlib + internal `pkg/metrics`. `pkg/figures/figures.go:3-13` stdlib + internal `pkg/results`. The internal chain `metrics → results → figures` is the extraction blocker for a Go port of `Scan`/`Generate` (sibling results-figures research §5.1); the seal layer alone has no such chain.
- `go.mod:1-3`: module `github.com/Realms4239/cgo`, `go 1.25`. Deps `fyne`/`bubbletea`/`lipgloss`/`golang.org/x/sys` (`go.mod:5-10`) are not needed by the seal layer.
- **No `syscall`, no network, no clock dependence** in the seal layer: `seal_findings` is pure file hashing (`tools/verify.py:48-62`). The only clock (`time.Now` in `pkg/figures/figures.go:77`) and the only subprocess (`audit-archive.py` in `tools/verify.py:68`) are both removable/testable seam points (§6).

### 3.2 Incoming (who needs the seal layer)

- `tools/verify.py` itself: `gate_frozen` is called by `--only frozen` and by the full six-gate run (`tools/verify.py:455-462` registry, `tools/verify.py:517-523` verdict 0/1/2). The pre-commit hook runs `python tools/verify.py --staged`, and any change to `tools/verify.py` forces all six gates (`tools/verify.py:464-471` `select_staged`).
- `pkg/api/server.go:896` (compare-cache key via `ProvenanceHash8`), `:975` (`/api/integrity`: `sha256`+`hash8` via `figures.ProvenanceSHA/Hash8`), `:1022` (report header), `:1079` (evidence bundle).
- `pkg/results` callers: `pkg/api/server.go:770,942,982,1004`, `pkg/api/duel.go:77`, `pkg/api/translate.go:25` (all `results.Scan`).
- Claim ledger: `docs/VERIFY.md:24-26` is the claim these gates keep true.

### 3.3 Licence

Repo root `LICENSE` is MIT (sibling deploy research §3/§6; `LICENSE:3` “Copyright (c) 2026 Meteolink contributors”). A standalone CLI repo inherits this cleanly only with the same licence file kept — anti-toy criterion §8.

## 4. Primary-source requirements

What each primary source **legitimately** requires of a standalone seal tool — and what it does not require:

### 4.1 SLSA (R1–R5) — the level vocabulary

- **Build L0:** “No guarantees” (R1, Build L0). **Build L1:** “Provenance exists” — producer follows a consistent process, provenance exists describing build platform/process/top-level inputs, and is distributed to consumers; provenance “may be incomplete and/or unsigned at L1” (R1, Build L1). **Build L2:** “Signed provenance, generated by a hosted build platform” — provenance tied to dedicated infrastructure through a digital signature, verification includes validating authenticity (R1, Build L2). **Build L3:** “Hardened builds” — prevents runs influencing one another and keeps signing secrets inaccessible to user-defined build steps (R1, Build L3).
- **Producer requirements (all of L1–L3):** choose an appropriate build platform, follow a consistent build process so verifiers can form expectations, and distribute provenance (R3, Producer). **Build-platform requirements:** provenance MUST unambiguously identify the output package by cryptographic digest and describe how it was produced; at L2 authenticity MUST be validatable (signature + trust definition); at L3 provenance MUST be strongly resistant to forgery by tenants with secrets inaccessible to user build steps (R3, Provenance generation table). Isolation: L2 hosted (not an individual workstation), L3 isolated/ephemeral environments (R3, Isolation strength).
- **Provenance predicate (R2):** `predicateType` MUST be `"https://slsa.dev/provenance/v1"`; REQUIRED for L1: `buildDefinition` (`buildType`, `externalParameters`) + `runDetails` (`builder.id`); `resolvedDependencies` are `[ResourceDescriptor]` with `uri`+`digest` (R2, Schema). `builder.id` is the sole determiner of SLSA Build level and MUST reflect the trust base (R2, Builder). Extension fields SHOULD be `<vendor>_<field>` and MUST NOT alter meaning of standard fields (R2, Extension fields).
- **Distributing provenance (R4):** attestations SHOULD be bound to artifacts, not releases; provenance SHOULD accompany the artifact at publish time; attestations MUST be published in at least one place and SHOULD be in more than one (source-repo release, package registry sidecar, transparency log); attestations SHOULD be immutable — a new attestation for the same artifact MUST NOT overwrite the old; in-toto recommends filename `<filename>.intoto.jsonl` (R4). Source-based ecosystems (Go via module proxy) SHOULD provide build provenance for source archives since the proxy transforms the repository (R4, Considerations for source-based ecosystems).
- **Verifying artifacts (R5):** Step 1 — check SLSA level: verify envelope signature from roots of trust, verify `subject` matches artifact digest, verify `predicateType`, look up level from (`public key`, `builder.id`); Step 2 — check expectations (`builder.id`, canonical source repo, `buildType`, `externalParameters`, reject unrecognized fields); Step 3 (optional) — check `resolvedDependencies` recursively (R5). Expectations may be trust-on-first-use, defined by producer, or defined in source (R5, Forming Expectations).

**Consequence for `cgo-seal`:** the current seal is honestly **Build L1 at best** — the tracked anchor + fault proofs exist (`docs/VERIFY.md:26`, `tools/verify.py:82-128`), but there is no signature, no hosted builder, no transparency log (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:90-95`). A standalone tool must (a) declare its achieved level per artifact the way R1/R3 demand, (b) emit SLSA-shaped provenance (R2 schema) so that R5 verification is possible, and (c) never claim L2/L3 without a signed hosted builder + isolated platform (R1/R3). Anything beyond that is citation theatre.

### 4.2 in-toto (R6) — the attestation envelope

The in-toto Attestation Framework provides a specification for generating verifiable claims about how a piece of software is produced (R6, overview). The spec decomposes into Envelope (authentication + serialization), Statement (binds to subject, identifies predicate type), Predicate (arbitrary typed metadata), Bundle (groups attestations) (R6, spec v1 README index). Parsing rules: consumers MUST ignore unrecognized fields; versioning is SemVer, TypeURI reflects major version; extension fields MUST follow the monotonic principle — ignoring a field MUST never turn a DENY into an ALLOW (R6, Parsing rules; R2 Extension fields states the same principle for SLSA provenance).

**Consequence for `cgo-seal`:** `valuation.json` and `manifest.json` are predicate-shaped, not full attestations — they have no envelope signature, no `subject` binding to the CSV digest, no `predicateType`. A conformant `cgo-seal` wraps each sealed bundle in an in-toto Statement (`subject` = sha256 of the corpus digest) so that R5 Step 1 verification can run. Today nothing does this.

### 4.3 CycloneDX SBOM (R7) — the dependency graph

CycloneDX SBOM inventories software components and services and the dependency relationships between them, including versions and hierarchical interconnections, in machine-readable form enabling vulnerability management and supply-chain risk identification (R7, SBOM capability). It “goes beyond simple inventories” via dependency hierarchy, not a flat package list (R7, Introduction).

**Consequence for `cgo-seal`:** the frozen corpus is *evidence data*, not a software dependency graph. An SBOM is the wrong primary for corpus integrity — the honest requirements are the seal (content) + SLSA/in-toto attestation (provenance). Where CycloneDX *is* legitimate: the `cgo-seal` CLI itself is shipped software — its own Go dependencies (`go.mod:5-10`) and release artifacts should publish a CycloneDX SBOM per artifact so consumers can inventory *the tool* (R7). Conflating the two (SBOM of the tool presented as proof of the corpus) would be a toy. The corpus SBOM-equivalent is the seal + per-run manifests + attestation bundle, not a CycloneDX document.

### 4.4 sha256sum / Coreutils (R8) — the file format

The GNU Coreutils manual documents §6.7 “sha2 utilities: Print or check SHA-2 digests” (R8, TOC). The local anchor format `<sha256>  book/data/runs/<run>/manifest.json` with two spaces, LF endings, sorted order (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`) is verified in-repo by `sha256sum -c docs/frozen-manifests.sha256` → 159 `OK` (`docs/VERIFY.md:26`) and by `seal_findings` parsing `digest  path` with `line.partition("  ")` (`tools/verify.py:53-57`). **Detailed `sha256sum` output-format rules (binary-mode `*` marker, backslash escaping of filenames, `--check` exit codes) are `unverified — primary source not found` in this pass** — only §6.7 existence was fetched, not the sha2-utilities section body.

**Consequence for `cgo-seal`:** it must preserve the exact two-space, LF, sorted format so `sha256sum -c` and `seal_findings` agree, and it must never silently re-emit with spaces/tabs/CRLF that one checker accepts and the other rejects. Anything about `sha256sum` semantics beyond the 159-OK run + local parser must be labelled unverified until the §6.7 body is fetched.

## 5. Gap analysis (current → releasable `cgo-seal`)

Ordered by severity. Each gap cites local evidence; primary-source side cites R# or is `unverified — primary source not found`.

1. **No signature, no hosted builder — the tool is L1 at best, and must declare it.** Anchor is unsigned (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:90-95` out of scope: “No signing key, remote timestamp, transparency log”). Per R1, L1 requires provenance *exists* (satisfied by the seal + manifests); L2 requires signed provenance from a hosted builder (not satisfied); L3 requires hardened isolation (not satisfied). Per R3, producer MUST distribute provenance (satisfied in-repo by tracked anchor, not yet by release artifact). Gap: no signed anchor, no transparency-log entry (R4 SHOULD publish in >1 place). Fix: sign the anchor and publish attestation alongside release (R4), declare L1 until L2 infrastructure exists (R1). Do not claim “verified” beyond what `docs/VERIFY.md:26` already narrows.
2. **Anchor → manifest → file chain has a known SKIP hole.** `gate_frozen` SKIPs when `book/` exists but audit or `data/runs` missing (`tools/verify.py:65-66`), and the VERIFY row notes the resulting exit-2 detail “says ‘book/ absent’ inaccurately” (`docs/VERIFY.md:26`). Per R5 verification, a verifier that answers “not run” instead of DENY on a partial tree cannot complete Step 1. Fix: distinguish `absent` (SKIP) from `partial` (FAIL, naming the missing piece); `cgo-seal` must never return exit 0 on a partial tree. This is the cheapest correctness fix and the one most likely to be skipped.
3. **`ProvenanceSHA` / `hash8` is not a corpus provenance.** Hashes only lexically-last `aqm_eval.csv` (`pkg/figures/figures.go:41-53`); mutating any of the other 158 runs does not move it; renaming a run changes which file is “last”. `docs/architecture.md:80-83` honestly says “sha256 du dernier `aqm_eval.csv`”, but the “triple provenance” framing (`pkg/figures/figures.go:39-40`) lets readers believe it covers the corpus. Per R2, `subject` MUST identify output artifacts by digest — a last-file digest cannot be the subject of a corpus attestation. Per R8, `hash8` is 8 hex chars (32 bits) — a short label, not a proof. Fix: `cgo-seal` computes a corpus manifest hash (sorted `path + sha256` over all runs, same shape as the seal) and binds both `ProvenanceSHA` (kept, renamed `last-file sha256`) and the corpus hash into the attestation subject.
4. **No SLSA/in-toto envelope at all.** `manifest.json` carries per-file `File`/`Sha256` pairs (`book/tools/audit-archive.py:203-216`) but has no `_type`, no `subject`, no `predicateType`, no signature. Per R2/R6, without a Statement binding `subject` to a digest and an Envelope signature, R5 verification cannot run. Fix: `cgo-seal emit` produces an in-toto Statement (`_type https://in-toto.io/Statement/v1`, `subject` = corpus digest, `predicateType` = `https://slsa.dev/provenance/v1` per R2) with `buildDefinition` (`buildType`, `externalParameters` = expectations inputs, `resolvedDependencies` = per-run manifest descriptors per R2) and `runDetails.builder.id` restricted to an honest identifier (today: `cgo-seal/<version>`, **not** a hosted-builder URI).
5. **`SeedRuns` + `HardwareRecommendation` leak into the evidence layer.** `SeedRuns` ships an unversioned, unhashed `runs-frozen.tar.gz` with no relationship to the seal (`pkg/results/seed.go:15-16`, `docs/VERIFY.md:26`); `HardwareRecommendation` ships a vendor claim with **unverified — primary source not found** provenance (CAKE/RFC 8290 informative only, MikroTik not fetched). Neither belongs in a verifier. Fix: `cgo-seal` has no embed, no vendor string, no shaping/measurement code; it reads a corpus path and answers. (`SeedRuns` moves to app `cmd`; hardware advice stays CGO-app-only, labelled unverified.)
6. **Silent skip makes “empty” indistinguishable from “corrupt”.** `Scan` returns `nil,nil` on no files and `continue`s on unreadable/missing-header (`pkg/results/results.go:71-100`); HTTP callers ignore the error (`groups, _ :=`). Per R5 Step 2, a verifier MUST fail on unrecognized/unexpected input rather than silently producing an ALLOW — silent skip violates the spirit of that requirement and of the in-toto monotonic principle (R6: ignoring MUST NOT turn DENY into ALLOW). Fix: `cgo-seal` reports skipped files by name with reason; empty corpus is an explicit error, never a zero-count PASS.
7. **The seed’s anti-traversal is good; the corpus path handling is not.** `SeedRuns` correctly rejects `..`/absolute (`pkg/results/seed.go:43-54`), but `Scan`’s glob (`pkg/results/results.go:66-70`) and `Generate`’s path writes (`pkg/figures/figures.go:26-35`) have no such guard, and `runFilter` is interpolated into a `filepath.Join` pattern (`pkg/results/results.go:67-69`). A standalone verifier must treat the corpus path as untrusted input (cf. `docs/VERIFY.md:69` `validRunID` on `POST /api/run/start`). Fix: canonicalise + prefix-check every corpus path before open; reject `..`/absolute in `--corpus`/`--anchor` arguments.
8. **No SBOM of the tool, no version identity.** Per R7, shipped software should have a machine-readable component + hierarchy inventory; per R4, attestations accompany artifacts. Today there is no `.goreleaser` config, no SBOM, no signed tag in the evidence layer. Fix: `cgo-seal` releases carry CycloneDX SBOM of its own deps, `cgo-seal version` == tag, and every release carries the in-toto bundle (R4 `<filename>.intoto.jsonl` pattern).
9. **`sha256sum` format parity is unproven beyond one run.** The seal’s two-space/LF/sorted contract (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:47-56`) is checked by exactly one `sha256sum -c` run (`docs/VERIFY.md:26`) and the local `partition("  ")` parser (`tools/verify.py:54`); the GNU §6.7 body was not fetched (R8, §4.4). Any filename needing escaping, or any `-c` edge (binary-mode marker, missing file, malformed line), is **unverified — primary source not found**. Fix: fetch R8 §6.7 body before claiming `sha256sum` compatibility beyond the happy path, and add a cross-check test (one file renamed to a name with a space → both checkers agree or both DENY, never split).

## 6. Target tool — `cgo-seal` (independent CLI repo)

Single-purpose, zero-spine design. It reads evidence, it never measures, never shapes, never re-seals. Per R4 it ships an attestation alongside every artifact; per R5 it verifies in three steps; per R6 it emits in-toto Statements.

```
cgo-seal verify  --corpus <runs-dir> --anchor docs/frozen-manifests.sha256 [--expect-counts 159/528/208/74/246]
                 [--audit book/tools/audit-archive.py] [--signed-anchor <sig>]
                 → 0 iff audit exit 0 with 2 OK lines AND seal_findings empty AND counts match
                 → 1 on any mismatch (names the run/file); 2 on absent corpus (SKIP, never green)
cgo-seal digest  --corpus <runs-dir> → corpus manifest hash (sorted "path + sha256", seal shape) + last-file sha256 + hash8 (label)
cgo-seal emit    --corpus <runs-dir> --anchor <seal-file> → in-toto Statement bundle (subject=corpus digest,
                 predicateType https://slsa.dev/provenance/v1, buildDefinition+resolvedDependencies per run,
                 builder.id=cgo-seal/<version>), written to <artifact>.intoto.jsonl (R4) + printed digest
cgo-seal version → prints tag version, exit 0
```

Design invariants (each kills one §5 gap):

1. **Pure functions, filesystem as input.** `seal_findings` is already pure (`tools/verify.py:48-62`) — port it nearly verbatim into Go internals. Clock, subprocess, and network stay in `cmd/` adapters only (gap §5.2 SKIP distinction enforced in the CLI exit mapping).
2. **Corpus digest = sorted manifest hash**, same shape as `docs/frozen-manifests.sha256` (R2 `subject`). `ProvenanceSHA` is kept for compat but renamed in docs to `last-file SHA-256`, never presented as corpus proof (gap §5.3); `hash8` is documented as a 32-bit label (R8 8-hex).
3. **Signed anchor is optional-verified, never assumed.** When absent the tool prints “static anchor only (docs/VERIFY.md:26)” and refuses `--require-signed` success without it (gap §5.1). SLSA level is declared in `emit` output (`builder.id` never claims a hosted builder until one exists, R1/R2).
4. **Path canonicalisation** on every argument before any open (gap §5.7); skipped files are reported, never silent (gap §5.6); partial tree is FAIL, absent tree is exit 2 (gap §5.2).
5. **No `embed`, no vendor strings, no CycloneDX of the corpus.** The tool’s own CycloneDX SBOM is a release asset (R7), not a corpus artifact (gap §5.5, §5.8).
6. **Audit is shelled, not re-implemented**; its stdout/exit are quoted into the bundle so the verdict is replayable (R4 immutable attestation, R5 reproducible verification).
7. **`sha256sum -c` parity is tested, not assumed** (gap §5.9): `digest` must pass both the Go parser and `sha256sum -c` on the same tree, including a space-in-filename fixture — once R8 §6.7 body is fetched.

Module shape: independent Go module (`go.mod` with its own path, `go 1.25` to match `go.mod:3` — or lower, stated in its `go.mod`). No dependencies beyond stdlib (`crypto/sha256`, `encoding/hex`, `encoding/json`, `encoding/csv`) — the verify layer is already stdlib-only (`tools/verify.py:2`), so the Go port has no reason to add one. Tag `vX.Y.Z`; never move a published tag.

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it. Slices 1–3 are the honesty floor; 4–6 are the attestation/distribution layer.

1. **S1 — Freeze + declare.** Snapshot §2 exported surface into `cgo-seal` README as the `0.1.0` declared public API. Port `seal_findings`/`gate_frozen`/`prove_frozen` unchanged (they carry the fault proofs, `tools/verify.py:82-128`). Acceptance: `go test ./...` green on the extracted tree; README lists every §2 identifier with its `file:line`; `--selftest` equivalents pass with both controls + fault cases (`tools/verify_test.py:74-88` seal cases ported).
2. **S2 — Fix SKIP honesty + path hygiene.** Distinguish absent (exit 2) vs partial/corrupt (exit 1, named) (§5.2); canonicalise corpus/anchor paths, reject traversal (§5.7); report skipped files (§5.6). Acceptance: temp-copy fault proof — partial tree → FAIL (not SKIP), one corrupt + 158 good runs → skipped name reported, counts still `159/528/208/74/246` on the intact copy (`docs/VERIFY.md:24`).
3. **S3 — Corpus digest + hash8 honesty.** Implement sorted-manifest hash; keep `ProvenanceSHA` as renamed last-file helper; emit both into the bundle subject (`subject` = corpus digest per R2); document `hash8` as 32-bit label. Acceptance: mutating any non-last file changes the corpus digest; two `emit` runs over unchanged data are byte-identical (no `time.Now`, no local path in the bundle except caller-supplied); `TestProvenanceHash8` (`pkg/figures/figures_test.go:55-73`) still passes as the last-file pin.
4. **S4 — Signed anchor + level declaration.** Accept a detached signature on the anchor; add `--require-signed`; declare SLSA level in every `emit` (`L1` until a hosted builder exists, R1/R2). Acceptance: unsigned anchor + `--require-signed` → exit ≠ 0; signed anchor with bad key → named failure; `emit` output contains `builder.id` + level that matches the actual signing infrastructure.
5. **S5 — in-toto bundle + SBOM + sha256sum parity.** `emit` writes `<artifact>.intoto.jsonl` (R4) with Statement + SLSA predicate (R2 schema: `buildDefinition`, `resolvedDependencies`, `runDetails.builder`); each release ships CycloneDX SBOM of `cgo-seal` deps (R7) + `go.mod`/`go.sum`; fetch R8 §6.7 body and add the space-in-filename cross-check (§5.9). Acceptance: bundle parses against an in-toto validator (R6 spec); SBOM parses as CycloneDX and lists direct deps; `subject` digest matches the corpus; `sha256sum -c` and Go parser agree on happy path + escaping fixture.
6. **S6 — Recertify + release.** Re-run `cgo-seal verify` against the live frozen corpus, compare to `docs/VERIFY.md:24-26` (`159/528/208/74/246`, seal 159 OK, six gates 6 passed); publish the attestation alongside the repo release (R4 multi-place SHOULD); write the VERIFY row with command, host, date. Acceptance: independent verifier re-runs the command and owns the `VERIFIED` row (builder may not certify its own number); `cgo-seal version` == tag == release assets.

Each slice is independently reviewable. Slices 4–5 are the only ones that change the trust claim (unsigned → signed, JSON → attestation) and each must be verifier-re-run before its row is claimed.

## 8. Anti-toy criteria (what would make this a toy — and the bar for “not a toy”)

A `cgo-seal` submission is a toy if it re-describes its own outputs instead of attesting against an external check. It is **not** a toy only if all of the following hold:

1. **Independent reproduction of the frozen corpus.** `cgo-seal verify` over the sealed corpus reproduces `159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells` (`docs/VERIFY.md:24`) and the seal 159-OK (`docs/VERIFY.md:26`) — not just its own fixture tests. If a counting-rule change moves any of those numbers without a named `VERIFIED` row, it rewrites history.
2. **Fault-proof parity.** Every witness in `prove_frozen` (`tools/verify.py:82-128`) and `test_seal` (`tools/verify_test.py:74-88`) is ported and still catches: rewritten manifest (audit passes, seal catches), mutated data file (audit catches), added/deleted run, absent anchor — with the intact-copy control passing. A verifier without a fault proof is an opinion.
3. **Provenance honesty, SLSA-typed.** This is **Build L1 at best** until a hosted builder signs (R1/R3): the tool declares its level, binds `subject` to the corpus digest (R2), emits the predicate schema (R2), and never presents the last-file hash (`pkg/figures/figures.go:41-53`), the 8-hex `hash8` label, or the vendor string (`pkg/results/hardware.go:7-12`) as corpus proof. CAKE/MikroTik/statistics claims stay **unverified — primary source not found**; detailed `sha256sum` format rules stay **unverified — primary source not found** until R8 §6.7 body is fetched.
4. **No silent skip, no silent zero, no green SKIP.** Partial tree, corrupt file, skipped run, and absent `book/` are four distinct answers (`FAIL` / named skip / exit-2), never a zero count read as a measurement (R5 Step 2 reject-unrecognized, R6 monotonic principle). If two corpora can yield the same verdict with no way to tell which files were skipped, it is a helper, not a sealer.
5. **Verify-from-a-clean-clone + `sha256sum` parity.** A fresh clone (no `book/`) answers exit 2 naming the absence — never exit 0 — and a restored copy (`docs/VERIFY.md:37` second-copy drill) verifies green; `sha256sum -c docs/frozen-manifests.sha256` → 159 `OK` agrees with `cgo-seal verify` on the same tree (R8). The tool must survive the R4 distribution story: artifact + attestation `subject` digest matches a corpus the verifier assembled independently.
6. **Version discipline.** Declared API, immutable `vX.Y.Z` tags, changelog in version terms, and a release that carries bundle + SBOM + `go.sum` together (R4/R7). A repo that silently rewrites the anchor between tags, or ships no declared API, is a toy regardless of code quality.

---
*Research only. No source was modified in this pass; implementation of `cgo-seal` waits for human approval per AGENTS.md (“Stop and ask” — new dependency, scope growth, frozen-data adjacent). `book/` was never touched.*

*Skipped / not checked: live bench re-measurement (no Linux 6.8 host in this pass — `docs/VERIFY.md:28` BLOCKED), cosign/Sigstore signing-key choice (**unverified — primary source not found** for this org), GNU §6.7 sha2-utilities body beyond TOC listing (R8, §4.4 — detailed format rules marked unverified above), Go-module distribution mechanics beyond `go.mod:1-3` (fetched in sibling attestor research, not re-fetched here). Risk the user must know: the current `ProvenanceSHA`/`hash8` covers only the last file, so any figure/report/integrity claim citing it as corpus proof is weaker than it reads (`pkg/figures/figures.go:41-53` vs `docs/VERIFY.md:26`); and the anchor is unsigned, so a writer who can rewrite the anchor too is out of scope by design (`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md:6-9`).*
