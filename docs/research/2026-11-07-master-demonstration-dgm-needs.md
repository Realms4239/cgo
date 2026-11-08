# Master Demonstration & DGM Needs — Deep Research

*Date: 2026-11-07 · Repo: `operator-cockpit` (CGo / Meteolink, VERSION 1.3.5, `go.mod:1-3` module `github.com/Realms4239/cgo`, Go 1.25) · Method: primary sources only; secondary write-ups not trusted. Claims that could not be sourced are marked `unverified — primary source not found`.*

---

## 1. Method & sources

| # | Claim → source → status |
|---|---|
| R1 | Repo product definition (real-time test bench + terminal/browser visualiser, AQM/BBR evidence, frozen CSV + SHA-256) → `README.md:5-6` → **verified (repo)** |
| R2 | Ten-package architecture, single Go binary with embedded React (`go:embed`), SSE 10 Hz hub, G0–G7 gates, freeze+manifest+`hash8` provenance → `docs/architecture.md:1-4`, `:23-41`, `:60-83` → **verified (repo)** |
| R3 | API surface (shape/run/watch/burst blocked 501 in observe mode; cost tiers; audit; remediation; integrity) → `docs/api.md:1-11`, `:59-110` → **verified (repo)** |
| R4 | Deployment matrix (Windows = observe, Linux = full), mono-operator security line (no auth; LAN only) → `docs/deploy.md:9-16`, `:18-28`, `:159-177` → **verified (repo)** |
| R5 | Claim ledger discipline (VERIFIED only on named command+host+date; BLOCKED otherwise) → `docs/VERIFY.md:7-18`, `:83-91` → **verified (repo)** |
| R6 | Frozen corpus numbers (159 runs / 528 rows / 208 valid / 74 degraded / 246 invalid / 22 cells) and seal `docs/frozen-manifests.sha256` → `docs/VERIFY.md:24-26` → **verified (repo)** |
| R7 | Slice index & live-blocker rows (eBPF, `.dist`, BBRv3, CHR, case study all BLOCKED/deferred) → `docs/superpowers/plans/README.md:25-58` → **verified (repo)** |
| R8 | Enterprise spec: M2 thesis title, ISGEI × DGM, five modules, honesty invariants (no-load⇒no-grade; unmeasured⇒Unavailable; 220 ms = DGM operator objective, not WMO mandate; BBR = BBRv1) → `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:10-21`, `:40-68` → **verified (repo)** |
| R9 | Case-study contract (read-only, grant a–d, pseudonym default, five deliverables, nothing executed) → `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md:16-27`, `:40-52`, `:75-78` → **verified (repo)** |
| R10 | Malagasy data prices: Yas Net Month 4.5 Go = 25 000 Ar (5 556 Ar/Go); GOLD 1 M Ar/mo 200 Mbps CIR; PLATINE 2.5 M Ar/mo 30 TB; retail tiers table → `docs/data-prices.md:23`, `:140-144`, `:148-156` → **verified (repo, sourced to yas.mg 31/08/2026)** |
| R11 | WIS 2.0: operational 01/01/2025, MQTT pub-sub over public Internet (vs GTS private links), WIS2 Node = MQTT broker + HTTP server, Global Broker/Cache/Catalogue/Monitor, topic hierarchy, WNM/WCMP2 → https://community.wmo.int/site/knowledge-hub/programmes-and-initiatives/wmo-information-system-wis/wis2-overview → **verified (primary: WMO)** |
| R12 | CAP 1.2: open non-proprietary XML alert format, `<alert>/<info>/<resource>/<area>` structure, urgency/severity/certainty triples → https://docs.oasis-open.org/emergency/cap/v1.2/CAP-v1.2-os.html §1.1–1.3 → **verified (primary: OASIS)** |
| R13 | FQ-CoDel / RFC 8290: hybrid DRR + CoDel AQM, flow queueing for bufferbloat, per-queue CoDel, Linux `fq_codel` canonical → https://www.rfc-editor.org/info/rfc8290 (Abstract, §1–3) → **verified (primary: IETF)** |
| R14 | Linux `tc(8)`: shaping/scheduling/policing on egress/ingress; classless qdiscs include `fq_codel`, `netem`, `tbf`, `pfifo_fast`; `cake` via `tc-cake(8)` → https://manpages.debian.org/bookworm/iproute2/tc.8.en.html (DESCRIPTION, CLASSLESS QDISCS) → **verified (primary: man page)** |
| R15 | Prometheus exposition: text format required (UTF-8, `# HELP`/`# TYPE`, `le="+Inf"` histogram rule, mandatory `Content-Type` since Prometheus 3.0) → https://prometheus.io/docs/instrumenting/exposition_formats/ → **verified (primary: prometheus.io)** |
| R16 | Grafana dashboards: panels/rows/queries over data sources; 150+ data-source plugins; sharing via links/snapshots/exports → https://grafana.com/docs/grafana/latest/dashboards/ → **verified (primary: grafana.com)** |
| R17 | Mosquitto: broker = MQTT; config via `mosquitto.conf(5)`; `$SYS/broker/...` status topics; listeners/TLS/auth/QoS configurable → https://mosquitto.org/man/mosquitto-8.html, https://mosquitto.org/man/mosquitto-conf-5.html (fetched via search extract; full page not fetched) → **partial — primary man page confirmed via search extract, not full fetch** |
| R18 | Go modules: module = versioned package collection; versions follow SemVer; `go.mod` directives (`module`, `go`, `require`, `replace`, `retract`, `exclude`); MVS selects build list → https://go.dev/ref/mod (“Introduction”, “Versions”, “Minimal version selection”) → **verified (primary: go.dev)** |
| R19 | SemVer 2.0.0: MAJOR.MINOR.PATCH increment rules; released versions immutable; 0.y.z = initial development; public API MUST be declared → https://semver.org/ (Summary, Rules 1–8) → **verified (primary: semver.org)** |
| R20 | GoReleaser SBOM: creates SBOMs per artifact via Syft; CycloneDX/SPDX output configurable; `dist` dir; supports checksums/signing/attestations → https://goreleaser.com/customization/sbom/ → **verified (primary: goreleaser.com)** |
| R21 | SLSA v1.0 (retired page, historical): L1 provenance exists; L2 + hosted + authentic; L3 + isolated + unforgeable; producer MUST choose platform, MUST distribute provenance → https://slsa.dev/spec/v1.0/requirements → **verified (primary: slsa.dev; page marked Retired — v1.2 current)** |
| R22 | CycloneDX SBOM: detailed inventory of components + dependencies/hierarchy; machine-readable; vulnerability/compliance integration → https://cyclonedx.org/capabilities/sbom/ → **verified (primary: cyclonedx.org / OWASP)** |
| R23 | OCI: three specs (Runtime, Image, Distribution) for open container formats/runtimes → https://opencontainers.org/ → **verified (primary: opencontainers.org)** |
| R24 | Dublin Descriptors, second cycle (Master): 90–120 ECTS (≥60 at second cycle); advanced knowledge for original thinking; problem-solving in new environments; multidisciplinary integration; critical/ethical analysis; autonomy → https://www.univ-tech.eu/dublin-descriptors (EUT+ alliance page citing Joint Quality Initiative 2004 + EQF 2008) → **verified (primary-citing: EUT+ page reproducing the JQI/EQF descriptors)** |
| R25 | ISGEI/MSTI-STR degree framework, ISGEI Madagascar thesis/jury rubric specifics → web search returned only non-primary results (Facebook page, generic formation portals) → `unverified — primary source not found` |
| R26 | DGM Ampandrianomby operational organisation, sadcnwp service cadences/deadlines → not found in fetched primary sources; repo contract names only "the national forecasting server" pseudonym (`2026-10-30-cgo-case-study-contract.md:34-36`) → `unverified — primary source not found` |
| R27 | MikroTik RouterOS CAKE help page → https://help.mikrotik.com/docs/spaces/ROS/pages/83950303/CAKE returned 404 → `unverified — primary source not found` (CAKE on Linux confirmed via `tc(8)` SEE ALSO `tc-cake(8)`; `pkg/remediate` RouterOS script exists in repo but unproven live) |
| R28 | Cobra / urfave-cli specifics → not fetched this session → `unverified — primary source not found` (repo CLI is hand-rolled in `cmd/cgo/main.go`; see §4) |

---

## 2. What a Master 2 actually needs

**Primary anchor: Dublin Descriptors (second cycle) + EHEA self-certification.** The JQI Dublin Descriptors define master-level learning outcomes as: advanced knowledge allowing original thinking; ability to solve problems in new/unfamiliar environments; integration of multidisciplinary knowledge; critical analysis and ethical reflection; high autonomy in research and learning (R24). EHEA self-certification requires national bodies to demonstrate alignment between national qualifications and these descriptors, with quality-assurance systems and diploma-supplement references (R24, citing Maguire 2011 via EUT+). For CGO the mapping is direct:

- **Original contribution** — the field-overlap is already framed: *"Conception et évaluation d'un banc d'essai reproductible pour la comparaison de politiques AQM et BBR sur des liens d'accès contraints, ISGEI × Météo Madagascar / DGM"* (`2026-10-25-...-enterprise-rebuild-design.md:10-16`). The contribution is a *reproducible evidence engine* (frozen corpus + gates G0–G7 + seal anchor), not a new qdisc algorithm. FQ-CoDel itself is an established standard (RFC 8290, R13) — novelty must be claimed in *measurement method*, not in queueing theory.
- **Reproducibility** — the repo's own ledger standard exceeds the descriptor floor: every VERIFIED row requires a named command, host, date and saved output (`docs/VERIFY.md:83-91`), and the six gates + selftest fault proofs run as `make verify` / `python tools/verify.py` (`docs/CHECKPOINTS.md:114-132`; `docs/VERIFY.md:42-43`). The frozen corpus seal (`docs/frozen-manifests.sha256`, 159 manifests) is proved by `sha256sum -c` + fault-proven seal (`docs/VERIFY.md:26`). This is the correct defence exhibit: *the jury re-runs the gates on their own host.*
- **Disciplinary breadth** — the project spans kernel netem/qdisc (`tc(8)` qdiscs, R14), AQM standards (RFC 8290, R13), economics (enterprise tariffs, R10), and service APIs (Prometheus exposition, R15; Grafana, R16). Each strand is anchored to a primary spec in §1 rather than to marketing claims.
- **What ISGEI/MSTI specifically demands** — `unverified — primary source not found` (R25): no fetchable ISGEI framework PDF, thesis template, or jury rubric was obtained. **Do not invent a jury rubric.** Concretely missing: required thesis word count, number of examiners, rubric weights (code vs writing vs demo), language of defence, and any prohibition on unpublished software in the appendix. These MUST be confirmed with the ISGEI secretariat or MSTI-STR before the defence; the Dublin-Descriptors mapping above is the fallback framework only.

**Defence demo (5-minute arc).** What a jury can verify in minutes without trust: (1) `python tools/verify.py` → six PASS; (2) `python book/tools/audit-archive.py` → both OK lines; (3) `sha256sum -c docs/frozen-manifests.sha256` → 159 OK. All three are air-gap-safe (stdlib/Python only) and already recorded as VERIFIED on the builder host (`docs/VERIFY.md:42`, `:24-26`). Live bench steps stay out of the demo: Linux acceptance, eBPF, BBRv3, CHR and `.dist` are BLOCKED rows (`docs/VERIFY.md:28-30`, `:52-53`, `:57`).

---

## 3. What Météo Madagascar really needs

**Verified operational needs (grounded in repo + primary docs only; speculation excluded).**

1. **Real-time alert/telemetry protection on constrained links.** The enterprise spec names the DGM scenario explicitly: *"download = bulk FROM the sink through shaped `e-wan` — the actual DGM scenario (200 MB EUMETSAT downlink vs CAP egress)"* (`2026-10-25-...-design.md:151-153`), with Model C measuring *"loaded 16 KB completion samples at 220 ms; breach flag < 95%"* (`:134-136`). This is consistent with WMO WIS 2.0's model — publish-subscribe over MQTT on the public Internet, nodes exposing an MQTT broker + HTTP download server, core data via Global Cache (R11) — i.e. small alert objects must survive bulk satellite/data downloads. CAP 1.2 (R12) is the alert *payload* vocabulary; WIS 2.0 is the *transport*. The 220 ms figure is the **DGM operator objective**, never a WMO mandate (invariants, `2026-10-25-...-design.md:58-60`; WMO defines "real time" as a few seconds to minutes).
2. **Cost-bounded bandwidth.** Patched tariff reality: Yas Net Month 4.5 Go = 25 000 Ar (5 556 Ar/Go), FTTH ≈ 490–660 Ar/Go, GOLD dedicated 1 M Ar/mo (200 Mbps CIR), PLATINE 2.5 M Ar/mo (30 TB) (`docs/data-prices.md:23`, `:33-40`, `:140-144`). The project already encodes these as tiers (`docs/api.md:23-25`) — the operator need is *waste cost in Ar/h per scenario*, not raw throughput.
3. **Router-grade remediation on MikroTik-class edge.** Remediation generates Linux `tc` and RouterOS queue scripts (`docs/api.md:101` `/api/remediate`); the case-study contract allows read-only inspection of `tc qdisc`/link-rate output *provided by the operator* (`2026-10-30-...-contract.md:20-21`). CHR live proof is BLOCKED (`docs/VERIFY.md:57`) — the need is verified as a *requirement*, not as current capability.
4. **Read-only evidence without operational risk.** The contract's seven prohibitions (no capture, probe, load, write, root, login, scan) and grant items a–d are the operator's actual risk posture (`2026-10-30-...-contract.md:29-38`, `:65-78`). VERIFIED status of the slice: **BLOCKED until the written grant exists** (`docs/VERIFY.md:60`).

**Speculation (NOT verified — excluded from requirements):** DGM Ampandrianomby site topology, sadcnwp host/service inventory, actual cadences/deadlines per service class, WIS2 Node registration status (centre-id, GISC assessment state) — all `unverified — primary source not found` (R26). The Global-Broker/registration flow is documented (R11) but Madagascar's specific registration state was not established from primary sources. **Do not claim CGO is a WIS2 Node or CAP producer** — CGO measures link behaviour; dissemination stays with wis2box-class tooling (R11: *"WIS2 in a box … open-source reference implementation of a WIS2 Node"*). The correct product boundary: CGO = link auditor/bench; WIS2 = transport; CAP = alert format.

---

## 4. Modular repo/tool model

**Per-module repos.** Current inventory (from §1 shell output): `pkg/api` (8 files), `pkg/audit` (15), `pkg/campagne` (20), `pkg/doctor` (1), `pkg/figures` (1), `pkg/iface` (1), `pkg/metrics` (5), `pkg/model` (2), `pkg/probe` (14), `pkg/profile` (1), `pkg/qdisc` (5), `pkg/remediate` (1), `pkg/results` (4); `cmd/cgo`, `cmd/cgo-chr`, `cmd/cgo-gui`, `cmd/meteolink`. `docs/architecture.md:10-21` already documents the ten-package responsibility split. Per Go module rules (R18), each split repo becomes its own module with its own `go.mod`, and consumers `require` it at a SemVer version — the current monorepo `go.mod:1-3` is the anti-pattern to unwind.

**Dependency order and boundaries (from `pkg/*` + `docs/architecture.md`):**

```
model (types: profiles P1-P4, qdiscs, CC, gates G0-G7, CSV headers)   ← no deps
  ↑                    ↑
metrics (Percentile, Summarize, CostARPerH)   probe (ping/small/bulk sondes)   profile (import registry)
  ↑                    ↑                          ↑
qdisc (tc: Reset, ApplyNetem, ApplyShaper, SumBytes/…Drops)  iface (twin sampler)
  ↑                    ↑
campagne (Matrix, RunEvent baseline→charge→recovery, Deps) + remediate (tc/RouterOS scripts)
  ↑                    ↑
audit (link audit → data/link_audit.csv) + results (Scan data/runs → medians) + figures (SVG + ProvenanceSHA/hash8)
  ↑                    ↑
api (REST + SSE hub 10 Hz, stdlib only) — doctor (health checks)
  ↑
cmd/cgo, cmd/meteolink (CLI + server) + web/frontend (React via go:embed, go:build constraint on dist)
```

- **Boundaries:** `model` is the only package with zero internal deps — extract it first. `qdisc` wraps `tc` subprocess calls (stdlib exec; `docs/architecture.md:97-103` confirms cross-compilable Go-only HTTP/SSE). TUI lives in `cmd/cgo/tui.go` (Bubble Tea, `go.mod:7-8`); GUI in `cmd/cgo-gui` (Fyne, `go.mod:6`, **pre-existing Linux cross-build failure** — `docs/VERIFY.md:27` — must be its own repo so it never blocks the core build).
- **Versioning:** SemVer rules (R19) per module; Go canonical versions are `v`-prefixed (`go.dev/ref/mod`, R18) — e.g. current release `1.3.5` (`VERSION`) tags as `v1.3.5`. Crucially, rule 3 (R19): released versions are immutable — the frozen corpus discipline (`book/data/runs` never touched, AGENTS.md) is the dataset analogue and MUST survive the split unchanged.
- **Open-source practice:** MIT licence (`LICENSE:1-3`); SLSA build levels (R21 — cheapest meaningful target is L2: hosted builder + authentic provenance, which GoReleaser + GitHub Actions provides); SBOM via GoReleaser/Syft emitting CycloneDX (R20, R22); OCI image spec for container distribution (R23). Maintenance: per-repo SemVer bump policy + `retract` directive for bad releases (Go modules, R18).

**Continuous maintenance:** each repo gets its own CI running the six-gate analogue (vet/test/archival audit/CLI exit/hygiene/links). The monorepo gains a top-level meta-repo pinning sibling versions via `require` + `replace` (Go modules, R18) — upgrading one module never rewrites frozen data (R19 rule 3 applied to data).

---

## 5. Demonstration design

**Failure → evidence → remediation arc (5 minutes, air-gap-safe):**

| Minute | Jury sees | Source of truth |
|---|---|---|
| 0:00–1:00 | `python tools/verify.py` → six PASS (frozen, provenance, tests, exit, hygiene, links) | `docs/VERIFY.md:42` (VERIFIED 2026-10-28) |
| 1:00–2:00 | `python book/tools/audit-archive.py` → both OK lines (159 runs / 528 rows / 22 cells) | `docs/VERIFY.md:24` |
| 2:00–3:00 | `sha256sum -c docs/frozen-manifests.sha256` → 159 OK; mutate one file in a temp copy → seal catches it | `docs/VERIFY.md:26` (fault proof recorded) |
| 3:00–4:00 | `cgo verify` / `cgo figures` regenerate from frozen CSV; `hash8` identical in Wall, Résultats, Provenance, `report.md` | `README.md:42-46`, `docs/architecture.md:80-83` |
| 4:00–5:00 | Verdict cascade: a fixture report without measured load renders `Unavailable`, never green | `docs/VERIFY.md:63` `TestVerdictNeverGreenWithoutGrade` — *unit fixtures only, never observed against a real audit* (stated honestly) |

**How to falsify/verify:** the ledger's fault-proof requirement — every gate carries a selftest proving it catches injected mutations (`tools/verify.py --selftest`, `docs/VERIFY.md:42`). The demo should invite the jury to break one: the seal (`docs/VERIFY.md:26`), the hook (`docs/VERIFY.md:43`), or a gate (`docs/VERIFY.md:42`).

**Air-gap:** all demo commands are local (Python + Go stdlib + `sha256sum`); dashboard runs single-binary with embedded React and no DB (`README.md:37`). No network required. The operator-cockpit branch runs offline — this is a design property, not a demo hack (`docs/architecture.md:97-103`).

**What stays explicitly unproven in the room:** Linux 6.8 live acceptance, eBPF, IFB ingress, BBRv3, `.dist`, CHR, the case study — all named BLOCKED conditions (`docs/VERIFY.md:28-30`, `:52-53`, `:57`, `:60`). Saying so is the credibility signal.

---

## 6. Anti-toy standards

| Area | Standard (primary) | CGO status |
|---|---|---|
| Versioning | SemVer 2.0.0: MAJOR.MINOR.PATCH rules; released versions immutable; declare public API (R19, https://semver.org/) | Single `VERSION 1.3.5` + MIT; no per-module SemVer yet — split requires it |
| Modules | Go modules: `go.mod` directives; canonical `v`+SemVer versions; `retract` for bad releases (R18, https://go.dev/ref/mod) | Monorepo `go.mod:1-60` (Fyne/BubbleTea/sys deps) — pre-split |
| SBOM | CycloneDX: component+dependency inventory, machine-readable (R22, https://cyclonedx.org/capabilities/sbom/) | None yet — GoReleaser `sboms` step (R20) is the adoption path |
| Supply chain | SLSA v1.0 L2: provenance exists + hosted + authentic (R21, https://slsa.dev/spec/v1.0/requirements) | GitHub Releases + `.sha256` archives (`README.md:79-103`) — checksums yes, signed provenance no |
| Metrics | Prometheus exposition: `# HELP`/`# TYPE`, `le="+Inf"`, mandatory Content-Type (R15) | SSE/JSON only (`docs/api.md`) — add `/metrics` exposition |
| Containers | OCI Runtime/Image/Distribution specs (R23, https://opencontainers.org/) | `docker compose` digest-pinned (`docs/deploy.md:182-183`) — matches OCI distribution model |
| Dashboards | Grafana: panels/rows/queries; 150+ data sources; shareable snapshots (R16) | Bespoke React dashboard embedded — federation via Grafana is optional, not required |
| Broker | Mosquitto `mosquitto.conf(5)`, `$SYS` topics, TLS/auth (R17) | No MQTT in repo (sibling research confirms) — WIS2 alignment deferred until CGO auth lands |
| Docs | Ledger discipline: claim→row→command+host+date (`docs/VERIFY.md:83-91`) | **Already at standard** — strongest anti-toy asset |
| Tests | Six gates + selftest fault proofs; Go suite + frontend 283/283 (`docs/VERIFY.md:72-81`) | **Already at standard** |
| Air-gap | Single static binary, embedded frontend, no service deps (`README.md:37`; `docs/architecture.md:97-103`) | **Already at standard** |

---

## 7. Proposed essential structure

**The project essentially is one evidence engine plus five things around it.** CGO = the reproducible measurement bench + audit + provenance system (`README.md:5`; `docs/architecture.md:1-4`). Around it: (1) **WIS 2.0/CAP are neighbours, not parts** — transport and payload standards CGO's *measurements* should satisfy (R11, R12); (2) **ops surface** — Prometheus exposition (R15) + Grafana panels (R16) layered over the existing SSE/API; (3) **distribution** — GoReleaser + SLSA L2 + CycloneDX SBOM + OCI image (R20–R23); (4) **modular Go modules** per §4 dependency order with SemVer (R18, R19); (5) **ledger discipline** — the VERIFY/CHECKPOINT system that already exceeds toy-project practice (`docs/VERIFY.md:83-91`).

Final one-line architecture: **`model → metrics/probe/profile → qdisc/iface → campagne/remediate → audit/results/figures → api/doctor → cmd/* + embedded React`, released as SemVer Go modules with SLSA provenance, CycloneDX SBOMs, and OCI images, demonstrated via the six-gate + seal + hash8 arc.**

---

## 8. Open decisions requiring human approval

| Decision | Owner / blocker | Condition |
|---|---|---|
| Any live bench run (Linux 6.8 acceptance, dual-ns, IFB) | Human approver (AGENTS.md: live bench needs explicit approval) | Stable Linux 6.8 host + kernel — `docs/VERIFY.md:28`, `:31` |
| eBPF object build | Human | `make bpf` on Linux 6.8; object never committed until then — `docs/VERIFY.md:29-30` |
| `.dist` calibration | Human + operator | Field trace ≥ 30 baseline samples — `docs/VERIFY.md:53` |
| BBRv3 | Human | Kernel ≥ 7.1 with BBRv3 — `docs/VERIFY.md:52` |
| CHR image + QEMU | Human | Install QEMU, download CHR, record `windows-tcg` proof — `docs/VERIFY.md:57` |
| WIS2 Node registration (if DGM wants one) | PR/NFP via WMO Secretariat form + GISC assessment | Only after MQTT endpoint operational — R11 registration steps; **not a CGO deliverable** |
| Case-study grant (sadcnwp service map) | Operator, written grant items a–d | `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md:65-78`; until then nothing executes |
| Monorepo → per-module repos split | Human (scope decision) | New dependency edge per module; frozen corpus unchanged (R19 rule 3) |
| Merge to `main` | Human only | `docs/CHECKPOINTS.md:314` |
| ISGEI/MSTI rubric | Human (institution) | `unverified — primary source not found` (R25) — obtain from secretariat before defence |

*Research only — no code modified, no commit made. Every factual claim above cites a primary source (URL or repo `file:line`); unknowns are marked `unverified — primary source not found` explicitly (R25–R27).*
