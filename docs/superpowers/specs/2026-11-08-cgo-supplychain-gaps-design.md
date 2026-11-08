# CGO Supply-Chain Gaps — Exhaustive Design (All Missing Tools)

**Status:** approved direction, design-only. **No code changed. No commit.**

**Date:** 2026-11-08
**Branch:** `openchamber/operator-cockpit` · **HEAD:** `60de44b` (feat: SLSA L2 + SBOM supply chain for pure modules)

**Working dir:** `C:\Users\ASUS\.local\share\opencode\worktree\1c66ad257b56747880c45bc5d4f7e50f8688e1eb\operator-cockpit`

**Scope:** This spec closes the supply-chain and missing-tool gaps in the entire 10k-line research corpus. It is the single implementation-ready design for every gap the corpus names. Read this file as the predecessor that wins over prose in any checkpoint.

**Corpus (read exhaustively, not skimmed):**

- 36 files `docs/research/2026-11-07-module-*.md` (9,511 lines): airgap-distribution, api-doctor-iface, auth-ratelimit, campagne, cgroups-container, cli-binary, cli-frontend, cost-economics, defence-demo, deploy-supplychain, docs-persistence, doctor-preflight, ebpf-probe, ebpf-tcpinfo, economics-cost, evidence-attestor, evidence-seal, frontend-citation, iface-sampler, kit-vm-profile, master-dgm, metrics, model, mqtt-wis2, nginx-ingress, probe-audit, profile-registry, prometheus-metrics, qdisc-remediate, rate-alerting, results-figures, routeros-chr, slo-runbook, tui-kit-vm, tui-wall, verification-gates.
- `docs/superpowers/specs/2026-11-07-cgo-modular-extraction-design.md` (1,500 lines) — prior comprehensive design expanded from 137 lines to cover every tool. This spec is its supply-chain successor; where both speak, this spec wins for SLSA/CycloneDX/OCI/tags/GUI/WIS2/eBPF and the 34-tool index.
- `docs/research/2026-11-07-synthesis-modular-enterprise-master.md` (87 lines), `docs/research/2026-11-07-master-demonstration-dgm-needs.md` (159 lines), `docs/research/2026-11-07-modular-toolkit-morph.md` (185 lines).

**Method:** Every normative claim cites `file:line` (repo) or `(R#)` (fetched primary). Anything without a fetched primary is marked `unverified — primary source not found`. The marker is load-bearing: it blocks ranking that claim as a requirement. Search for the phrase to find all deliberate unknowns.

**How to use:** §0 Global Constraints are MUST for every slice. §1 is the 34-tool index. §§2–9 are the seven deep sections (SLSA L2, CycloneDX, Signed OCI, Tags, GUI, WIS2, eBPF). §10 is every other gap tool, one subsection per tool. §§11–20 are the cross-cutting contracts. §§21–24 are slices, supply-chain procedure, ledger/citation, and open decisions. Part II is Appendices A–L.

---
---

# PART I — GLOBAL CONSTRAINTS (§§0–24)

## 0. Global Constraints (MUST, no slice may violate)

### 0.1 The operator objective (220 ms)

- The CAP deadline objective is **220 ms with 95% target** (`pkg/metrics/cost.go:80-83` `CAPDeadlineMs=220`, `CAPBreachThresholdPct=95.0`).
- It is a **DGM operator objective, not a WMO mandate** (`cost.go:75-79`; WIS 2.0 Guide WMO-No. 1061, Vol. II, APPROVED v1.3.0, 2026-06-15, defines "real time" as "a few seconds to a few minutes" with no numeric per-bulletin threshold). Never write "WMO mandate" or "seuil OMM". Every user-facing string says "objectif DGM, cadre WIS 2.0".
- `AssessCAPSLA` breaches iff `pct < 95` strictly (`cost.go:105`); exactly 95% passes. Deadlines are strict `<` (`DeadlineOKPct`, `metrics.go:53`): exactly 220 ms breaches. Both boundaries are pinned by tests and must survive extraction unchanged.
- Parameters are caller-supplied (`DeadlineTarget{DeadlineMs, TargetPct}`), with `CAPOperationalTarget={220,95}` as an explicitly labelled opt-in default. No tool hard-codes 220/95 as a universal truth.

### 0.2 Absent ≠ 0 (everywhere)

- A missing measurement is **absent, never 0**. `Valid=false` emits no rate; `iface_valid:false` omits `cgo_iface_rx/tx_mbps` entirely (`pkg/iface/iface.go:28-41`).
- Companion flags are the only zero: `cgo_iface_valid 0` + `cgo_iface_gap{reason}`; `cgo_cap_sla_measured 0`; `Measured:false`. Never `0` as a measurement.
- Prometheus `absent()` / `absent_over_time()` is the absence alert shape. `== 0` on an empty vector matches nothing and must never be used as an absence test.
- `BuildEconomics` with no ratio/transport emits **zero scenarios**, never fabricated zeros (`cost.go:201-203`). `AssessCAPSLA` with no samples returns `Measured:false` (`cost.go:99-102`). `Scan` empty returns explicit empty-with-diagnostics, never success-with-zero.
- The `pkg/api/metrics.go:46-53` six static `"0"` endpoint is the canonical anti-pattern. Any slice that ships hard-coded `"0"` outside `*_measured/_valid` companions fails review.

### 0.3 Frozen corpus untouched

- `book/data/runs` and the frozen corpus: **no new run, rewrite, or reshape**.
- `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv` schemas: **never touch**. Any CSV header reorder is a MAJOR bump with migration note, never a patch.
- `epbfBBR.md`, `backup/`, `data/`, user documents: never touch.
- `.dist` execution (no field trace at the sample floor): never execute.
- The frozen counts **159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells** (`docs/VERIFY.md:24`) must be reproduced by every attestation path until a named `VERIFIED` row moves them. A counting-rule change without a named row rewrites history and is a FAIL.
- `ProvenanceSHA` hashes only the lexically-last `aqm_eval.csv` (`pkg/figures/figures.go:41-53`); it is a **last-file label, not corpus proof**. The corpus digest (sorted manifest hash) is the subject. Never present `hash8` (8 hex, 32 bits) as a proof.

### 0.4 Human gates (stop and ask)

- **Never:** live bench without explicit human approval · destructive actions · merge to `main` without human decision · secret exposure · frozen-data change · **new dependency without approval**.
- Each new dependency is a stop-and-ask: `cilium/ebpf` (eBPF), `prometheus/client_golang` (if exporter delegates), `golang.org/x/time/rate` (if Go limiter chosen), MQTT client lib (WIS2), QEMU + CHR image (CHR), RabbitMQ/Mosquitto broker, any YAML lib, any frontend framework swap. The default is **stdlib-only, zero new `require`** and every target below is written to hold with zero.
- Builder ≠ verifier. The builder never certifies its own number; the verifier re-runs the command and owns the `VERIFIED` row (`docs/VERIFY.md:83-91`). "Windows-green" is not "done".
- At most three active slices. A fourth active slice is refused until one closes or is deferred.
- `git add -A` is forbidden. Stage explicit paths. Never commit: `epbfBBR.md`, `critical-review-*.md`, `session-*.md`, `handoff-*.md`, `backup/`, `data/`, `book/` changes, `testbed-bootstrap/`, `unp*/`, `*.docx`, `*.pdf`, `cgo.exe`, archives.
- Never `git merge` to `main` without human decision. Never run runtime `clang/gcc/llvm/linux-headers` on the VM; the eBPF object is prebuilt.
- Never touch Graylog or `100.114.101.13`.

### 0.5 The one command

- `make verify` — all six gates. `make verify-fast` — no Go suite. `make verify-selftest` — the gates' own fault proofs. On a host without `make`, run `python tools/verify.py` with the same flags and say so.
- Exit `0` = every gate passed. Exit `1` = named gate failed. Exit `2` = a gate could not run; **2 is not green**, and the skipped gate is named. Never report missing prerequisite as passed.
- Verdict mapping is `1` if any FAIL, else `2` if any SKIP, else `0` (`tools/verify.py:35-36`). A green run on a partial tree is a false claim.

### 0.6 Language

- French comments and user-facing text; English identifiers, specs, plans, gate messages, and ledgers. A file keeps the language it already uses. Operator banners (IFI/WIS/CAP verdict lines) keep their French contract text verbatim.

### 0.7 Honesty invariants (docs/VERIFY.md + code)

- No-load ⇒ no-grade (`TestVerdictNeverGreenWithoutGrade`).
- Unmeasured ⇒ `Unavailable`, never zero, never pass.
- Absent severs from measurement.
- 220 ms = operator objective (above). `bbr` = BBRv1 in the thesis. Audit says "achieved saturation throughput", never "link capacity".
- `RECORDED` is not promoted by confidence, repetition, or later agreement (`docs/VERIFY.md:87-88`).
- Every slice ships a README declaring the public API (SemVer §1). No README without `unverified — primary source not found` where applicable is a toy.

---

## 1. Target Module Structure — The 34 Tools (Index)

All paths are placeholders. The Go Modules Reference requires choosing unique controlled paths **before first publish**; reusing `github.com/Realms4239/cgo` for v2+ needs a `/v2` suffix. Suggested layout: one repo per versioned unit, `go 1.25` floor from `go.mod:3` (mandatory minimum since Go 1.21). Every module carries `go.mod`, `README.md`, `LICENSE` (MIT), `CHANGELOG.md`, tests, SBOM (CycloneDX), and SLSA provenance. Founder-cited lines below are the registration pin for each row.

| # | Tool (repo/module) | What it does | Founder pin |
|---|---|---|---|
| T01 | `cgo-model` | Domain types P1–P4, qdiscs, CCs, gates G0–G7, events, headers, `ParseCCList` | `pkg/model/types.go:11-184`, `cc.go:8-19` |
| T02 | `cgo-metrics` | Percentiles, summaries, bufferbloat grades, JFI, stats, verdict cascade | `pkg/metrics/metrics.go:13-219`, `bufferbloat.go`, `stats.go`, `verdict.go` |
| T03 | `cgo-cost` | Models A/B/emergency economics + 6 retail tiers + CAP SLA | `pkg/metrics/cost.go:12-238`, `metrics.go:72-156` |
| T04 | `cgo-profile` | Registry + persistence (`Import/Load/Delete/Reserved`, `data/profiles.json`) | `pkg/profile/profile.go:12-93` |
| T05 | `cgo-results` | Frozen scan → medians, quarantine honesty, provenance `Runs` | `pkg/results/results.go:13-236`, `csvrows.go` |
| T06 | `cgo-figures` | SVG bars+scatter + `ProvenanceSHA/Hash8` | `pkg/figures/figures.go:17-156` |
| T07 | `cgo-qdisc` | `tc` Reset/ApplyNetem/ApplyShaper/Stats/Calibration | `pkg/qdisc/qdisc.go:23-149`, `ge.go`, `dist.go`, `stats.go`, `ingress.go` |
| T08 | `cgo-remediate` | Pure Linux + RouterOS generators from measured `Link` | `pkg/remediate/remediate.go:22-417` |
| T09 | `cgo-probe` | Ping/small/bulk, CC sockets, TCP_INFO fallback | `pkg/probe/probe.go:28-139`, `bulk*.go`, `*_linux.go` |
| T10 | `cgo-bpf-observer` | eBPF sidecar + loader (CO-RE object, maps, fallback ladder) | `pkg/probe/bpf/types.go`, `status.go`, `bpf_probe.c`, `BUILD.md` |
| T11 | `cgo-audit` | Phased audit → sealed report + telemetry + RouterOS compare | `pkg/audit/audit.go:16-717`, `report.go`, `telemetry.go` |
| T12 | `cgo-campagne` | Matrix orchestration, watch sentinel, valuation sidecar | `pkg/campagne/campagne.go`, `matrix.go`, `watch.go` |
| T13 | `cgo-api` | REST + SSE hub (stdlib transport) | `pkg/api/server.go:380-1432`, `sse.go`, `metrics.go` |
| T14 | `cgo-doctor` | Read-only capability preflight | `pkg/doctor/doctor.go:16-183` |
| T15 | `cgo-iface` | `/proc/net/dev` rate sampler with gap contract | `pkg/iface/iface.go:16-200` |
| T16 | `cgo-exporter` | Hardened Prometheus exposition + `rules.yml` compatible series | `pkg/api/metrics.go:10-64`, `deploy/prometheus/rules.yml` |
| T17 | `cgo-slo` | SLO generator (spec→rules/test/runbook/alertmanager) | `deploy/prometheus/rules.yml`, `docs/runbooks/smoke-alerts.md` |
| T18 | `cgo-alerting` | Probe-dead/iface/SLO alert paths (or folded into exporter+slo) | `pkg/campagne/watch.go`, `rules.yml`, `pkg/api/metrics.go` |
| T19 | `cgo-guard` | Bearer + CSRF + rate-limit + Nginx fragment | `pkg/api/auth.go`, `server.go:1438-1450`, `deploy/nginx/cgo.conf` |
| T20 | `cgo-ingress` | TLS reverse-proxy conf as its own OCI family | `deploy/nginx/cgo.conf:1-35` |
| T21 | `cgo-server` | `cmd/cgo --serve` composition + embedded SPA | `cmd/cgo/server.go:54-391`, `web/frontend/embed.go` |
| T22 | `cgo-kit` | Deployment engine + VM drivers + bench scripts (binary, not library) | `internal/kit`, `internal/vm`, `kit/*.sh` |
| T23 | `cgo-measure` | `audit + run + verify/figures/testbedsrv/calibrate` verbs | `cmd/cgo/main.go:148-298`, `runcli.go`, `calibrate.go` |
| T24 | `cgo-tui` | Operator + kit TUIs (Bubble Tea, promoted from `package main`) | `cmd/cgo/tui.go`, `tui_audit.go`, `kittui.go` |
| T25 | `cgo-wall` | Web wall + Grafana + Prometheus display (display-only) | `web/frontend/src/App.tsx`, `LiveView.tsx`, `VerdictHero.tsx` |
| T26 | `cgo-gui` | Fyne/Win32 dual-view console (separate repo + build constraints) | `cmd/cgo-gui/` (core/win32/fyne) |
| T27 | `cgo-chr-validate` | RouterOS QEMU/serial validator + parser/diff/proof | `cmd/cgo-chr/`, `pkg/audit/routeros_*.go`, `kit/chr/` |
| T28 | `cgo-verify` | Six-gate runner core + hooks + project adapter | `tools/verify.py:1-527`, `tools/verify_test.py` |
| T29 | `cgo-seal` / `cgo-attest` | Corpus seal + provenance + headers + sidecar + in-toto bundle | `docs/frozen-manifests.sha256`, `verify.py:48-191` |
| T30 | `cgo-ledger` / `gate` / `cite` | Ledger writer / gate runner / citation emitter | `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md` |
| T31 | `cgo-cite` | Citation purity + proof panel (TS + Go) | `web/frontend/src/lib/citation.ts`, `AuditReportCard`, `ArchiveTable`, `ProofDrawer` |
| T32 | `cgo-wis2-probe` | Lab MQTT observer (DEFERRED, never inside CGO) | (absent — grep zero in `pkg/`+`cmd/`; §7) |
| T33 | `cgo-demo` | Doctor-verify sequencer of the falsifiable 5-min arc | `tools/verify.py`, `audit-archive.py`, `frozen-manifests.sha256` |
| T34 | `cgo-dist` / `cgo-isolate` | Bundle builder + cgroup/OCI isolation + air-gap loader | `internal/kit/package.go`, `deploy/Dockerfile`, `Makefile` |

**Dependency DAG (no cycles).** Layer 0 (stdlib-only): `model`, `iface`, `doctor` parser, `remediate` generators, `verify` core, pure predicates. Layer 1: `metrics`, `profile`, `qdisc`, `probe`, `cost`. Layer 2: `results`, `figures`, `audit`, `campagne`. Layer 3: `api`, `doctor`, `iface`, `exporter`, `slo`, `guard`. Layer 4 (binaries): `server`, `kit`, `measure`, `tui`, `wall`, `gui`, `chr-validate`, `verify`, `seal/attest`, `ledger/cite`, `cite`, `wis2-probe`, `demo`, `dist/isolate`. Sidecars read evidence, never write it (Correction 2: sidecars never inject the write that overwrites evidence).

**Cutover rule (every sibling-upgrade):** (1) publish new module tag, (2) bump `require`, (3) run `go mod tidy`, (4) verify `go list -m all`, (5) `goreleaser release --clean` dry-run, (6) update every `keep_version` entry in `mise.lock`/`package.json`, (7) run `make verify-fast` + selftest, (8) verifier re-runs. No `replace` survives a tag. No tag ever moves.

---

## 2. SLSA Provenance — Full SLSA v1.0 + v1.1 + v1.2 (L2) Requirements

Primary sources (this pass): SLSA spec pages via `slsa.dev` (fetched this session), distilled below verbatim in substance. The React docs and SLSA requirement text are citations, not paraphrase risk: short phrases are quoted. Anything SLSA-conventional but not fetched is marked `unverified — primary source not found` where it appears in the research files.

### 2.1 Levels (SLSA v1.0 `spec/v1.0/levels`, v1.1 `spec/v1.1/provenance` + `requirements`, v1.2 whats-new)

| Level | Name | What it requires |
|---|---|---|
| **L0** | No guarantees | Nothing. Provenance absent, builders arbitrary. This is the repo today for distribution artifacts (from-source build + manual copy, `docs/deploy.md:42-61,108-118`) and the ceiling for the frozen seal if read as an attestation (unsigned static anchor, `docs/VERIFY.md:26` "Proves a static anchor only: not a signature"). |
| **L1** | Provenance exists | Producer follows a consistent build process; provenance exists describing build platform/process/top-level inputs and is distributed to consumers. Provenance **may be incomplete and/or unsigned** at L1 ("best effort" completeness acceptable). The platform **MUST** generate provenance that unambiguously identifies the output package by cryptographic digest. |
| **L2** | Signed provenance, generated by a hosted build platform | L1 + provenance is **Authentic**: signed by a key accessible only to the platform component (via transparency log or timestamping — RECOMMENDED), generated by dedicated (hosted) infrastructure, validatable by consumers. L2 adds hosted (not an individual workstation). Signing-key custody and hosted-builder choice: `unverified — primary source not found` for this org (no policy doc read); candidate routes recorded in §2.5. |
| **L3** | Hardened builds (unforgeable + isolated) | L2 + provenance is **Unforgeable**: signing secrets in secure management inaccessible to user-defined build steps, every field generated/verified by the platform. Builds are **isolated/ephemeral**: no cross-build influence, no cache poisoning, ephemeral environments. L3 is a stated non-goal until builds are isolated — claiming L3 without the platform is exactly what SLSA warns provenance may be bypassed at L1. |

**SLSA v1.2 additions (fetched "What's new" page).** Source track + VSA (Verification Summary Attestation) guidance. The spec's `predicateType` URI carries the major version; consumers MUST ignore unrecognized fields. Extension fields SHOULD be `<vendor>_<field>` and MUST NOT alter the meaning of standard fields (monotonic principle: ignoring MUST never turn DENY into ALLOW). The frozen system is honestly **L0+ / L1-at-best**: tracked anchor + fault proofs exist (`VERIFY.md:26`, `verify.py:82-128`), but there is no signature, no hosted builder, no transparency log (`specs/2026-10-28-cgo-corpus-seal-design.md:90-95`). It must declare L1, emit the predicate schema so v1.2-era verification can run, and never claim L2/L3 without signed hosted isolated builds.

**SLSA v1.1 `provenance` predicate (R2 of seal research).** `predicateType` MUST be `"https://slsa.dev/provenance/v1"` (carries major version). REQUIRED at L1 for `buildDefinition` (`buildType` URI + `externalParameters` + `resolvedDependencies[]`) + `runDetails` (`builder.id` URI + `metadata.invocationId/startedOn/finishedOn`). `resolvedDependencies` are `[ResourceDescriptor]` with `uri` + `digest` (`sha256`, `gitCommit`, …). `builder.id` is the **sole determiner** of the SLSA Build level and MUST reflect the trust base; consumers MUST accept only specific signer-builder pairs. Parsing: consumers MUST ignore unrecognized fields.

**Distributing provenance (v1.0 R4).** Attestations SHOULD be bound to artifacts, not releases; accompany the artifact at publish time; be published in at least one place, SHOULD be in more than one (source-repo release, package registry sidecar, transparency log); SHOULD be immutable — a new attestation for the same artifact MUST NOT overwrite the old; in-toto recommends `<filename>.intoto.jsonl`. Source-based ecosystems (Go via module proxy) SHOULD provide build provenance for source archives since the proxy transforms the repository.

**Verifying artifacts (v1.0 R5).** Step 1 (level): verify envelope signature from roots of trust, verify `subject` matches artifact digest, verify `predicateType`, look up level from (public key, `builder.id`). Step 2 (expectations): check `builder.id`, canonical source repo, `buildType`, `externalParameters`, reject unrecognized fields. Step 3 (optional): check `resolvedDependencies` recursively. Expectations may be trust-on-first-use, defined by producer, or defined in source.

### 2.2 SLSA requirements per tool

| Req | What it demands of every shipped artifact |
|---|---|
| **Provenance Exists** (L1) | `provenance.json` per asset: `{artifact digest (sha256), cgo_commit, builder, cgo-dist.toml hash, toolchain pin}`. The `frozen` row's `docs/VERIFY.md:93-115` signature block is the human-readable ancestor ("Session id, host, branch/HEAD, per-gate table") — the machine `runDetails` is its structured twin. VERIFY rows already carry the R1-shape fields informally: Claim ≈ `subject`, Evidence ≈ `buildDefinition.externalParameters`, Date/Host ≈ `runDetails.metadata`, seal ≈ `resolvedDependencies`. Missing: `predicateType`, authenticity, `builder.id` URI. |
| **Provenance is Authentic** (L2) | Sign provenance with a platform-held key; publish in ≥1 place (release assets + OCI referrer); verify with `cosign verify <image>@<digest>` + provenance check green in CI. Until hosted signing exists, state L1 in release notes — never imply L2/L3. |
| **Provenance is Unforgeable** (L3) | Non-goal until isolated/ephemeral builds exist. When reached: secrets in secure management, every field platform-generated. |

### 2.3 Gap analysis (SLSA)

1. At best the repo claims SLSL L0 (no provenance distributed). `docs/VERIFY.md:26` seals the *frozen corpus* with a hash list and explicitly narrows itself ("Proves a static anchor only: not a signature") — correct, and exactly the ceiling the *distribution* currently shares. A consumer of a release bundle today has nothing equivalent even at that level.
2. No `.goreleaser.yaml` → no `sboms:` block → no SBOM ever cataloged. No `checksum:` → hand `sha256sum` drifts. No reproducible build definition → host-dependent `go build` + whatever `web/frontend/dist` holds; two runs differ (frontend timestamps, build paths). "Follow a consistent build process" cannot be claimed, let alone proven.
3. `kit verify` (`package.go:147`) recomputes hashes on the VM with **no expected-value anchor** — a measurement, not a verification. R4 (Provenance Exists) + R3 (build-generated checksums *with* the release) both require the anchor *beside* the artifact.

### 2.4 Target (per artifact family)

- **Binaries + archives:** `.goreleaser.yaml` `builds` (`cmd/cgo`, linux/amd64 + windows/amd64, `-ldflags "-X main.version={{.Version}}"`), `archives` (`cgo-<os>-<arch>.tar.gz`/`.zip`), `checksum` (sha256), `sboms:` (R2/R3 pipeline, §3), `signs:` for archives/checksums. GitHub release publishes SBOM + checksums as first-class assets.
- **OCI images:** multi-arch index, `org.opencontainers.image.*` labels at build time (revision = git SHA; exact label-key namespace `unverified — primary source not found` beyond image-spec repo layout in this pass), digest-pinned publish, cosign sign, provenance as OCI referrer/attestation. GoReleaser's image non-catalog limitation means OCI SBOM is off-image (§3).
- **Air-gap bundle:** bundle-level `provenance.json` (L1 digest + inputs + builder + toml hash) + `MANIFEST.json` + detached `checksums.txt` + signed checksums. Bundle states its level per bundle (§3.4).
- **Corpus seal:** `cgo-seal emit` produces in-toto Statement (`subject`=corpus digest, `predicateType=https://slsa.dev/provenance/v1`, per-run `resolvedDependencies`, `builder.id=cgo-seal/<version>` honest id, never a hosted-builder URI until one exists). `cite <row-id>` produces the same for any `VERIFIED` row.
- **SLSA level declaration is itself a shipped artifact.** Every bundle and every release note states the achieved level (L1, or L2 if hosted signing is in place). Never imply more.

### 2.5 Key-custody decision (OPEN, human-gated)

Keyless OIDC (Sigstore) vs. self-managed keys: `unverified — primary source not found` for this org. R6 (sigstore index pages fetched; detailed key-custody/bundle-format requirements `unverified` in this pass). Decision recorded in `cgo-dist.toml`, not assumed. The bundle must carry verifiable signatures such that verification evaluates with the trust root shipped *inside* the bundle or pre-installed — not fetched at verify time. Transparency-log *online* verification semantics: `unverified` in this pass.

### 2.6 Per-tool slices (SLSA)

- **SLSA-S1 — L1 everywhere (no signing infra).** Publish `provenance.json` with every release; state L1. Acceptance: every asset has digest + commit + builder; release notes state L1.
- **SLSA-S2 — L2 sign + host (hosted CI).** Cosign sign + provenance attestation behind hosted CI. Acceptance: `cosign verify` + provenance check green in CI; compose pulls the digested image.
- **SLSA-S3 — Corpus attestation (L1 corpus).** `cgo-seal emit` / `cgo-attest emit` produces in-toto Statement with `subject`=corpus digest, `predicateType=https://slsa.dev/provenance/v1`, per-run `resolvedDependencies` (159 manifests), `builder.id=cgo-seal/<version>`. Acceptance: bundle parses against in-toto validator; `subject` digest matches `sha256sum -c docs/frozen-manifests.sha256` (159 OK).
- **SLSA-S4 — Signed anchor + L2 corpus (conditional).** Accept detached GPG signature on anchor; `--require-signed`; declare L1 until hosted builder exists. Acceptance: unsigned + `--require-signed` → exit ≠ 0; signed with bad key → named failure.
- **SLSA-S5 — CI hosted builder (L2 for distribution).** Hosted CI builds + signs + publishes provenance. Acceptance: provenance signature validates from fresh clone with trust root in bundle; release notes state L2.

### 2.7 Anti-toy refusal list (SLSA)

1. No provenance file beside the artifact — "trust the zip" fails.
2. Level implied but not earned (claiming L2/L3 on a workstation build) — citation theatre, fails.
3. Unsigned treated as signed — a hash list presented as signing fails.
4. Provenance that does not identify the artifact by digest fails R5 S3.
5. New attestation overwriting old for same artifact fails R4 immutability.
6. Builder id that is a hostname string presented as hosted-builder URI fails R1 Builder.

### 2.8 DEFERRAL conditions (SLSA)

- L2/L3 stay DEFERRED until hosted builder + key custody decision (human, §2.5). Until then every release states L1.
- Unforgeable/isolated (L3) is a stated non-goal until builds are isolated/ephemeral. Do not scaffold an L3 claim.
- Corpus signing needs the signed-anchor decision (S4) and remains L1 until then.

---

## 3. CycloneDX SBOM — Full Spec
### 3.1 What CycloneDX requires (R7 + v1.5 JSON Reference)
CycloneDX "represents comprehensive inventories of software components, dependencies, and relationships… in a machine-readable format" to identify outdated components, licensing conflicts, vulnerabilities (Introduction). It "goes beyond simple inventories" via dependency hierarchy, not a flat package list.
Top-level envelope (v1.5 JSON): `bomFormat: "CycloneDX"` (MUST be exactly that), `specVersion`, `serialNumber` SHOULD have unique serial (RFC-4122 UUID) even if contents unchanged, `version: 1` bumped on modification, `metadata.timestamp` (date-time), `components[]` each with `type` (application/framework/library/container/platform/os/device/device-driver/firmware/file/ml-model/data), `name` REQUIRED, `version`, `hashes[]` (`alg` ∈ MD5/SHA-1/SHA-256/…/BLAKE3 + `content`), `licenses[]` (SPDX id or named), `supplier` name/url/contact, `scope` ∈ required/optional/excluded, unique `bom-ref` per component.
The hierarchical dependency graph is the requirement, not a bare package list. A bare file list is not an SBOM.
### 3.2 GoReleaser SBOM pipeline (R2/R5)
GoReleaser `sboms:` catalogs one or more artifacts into documents in `dist/`; defaults `cmd: syft`, `artifacts: archive`, documents `{{ .ArtifactName }}.sbom.json`, `args: ["$artifact","--output","spdx-json=$document","--enrich","all"]`; `artifacts` may be `archive`, `binary`, `source`, `package`, `installer`, `diskimage` (Pro), or `any`.
**Limitation (R2, verified): "Container images generated by GoReleaser are not available to be cataloged by the SBOM tool."** A GoReleaser SBOM pipe alone leaves the OCI image uncovered; the image SBOM must be produced off-image and attached as a separate asset/attestation. Exact syft/cosign CLI flags beyond the R2/R6 index pages: `unverified — primary source not found` in this pass and not used as requirements.
### 3.3 Per-tree gap (today)
Zero `cosign`/`sigstore`/`syft`/`sbom`/`cyclonedx` *configuration* in-tree (grep). `.goreleaser.yaml` absent. `sboms:` absent → no SBOM ever cataloged (R2). `internal/kit/package.go` zips contain binaries + scripts + README and nothing describing components (R5 unmet). Hygiene (`verify.py:344-388`) is a denylist (absence of known-bad shapes), never a positive `components[]` enumeration (R5 gap). No `serialNumber`/`timestamp`/`specVersion` envelope fields emitted.
### 3.4 Target SBOM set (every release)
| Artifact | SBOM source | Format | Location |
|---|---|---|---|
| `cgo-<os>-<arch>.tar.gz`/`.zip` (archives) | `syft $artifact --output cyclonedx-json` via GoReleaser `sboms:` | CycloneDX JSON | release asset + bundle file `SBOM.cyclonedx.json` |
| `cgo` binary (direct) | same pipe (`artifacts: binary` or `any`) | CycloneDX JSON | release asset |
| OCI image | **off-image** `syft ghcr.io/.../cgo@sha256:... --output cyclonedx-json` → referrer/attestation (R2 limitation) | CycloneDX JSON | referrer or bundle file |
| Each Go module (`go.mod` deps) | GoReleaser `sboms:` or tool SBOM | CycloneDX `components[]` from `go.mod:5-10` expanded transitively | release asset per module |
| `cgo-attest`/`cgo-seal`/`cgo-verify` tool deps | tool SBOM | CycloneDX | release asset ("SBOM of the tool", never presented as proof of the corpus) |
| **Frozen corpus** | **NOT an SBOM** — seal + per-run manifests + attestation bundle is the SBOM-equivalent. Conflating the two would be a toy. | seal shape (§2/§29) | `docs/frozen-manifests.sha256` + bundle |
### 3.5 Format choice
Pick one (CycloneDX vs SPDX) per release and ship it **as a file in the bundle** — a URL is useless offline. CycloneDX is the default in this spec (`--output cyclonedx-json`, matching R2 `spdx-json` alternative). SPDX kept as an `alternative` if a consumer needs it; never both silently.
### 3.6 Per-tool slices (CycloneDX)
- **SBOM-S1 — Archives + binaries (GoReleaser).** Add `.goreleaser.yaml` `sboms:` + `checksum:`. Acceptance: pre-release emits archives + `checksums.txt` + per-archive CycloneDX JSON; JSON parses, lists `go.mod:5-10` deps expanded, `sha256sum -c` passes.
- **SBOM-S2 — OCI image SBOM (off-image).** `syft <image> --output cyclonedx-json` → bundle file / referrer. Acceptance: image SBOM present in bundle as file; `components[]` covers direct deps expanded.
- **SBOM-S3 — Tool SBOMs + envelope fields.** Every module/tool release ships its own CycloneDX with `bomFormat/specVersion/serialNumber/version/timestamp/components{hashes,licenses,supplier,scope,bom-ref}`. Acceptance: SBOM parses against CycloneDX v1.5 reference fields; `bomFormat == "CycloneDX"`.
### 3.7 Anti-toy (CycloneDX)
1. Toy SBOM (bare file list, no dependency graph) fails — R7 requires hierarchical relationships.
2. URL-only SBOM (not a file in bundle) fails — offline useless.
3. Tool SBOM presented as corpus proof fails — the corpus SBOM-equivalent is the seal + manifests + bundle.
4. Denylist presented as inventory fails — `hygiene` proves absence, not components.
### 3.8 DEFERRAL (CycloneDX)
No deferral — SBOM-S1..S3 are in the critical path before any Grafana dashboard. OCI label-key namespace for SBOM attachment: `unverified — primary source not found` beyond repo layout; use registry referrer or bundle file whichever the operator's registry supports (human decision).

---

## 4. Signed OCI — Full Image / Distribution / Runtime Specs
### 4.1 The three specs (R8/R23)
OCI is three specs — Runtime (run a filesystem bundle), Image, Distribution — implementation downloads image, unpacks to runtime bundle, runtime runs it (`R8` overview). The `scratch` + `COPY` output must therefore be a valid image manifest: `config` descriptor + ordered `layers` descriptors (Image Manifest: "base layer at index 0… final filesystem layout MUST match applying layers to an empty directory").
Descriptor `digest` is `algorithm:encoded`, e.g. `sha256:<64 hex>` (Digests). "Retrieved content SHOULD be verified against this digest when consumed via untrusted sources." "Implementations MUST implement SHA-256 digest verification."
Distribution pull: reference "**MUST be either (a) the digest of the manifest or (b) a tag. `MUST NOT` be in any other format.**" "If the `name` part of a manifest request is a digest, clients SHOULD verify the returned manifest matches this digest." **`:latest` is case (b), a mutable tag — by R7 the registry may return a different manifest for the same tag tomorrow. "Digest-pinned" means case (a) `@sha256:...` and nothing else.** Tag vs digest: "A manifest digest may have zero, one, or many tags referencing it" — pinning by digest does not require the tag to exist at pull time.
`EXPOSE` "doesn't actually publish the port… functions as a type of documentation" (Dockerfile ref). Publish is via `-p`/`-P`; `deploy/docker-compose.yml:4-5` does the real publish.
`FROM scratch` "signals… that you want the next command in the `Dockerfile` to be the first filesystem layer" and the image "can only be successfully executed as long as the `hello` binary doesn't have any runtime dependencies". Exec-form `ENTRYPOINT` is "the preferred form" — `deploy/Dockerfile:4` complies.
### 4.2 Config / runtime needs (R5 OCI Runtime Spec `config-linux.md`)
The runtime config that *actually* carries limits is `linux.resources`: `cpu{shares,quota,burst,period,realtimeRuntime,realtimePeriod,cpus,mems}` and `memory{limit,reservation,swap,kernel,kernelTCP,swappiness,disableOOMKiller,useHierarchy,checkBeforeUpdate}`. `checkBeforeUpdate` exists precisely because "when cgroup v2 is used, an OOM killer is invoked" [for a limit below current usage]. `quota`/`burst`/`period` are microseconds, matching R1 `cpu.max` units. "Runtimes MAY attach the container process to additional cgroup controllers beyond those necessary." So the container's cgroup tree may contain more than `cpu`/`memory`.
### 4.3 Signing layer (cosign/sigstore)
Signing containers and verifying signatures are separate, first-class operations (R6 index: "Signing Containers", "Verifying Signatures", "In-Toto Attestations", "Key Management"). Offline consequence: the bundle must carry verifiable signatures (or at L1, detached checksums + provenance) such that `cosign verify <image>@<digest>` (or blob equivalent) can be evaluated with the trust root shipped *inside* the bundle or pre-installed — not fetched at verify time. Transparency-log *online* verification semantics: `unverified — primary source not found` in this pass (only docs landing pages fetched). Key-custody choice (keyless OIDC vs self-managed) is `unverified — primary source not found` for this org (§2.5).
### 4.4 `FROM scratch` + labels routing (R6/R9/R10)
Today `deploy/Dockerfile:1-4`: `FROM scratch`, `COPY bin/cgo /cgo`, `EXPOSE 9090`, `ENTRYPOINT ["/cgo","--serve"]`. No `USER`, `HEALTHCHECK`, `LABEL`, `STOPSIGNAL` (all Dockerfile instructions, R9 table). No `--platform`, no multi-stage, no digest on `FROM`. `org.opencontainers.image.*` labels (title/version/revision/source/created) are the standard annotation surface (R6 `annotations.md`); exact label-key namespace: `unverified — primary source not found` in this pass beyond the image-spec repo layout. Archive is implicit: binary is `bin/cgo` whatever the build host produced; no `GOOS/GOARCH` pin or multi-arch manifest (R6 image-index concept, not used; a Windows `.exe` copied in breaks silently; cf. `docs/VERIFY.md:27` cross-build note on `cgo-gui`).
### 4.5 `docker-compose.yml` + `Dockerfile` surface (today)
| Key | Value | Location | Meaning |
|---|---|---|---|
| `services.cgo.image` | `ghcr.io/realms4239/cgo:latest` | `docker-compose.yml:3` | Mutable tag (R2/R7), **not** a digest |
| `ports` | `127.0.0.1:9090:9090` | `:4-5` | Loopback-only publish (`-p` semantics) |
| `deploy.resources.limits.cpus` | `"1.0"` | `:9` | Fractional CPU; platform MUST prevent exceeding (R4) |
| `deploy.resources.limits.memory` | `512M` | `:10` | Byte-string; platform MUST prevent (R4) |
| `environment` | `CGO_API_TOKEN=${CGO_API_TOKEN:?set}` | `:11-12` | Required-at-up env; fail-fast (`:?set` semantics `unverified — primary source not found` in this pass) |
| `restart` | `unless-stopped` | `:13` | Top-level service key, not `deploy.restart_policy` |
Absent: `deploy.resources.reservations`, `limits.pids`, `cap_add: [NET_ADMIN]`, `devices`, `healthcheck`, second service (no Nginx/Prometheus). See §23 isolate for the full cgroup treatment (T34).
### 4.6 Gap: the `:latest` contradiction
`docs/deploy.md:183` says "digest-pinned image … rollback is compose pull of the prior digest" while `docker-compose.yml:3` ships `:latest` (mutable). R4 (semver) §3 forbids modifying a released version's contents. R7 makes tags mutable. The doc is literally false of the artifact. Fix is a versioned tag + digest in compose and a rollback note naming *which* prior release (§4.8, §23.9).
### 4.7 `DIRECTIVE KEY + capability gap`: container cannot shape
No `NET_ADMIN` capability (compose) vs Linux-full shaping requirement (`docs/deploy.md:79-87` `CAP_NET_ADMIN` for `tc`/netem). Container as shipped = observation mode only. The systemd binary path (`setcap`, `docs/deploy.md:82-87`) and container path diverge silently. One-line fix (`cap_add: [NET_ADMIN]`) plus README note — or restrict compose to observe mode explicitly. Whether container `NET_ADMIN` reaches the host qdisc: `unverified — primary source not found` (netns-scoped; flag for owner; README must state which: in-container `tc` vs host-side `cgo shape --restore`).
### 4.8 Target OCI publication (per image)
- Multi-arch image index from the scratch Dockerfile, with `org.opencontainers.image.title/version/revision/source/created` labels at build time (revision = git SHA).
- Publish to `ghcr.io/realms4239/cgo` **by digest**; change compose to `image: ghcr.io/realms4239/cgo@sha256:...` (or at minimum `:<semver>`), so the deploy.md rollback sentence becomes literally true.
- Sign with cosign and attach provenance as OCI referrer/attestation so SLSA consumers can verify (L2). OCI SBOM off-image → referrer or bundle file (§3.2).
- Advertise SLSA level per release (L1 with release; L2 once hosted CI signs; L3 non-goal).
- Compose: `cap_add: [NET_ADMIN]`, `healthcheck:` calling `GET /api/health`, `digest.txt` sidecar recording prior digest before upgrade (`{"digest":"sha256:...","size":...}` descriptor shape, R6).
### 4.9 Per-tool slices (OCI)
- **OCI-S1 — Make deploy docs true (no signature infra).** Compose → `@sha256` + `cap_add`, or split `compose.observe.yml`/`compose.full.yml`; `cgo.conf` → template cert paths + non-colliding upstream; rollback names actual prior digest. Acceptance: `docker compose config` succeeds, `nginx -t` passes with test cert, rollback matches compose. Priority: the two lies (digest, shaping).
- **OCI-S2 — Digest-pinned publish + inspect gate.** Replace `:latest` with `@sha256` (R7); `digest.txt` prior; CI `docker inspect` asserting 536870912/1000000000. Acceptance: tampered digest fails; inspect matches.
- **OCI-S3 — Sign + provenance (L1→L2).** Digest-pinned publish, labels, cosign sign, SLSA attestation behind hosted CI. Acceptance: `cosign verify <image>@<digest>` + provenance green in CI.
### 4.10 Anti-toy (OCI)
1. Mutable identity (`:latest`, branch tags, hand-copied binaries) fails — digest must be immutable once published.
2. Unverifiable payload (no checksum/SBOM beside binary/image) fails.
3. PDF-ish identity (no labels, `cgo version`≠tag, no revision) fails L1 minimum.
4. Capability lies (compose promises Linux-full the container cannot exercise; TLS conf fails `nginx -t`) fails.
5. Secret in image (token baked/logged) fails (`CGO_API_TOKEN:?set` + hygiene must stay green).
6. No rollback (no prior digest + one-command restore) is a trap.
### 4.11 DEFERRAL (OCI)
No deferral — OCI-S1..S3 are critical-path. `docker save/load` layout at OCI-spec level: `unverified — primary source not found` in this pass (R1 covers image format + Distribution API, not save-file layout); bundle consumes whatever the runtime exports.

---

## 5. Per-module Tags / Cutover — SemVer 2.0.0 + Go Modules Canonical Versions
### 5.1 Rule (SemVer 2.0.0, fetched)
`X.Y.Z`; MAJOR = incompatible API, MINOR = backward-compatible feature, PATCH = backward-compatible bug fix; `1.0.0` defines the public API (MUST declare a public API — SemVer §1); released version contents MUST NOT be modified (§3); pre-1.0 MAY change (§4 `0.y.z`); `vX.Y.Z` is a tag, not a version. A public API must be declared before assigning MINOR/MAJOR; without it, `1.3.5` "is stable" is `unverified — primary source not found`.
Go Modules Reference (fetched): module path is unique ID + import prefix; `go` mandatory minimum since 1.21; `require`/`replace`/`exclude`/`retract`; MVS. A module is "released, versioned, and distributed together". `v` + SemVer tag is the canonical version. Major v2+ needs `/vN` suffix (import-compatibility rule), forbidden for v0/v1.
### 5.2 Version state (today)
Canonical `KitVersion = "1.3.5"` (`internal/kit/kit.go:52`), `VERSION:1` = `1.3.5`, frontend `package.json:4` = `1.3.5`, `cmd/cgo/main.go:41` default overridden `-X main.version={{.Version}}`. `-X` release must equal tag (= KitVersion) or CLI (`kit next`) and GUI diverge (`main.go:36-40`). Docs stale: `docs/deploy.md:1,143` and `docs/api.md:17` show `1.2.3`. No `.goreleaser.yaml` derives `-ldflags -X` consistently.
Per R4 §1/§4: declare public API (CLI flags + exit codes + `/api/health|state|results|stream` + audit report JSON/MD) before 1.x split; `0.y.z` MAY change.
### 5.3 Go Modules canonical discipline
- Paths must be unique-controlled and chosen **before first publish** (else rename). Nested vs separate repo: human decision per module (breaks import path under R2 rule 8 + suffix rules). Paths in this spec are placeholders.
- `require` records a module path + version (`v1.2.3` or pseudo-version). Tags immutable once pushed; retraction needs a newer tag (`retract`, never silent rewrite). `go.mod` records module identity + dependency minima; `go list`/`go mod tidy` resolve via minimal version selection.
- `replace` applies only in the main module; may be used transiently during migration, must not survive a tag.
- `go` directive = minimum toolchain. Extraction must bump/pin it deliberately; `go 1.25` today.
### 5.4 Cutover procedure (every module, no exception)
```
1. Publish new module tag (v0.1.0, immutable)
2. Bump `require` in consumer go.mod (no `replace` in tag)
3. `go mod tidy`
4. Verify `go list -m all` (no internal `pkg/*` in new module; no `fyne` unless justified)
5. `goreleaser release --clean` dry-run (checksums + SBOM present)
6. Update every `keep_version` entry (mise.lock / package.json / Vite pin)
7. Run `make verify-fast` + `python tools/verify.py --selftest`
8. Verifier re-runs (builder never certifies) before `VERIFIED` row
```
A sibling-upgrade that rewrites frozen data, moves a tag, or leaves `replace` is a FAIL. `code` change + `require` + `tidy` + list + dry-run + lock + verify-fast + verifier is the *atomic* cutover; partial cutovers are not commits.
### 5.5 Per-module version mapping (examples)
| Change | SemVer | Example |
|---|---|---|
| Dashboard wording only (no surface change) | PATCH | typo fix in `AuditReportCard` |
| New CLI verb, new `/api` field additive, new `kit` action | MINOR | `cgo kit backup [dest]` |
| Rename exported symbol, remove field, break JSON | MAJOR + `/v2` | `Sample` field rename |
| Emitted `tc` line change, CSV schema change | MINOR at least (public API includes generated config/frozen files) | CAKE knob diff |
| `0.y.z` extraction (pre-1.0) | MAY break freely, but must say so | first `v0.1.0` |
### 5.6 No-new-dependency gate
Every extraction adds **no new external dep beyond stdlib unless approved**. Candidates needing approval: `cilium/ebpf` (T10), `prometheus/client_golang` (T16 if delegates), `golang.org/x/time/rate` (T19 if Go limiter), MQTT client lib (T32), QEMU/CHR (T27), YAML lib (T04/T22 if hand-parser grows). Default shippers: stdlib `net/http`, `crypto/subtle`, `os`, `strings`.
### 5.7 Anti-toy (versioning)
1. Mutable version identity (`:latest`, branch tags, hand-copied binaries, `cgo version` drift) fails.
2. No declared public API (README + godoc) fails SemVer §1 — MAJOR/MINOR unverifiable.
3. Tag moved or contents edited post-release fails immutability.
4. `replace` left in release tag fails.
5. Three-copy version (`KitVersion`/`VERSION`/`package.json`) disagreeing fails single-source rule.
### 5.8 DEFERRAL (versioning)
None — version single-sourcing is Slice 0 (S5 in cost §10-T03, S5 in CLI §10-T19, S3 in model §10-T01). Org-preferred module path (`cgo-cli` vs `/v2`): `unverified — primary source not found` (no org naming doc). Keep `v0.1.0` until deprecation cycles complete; do not claim `1.0.0` early.

---

## 6. cgo-gui Separate Repo — Fyne Cross-Compile Analysis + Build Constraints
### 6.1 What cgo-gui is (today)
Shared core, two views: `core.go:1-4` "Coeur applicatif PARTAGÉ (Windows + Linux) : aucune API Win32, aucun widget — que du pur Go (kit, vm, stdlib). Les deux vues pilotent les mêmes actions". Windows view is `//go:build windows` (`main.go:1`), raw Win32 via `golang.org/x/sys/windows` and LazyDLL `user32.dll`/`gdi32.dll` (`win32.go:3-41`), window + pump on one OS thread with `runtime.LockOSThread()` (`window.go:20-73` — "without it Go migrates goroutine and PostMessage freezes"). Window `winW=940/winH=700`, control IDs (`main.go:21-34` `idVMList=101…idNext=108`). Actions emit `cgo.exe kit …` beside GUI and stream output; interactive actions open visible console (`actions.go:217-270`).
Linux view is `//go:build linux` (`fygui_linux.go:1`), using `fyne.io/fyne/v2` + `app` + `container` + `widget` (`:19-22`); same buttons/journal/anti-double-click, only rendering changes (`:3-6`); refuses without `DISPLAY`/`WAYLAND_DISPLAY`, redirecting to `cgo kit tui` (`:731-735`). Shared button inventory groups `groupActions/access/deploy/control` (`core.go:75-103`); reverse `buttonAction(id)→(kind,args)` (`:247-313` kinds `direct:/bg:/console:`); `isKnownSuiteVerb` (`:112-123`). Windows dispatch `runKind` (`actions.go:381-518`), Linux `dispatchKind` (`fygui_linux.go:420-562`); shared `suiteStep/suiteContinue` anti-loop (`:520-589` + `:564-624`).
No Go `flag`/`cobra`/`urfave` surface. Entrypoint locates `cgo.exe` beside it (fatals if absent) (`main.go:79-107`); Linux locates sibling `cgo` (`fygui_linux.go:731-747`, `cgoExe` at `:643-645`). Log mirroring `cgo-gui-<date>.log` beside exe else `%TEMP%` (`core.go:23-66`).
### 6.2 The blocking failure (pre-existing, verified)
`docs/VERIFY.md:27`: "`GOOS=linux GOARCH=amd64 go build ./...` → exit 1, sole failing package `github.com/Realms4239/cgo/cmd/cgo-gui` (`fyne.io/fyne/v2/app` → `go-gl/gl` v2.1 'build constraints exclude all Go files'; `CGO_ENABLED=0`, no `x86_64-linux-gnu-gcc`); `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` → exit 0." Repeated VERIFY `:32` (Slice 6) and `:41` (Slice 10) with `pkgs=$(go list ./... | grep -v '/cmd/cgo-gui$')` → exit 0. `CHECKPOINTS.md:256-260` same. Header: "`cmd/cgo-gui` excluded (needs cgo/go-gl)" — gates deliberately exclude it, not that it builds.
**Root cause chain:**
1. `fygui_linux.go:1` is `//go:build linux` with **no exclusion tag**, unconditionally importing `fyne.io/fyne/v2` (`:19-22`). On a Windows host cross-compiling to Linux, Go disables CGO when it cannot find a suitable cross toolchain (`CGO_ENABLED=0`, no `x86_64-linux-gnu-gcc` per VERIFY).
2. Fyne's Linux GL backend (`go-gl/gl/v2.1`) provides zero Go files when CGO is disabled — hence "build constraints exclude all Go files". Fyne's exact apt/pkg-config prerequisites were not fetched — `unverified — primary source not found`.
3. Windows build never hits this because `main.go:1` is `//go:build windows` and imports no Fyne (`:10-19`). Linux file is a parallel `main()` (`:731-747`) chosen solely by GOOS. No intermediate `//go:build linux && cgo` guard or stub `main` exists for the CGO-free Linux case.
**Why it matters:** `go test/vet/build ./...` are the project's own gates (`Makefile:9-15`, `verify.py`). A red `./...` on Linux means cross-build cannot be `VERIFIED`; `VERIFY.md:27` records it as VERIFIED only by excluding the package (correct "recorded, not hidden", but gate stays red). Pre-existing (`35bb021` ancestor) — any new Linux CI running `./...` fails from HEAD backwards, masking new regressions. `.goreleaser.yaml:8-33` never builds `cgo-gui`; `Makefile:6-7` builds only `cmd/cgo`. GUI is undeployed via release automation on both OSes — invisible in shipped artifacts, visible in every `./...`.
### 6.3 Fyne cross-compile analysis (full)
- **Windows (native):** `x/sys v0.36.0` (`go.mod:9`) + raw Win32 + `LockOSThread`. No Fyne on Windows path (files are `//go:build windows`). Builds with `GOOS=windows GOARCH=amd64 go build ./cmd/cgo-gui` → exit 0 (acceptance S1 below).
- **Linux (CGO enabled, native toolchain):** `fyne v2.8.1` (`go.mod:6`) + transitive `go-gl/gl`, `glfw`, `image`, `oksvg`, `systray` (`go.mod:12-60`). Needs CGO + GL toolchain (GCC, X11/Wayland, OpenGL, pkg-config dev packages). Those system prerequisites are **not documented in-repo and Fyne docs were not fetched** — `unverified — primary source not found`.
- **Linux (CGO_ENABLED=0 cross from Windows):** `go-gl/gl` returns zero files → hard fail. This is the *only* failing configuration and it is the default `go build ./...` on a Windows dev host with `GOOS=linux`.
- **Linux (CGO enabled, no DISPLAY):** refuses with the `cgo kit tui` redirect (`fygui_linux.go:731-735`). Correct behaviour; the build-constraint failure above prevents this code from ever being reached in a cross build.
- **Behaviour drift risk:** `actions.go:381-518` vs `fygui_linux.go:420-562` (two dispatch loops), `window.go:177-237` vs `fygui_linux.go:648-729` (two layouts). Only `buttonAction`/`isKnownSuiteVerb` pure functions mitigate; two event loops (Win32 pump vs Fyne `pump()` `fygui_linux.go:78-117`) remain.
### 6.4 Minimal fix shape (no new dep, per Go Modules)
Isolate the CGO-dependent Linux view behind `//go:build linux && cgo` + a `//go:build !linux || !cgo` stub `main()` that prints the honest no-display redirect (`fygui_linux.go:732-735` logic) and exits 2; or extract `cmd/cgo-gui` into its own `go.mod` excluded from the monolith `./...`. The stub needs no `require` (stdlib only), satisfying Go Modules require rules (only imported paths need listing).
### 6.5 Separate-repo structure (target)
```
cgo-gui/  (own repo + own go.mod; monolith ./... goes green on CGO_ENABLED=0)
  go.mod                          # module .../cgo-gui, go 1.25
  core.go                         # shared pure Go (kit,vm,stdlib) — extracted verbatim
  win/                            # Windows view (//go:build windows): main.go, win32.go, window.go, actions.go, elevate.go
  linux/                          # Linux Fyne view (//go:build linux && cgo): fygui_linux.go
  stubs/                          # CGO-free stub main (//go:build !linux || !cgo): redirect + exit 2
  kit/  (or pkg/kit-client)       # thin client to cgo-kit (sibling binary), NEVER internal/kit import
  README.md                       # Fyne apt/pkg-config prerequisites (once fetched) + CGO requirement
  Makefile                        # go build ./... (CGO_ENABLED=1 for Linux view); GOOS=windows build for Win view
```
- **Linking rule:** `cgo-gui` imports **nothing** from the monolith's `internal/` (Go internal rule forbids it) and nothing from `github.com/Realms4239/cgo`. It talks to `cgo`/`cgo-kit` as subprocess (the existing `cgo.exe kit …` composition) or via a thin client. No `replace` directives, no `exclude`/`retract` at birth.
- **Build constraints:** `linux && cgo` for Fyne; `!linux || !cgo` for stub; `windows` for Win32. Every build-tag file carries its truth (`fygui_linux.go` line 1, `main.go` line 1).
- **Versioning:** own SemVer `v0.1.0` start (`0.y.z` MAY change); `v1.0.0` only when button→verb contract declared. `KitVersion` == tag == `cgo version` must agree (single-source, §5.4).
### 6.6 Per-tool slices (GUI)
- **S1 — Isolate `./...` green (fixes blocking failure).** Add build tags + stub `main` (or separate `go.mod`). Acceptance: `GOOS=linux GOARCH=amd64 go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./cmd/cgo-gui` → exit 0; `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go vet ./...` → exit 0.
- **S2 — Separate repo + thin client.** Move gui to `cgo-gui` repo; replace `internal/kit` import with subprocess/kit client. Acceptance: monolith CI no longer lists `cmd/cgo-gui` in `./...`; gui repo builds standalone with `GOPROXY=off`.
- **S3 — Fyne prerequisite docs (fetch first).** Fetch Fyne Linux prerequisites (GCC/X11/Wayland/OpenGL/pkg-config) or mark `unverified`. Acceptance: README documents prerequisites with primary citation; Linux CI image pinned.
- **S4 — Dual-view parity test.** Shared `buttonAction`/`isKnownSuiteVerb` golden across both views. Acceptance: 22 buttons map identically (Win vs Linux fixtures).
### 6.7 Anti-toy (GUI)
1. Green `./...` cross-platform (Linux + Windows + vet) — a red `./...` on Linux is a FAIL.
2. No fake hypervisor / no fake live data (stub must be labelled stub, excluded from "driver" claims).
3. Version identity single-sourced (KitVersion/tag/`cgo version` agree).
4. No `fyne` in monolith `go list -m all` after extraction — the GUI stack must never re-enter the core build list.
### 6.8 DEFERRAL (GUI)
- Linux Fyne CI image stays BLOCKED until S3's prerequisites are fetched/provisioned. Windows view ships first if needed; do not block the core on the Linux view.
- VMware `vmrun` semantics: `unverified — primary source not found` (Broadcom portal returned no manual on 2026-11-07). Treat `vm.go:64-125` as local evidence only.

---

## 7. WIS2 Probe — Full WMO WIS 2.0 + CAP 1.2 + Mosquitto Spec
### 7.1 The causal boundary (non-negotiable)
CGO measures links; a WIS2 Node is a separate dissemination system. WIS 2.0 is the *transport* (MQTT publish-subscribe over the public Internet) [R1]; CAP 1.2 is the *payload vocabulary* for all-hazard alerts [R2]; CGO's job is to measure whether a constrained link can carry small alert/telemetry objects on time [R4, R5]. Confusing measurement with dissemination is the overclaim this module must not make. Never present 220 ms as a WMO mandate; never present CGO as a WIS2 Node.
### 7.2 What is missing (verified)
1. **No MQTT plane at all.** No broker, no client, no topic, no MQTT dependency in the product [R9, R10]. Only live distribution is HTTP SSE inside one host (`/api/stream` 10 Hz, 64-sub cap) — a dashboard feed, not dissemination, does not leave the bench host.
2. **No WIS2 Node.** A Node is *two* public endpoints — MQTT broker publishing WIS2 notifications *and* HTTP server for data download — plus WCMP2 metadata records and per-granule notifications [R1]. None exist in CGO [R10].
3. **No CAP producer/consumer.** CGO computes a CAP *SLA* (Model C, 16 KB completions vs 220 ms/95%) [R5] but produces no CAP XML, validates none against R2 schema, implements neither §4.3 producer nor §4.4 consumer conformance [R2, R10]. "CAP" in CGO today means "deadline objective", never "CAP message" [R5, R6].
4. **Already named as deferred, not forgotten.** Hardening spec lists *"WIS 2.0 / CAP publication; Mosquitto multi-site"* under DEFERRED with unblock condition *"primary docs fetched"* [R8]. This section is the first half of that unblock condition (R1–R3 fetched); the second half — a human decision that dissemination is in scope at all — is not taken here.
### 7.3 Primary sources (fetched)
| # | Topic | What it says |
|---|---|---|
| R1 | WMO WIS 2.0 overview (community.wmo.int) | Operational 01/01/2025; unlike GTS private links, WIS 2.0 uses public Internet + MQTT pub-sub; each publisher must operate a WIS2 Node; Global Broker / Global Cache / Global Discovery Catalogue / Global Monitor; topic hierarchy `origin/a/wis2/<centre-id>/data/core|recommended`, `cache/a/wis2/<centre-id>/data/core`, `origin/a/wis2/<centre-id>/metadata`; WNM with canonical HTTP(S) download link; WCMP2 discovery metadata; `wis2box` = FOSS reference Node; registration via PR/NFP + centre-id + GISC assessment, only for an *operational* Node with public MQTT regularly publishing. |
| R2 | CAP 1.2 (OASIS Standard, 01 July 2010) | §1.1 open, non-proprietary format for all-hazard alerts, no particular application/telecom method; §1.3 `<alert>` → `<info>` (urgency/severity/certainty triples) → `<resource>`/`<area>`; §3.2.1 namespace `urn:oasis:names:tc:emergency:cap:1.2`, required `<alert>`: `identifier`, `sender`, `sent`, `status`, `msgType`, `scope`; §3.2.2 required `<info>`: `category`, `event`, `urgency`, `severity`, `certainty`; §4.3/§4.4 conformance as Message Producer/Consumer. |
| R3 | Eclipse Mosquitto `mosquitto.conf(5)` man | Config file for the broker; `listener` + per-listener TLS (`cafile`/`certfile`/`keyfile`, `require_certificate`, `use_identity_as_username`/`use_subject_as_username`, `tls_version`); `password_file`, `acl_file` (`topic [read|write|readwrite|deny] <topic>`, `user <username>`, `pattern` with `%c`/`%u`), `allow_anonymous` (defaults false unless no listener defined), `max_qos`, `retain_available`, `persistence`, `sys_interval` (`$SYS` hierarchy), bridge options (`topic <pattern> [[in|out|both] qos] [local-prefix remote-prefix]`, bridge TLS `bridge_cafile`/`bridge_certfile`/`bridge_keyfile`, bridge notification topic `$SYS/broker/connection/<id>/state`). |
| R4 | Repo identity | CGO = Congestion Ground-truth Observatory: reproducible bench for comparing AQM/BBR on constrained access links (`specs/2026-10-25:10-21`). |
| R5 | DGM scenario + Model C | Download audit = bulk FROM sink through shaped `e-wan` = *"the actual DGM scenario (200 MB EUMETSAT downlink vs CAP egress)"*; Model C: loaded 16 KB at 220 ms, breach `< 95%`; `cost.go:75-79` states 220 ms *"is NOT a threshold imposed by the WMO"*. |
| R6 | Ledger status of CAP objective | `VERIFY.md:48` `RECORDED` contractual scenarios, *"Not re-run as a measurement here"*. |
| R7 | Case-study posture | Read-only only; never capture/probe/load/write/root/login/scan/query; BLOCKED on written grant (a)–(d); nothing read/sent/executed. |
| R8 | Hardening deferral | §6.3 DEFERRED: *"WIS 2.0 / CAP publication; Mosquitto multi-site"* — unblock = primary docs fetched, image procured, human grant signed. |
| R9 | Current live surface (no MQTT) | Live data is SSE over HTTP: `/api/stream` 10 Hz (ring 2048, `503` > 64), `/api/events` last 50; `go.mod:5-10` fyne/bubbletea/lipgloss/x-sys only — no MQTT client. |
| R10 | Absence of MQTT/WIS/CAP code | grep `mqtt\|mosquitto\|wis2\|paho` in `pkg/` + `cmd/` → **no matches** (2026-11-07); repo-wide hits only docs/specs/plans, never `pkg/`/`cmd/`. No broker/client/topic/CAP XML exists. |
| R11 | WNM/WCMP2/full hierarchy schemas | Linked from R1 as WMO-No. 1060 Vol. II App. D/E + WMO-No. 1061; not fetched. Exact WNM fields, WCMP2 fields, full annex beyond `origin/.../core|recommended` shapes: `unverified — primary source not found`. |
| R12 | `warnings.wmo.int` portal / mandated QoS/retain/CAP-over-WIS2-topic mapping | Not fetched; R1 states no QoS value, retain policy, or CAP-on-data-topic mapping. Whether CAP travels as WIS2 data granules, and at what QoS: `unverified — primary source not found`. |
| R13 | Madagascar/DGM centre-id, GISC, registration state | Not established. Any claim DGM holds centre-id / registered Node / Global Broker subscription: `unverified — primary source not found`. |
| R14 | `wis2box` deployment specifics | R1 names it only as *"[WIS2 in a box] (wis2box): FOSS reference implementation of a WIS2 Node"* with docs link. Install/config/payloads/CAP support: `unverified — primary source not found`. |
### 7.4 Responsibility boundary (what the probe owns)
| Responsibility | Owner | Primary |
|---|---|---|
| Measure whether constrained link delivers 16 KB within 220 ms under load (Model C) | **CGO (existing)** — unchanged | R5 (`cost.go:75-112`) |
| Carry small objects/notifications over public Internet between sites | **MQTT broker layer (missing)** — Mosquitto-class, TLS + ACL per R3 | R1 (WIS 2.0 = MQTT pub-sub over public Internet), R3 |
| Publish data + metadata notifications and serve downloads | **WIS2 Node (missing, separate system)** — broker + HTTP + WCMP2 + registration | R1 |
| Express all-hazard alert in standard message format | **CAP 1.2 producer (missing)** — only if operator wants CAP alerts | R2 |
| Decide DGM centre-id / registration / GISC path | **Human / DGM / WMO Secretariat (out of scope for any code)** | R1, R13=`unverified` |
**Non-responsibility:** the probe must not turn CGO into a Node by default. Measurement evidence (sealed `audit_report.json`, `telemetry.csv`) stays in CGO's evidence model [R4]; the probe *reads* a sealed report as input and publishes *derived* notifications. It does not re-grade, re-measure, or re-seal anything.
**Why separate, not a CGO feature:** (1) single Go binary with embedded React and no DB, frozen CSV schemas, ledger discipline (`VERIFIED` requires named command) [R4, R6] — adding public-Internet broker + TLS + ACL + registration widens the trust boundary to exactly the plane the hardening spec deferred until authN/Z lands [R8]; (2) case-study contract forbids CGO from connecting to operator infrastructure — *"CGO itself connects to nothing"* [R7] — a probe that *does* connect is a different trust class and must live in a different repo with its own grant boundary; (3) a missing tool that is also *separate* can be refused, revoked, or left uninstalled without touching the measurement bench. Reversibility is the point.
### 7.5 Exported API — PROPOSED (not a standard)
Everything in this subsection is a **proposal** for the future independent tool. Not a WMO requirement, not an OASIS claim, not a CGO commitment. Standard shapes marked [R1]/[R2]/[R3]; inventions marked **PROPOSED**.
#### 7.5.1 Topic tree
Standard (from R1):
```
origin/a/wis2/<centre-id>/data/core
origin/a/wis2/<centre-id>/data/recommended
origin/a/wis2/<centre-id>/metadata
cache/a/wis2/<centre-id>/data/core
```
`origin/a/wis2/<centre-id>/data/core` implies free unrestricted access (Resolution 1); Global Cache downloads it and re-announces on `cache/a/wis2/<centre-id>/data/core` [R1]. `.../data/recommended` may require access control and is *not* downloaded by Global Caches [R1]. `.../metadata` carries WCMP2 canonical link; Global Discovery Catalogue updated from it [R1].
**PROPOSED lab namespace (default until centre-id exists):**
```
cgo/v1/probe/<site-id>/link-quality        # PROPOSED — lab only, never on a WIS2 Node
cgo/v1/probe/<site-id>/cap-sla              # PROPOSED — lab only
cgo/v1/probe/<site-id>/status/$SYS-like     # PROPOSED — use broker $SYS instead where possible [R3]
```
Rationale: `origin/a/wis2/...` requires centre-id proposed by Member and endorsed by WMO Secretariat [R1]. Publishing under `origin/a/wis2/<invented-id>/...` before assignment would be a false WIS2 claim. `cgo/v1/...` keeps the lab tool honest and is what Mosquitto ACLs can scope (`acl_file` topic rules, `pattern` with `%u`/`%c`) [R3].
**PROPOSED WIS2 mapping (only after centre-id assignment + GISC assessment [R1]):**
| Probe output | WIS2 topic | Class |
|---|---|---|
| Sealed audit summary (derived JSON, §7.5.2) | `origin/a/wis2/<centre-id>/data/recommended/cgo/link-quality` | PROPOSED sub-path under standard `.../data/recommended` prefix [R1]; `recommended` because access control applies and Global Cache must not ingest it [R1] |
| WCMP2 record for the above dataset | `origin/a/wis2/<centre-id>/metadata` | standard [R1]; record schema = `unverified` [R11] |
| CAP XML (§7.5.3) | HTTP download URL referenced from WNM canonical link, *not* MQTT payload | PROPOSED separation; whether CAP travels as WIS2 data granule = `unverified` [R12] |
#### 7.5.2 Notification payload (PROPOSED lab JSON)
Standard (from R1): WNM contains canonical link indicating HTTP(S) endpoint for download [R1]. Exact WNM fields = `unverified` [R11].
```json
{
  "schema": "cgo.probe.notification/v1",
  "site_id": "PROPOSED — operator-assigned, never a hostname",
  "audit_id": "<id from sealed audit_report.json>",
  "seal_sha256": "<SHA-256 of telemetry.csv — copied, never recomputed differently>",
  "measured_at": "2026-11-07T00:00:00Z",
  "link_quality": {
    "cap_deadline_ms": 220,
    "cap_ok_pct": 96.5,
    "cap_breach": false,
    "basis": "loaded 16 KiB completions"
  },
  "download_url": "https://<probe-http-host>/<audit_id>/audit_report.json",
  "cap_url": "https://<probe-http-host>/<audit_id>/alert.xml"
}
```
`cap_deadline_ms`/`cap_breach` semantics copied from `AssessCAPSLA` (`Measured=false` ⇒ no value, never a pass) [R5]. 220/95 stay labelled *"objectif DGM, cadre WIS 2.0"* — never WMO mandate [R5]. `seal_sha256` is a *copy* of the CGO seal [R4]; probe never re-seals. Report without seal is refused (fail-closed), mirroring *"unmeasured ⇒ `Unavailable`, never zero, never pass"* [R4].
#### 7.5.3 CAP mapping (PROPOSED, only when operator asks for CAP alerts)
Standard (from R2): CAP 1.2 XML, namespace `urn:oasis:names:tc:emergency:cap:1.2` [§3.2.1]. Required `<alert>`: `identifier`, `sender`, `sent`, `status`, `msgType`, `scope` [§3.2.1]. Required `<info>`: `category`, `event`, `urgency`, `severity`, `certainty` [§3.2.2]. `status` ∈ {Actual, Exercise, System, Test, Draft}; `msgType` ∈ {Alert, Update, Cancel, Ack, Error}; `scope` ∈ {Public, Restricted, Private} [§3.2.1]. `sender` globally unique (e.g. domain-based) [§3.2.1]. Producer §4.3, consumer §4.4 [R2].
| CAP element [R2] | PROPOSED value | Source/rule |
|---|---|---|
| `<identifier>` | `<audit_id>-<sha256-prefix>` (unique per message, no spaces/commas) | PROPOSED; uniqueness [§3.2.1] |
| `<sender>` | operator domain-based id, e.g. `cgo-probe/<site-id>.dgm.example` | PROPOSED; globally-unique [§3.2.1] |
| `<sent>` | audit `measured_at` in DateTime format, UTC as `-00:00` | rule [§3.2.1]; format [§3.2.2 `effective`] |
| `<status>` | `Actual` | PROPOSED; domain [§3.2.1] |
| `<msgType>` | `Alert` on first breach, `Update`+`<references>` on recovery, `Cancel` to retract | PROPOSED sequencing; domain + `<references>` form `sender,identifier,sent` [§3.2.1] |
| `<scope>` | `Private` (named operator addresses) or `Restricted` | PROPOSED; domain [§3.2.1] — never `Public` without operator sign-off |
| `<info><category>` | `Infra` (telecom infrastructure) | PROPOSED from R2 domain [§3.2.2] |
| `<info><event>` | `Link quality below operator objective` | PROPOSED text; element required [§3.2.2] |
| `<info><urgency>/<severity>/<certainty>` | derived from `cap_ok_pct` bands — **bands PROPOSED and must be operator-signed** | domains [§3.2.2]; no band table in R2 — inventing one silently = toy |
| `<info><parameter>` | `valueName="CGO_CAP_OK_PCT"`, `value="<cap_ok_pct>"`; `valueName="CGO_SEAL_SHA256"`, `value="<seal>"` | PROPOSED use of optional `<parameter>` [§3.2.2] |
| `<resource>` | sealed `audit_report.json` / `telemetry.csv` via `<uri>` | PROPOSED; `<resource>` = optional digital-asset reference [§1.3.3] |
**Hard boundary:** healthy probe (no breach) publishes **no CAP message**. CAP `msgType=Alert` carries *"initial information requiring attention"* [§3.2.1]; routine telemetry is not an alert. Steady-state channel is JSON notification (§7.5.2), not CAP.
#### 7.5.4 Broker-facing contract (Mosquitto, R3)
Standard capabilities proposal relies on:
- Separate listeners per trust class: `listener <port> [bind]` [R3]; TLS listener with `cafile`/`certfile`/`keyfile` [R3].
- Client auth: `require_certificate true` + `use_identity_as_username true` makes certificate CN the ACL username, password unused [R3]; or `password_file` + `allow_anonymous false` [R3].
- Per-user topic control: `acl_file` with `topic [read|write|readwrite|deny] <topic>` and `user <username>` lines; `pattern` substitution `%c` (client id) / `%u` (username) [R3].
- QoS cap per listener: `max_qos` [R3]; retained-message support: `retain_available` [R3]; offline durability: `persistence` + `persistence_location` [R3]; broker self-reporting: `sys_interval` publishes `$SYS` hierarchy [R3].
- Multi-broker fan-out: `connection` bridges with `topic <pattern> [[in|out|both] qos-level] [local-prefix remote-prefix]`, bridge TLS (`bridge_cafile`, `bridge_certfile`, `bridge_keyfile`), and `$SYS/broker/connection/<id>/state` notifications [R3].
**PROPOSED profile (lab):**
```ini
# PROPOSED mosquitto.conf fragment — lab only, not a WIS2 Node config
listener 8883
cafile /etc/cgo-wis2/ca.crt
certfile /etc/cgo-wis2/broker.crt
keyfile /etc/cgo-wis2/broker.key
require_certificate true
use_identity_as_username true
allow_anonymous false
max_qos 1
retain_available true
persistence true
persistence_location /var/lib/cgo-wis2/
sys_interval 10
acl_file /etc/cgo-wis2/acl
```
`max_qos 1` and retain policy are **PROPOSED**: R1 states no QoS value and R3 only defines the `max_qos` knob [R3, R12]. QoS 1 proposed because QoS 0 is fire-and-forget and notification must survive constrained link — engineering choice, not WMO mandate, recorded as such. Bind to specific interface/address (`listener <port> <bind>` [R3]); public-Internet listener is a *separate, later* decision gated on registration [R1], never lab default.
### 7.6 Dependencies (WIS2)
| Dependency | Kind | Status/gate | Primary |
|---|---|---|---|
| Eclipse Mosquitto broker (external binary) | Runtime, not Go module | New external dependency — human approval before scaffolding (stop-and-ask) | R3 |
| MQTT client library (Go) | New Go module dep | **None exists today** [R9 `go.mod:5-10`]; adding one is stop-and-ask; candidate choice `unverified` (no lib evaluated) | R9, R10 |
| CAP XML generation | `encoding/xml` (stdlib) | No new dep proposed; schema validation vs R2 §3.4 is the work | R2 §3.4 |
| `wis2box` (FOSS reference Node) | Alternative runtime | **Not proposed as dependency of probe.** Deploying it is a *separate* decision by Node operator. Install/config specifics `unverified` [R14] | R1, R14 |
| CGO sealed report (`audit_report.json` + `telemetry.csv`) | Input artifact, read-only file | Probe consumes CGO output; does not import CGO packages, re-grade, or write into `book/data/runs` (frozen, AGENTS.md) | R4, R7 |
| Centre-id, GISC, Global Broker subscription | Organisational | External to code; required before any `origin/a/wis2/...` publishing [R1]; Madagascar state `unverified` [R13] | R1, R13 |
| Operator written grant | Human | Required before probe reads *anything* operator-side, per grant items (a)–(d) [R7] | R7 |
**Explicitly not a dependency:** frozen corpus, CSV schemas, six-gate verification system [R4, R6]. Probe adds no gate to CGO and must not weaken one.
### 7.7 Primary-source requirements (what must be fetched before building)
Each item is a *requirement on the builder*, not a claim. Status = this pass.
| # | Requirement | Why | Status |
|---|---|---|---|
| Q1 | Fetch WMO-No. 1060 Vol. II App. D (Topic Hierarchy) + App. E (WNM, WCMP2) | R1 links them but does not reproduce schemas; §7.5.2/§7.5.3 cannot be implemented from R1 alone | `unverified` [R11] |
| Q2 | Fetch WMO-No. 1061 Guide §2.6.1.1 (Node registration/decommissioning) | R1 summarises registration; normative steps/assessment in Guide | `unverified` [R11] |
| Q3 | Establish whether CAP documents are carried as WIS2 data granules, and at what QoS/retain | §7.5.3 separation (CAP via HTTP canonical link, not MQTT payload) is proposal until answered | `unverified` [R12] |
| Q4 | Obtain operator centre-id status or written decision not to register | Determines whether §7.5.1 WIS2 mapping is ever used or lab namespace permanent | `unverified` [R13] |
| Q5 | Operator-signed CAP severity bands (`cap_ok_pct` → urgency/severity/certainty) | R2 defines domains but no mapping from link metric to them [§3.2.2]; unsigned bands = invented thresholds | Not a fetch — a signature |
| Q6 | Evaluate `wis2box` docs (install, accepted payloads, CAP support) | Decides probe→`wis2box` vs probe→own broker; R1 only names it [R1] | `unverified` [R14] |
| Q7 | Select MQTT client library with primary source (release notes/API docs) | No library in `go.mod` today [R9]; choice must be citable | Not started |
**Build rule:** no slice in §7.8 starts until its Q-items close. In particular, S4 (WIS2 publishing) is blocked on Q1–Q4 exactly as hardening spec's DEFERRED requires [R8].
### 7.8 Gap analysis — why CGO is NOT a WIS2 Node
| WIS2 Node requirement [R1] | CGO today | Gap |
|---|---|---|
| MQTT broker over public Internet, publishing WIS2 notifications [R1] | No MQTT code at all [R10]; HTTP SSE on one host [R9] | **Absent.** Whole plane |
| HTTP server enabling download of data and metadata [R1] | Dashboard on `10.99.0.2:9090` serves UI/API, not data granules for WIS2 network [R4] | **Absent as Node function** |
| WCMP2 records; WNM on `origin/a/wis2/<centre-id>/metadata` with canonical link [R1] | No WCMP2, no WNM [R10]; schema unfetched [R11] | **Absent** |
| Topic hierarchy `origin/a/wis2/<centre-id>/data/core|recommended` [R1] | No topics; centre-id never assigned [R10, R13] | **Absent** |
| Registration: PR/NFP, centre-id endorsement, WIS2 Register, GISC assessment, Global Broker — only for *operational* Node with public MQTT *regularly publishing* [R1] | No registration/assessment/subscription; R1 Note 2: only considered for already-operational endpoint | **Absent, and order matters:** broker first, paperwork after |
| 24/7 operation with Global Monitor [R1] | Bench + audit tooling; no 24/7 dissemination duty | **Different duty.** CGO availability = "run audit, seal evidence" [R4] |
**Causal statement:** CGO's chain (raw probes → freeze → manifest → `hash8`) [R4] proves *what the link did*. A WIS2 Node's chain (data granule → WNM with canonical link → Global Cache/Catalogue) [R1] proves *that the network can fetch the data*. Different claims, operators, failure modes. A tool that measures the link cannot, by measuring, become entitled to publish for the network — and a Node that publishes cannot, by publishing, prove the link was good. The probe sits *between* them: reads sealed evidence, emits notifications, claims neither role.
### 7.9 Target tool — `cgo-wis2-probe` (independent repo)
**Name:** `cgo-wis2-probe` — one independent repository, name fixed by human at creation; not created by this research. **Not** a package, directory, or branch inside `operator-cockpit`.
**Boundary (from R7):** probe is first CGO-adjacent component allowed to *connect* — and only to a broker the human configures, only after operator grant [R7]. Everything else (frozen corpus, CSV schemas, CGO gates) stays untouched [R4, AGENTS.md].
**Inputs (read-only):** `data/audits/<id>/audit_report.json` + `telemetry.csv` + seals [R4]. **Outputs:** MQTT notifications (§7.5.2) and, on operator demand, CAP XML (§7.5.3). **Never:** re-grading, re-sealing, writing into CGO, or publishing under `origin/a/wis2/...` before Q4 closes [R13].
#### Slices (dependency order; each gated on §7.7)
| Slice | Scope | Acceptance (all required) |
|---|---|---|
| **S0** | Close Q1–Q4 (fetch WNM/WCMP2/hierarchy/Guide; establish centre-id status; decide CAP-carriage) | Sources table updated with fetched URLs; any item still unfetched stays `unverified` and blocks S4 |
| **S1** | Lab Mosquitto: TLS listener, `require_certificate` + `use_identity_as_username`, `acl_file`, `max_qos`, `persistence`, `$SYS` (`sys_interval`) [R3] | Anonymous refused; wrong-CN cert refused; `mosquitto_sub` with valid cert receives only its ACL topics; broker survives restart with retained messages (persistence proof) |
| **S2** | Probe publisher, lab namespace only (`cgo/v1/probe/<site-id>/...`, §7.5.1) | Sealed report → notification round-trip on lab broker; unsealed refused; `Measured=false` publishes *nothing* (never pass by default [R5]) |
| **S3** | CAP producer (§4.3 conformance [R2]) | XML validates vs R2 §3.4 schema; `<sender>` globally unique; `Update`/`Cancel` carry correct `<references>` [§3.2.1]; healthy probe emits no CAP |
| **S4** | WIS2 publishing (only if Q4 grants centre-id and Q1–Q3 close) | WNM on `origin/a/wis2/<centre-id>/metadata` with canonical link [R1]; data on `.../data/recommended` [R1]; GISC path documented — until then **does not run** |
| **S5** | Bridge / multi-site (Mosquitto `connection` + `topic ... in/out/both`, bridge TLS [R3]) | Two lab brokers exchange notifications with prefixes intact; bridge state on `$SYS/broker/connection/<id>/state` [R3]; loop/duplicate behaviour documented |
| **S6** | Evidence + ledger (probe's own `VERIFY`-style rows) | Every claim row names command + host + date + saved output [R6]; fault proofs in §7.10 pass; `BLOCKED` row names condition and stays blocked |
**Sequencing constraint (from R8):** S1 must not precede CGO's own authN/Z decision — hardening defers Mosquitto precisely because a second unauthenticated plane would compound exposure. Probe's broker is a *new* plane; the human sequences it.
### 7.10 Anti-toy criteria (WIS2)
A toy here = demo that *looks* like WIS2/CAP dissemination but is a mock broker, hardcoded payload, or screenshot. Each falsifiable; failure is `REFUTED`, not quiet fix [R6].
1. **Real broker, real TLS.** S1 passes against real Mosquitto binary with certificate auth [R3]. Mock/in-process fake broker in acceptance test = toy, refused. Fault proof: replace valid client cert with expired one → refused.
2. **ACLs enforced, not decorated.** Valid cert but no ACL for `cgo/v1/probe/<other-site>/...` must be denied publish/subscribe [R3 `acl_file`]. Mutate ACL to `deny` → same publish must fail. (Pattern `%u`/`%c` [R3] own test: site A cannot write site B's prefix.)
3. **No green without measured input.** Report with `Measured=false` / no loaded samples → publishes nothing and logs refusal [R5]. Fabricated `cap_ok_pct` with no seal → refused. This is CGO's invariant applied to probe [R4].
4. **CAP schema-valid or it does not ship.** Every emitted CAP XML validates vs R2 §3.4. Fault proof: hand-edit one `<urgency>` to out-of-domain → validation must fail and message must not publish. §4.3 producer conformance or nothing [R2].
5. **220 ms framing survives trip.** Any human-facing string says *"objectif DGM, cadre WIS 2.0"* and never presents 220 ms as WMO threshold [R5]. Automated check: grep output corpus for `WMO mandate` / `seuil OMM` → must be empty.
6. **No false WIS2 identity.** Until Q4 closes, probe must never publish to `origin/a/wis2/...` [R1]; test asserts lab config contains no `origin/a/wis2` topic. After Q4, only assigned centre-id appears, GISC record on file [R1].
7. **Persistence proven by killing broker.** Stop Mosquitto mid-queue, restart, assert retained/queued notification delivered [R3 `persistence`]. Clean shutdown does not count.
8. **Bridge loops tested, not assumed.** Two brokers bridged both ways (`topic ... both` [R3]) must not amplify; duplicate-delivery measured and recorded, not waved away.
9. **Evidence discipline matches CGO's.** Same ledger standard: `VERIFIED` only on named command + host + date + saved output; `BLOCKED` names condition [R6]. Claim without row = hypothesis, labelled as such.
10. **Case-study grant gates everything operator-side.** No probe run reads operator config, names real service/host, or publishes on behalf of DGM without grant (a)–(d) [R7]. Default identity pseudonym, exactly as contract requires [R7].
### 7.11 DEFERRAL (WIS2)
- **Entire tool DEFERRED** until Q1–Q7 close + human dissemination-scope decision + operator grant. S0 fetches Q1–Q4; S4 needs Q1–Q4 + centre-id assignment + GISC assessment. No `origin/a/wis2/...` publication before then (lab namespace only).
- Mosquitto multi-site deferred per R8 until CGO authN/Z lands (exposure compounding).
- Consumer/CAP publication deferred until producer validated (producer §4.3 before consumer §4.4).
- `wis2box` as Node implementation deferred to Node operator (not probe dependency).

---

## 8. eBPF Observer — Full Kernel eBPF/BTF/CO-RE/cilium-ebpf/libbpf Analysis
### 8.1 Why eBPF matters (Master/IP bottleneck)
Frozen corpus 159/528/208/74/246/22 (`VERIFY.md:24`) never re-measured (umbrella invariant 3). Every *new* measurement must be defensible in front of a Master jury that knows the proxy flaw:
- Weak link is the retransmit numerator. `pkg/probe/tcpinfo_linux.go:15` hardcodes `retransMSS=1448` and `:43` returns `Total_retrans*1448`. Header: "TCP_INFO rend des SEGMENTS réémis ; ×1448 les convertit en octets comme le fait le mémoire (wasted = drops × 1448). Proxy documenté, pas une mesure par paquet." Ground-truth spec §0 expands: proxy "is wrong under offload, fragmentation, or a non-1448 MSS" (segment count × guessed MSS, not byte count per packet). No jury accepts Model A numerator as guessed MSS under GSO/TSO/GRO (kernel keeps super-packets up to 64 KiB TSO / ~25 KiB GRO — R1 §5.2.4 notes TSO/GRO "mega-packets severely impact … scheduling").
- Model A/B economics inherit the flaw. `CapacityWasteARPerHour = (wasted/transported) × (MonthlyOpEx/720)` and `MeteredWasteARPerHour = (wasted_bytes/2^30) × tariff × (3600/window)` where `wasted` derives from decision #9. If ratio is proxy, Ar/h is proxy. Ground-truth spec §§4.2–4.3 supersede decision #9 "for hosts where eBPF is available" while preserving honesty (`measured` TCP_INFO never confused with `estimated` proxy), and add `source: ebpf|tcpinfo|""` + `basis: measured|estimated|""` and `kernel_telemetry` report block (§4.5). Economics bullet survives only if source recorded per window.
- Bufferbloat/AQM claims need drop truth per class. RFC 8290 defines FQ-CoDel as "combined packet scheduler and AQM algorithm" (R1 §1) with per-queue CoDel (R1 §2 → RFC 8289), flow hashing 5-tuple + salted Jenkins (R1 §4.1), DRR new/old two-tier (R1 §4.2), defaults interval 100 ms (R1 §5.2.1), target 5 ms (R1 §5.2.2), quantum 1514 (R1 §5.2.4), flows 1024 (R1 §5.2.5), only MUST "some packet limit MUST be enforced" (R1 §5.2.3). Defines *none* of CGO probes/phases/schema/220/95. Legitimate use is framing why loaded latency is measured. CAKE has no RFC (R1 §10.2 informative, §7 future work). Thesis citing RFC 8290 as CAKE authority is citation theatre; thesis measuring EF vs BE under CAKE/fq_codel with kernel drop counters is experiment. eBPF gives that drop leg. `.dist`/shape (dwell) cannot be true without per-packet sojourn — eBPF is *precondition* for `.dist` even though `.dist` is blocked (VERIFY `:34`).
### 8.2 What "ground truth" concretely means (and does not)
| Signal | Hook (C source) | Unit claimed | Not claimed |
|---|---|---|---|
| Retransmit bytes | `raw_tp/tcp_retransmit_skb` — `bpf_probe.c:111-147`. Comment `:6-9`: "`skb->len` est déjà le payload TCP, TSO compris, et `transport_header` porte encore le sentinelle `0xFFFF`. Ne rien soustraire." TP_PROTO gives `(sk, skb)` so raw TP reads final `skb->len`. Spec §2.1: `tcp_transmit_skb(clone_it=1)` pushes headers on clone; queued skb keeps payload, so `payload = skb->len`. Kernel itself does `tp->bytes_retrans += skb->len` (spec §2.3). | Submitted retransmit payload bytes, successful transmits only, after fragmentation/collapse (TSO-inclusive). | ACKed/delivered bytes; failed attempts (no tracepoint fire); wire bytes incl. headers (not observable at this hook, spec §2.3 — stored as no counter, not estimated). |
| Socket state | `fentry/tcp_rcv_established` primary + `kprobe/tcp_rcv_established` fallback — `:175-185`. `sample_sock` (`:149-173`) reads `srtt_us >> 3` (kernel left-shifted by 3), `snd_cwnd`, `snd_ssthresh`, `skc_state`, throttled 50 ms (`SAMPLE_NS 50000000ULL`, `:27`). | Sampled `srtt_us` (unshifted), `snd_cwnd`/`snd_ssthresh` in segments, state, ≤20 Hz/socket. | Per-packet RTT. Sample is "monitoring, never per-packet" (spec §2.4). |
| Qdisc drops | `raw_tp/kfree_skb` — `:187-251`. Filters `reason == SKB_DROP_REASON_QDISC_DROP` via `bpf_core_enum_value` (`:195-197`), parses IPv4/TCP headers (`:211-242`), attrs per-flow only if flow entry already exists, else global only (`:244-249`). Linux 6.8 has **no** `qdisc_drop` tracepoint (introduced v7.1, `ff2998f29f39` — spec §2.1); 6.8's `__qdisc_drop()` in `include/net/sch_generic.h` defers to `kfree_skb_list_reason()` → `__kfree_skb_reason()` with that reason, so `kfree_skb` is correct 6.8 hook. On 7.1+ native `qdisc_drop` can replace it (spec §§2.1, 7.12). | Host-wide qdisc-drop packets/bytes (`global_stat` `:52-58`), per-flow drop packets/bytes when 4-tuple parses and flow exists. | Queue dwell / sojourn / queue delay — "no enqueue/dequeue timestamp programs and no skb-pointer correlation map" (spec §2.6). Drop timing is free event, possibly batched, not enqueue timestamp (spec §7.6). |
Tracepoint semantics support split: off = branch+space only; on = synchronous callback in caller context (R5 §Purpose). "Zero userspace overhead" means *no per-drop syscall* (counters updated in-kernel, read at window end) — not zero cost; the 2% overhead gate (§6.5) bounds it.
Maps: `flows` = `BPF_MAP_TYPE_HASH` max 4096 (`:60-65`), `globals` = `BPF_MAP_TYPE_PERCPU_ARRAY` max 1 (`:67-72`). IPv6/non-IPv4 retransmits and map-full overflows increment `global_stat` so totals never silently drop (`:121-126`, `133-141`). Per-flow key omits netns — same-tuple collision across namespaces merges counters. Spec §3.2 names this documented bench-only `shortcut` to upgrade to `(netns, tuple)` if background traffic can collide.
### 8.3 BLOCKED rows make this the bottleneck
| VERIFY row | Status | Gates |
|---|---|---|
| "eBPF retransmit / RTT / qdisc drops" (`:29`) | **BLOCKED** — No `bpf_probe.o`, no Linux 6.8 run; Windows cannot build; CO-RE source + `make bpf` landed 2026-10-29, object not built: CGO-DEV VM SSH rejects every key/user, PC2 WSL 6.18 without clang/bpftool/tc. | Every `source=ebpf` sentence. |
| "eBPF CO-RE source + build target" (`:30`) | **RECORDED** — `bpf_probe.c`, `BUILD.md`, `vmlinux.h` gitignore, `make bpf/bpf-check` committed 2026-10-29; `make` absent on Windows; "No object, no hash, no loader yet; do not commit object until `make bpf/bpf-check` run on Linux 6.8." | Provenance chain. |
| "Linux live acceptance (G0–G7)" (`:28`) + "IFB" (`:31`) + "PC2/nemo" (`:56`) | **BLOCKED** (6.8 bench unreachable; CGO-DEV SSH reject; PC2 WSL 6.18.54.1 lacks `tc`/BBR/clang/bpftool) | Bench truth for *all* live cells. |
| "BBRv3 3×3 matrix" (`:33`) + engine (`:52`) | **BLOCKED** — No 7.1 kernel; VM exposes only reno, cubic, bbr. | Third CC row. |
| "`.dist` netem" (`:34`) + engine (`:53`) | **BLOCKED** — No field trace ≥30 baseline samples; two SMOKE-TEST 19+25 rows. | Trace-driven calibrator. |
| "eBPF Go half — contract, ladder, report, health, badge, script" (`:58`) | **VERIFIED 2026-10-30** — `go build ./... && go test ./pkg/probe/ ./pkg/probe/bpf/ ./pkg/audit/ ./pkg/api/ -count=1` 0; ladder tests cover every rung (Windows, absent object, no BTF, no CAP_BPF, ready); `/api/health` `{"source":"tcpinfo","reason":"hôte non-Linux"}`; `bash kit/host-readiness.sh` n/a rows on Windows. "No object exists and nothing attached: slice-4 live row stays BLOCKED." | Windows side done. |
**Consequence:** thesis is legible as honest engineering (contract + fallback + report block verified) but illegible as measurement science (no `source=ebpf` cell has ever run). The seam is the critical path: only BLOCKED row whose clearing turns a `tcpinfo` proxy column into a kernel byte column; BBRv3 (27-event 3×3 `bbrv3-matrix-design.md:148-164`) and `.dist` both ride on it. IP-wise product is the *fallback ladder + provenance* pattern (typed `UnavailableError` vs `BuildError`, SHA-pinned object, coverage-ranked ratio), reusable even if kernel never cooperates.
### 8.4 Exported API (file:line)
#### 8.4.1 `pkg/probe` (non-test)
`PingSample{RTTms}` (`probe.go:22`); `Ping(ctx,target,count,intervalMs)` (`:28-43`, `-c/-i` Unix, `-n` Win, `count×interval+5000ms`, `len(out)==0` only error); `SmallClient()` (`:93-96`, shared 2s, `once+markedClient` `:108-111`); `SmallClientTimeout` (`:104-106`, fresh, `max(2s,4×delay)`); `SmallObject` (`:122-139`, GET→`io.Discard`, ms `us/1000.0`); `BulkSendTo` (`bulk.go:11-17`); `BulkSend` (`:20-40`, 64KiB zeros, 2s deadline, `Write`-error→end); `BulkReceive` (`:43-75`, 256KiB, cancel via `SetReadDeadline(pastTime())` `:77`); `BulkSource` (`bulksource.go:14-31`); `BulkRecvSource` (`:39-49`, `"D"/"D:<cc>\n"`); `BulkDownloadTo` (`:55-61`); `BulkDownloadNTo` (`:68-90`, N parallel, `[]uint64` dial order); `BulkSendNStats/BulkDownloadNStats` (`bulkstats.go:60-73`, `bulkNStats` skips dial-fails `:36-38`, `keepConn` withholds Close so TCP_INFO sees live fd `:17-25`, sums `RetransBytes`, `oneTCP` if ≥1 sampled `:43-56`); `ParseDownloadHello` (`bulksource.go:94-103`); `FlowResult{Bytes}` (`bulkflows.go:14-16`); `BulkSendNFlows` (`:23-36`, 3rd arg ignored); `BulkSendNTo` (`:42-64`, `""`→`model.Cubic` `:44`, `n<=1` single); `BulkStats{Bytes,RetransBytes,RetransMeasured}` (`bulkstats.go:11-15`, **no `Source`/`Kernel` yet** — spec §4.2 proposes `Source string` + `Kernel *KernelTelemetry`); `DialWithCC` (`bulk_cc_linux.go:14-28`: `net.Dialer{5s}` + `TCP_CONGESTION` `tcpCongestion=0x0d`; stub `bulk_cc_other.go:11-14` plain dial; R11 §`TCP_CONGESTION` per-socket allows string); `SetConnCC` (`ccsock_linux.go:15-34`; stub `ccsock_other.go:10`); `TCPRetransBytes` (`tcpinfo_linux.go:20-44`: `*net.TCPConn`→`File()`→`getsockopt(TCP_INFO)`→`Total_retrans×1448`; stub `tcpinfo_other.go:14`; raw syscall only, stdlib comment rejects `golang.org/x/sys` `:30-35`); `ErrTCPInfoUnavailable` (`tcpinfo_other.go:12`); `KernelSourceInfo{Source,Reason}` (`kernelsource.go:12-15`, Source∈`ebpf`/`tcpinfo`/`""`); `KernelSource()` (`:34-38`, `kernelFacts` `:19-31`: `/sys/kernel/btf/vmlinux` via `os.Stat`, caps approx `os.Geteuid()==0` `:26-30`).
#### 8.4.2 `pkg/probe/bpf` (non-test)
`Flow{Local,Peer,RetransPayload,RetransEvents,DropPackets,DropBytes,SRTTus,SndCwnd,SndSsthresh,State,Samples}` (`types.go:4-16`); `Summary{Available,Source,Attach,FallbackReason,ObjectSHA256,FlowsSeen,FlowsTotal,RetransPayload,RetransEvents,QdiscDrops,QdiscDropBytes,QdiscAvailable,MapFull,Flows}` (`:20-35`, Source∈`ebpf`/`tcpinfo`/`""`); `Globals{RetransPayload,RetransEvents,DropPackets,DropBytes,MapFull,QdiscAvailable}` (`:40-47`); `Status(goos,btf,caps,objectErr)` (`status.go:11-25`, injected for Windows rung tests, French reason `hôte non-Linux`/object error/`BTF absent`/`CAP_BPF absent`/else `ebpf`); `ObjectSHA256()` (`:28-35`, `sha256(ObjectBytes)`); `ObjectBytes()` (`assets_embedded.go:10-15` `//go:build ebpf_object`, `//go:embed bpf_probe.o`; stub `assets_absent.go:7-9` `//go:build !ebpf_object` returns `BuildError`); `UnavailableError{Reason}`/`BuildError{Reason}` + `Error()` (`errors.go:4-11`).
#### 8.4.3 Build assets (`pkg/probe/bpf/`)
`bpf_probe.c` (CO-RE, GPL-2.0; hooks `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`+`kprobe/tcp_rcv_established`, `raw_tp/kfree_skb`; `vmlinux.h`, `<bpf/bpf_helpers.h>`, `<bpf/bpf_tracing.h>`, `<bpf/bpf_core_read.h>`, `<bpf/bpf_endian.h>`; `bpf_probe.c:1-253`); `BUILD.md` (maintainer pin: "Source et cible posées 2026-10-29. L'objet n'est pas encore construit" `:8-12`, clang-18/libbpf-dev/bpftool/Ubuntu 24.04/kernel 6.8.0 `/sys/kernel/btf/vmlinux` `:16-19`, `make bpf` generates `vmlinux.h` → CO-RE compile → `bpf_probe.o.sha256` `:29-31`, verify `readelf -S … | grep -E 'raw_tp|fentry|kprobe'` `:35-41`, DWARF embeds source path so other dirs change hash `:45-48`); assets `assets_embedded.go`/`assets_absent.go` build-tag gate; `Makefile` `bpf` (bpftool dump → clang-18 `-O2 -g -target bpf -D__TARGET_ARCH_x86` → `sha256sum` → `.sha256`) + `bpf-check` (`:44-59`).
### 8.5 Dependencies (eBPF)
- **Stdlib only for probe today** (`probe.go:6-19`, `bulk.go:3-8`, `bulkstats.go:3-7`, `bulksource.go:3-9`, `bulkflows.go:3-9`, `bulk_cc_linux.go:5-10`, `tcpinfo_linux.go:5-10`, `dscp_linux.go:5-10`, `ccsock_linux.go:5-8`). `tcpinfo_linux.go:30-35` explicitly rejects `golang.org/x/sys` for raw `SYS_GETSOCKOPT`. None of `go.mod:5-10` (fyne/bubbletea/lipgloss/x-sys) needed.
- **Two local edges:** `pkg/model` only for `Cubic` (`bulkflows.go:8,44`) and `pkg/probe/bpf` ladder; `status.go:3-6` only `crypto/sha256,encoding/hex`; `assets_embedded.go:5` `embed`. Module `github.com/Realms4239/cgo` `go 1.25` (`go.mod:1-3`).
- **Build axes:** `linux` vs `!linux` (CC/DSCP/TCP_INFO) and `ebpf_object` for embedded vs absent (`assets_embedded.go:1`/`assets_absent.go:1`). Fuller loader axis (`linux && amd64` real loader vs stub, spec §3.5/§4.1 `Check()`/`Start()`/`Lookup`/`Snapshot`/`Globals`/`Attach`/`Close`) is **designed but not present** — no `session_linux.go`/`session_stub.go` today.
- **Build toolchain (maintainer-only, Linux 6.8):** Ubuntu 24.04, kernel 6.8.0, clang-18 (`BPF_CLANG=clang-18`), libbpf-dev headers (`<bpf/*.h>`), bpftool (`vmlinux.h` from `/sys/kernel/btf/vmlinux` per kernel docs). `.BTF`/`.BTF.ext` require clang ≥8.0 or pahole. `vmlinux.h` gitignored; only object + `.sha256` is the deliverable.
- **Missing runtime dep:** `go.mod` has **no** `github.com/cilium/ebpf` (or libbpf binding) — loader and map-read code do not exist. Spec §3.5 says `go.mod` *will* pin a specific cilium/ebpf release for `session_linux.go`; pin absent. Adding it is stop-and-ask per AGENTS.md.
- **Host capabilities for live attach:** `CAP_BPF` (since 5.8) = privileged BPF; `CAP_PERFMON` (since 5.8) = perf/BPF perf; `CAP_NET_ADMIN` = TOS/interface; root bypasses unless bounding set dropped cap; `CAP_SYS_ADMIN` overloaded legacy (avoid for new features). Spec §3.4 requires `CONFIG_DEBUG_INFO_BTF` + `/sys/kernel/btf/vmlinux` + `CAP_BPF` + `CAP_PERFMON` (+ `CAP_NET_ADMIN` already for `tc`); root supported path, `sudo setcap cap_bpf,cap_perfmon,cap_net_admin+ep ./cgo` fallback. `kernelFacts` (`kernelsource.go:29-30`) approximating caps as `euid==0` over-claims vs capabilities doc — root bypasses *unless* bounding set dropped cap. Real capget is a gap (§8.7).
**Incoming:** `docs/architecture.md:60-74` gates G0–G7 on probe outputs; audit/report consumes probe stats but must not import loader directly — spec §4.1 aliases `probe.KernelTelemetry = bpf.Summary` so `pkg/audit` never imports loader. Sibling modules call `probe.`/`audit.` — extraction must keep every call site compiling or provide adapter.
**Licence:** repo root MIT (sibling `probe-audit.md:173`); `cilium/ebpf` MIT (repo README). BPF C object GPL-2.0 (`bpf_probe.c:1` vs cilium example `Dual MIT/GPL`: GPL helpers require GPL-compatible BPF licence). Standalone observer must keep both licences with files; do not relicense object.
### 8.6 Primary-source requirements (eBPF)
#### 8.6.1 Kernel BTF + CO-RE (kernel docs)
BTF has two parts — kernel API + ELF file format. Kernel API involves `BPF_BTF_LOAD` → `BPF_MAP_CREATE`/`BPF_PROG_LOAD` with `prog_btf_fd`/`func_info`/`line_info`. ELF carries `.BTF` + `.BTF.ext` with `core_relo` sub-section; see LLVM relocation docs for relocation types and CO-RE relocation kinds/records. Function/line info enables "source annotated translated byte code, jited code and verifier log".
CO-RE = Compile Once – Run Everywhere. Producing it needs running kernel's BTF at `/sys/kernel/btf/vmlinux` → `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h`; including `vmlinux.h` eliminates system-wide kernel-header dependency. libbpf resolves recorded BTF relocations against runtime BTF and patches offsets at load. `BPF_CORE_READ` records field info for relocation; generic helpers in `bpf_helpers.h`; see ebpf-go annotations 2–3 (`__uint`/`__type`/`SEC`).
**Concrete requirements:** (a) commit prebuilt `bpf_probe.o` + pinned SHA-256 + `BUILD.md` with clang version, libbpf headers, BTF source, exact command (spec §§3.5, 8); (b) keep `vmlinux.h` generated + gitignored; (c) use `-g` and CO-RE relocations so `BPF_CORE_READ`/`bpf_core_enum_value` offsets are patched at load instead of hardcoded (C already does — `:105-108`, `195-197`); (d) make build reproducible — same toolchain + same BTF + same workdir must reproduce pinned SHA, else investigate (spec §3.5); DWARF embeds source paths, so other dirs change hash (`BUILD.md:45-48`); (e) verify sections `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`, `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb` via `readelf` (`BUILD.md:35-41`) and pin SHA via `TestEmbeddedObjectPinned` without `make`.
**What BTF does not do:** debug/metadata, not magic offsets. Does not make verifier-unsafe program load, nor make `tcpinfo` byte-exact. Missing CO-RE field fails whole collection load (spec §3.3) and falls back to TCP_INFO — honest behaviour.
#### 8.6.2 libbpf / cilium-ebpf loader lifecycle (libbpf Overview, ebpf-go)
libbpf four phases: **Open** (parse ELF, discover maps/progs/globals) → **Load** (create maps, resolve relocations, verify+load progs; no prog executed yet) → **Attach** (to tracepoints/kprobes/cgroup/packet hooks; work happens here) → **Tear down** (detach, unload, destroy maps). Skeleton (`<name>__open/__load/__attach/__destroy`) embeds bytecode so "no extra files to deploy".
ebpf-go mirrors natively: ELF → `LoadCollectionSpec` → `NewCollection` (loads maps+progs, `Collection.Close()` frees) or `LoadAndAssign` (selective load into struct, `Close()` pattern). Canonical Go app: `rlimit.RemoveMemlock()` for <5.11, `loadCounterObjects(&objs,nil)`, `link.AttachXDP`/`AttachTracepoint`/…, `objs.Close()` + `link.Close()`, map `Lookup` after null-check. Also ships `features` (bpftool equivalent), `btf` reader, `pin` (bpffs), `cmd/bpf2go` (compile C + generate Go + embed; endianness `*_bpfel.go`/`*_bpfeb.go`).
**Concrete requirements:** (a) pin specific cilium/ebpf release in `go.mod` (spec §3.5) — today absent (§8.5); (b) implement Open/Load/Attach/Tear-down in Go with verifier log captured; (c) `rlimit.RemoveMemlock()` best-effort on 6.8 memcg accounting (spec §3.3); (d) per-session `Collection` (no global session state) with concurrent-safe `Lookup`/`Snapshot`; (e) build-tag split: real loader `//go:build linux && amd64`, stub elsewhere with identical exported API (spec §3.5, §4.1). R9 also requires `//go:build ignore` on C files co-located with Go to avoid "C source files not allowed when not using cgo", plus tool dependency `go get -tool github.com/cilium/ebpf/cmd/bpf2go`.
**What loader does not do:** loads and attaches; does not author BTF, nor validate BPF C correctness beyond kernel verifier.
#### 8.6.3 Verifier + tracepoint + capability discipline
Verifier does DAG (disallow loops, unreachable) + path simulation (register/stack types). Enforces NULL-checked `map_lookup_elem` before access — `PTR_TO_MAP_VALUE_OR_NULL` becomes `PTR_TO_MAP_VALUE` only when checked `!= NULL`; un-checked yields `R0 invalid mem access 'map_value_or_null'`, misaligned yields `misaligned access off …`. C already respects: `flow_get_or_create` handles `BPF_NOEXIST` + `map_full` (`:80-99`) and callers handle NULL (`:83-99`, `132-141`).
Tracepoints are runtime hooks: off = branch+space only; on = synchronous callback in caller context. Three hooks therefore run in transmit/ACK/free context with target CPU preemption state — no per-packet userspace syscall, only end-of-window map reads (spec §2.1).
Capabilities: `CAP_BPF`/`CAP_PERFMON` are *weakest* caps covering privileged BPF / perf-BPF; `CAP_SYS_ADMIN` overloaded, avoid for new features. Loader must *check* them (capget) and fall back with typed reason, not approximate via euid or `stat` alone. Doctor must not claim attach capability from load-only `Check()` (spec §4.1 `Check()` = verify hash then load-only; doctor detail says "loadable (BTF + CAP_BPF); attach is verified by audit").
**What these do not do:** verifier proves safety, not metric correctness. ახმიენტობა (verify + load ≠ attach) and register tracking do not prove `skb->len` is payload — *semantic* claim grounded in `tcp_transmit_skb` / `tp->bytes_retrans` code, cited at spec §§2.1–2.3, not in verifier docs. Keep separate in thesis.
#### 8.6.4 RFC 8290 + tc(8) family
Per §8.1, RFC 8290 legitimately frames *why* FQ-CoDel/CoDel queue control matters (target/interval vocabulary for loaded-latency measurement). Does not normatively define ping, bulk, audit phases, DSCP marking, CAP 220 ms, or report schema. CAKE has no RFC (R1 §10.2 informal). Releasable module must cite RFC 8290 only for FQ-CoDel mechanics, cite `tc-cake(8)` for CAKE `ingress`/`diffserv4`/`ack-filter`, and cite kernel docs for eBPF — never cross-cite.
tc(8) vocabulary: shaping egress, scheduling egress, policing ingress; CAKE `ingress` counts already-traversed dropped bytes; netem impairs outgoing; TBF pure non-work-conserving shaper. Standalone observer must *document generated `tc` strings* against tc syntax and must *not* claim to apply shaping — `pkg/qdisc` owns application; observer only *reads* kernel after `tc` acted.
#### 8.6.5 Go module hygiene
Standalone `cgo-ebpf-observer` lives behind own `go.mod` + controlled path; v2+ needs `/v2`; pre-1.0 (`0.y.z`) may break API, 1.0.0 is promise. `require` is `path version` (SemVer tags, immutable); `go` directive minimum since 1.21. `ebpf_object` `embed` and raw `syscall` need no `require`; only new loader dep forces `require` + `replace` during migration if paths change.
### 8.7 Gap analysis (eBPF)
Ordered by severity. Local cites `file:line`; primary cites R# or `unverified — primary source not found`.
1. **Object-less and loader-less — `ebpf` tier cannot run.** `BUILD.md:8-12`: "L'objet n'est pas encore construit : aucun hôte Linux 6.8 stable n'était joignable." `go.mod:5-10` has no loader dep. Go has only `ObjectBytes`, `Status`, `ObjectSHA256` — none of the four libbpf phases exists in Go. Spec §4.1 designed API (`Check`, `Start`, `Lookup`, `Snapshot`, `Globals`, `Attach`, `Close`) absent (`glob pkg/probe/bpf/*` returns only `types.go`, `status.go`, `errors.go`, `assets_*.go`, `bpf_probe.c`, `BUILD.md`, tests). `KernelSource` (`kernelsource.go:34-38`) will always fall to `tcpinfo` (`status.go:15-17` via `BuildError`, `assets_absent.go:8`). Until object + SHA built on pinned 6.8/clang-18/libbpf/bpftool stack (`BUILD.md:16-31`) and loader attaches three hooks and reads `flows`/`globals`, every `ebpf` claim is toy (see §8.10). Sole reason VERIFY `:29` stays BLOCKED.
2. **Host-fact heuristics lie about caps and BTF.** `kernelFacts` (`kernelsource.go:19-31`) uses `os.Stat("/sys/kernel/btf/vmlinux")` alone for BTF and `os.Geteuid()==0` for caps. Root bypasses *unless* bounding set narrowed, so euid==0 over-claims; BTF *presence* ≠ parseable BTF (`CONFIG_DEBUG_INFO_BTF` per spec §3.4). Fix: parse BTF (via `cilium/ebpf/btf` or `bpftool` READ, not stat) and use `capget` to test `CAP_BPF`/`CAP_PERFMON` in effective set; doctor `warn` detail must name missing fact (spec §4.5).
3. **Coverage contract not yet in code.** `bulkstats.go:29-56` (`bulkNStats`) has no `Source`/`Kernel` plumbing; `BulkStats` (`:11-15`) has no `Source`/`Kernel`; no `kernelSession` interface seam (`Start()` factory) from spec §4.2, no sender-side key derivation (upload = client `LocalAddr/RemoteAddr`, download = reversed), no query-before-real-close before `keepConn` release, and no all-or-nothing gate ("require *all* dialed flows observed with ≥1 sample/event; qdisc drop alone does not prove retransmit accounting; missing ≠ zero; exact zero has a sample" — spec §4.2). `BulkSendNStats`/`BulkDownloadNStats` (`:60-73`) still sum TCP_INFO only. Any extraction must move this logic into new module (§8.9), not duplicate in `pkg/audit`.
4. **Probe fidelity gaps carry into observer if left as-is.** *Ping by exec + locale parsing.* `parsePing`/`parsePingLine` (`probe.go:49-85`) tolerates `temps=`/`temps<` and `<1ms`→`v/2` (`:81-83`), but ping-output contract has no primary source (`unverified`); `Ping` returns partial output as success when output exists even on non-zero exit (`:38-43`), and `parsePing` sorts, losing per-packet order. Windows `-n` has no `-i` (`:31-34`) — interval silently ignored. *Small-object timeout triad.* Shared 2s (`:93-96`) vs fresh bench (`:104-106`) vs audit default 8s (`pkg/audit/audit.go:231` sibling) — three timeouts, no rule; `SmallClientTimeout` allocates fresh `Transport` per call (`:105`), defeating reuse. *Bulk payload heuristics.* Zero-filled 64 KiB writes (`bulk.go:22`) measure compressible-zero goodput (no entropy; fill policy `unverified`); 2s write deadline (`:33`) comment-justified only; `BulkSend` swallows `Write` errors as clean end (`:36-38`); `bulkNStats` skips dial failures (`:36-38`) so `Bytes` < `n` silently; `RetransMeasured` = true if *any* socket sampled (`:48-56`) while `RetransBytes` sums only sampled — partial basis. *Byte-unit mismatch.* `×1448` (`tcpinfo_linux.go:15,43`) counts segments × MSS; eBPF `skb->len` counts payload TSO-inclusive (`:119`). Two tiers count different things; no code reconciles them (spec §2.3).
5. **Report seam designed, not wired.** Spec §4.2–§4.5 proposes `BulkStats.Source`/`Kernel`, `metrics.RetransRatio.Source`, `audit.LoadOutcome.RetransSource`/`Kernel`, `Result.KernelUp/Down`, `Report.KernelTelemetry{Up,Down}`, Markdown "Télémétrie noyau" block, doctor `ebpf` check (`ok`/`warn`/`fail` on `Check()`), and `pkg/audit/kernel_adapter.go` decoupling — none of which exists in current `pkg/probe` tree for this seam. Observer must publish `bpf.Summary`/`Flow`/`Globals` cleanly so audit can consume it without importing loader (spec §4.1 alias `probe.KernelTelemetry = bpf.Summary`).
6. **Bench identities and tunables hardcoded for bench, not API.** `BulkSendNTo` defaults `""`→`model.Cubic` (`bulkflows.go:43-45`); `BulkSendNFlows` 3rd arg ignored (`:23`); `DialWithCC` 5s dial timeout (`bulk_cc_linux.go:15`); `SmallClient` 2s; `SAMPLE_NS` 50ms (`:27`); `MAX_FLOWS` 4096 (`:26`). Releasable API needs these as explicit `Options`/`Config`, not globals (sibling §5.6 class).
7. **Versioning/schema vacuum.** No `LinkAuditSchema` version in `pkg/probe`; standalone module needs own schema stamp so `telemetry.csv`/`link_audit.csv` freeze (AGENTS.md "Never touch" + invariant 3) is versioned, not fiat. Today `AppendLinkAudit` self-migrates headers (sibling §5.10) — observer should keep schemas as files with versions.
**Sibling TCP_INFO depth (ebpf-tcpinfo research, R1–R9):** App lifecycle Open→Load→Attach→Tear-down; skeleton embeds bytecode; CO-RE via `/sys/kernel/btf/vmlinux` + `bpftool dump format c > vmlinux.h`; `BPF_MAP_CREATE`/`LOOKUP`/`UPDATE`/`DELETE` hash/array/percpu types; helpers `bpf_map_lookup_elem`/`update_elem(BPF_NOEXIST)`/`ktime_get_ns`/`probe_read_kernel`; `TCP_INFO` since 2.4 returns `struct tcp_info` from `/usr/include/linux/tcp.h`; `TCP_CONGESTION` per-socket string; `TCP_INFO=11`, `TCP_CONGESTION=13 (0x0d)`; `struct tcp_info` fields incl. `tcpi_total_retrans` (`__u32`), `tcpi_bytes_retrans` (`__u64`), `tcpi_delivery_rate`, `tcpi_min_rtt`; `CAP_BPF` (5.8+) privileged BPF, split from `CAP_SYS_ADMIN`; `cilium/ebpf` pure-Go loader with `link`/`btf`/`rlimit`/`features`/`bpf2go`, MIT, amd64/arm64; BBRv3 model-based (delivery-rate/RTT/loss → `max_bw`/`min_rtt`/BDP; pacing/`cwnd`/`send_quantum`; Startup→Drain→ProbeBW/ProbeRTT; `C.SMSS` SHOULD be payload-only). Observer does NOT implement BBR — only selects CC by name and measures outcomes; must never claim BBRv3 exactness; `min_rtt`/`max_bw` model state inside sender is `unverified` from this module's vantage.
### 8.8 Additional unverified kernel facts (do not promote)
Each stated in repo as fact, but no fetched primary corroborates it. Stays `unverified — primary source not found` until kernel-doc/manual fetch covers them:
1. `tcp_retransmit_skb` raw args are `(sk, skb)` and `kfree_skb` raw args are `(skb, location, reason)` — `bpf_probe.c:114` vs `:190`, spec §2.1–§2.2. Fetched kernel docs index does not list these prototypes.
2. `srtt_us` stored left-shifted by 3, hence `>>3` (`:167`) and spec §2.4. No fetched source shows `struct tcp_sock` `srtt_us` encoding.
3. `skb->len` at retransmit tracepoint is payload with `transport_header==0xFFFF` sentinel (`:6-9`). Behavioural claim about `tcp_transmit_skb` cloning; no fetched source.
4. `SKB_DROP_REASON_QDISC_DROP` timing: `__qdisc_drop()` static-inline + `tcf_set_drop_reason()` + `kfree_skb_list_reason()` path, and 6.8-vs-7.1 `qdisc_drop` tracepoint history (commit `ff2998f29f39`, v7.1) — spec §2.1. Kernel git history not fetched.
5. `SAMPLE_NS=50ms`, `MAX_FLOWS=4096`, "64-octet `flow_stat` matches Go `bpfFlowStat`" (`:26`, `:49`) — design constants, not kernel facts; Go mirror struct does not exist on disk (`types.go` has no `bpfFlowStat`).
`dscpEF=46` (RFC 3246) and `46<<2=184` TOS (`dscp_linux.go:16,24`) — RFC 3246 not fetched; mark `unverified` for EF value as primary claim (code is the claim). `retransMSS=1448` (veth MTU 1500) (`tcpinfo_linux.go:15`) — bench-measured proxy; no fetched source derives 1448. Short-`struct tcp_info` prefix-fill behaviour (comment `:32-34`) — kernel-prefix behaviour is **not in** fetched man/UAPI; `unverified`.
IPv6/non-IPv4 retransmits go to `globals` only (`:121-128`); flow key omits netns (spec §3.2 shortcut `netns,tuple` upgrade noted). Per-flow drop attribution requires existing flow entry (`:244-249`); else global only. All faithful to source but **unverifiable from fetched sources** as production guarantees — scope statements, not kernel theorems.
### 8.9 Target tool — `cgo-ebpf-observer` (single independent repo)
One repo `cgo-bpf-observer` (candidates `cgo-bpf-observer`/`cgo-tcpinfo` collapse into one module with two entry points — ponytail rule 1: one repo, thin CLI; two repos would duplicate module boundary). Library is product; CLI is sugar. No metrics/audit/remediate/model imports — `go list -m all` must show stdlib + `cilium/ebpf` only.
```
cgo-bpf-observer/
  go.mod                          # module github.com/Realms4239/cgo-bpf-observer; go 1.25; require cilium/ebpf <pinned>, x/sys
  README.md                       # what it measures, what it refuses to claim, BLOCKED preconditions
  Makefile                        # bpf / bpf-check / test / verify   (same as root Makefile:44-59 shape)
  cmd/cgo-bpf-observer/main.go   # CLI: attach + dump Summary as JSON; `tcpinfo` subcommand
  internal/probe/                 # moved pkg/probe/*.go
  internal/probe/bpf/             # moved pkg/probe/bpf/{types.go,status.go,errors.go,assets_*.go,bpf_probe.c,BUILD.md}
  internal/probe/bpf/session_linux.go   # loader: cilium/ebpf Open/Load/Attach/Tear-down
  internal/probe/bpf/session_stub.go    # //go:build !linux || !amd64 identical API, typed UnavailableError
```
`internal/` (not `pkg/`) so external callers consume **one** public package (`probe`) + returned JSON/Summary values. Loader types stay internal; `probe.KernelTelemetry = bpf.Summary` alias lives at seam (as ground-truth spec intends).
```go
package probe  // stdlib + cilium/ebpf only
// — TCP_INFO tier —
func TCPRetransBytes(c net.Conn) (uint64, error)
var ErrTCPInfoUnavailable error
// + TCPInfoSnapshot returns full struct tcp_info (optional; default bytes_retrans + del. rate)
// — CC selection —
func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error)
func SetConnCC(conn net.Conn, cc string) error
// — Kernel tier —
func KernelSource() KernelSourceInfo
type KernelSourceInfo struct{ Source, Reason string }  // Source ∈ ebpf|tcpinfo; Reason names missing fact
// — Session (new; designed in ground-truth spec §4.1) —
type Session interface {
    Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
    Snapshot() []bpf.Flow
    Globals() (bpf.Globals, error)
    Attach() string  // "fentry"|"kprobe"
    Close() error
}
func Open() (*Session, error)   // fresh maps per session; load+attach, all-or-nothing per load
func Check() error              // SHA-verify then load-only (no attach); BuildError|UnavailableError
// — Bulk with kernel truth —
type BulkStats struct {
    Bytes []uint64 `json:"bytes"`
    RetransBytes uint64 `json:"retrans_bytes"`
    RetransMeasured bool `json:"retrans_measured"`
    Source string `json:"source,omitempty"`          // + ebpf|tcpinfo|""
    Kernel *bpf.Summary `json:"kernel,omitempty"`    // + coverage + per-flow
}
func BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)
func BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)
```
CLI: `cgo-bpf-observer capture --flows 4 --duration 10s --cc bbr --out summary.json`; `cgo-bpf-observer tcpinfo --conn tcp://host:port`; `cgo-bpf-observer doctor` (`Check()` → ok|warn|fail, names missing fact).
**Non-goals (explicit):** No BBR implementation — CC is socket option name. No qdisc/tc control, no netns, no economics/audit/report. No IPv6 per-flow keys, no queue-dwell/ECN/ACK pairing — inherited 6.8 scope.
**Design notes (from ebpf-probe §6):** (a) observer has zero `metrics`/`audit`/`remediate`/`model` imports; (b) callers never import `probe/bpf` loader directly — they consume `bpf.Summary`/`Flow`/`Globals` via alias or `KernelTelemetry` interface; (c) every "absent" is flag/error/`nil`, never silent zero; (d) CC/DSCP-marking, sinks, tunables are explicit values; (e) CLI optional sugar — library is product. `Options{CC,Flows,SampleNS,MaxFlows,Attach}`, `PingRunner` exec seam, `PingResult{Attempts,Received,Samples}`, `SmallClient{Timeout,MarkEF}`, `SmallObject`, `BulkUpload/BulkDownload`, `ObjectSHA256`, `KernelSource` ladder with real capget + BTF parse.
### 8.10 Slices (eBPF, stop-and-ask gates)
Stop-and-ask (AGENTS.md): new dependency (`cilium/ebpf`), live bench run, frozen-data touch — each requires human approval before acting. No slice touches code in this pass.
| Slice | Scope | Acceptance (named command / host) |
|---|---|---|
| **0 — Extraction & pin** | Move `pkg/probe`+`bpf` to new module; `go.mod` stdlib-only; keep all APIs byte-identical; add `make bpf/bpf-check` Makefile shape. No loader, no dep. | `go vet ./...`; `go test ./... -count=1` (Windows); `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` 0; `go list -m all` stdlib only. |
| **1 — TCP_INFO standalone** | Extract `TCPRetransBytes` + `DialWithCC`/`SetConnCC` + `keepConn`/`bulkNStats` + `BulkStats` as lib+CLI (`tcpinfo`). Add optional full-`struct tcp_info` + `tcpi_bytes_retrans` exact-byte path alongside 1448 proxy, never silent. | `go test ./...` incl. non-TCP-error + multi-flow `RetransMeasured`; live `cgo-bpf-observer tcpinfo` vs loopback TCP shows non-zero `tcpi_total_retrans` under `tc netem loss` (manual, human-approved). |
| **2 — Object + loader** | Build `bpf_probe.o`+`.sha256` on pinned 6.8/clang-18/[R1] stack; pin `cilium/ebpf`; implement `Open/Check/Lookup/Snapshot/Globals/Attach/Close` [R1][R8]; replace `kernelFacts` stat/euid with BTF-parse + `capget` [R1][R7]; split `session_linux.go` (`linux && amd64`) / stub. Verifier log captured. | On 6.8 VM with BTF + `CAP_BPF`: `doctor ok`; loopback flow in `Snapshot()` with `samples>0`; forced retransmit yields `retrans_payload>0` same order as TCP_INFO; without caps → `tcpinfo` named reason. `make bpf-check` 0; `TestEmbeddedObjectPinned` passes. |
| **3 — Coverage contract & audit wiring** | `BulkStats.Source/Kernel`, `kernelSession` seam, sender-key derivation (upload client, download reversed), query-before-close, all-or-nothing gate (all dialed flows observed ≥1 sample/event; qdisc-drop alone ≠ observed; missing ≠ zero; exact zero has sample). Wire `KernelSource`→`Summary`→`KernelSample` neutral adapter. | Unit: partial + map-full + zero-with-sample + zero-without-event fall back to `tcpinfo` with coverage `n/N` + reason. `dualns` upload+download `source=ebpf`, coverage 4/4 (human-approved). |
| **4 — Hardening** | Pin + fuzz key-encoding (host-order addr/port), perf overhead (<2% per spec §6.5), `vmlinux.h` reproducibility docs, 7.1+ `qdisc_drop` note (later gate). | Overhead run recorded; key test green; reproducible SHA documented. Human decides 7.1 port. |
| **5 — 7.1+ upgrade + `.dist` correlation (conditional)** | Once kernel ≥7.1, swap `kfree_skb` filter for native `qdisc_drop` (spec §§2.1, 7.12) with new object+hash; correlate drop corpus with `.dist` traces (30-sample floor) and BBRv3 3×3 (require 7.1 + BBRv3 registered). Cross-kernel numeric merging with frozen 6.8 corpus forbidden (BBRv3 spec §6). | 7.1 run produces `qdisc_available=true` via native hook and `.dist`-gradeable trace ≥30 samples, each in own run set labelled `bench=dualns`/`kernel=7.1`. |
| **6 — Publish** | Tag `0.x.y`; audit consumes `BulkStats.Source`/`Kernel` (spec §4.1 alias) without re-importing loader; VERIFY rows updated with named command + host + date. Until `VERIFIED`, stay BLOCKED — never promote by confidence. | — |
### 8.11 Anti-toy criteria (eBPF)
A build failing **any** of these is toy — fixtures only, never called `ebpf` in README/report:
1. **Object real:** `bpf_probe.o` + `.sha256` committed, built on pinned 6.8/clang-18 stack; `make bpf-check` exit 0; `readelf -S` shows `raw_tp/*`, `fentry/*`, `kprobe/*`; embedded SHA == pinned SHA (`TestEmbeddedObjectPinned`).
2. **Loader real:** `go.mod` pins `cilium/ebpf`; `Open` implements Open/Load/Attach/Tear-down with verifier-log capture; `Snapshot()` under real traffic returns `samples>0`; `Close()` detaches cleanly. Unit-only `Status` without attach does not count.
3. **Attribution exact:** bulk with `tc netem loss` produces `retrans_payload>0` and `retrans_events>0` on sender-side flows, same order-of-magnitude as TCP_INFO `Total_retrans*1448` on same window; `srtt_us`/`snd_cwnd` samples non-stale (>50 ms freshness). Zero-flow fixture JSON does not count.
4. **Fallback named, never silent:** every non-`ebpf` window carries `Source=tcpinfo|""` + `FallbackReason` naming missing fact (non-Linux, missing BTF, no `CAP_BPF`, verifier/CO-RE, attach, partial coverage, map-full, remote sender) — ladder in `status.go:11` plus spec §5 table. Missing ≠ zero; exact-zero requires sample.
5. **No silent MSS:** TCP_INFO path labels byte count as proxy (`×1448`) unless `tcpi_bytes_retrans` exact path used; eBPF labels payload-vs-wire honestly (`skb->len` semantics stated, never "ACKed").
6. **Cross-build + Windows gates green:** `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` exit 0; `go vet` + `go test ./... -count=1` exit 0 on Windows; `GOOS=linux go vet` where supported.
7. **Independent verification:** live attach, overhead (<2%), dualns 4/4 claims re-run by verifier who is not builder, on named host, with date + command + output/quoted checkpoint (VERIFY row rule). Windows-green explicitly not "done".
8. **No BBRv3/retransmodel smuggling:** CC selection success never presented as BBR model evidence (`max_bw`/`min_rtt` not exposed). BBRv3 matrix/engine remain separate BLOCKED rows until 7.1 host + human-approved run exists.
Additional (A1–A8 from ebpf-probe §8): A1 Object provenance (empty→toy); A2 Load≠attach (load marketed as attach→toy); A3 All-or-nothing coverage (partial summed as complete→toy); A4 Submitted vs delivered honesty ("Delivered/ACKed" wording→toy); A5 Unit honesty (proxy labelled payload→toy); A6 Fallback ladder completeness (any rung returns `ebpf` or panics→toy); A7 Verifier discipline (suppressed verifier→toy); A8 Overhead bound (≥2% shipped unnamed→toy). Overhead gate: 60 s audit eBPF on/off: goodput delta <2%, all phases run, no audit error.
### 8.12 DEFERRAL (eBPF)
- **Live-attach + overhead + dualns 4/4:** DEFERRED until Linux 6.8 build host gate (human) + object built + loader written. Windows body: `VERIFIED` rows must say "Windows body only", never "verified" where the condition is a Linux kernel feature (`cgo-bpf-observer` VERIFY rows checklist, appendix D).
- **7.1+ `qdisc_drop` swap:** DEFERRED until target kernel ≥7.1 (human). Until then `kfree_skb` hook is the 6.8-correct hook; do not ship `qdisc_drop` on 6.8 (tracepoint absent).
- **BBRv3 3×3 + `.dist` ≥30 floor:** DEFERRED until 7.1 kernel + BBRv3 registered + field trace. Cross-kernel numeric merging with frozen 6.8 corpus forbidden.
- **TCP_INFO exact-byte (`tcpi_bytes_retrans`) as default:** DEFERRED until Slice 1 live validation; proxy stays labelled until then.

---

## 9. Per-Module Tags / Cutover Detail + Rollback (per-tool procedure)

### 9.1 Per-tool tag procedure (uniform)
```
git tag <tool>/v0.1.0          # canonical v + SemVer; never move; never edit contents post-release
go list -m all                  # in consumer: expect <tool> v0.1.0, no replace
goreleaser release --clean --dry-run   # SBOM + checksums + provenance draft
# cutover:
go get <tool>@v0.1.0 && go mod tidy && go list -m all
make verify-fast               # + python tools/verify.py --selftest (six caught)
# verifier re-runs; only then VERIFIED row moves (builder != verifier)
```
`replace` allowed during migration (main-module only) and **must be dropped before tagging**. `exclude`/`retract` only via policy (bad release → newer tag + retract line, never silent rewrite). `go.mod` `require` bump is part of the same commit as consumer code using it — partial upgrades are not commits.
### 9.2 Per-tool cutover + rollback hooks
| Tool family | Cutover hook | Rollback (one command) |
|---|---|---|
| Pure libs (model/metrics/cost/profile/results/figures/qdisc/remediate/probe/iface/doctor/api) | `go get` + `tidy` + `list` + dry-run | `go get <tool>@<prior v0.1.x>` + `tidy` + `make verify-fast` |
| Binaries (server/kit/measure/tui/wall/gui) | tag + `goreleaser` + `keep_version` lock | prior tag + `docker compose pull <prior digest>` (or prior archive + `SHA256SUMS`) |
| Supply-chain (dist/isolate/ingress/exporter/slo/guard) | digest-pin + SBOM + sign + provenance | `compose pull prior digest` + `nginx -t && systemctl reload nginx` + `cgo shape --restore` |
| Evidence (seal/attest/ledger/cite) | anchor + bundle atomically | prior anchor + prior bundle (both kept; never overwrite) |
| DEFERRED (wis2-probe, bpf-observer live) | gate on §7.7/§8.12 Q-items | no rollback — unbuilt; if built, prior tag |
### 9.3 `keep_version` lock update (every cutover)
Update every `keep_version` entry in `mise.lock` / `package.json` / Vite pin that names the tool version. Diff must show only version bumps, no logic change piggy-backing, or the slice is split (one logical group per commit).

---

## 10. Every Other Gap Tool — Per-Tool Specs (T01–T34)

*Each tool block: Responsibility · Exported API (file:line) · Dependencies · Primary-source requirements (R# cited) · Gap analysis · Target repo/module structure · Per-tool slices · Anti-toy refusal list · DEFERRAL conditions (human gates).* The founder line for each block is its registration pin; read it in the implementation plan before coding.

### 10-T01 cgo-model (pkg/model)
- **Responsibility:** Domain-types package, not an application (`docs/architecture.md:21`). Profile value object with asymmetric rule (`CapacityMbps` reference, `CapacityUpMbps` optional, absent=symmetric, `up` capped to `down` — `types.go:7-11,30-41`), axes (qdisc, CC, G0–G7, statuses, phases, `Event`, headers — `types.go:70-184`), `ParseCCList` kernel-order dedupe (`cc.go:5-19`). Does NOT shape/probe/store/serve HTTP.
- **Exported API:** `Profile` (`:11-28`); `CapDown` (`:31`); `CapUp` (`:36-41`); `Profiles` P1 80/20/2/0, P2 20/100/15/0.5, P3 5/600/30/1, P4 100/40/20/0.3 (`:43-51`); `ProfilesMu` (`:53-57`); `BuiltinProfileIDs` (`:59-68`); `Qdisc` `PfifoFast/FqCodel/Cake` `AllQdiscs` (`:70-78`); `CC` `Cubic/BBR/BBRv3` `AllCC/KnownCC` (`:80-92`); `Gate/G0…G7/GateCount=8/GatePass/Degraded/Invalid` (`:95-115`); `PhaseBaseline/Charge/Recup` + `30/120/30` (`:119-125`); `Event` frozen 23-col (`:128-165`); `AQMEvalHeader/LinkAuditHeader` (`:171-184`); `ParseCCList` (`cc.go:8-19`). Tests: `TestParseCCList` (`cc_test.go:8-22`), `TestBBRv3IsKnownButNotDefault` (`:24-39`), `TestProfileAsymmetricCapacity` (`profile_asym_test.go:8-24`).
- **Dependencies:** Stdlib only (`"sync"`, `"strings"`). Imported by `profile.go:9`, `campagne`, `qdisc`, `audit`, `api`, `probe`, `doctor`, `figures`, `cmd/cgo`. No own `go.mod` today (only `go.mod:1-3`).
- **Primary-source requirements:** Go Modules (path=ID, `go` mandatory minimum since 1.21, `/vN` v2+; `module` directive + prefix; `require`/`replace`/`exclude`/`retract`); SemVer (§1 MUST declare public API; §3 immutable; §6 PATCH compat fix; §7 MINOR compat functionality; §8 MAJOR incompat; §4 `0.y.z` MAY change); RFC 8290 §1 `FqCodel` names Linux discipline only (R3 §1 "mainline Linux kernel as the fq_codel queueing discipline"); CAKE/PFIFO `unverified — primary source not found`.
- **Gap analysis:** G1 exported mutable globals invite races. G2 deprecated `LossBurstP/R/H/K` (`:18-27`) leak into JSON. G3 `BuiltinProfileIDs` init-frozen from mutable map (stale). G4 P1–P4 literals lack measured-vs-illustrative tag. G5 CSV headers frozen by policy, not versioned contract.
- **Target:** `cgo-model` (own `go.mod`, MIT, `go 1.25`). Drop `LossBurst*` (separate `LegacyBurst` type only for CGO CSV importer, migration note); `Profile.Validate()` promotes HTTP bounds (1–1000, 0–600, 0–100, 0–10, `docs/api.md:51`); `Registry` (`Get/Set/List/BuiltinIDs`, `Default()` P1–P4) replaces globals (deprecated shim one cycle, R2 deprecation → MINOR with ≥1 minor before removal); `SchemaVersion` + `ColumnIndex(name)`. Stdlib-only.
- **Slices:** M-1 Freeze & document (godoc every export, README "types + parser only, no shaping/probing" + R3 citation; `go vet/test` green). M-2 Extract (`v0.1.0`, moved tests, `GOFLAGS=-mod=readonly` offline green; R2 rule 4 unstable). M-3 Re-point CGO (`make verify` green; headers byte-identical). M-4 Harden (`LegacyBurst` path, example `examples/profiles/main.go` offline).
- **Anti-toy:** A1 Declared API (SemVer §1). A2 Independent SemVer. A3 Immutable releases. A4 Own `go.mod`, zero requires, offline build. A5 No mutable global without sync contract. Failure A1–A5 = toy.
- **DEFERRAL:** CAKE/PFIFO normative pages + P4 Starlink provenance stay `unverified` until fetched (docs only, not gating). Module path + license: human decision before M-2.

### 10-T02 cgo-metrics (pkg/metrics)
- **Responsibility:** Calculation core (`docs/architecture.md:16`, `metrics.go:1-2`): summaries → bufferbloat grading → economics scenarios + legacy tier → CAP SLA → verdict cascade + VoIP + inference. No shaping/probing/storage/HTTP. CAP 220/95 is DGM objective (`cost.go:75-83`).
- **Exported API:** `Percentile(p)` (`:13-26` pre-sorted, `n==0→0`), `Summary{P50,P95,Median,IQR}` (`:28-31`), `Summarize` (`:34-44`), `DeadlineOKPct` strict `<` (`:47-58`), `MB/GB=1024ⁿ` (`:72-75`), `PriceTier/Tiers(6)/DefaultTier/SetDefaultTier/ActiveTier*/TierByName` (`:78-141`, `tierMu` `:99`), `CostARPerH/HTier` `×20` (`:146-156`), `JFI` (`:164-177` `(Σx)²/(n·Σx²)`, `unverified` provenance), `VoIPR` (`:190-219`, comment/code mismatch, ITU `unverified`); `BufferbloatGrade/Worst/Verdict` A+<5/A<30/B<60/C<200/D<400/F≥400 (`bufferbloat.go:11-63`, Waveform `unverified`, NaN→`""`); `Economics*` half (`cost.go`, →T03); `Median/Quartiles/IQR/BootstrapMedianCI95/MannWhitneyTwoSided/CliffDelta/Interp` (`stats.go:32-149`, all stats `unverified`, French labels); `VerdictLevel/Neutral/Red/Amber/Green/VerdictInput/Verdict/AssessVerdict` first-wins cascade (`verdict.go:8-68`, hardcoded French). 34 `Test*` pin bands/zero-window/no-scenarios/never-green (`docs/VERIFY.md:63`).
- **Dependencies:** Stdlib only; no internal imports (leaf). Incoming: `campagne`,`audit`,`api`,`results`,`cmd/cgo/tui_audit.go:33,173`. Today `github.com/Realms4239/cgo` `go 1.25`.
- **Primary-source:** RFC 8290 Experimental (Status "not Standards Track"), MUST only packet-limit (§5.2.3), defaults 100ms/5ms/1514/1024 (§5.2.1/2/4/5). CAKE no RFC (§7/§10.2 informative). RFC defines none of percentiles/bands/CAP/JFI/E-model. Prometheus text: LF-terminated last line, one HELP/TYPE before first sample, unique metric+labels, `NaN/+Inf` valid, histogram `_bucket{le}` + mandatory `{le="+Inf"}==count` ordered. Go Modules + SemVer: new path before publish, MUST declare API.
- **Gaps:** G1 global mutable `Tiers/DefaultTier/Plan*/Tariff*`. G2 French strings are API (`BufferbloatVerdict`,`AssessVerdict`,`CliffInterp`,`Disclaimer`,CAP label). G3 Madagascar globals, stale Airtel (5556 vs doc 3333). G4 No exposition (only `pkg/api/metrics.go:17-41`). G5 `VoIPR` mismatch, ITU `unverified`. G6 Absent 0 vs `""` vs NaN inconsistent. G7 Two quantile/cost/unit paths (`Percentile` vs `quantileSorted`; `×20` vs explicit window; `GB=1024³` vs `1e6`). G8 `Percentile` pre-sorted contract implicit.
- **Target:** `cgo-metrics` (economics → T03 `cgo-cost`). `Summary{N,P50,P95,IQR}` single path; `DeadlineTarget` param (`CAPTarget={220,95}` opt-in); `Grade` codes out; tariffs injected (`DefaultMGATariffs()`), delete ×20; `Level` codes; `VoIPR` only after G.107 fix; `RenderText` NEW per Prometheus R2.
- **Slices:** S1 Freeze+declare (port 34 tests). S2 De-globalise (adapter, race clean). S3 Unify+boundaries (exactly-220/95/NaN/empty). S4 Codes + exposition golden tests. S5 Tag `v0.1.0`, migrate, no `replace`.
- **Anti-toy:** 1. Hand vectors per family. 2. Provenance honesty (`unverified` until fetched). 3. No silent zero. 4. No global mutable. 5. Real consumer + real R2 scrape. 6. Version discipline.
- **DEFERRAL:** G.107/Jain/Waveform refs stay `unverified` until fetched; `VoIPR` stays experimental until spec-vector pinned. Orange/Airtel tiers excluded from defaults until values fetch succeeds.

### 10-T03 cgo-cost (economics split from metrics)
- **Responsibility:** `cost.go:8-9`: what does wasted bandwidth cost in Ariary under which contract? Model A CIR depreciation, Model B metered + emergency, Model C CAP SLA (220/95), assembly `BuildEconomics` (`:192-238`). Retail tiers (`metrics.go:60-156`; `GET /api/cost/tiers|enterprise`, `POST /api/cost/tier`, `server.go:417-475`). Never probes/shapes/stores/FX.
- **Exported API:** `EnterprisePlan` (`:12-17`), `PlanGOLD/PLATINE` (`:20-24` GOLD 1M/200Mbps CIR, PLATINE 2.5M/FUP 30T), `AmortizedARPerSecond` 720h (`:26-30`), `CapacityWasteARPerHour` (`:36-41`, `transported<=0→0`), `B2BTariff` (`:44-50`), `TariffWholesale/APN` (`:52-57` 81.38 derived, 150 `Quoted:true` `unverified`), `RetailEmergencyARPerGB=5556` (`:62`, 25000/4.5), `MeteredWasteARPerHour` (`:67-73`, `window<=0→0`, ×20 is 180s case), `CAPDeadlineMs=220/Threshold=95.0` (`:80-83`), `CAPSLA/AssessCAPSLA` empty→`Measured:false` (`:88-112`, breach `pct<95`), `RetransRatio` (`:117-120` `measured/estimated/""`), `EconomicsScenario/Numbers` (`:123-143`, `omitempty`), `Economics/PrimaryName/Disclaimer/BuildEconomics` (`:159-238`, no ratio/transport → zero scenarios `:201-203`; primary 4g/5g→APN, vsat→wholesale, else→GOLD `:170-183`); tiers `MB/GB/PriceTier/Tiers(6)/DefaultTier/SetDefaultTier/ActiveTier*/TierByName/CostARPerH/HTier` (`metrics.go:72-156`, ×20 hardcoded `:155`); endpoints `GET tiers/enterprise` (`server.go:417-452`, single-source no recopy `:420-422`), `POST tier` 409 running/400 unknown (`:455-475`).
- **Dependencies + tariff primaries (re-checked 07/11/2026):** Stdlib only. Yas Net Month 25k/4.5Go CONFIRMED; GOLD 1M/200Mbps + PLATINE 2.5M/FUP 30T CONFIRMED; wholesale 81.38=2.5M/30,720 CONFIRMED; `yas-month-100gb` 2,000 STALE (now 250Go→800); Airtel 5,556 STALE (doc 3,333; live JS-gated); Orange Be Connect JS-gated; APN 150 quote `unverified`; FTTH 490 CONFIRMED/inconclusive two passes.
- **Primary-source requirements:** WMO Guide §23 "real time" = seconds-to-minutes, no numeric threshold → deadline/target MUST be caller params. Yas grids fetched price+volume → every constant needs operator+offer+date, re-fetchable. Airtel/Orange Be Connect JS-gated → exclude until values fetch; voice+data Be MUST NOT ingest as pure data (4–7× overstatement). Go Modules + SemVer: new path, declare API, immutable tags. Ariary MGA 969, 5 iraimbilanja; `GB=1024³` (`metrics.go:74`); operator Go=1024Mo; decimal `1e6` (`cost.go:218`) labelled. No FX (XE 4,525.18 vs doc 4,335 stale).
- **Gaps:** G1 live drift (100gb, Airtel, Fibre FUPs). G2 global mutable pricing. G3 ×20 vs explicit window. G4 quoted 150 ranked primary for 4g/5g. G5 primary map is guess. G6 French strings API. G7 no provenance fields. G8 mixed-bundle contamination. G9 bare `0` indistinguishable from no-waste.
- **Target:** `Tariff{Name,ARPerGB,Quoted,Source,AsOf,Model}`, `Plan{Name,MonthlyOpExMGA,CIRMbps,FUPBytes,Source,AsOf}`; waste → `(value,measured)`; `CAPOperationalTarget`; `DefaultMGATariffs(asOf)` (fix 100gb→250Go/800, add 30000/75000, exclude Airtel/Orange); CLI prints `Economics` JSON; optional `RenderText`.
- **Slices:** S1 Freeze+provenance types. S2 De-globalise+delete ×20. S3 Drift repair+`check-tariffs.sh` (re-fetch R1–R6, fail on drift, 404/JS→`unverified`). S4 Measured discipline (primary caller-declared). S5 Tag+consume (wire shape preserved).
- **Anti-toy:** 1. Hand arithmetic (81.38, 5,556, GOLD hourly, ×20@180s). 2. Provenance on every number (`unverified` exact phrase when failing). 3. No silent zero. 4. No global mutable. 5. Real consumer. 6. Version discipline (silent tag edit = toy).

### 10-T04 cgo-profile (pkg/profile + model profile types)
- **Responsibility:** Registry + persistence: `Profile` value (`CapUp/CapDown`), built-ins P1–P4, global `Profiles` + `ProfilesMu` + `BuiltinProfileIDs`, ops `Reserved/Import/Load/Delete` via `data/profiles.json`. HTTP: `GET /api/profiles` sorted `imported` (`server.go:531-559`, `api.md:42`), `GET /api/profile/list` (`:1252-1263`, `api.md:43`), `POST /api/profile/import` CSV positional or JSON (`:1158-1230`, `api.md:44`), `POST /api/profile/delete` (`:1233-1251`). Negative: no shaping/probing/matrix-resolve/match.
- **Exported API:** `Profile/CapDown/CapUp/Profiles/ProfilesMu/BuiltinProfileIDs` (§10-T01, `types.go:11-68`); `Reserved(id)` (`profile.go:18-20`); `Import(p)` (`:23-46` empty ID nil silently — library vs HTTP disagree); `Load()` (`:50-67` silent no-op on missing/unparseable); `Delete(id)` (`:71-93`). Tests: P3 literals (`profile_test.go:11-19`), round-trip (`:22-44`), reserved refused/ghost-P3/unknown delete (`reserved_test.go:25-72`), concurrent 8×25 (`profile_concurrent_test.go:15-51`), asymmetry silent cap (`profile_asym_test.go:8-24`), API CSV/extended/P4-not-imported (`profile_import_test.go`, `profile_fullcols_test.go`, `csvmapping_test.go:99-126`).
- **Dependencies:** stdlib + `model` only (`profile.go:9`). Incoming: `pkg/api` (`server.go`), `pkg/campagne/config.go:71-84` (RLock resolve), `pkg/audit/match.go:16-56` (Load+RLock), `matrix.go:179-181`/`watch.go:82-84` (RLock per cell), `calibrate.go:37-39`, `server.go:287-288` shaping, frontend `profiles.ts:3-12` + `BenchConfig.tsx:90-94`.
- **Primary-source requirements:** Go Modules (§4.1: own `go.mod`, path=prefix, immutable; `replace` main-module only); SemVer (§4.2: declare API before 1.0.0; MAJOR/MINOR/PATCH; deprecate one cycle — governs `LossBurst*` removal + `Profiles` shim); RFC 8290 negative (§4.3: no link-profile registry in RFC — `DelayMs` is netem/regime param, not CoDel `target` sojourn; docs must state distinction, not borrow RFC language); Prometheus labels (§4.4: `[a-zA-Z_][a-zA-Z0-9_]*`, `__` reserved, value change = new series, empty = absent — unbounded IDs = cardinality hazard; `CapacityUpMbps=0` as empty label vanishes); honesty §4.5 (R2 rule 3 immutable; reorder JSON/CSV = MAJOR).
- **Gaps:** Silent failures (`Import` empty-nil vs API-400; `CapUp` silent cap; CSV errors→0.0 `:1181`; unreadable JSON silent; no `Validate()`; HTTP bounds 1–1000/0–600/0–100/0–10 `api.md:51` only at shape edge). `Load` write on every GET (`server.go:532,1253`, `match.go:20`) + lock across disk IO. Reserved duplicated (literals `:43-51` vs predicate `:19` — ghost-P3 incident). CWD-relative unversioned (`:12`), no `schema_version`, positional CSV sniffing (`row[0]=="id"`). `LossBurst*` zombie exported. No bounded Prometheus label policy.
- **Target:** `Registry` (`Get/Set/List/Reserved/BuiltinIDs`, `Default()`) + `Validate` (no silent-nil; promote HTTP bounds) + `Store` injection (`FileStore(path)`, missing=empty honest, corrupt=error) + `ParseCSV` header-based name-matched + `ImportResult` per-line errors + `Export` byte-stable + CLI `list|import|export|delete|validate` + bounded `LabelValues()` (`P1..P4` or `builtin/imported`, lint regex). Deprecate `Profiles` shim one minor cycle. Schema `schema_version:1` + migration note.
- **Slices:** PR-1 Freeze+document. PR-2 Extract library (`Registry` + shim, `Validate`, `Store`, moved tests, `Reserved` one authority). PR-3 Store+CLI round-trip (`import→export→import` byte-stable; bad line exits non-zero; version field). PR-4 Re-point CGO (`make verify` green, headers byte-identical, chdir tests removed). PR-5 Harden (bounded helper, `LegacyBurst` path, example, retraction policy).
- **Anti-toy:** A1–A10 (see spec appendix): declared API; independent SemVer; immutable; own `go.mod` stdlib-only; no raw global; validation at boundary; normative cited `unverified` labelled; bounded exposure; per-module CI; license + `schema_version:1`. Failure A1–A5 = toy.
- **DEFERRAL:** Module path + license + cap-policy (`up>down` silent-cap vs error) are human decisions before PR-2 (affects API shape + CLI exits). CAKE/PFIFO/netem normative pages + P1–P4 value provenance stay `unverified` (docs only).

### 10-T05 cgo-results (pkg/results)
- **Responsibility:** Frozen-corpus aggregation core (`docs/architecture.md:17`, `results.go:49-65`): aggregate every `aqm_eval.csv` under data dir, group by `profile|qdisc|cc|direction`, compute medians. Five things: header-name CSV tolerating 17/18/19-col historic layouts (`csvrows.go:9-45`, never positional — `qdi_ms` insertion bug `:11-15`); cell aggregation quarantine honesty (`valid:208,degraded:74,quarantined:246`, `docs/VERIFY.md:51`); provenance `Group.Runs` only runs with ≥1 non-invalid row (`:43-46,146-148`); best-per-profile + hardware string (`:207-234`, `hardware.go:7-12`); `SeedRuns` anti-traversal extract (`seed.go:15-79`). No shaping/probing/HTTP/exposition.
- **Exported API:** `ReadAQM` (`FieldsPerRecord=-1` `:19-45`, keep short rows); `ColIndex/Field` (`:49-65`); `Group` (`:13-47` cell, `Count/Quarantined/Degraded`, all-rows medians+IQRs, valid-only `Smallp95*/Deadline*` with N/median/IQR/CI95, `Best`, `HardwareRecommendation`, `Runs` sorted); `Scan(dataDir,runFilter)` (`:65-236` glob `*/aqm_eval.csv`, skip `run-smoke*` `:88-90`, silent-skip unreadable/missing-header `:93-100`, direction default `up`, up-before-down `:188-206`, `Best` min `Smallp95Median` `>0` `:208-220`); `HardwareRecommendation` (`hardware.go:7-12` French MikroTik, `unverified`); `SeedRuns` (`seed.go:24-79` embed `runs-frozen.tar.gz` `:15-16`, O_EXCL, reject `..`). Tests: `TestScanReadsByName` (`results_test.go:70-121`), `TestScanMixedLayouts` (`:126-160`), `TestGroupRunsExcludeInvalidOnlyRuns` (`:33-54`), `TestValidOnlyFields`, direction split/default/order-stable 30-iter (`direction_test.go:9-78`, `direction_order_test.go:13-33`), `TestHardwareSingle` (`hardware_test.go:9-22`), `TestSeedRunsExtracts/NoOp/RejectsEvil` (`seed_test.go:12-96`).
- **Dependencies:** `results.go` stdlib + `pkg/metrics` (`Summarize/Quartiles/BootstrapMedianCI95/Median`, `:166-183`); `csvrows.go` stdlib; `hardware.go` `fmt`; `seed.go` stdlib. No `syscall`. Incoming 7× per request (`server.go:770,942,982,1004`, `duel.go:77`, `translate.go:25`) + `Generate` (`:1084-1090`) + CLI `cgo figures` (`main.go:123`).
- **Primary-source requirements:** RFC 8290 Experimental; MUST only packet-limit (§5.2.3); defaults 100ms/5ms/1514/1024. CAKE no RFC (informative §7/§10.2) — never cite for CAKE. RFC defines no CSV schema/cell/median/Best/SVG/SHA. Prometheus text: TYPE before first sample, unique metric+labels, `+Inf` bucket. Go embed: `[]byte` package-scope (`seed.go:15`). SemVer: declare API, `vX.Y.Z` tags not SemVer strings, tarball unhashed violates §3 immutability.
- **Gaps:** G1 chain `metrics→results→figures` blocks one-module extraction. G2 No exposition (JSON/SVG only). G3 `ProvenanceSHA` hashes last-file only (`figures.go:41-53`) — mutating other 158 does not move it. G4 SVG embeds `time.Now()` + local path — not reproducible. G5 `SeedRuns` unversioned/unhashed snapshot coupled to lib. G6 `HardwareRecommendation` inside `Scan` every group, French, `unverified` — must leave core. G7 Silent skip (`nil,nil` empty, `continue` unreadable/missing-header) — empty indistinguishable from corrupt. G8 `Best` on all-rows (incl. degraded) vs valid-only decision. G9 Chart math no guards/scales. G10 Recompute `Bootstrap×2/group` every `Scan`, no cache.
- **Target:** `Row{Profile,Qdisc,CC,Direction,Gate,Values}`, `File{Name,Header,Rows}`, `Scan([]File) ([]Group,[]Skipped,error)` (skipped named); `Group{Summary,Valid Summary,Runs}` delete `Best`+`HardwareRecommendation` (adapter advisory `unverified`); `Provenance{SHA256,Hash8,Files,At caller-supplied}` corpus manifest hash; `RenderBar/RenderScatter([]Group,Opts)` value-in; `RenderText` NEW per R2; `SeedRuns` → app `cmd` with version+SHA pin.
- **Slices:** S1 Freeze+declare. S2 Break metrics chain (vendor stats, explicit file list, diagnostics). S3 Fix provenance+reproducibility. S4 Decoupled adapters (seed/hardware out, `Best` decided). S5 Renderer+exposition (R2 goldens). S6 Tag `v0.1.0`, migrate, frozen counts still 159/528/208/74/246.
- **Anti-toy:** 1. Reproduce frozen corpus 159/528/208/74/246/22 (`VERIFY.md:24`) + integrity shape (`:51`). 2. Hand vectors (`Smallp95Median==328.2` not QDI `57.4`, `Goodput==18.2` not deadline `60.2`) + direction/order tests ported unchanged. 3. Provenance honesty (corpus vs last-file, CAKE/MikroTik `unverified`). 4. No silent skip/zero. 5. Byte-stable figures + real R2 scrape. 6. Version discipline (no silent tarball swap).
- **DEFERRAL:** CAKE/MikroTik/statistics primaries stay `unverified` (docs only). `book/` never touched.

### 10-T06 cgo-figures (pkg/figures)
- **Responsibility:** SVG renderer + provenance hasher (`docs/architecture.md:18`, `figures.go:15-37`): walks runs via `results.Scan`, emits `small_p95.svg` (bars) + `scatter.svg` (goodput×small) fixed 800×400 with RDF metadata. No shaping/probing/HTTP/exposition.
- **Exported API:** `Generate(dataDir,outDir)` (`:17-37`, `no data` on empty); `ProvenanceSHA(dataDir)` (`:41-53`, lexically-last CSV only); `ProvenanceHash8` (`:56-61`, first 8 hex). Unexported `provenanceMeta` (`:63-79`), `writeBar` (`:81-116`), `writeScatter` (`:118-156`). Tests: `TestFiguresRDF` (`:16-51`), `TestProvenanceHash8` (`:55-73`).
- **Dependencies:** stdlib + `pkg/results`. No third-party. Incoming: `POST /api/figures/regen` (`server.go:1084-1090`), CLI `cgo figures` (`main.go:123`), `/api/integrity`/report/compare-cache (`:896,975,1022,1079`).
- **Primary-source requirements:** Same RFC 8290 / Prometheus / Go Modules / SemVer family as T05 + `embed` rules (package-scope `[]byte`, relative patterns). SVG reproducibility has no RFC: caller-supplied timestamp/source is the contract.
- **Gaps:** G1–G10 as T05 G1–G10 viewed from figures: drags `results`+`metrics`; last-file-only hash; `time.Now()`+path leak; zero-width bars no guard; no tick axes; overlap labels; hardcoded fonts/colors (`#070707`, Cormorant, JetBrains Mono, `:97,108`).
- **Target:** `RenderBar/RenderScatter([]Group,Opts) ([]byte,error)` data-driven size, tick axes; `Generate` deleted from core (adapter only). No `embed`, no hardware string.
- **Slices:** See T05 S3–S5 (provenance, adapters, renderer+exposition). Slice acceptance: mutating any non-last file changes corpus digest; two `Generate` byte-identical; exposition goldens vs R2 parser; chart pixel pins for 2-group fixture.
- **Anti-toy:** Same six as T05 plus: byte-identical two renders, at least one out-of-tree importer, one R2 scrape succeeds.
- **DEFERRAL:** MikroTik/CAKE/statistics primaries `unverified` stay labelled.

### 10-T07 cgo-qdisc (pkg/qdisc)
- **Responsibility:** Bench `tc` execution behind runner seam (`qdisc.go:1-13`): `TCRunner{Run(args)}` (`:23-25`), `ExecRunner` `sudo tc` (`:27-32`), `NsRunner{Ns,Env}` `ip netns exec` (`:37-54`), `FakeRunner` (`:57-62`). Two-hop topology `veth client ← netem · veth shaping+aqm → veth server` (classless at root only). Egress `ApplyShaper` (`:88-143`): CAKE root/child `10:` fallback `1:`, fq_codel child `20:` of `tbf 10:` with RFC defaults, pfifo_fast = TBF alone (magic default R7). Impairment `ApplyNetem/Physics/GE/Distribution` (`ge.go:41-92`). Ingress `ApplyIngressShaper` + read-only `CheckIngress` (`ingress.go:28-88`). Stats `PollStats` `tc -s` (`stats.go:22-28`). Calibration `Build/Write/Load/Render` (`dist.go:73-303`, fail-closed sha256). No remediation/probing/HTTP/CSV.
- **Exported API:** `TCRunner/ExecRunner/NsRunner/NsCommand/FakeRunner` (`qdisc.go:23-62`, `ingress.go:13-23`); `Reset(r,iface)` (`:67-74`); `ApplyNetem(r,iface,delay,jitter,loss)` (`:79-81`); `ApplyShaper(r,iface,q,cap,rtt)` (`:88-143`); `GEModel/GEMeanLoss/ApplyNetemGE/Physics/Distribution` (`ge.go:19-73`); `Calibration/BuildCalibration/Render/WriteCalibration/LoadCalibration` (`dist.go:33-303`); `Stats/PollStats/SumDrops/SumBytes` (`stats.go:10-105`); `IngressRunner/NsCommand/ApplyIngressShaper/CheckIngress` (`ingress.go:13-88`).
- **Dependencies:** `pkg/model` only (`Qdisc` `qdisc.go:20`/`ingress.go:8`, `Profile` `dist.go:17`), stdlib (`os/exec,fmt,strings`). Incoming: `campagne` (`campagne.go:360-368,391,415,428,495,596-597`, `dualns.go:46-173`, `prod.go:41-80`, `watch.go:34,77`). Kernel `tc` + `cake,fq_codel,tbf,netem,pfifo` qdiscs, IFB + `act_mirred/cls_matchall`.
- **Primary-source requirements:** RFC 8290 §5.2.1–5.2.5 defaults (100ms/5ms/MUST limit/1514/1024) + `tc-fq_codel(8)` mirrors. `tc-cake(8)` (bandwidth, rtt internet 100ms, overhead [-64,256], `diffserv4/triple-isolate/dual-dsthost/nat/ack-filter/wash/ingress`). `tc(8)` (egress shaping, ingress policing, `replace` atomic, classless at root, `-s` more stats). `tc-netem(8)` (delay/jitter, gemodel 4-arg, uniform/normal/pareto). `tc-tbf(8)` (pure shaper, burst bucket, limit/latency exclusive). `tc-pfifo_fast(8)` (automatic default, 3-band TOS, no stats, invisible in `ls`). Go Modules/SemVer: local `Kind` instead of `model.Qdisc`; `v0.1.0`, minor compatible-add, major `/v2`.
- **Gaps:** G1 fq_codel params pinned bench (`:126`) defaulted production — implicit agreement. G2 CAKE knobs differ bench (bandwidth+rtt) vs production (+ack-filter/nat/diffserv4/isolate). G3 TBF burst fixed 256kbit vs MTU-bytes — neither sizes from rate. G4 "pfifo_fast" three meanings (bench TBF+default, production TBF+pfifo 1000, API label) none is R7 fifo. G5 Stats text scrape fragile, `SumBytes` leaf heuristic fails exotic, pfifo invisible ⇒ 0 indistinguishable. G6 RouterOS `unverified` (see T08). G7 `.dist` format `unverified` (custom name not in `tc-netem` delay list; `Load` counts lines not values). G8 Dead code (`fqCoDelString` discarded, `execCmdRunner` in non-test). G9 `Reset` English-substring tolerance. G10 `CheckIngress` names not effect.
- **Target:** `module .../cgo/qdisc`: local `Kind` (`cake/fq_codel/tbf-fifo` — no pfifo label), `ShaperParams{Kind,Capacity,RTT,CAKE,FqCodel,TBF Burst derived}`, `NetemParams`, `PollStats` parses `tc -j` (text fallback flag), `SumLeafBytes` renamed, `Calibration` parse+range-check. Two modules (`qdisc` bench exec vs `remediate` pure gen) — owners/gates differ.
- **Slices:** S1 Decouple `model` (local `Kind`+`Profile` fields; `grep pkg/model pkg/qdisc` empty). S2 One `FqCodelParams` shared (R1 five values asserted). S3 `tc -j` parser + pfifo-invisible explicit error. S4 Burst from rate both packages (`burst>=rate/HZ` at 5/20/80/1000). S5 CAKE knobs unified or delta named. S6 RouterOS verify-or-fence (see T08). S7 Calibration parse+range. S8 Module split `v0.1.0`.
- **Anti-toy:** 1. Real `tc` on Linux VM (recorded `tc -s qdisc show`; pfifo-invisibility included; `Reset` vs real error text). 2. RFC 8290 five params asserted in both packages. 3. CAKE equivalence proven or delta named. 4. Stats never lie (explicit empty, leaf-only goodput). 5. Nothing unmeasured emitted (`Validate`, `<30` samples, fail-closed load, 404). 6. RouterOS verified or fenced `unverified`. 7. Version numbers mean what they say (tc-line change = minor at least).
- **DEFERRAL:** RouterOS syntax + `.dist` custom name + GE mean-loss formula + `ffff:` handle stay `unverified` until kernel source / real device fetch confirms them (docs-only fences).

### 10-T08 cgo-remediate (pkg/remediate) + cgo-chr-validate
- **Responsibility:** `remediate.go:1-8`: remediation only from measured numbers. `Link{Interface,RouterOSInterface,RouterOSLanInterface,CapacityMbps,RTTms,Overhead,MTU}` (`:22-39`); `Validate` refuses empty/invalid/unmeasured with `ErrUnmeasured` (`:77-101`); rate = `Capacity×0.95` (`:68-74` `unverified`). Four pure generators: `LinuxTC` egress (`:135-194`), `LinuxTCIngress` via IFB (`:215-260`), `RouterOS` upload (`:268-322`), `RouterOSIngress` download (`:330-384`, disjoint `cgoingress-*`). `Recommend` (`:400-417`) one-line policy `unverified`. `cmd/cgo-chr` auxiliary QEMU/serial orchestrator (`main.go:1-5`): flags `--report/--image/--level/--timeout/--out/--qemu/--export-file/--dry-run` (`:42-50`), `BuildQEMUArgs` (`:23-39`), `run` (`run.go:19-107`), `drive/expect` serial (`serial.go:23-121`), `Transport/qemuTransport` (`transport.go:8-23`). `pkg/audit` RouterOS subset `ParseRouterOSConfig/CompareRouterOS/WriteCHRProof` (`routeros_config.go:9-85`, `routeros_compare.go:11-64`, `routeros_proof.go:12-76`, schema `cgo.chr-proof/1`). `kit/chr/` wrapper + fixtures + empty `chr-image.sha256` placeholder.
- **Exported API:** `Link,QdiscCake/FqCodel/Pfifo,IngressIFB="ifb0",DefaultRouterOSLan="bridge",RouterOSLan(),ShapedRate,ErrUnmeasured,Validate,LinuxTC/LinuxTCIngress/RouterOS/RouterOSIngress,Recommendation/Recommend` (`remediate.go:22-417`). `RunConfig{QEMU,Image,Level,Timeout}` (`cgo-chr/main.go:14-19`), `BuildQEMUArgs` (`:23-39`), `Transport/qemuTransport` (`transport.go:8-23`). `RouterOSObject/Config/Get/ParseRouterOSConfig`, `RouterOSCheck/CompareRouterOS`, `CHRInfo/ImportResult/CHRProof/FileSHA256/WriteCHRProof` (`routeros_*.go`). `Remediation{RouterOS,RouterOSIngress}` (`report.go`).
- **Dependencies:** `pkg/remediate` zero internal (stdlib `fmt,math,strings`). `cmd/cgo-chr` only `pkg/audit` (`run.go:11`) + stdlib. `pkg/audit` CHR subset only needs parser/diff/proof types, not full audit. Runtime: QEMU `qemu-system-x86_64` (`main.go:47`), CHR image licensed never committed (`README.md:12-14`), RouterOS console `login:/[admin@…] >` `/export terse`.
- **Primary-source requirements:** MikroTik CAKE (R1: bandwidth, RTT, Diffserv, Flow Mode, ACK Filter, NAT, Wash; `triple-isolate` default, `dual-dsthost` ingress, `diffserv4` 4 tins), Queues (R2: `/queue type`+`/queue tree`, `max-limit`, defaults `cake-rtt 100ms`,`cake-flowmode triple-isolate`,`cake-diffserv diffserv3`,`cake-ack-filter none`,`cake-nat no`,`cake-wash no`; FQ-Codel/CAKE since 7.1beta3 with `fq-codel-interval 100ms/target 5ms/flows 1024/limit 10240/quantum 1514`), Queue types (R3: BFIFO/PFIFO/CAKE/CoDel/FQ_CoDel/MQ_PFIFO/RED/SFQ), RFC 8290 §5.2 defaults. RouterOS `kind=fq-codel` spelling, `/queue` grammar, CHR image/license/console: `unverified — primary source not found` (404/Anubis). `cake-ack-filter=filter`, `cake-wash=yes`, `kind=sfq sfq-perturb=10`, `kind=fifo` bare, 95% headroom, `Recommend` thresholds, ≥7.10 gate: `unverified`.
- **Gaps:** G1 "No fq_codel, emit sfq" (`:300-310`) contradicts R2/R3 (FQ-Codel exists since 7.1beta3) — emits weaker SFQ while claiming honesty. G2 `TestRouterOSEmitsOnlyValidKinds` (`remediate_test.go:249-297`) pins `cake|sfq|fifo`, forbids correct `fq-codel`. G3 ≥7.10 header vs R2 7.1beta3. G4 `linux-namespace` label only (tcg→kvm, no attach). G5 `imageHash` `os.ReadFile` whole image + pin file never checked. G6 `Validate` forbids negative overhead R1/R2 allow (-64…256); RouterOS ignores overhead. G7 Serial framing only `fakeCHR`-pinned, no real transcript. G8 Upload+ingress concat interleaves cleanup (safe only by disjoint names). G9 `fakeCHR` `/export` shape vs real `terse` `unverified`. G10 CHR BLOCKED (`VERIFY.md:57`, 12/12 replay only, no QEMU/image). G11 Mutation test missing named-failure assert. G12 Two "default LAN" provenances collide (`"bridge"` never observed, `VERIFY.md:64`).
- **Target:** `cgo-chr-validate` (stdlib-only): `Report/Remediation` JSON shapes, `ParseRouterOSConfig/CompareRouterOS/WriteCHRProof` unchanged, `BuildQEMUArgs/Run/RunConfig`, generation oracle `RouterOS/RouterOSIngress` vendored. `Level` stays `windows-tcg|linux-namespace` until G4 implemented; refuses `linux-namespace` proof without attach. Layout `cmd/cgo-chr-validate`, `pkg/chrvalidate`, `pkg/remediateoracle`, `kit/chr/` repointed, `docs/VERIFY-IMPORT.md` quotes BLOCKED row.
- **Slices:** S1 Vendor-truth fix in monolith first (emit `kind=fq-codel`, admit regex, `cake-overhead` negative, header cites R2). S2 Strengthen tests (`TestRunReplayMutatedExportFails` named assert; streaming hash + pin). S3 Extract repo (no `fyne/bubbletea/lipgloss/x-sys` via `go list -m all`; replay+dry-run green). S4 Real `windows-tcg` acceptance (human installs QEMU+CHR, `proof.json` `cgo.chr-proof/1` all-pass, row BLOCKED→VERIFIED with command+host+date). S5 `linux-namespace` implement-or-rename. S6 Tag `v0.1.0`, notes quote VERIFY row.
- **Anti-toy:** 1. Real CHR boot with transcript + matching SHAs, not `fakeCHR`. 2. Vendor truth R1/R2/R3 for every `cake-*`/`fq-codel-*`; `unverified` listed in README table. 3. Named failure (mutated export fails one check, duplicate fails, refused seals `Accepted=false` verbatim). 4. Image = image (streaming SHA vs pin before boot; absent QEMU exit 2). 5. Levels mean what they say (no débit/latence/prod change). 6. Nothing unmeasured validated (refuse no script exit 2; regenerate byte-identical; echo `routeros_lan_assumed`). 7. Independence real (`go list -m all` stdlib only, builds without monolith). 8. Version numbers honest (emitted-string change = minor at least; quote BLOCKED row).
- **DEFERRAL:** Real CHR boot DEFERRED until human installs QEMU + downloads CHR + fills `chr-image.sha256` (`kit/chr/README.md:10-14`). `linux-namespace` level DEFERRED until attach implemented or renamed. RouterOS/CHR/OpenWrt pages 404/Anubis stay `unverified` until fetch succeeds.

### 10-T09 cgo-probe (pkg/probe sans bpf loader)
- **Responsibility:** Active measurement primitives (`probe.go:1-4`): ICMP ping exec + `time=` parse, HTTP small-object via shared/fresh clients, TCP bulk flood up/down N-flow, per-socket CC + TCP_INFO truth, kernel-tier ladder (eBPF vs TCP_INFO). No shaping/percentiles/CSV/report.
- **Exported API:** `PingSample`, `Ping(ctx,target,count,intervalMs)` (`probe.go:22,28-43`); `SmallClient`, `SmallClientTimeout`, `SmallObject` (`:93-139`); `BulkSendTo/BulkSend/BulkReceive`, `BulkSource/BulkRecvSource/BulkDownloadTo/NTo,ParseDownloadHello` (`bulk.go:11-75`, `bulksource.go:14-103`); `FlowResult,BulkSendNFlows,NTo`, `BulkStats,BulkSendNStats/DownloadNStats` (`bulkflows.go:14-64`, `bulkstats.go:11-73`); `DialWithCC/SetConnCC` (linux + `!linux` no-op) (`bulk_cc_linux.go:14`, `ccsock_linux.go:15`); `TCPRetransBytes`, `ErrTCPInfoUnavailable` (`tcpinfo_linux.go:20`, `tcpinfo_other.go:12`); `KernelSourceInfo,KernelSource` (`kernelsource.go:12,34`).
`Ping` `-c/-i` Unix, `-n` Win, `count*interval+5s`, `len(out)==0` only error; `<1ms→v/2` (`:81-83`); sort discards order. `SmallObject` GET→`io.Discard`, DSCP EF `IP_TOS=46<<2` (`dscp_linux.go:12-30`; RFC 3246 §2.7 EF 101110). `BulkSend` 64KiB zeros, 2s deadline, `Write`-error→end. `DialWithCC` `TCP_CONGESTION=0x0d` string (R8 per-socket). `TCPRetransBytes` `SYS_GETSOCKOPT(TCP_INFO)` → `Total_retrans×1448` (`:15,43`).
- **Dependencies:** Stdlib only + `pkg/model.Cubic` (`bulkflows.go:44`) + `pkg/probe/bpf` ladder; `tcpinfo` rejects `x/sys` for raw syscall (`:30-35`). Axes `linux|!linux`, `ebpf_object`. `go.mod:5-10` none needed.
- **Primary-source requirements:** RFC 8290 §§5.2 framing only (CAKE no RFC). tc(8) family: shaping egress, CAKE `ingress` counts dropped bytes (R4), fq_codel defaults `limit 10240/flows 1024/target 5/interval 100/quantum 1514` (R5), netem impairs outgoing (R6), TBF pure shaper (R7). TCP (R8): `TCP_CONGESTION` per-socket, `TCP_INFO` `struct tcp_info` from `/usr/include/linux/tcp.h` not portable. BTF (R9) `/sys/kernel/btf/vmlinux`, libbpf Open→Load→Attach (R10), verifier NULL-checks (R11), tracepoints (R12), caps `CAP_BPF` since 5.8 / `CAP_NET_ADMIN` TOS (R13). Waveform bands, JFI, G.107, RouterOS, `ping(8)` output, `Total_retrans×1448`, 220/95: `unverified — primary source not found` (local framing only `cost.go:75-83`).
- **Gaps:** G1 eBPF object-less + loader-less (`BUILD.md:8-12`), no Open/Load/Attach. G2 `audit.go` 865-line god-harness (4 other packages). G3 Bench IDs hardcoded (`sinks.go:41-48`, `presets.go:18-39`). G4 Global registry read in measurement (`match.go:20-22`). G5 Fidelity: ping locale/heuristic, 3 small timeouts, zero-fill goodput + swallowed writes + skipped dials + `RetransMeasured=any`, ×1448 vs `skb->len` mismatch. G6 Invisible tunables (`Flows=4`, `Direction ""→up`, `Duration≤0→10s`, BBR default, `pingPerCall=5`, 300ms/5s, mini-drain 5s+3s). G7 Loss denominator assumes fake contract (`pingCalls×5`). G8 Throughput fallback reintroduces invented duration. G9 `BuildReport` mixes economics/remediation. G10 CSV self-migration, no schema version (plus `timeline.points()` `max+1` alloc, `SealSource` whole-CSV memory, `kernelFacts` stat/euid overclaim).
- **Target:** Probe `github.com/<org>/cgo-probe`: `PingResult{Attempts,Received,Samples}`, `PingRunner` seam, `SmallClient{Timeout,MarkEF}`, `BulkOptions{CC,Flows}`, `BulkStats{PerFlow,RetransBytes,Measured,Proxy}`, `TCPRetransBytes` (`ErrNotTCP/ErrNotLinux`), `KernelTier` (Open/Summary), `KernelSource` real caps/BTF. Audit `github.com/<org>/cgo-audit`: `Config{Duration,Flows,Direction,Target,SmallURL,BulkAddr,Ping/Small/Load funcs,OnProgress,Kernel}`, `PingFunc` returns `Attempts`, `Engine{Stats}`, `Run`, `BuildReport(res,SourceSeal)` (no `remediate.Link`), `LinkAuditSchema="cgo.link-audit/1"`, `AppendLinkAudit(w,schema,res)`. Probe stdlib+loader only; audit never imports `probe/bpf` directly.
- **Slices:** S1 Freeze+declare (port tests incl. honesty suite). S2 De-globalise+honest sockets (delete `model.Cubic`, `MarkEF`, `Attempts`, drop throughput fallback). S3 Kernel truth (build object, loader, real BTF/capget). S4 De-bench+de-globalise (inject sinks/presets/profile, split `BuildReport`, version schema). S5 Tag `v0.1.0` both.
- **Anti-toy:** 1. Live-bench proof per family (dualns + `tc` impairment cited). 2. Kernel-tier proof or honest fallback with named missing fact. 3. Honesty invariants as unmodified tests (no-load⇒no-grade, <2 flows⇒`JFIMeasured=false`, etc.). 4. Provenance honesty (bench IPs removed; DSCP/fq_codel cited; rest `unverified`). 5. No silent zero/self-migration. 6. Real consumer + `tc` remediation `CompareRouterOS`. 7. Version discipline.
- **DEFERRAL:** Live bench runs DEFERRED until human approves 6.8 host. Kernel-truth (object+loader) DEFERRED per §8.12. Bench IPs/presets/tariffs stay `unverified`-labelled deploy constants until injected as caller data.

### 10-T10 cgo-bpf-observer (pkg/probe/bpf) — deep section §8
- **Responsibility:** CO-RE trends sidecar: object `bpf_probe.o` (`bpf_probe.c:1`) + Go contract types/status/hash. Fallback ladder `probe.KernelSource` (`kernelsource.go:34`) chooses `ebpf|tcpinfo`. Does NOT own topology/matrix/metrics/report/qdisc/DSCP beyond small-object bit.
- **Exported API:** `bpf.Flow` (`types.go:4-16`); `bpf.Summary` (`:20-35`); `bpf.Globals` (`:40-47`); `bpf.Status(goos,btf,caps,objectErr)` (`status.go:11-25`); `bpf.ObjectSHA256/ObjectBytes` (`:28`, `assets_embedded.go:10`, `assets_absent.go:7`); `bpf.UnavailableError/BuildError` (`errors.go:4-11`); Programs `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`, `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb`; maps `flows HASH 4096` + `globals PERCPU_ARRAY 1` (`bpf_probe.c:60,112,176,182,188`); Missing on disk (must implement): `bpf.Start`, `Lookup/Snapshot/Globals/Attach/Close`, `Check`, `KernelTelemetry` alias, `BulkStats.Source/Kernel` (spec `2026-10-27-cgo-ebpf-ground-truth-design.md:393-537`).
- **Dependencies:** Module `github.com/Realms4239/cgo` `go 1.25`; NO `cilium/ebpf` today; `pkg/probe` stdlib + `pkg/model`; `pkg/probe/bpf` stdlib (`sha256,hex,embed`) only, zero `bpf(2)` calls. Build axes `linux|!linux`, `ebpf_object`. Toolchain: Ubuntu 24.04, kernel 6.8.0, clang-18, libbpf-dev, bpftool (`BUILD.md:16`, `Makefile:46-56`); runtime no clang (CO-RE). Lineage: `kernel_adapter.go:6,20` neutral `KernelSample`, `kernel_bridge.go:12` maps Summary.
- **Primary-source requirements:** Full §8.6 (BTF/CO-RE, libbpf/cilium lifecycle, verifier/tracepoint/caps, RFC 8290+tc, Go module hygiene). Tracepoint arg orders, `srtt_us>>3`, `skb->len==payload`, 6.8/7.1 `qdisc_drop` history, 1448 MSS, EF 46, short-struct prefix-fill: `unverified — primary source not found`.
- **Gap analysis:** §8.7 7.1–7.5. Object-less+loader-less (critical). TCP_INFO honest but lossy. Unverified kernel facts as comments. Small-object/bench constants. IPv6/netns/drop limits.
- **Target structure:** §8.9 one repo `github.com/Realms4239/cgo-bpf-observer` `go 1.25` + `cilium/ebpf` pinned; `internal/probe` + `internal/probe/bpf` + `session_linux.go`/`session_stub.go`; public `probe` package only; CLI `capture/tcpinfo/doctor`; Non-goals explicit.
- **Slices:** §8.10 Slices 0–6.
- **Anti-toy:** §8.11 (8 criteria + A1–A8).
- **DEFERRAL:** §8.12 (live-attach/overhead/dualns 4/4 until 6.8 host; 7.1+ `qdisc_drop` until ≥7.1; BBRv3 3×3 + `.dist` ≥30 until 7.1 + field trace; exact-byte default until Slice 1).

### 10-T11 cgo-audit (pkg/audit)
- **Responsibility:** Phased link audit + sealed operator report (`audit.go:199-204`): 60s = 10s quiescent + 40s saturated + 10s drain (`:719-748`). Three-window harness (baseline/direction load 5s mini-drain+3s settle `:414-424`, `collectWindow` 5/call `:309` + 400ms `:344-373`, `runLoadWindow` 300ms/5s `:380-409`). Honesty invariants: zero ping→error, no unmeasured grade (`:562-565`), `JFIMeasured=false` <2/all-zero (`:184-197`), `DrainMeasured=false` uncollected (`:567-607`), `CAP Measured=false`→blank CSV (`:114-122`), telemetry absent→absent line, loss=received vs attempted (`pingCalls×5`), throughput real bytes/real `Elapsed`, UTF-8 gate, cancel→no partial. Raw evidence + double seal (`telemetry.csv` exact header, `Provenance` row+telemetry), reproducibility seal (`Invocation` normalized Params POSIX command), `BuildReport` overwrites capacity from measurement, refuses without RTT+capacity, emits `LinuxTC/RouterOS` via `remediate`, `Unavailable` reasons. Bench discovery (`ResolveSinks` `10.30.0.2:5201`→`10.200.0.1:5201`→`8.8.8.8`), `MatchProfile` 2-axis P1–P4, `Presets` Yas/Madagascar. No `tc` shaping (delegated).
- **Exported API:** `Params` (`audit.go:16-38`), `Result` 30+ fields (`:41-144`), `Deps{Ping,Small,BulkLoad,OnProgress}` (`:146-152`), `Progress` (`:157-169`), `LoadOutcome` (`:175-182`), `Run` (`:204-717`), `Invocation` (`invocation.go:15-34`), `KernelSample/KernelTelemetryFrom` (`kernel_adapter.go:6-35`), `TelemetryRow/ArtifactID/WriteTelemetry/ReadTelemetry` (`telemetry.go:20-117`), `TimelinePoint` (`timeline.go:7-14`), `AppendLinkAudit` (`writer.go:13-84`), `MatchProfile` (`match.go:16-56`), `Preset/Presets` (`presets.go:7-39`), `Sinks/ResolveSinks/EnvOrBulk` (`sinks.go:14-57`), `Report/RouterOSValidation/KernelTelemetry/KernelFlow/Phases/Measured/Unavailable/Provenance/Remediation` (`report.go:25-248`), `BuildReport` (`:255-504`), `SealSource/SealTelemetry/RenderJSON/Markdown/SortUnavailable/WriteArtifacts` (`:508-847`), `DefaultDataDir/ReportArtifactName/.../ArtifactDir/File/Pointer` (`artifacts.go:11-55`), `RemediationState/AvailableState/ObservationState/IsAvailable` (`remediation_state.go:8-39`), `RouterOSObject/Config/Get/ParseRouterOSConfig` (`routeros_config.go:9-85`), `RouterOSCheck/CompareRouterOS` (`routeros_compare.go:11-64`), `CHRInfo/ImportResult/CHRProof/FileSHA256/WriteCHRProof` (`routeros_proof.go:12-76`).
- **Dependencies:** Stdlib + local `pkg/metrics` (Summarize/Grade/JFI/CAP/VoIPR/Economics), `pkg/remediate` (Link/Recommend/ShapedRate/LinuxTC/RouterOS…), `pkg/model` (Profiles/Mu, LinkAuditHeader), `pkg/profile` `Load()` (`match.go:7`), `pkg/probe` defaults (`audit.go:215-268`). Whole module `go 1.25`.
- **Primary-source requirements:** Same family as T09/T10: RFC 8290 framing only; tc(8) R3 shaping egress; CAKE `ingress` R4; fq_codel defaults R5; netem impairs outgoing R6; TBF pure shaper R7; TCP R8; BTF/CO-RE R9/R10; verifier R11; tracepoints R12; caps R13. Waveform/JFI/G.107/RouterOS/`ping`/`×1448`/220-95: `unverified — primary source not found`.
- **Gap analysis:** See T09 G1–G10 (object-less, god-harness, hardcoded bench, global registry, fidelity, tunables, loss contract, throughput fallback, mixed `BuildReport`, unversioned CSV). Plus `timeline.points()` `max+1` alloc, `SealSource` whole-CSV memory, `kernelFacts` stat/euid overclaim.
- **Target module:** See T09 audit shape: `Config` explicit, `Run` normalized, `BuildReport(res,SourceSeal)` only measured/provenance/invocation, `AppendLinkAudit(w,schema,res)` caller-owned file, no `model`/`profile`/`remediate` imports in engine.
- **Slices:** S1 Freeze+declare. S2 De-globalise probe. S3 Kernel truth (see §8.10-2). S4 De-bench+split report+schema version (`grep 10.30|10.200|8.8.8.8 pkg/audit` empty, `go list -deps` no model/profile/remediate). S5 Tag.
- **Anti-toy:** See T09 (7 criteria) — honesty suite unmodified first, then kept green through S2–S4.
- **DEFERRAL:** Live bench runs DEFERRED until human approves 6.8 host. Benchmark/Throughput etc. stay `unverified` until bench run exists.

### 10-T12 cgo-campagne (pkg/campagne, last)
- **Responsibility:** Single-cell orchestration `RunEvent` (`campagne.go:302-703`, physics `dualns.go:46-62`, shaper `campagne.go:336-371`, collect `:439-466`, bulk CC `:499-544`, CPU `:550-582`, G0–G7 `:373-74,472-481,587-657,678-701`, freeze `model.Event`); matrix `startMatrix` (`matrix.go:103-258`, `Window seen` resume, per-cell skip `:43-51,213-250`); live diffusion `Live/Snapshot`+`Deps.Snapshot` (`campagne.go:162-216`) + `StartWatch` surveil (`watch.go:23-100`); persistence `Writer` G5 dedupe + manifest/quarantine SHA (`writer.go:17-200`), `buildMatrixConfig` (`config.go:71-121`), `valuation.json` (`valuation.go:50-95`); host preflight `CapabilitiesNow/DefaultCCs/PlaneReady/BaselineStable(G6)/BurstGE`; production `ProdDeps/Down/DualNS`; policy G6 jitter `max(5,20%·median)+2·JitterMs` (`baseline.go:45,65-72`), G4 direction floor (`campagne.go:649-655`), CPU nil-on-unmeasured. Delegates shaping (`qdisc`), probes (`probe`), percentiles (`metrics`), HTTP (`OnSnap` callback `:97`). Run endpoints `POST /api/run/start|stop|skip`, `GET /api/results|delta|rows|events` (`api.md:59-73`); `cc=bbr3` gated exact-token (`api.md:71-73`, `ccavail_linux.go:12-18`, `matrix.go:138-140`).
- **Exported API:** `Deps` 30+ fields (`campagne.go:20-98`), `Snapshot(phase,load,ev,base,chg,bulk,prof,gates)` (`:104-140` dead base/chg/bulk), `Live/Snapshot` 22 JSON fields (`:162-198`), `NewLive/Set/Get/SetRunning` (`:200-216`), `RunEvent` (`:302-703`); `BaselineStable` (`baseline.go:52-61`); `Capabilities/CapabilitiesNow` (`capabilities.go:8-34`); `DefaultCCs` (`ccavail.go:27-36` `cubic+bbr` always, `bbr3` exact-token); `OnQuarantine/OnSkip` (`matrix.go:14,19`), `Matrix{RunID,Total,Done,Running}` (`:24-35`), `Skip/IsRunning/Stop/StartMatrix/StartMatrixFiltered/WithID` (`:44-267`); `PlaneReady` 6s (`plane.go:19-32`); `ProdDeps/Down` (`prod.go:24-33`), `BurstGE` P2/P3 only (`dualns.go:18-26` `unverified`), `ProdDepsDualNS` (`dualns.go:112-181`); `Writer/OpenRun/Append/Freeze` (`writer.go:17-200` G5 dedupe, 23-col `:145-149`); `ValuationLabels/EventValuation/Sidecar/WriteValuation` (`valuation.go:14-111`); `OnWatchAlert/StartWatch` 10×/5min (`watch.go:15,23`). ~60 `Test*` (G6, bbr3 token, live publish, skip/fail, ingress, G4-down, tier freeze).
- **Dependencies:** `model+metrics+probe+qdisc` + stdlib; no third-party direct (`go.mod:5-10` none imported). Linux `tc`+qdiscs (`modprobe sch_netem/sch_fq_codel/sch_cake` `modprobe_linux.go:10-12`); `/proc/stat`, CC file, osrelease (`unverified` formats); netns topology `veth-c/s`, `ns_wan_emulator/edge_router/client/server`; filesystem `data/runs/<run_id>/{aqm_eval.csv,manifest.json,quarantine.json,valuation.json}` frozen schema (AGENTS.md never touch).
- **Primary-source requirements:** RFC 8290 §§5.2 defaults (CAKE no RFC); Prometheus data model (metric `[a-zA-Z_:]`, label `[a-zA-Z_]`, `__` reserved, empty=absent, float+ms) — campagne emits NO exposition today (`Snapshot` JSON SSE, not text); Go Modules/SemVer (new path, `go` mandatory, `v0.1.0` unstable).
- **Gaps:** G1 `Deps` god-object (30+ fields, `defaults()` silent, dead `Snapshot` params). G2 AQM inherited not asserted (no `FqCodelParams` exposed; CAKE bench vs prod delta). G3 No exposition (`Snapshot` not HELP/TYPE). G4 Package-global hooks `OnQuarantine/OnSkip/OnWatchAlert` (two matrices cross-fire). G5 Diamond hard deps + `ProfilesMu` global read. G6 Linux-only hidden (fail mid-cell not at construction; silent `sudo -n modprobe`). G7 Filesystem/frozen-schema coupling (string concat, `run-<unix>` collision). G8 Heuristics `Wasted=Drops×1448`, tier fallback silent, G4 floor switch, `g<=2500` clamp — all `unverified`. G9 Dead code (`dupes`, dead params, duplicate bbr3 gate). G10 Error strings not typed. G11 `QDIPct=P95−P50` called reference queue-delay (not RFC sojourn), `BurstGE` `unverified`, `jitterAllowance=2.0` `unverified`.
- **Target:** Split `Topology/Probes/Hooks/CellPolicy`; local `Profile/Qdisc/CC/Event` (no global read); `Hooks` per-`Deps`; typed errors (`ErrBurstInvalid/IngressPreflight/ShapeInvalid/CCUnavailable/DuplicateCell/UnsupportedPlatform`); `RunStore` interface (file + in-memory); `FqCodelParams/CAKEParams` explicit; `RenderSnapshotText` per R2.
- **Slices:** S1 Decouple globals. S2 Split `Deps`, drop dead params. S3 Typed errors + fail-closed. S4 `RunStore` + `filepath.Join` + unique runID. S5 AQM assertion. S6 Exposition. S7 Split `v0.1.0`. S8 Verify.
- **Anti-toy:** 1. Real `tc`+netns full cell+matrix (recorded `tc -s`). 2. Five R1 params asserted. 3. Gates mean what they say (G6 thrash fails, `cpuGate` nil never true, heuristics labelled). 4. No global mutable (grep `var On` empty, race clean). 5. Frozen means frozen (Freeze refuses bad JSON, Append refuses dupes, empty axes refused). 6. Real consumer + R2 scrape. 7. Version discipline (tc-line/CSV change = minor+).
- **DEFERRAL:** DualNS/BurstGE bench topologies DEFERRED until human approves bench. No campaign run without approval.

### 10-T13 cgo-api (pkg/api)
- **Responsibility:** REST + SSE surface (stdlib transport, `sse.go:1-2`): 40+ routes (`server.go:380-1432`, `docs/api.md:13-116`), `writeJSON` vs `writeErr` (JSON if Accept/Content-Type mentions it), 1MiB cap (`MaxBytesReader 1<<20`), SSE hub 10Hz (`frameHz=10`, `sse.go:14`) delta+2048-ring+64-cap+`Last-Event-ID`+`event: backpressure` (`sse.go:71-215`), `/metrics` builder (`metrics.go:17-64`, see T16), hardening `requireAuth`+`crossSiteGuard`+`observeBlocked` 501+`validRunID`, audit lifecycle (`server_audit.go:26-499`), `HandleTranslate` (`translate.go:16-125`), `handleDuel` (`duel.go:62-120`), `kernelTelemetryFromSummary` (`kernel_bridge.go:12-30`). No shaping/probing/percentile itself (`Deps.ShapeFn/StartFn/BurstFn`). Composition in `cmd/cgo/server.go:294-337`.
- **Exported API:** `Param` 16 entries (`server.go:92-121`), `SchemaParams()` (`:134`), `ShapeReq` (`:137-143`), `RunOpts` (`:150-170`), `Deps` 14 fields (`:173-216` leaks `audit.Sinks/Deps`), `observeBlocked` (`:220-229`), `RecordEvent` 50-ring (`:233,262-269`), `Handler{http.Handler;hubClose}` (`:237-240`), `CloseHub` (`:243-247`), `New(d Deps) Handler` (`:380-1432` wraps `crossSiteGuard`), `maxSubs=64` (`:314`), `validRunID` (`:320`), `shapeApply` bounds 1–1000/0–600/0–100/0–10 (`:277-312`), `readValidMetric` (`:334-378`), `auditParamsFromRequest` (`:1488-1524`); `Hub/NewHub/Serve/Close` (`sse.go:18-69`); `HandleTranslate` (`translate.go:16-125`); `renderMetrics/handleMetrics/requireAuth/handleDuel/handleAudit*/kernelTelemetryFromSummary` unexported (surface via HTTP).
- **Dependencies:** `server.go:3-30` stdlib + `audit,campagne,figures,metrics,model,probe,profile,results,frontend`; `sse.go:4-12` stdlib; `metrics.go:3-8` stdlib; `duel.go:3-13` stdlib+metrics/results; `server_audit.go:3-24` stdlib+audit/metrics/model/profile/remediate; `kernel_bridge.go:3-6` audit+`probe/bpf`. Globals: `auditMu/lastAudit*` (`:249-256`), `eventsRing` (`:259-260`), `shape*` (`:272-275`), `compareCache` (`:326-330`), `duelDir` (`duel.go:17`). Hardcoded `"data/runs|figures|link_audit.csv"` (`:335,766,794,925,982,1004,1092,1110,1286,1331,1363`, `duel.go:17`, `server_audit.go:184,393,437`).
- **Primary-source requirements:** Prometheus text (`text/plain; version=0.0.4`, HELP/TYPE one-per-metric before first sample, `counter/gauge` vocabulary, EBNF, LF, grouping, `ParseFloat`+ms); Go Modules (released/versioned/distributed together, path=ID, `/vN`, `go` mandatory); SemVer (§1 MUST declare, §3 immutable, §6–8 PATCH/MINOR/MAJOR); `proc_pid_net(5)` (16-field header, per-netns via `/proc/self/net`); IP Sysctl CC STRING; `ethtool(8)` `-k` query / `-K` tso/gso/gro; WHATWG SSE (`text/event-stream`, `data/event/id/retry/`:, blank-line dispatch, `Last-Event-ID`, `retry` digits-only); Go `net/http` (`"METHOD /path"`, `MaxBytesReader`, `Flusher`). tc/ip-netns/caps/BBR/Sec-Fetch/RFC6750/RouterOS/SNMP: `unverified`.
- **Gaps:** G1 `/metrics` static-zero facade (BLOCKER). G2 Shared globals + CWD paths (BLOCKER — two `New` race). G3 Hub exportedness mismatch (`Publish/SSE` unexported, `Hub` uncompletable). G4 Untyped `any` snapshots + `Deps` leakage + stringly `Mode`. G5 Unbounded exec + auth edges (no ctx, `HasPrefix 127.0.0.1` misses `::1`, `==` not constant-time, env leaks, `Print()` hardcoded auto). G6 Minor: `writeJSON` partial 200, silent `results.Scan` ignore, `validRunID` dot rule, 422 vs 404, `Sample` JSON bloat, no escaping, path collision.
- **Target:** `cgo-api` + `api/sse` + `api/metrics`: `Snapshot` typed + `SnapshotProvider`, `Engine`, `DoctorSource`, `AuditSink`, `Store` (kills `data/…`), `ServerOpts{Engine,Store,Doctor,Audit,Snapshot,Mode,Version,FS,Auth,Metrics}`, `New(opts) *Server` per-instance (no globals), exported `sse.NewHub/Publish/SSE/Close`, `metrics.Registry.Observe` (no zeros), `Mode` enum + `ParseMode`, `MaxBytes` opt. Version: minor additive, major+`/v2` if removed/renamed.
- **Slices:** S1 `iface` module. S2 `doctor` decouple (absorb `ParseCCList`). S3 `sse` subpackage (external two-client resync test). S4 Transport core (`*Server`, no `data/…`, two-`New` isolation). S5 Handlers via `Store`/`AuditSink` (stub test, cwd empty). S6 Real `/metrics` (absent not zero). S7 Release train (`v0.1.0` tags, `make verify`).
