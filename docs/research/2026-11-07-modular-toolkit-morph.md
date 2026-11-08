# Morphing CGO into Modular Open-Source Tools — Primary-Source Research

*Date: 2026-11-07 · Scope: how to split the CGO monolith into reusable, framework-grade tools without losing frozen evidence. Research only — no code changed.*

## 1. Method & sources (table)

| # | Source (primary) | URL | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 (FQ-CoDel) | https://www.rfc-editor.org/rfc/rfc8290.txt | 200 | Q_disc / QoS semantics |
| R2 | RFC 3095 (ROHC) | https://www.rfc-editor.org/rfc/rfc3095.txt | 200 | header compression context |
| R3 | Go Modules Reference | https://go.dev/ref/mod | 200 | module boundaries |
| R4 | go.mod reference | https://go.dev/doc/modules/gomod-ref | 200 | single-module current state |
| R5 | Go `embed` package | https://go.dev/pkg/embed/ | 200 | frontend / object embed |
| R6 | Prometheus exposition formats | https://prometheus.io/docs/instrumenting/exposition_formats/ | 200 | honesty exporter spec |
| R7 | Nginx `limit_req` module | https://nginx.org/en/docs/http/ngx_http_limit_req_module.html | 200 | rate-limit shape precedent |
| R8 | Mosquitto `mosquitto.conf` man (v5) | https://mosquitto.org/man/mosquitto-conf-5.html | 200 | MQTT topics/QoS/TLS/auth |
| R9 | `tc(8)` man page | https://man7.org/linux/man-pages/man8/tc.8.html | 200 | tc-apply, qdisc/class/filter |
| R10 | Kernel CFS Bandwidth Control | https://www.kernel.org/doc/html/latest/scheduler/sched-bwc.html | 200 | cgroups v2 `cpu.max`/`memory.*` |
| R11 | GoReleaser | https://goreleaser.com/ | 200 | versioning / signing / SBOM |
| R12 | SLSA spec (about) | https://slsa.dev/spec/v1.2/about | 200 | supply-chain attestation |
| R13 | CycloneDX SBOM | https://cyclonedx.org/capabilities/sbom/ | 200 | SBOM format |
| R14 | Semantic Versioning 2.0.0 | https://semver.org/ | 200 | semver policy |
| R15 | OCI overview | https://opencontainers.org/about/overview/ | 200 | OCI image / air-gap distribution |
| R16 | Cobra (upstream repo) | https://github.com/spf13/cobra | 200 | CLI pattern reference |
| R17 | iperf3 (upstream repo) | https://github.com/esnet/iperf | 200 | library-vs-binary pattern |
| — | `go.dev/blog/go-modules` | (old blog URL) | **404** | marked unverified — primary source not found at that path |
| — | `go.dev/ref/embed` | (wrong path) | **404** | unverified — correct path is R5 |
| — | Kernel `sch_cake.html` (that exact path) | https://www.kernel.org/doc/html/latest/networking/scheduler/sch_cake.html | **404** | unverified — primary source not found at that path |
| — | MikroTik `help.mikrotik.com` CAKE pages probed | https://help.mikrotik.com/docs/spaces/ROS/pages/152321195/CAKE | **404 (page)** | RouterOS CAKE details unverified — primary source not found |
| — | Mosquitto `/documentation/` index | https://mosquitto.org/documentation/ | **ECONNRESET** | unverified — fetch failed |
| — | `warnings.wmo.int` | https://warnings.wmo.int/iwp/OPCMMS/en/ | **ENOTFOUND** | WIS 2.0 / CAP endpoint unverified — primary source not found |
| — | `tcutils.org`, `cakeqos.com`, `flent-utils.github.io` (404) | — | **fail/404** | flent/tcutils structure unverified from first-party docs |
| — | `datatracker.ietf.org/doc/html/rfc6099` | — | **404** | unverified (wrong number) |

Method: read `go.mod`, `README.md`, `docs/architecture.md`, `docs/api.md`, `docs/superpowers/plans/README.md`; list `pkg/`, `cmd/`, `internal/` exports with `go list`/`grep`; fetch primary docs above. Every technical claim below cites either `file:line` or `(R#)`. Anything else is labelled **unverified — primary source not found**.

## 2. Current module inventory (package → responsibility → importable? → evidence)

One Go module: `github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1-3`, R4). No nested `go.mod` exists under `pkg/` — everything is one published module, so there is no independent versioning per package (R3). Depends only on `fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys` directly (`go.mod:5-10`); the rest is GUI/terminal transitive deps (`go.mod:12-60`).

| Package | Responsibility (current) | Exported API (sample) | Independently importable? | Evidence |
|---|---|---|---|---|
| `pkg/model` | Domain types: profiles P1–P4, qdiscs, CCs, G0–G7 gates, events, CSV headers | `type Profile`, `type Qdisc string`, `type Event`, `DefaultCCs() []string` | Yes as Go package (zero external deps beyond stdlib), but same module — no own version | `docs/architecture.md:21` |
| `pkg/metrics` | `Percentile`, `Summarize`, `DeadlineOKPct`, `CostARPerH`, tiers, economics, JFI, Mann-Whitney | `func Percentile`, `func CostARPerH`, `type PriceTier` | Yes (stdlib only), same module | `docs/architecture.md:16` |
| `pkg/profile` | Dynamic profile registry (import CSV/JSON) | `func Import(p model.Profile) error`, `func Delete`, `func Reserved` | Yes, same module; imports `pkg/model` | `docs/architecture.md:20` |
| `pkg/probe` | Concurrent probes: ping, small-object, bulk (up/down, N flows) + Linux CC socket tuning + eBPF embed sidecar | `func Ping`, `func SmallObject`, `func BulkSendTo`, `func BulkDownloadTo` | Yes, same module; Linux/Windows split via `_linux`/`_other` files | `docs/architecture.md:15`, `pkg/probe/bpf/assets_embedded.go:7` (`//go:embed bpf_probe.o`) |
| `pkg/qdisc` | tc layer: `Reset`, `ApplyNetem`, `ApplyShaper`, stats, distribution/GE netem, ingress/IFB | `func Reset`, `func ApplyNetem`, `func ApplyShaper`, `func ApplyNetemPhysics` | Yes, same module; shells `tc` as subprocess (cross-compilable) | `docs/architecture.md:14`, `docs/architecture.md:95-103` |
| `pkg/campagne` | Matrix orchestration (`Matrix`, `StartMatrixWithID`, `RunEvent`, `Deps`, dualns writer) | `type Matrix`, `type Deps`, `func RunEvent` | Yes, same module; depends on `model`+`qdisc`+`probe` | `docs/architecture.md:13` |
| `pkg/audit` | Client link audit (idle/loaded), sealed reports, telemetry, RouterOS config compare | `func Run`, `func BuildReport`, `func CompareRouterOS` | Yes, same module; depends on `metrics`+`remediate`+`model` | `docs/architecture.md:19` |
| `pkg/remediate` | Remediation scripts: Linux tc + RouterOS `/queue tree` | `func LinuxTC`, `func RouterOS`, `func Recommend` | Yes, same module | `docs/api.md:101` (`/api/remediate`) |
| `pkg/results` | Scan frozen runs → per-cell medians, quarantine, best cell; seed archive | `func Scan`, `func ProvenanceSHA/Hash8`, `func SeedRuns` | Yes, same module; `//go:embed runs-frozen.tar.gz` at `pkg/results/seed.go:15` | `docs/architecture.md:17` |
| `pkg/figures` | SVG generation (bars, scatter) + provenance SHA | `func Generate`, `func ProvenanceSHA` / `ProvenanceHash8` | Yes, same module | `docs/architecture.md:18` |
| `pkg/iface` | Live `/proc/net/dev` sampler, 10 Hz window | `func New`, `type Sampler` | Yes, same module | `docs/architecture.md:87-93` (`iface_*` fields) |
| `pkg/doctor` | Preflight checks (os, tc, cap_net_admin, bbr, ping) | `func Report(mode string) (string, []Check)` | Yes, same module | `docs/api.md:18` (`/api/doctor`) |
| `pkg/api` | HTTP + SSE surface (stdlib only): REST, 10 Hz hub, translate | `func New(d Deps) Handler`, `func NewHub() *Hub`, `func HandleTranslate` | Yes, same module; stdlib only | `docs/architecture.md:12` |
| `cmd/cgo` | Single binary: CLI + server + setup/kit/tui | `package main` + flag parsing | Binary, not importable | `docs/architecture.md:4` |
| `cmd/meteolink` | Alias re-exec of sibling `cgo` binary | `package main` | Binary alias; `cmd/meteolink/main.go:9-14` | re-exec, not a library |
| `cmd/cgo-chr` | MikroTik CHR (QEMU) runner/proof | `type CHRProof` | Binary | `docs/superpowers/plans/README.md:44` (slice 16) |
| `cmd/cgo-gui` | Fyne desktop console (27 buttons) | `package main` | Binary; heavy dep (Fyne via `go.mod:6`) | `README.md:77` |
| `internal/kit` | Deployment engine (21 actions), VM scan/deploy/secure | `package kit` (exported actions) | Internal — not importable outside module by Go rule | `README.md:135-151` |
| `internal/vm` | VM discovery (VMware/VBox), NAT, guest mgmt | `package vm` | Internal | `internal/` Go rule (R3) |
| `web/frontend` | React dashboard, built by Vite, embedded via `go:embed all:dist` | `web/frontend/embed.go:11` | Not Go-importable; `README.md:37` |

Cross-cutting facts from source:

- Dashboard is embedded, not served from disk: `web/frontend/embed.go:11` (`//go:embed all:dist`); README calls it "un seul binaire" (`README.md:37`, R5).
- SSE hub: 10 Hz, ring 2048, 64-subscriber cap, backpressure event, `Last-Event-ID` replay (`docs/architecture.md:25-41`).
- tc semantics are reset-then-apply: `Reset` → `ApplyNetem` (root `1:`) → `ApplyShaper` child (`docs/architecture.md:46-53`); matches `tc` qdisc/class model (R9) and FQ-CoDel (R1).
- P1–P4 profiles: fibre 80 / 4G 20 / VSAT 5 / Starlink 100 Mbit/s (`README.md:40`).
- Modes: Linux full vs Windows observe (shape/run/watch/burst → 501) (`docs/architecture.md:87-89`, `README.md:303`).

## 3. Primary-source requirements for modularity/open-source

- **One module = one version.** Per Go Modules Reference (R3), a module is a collection of packages versioned together as `module@version`; nested `go.mod` files create submodules with independent versions. CGO has exactly one module (`go.mod:1`) so `pkg/model` cannot be versioned independently today.
- **`internal/` is unimportable by design.** Go restricts `internal/` imports to within the parent module (R3). `internal/kit`, `internal/vm` cannot be published as tools.
- **`go:embed` bakes assets at compile time** (R5). `web/frontend/embed.go:11` and `pkg/probe/bpf/assets_embedded.go:7` mean the binary is self-contained but frontend changes require rebuild; that's a strength for air-gap (no runtime deps, `README.md:37`) and a weakness for versioning (embed hash ties all assets to one release).
- **HTTP surface is hand-rolled stdlib, no web framework** (`docs/architecture.md:97-99`). That's credible for a single binary; for a toolkit, each module needs its own API doc (`docs/api.md` currently documents the monolith, not a library).
- **Raw metrics out** — Prometheus exposition format (R6) defines the textual format for scrapers. CGO exposes `/api/*` JSON (docs/api.md) but has no Prometheus exposition endpoint: no `http`-style exposition, no counters/gauges per (R6).
- **tc application is subprocess-based** (`docs/architecture.md:102-103`): cross-compilable, but tc behaviour is host-dependent (Linux only for shaping; Windows = observe, `docs/api.md:9-11`). This is the correct upstream constraint (R9) and must be surfaced per-module, not hidden.
- **CAKE/FQ-CoDel standards**: CAKE is documented in the Linux scheduler docs (R1 covers FQ-CoDel; CAKE specifics: **unverified — primary source not found at the probed path**). RouterOS `/queue tree` + IFB and MikroTik CAKE parameter pages: **unverified — primary source not found at the probed URLs** (§1). WIS 2.0 / CAP (`warnings.wmo.int`): **unverified — primary source not found** (DNS failure). Slice 13's WIS 2.0 SLA paragraph (`plans/README.md:229`) therefore has no fetched primary source yet.
- **MQTT (Mosquitto)**: topics/QoS/TLS/auth are configured in `mosquitto.conf` (R8). The exact conf-page for v5 was fetched (200); the index page failed (ECONNRESET) — per-topic wildcard/auth specifics are **unverified — primary source not found** beyond the man page.
- **Nginx `limit_req`** (R7) is an HTTP-layer rate limiter. It is a different enforcement point than edge shaping via `tc`/CAKE (R9) — conflating the two would be a modelling error.
- **cgroups v2 budgets**: `cpu.max`/`memory.max` semantics per Kernel CFS Bandwidth Control (R10). CGO caps bench throughput via tc shaping, not cgroup limits — cgroups are **unverified** as a CGO control (no evidence in source searched).
- **Supply chain**: semver (R14), GoReleaser + signing/SBOM (R11), SLSA provenance (R12), CycloneDX SBOM (R13), OCI distribution (R15). CGO releases are `tar.gz` + `.sha256` only (`README.md:84-85`) — no SBOM, no signatures, no provenance records. See §7.

## 4. Gap analysis (monolith vs required)

| Required for "framework-grade" | Status in CGO | Evidence | Gap |
|---|---|---|---|
| Per-package versioning | Single module, single binary | `go.mod:1` | No independent semver; one bump changes everything (R3, R14) |
| Independent import | All `pkg/*` technically importable, but coupled via `model` + shared `Deps` | `cmd/cgo/main.go` imports `audit`+`campagne`+`kit` | Importing `metrics` pulls the same module graph (no submodule) |
| Reusable binaries | One binary (`cgo`) + alias (`meteolink`) + GUI (`cgo-gui`) + `cgo-chr` | `cmd/*` | No `cgo-audit`, `cgo-figures`, `cgo-probe` CLIs; audit/figures are subcommands only |
| Public API docs per module | Monolith API doc | `docs/api.md:1-3` | Library consumers have no `pkg/*` godoc split |
| Prometheus metrics | JSON API only | `docs/api.md` | No exposition (R6) |
| MQTT publishing | None found | grep-style absence: no MQTT code in `pkg/` inventory (§2) | Pasted-context demand (Mosquitto) unimplemented |
| RouterOS validation | RouterOS script generator only, no live validator/proof loop | `pkg/remediate`, slice 16 deferred (`plans/README.md:44`) | No CAKE validator; CHR proof blocked on QEMU (`plans/README.md:44`) |
| Signed/SBOM releases | SHA256 files only | `README.md:84-85` | No sigstore/SLSA attestation (R12), no SBOM (R13) |
| Air-gap distribution | Static single binary + embedded frontend | `README.md:37`, R5 | No signed OCI image, no offline bundle doc (R15) |
| Tests | Strong (`python tools/verify.py` 6 gates) | `plans/README.md:352-365` | Gates exist but per-module test ownership is unclear for extraction |

## 5. Target modularization (table: package → module/tool → API → consumer → slice)

Design principles (R3, R11, R14): (1) keep `pkg/model` as the shared kernel — every other module depends on it first; (2) promote pure-logic packages to submodules before anything that shells out; (3) `internal/` and `cmd/` binaries never become libraries; (4) each promoted module gets `go.mod`, `README.md`, semver, and tests before it's "done".

| Package | Decision | Target module / tool | Public API surface | Depends on | Inside CGO consumer | Slice (dependency order) |
|---|---|---|---|---|---|---|
| `pkg/model` | **Promote** | `github.com/Realms4239/cgo/model` (submodule) | `Profile`, `Qdisc`, `Event`, `DefaultCCs`, gate constants | stdlib only | all below | S-A (first — nothing else moves first) |
| `pkg/metrics` | **Promote** | `cgo/metrics` | `Percentile`, `Summarize`, `CostARPerH(Tier)`, `PriceTier`, `JFI`, `AssessVerdict` | `model` | `audit`, `campagne`, `api` (`cost/*`) | S-A (with model) |
| `pkg/results` | **Promote** | `cgo/results` | `Scan`, `ProvenanceSHA/Hash8`, `VerifyManifest` | `metrics`, `model` | `figures`, `api` (`integrity`, `results`) | S-B (after metrics) |
| `pkg/figures` | **Promote** | `cgo/figures` | `Generate`, `ProvenanceSHA` | `results`, `model` | `api` (`figures/regen`) | S-B (after results) |
| `pkg/qdisc` | **Promote** | `cgo/qdisc` | `Reset`, `ApplyNetem`, `ApplyShaper`, `ApplyNetemPhysics`, `PollStats`, `GEModel` | `model` | `campagne`, `remediate` | S-C (after model; Linux-gated) |
| `pkg/remediate` | **Promote** | `cgo/remediate` (+ `cgo-chr` proof CLI reuse) | `LinuxTC`, `RouterOS`, `Recommend`, `CompareRouterOS` | `model`, `qdisc` | `audit` (`BuildReport`), `api` (`remediate`) | S-C (after qdisc) |
| `pkg/probe` | **Promote** | `cgo/probe` | `Ping`, `SmallObject`, `BulkSendTo`, `BulkDownloadTo`, `DialWithCC`, `SmallClient` | `model` | `audit`, `campagne`, `api` (`burst`) | S-D (after model; platform-split) |
| `pkg/audit` | **Promote** | `cgo/audit` (library) + `cgo-audit` binary | `Run`, `BuildReport`, `ResolveSinks`, `CompareRouterOS`, `Presets` | `metrics`, `probe`, `remediate`, `model` | `api` (`audit/*`), `cmd/cgo` | S-D (after probe+remediate) |
| `pkg/iface` | **Promote** | `cgo/iface` | `New`, `Sampler.Sample()` | stdlib only | `api` (hub clock) | S-E (last; isolated) |
| `pkg/profile` | **Promote** | `cgo/profile` | `Import`, `Delete`, `List`, `Reserved` | `model` | `api` (`profile/*`) | S-E |
| `pkg/doctor` | **Keep-as-internal** (ship as `cgo doctor` subcommand of each binary, not a module) | — | `Report(mode)` stays internal to binaries | `model`, `probe` | `api` (`doctor`), `cmd/cgo` | — (no slice; document as internal) |
| `pkg/campagne` | **Keep-as-internal** (orchestrator of the monolith; becomes the runner inside `cgo` binary only) | — | — | `model`, `qdisc`, `probe`, `metrics` | `cmd/cgo`, `api` (`run/*`) | — (monolith core; extract only if a second runner appears) |
| `pkg/api` | **Keep-as-internal** (replaced per-module by each tool's own HTTP surface) | — | — | all above | `cmd/cgo --serve` | — (monolith shell; each promoted module grows its own server in its slice) |
| `cmd/cgo`, `cmd/meteolink`, `cmd/cgo-gui`, `cmd/cgo-chr` | **Keep** | Binaries | CLI flags documented in `docs/api.md:167-180` | `internal/*`, promoted modules | — | Migration slices progressively re-point `cmd/cgo` at the submodules |
| `internal/kit`, `internal/vm` | **Keep internal** (never published; move into each binary that needs them) | — | — | stdlib + `golang.org/x/sys` | `cmd/cgo` | — |
| `web/frontend` | **Separate repo at extraction** | `cgo-console` npm package → consumed via `go:embed` (R5) | React components / REST client | promoted module REST APIs | `cmd/cgo` | S-F (after all API modules exist) |

Target repo layout after S-A..S-F (R3, R14, R11):

```
github.com/Realms4239/cgo/            # monolith shell: cmd/*, pkg/{api,campagne,doctor}, internal/*
  model/go.mod  metrics/go.mod  results/go.mod  figures/go.mod
  qdisc/go.mod  remediate/go.mod  probe/go.mod  audit/go.mod (+ cgo-audit/ binary)
  iface/go.mod  profile/go.mod
```

Each `go.mod` references published versions of the others (no `replace` directives in release tags). Binaries (`cgo`, `cgo-audit`, `cgo-figures`, `cgo-verify`) each ship their own release archives (R11).

## 6. Build-what-we-lack opportunities (each: gap, standard, tool, API)

1. **Prometheus honesty exporter** — Gap: `/api/*` is JSON-only (`docs/api.md`); scrapers and Grafana can't consume it; no `openmetrics` text (R6). Standard: Prometheus exposition format (R6). Tool: `cgo/prometheus` submodule + `cgo-exporter` binary. API: `Handler()` (http.Handler), `Collect(state) []Metric` — never synthesizes values, only re-exports measured `pkg/metrics` outputs with `honesty` labels.
2. **Sealed-evidence attestor** — Gap: integrity is `sha256` + `hash8` in-band (`docs/architecture.md:76-83`, `README.md:163`); no detached signature or SLSA-style attestation (R12). Standard: SLSA provenance + CycloneDX SBOM adjuncts (R12, R13). Tool: `cgo/attest` submodule + `cgo attest sign|verify`. API: `Attest(dir) (*Attestation, error)`, `Verify(path) error`.
3. **Profile importer / converter** — Gap: ad-hoc CSV/JSON import (`/api/profile/import`, `docs/api.md:44`); no round-trip, no validation report, no versioned schema. Standard: `pkg/profile` + `pkg/model` types as the versioned schema. Tool: `cgo/profile` submodule + `cgo profile import|export`. API: `Import(r io.Reader) (*ImportResult, error)` (already the shape in `pkg/profile`).
4. **Cost-tier calculator** — Gap: tiers are baked in `pkg/metrics` + `docs/data-prices.md` research; no standalone calculator for an operator to validate a landed link before an audit. Standard: `pkg/metrics.CostARPerH` + economics types. Tool: `cgo/cost` binary (smallest useful subcommand). API: `CostARPerH`, `PriceTier`, `Tiers()`.
5. **RouterOS CAKE validator** — Gap: generator only (`pkg/remediate`); no checker that a running RouterOS `/queue tree` matches a remediation script; CHR proof blocked on QEMU (`plans/README.md:44`). Standard: RouterOS CLI (`/queue tree`, CAKE parameters) — **unverified — primary source not found** (probed pages 404, §1). Tool: `cgo/remediate` gains `Validate(device) []RouterOSCheck` (compare direction already exists: `CompareRouterOS`). API: `ParseRouterOSConfig`, `CompareRouterOS`, `Validate`.
6. **MQTT bridge for field alerts** — Gap: no MQTT code in inventory (§2). Standard: Mosquitto `mosquitto.conf` topics/QoS/TLS/auth (R8). Tool: `cgo/mqtt` (optional, not in critical path). API: `Publish(ctx, audit.Report) error` — QoS 1, TLS client-cert required; specifics **unverified — primary source not found** beyond R8.
7. **qdisc policy planner** — Gap: `pkg/qdisc` applies commands; nothing validates a policy *before* applying it (e.g., "CAKE on a 5 Mbit/s VSAT link" vs capacity). Standard: `tc` qdisc/class model (R9) + FQ-CoDel (R1). Tool: `cgo/qdisc plan <policy.yaml>` — dry-run validator only. API: `Plan(qdisc, capacity, delay…) ([]string, error)` (reuses current `Apply*` command builders, adds validation).

Each of these is a standalone open-source tool *because* the capability is missing from the current repo inventory (§2) — not because of any imported claim.

## 7. Anti-toy standards (versioning, SBOM, docs, air-gap distribution)

| Standard | Primary source | Current state | Required action |
|---|---|---|---|
| Semver per module | R14 | Single version `1.3.5` (`README.md:1`, `web/frontend/package.json:4`) | Version each submodule independently (`vMAJOR.MINOR.PATCH`); `cgo` binary version = composition, not sum |
| SBOM | R13 | None (`tar.gz` + `.sha256`, `README.md:84-85`) | Emit CycloneDX SBOM per submodule (`go mod vendor` + GoReleaser `sbom:` hooks, R11, R13) |
| Signing / distrusted build | R11 (GoReleaser signing), R12 (SLSA) | Checksums only | Sign release artifacts (cosign/gpg via GoReleaser, R11); attach SLSA provenance records (R12) — records builds, not measurements |
| Docs | R3 (godoc per module), R5 | Monolith `docs/api.md` | Each submodule: `README.md` + godoc + CHANGELOG; `docs/api.md` stays for the shell |
| Tests | `tools/verify.py` 6 gates (`plans/README.md:352-365`) | Monolith gates | Per-module `go test ./...` in its module CI; keep the 6-gate suite for the shell composition |
| Air-gap distribution | R5 (embed), R15 (OCI) | Static binary, no OCI | Publish OCI image of `cgo` (R15) + offline bundle (binary + SBOM + attestations + signed checksums) for validator machines on no-network sites |
| Reproducible builds | R11 | `cgo setup` builds locally (`README.md:64-73`) | `GOFLAGS=-trimpath`, recorded toolchain (`go.mod:3`), provenance (R12) |

## 8. Slices in dependency order

| # | Slice | Delivers | Depends on | Primary source for extraction pattern |
|---|---|---|---|---|
| S-A | `model` + `metrics` submodules | Two `go.mod`s, versioned, tested independently | — | R3 (modules), R14 (semver) |
| S-B | `results` + `figures` submodules | Frozen-evidence + SVG libraries | S-A | R3 |
| S-C | `qdisc` + `remediate` submodules | tc + RouterOS libraries, dry-run `Plan` | S-A (`model`) | R9 (`tc`), R1 (RQ) |
| S-D | `probe` + `audit` (+ `cgo-audit` binary) | Probes and sealed audit as a tool | S-A (+ S-C for remediation) | R3 |
| S-E | `iface` + `profile` submodules | Sampler + registry libraries | S-A | R3 |
| S-F | `web/frontend` as versioned package; per-module servers replace monolith `pkg/api` | Frontend consumed like any other versioned dep (R5) | S-A..S-E (REST APIs) | R5 (embed) |
| S-G | Release engineering: GoReleaser, signing, SBOM, OCI, docs/lint | Anti-toy gates of §7 | S-A..S-F | R11, R12, R13, R15 |
| S-H (opportunistic) | `prometheus` exporter, `attest`, `cost` CLI, RouterOS validator, MQTT bridge (§6) | Standalone tools | S-A..S-E | R6, R12, R8 |

`pkg/doctor`, `pkg/campagne`, `pkg/api`, `internal/*` are deliberately **not** extracted — they are the shell that composes the modules (see §5). Extracting them would duplicate orchestration without a second consumer.

## 9. Open decisions requiring human approval

1. **Submodule vs separate repos.** Nested `go.mod`s (R3) keep one Git repo; separate repos simplify publishing and CODEOWNERS. Needs human choice before S-A — affects every later slice.
2. **RouterOS CAKE facts.** §1 fetches for RouterOS/CAKE returned 404s. Before building the validator (§6.5), someone must supply the primary RouterOS v7 page or grant QEMU/CHR access (slice 16 blocker, `plans/README.md:44`).
3. **WIS 2.0 / CAP alignment.** The SLA paragraph in slice 13 (`plans/README.md:229`) has no fetched primary source (`warnings.wmo.int` DNS failure, §1). Decide whether CGO claims WIS 2.0 conformance or drops it.
4. **MQTT scope.** Publishing audit results via Mosquitto (R8) is the largest "new capability" in §6. Confirm it's wanted before S-H, or leave it documented as out of scope.
5. **Release authority.** GoReleaser signing keys, SBOM policy, and SLSA provenance (R11, R12) need a named owner — decide who holds the keys.
6. **Scope guard.** Per the active slice rule (`AGENTS.md` §Cadence), at most 3 active slices; S-A..S-H above is the full map — the human picks which 3 start first.

---

*Evidence base: repo files cited inline (`go.mod`, `README.md`, `docs/architecture.md`, `docs/api.md`, `docs/superpowers/plans/README.md`); primary docs R1–R17; 404/DNS failures tabled in §1. No marketing numbers; no unsourced claims.*
