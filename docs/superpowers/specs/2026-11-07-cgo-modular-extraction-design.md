# CGO Modular Extraction — Twelve Tools (Design)

**Status:** approved direction, pending implementation plan.
**Date:** 2026-11-07
**Scope:** extract twelve independently versioned tools from the monolith in dependency order, each with own `go.mod`, README, semver, tests, SBOM, and SLSA provenance. No new measurement science, no live runs, no frozen-corpus change.

## Context

Primary-source research in `docs/research/` (40 files, 10,091 lines, including `2026-11-07-modular-toolkit-morph.md` and `2026-11-07-master-demonstration-dgm-needs.md`) establishes:

- One Go module today (`github.com/Realms4239/cgo`, `go.mod:1`); no per-package versioning.
- Pure-logic packages (`model`, `metrics`, `results`, `figures`) are stdlib-only and are the correct first extractions.
- `internal/kit`, `internal/vm`, `pkg/campagne`, `pkg/api`, `pkg/doctor` are the shell — deliberately kept, not extracted.
- Supply chain is the biggest anti-toy gap: `tar.gz` + `.sha256` only, no SBOM/signing/SLSA/OCI.
- Frozen corpus and sealed evidence (CSV + `manifest.json` + `hash8` + dual seals) are the strongest enterprise asset and must survive extraction unchanged.
- 220 ms / 95% is a DGM operator objective, not a WMO mandate (`pkg/metrics/cost.go:75-83`).
- CHR, WIS 2.0/CAP, eBPF, BBRv3, `.dist`, live bench, and case study remain BLOCKED/deferred with named conditions.

## Problem statement

A monolith cannot be versioned, cannot be imported per-feature, and cannot be maintained by more than one human. The jury will call it a toy until each feature is an independently consumable tool with tests, docs, SBOM, and signed provenance.

## Users

- **Primary: Master 2 jury / evaluator** — needs a reproducible evidence engine plus a falsifiable, air-gap-safe demonstration.
- **Secondary: DSI / NOC operator** — needs authenticated, rate-limited, scrapeable services with a working rollback at 3 a.m.
- **Stakeholder: DGM** — needs a system that can be patched, audited and supported after the author leaves.

## Scope

### In scope
- Twelve extractions sequenced 0–7 (see “Slices” below), each merged only when green.
- Per-module `go.mod`, `README.md`, semver, tests, SBOM (CycloneDX), signed provenance (SLSA L2 target).
- Monorepo `require` + `replace` bridge until cutover.
- Build-what-we-lack tools: `cgo-attest`, `cgo-exporter`, `cgo-cost` (proposed in research; scoped here as design only).

### Out of scope
- New measurement science (eBPF object/loader, BBRv3, `.dist` calibration).
- WIS 2.0 / CAP publication; Mosquitto fan-out; CHR boot.
- Live bench runs of any kind.
- Rewriting frozen corpus, API schemas, or cockpit UI beyond the citation guard.

## Product constraints

- A missing measurement is **absent**, never `0`. Prometheus `absent` ⇒ no sample.
- `load_attempted` vs `load_measured` are always exported as a pair; breach never exported without its measured flag.
- Citation is `""` without `provenance.source_sha256`. Single guard shared by card and archive.
- 220 ms / 95% labelled “operator objective” everywhere.
- No number leaves the repo without a verifier pass (`AGENTS.md`). Exit 2 is not green; BLOCKED rows name their unblock condition.
- No new Go dependency without explicit human approval. Every extraction adds no new external dep beyond stdlib unless approved.

## Architecture

```
model → metrics/probe/profile → qdisc/iface → campagne/remediate → audit/results/figures → api/doctor → cmd/* + embedded React
```

### Module map (12 tools)

| # | Tool (repo) | Responsibility | Public API | Depends on |
|---|---|---|---|---|
| 1 | `cgo/model` | Profiles P1–P4, qdiscs, CCs, G0–G7 gates, CSV headers | `Profile`, `Qdisc`, `Event`, `DefaultCCs` | stdlib only |
| 2 | `cgo/metrics` | `Percentile`, `Summarize`, `CostARPerH`, tiers, JFI, verdicts | `Percentile`, `CostARPerH(Tier)`, `PriceTier`, `JFI` | `model` |
| 3 | `cgo/results` | Frozen scan → medians, quarantine, best cell | `Scan`, `ProvenanceSHA/Hash8`, `VerifyManifest` | `metrics`, `model` |
| 4 | `cgo/figures` | SVG generation + provenance SHA | `Generate`, `ProvenanceSHA` | `results`, `model` |
| 5 | `cgo/qdisc` | `tc` Reset/ApplyNetem/ApplyShaper, stats, GE netem, IFB | `Reset`, `ApplyNetem`, `ApplyShaper` | `model` |
| 6 | `cgo/remediate` | Linux `tc` + RouterOS `/queue tree` scripts; `cgo-chr-validate` CLI | `LinuxTC`, `RouterOS`, `Recommend` | `model`, `qdisc` |
| 7 | `cgo/probe` | Ping, small-object, bulk (up/down, N flows) + TCPSock | `Ping`, `SmallObject`, `BulkSendTo`, `BulkDownloadTo` | `model` |
| 8 | `cgo/audit` (+ `cgo-audit` CLI) | 60 s phased audit → sealed report | `Run`, `BuildReport`, `CompareRouterOS` | `metrics`, `probe`, `remediate`, `model` |
| 9 | `cgo/api` (+ sse/metrics) + `cgo/doctor` + `cgo/iface` | REST/SSE hub, `/metrics` honesty contract, preflight, `/proc/net/dev` sampler | `New(d Deps) Handler`, `Report(mode)`, `New()` sampler | promoted modules |
| 10 | `cgo/cite` + proof panel | Seal-gated citation + accessible proof UI | `auditCitation(report) string` | `AuditReport` types |
| 11 | `cgo/guard` + `cgo/attest` + `cgo/exporter` | Bearer auth + Nginx rate limit; corpus-digest attestation; honest Prometheus | `requireAuth`, `attest`, `renderMetrics` | promoted modules |
| 12 | `cgo-lite` + `cgo-demo` | Digest-pinned OCI + `cpu.max`/`memory.max` + rollback; doctor-verify falsifiable arc | `doctor-verify` CLI | promoted modules |

**Kept internal (never libraries):** `pkg/campagne`, `pkg/api` shell, `pkg/doctor`, `internal/kit`, `internal/vm`, `cmd/cgo*` binaries. `cmd/cgo-gui` (Fyne) is its own repo.

### Data flow

```
Operator → Nginx (TLS, limit_req, auth) → Go binary (handlers, /metrics, sealed APIs)
                                           ├─ Prometheus scrape → /metrics → Grafana
                                           ├─ Node/Blackbox exporters → Prometheus
                                           └─ Sealed freeze → CSV + manifest + hash8
```

No new measurement is invented by the ops plane. Grafana renders; Go measures; Nginx guards.

### Normalization

- **State:** one `viewState`/`health`/`archive` rule for `Valide`, `Observation seule`, `Dégradé`, `Preuve manquante`.
- **Citation:** one guard — disabled without `source_sha256`.
- **Metrics:** Prometheus mapping preserves `iface_valid:false ⇒ absent`.
- **Tokens:** cockpit tokens frozen in Slice 27; no new visual language.

## Error handling

- Unauthenticated mutation → `401` with `WWW-Authenticate`; CSRF failure → `403`; rate-limited → `503` (Nginx `limit_req_status`).
- Invalid `run_id` → `400` (`validRunID`); tier change mid-campaign → `409`.
- Scrape failure → target-down alert; `iface_valid:false` ⇒ gap, not zero.

## Testing & verification

- Per-module `go test ./...` in its module dir; monorepo suites stay green during extraction.
- Frontend: 321 Vitest + Playwright cockpit matrix (890×768 / 1920×1080, no horizontal overflow, 1–4 inert in text fields, no per-row waterfall on 179 rows).
- Gates: `python tools/verify.py` six gates PASS (frozen 159 runs, provenance 5 eras, tests, exit, hygiene, links). Exit 2 is not green; BLOCKED rows stay named.
- Jury air-gap arc: six gates → corpus audit → seal + fault proof → `hash8` in four places → `Unavailable` never green. All live claims stay BLOCKED on screen.

## Slices in dependency order

0. **Contract** — `model` public API frozen, README, semver policy.
1. **Pure first** — `metrics`, then `results`/`figures`.
2. **Linux-gated** — `qdisc`, `remediate`.
3. **Experiment** — `probe`, `audit` (+ CLI).
4. **Ops surface** — `api`/`doctor`/`iface`.
5. **Presentation** — `cite` + proof panel.
6. **Ops plane** — `guard`, `attest`, `exporter`, `lite`, `demo`.
7. **Cutover** — monorepo pins via `require` + `replace`; each module tagged, SBOM, signed.

At most three active slices (AGENTS.md cadence). Each slice merges only when green.

## Open decisions requiring human approval

1. Live bench window + host (Linux 6.8).
2. CHR image licence + version pin + download approval.
3. Each new dependency (Nginx, Prometheus, Grafana, exporters, OCI runtime, Mosquitto deferred).
4. WIS 2.0 / CAP scope and primary-doc fetch.
5. Case-study operator grant a–d.
6. Port / network exposure (mgmt subnets, TLS cert issuance).
7. Tariff re-source before any number leaves the repo.
8. Deletion of microsecond / RAM / 3-file sentences from thesis prose.

## Definition of done

- All twelve extractions merged in sequence, each with green gates, SBOM, and signed provenance.
- Monorepo still builds; no frozen-corpus edit; no live run; no new Go dependency without signed human gate.
- Jury air-gap arc reproduces offline; live claims remain explicitly BLOCKED where unproven.
- Research index (`docs/research/README.md` + 40 files) cited for every MUST above.
