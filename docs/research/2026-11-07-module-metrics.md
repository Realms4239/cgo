# pkg/metrics as a standalone open-source tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/metrics` only, read as if extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — FQ-CoDel | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | What RFC 8290 does/does not normatively define (bufferbloat, AQM, CAKE status) |
| R2 | Prometheus exposition formats | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched | Text-format contract a metrics library must meet |
| R3 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | fetched | Module path/versioning, `go` directive, extraction mechanics |
| R4 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | Public-API declaration, MAJOR/MINOR/PATCH rules |
| — | ITU-T G.107 E-model (for `VoIPR`) | — | **unverified — primary source not found** (not fetched; only local comment `pkg/metrics/metrics.go:179-189` cites it) | `VoIPR` formula provenance |
| — | Jain (1984) fairness-index paper (for `JFI`) | — | **unverified — primary source not found** | `JFI` formula provenance |
| — | Mann-Whitney / Cliff's δ / Vargha-Delaney / bootstrap references | — | **unverified — primary source not found** | `stats.go` provenance |
| — | Waveform bufferbloat bands (for `BufferbloatGrade`) | — | **unverified — primary source not found** (only local comment `pkg/metrics/bufferbloat.go:6-7` and `docs/methodology.fr.md:20` cite it) | Grade bands provenance |
| — | CAKE scheduler specification | — | **unverified — primary source not found** (RFC 8290 lists CAKE only as informative ref `[CAKE]`, §10.2, and as "currently being developed" future work, §7) | `cake` label provenance |

Method: read all five `pkg/metrics/*.go` sources (not tests except to note pinned behaviour), `docs/architecture.md:16`, `docs/data-prices.md`, `go.mod:1-3`, `pkg/api/metrics.go` (the only Prometheus exposition in-repo), and grep importers of `pkg/metrics`. Every technical claim below cites `file:line` or `(R#)`. Anything else is labelled **unverified — primary source not found**.

## 1. Responsibility

Today `pkg/metrics` is the CGO **calculation core**: `docs/architecture.md:16` describes it as "`Percentile`, `Summarize`, `DeadlineOKPct`, `CostARPerH`" — that row is stale, the package now has five source files and ~40 exported identifiers (§2). `pkg/metrics/metrics.go:1-2` states its contract: "percentile, deadline et coût. Le navigateur ne recalcule rien ; Go est le seul calculateur." Concretely, across the five files it does five things:

- **Distribution summaries over samples** — percentile, median/IQR, deadline compliance (`pkg/metrics/metrics.go:13-58`, `pkg/metrics/stats.go:32-44`).
- **Bufferbloat grading of a latency delta under load** — A+..F bands, worst-of-two-directions, operator-language verdict (`pkg/metrics/bufferbloat.go:11-63`).
- **Enterprise economics scenarios (models A/B/emergency) from wasted bytes** — CIR depreciation, per-GB metered waste, retail emergency penalty, with contractual-scenario disclaimer (`pkg/metrics/cost.go:28-238`), plus a legacy retail-tier cost path (`pkg/metrics/metrics.go:85-156`).
- **CAP SLA assessment** — 16 KiB loaded completions vs a 220 ms / 95 % objective (`pkg/metrics/cost.go:80-112`).
- **Operator verdict cascade + voice-quality proxy + inference helpers** — first-wins verdict (`pkg/metrics/verdict.go:35-68`), simplified E-model R (`pkg/metrics/metrics.go:190-219`), bootstrap CI, Mann-Whitney, Cliff's δ (`pkg/metrics/stats.go:55-149`).

What it deliberately does **not** do: it applies no shaping, runs no probes, stores no data, and exposes no HTTP/Prometheus surface — shaping is `pkg/qdisc` and probes are `pkg/probe` (`docs/architecture.md:14-15`), and the only exposition builder lives in `pkg/api/metrics.go:17-41`. That negative scoping is what makes extraction plausible: the package is already dependency-free (§3) and heavily imported by `pkg/campagne`, `pkg/audit`, `pkg/api`, `pkg/results`, and `cmd/cgo` (grep `metrics\.` across `pkg/`+`cmd/`), so it is the calculation leaf of the call graph, not a hub.

The CAP objective is an **operational DGM target, not an imposed threshold**: `pkg/metrics/cost.go:75-83` says so explicitly — "Ce n'est PAS un seuil imposé par l'OMM : le Guide WIS 2.0 (WMO-No 1061) définit le « temps réel » comme « de quelques secondes à quelques minutes », sans seuil numérique par bulletin". The WMO claim itself is **unverified — primary source not found** (not fetched in this pass; only this local comment states it), and `docs/VERIFY.md:48` records the row as `RECORDED` contractual scenarios, "Not re-run as a measurement here". Any standalone module must keep that framing or it misattributes the 220 ms / 95 % numbers.

## 2. Exported API (file:line)

Complete inventory of exported identifiers in the five source files. Unexported helpers (`numbers` `pkg/metrics/cost.go:148`, `primaryName` `pkg/metrics/cost.go:170`, `humanMGA` `pkg/metrics/cost.go:241`, `sortedCopy`/`quantileSorted` `pkg/metrics/stats.go:9-29`, `lcg` `pkg/metrics/stats.go:47-52`, `tierMu` `pkg/metrics/metrics.go:99`) are not API.

### 2.1 `pkg/metrics/metrics.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Percentile(sorted, p)` | func | `pkg/metrics/metrics.go:13-26` | Linear interpolation on a **pre-sorted** slice; no sort performed, no empty/p bounds guard beyond `n==0 → 0` (`pkg/metrics/metrics.go:14-20`). Caller contract is in the parameter name only. |
| `Summary` | struct | `pkg/metrics/metrics.go:28-31` | Fields `P50, P95, Median, IQRLow, IQRHigh` (no JSON tags — unlike the cost structs). `P50` duplicates `Median` (`pkg/metrics/metrics.go:37-43`). |
| `Summarize(v)` | func | `pkg/metrics/metrics.go:34-44` | Copies, sorts, renders median + IQR + p50/p95 via `Percentile`. |
| `DeadlineOKPct(completionsMs, dueMs)` | func | `pkg/metrics/metrics.go:47-58` | Strict `< dueMs` (`pkg/metrics/metrics.go:53`): a completion exactly at the deadline counts as breach. Empty → 0. |
| `MB`, `GB` | consts | `pkg/metrics/metrics.go:72-75` | `1024*1024`, `1024^3` (binary units; note `pkg/metrics/cost.go:218` uses decimal `1e6` in a basis string — unit inconsistency, §5). |
| `PriceTier` | struct | `pkg/metrics/metrics.go:78-82` | `Name`, `ARGB`, `Bytes` with JSON tags. |
| `Tiers` | var `[]PriceTier` | `pkg/metrics/metrics.go:85-92` | Six Madagascar retail tiers (Yas/Airtel/Orange). **Exported mutable slice** — any caller can rewrite prices (§5). |
| `DefaultTier` | var `PriceTier` | `pkg/metrics/metrics.go:95` | `Tiers[1]` (`yas-month-4.5gb`). Exported mutable; the only safe readers are `ActiveTier*` under `tierMu`. |
| `SetDefaultTier(name)` | func | `pkg/metrics/metrics.go:103-113` | Exact-name match only; unknown tier errors (no silent default). |
| `ActiveTierName()` / `ActiveTier()` / `TierByName(name)` | funcs | `pkg/metrics/metrics.go:116-141` | RWMutex-guarded readers (`tierMu` `pkg/metrics/metrics.go:99`). |
| `CostARPerH(wastedBytes)` | func | `pkg/metrics/metrics.go:146-150` | Legacy path: `wasted/GB * ARGB * 20` — the `×20` is a hardcoded 3-minute window extrapolation (`pkg/metrics/metrics.go:153-156`). Superseded semantically by `MeteredWasteARPerHour` with explicit `windowSeconds` (`pkg/metrics/cost.go:67-73`); both remain exported and must not both be presented as canonical. |
| `CostARPerHTier(wastedBytes, tier)` | func | `pkg/metrics/metrics.go:153-156` | Same `×20` with explicit tier. |
| `JFI(xs)` | func | `pkg/metrics/metrics.go:164-177` | Jain's index `(Σx)²/(n·Σx²)` over **per-flow** values; `len<2` or zero sum-of-squares → 0. Formula provenance: **unverified — primary source not found**. |
| `VoIPR(oneWayDelayMs, jitterMs, lossPct)` | func | `pkg/metrics/metrics.go:190-219` | "E-model simplifié (ITU-T G.107)" per comment (`pkg/metrics/metrics.go:179-189`). Clamps delay at 400 ms (`pkg/metrics/metrics.go:194-196`), clamps R to [0,100] (`pkg/metrics/metrics.go:212-217`). **Comment/code mismatch:** the comment states `Id = Id(0) + 0.024·d + 0.11·(d−177.3)/[d−177.3+ε]` (`pkg/metrics/metrics.go:186`) but the code computes `id += 0.11 * (d - 177.3)` with no denominator (`pkg/metrics/metrics.go:198-199`). ITU provenance: **unverified — primary source not found**. |

### 2.2 `pkg/metrics/bufferbloat.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `BufferbloatGrade(deltaMs)` | func | `pkg/metrics/bufferbloat.go:11-29` | Bands A+ <5, A <30, B <60, C <200, D <400, F ≥400 ms. Comment attributes them verbatim to Waveform, consulted 2026-09-03 (`pkg/metrics/bufferbloat.go:6-7`): **unverified — primary source not found**. NaN → `""` (honest unavailable, `pkg/metrics/bufferbloat.go:12-14`); negative delta → A+ by the `<5` branch. |
| `BufferbloatWorst(downDelta, upDelta)` | func | `pkg/metrics/bufferbloat.go:35-43` | `max` of the two; NaN on one side defers to the other; both NaN → NaN. |
| `BufferbloatVerdict(grade)` | func | `pkg/metrics/bufferbloat.go:47-63` | French operator-language remediation per grade; unknown grade → `""`. Hardcoded French is the public string a standalone module would ship (§5). |

### 2.3 `pkg/metrics/cost.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `EnterprisePlan` | struct | `pkg/metrics/cost.go:12-17` | `Name`, `MonthlyOpExMGA`, `CIRMbps`, `FUPBytes` with JSON tags. |
| `PlanGOLD`, `PlanPLATINE` | vars | `pkg/metrics/cost.go:20-24` | Yas Grande Entreprise, relevé 31/08/2026 (comment `pkg/metrics/cost.go:19`); PLATINE FUP 30 TiB → 81.38 Ar/Go (`pkg/metrics/cost.go:22`). Evidence in `docs/data-prices.md:140-142`. |
| `AmortizedARPerSecond(monthlyOpEx)` | func | `pkg/metrics/cost.go:28-30` | Month = 720 h convention (`pkg/metrics/cost.go:26-27`). |
| `CapacityWasteARPerHour(monthlyOpEx, wastedBytes, transportedBytes)` | func | `pkg/metrics/cost.go:36-41` | Model A (CIR depreciation). `transportedBytes <= 0` → 0. |
| `B2BTariff` | struct | `pkg/metrics/cost.go:44-50` | `Name`, `ARPerGB`, `Quoted` (tariff on operator quote, not public grid). |
| `TariffWholesale`, `TariffCorporateAPN` | vars | `pkg/metrics/cost.go:52-57` | 81.38 Ar/Go (derived from PLATINE: `pkg/metrics/cost.go:53`) and 150 Ar/Go `Quoted:true` (operator quote, oct. 2026: `pkg/metrics/cost.go:55`). |
| `RetailEmergencyARPerGB` | const | `pkg/metrics/cost.go:62` | 5556 (Yas Net Month 4,5 Go), penalty-only per comment (`pkg/metrics/cost.go:59-61`). |
| `MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds)` | func | `pkg/metrics/cost.go:67-73` | Model B; `windowSeconds <= 0` → 0. Generalises the legacy `×20` (`pkg/metrics/cost.go:64-66`). |
| `CAPDeadlineMs`, `CAPBreachThresholdPct` | consts | `pkg/metrics/cost.go:80-83` | 220 and 95.0. DGM operational objective, not WMO-mandated (§1). |
| `CAPSLA` | struct | `pkg/metrics/cost.go:88-95` | `Measured`, `OKPct`, `Breach`, `Label`, `TargetPct`, `DeadlineMs` with JSON tags; targets travel with the verdict (`pkg/metrics/cost.go:85-87`). |
| `AssessCAPSLA(loadedSmallCompletionsMs)` | func | `pkg/metrics/cost.go:99-112` | Empty → zero `CAPSLA{}` (`Measured:false`), never a pass by default (`pkg/metrics/cost.go:97-98`). Breach iff `pct < 95` (`pkg/metrics/cost.go:105`): exactly 95 passes. Breach label embeds the DGM/WIS 2.0 framing (`pkg/metrics/cost.go:109`). |
| `RetransRatio` | struct | `pkg/metrics/cost.go:117-120` | `Ratio` + `Basis` (`"measured"` / `"estimated"` / `""` = unmeasured: `pkg/metrics/cost.go:114-116`). |
| `EconomicsScenario`, `EconomicsNumbers` | structs | `pkg/metrics/cost.go:123-143` | Scenario + its numeric basis in its own model; `omitempty` makes unused model fields disappear from JSON instead of showing as measured zeros (`pkg/metrics/cost.go:136-143`, `pkg/metrics/cost.go:145-147`). |
| `Economics` | struct | `pkg/metrics/cost.go:159-168` | Whole-audit valuation: scenarios, primary, retrans ratio, disclaimer (French, dated: `pkg/metrics/cost.go:187`). |
| `PrimaryName(linkType)` | func | `pkg/metrics/cost.go:183` | Wraps unexported `primaryName` (`pkg/metrics/cost.go:170-179`): 4g/5g → APN, vsat → wholesale, else (fiber/other/empty) → GOLD. |
| `EconomicsDisclaimer` | const | `pkg/metrics/cost.go:187` | Single-sourced French disclaimer ("Scénarios contractuels … pas des mesures"). |
| `BuildEconomics(linkType, ratio, transportedBytes, wastedBytes, loadSeconds)` | func | `pkg/metrics/cost.go:192-238` | No measured ratio or no transport → scenarios empty, no fabricated zeros (`pkg/metrics/cost.go:201-203`). Assembles A (2 plans) + B (2 tariffs) + emergency, selects primary by name (`pkg/metrics/cost.go:205-236`). |

### 2.4 `pkg/metrics/stats.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Median(xs)` / `Quartiles(xs)` / `IQR(xs)` | funcs | `pkg/metrics/stats.go:32-44` | Linear-interpolation quantile on a sorted copy — a **second implementation** of `Percentile` (`pkg/metrics/stats.go:15-29` vs `pkg/metrics/metrics.go:13-26`); empty → 0 by house JSON convention (`pkg/metrics/stats.go:16-17`). |
| `BootstrapMedianCI95(xs, resamples, seed)` | func | `pkg/metrics/stats.go:55-73` | Percentile CI of the median; Park-Miller LCG, fixed seed for stability (`pkg/metrics/stats.go:46-52`). Empty → `(0,0)`. |
| `MannWhitneyTwoSided(a, b)` | func | `pkg/metrics/stats.go:76-115` | U + two-sided p, normal approximation with continuity correction, tie-averaged ranks (`pkg/metrics/stats.go:90-114`). |
| `CliffDelta(a, b)` / `CliffInterp(d)` | funcs | `pkg/metrics/stats.go:118-149` | Cliff's δ and Vargha-Delaney magnitude thresholds (0.11/0.28/0.43, `pkg/metrics/stats.go:139-148`), French labels ("négligeable"…"grand"). All statistical provenance: **unverified — primary source not found**. |

### 2.5 `pkg/metrics/verdict.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `VerdictLevel` + `VerdictNeutral/Red/Amber/Green` | type + consts | `pkg/metrics/verdict.go:8-15` | Four operator-language levels; red has a single level with two words ("une mission manquée et une saturation sont deux urgences différentes", `pkg/metrics/verdict.go:5-7`). |
| `VerdictInput` / `Verdict` | structs | `pkg/metrics/verdict.go:19-30` | Input is facts only — grade from `BufferbloatGrade`, SLA from `AssessCAPSLA`, nothing recomputed (`pkg/metrics/verdict.go:17-18`). |
| `AssessVerdict(in)` | func | `pkg/metrics/verdict.go:35-68` | First-wins cascade: not-attempted → neutral; load-not-established → neutral; SLA breach → red; grade `""` → neutral; D/F → red; C → amber; A+/A/B (+measured SLA) → green; unknown grade → neutral, never green by default (`pkg/metrics/verdict.go:62-66`). Sentences are hardcoded French with embedded CAP numbers (`pkg/metrics/verdict.go:43-58`). |

### 2.6 Tests (pin behaviour, not API)

34 `Test*` functions across 8 test files (`bufferbloat_test.go:3`, `cost_test.go:11`, `economics_numbers_test.go:4`, `jfi_test.go:1`, `metrics_test.go:6`, `stats_test.go:5`, `tier_test.go:1`, `verdict_test.go:3`). They pin, among others: Waveform band edges and NaN honesty (`pkg/metrics/bufferbloat_test.go:24-51`), zero-window/zero-transport guards (`pkg/metrics/cost_test.go:22-49`), no-scenarios-without-ratio (`pkg/metrics/cost_test.go:121-132`), primary mapping per link type (`pkg/metrics/cost_test.go:106-113`), and never-green-without-load (`docs/VERIFY.md:63`).

## 3. Dependencies

### 3.1 Outgoing (what `pkg/metrics` needs) — stdlib only

- `pkg/metrics/metrics.go:5-10` imports `fmt`, `math`, `sort`, `sync`. `pkg/metrics/bufferbloat.go:3` imports `math`. `pkg/metrics/cost.go:3-6` imports `fmt`, `strconv`. `pkg/metrics/stats.go:4-7` imports `math`, `sort`. `pkg/metrics/verdict.go:3` imports `fmt`. All are standard-library packages — per the Go Modules Reference, stdlib import paths contain no dot in the first element and resolve locally, not by download (R3, §"Module paths"). No `require` line in `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed by this package.
- **No internal imports** either: no `pkg/model`, `pkg/probe`, or other `pkg/*` import appears in any of the five files. This is the property that makes a zero-dependency extraction possible without first untangling a dependency chain.
- The module as a whole is `github.com/Realms4239/cgo` at `go 1.25` (`go.mod:1-3`). `pkg/metrics` inherits that module path and Go version today; as a standalone it would need its own `go.mod` (R3, §"go" directive: since Go 1.21 the `go` line is a mandatory minimum, and toolchains refuse newer-than-declared modules).

### 3.2 Incoming (who needs `pkg/metrics`)

Grep `metrics\.` across `pkg/` + `cmd/` shows the package is load-bearing in: `pkg/campagne` (summaries, deadline %, JFI, VoIPR, tier freeze, economics valuation — `pkg/campagne/campagne.go:135-711`), `pkg/audit` (summaries, grades, CAP SLA, economics, verdict — `pkg/audit/audit.go:467-650`, `pkg/audit/report.go:318-361`), `pkg/api` (tiers/plans/tariffs catalogue — `pkg/api/server.go:418-449`, `/metrics` flags — `pkg/api/server_audit.go:265`), `pkg/results` (summaries, quartiles, bootstrap CI — `pkg/results/results.go:166-183`), and `cmd/cgo` (TUI verdict + CAP constants — `cmd/cgo/tui_audit.go:33-173`). Extraction must keep every one of these call sites compiling, or vendor a compatibility shim (see §6, slices).

### 3.3 Licence

`LICENSE` at repo root is MIT ("Copyright (c) 2026 Meteolink contributors", `LICENSE:3`). A standalone module inherits this cleanly only if the extracted repo keeps the same licence file — required for open-source use, listed as an anti-toy criterion (§7).

## 4. Primary-source requirements

What each primary source **legitimately** requires of a standalone metrics module — and, just as importantly, what it does not require:

### 4.1 RFC 8290 (R1) — FQ-CoDel/CAKE

- RFC 8290 is **Experimental, "not an Internet Standards Track specification"** (R1, Status of This Memo), and it normatively depends on CoDel (RFC 8289) for the AQM half (R1, §2: "we refer the reader to [RFC8289]"). The only MUST in the parameter section is the packet limit: "some packet limit MUST be enforced" (R1, §5.2.3). Default parameter values an extracted module may cite: interval 100 ms (R1, §5.2.1), target 5 ms with a low-bandwidth floor tied to MTU transmission time (R1, §5.2.2), quantum 1514 bytes (R1, §5.2.4), 1024 flows (R1, §5.2.5).
- CAKE has **no RFC**: RFC 8290 cites it only as an informative reference (R1, §10.2 `[CAKE]`) and as "currently being developed as a further refinement" (R1, §7). A module must therefore never cite RFC 8290 as authority for CAKE behaviour — CAKE provenance stays **unverified — primary source not found** in this pass.
- Critically, **RFC 8290 defines none of this package's interfaces**: no percentile, no A+..F bands, no 220 ms / 95 % objective, no JFI, no E-model R, no Ar/Go economics. Its legitimate use for the module is framing only — sojourn-time/target/interval vocabulary for explaining *why* loaded latency is measured (bufferbloat = "dark buffers", R1, Abstract; queue delay vs link busy, R1, §2). Anything stronger would be citation theatre.
- The package's actual band authority is Waveform (local comment `pkg/metrics/bufferbloat.go:6-7`), its CAP authority is a DGM operational objective (`pkg/metrics/cost.go:75-83`), its R authority is claimed as ITU-T G.107 (`pkg/metrics/metrics.go:179`) — all three are **unverified — primary source not found** at primary level in this pass, and must be labelled as such in the module README until fetched.

### 4.2 Prometheus exposition (R2)

- Since Prometheus 2.0, **all processes exposing metrics must use a text format by default** (R2, intro). The text format is line-oriented, LF-separated, the last line must end with LF, empty lines are ignored (R2, "Prometheus Text Format"). Lines with `#` are comments unless the first token is `HELP`/`TYPE`; only one HELP and one TYPE per metric name; TYPE **must appear before the first sample** of that name; absent TYPE means `untyped` (R2, "Comments, help text, and type information").
- Sample grammar: `metric_name_or_labels value [timestamp]`; `value` is Go `ParseFloat` syntax and **`NaN`, `+Inf`, `-Inf` are valid values** (R2, "The remaining lines describe samples"); timestamp is int64 milliseconds since epoch (R2). All lines for one metric form a single group, HELP/TYPE first, reproducible sorting preferred (R2, "Grouping and sorting"); **each line must have a unique metric-name+labels combination, otherwise ingestion behaviour is undefined** (R2).
- Histogram/summary conventions: `_sum` and `_count` companion samples, `{quantile="y"}` per summary quantile, `{le="y"}` per histogram bucket with a mandatory `{le="+Inf"}` equal to `_count`, buckets/quantiles in increasing order (R2, "Histograms and summaries").
- Consequence for this package: `pkg/metrics` today emits **no exposition at all** — the only builder is `pkg/api/metrics.go:17-41`, hand-rolled, which already deviates from R2 (no timestamps, no `\`/`"`/LF escaping of label values, always emits `{...}` even for label-less samples, no `_sum`/`_count`/quantile conventions, counter/gauge inferred from `_total` suffix). A standalone module that claims Prometheus compatibility must be written against R2 rules above, not by copying `pkg/api/metrics.go`. OpenMetrics (1.0/2.0 draft, `# EOF` trailer) is the alternative target (R2, "OpenMetrics Text Format") — the module must pick one and declare it.

### 4.3 Go modules / semver (R3, R4)

- A module's **module path is its unique identifier combined with its version, and the prefix of every package path in it** (R3, §"module" directive); v2+ paths must end in `/v2` (R3). Today's `github.com/Realms4239/cgo` (`go.mod:1`) is a monorepo path — a standalone `pkg/metrics` needs a new path it controls (R3, "Notes" under `module`), e.g. `<org>/meteo-metrics`, chosen before first publish to avoid a later rename.
- Software using SemVer **MUST declare a public API** in code or documentation, and it SHOULD be precise and comprehensive (R4, spec §1). Versioning rules: MAJOR for incompatible API changes, MINOR for backward-compatible functionality (and deprecations), PATCH for backward-compatible bug fixes (R4, spec §§7–9); **once released, a version's contents MUST NOT be modified** (R4, spec §3); 0.y.z MAY change and is not stable (R4, spec §4) — so a pre-1.0 extraction may break API freely but must say so, and the 1.0.0 release is the moment the §2 table becomes a compatibility promise.
- Go versions are `vX.Y.Z` git tags, **not** SemVer strings (`"v1.2.3" is not a semantic version` — R4, FAQ); the `require` form is `module-path v1.2.3` or a Go-generated pseudo-version (R3, §"require"). Tags must be immutable once pushed, or the checksum database will flag mismatches (R3, §"retract" notes).

## 5. Gap analysis (current → releasable)

Ordered by severity. Each gap cites the local evidence; the primary-source side cites R# or is marked unverified.

1. **Global mutable exported state.** `Tiers` (`pkg/metrics/metrics.go:85-92`) and `DefaultTier` (`pkg/metrics/metrics.go:95`) are exported vars any caller can overwrite without the `tierMu` lock (`pkg/metrics/metrics.go:99`); `PlanGOLD`/`PlanPLATINE`/`Tariff*` are exported vars too (`pkg/metrics/cost.go:20-57`). In a library this is a data race and a silent-pricing-corruption surface. Fix: unexported storage + constructor/registry, or frozen values. This is the single largest blocker to calling the package a releasable library.
2. **Hardcoded French operator strings are the API.** `BufferbloatVerdict` (`pkg/metrics/bufferbloat.go:47-63`), `AssessVerdict` sentences (`pkg/metrics/verdict.go:35-68`), `CliffInterp` labels (`pkg/metrics/stats.go:137-149`), `EconomicsDisclaimer` (`pkg/metrics/cost.go:187`) and the CAP breach label (`pkg/metrics/cost.go:109`) are French sentences baked into returned values, and `CAPSLA.Label` even embeds "objectif DGM … cadre WIS 2.0" (`pkg/metrics/cost.go:109`). A standalone module needs stable machine codes (`grade`, `level`, `breach`) with presentation localised by the caller, or must declare French as its presentation language in the module contract (R4 spec §1: the declared public API must be precise — a French sentence is precise but pins the audience).
3. **Madagascar pricing as package-level constants.** The six retail tiers (`pkg/metrics/metrics.go:85-92`), two enterprise plans, two B2B tariffs and the emergency rate (`pkg/metrics/cost.go:20-62`) are dated operator research (`docs/data-prices.md`, relevé 31/08/2026), and at least one is already stale against its own source doc: `airtel-month-4.5gb` at 5556 Ar/Go (`pkg/metrics/metrics.go:88`) duplicates Yas, while `docs/data-prices.md:46-54` shows Airtel's current Net Mlay-25000 now gives 5 Go + 2,5 Go bonus (3333 Ar/Go effective). The numbers are contractual scenarios, not measurements (`pkg/metrics/cost.go:8-9`, `docs/VERIFY.md:48`). A standalone module must inject tariffs as caller-supplied data (or ship them as an explicitly dated, opt-in dataset) rather than posing as universal constants.
4. **No Prometheus exposition in the package.** The only exposition builder is hand-rolled in `pkg/api/metrics.go:17-41` and deviates from R2 (no timestamps, no label-value escaping of `\`, `"`, LF required by R2 "escaped_string", no `_sum`/`_count`, no quantile/bucket conventions, always-braced samples). R2 requires a unique metric+labels per line and TYPE-before-first-sample; the current builder mostly satisfies grouping/sorting (it sorts metric names `pkg/api/metrics.go:19-23` and label parts `pkg/api/metrics.go:36`) but meets none of the hard escaping/typing rules robustly. The module needs a purpose-built `Exposition` function written against R2, with tests covering escaping, `+Inf` bucket, and HELP/TYPE ordering.
5. **`VoIPR` comment/code mismatch + unverified ITU provenance.** Comment formula includes `0.11·(d−177.3)/[d−177.3+ε]` (`pkg/metrics/metrics.go:186`); code implements `0.11*(d-177.3)` (`pkg/metrics/metrics.go:198-199`). Either the comment or the code is wrong, and neither is verified against ITU-T G.107 (**unverified — primary source not found**). For a standalone tool this is a correctness-and-trust gap: re-derive from G.107, fix one side, and pin with test vectors from the spec (the existing `pkg/metrics/metrics_test.go:70-86` vectors are self-referential, not spec-derived).
6. **Inconsistent "absent" policy (`0` vs `""` vs NaN).** `Percentile`/`Summarize`/`Median`/bootstrap return 0 on empty (`pkg/metrics/metrics.go:16-17`, `pkg/metrics/stats.go:16-17`, `pkg/metrics/stats.go:56-58`), `BufferbloatGrade` returns `""` on NaN (`pkg/metrics/bufferbloat.go:12-14`), `BufferbloatWorst` propagates NaN almost-cleanly (`pkg/metrics/bufferbloat.go:35-43`), `AssessCAPSLA` returns `Measured:false` (`pkg/metrics/cost.go:99-102`). R2 makes NaN a *valid* exposition value (R2, "value is a float … NaN … valid"), so conflating absent with measured-zero corrupts downstream scrapes. The module must pick one policy (recommend: explicit `Measured` flags / error returns, never silent 0) and document it per function.
7. **Two quantile implementations, two cost paths, two unit conventions.** `Percentile` (`pkg/metrics/metrics.go:13-26`) vs `quantileSorted` (`pkg/metrics/stats.go:15-29`); legacy `CostARPerH` with hardcoded `×20` (`pkg/metrics/metrics.go:155`) vs `MeteredWasteARPerHour` with explicit window (`pkg/metrics/cost.go:67-73`); binary `GB = 1024³` (`pkg/metrics/metrics.go:74`) vs decimal `wastedBytes/1e6` in a basis string (`pkg/metrics/cost.go:218`). Each duplicate is a drift surface; the target API (§6) keeps one of each.
8. **`Percentile` caller contract is implicit.** It takes a pre-sorted slice (`pkg/metrics/metrics.go:13`) but never checks sortedness; `Summarize` sorts its own copy (`pkg/metrics/metrics.go:34-36`) so the trap only bites direct `Percentile` callers (today: `pkg/results/results.go:166-171` only via `Summarize`, but a library user gets no guard). Either sort inside (document the cost) or expose `Sorted` in the name/docs and validate in tests.
9. **Boundary semantics are unpinned in docs.** `DeadlineOKPct` is strictly `<` (`pkg/metrics/metrics.go:53`) so exactly-220 ms breaches; `AssessCAPSLA` breaches iff `pct < 95` (`pkg/metrics/cost.go:105`) so exactly 95 % passes; `JFI` returns 0 for `len<2` (`pkg/metrics/metrics.go:165-167`) which reads as "maximally unfair" rather than "undefined". These are correct-or-not design choices, but a public API (R4 spec §1) must state them.
10. **No module identity, no README, no timestamped data provenance.** Today the package is `github.com/Realms4239/cgo/pkg/metrics` on `go 1.25` (`go.mod:1-3`) with no package doc beyond two comment lines (`pkg/metrics/metrics.go:1-2`). A standalone release needs its own `go.mod`, a README declaring the public API (R4), the exposition format choice (R2), the French-language presentation decision (§5.2), and the dated tariff provenance (redirect to `docs/data-prices.md`-equivalent evidence).

## 6. Target module API

Proposed contract for a standalone module (illustrative names; the shape is the deliverable, not the spelling). Per R4, this declaration **is** the public API once published, so every item below is a 1.0.0 compatibility commitment (R4, spec §§5–8). Per R3, it would live behind a new module path with its own `go.mod` and `go` directive; pre-1.0 (`0.y.z`) it MAY change (R4, spec §4).

```go
// Samples — one canonical summary path (kills gap §5.7/§5.8).
type Summary struct { N int; P50, P95 float64; IQR float64 } // JSON-tagged; Q1/Q3 optional
func Summarize(xs []float64) Summary                 // sorts a copy; N==0 => Measured=false, never a bare 0? (gap §5.6)
func PercentileSorted(sorted []float64, p float64) float64 // name states the contract; or drop it and keep only Summarize

// Deadline — explicit, parameterised (220/95 become caller values, not constants).
type DeadlineTarget struct { DeadlineMs, TargetPct float64 }
type Result struct { Measured bool; OKPct float64; Breach bool; Target DeadlineTarget }
func AssessDeadline(completionsMs []float64, t DeadlineTarget) Result // strict-< boundary documented (gap §5.9)
var CAPTarget = DeadlineTarget{220, 95} // the DGM operational objective, opt-in, labelled "operational target — not a WMO mandate"

// Bufferbloat — machine codes out, language in the caller (gap §5.2).
type Grade string; const (GradeAPlus Grade = "A+"; GradeA = "A"; GradeB = "B"; GradeC = "C"; GradeD = "D"; GradeF = "F"; GradeUnknown Grade = "")
func GradeDelta(deltaMs float64) Grade          // NaN => GradeUnknown; bands in a documented table
func WorstDelta(down, up float64) float64       // NaN-aware (already correct: pkg/metrics/bufferbloat.go:35-43)
// BufferbloatVerdict(grade) string: DELETE from core; move example strings to an optional subpackage or docs.

// Economics — tariffs injected, never global (gaps §5.1/§5.3).
type Tariff struct { Name string; ARPerGB float64; Quoted bool; Source string; AsOf string } // provenance fields required
type Plan struct { Name string; MonthlyOpExMGA float64; CIRMbps, FUPBytes float64 }
func MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) float64 // keep (pkg/metrics/cost.go:67-73)
func CapacityWasteARPerHour(monthlyOpEx, wasted, transported float64) float64            // keep (pkg/metrics/cost.go:36-41)
func DefaultMGATariffs() (plans []Plan, tariffs []Tariff) // the dated dataset as data, behind an explicit constructor
// DELETE: Tiers/DefaultTier/SetDefaultTier/ActiveTier*/CostARPerH(×20). Callers build a Tariff and use MeteredWaste.

// Verdict — levels + codes; sentences belong to the app (gap §5.2).
type Level int; const (LevelUnknown Level = iota; LevelRed; LevelAmber; LevelGreen)
type Verdict struct { Level Level; Code string } // "no-load" | "load-not-established" | "sla-breach" | "saturated" | "fragile" | "healthy" | "unavailable"
func Assess(input Input) Verdict // same first-wins cascade (pkg/metrics/verdict.go:35-68), no French strings

// Voice — after re-derivation (gap §5.5).
func VoIPR(oneWayDelayMs, jitterMs, lossPct float64) float64 // only once the G.107 formula is fixed and spec-vector-pinned

// Inference — keep, but document provenance or mark experimental.
func BootstrapMedianCI95(...); func MannWhitneyTwoSided(...); func CliffDelta(...); func CliffInterp(...) float64 // return a magnitude, not a French label

// Exposition — NEW, written against R2 (gap §5.4).
type Sample struct { Metric string; Labels map[string]string; Value float64; TimestampMs int64 }
func RenderText(samples []Sample) (string, error) // HELP/TYPE first, unique metric+labels enforced, escaping per R2, LF-terminated; or declare OpenMetrics + `# EOF`
```

Design notes: (a) machine-readable codes everywhere a sentence exists today; (b) data injected, never module-global; (c) one quantile path, one cost path, binary-vs-decimal unit stated per field; (d) every "absent" returns a flag or error, never a silent 0 (R2 valid-NaN noted as an option the API must choose, not drift into).

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it.

1. **Slice 1 — Freeze + declare.** Snapshot the current exported API (§2) into the module README as the `0.1.0` declared public API (R4 spec §1); create the new `go.mod` (R3); keep `LICENSE` (MIT, `LICENSE:3`); port all 34 existing tests unchanged (they pin behaviour today, §2.6). Acceptance: `go test ./...` green on the extracted tree; README lists every §2 identifier.
2. **Slice 2 — De-globalise.** Replace `Tiers`/`DefaultTier`/`tierMu` (`pkg/metrics/metrics.go:85-141`) and exported plan/tariff vars (`pkg/metrics/cost.go:20-57`) with injected `Tariff`/`Plan` values + `DefaultMGATariffs()` constructor; delete `SetDefaultTier`/`ActiveTier*`; keep in-repo callers (`pkg/campagne`, `pkg/audit`, `pkg/api`) compiling via a thin adapter that holds the old global. Acceptance: no exported mutable var remains (`go vet`-visible + grep); race detector clean under concurrent tier reads/writes (`pkg/campagne/tier_freeze_test.go` scenario).
3. **Slice 3 — Unify + pin boundaries.** One quantile path, one cost path, one unit convention; document/test strict-`<` deadline, `<95` breach, `JFI len<2` semantics; fix `VoIPR` comment/code mismatch against a fetched G.107 (or mark the function experimental in docs until then). Acceptance: boundary tests for exactly-220 ms, exactly-95 %, NaN/empty inputs; `VoIPR` matches spec-derived vectors.
4. **Slice 4 — Machine codes + exposition.** Replace all returned sentences with codes (§6); add `RenderText` (or OpenMetrics) written against R2 with tests for escaping, unique-metric+labels, HELP/TYPE-before-sample, `+Inf` bucket, LF final line; keep French strings only in a docs example or a `fr` subpackage. Acceptance: exposition golden tests pass against an R2-conformant parser; no French string returned by any core function.
5. **Slice 5 — First tag.** Resolve module path (R3), tag `v0.1.0`, publish; `pkg/campagne`/`pkg/audit`/`pkg/api`/`pkg/results` migrated to the new module version in-repo (or run dual-path with the adapter from Slice 2 until migration lands). Acceptance: `go list -m all` shows the new module; in-repo `go test ./...` + `python tools/verify.py` green; no `replace` directives left in `go.mod`.

Each slice is independently reviewable; Slices 3–4 are the only ones that change behaviour-visible semantics and each should be verifier-re-run before its `VERIFIED` row is claimed (per the repo's builder/verifier split, `AGENTS.md` Roles).

## 8. Anti-toy criteria (what would make this a toy — and the bar for "not a toy")

A metrics module is a toy if it re-describes its own outputs instead of measuring or computing against an external check. This extraction is **not** a toy only if all of the following hold:

1. **Independent verification per number family.** Percentiles vs a reference implementation on known vectors; CAP/deadline vs hand-computed boundary cases (exactly-220 ms, exactly-95 %); `JFI` vs known equal/unequal flow vectors; economics vs hand arithmetic from `docs/data-prices.md` rows (e.g. PLATINE 2 500 000 / 30 720 GiB = 81.38, `pkg/metrics/cost.go:22`); `VoIPR` vs ITU-T G.107 text once fetched (until then it **is** a toy in the trust sense — self-referential tests at `pkg/metrics/metrics_test.go:70-86` do not count).
2. **Provenance honesty.** Every constant carries source + date (Yas/Airtel/Orange rows point at `docs/data-prices.md` URLs; CAP 220/95 labelled "DGM operational objective, not a WMO mandate" per `pkg/metrics/cost.go:75-83`; waveform bands and G.107/Jain/statistical claims labelled **unverified — primary source not found** until the primaries are fetched — the exact phrase required by the brief, used above for each). A module that presents a French sector tariff or a 2026 operator quote as a universal constant is a toy.
3. **No silent zero.** Absent data is `Measured:false` / error / (if chosen) NaN — never a 0 that a dashboard or Prometheus scrape will read as a measurement (see gap §5.6; R2 makes the NaN-vs-0 choice load-bearing).
4. **No global mutable state.** Library users must not be able to rewrite each other's prices (gap §5.1). If two importers can corrupt each other's results through package-level vars, it is an app helper, not a module.
5. **A real consumer + real scrape.** At least one out-of-tree importer (not only `pkg/campagne`/`pkg/audit` in this repo) and one R2-conformant scrape succeeding against the module's own exposition (HELP/TYPE, escaping, `+Inf` bucket, LF termination) — otherwise the Prometheus claim is decoration.
6. **Version discipline.** A declared API (R4 spec §1), tags `vX.Y.Z` treated as immutable (R3 §"retract"), and a changelog phrased in MAJOR/MINOR/PATCH terms (R4 specs §§7–9). A repo that edits a released tag, or ships no declared API, is a toy regardless of code quality.

---
*Teams note: slices 1–5 are proposals in this research file only. No `pkg/metrics` source was modified in this pass; implementation waits for human approval per `AGENTS.md` ("Stop and ask" — new dependency, scope growth).*
