# pkg/metrics economics (CostARPerH, tiers, enterprise cost) as standalone tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/metrics/cost.go` + cost tiers in `pkg/metrics/metrics.go:60-156` + `docs/data-prices.md` + `docs/api.md` cost endpoints + Prometheus exposition in `pkg/api/metrics.go`. Research only — no code, no commit.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status 2026-11-07 | Used for |
|---|---|---|---|---|
| R1 | Yas Madagascar — Forfaits mobiles → Internet (officiel) | https://www.yas.mg/particulier/forfaits-mobiles/internet/ | fetched (price+volume+validity in page) | Retail grid |
| R2 | Yas Madagascar — Internet et connectivité, Grandes Entreprises (officiel) | https://www.yas.mg/entreprise/large-business/internet-et-connectivite/ | fetched | GOLD / PLATINE / TELMA NET |
| R3 | Yas Madagascar — La Fibre (officiel) | https://www.yas.mg/particulier/internet-fixe/fibre/ | fetched, extracted text = marketing + phones, **no fibre price/FUP** in this pass | FTTH re-check → unverified — primary source not found |
| R4 | Airtel Madagascar — databundle (officiel) | https://www.airtel.mg/internetservice/databundle | fetch returned no output this pass | Airtel values → unverified — primary source not found |
| R5 | Orange Madagascar — Be (officiel) | https://www.orange.mg/fr/be.html | fetched (volumes/detail amounts visible; "Prix" blank) | Be volumes only; price → unverified — primary source not found |
| R6 | Orange Madagascar — Akama et Be Connect (officiel) | https://www.orange.mg/fr/akama-et-be-connect.html | fetched (names+volumes+bonus+validity visible; "Prix" blank) | Be Connect volumes/bonus; price → unverified — primary source not found |
| R7 | Prometheus — Exposition formats | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched | Text-format contract |
| R8 | Go Modules Reference — go.mod file reference | https://go.dev/doc/modules/gomod-ref | fetched | Module path, `go` directive, versions, retract |
| R9 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | Declared public API, MAJOR/MINOR/PATCH |
| — | APN corporate 150 Ar/Go (devis opérateur) | — | no quote document in repo, no public grid found | unverified — primary source not found |

Local sources read in full: `pkg/metrics/cost.go` (251 lines), `pkg/metrics/metrics.go` (219 lines, cost part 60-156), `pkg/metrics/cost_test.go`, `pkg/metrics/economics_numbers_test.go`, `pkg/metrics/tier_test.go`, `pkg/metrics/metrics_test.go`, `docs/data-prices.md` (176 lines), `docs/api.md:23-25`, `pkg/api/server.go:417-475`, `pkg/api/metrics.go:15-64`, `pkg/campagne/campagne.go:707-712`, `pkg/campagne/valuation.go`, `pkg/audit/report.go:333`, `go.mod:1-3`.

## 1. Responsibility

`pkg/metrics/cost.go:1-251` answers one question: **what does wasted bandwidth cost, in Ariary, under which contract?** Its header states the doctrine: tariffs are contractual scenarios sourced in `docs/data-prices.md`, not measurements (`pkg/metrics/cost.go:8-9`). Concretely it does three things:

- **Model A — dedicated CIR depreciation.** Two Yas Grande Entreprise plans (`pkg/metrics/cost.go:20-24`) price capacity waste hourly (`pkg/metrics/cost.go:36-41`) and per-second (`pkg/metrics/cost.go:28-30`). A dedicated link does not bill the byte; it bills the capacity bufferbloat makes unusable (`pkg/metrics/cost.go:32-35`).
- **Model B — metered per-GB waste.** Two B2B tariffs (`pkg/metrics/cost.go:52-57`) price wasted bytes against a measured load window (`pkg/metrics/cost.go:67-73`), plus a retail emergency penalty (`pkg/metrics/cost.go:62`).
- **Model C — CAP SLA.** 16 KiB loaded completions judged against a 220 ms / 95 % DGM operational objective (`pkg/metrics/cost.go:80-112`), explicitly **not** a WMO mandate (`pkg/metrics/cost.go:75-79`).
- **Whole-audit assembly.** `BuildEconomics` (`pkg/metrics/cost.go:192-238`) emits all five scenarios with a primary selected by declared link type, and refuses to fabricate any number without a measured/estimated retransmit ratio (`pkg/metrics/cost.go:201-203`).

The retail-tier half of the cost surface lives in `pkg/metrics/metrics.go:60-156`: six `PriceTier`s and `CostARPerH`/`CostARPerHTier` with a hardcoded ×20 (`pkg/metrics/metrics.go:153-156`). It is exposed beside the enterprise catalogue through `docs/api.md` endpoints `GET /api/cost/tiers`, `POST /api/cost/tier`, `GET /api/cost/enterprise` (`pkg/api/server.go:417-475`). What the module deliberately does **not** do: probe, shape, store, or convert currency — there is no FX code in `cost.go`, and that absence is a feature (§4.4).

Out of scope for a cost extraction: percentile/summary, bufferbloat grading, verdict cascade, VoIP R, inference helpers (`pkg/metrics/metrics.go:13-58,158-219`, `pkg/metrics/bufferbloat.go`, `pkg/metrics/verdict.go`, `pkg/metrics/stats.go`). Those are the sibling `pkg/metrics` module (see `docs/research/2026-11-07-module-metrics.md`); `cgo-cost` below extracts only the economics surface.

## 2. Exported API (file:line)

### 2.1 `pkg/metrics/cost.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `EnterprisePlan` | struct | `pkg/metrics/cost.go:12-17` | `Name`, `MonthlyOpExMGA`, `CIRMbps`, `FUPBytes` (JSON). CIR and FUP are exclusive by construction. |
| `PlanGOLD`, `PlanPLATINE` | vars | `pkg/metrics/cost.go:20-24` | 1 000 000 Ar/mo, 200 Mbps CIR; 2 500 000 Ar/mo, FUP 30 TiB. Re-confirmed live R2 this pass (§3). |
| `AmortizedARPerSecond(monthlyOpEx)` | func | `pkg/metrics/cost.go:28-30` | Contractual month = 720 h (`pkg/metrics/cost.go:26-27`). |
| `CapacityWasteARPerHour(monthlyOpEx, wasted, transported)` | func | `pkg/metrics/cost.go:36-41` | Model A. `transported <= 0` → 0. |
| `B2BTariff` | struct | `pkg/metrics/cost.go:44-50` | `Name`, `ARPerGB`, `Quoted` (operator quote, not public grid, `pkg/metrics/cost.go:47-49`). |
| `TariffWholesale`, `TariffCorporateAPN` | vars | `pkg/metrics/cost.go:52-57` | 81.38 Ar/Go derived from PLATINE; 150 Ar/Go `Quoted:true` — unverified — primary source not found (§3). |
| `RetailEmergencyARPerGB` | const | `pkg/metrics/cost.go:62` | 5556 = 25 000 / 4.5 (Yas Net Month 4.5 Go, R1-confirmed §3). Penalty-only per comment (`pkg/metrics/cost.go:59-61`). |
| `MeteredWasteARPerHour(wasted, tariff, windowSeconds)` | func | `pkg/metrics/cost.go:67-73` | Model B. `windowSeconds <= 0` → 0. ×20 is the 180 s-window case (`pkg/metrics/cost.go:64-66`). |
| `CAPDeadlineMs`, `CAPBreachThresholdPct` | consts | `pkg/metrics/cost.go:80-83` | 220 / 95.0. DGM objective, not WMO (`pkg/metrics/cost.go:75-79`). |
| `CAPSLA`, `AssessCAPSLA(completions)` | struct+func | `pkg/metrics/cost.go:88-112` | Empty → `Measured:false` (`pkg/metrics/cost.go:99-102`). Breach iff pct strictly < 95 (`pkg/metrics/cost.go:105`). |
| `RetransRatio` | struct | `pkg/metrics/cost.go:117-120` | `Ratio` + `Basis` (`measured`/`estimated`/`""`=unmeasured). |
| `EconomicsScenario`, `EconomicsNumbers` | structs | `pkg/metrics/cost.go:123-143` | Per-model numbers; `omitempty` keeps unused fields out of JSON (`pkg/metrics/cost.go:136-147`). |
| `Economics`, `PrimaryName(linkType)`, `EconomicsDisclaimer`, `BuildEconomics(...)` | structs+funcs+const | `pkg/metrics/cost.go:159-238` | Primary map: 4g/5g→APN, vsat→wholesale, else→GOLD (`pkg/metrics/cost.go:170-183`). No ratio/transport → zero scenarios (`pkg/metrics/cost.go:201-203`). Disclaimer French, dated (`pkg/metrics/cost.go:187`). |

Unexported (not API): `numbers` (`pkg/metrics/cost.go:148`), `primaryName` (`pkg/metrics/cost.go:170`), `humanMGA` (`pkg/metrics/cost.go:241`).

### 2.2 Retail tiers — `pkg/metrics/metrics.go` + `docs/api.md`

| Export / endpoint | Location | Notes |
|---|---|---|
| `MB`, `GB` | `pkg/metrics/metrics.go:72-75` | `1024*1024`, `1024^3` (binary; contrast decimal `1e6` in basis string `pkg/metrics/cost.go:218`, §5). |
| `PriceTier`, `Tiers` (6 entries), `DefaultTier` | `pkg/metrics/metrics.go:78-95` | yas-day-1gb 1000, yas-month-4.5gb 5556, airtel-month-4.5gb 5556, yas-month-100gb 2000, yas-ftth-100gb 490, orange-month-5gb 2000. Default = `Tiers[1]` (`pkg/metrics/metrics.go:95`). **Exported mutable slice/var.** |
| `SetDefaultTier`, `ActiveTierName`, `ActiveTier`, `TierByName` | `pkg/metrics/metrics.go:103-141` | RWMutex-guarded (`tierMu`, `pkg/metrics/metrics.go:99`); unknown tier errors, no silent default. |
| `CostARPerH`, `CostARPerHTier` | `pkg/metrics/metrics.go:146-156` | `wasted/GB * ARGB * 20` — hardcoded 3-min window extrapolation (`pkg/metrics/metrics.go:155`). |
| `GET /api/cost/tiers` | `pkg/api/server.go:417-419`, `docs/api.md:23` | Serialises `metrics.Tiers` + active name verbatim — single source, no recopy (`pkg/api/server.go:420-422`). |
| `GET /api/cost/enterprise` | `pkg/api/server.go:423-452`, `docs/api.md:25` | Plans (CIR xor FUP, never zero), tariffs (`quoted:true` surfaced), `emergency_ar_per_gb`, `primary_mapping` via `PrimaryName`, disclaimer. |
| `POST /api/cost/tier` | `pkg/api/server.go:455-475`, `docs/api.md:24` | 409 while campaign runs (tier frozen with run); unknown tier 400. |

Callers that pin behaviour: `pkg/campagne` frozen-tier (`pkg/campagne/campagne.go:707-712`, `pkg/campagne/valuation.go:76`), `pkg/audit/report.go:333`, tests `pkg/metrics/cost_test.go`, `economics_numbers_test.go`, `tier_test.go`, `metrics_test.go:26-68`.

## 3. Deps (tariff sources with dates)

### 3.1 Code deps — stdlib only

`cost.go` imports only `fmt`, `strconv` (`pkg/metrics/cost.go:3-6`); tiers import `fmt`, `math`, `sort`, `sync` (`pkg/metrics/metrics.go:5-10`). No internal `pkg/*` import. Extraction needs no `require` — `go.mod:5-10` (Fyne, Bubbletea, Lipgloss, x/sys) is untouched by this module. Module identity today: `github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1-3`); a standalone needs its own `go.mod` and path (R8: module path is the unique identifier and package-path prefix; `go` directive is a mandatory minimum since Go 1.21).

Incoming callers (must keep compiling or be shimmed): `pkg/campagne` (tier freeze + `CostARPerH`), `pkg/audit` (`BuildEconomics`, `AssessCAPSLA`), `pkg/api` (catalogue endpoints), `cmd/cgo` (TUI CAP constants).

### 3.2 Tariff sources — original relevé + live re-check 07/11/2026

| Value in code | Doc source | Original date | Live re-check 07/11/2026 |
|---|---|---|---|
| GOLD 1 000 000 Ar, 200 Mbps; PLATINE 1 2 500 000 Ar, FUP 30 To; PLATINE 2 4 000 000 Ar sans FUP; TELMA NET 20 833 Ar/4.5 Go | `docs/data-prices.md:140-142` (yas.mg Grande Entreprise) | 31/08/2026 | **CONFIRMED** — R2 lists GOLD 1 000 000 Ar 30 jours "Jusqu'à 200 MPBS NATIONAL", PLATINE 1 2 500 000 Ar "FUP 30 To", PLATINE 2 4 000 000 Ar "Sans FUP", TELMA NET 20 833/62 500/104 167/124 166 Ar. |
| `TariffWholesale` 81.38 Ar/Go | derived PLATINE: 2 500 000 / 30 720 GiB (`pkg/metrics/cost.go:22,53`) | derived 31/08/2026 | Arithmetic holds against R2-confirmed inputs (2 500 000/30 720 = 81.3802…). "Go" = 1024³ binary convention (§4.4). |
| `RetailEmergencyARPerGB` 5556; `yas-month-4.5gb` 5556 | Yas Net Month 25000: 25 000 Ar / 4.5 Go (`docs/data-prices.md:23`) | 31/08/2026 | **CONFIRMED** — R1 lists Net Month 25000 at 25 000 Ar, 4.5 Go, 30 jours. 25 000/4.5 = 5 555.6. |
| `yas-day-1gb` 1000 | Ye'low One 1 000 Ar / 1 Go / 24 h (`docs/data-prices.md:18`) | 31/08/2026 | **CONFIRMED** — R1 lists Ye'low One 1 000 Ar, 1 Go, 24H. |
| `yas-month-100gb` 2000 (`pkg/metrics/metrics.go:89`) | Net Month 200000: 200 000 Ar / 100 Go (`docs/data-prices.md:27`) | 31/08/2026 | **STALE** — R1 now lists Net Month 200000 at 200 000 Ar / **250 Go** → 800 Ar/Go. Tier name and rate wrong. |
| (not in code, affects doc §6) Net Month 75000 / 30000 | `docs/data-prices.md:24` (75 000 / 15 Go) | 31/08/2026 | **STALE** — R1 lists Net Month 75000 at 75 000 Ar / **75 Go**, and a new Net Month 30000 at 30 000 Ar / **30 Go**. Unseen by code tiers. |
| `airtel-month-4.5gb` 5556 (`pkg/metrics/metrics.go:88`) | Airtel Net M'Lay 25000 (4.5 Go) 2024 grid (`docs/data-prices.md:67`) | 16/03/2024 blog | **STALE, re-check impossible** — doc's own updated grid (`docs/data-prices.md:46-54`, 05/10/2026) gives Net Mlay-25000 = 5 Go + 2.5 Go bonus = 3 333 Ar/Go effective; live page R4 returned no output → **unverified — primary source not found** for current Airtel values. |
| `orange-month-5gb` 2000 (`pkg/metrics/metrics.go:91`) | Be Connect 5 Go ~10 000 Ar (presse + internet-mg.org, `docs/data-prices.md:87,114`) | 07/2026 presse | **Unverified** — R6 confirms names/volumes/bonus only (Be Connect 5Go: 5 Go + 300 Mo Spotify + 2.5 Go OM, 30 j; 5Go Hebdo: 5 Go + 150 Mo + 1.25 Go, 7 j); "Prix" blank → **unverified — primary source not found**. R5 Be mixed bundles are voice+data (e.g. Be 2000: 2 000 Ar credit + 100 Mo, 3 j), not pure data. |
| `yas-ftth-100gb` 490 (`pkg/metrics/metrics.go:90`) | Fibre IN 49 000 Ar / FUP 100 Go (`docs/data-prices.md:37`) | 05/10/2026 | **Re-check inconclusive this pass** — R3 extracted text has no fibre price/FUP → **unverified — primary source not found** (do not treat as confirmed; do not treat as stale either). |
| `TariffCorporateAPN` 150 Ar/Go `Quoted` | operator quote, oct. 2026 (`pkg/metrics/cost.go:55`) | oct. 2026 | **Unverified — primary source not found** (no document in repo, no public grid; code correctly labels `quoted:true`). |

Licence: MIT ("Copyright (c) 2026 Meteolink contributors", `LICENSE:3` per companion research) — extractable as-is; verify `LICENSE` at repo root before extract.

## 4. Primary-source requirements

### 4.1 Yas / Orange / Airtel tariff pages (R1–R6)

- **Yas is a publishable grid** (R1, R2 fetched with full price+volume+validity). Every Yas-derived constant must carry operator + offer name + relevé date, and cgo-cost must be re-fetchable against these URLs: the 2026-11-07 live check already invalidated one tier value (Net Month 200000) and two doc FUP/volume rows (Net Month 75000, 30000). Tariff staleness is not hypothetical — it happened within this pass.
- **Airtel and Orange Be Connect are JS-gated** (R4 no output; R5/R6 "Prix" blank with volumes present). Until a values-bearing fetch exists, any Airtel/Orange price in cgo-cost is **unverified — primary source not found** and must be excluded from defaults. Orange mixed *Be* bundles (R5) are voice+data: dividing the full bundle price by data volume alone (doc §3bis, 20 000–40 000 Ar/Go effectifs) overstates waste cost 4–7× versus the pure-data reference — a pure-data calculator must not ingest them as data tariffs without a voice/SMS valuation.
- The APN 150 Ar/Go corporate tariff, being a negotiated quote, must stay `Quoted:true` with source/date fields, and must never be presented as a public grid. A 404/JS change on an operator page must downgrade affected tiers to unverified, never silently keep them green.

### 4.2 Prometheus exposition (R7)

- Since Prometheus 2.0, **all processes exposing metrics must use a text format by default** (R7, intro). The text format is line-oriented, LF-separated, the last line must end with LF, empty lines are ignored (R7, "Prometheus Text Format"). Lines with `#` are comments unless the first token is `HELP`/`TYPE`; only one HELP and one TYPE per metric name; TYPE **must appear before the first sample** of that name; absent TYPE means `untyped` (R7, "Comments, help text, and type information").
- Sample grammar: `metric_name_or_labels value [timestamp]`; `value` is Go `ParseFloat` syntax and **`NaN`, `+Inf`, `-Inf` are valid values** (R7, "The remaining lines describe samples"); timestamp is int64 milliseconds since epoch (R7). All lines for one metric form a single group, HELP/TYPE first, reproducible sorting preferred (R7, "Grouping and sorting"); **each line must have a unique metric-name+labels combination, otherwise ingestion behaviour is undefined** (R7).
- Histogram/summary conventions: `_sum` and `_count` companion samples, `{quantile="y"}` per summary quantile, `{le="y"}` per histogram bucket with a mandatory `{le="+Inf"}` equal to `_count`, buckets/quantiles in increasing order (R7, "Histograms and summaries").
- Consequence: `pkg/metrics` today emits **no exposition at all** — the only builder is `pkg/api/metrics.go:17-41`, hand-rolled, which already deviates from R7 (no timestamps, no `\`/`"`/LF escaping of label values, always emits `{...}` even for label-less samples, no `_sum`/`_count`/quantile conventions, counter/gauge inferred from `_total` suffix, and its samples are stub audit flags `pkg/api/metrics.go:47-52`, not cost metrics). A standalone cgo-cost that claims Prometheus compatibility must be written against R7 rules above, not by copying `pkg/api/metrics.go`. OpenMetrics (1.0/2.0 draft, `# EOF` trailer) is the alternative target (R7, "OpenMetrics Text Format") — the module must pick one and declare it.

### 4.3 Go modules / semver (R8, R9)

- A module's **module path is its unique identifier combined with its version, and the prefix of every package path in it** (R8, §"module"); v2+ paths must end in `/v2` (R8, §"module-path"). Today's `github.com/Realms4239/cgo` (`go.mod:1`) is a monorepo path — a standalone `cgo-cost` needs a new path it controls (R8, "Notes" under `module`), chosen before first publish to avoid a later rename.
- Software using SemVer **MUST declare a public API** in code or documentation, precisely and comprehensively (R9, spec §1). Versioning rules: PATCH for backward-compatible bug fixes, MINOR for backward-compatible functionality (and deprecations), MAJOR for incompatible API changes (R9, spec §§6–8); **once released, a version's contents MUST NOT be modified** (R9, spec §3); 0.y.z MAY change and is not stable (R9, spec §4) — so a pre-1.0 extraction may break API freely but must say so, and the 1.0.0 release is the moment the §6 table becomes a compatibility promise (R9, spec §5).
- Go versions are `vX.Y.Z` git tags, **not** SemVer strings (`"v1.2.3" is not a semantic version` — R9, FAQ); the `require` form is `module-path v1.2.3` or a Go-generated pseudo-version (R8, §"require"). Tags must be immutable once pushed, or the checksum database will flag mismatches (R8, §"retract"); `retract` is the only sanctioned way to withdraw a published version.

### 4.4 Ariary units & byte conventions (code-primary: `pkg/metrics/metrics.go:72-75`, `pkg/metrics/cost.go:22-23`)

- Code defines `GB = 1024³` (`pkg/metrics/metrics.go:72-75`) and `FUPBytes = 30·1024⁴` (`pkg/metrics/cost.go:23`), recovers the wholesale rate as "2 500 000 / 30 720 GiB" (`pkg/metrics/cost.go:22`) and calls the result "Ar/Go". The operators' "Go" is convention-loaded: the 81.38 derivation assumes operator Go = 1024 Mo. If an operator's Go were 1000 Mo, every Ar/Go shifts 2.4 %. cgo-cost must state its convention per field ("Go = 1024³, as used by the PLATINE derivation") and never mix the decimal `wastedBytes/1e6` basis-string usage (`pkg/metrics/cost.go:218`) with binary divisions without labelling.
- **No FX in the calculator.** The doc's USD/MGA reference (4 335 Ar, 31/08/2026, `docs/data-prices.md:4`) is a dated comparison, not a calculator input — there is no FX code in `cost.go`. cgo-cost must price in Ar only; any USD view is a dated, optional, caller-side conversion, never a hardcoded constant. (Live FX rate: unverified — primary source not found in this pass; do not hardcode.)

## 5. Gap analysis (current → releasable)

Ordered by severity. Local `file:line` for code; R# for primaries.

1. **Live tariff drift already present.** `yas-month-100gb` (2 000 Ar/Go) contradicts R1 (250 Go → 800 Ar/Go); `airtel-month-4.5gb` (5 556) contradicts the doc's own 05/10/2026 grid (3 333 effective) and the live page returned no output (R4 → **unverified — primary source not found**); Net Month 75000/30000 volumes moved (R1). Any tier shipped without a dated provenance string is a latent lie.
2. **Global mutable exported pricing.** `Tiers`/`DefaultTier` (`pkg/metrics/metrics.go:85-95`) and `Plan*`/`Tariff*` (`pkg/metrics/cost.go:20-57`) are exported vars any importer can rewrite, racing the `tierMu` guard (`pkg/metrics/metrics.go:99`). A library must not let one importer reprice another's scenarios.
3. **Two cost paths, two window conventions.** Legacy `CostARPerHTier` hardcodes ×20 / 3-min window (`pkg/metrics/metrics.go:153-156`) while `MeteredWasteARPerHour` takes an explicit measured `windowSeconds` (`pkg/metrics/cost.go:67-73`). The campaign keeps ×20 for historical continuity (`pkg/campagne/campagne.go:707-712`); the audit passes its measured window — but a standalone tool presenting both as canonical invites a 20× error. One path must go.
4. **Quoted vs public conflation risk.** APN 150 is a quote (**unverified — primary source not found**, §3) yet `BuildEconomics` ranks it as the primary for all 4g/5g links (`pkg/metrics/cost.go:170-179`). The API does surface `quoted:true` (`pkg/api/server.go:440-443`), but a calculator must force the caller to opt into quoted tariffs explicitly and print the disclaimer on every output.
5. **Primary mapping is a contract guess, not a measurement.** fiber/other→GOLD, vsat→wholesale, 4g/5g→APN (`pkg/metrics/cost.go:170-183`) is a correspondence of scenarios, as the disclaimer admits (`pkg/metrics/cost.go:187`). Standalone, the primary must be caller-declared, with the mapping offered only as a Madagascar default preset.
6. **French operator strings are the API.** `EconomicsDisclaimer` (`pkg/metrics/cost.go:187`, dated Yas+APN provenance in French), CAP breach label (`pkg/metrics/cost.go:109`), and basis strings (`pkg/metrics/cost.go:209,218,226`) embed language and numbers. A library must return machine codes + structured provenance (`source`, `as_of`, `quoted`, `model`) and let the caller localise presentation.
7. **No provenance fields on tariffs.** `B2BTariff` has `Name`/`ARPerGB`/`Quoted` only (`pkg/metrics/cost.go:44-50`); `PriceTier` has `Name`/`ARGB`/`Bytes` (`pkg/metrics/metrics.go:78-82`). Neither can carry offer URL, relevé date, HT/TTC, or bonus-inclusive flag — the four facts that distinguish a verified row (§3) from stale ones. `EnterprisePlan` likewise lacks `source`/`as_of`.
8. **Mixed-bundle contamination (Orange).** `orange-month-5gb` (2 000 Ar/Go) derives from presse Be Connect (R6 JS-gated, §3), while verifiable Orange data (R5) is voice+data at 20 000–40 000 Ar/Go effectifs. Ingesting either without bundle-composition metadata is a category error.
9. **Absent-data policy is sound in `cost.go`, ragged at the edge.** `BuildEconomics` emits no scenarios and `AssessCAPSLA` returns `Measured:false` without a ratio/sample (`pkg/metrics/cost.go:99-102,201-203`) — the correct invariant (never a zero that reads "measured"). But `CapacityWasteARPerHour` and `MeteredWasteARPerHour` return bare `0` on zero-transport/zero-window (`pkg/metrics/cost.go:37-40,68-70`): a caller cannot distinguish "no waste" from "no measurement". cgo-cost needs `(value, measured bool)` or error returns and must keep the §2 no-fabricated-zero tests.
10. **No Prometheus exposition in the package** (§4.2): the only builder is hand-rolled and non-conformant (`pkg/api/metrics.go:17-41`). The module needs a purpose-built exposition function (or an explicit non-Prometheus declaration), with tests for escaping, unique metric+labels, HELP/TYPE ordering, `+Inf` bucket, LF final line (R7).

## 6. Target tool: `cgo-cost` as independent calculator repo

Proposed contract (shape is the deliverable; names may rename at 0.x). R8: new module path + own `go.mod`; R9: this table *is* the declared public API at 1.0.0.

```go
// Provenance is mandatory, not a comment.
type Tariff struct {
    Name    string  `json:"name"`             // e.g. "yas-net-month-25000"
    ARPerGB float64 `json:"ar_per_gb"`        // 25000/4.5 = 5555.56
    Quoted  bool    `json:"quoted,omitempty"` // true iff devis opérateur
    Source  string  `json:"source"`           // operator URL — empty => reject
    AsOf    string  `json:"as_of"`            // relevé date YYYY-MM-DD
    Model   string  `json:"model"`            // "A" | "B" | "emergency" — no cross-model reuse
}
type Plan struct {
    Name           string  `json:"name"`
    MonthlyOpExMGA float64 `json:"monthly_opex_mga"`
    CIRMbps, FUPBytes float64 `json:"cir_mbps,omitempty"` // xor, never both zero
    Source         string  `json:"source"`
    AsOf           string  `json:"as_of"`
}
type EconomicsNumbers struct { /* keep pkg/metrics/cost.go:134-143 semantics, binary-Go convention stated */ }
type Economics struct {
    /* scenarios + primary (CALLER-DECLARED, not inferred) + retrans ratio + disclaimer */
}

func CapacityWasteARPerHour(monthlyOpEx, wasted, transported float64) (float64, bool) // bool = measured
func MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) (float64, bool)
func AmortizedARPerSecond(monthlyOpEx float64) float64 // 720 h convention, documented
func BuildEconomics(primary TariffOrPlan, ratio RetransRatio, transported, wasted, loadSeconds float64) Economics

type DeadlineTarget struct{ DeadlineMs, TargetPct float64 }
func AssessDeadline(completions []float64, t DeadlineTarget) DeadlineResult // strict-< documented
var CAPOperationalTarget = DeadlineTarget{220, 95} // "DGM operational target — not a WMO mandate"

func DefaultMGATariffs(asOf string) (plans []Plan, tariffs []Tariff, err error)
// Dated Madagascar preset; Airtel/Orange Be Connect excluded as unverified until R4/R6 fetch succeeds.
// yas-month-100gb corrected to 250 Go / 800 Ar/Go per R1 (2026-11-07).
```

Design notes: (a) tariffs are data, never globals — no `Tiers`/`DefaultTier`/`tierMu`; (b) one metered path with explicit window (delete ×20 legacy); (c) machine codes out, presentation in the caller; (d) every refused/unknown input errors — no silent tier fallback (`pkg/api/server.go` 400/409 semantics preserved as errors); (e) JSON catalogue endpoint mirrors `GET /api/cost/enterprise` shape (`pkg/api/server.go:445-451`) so cgo can consume it unchanged; (f) CLI `cgo-cost --wasted-bytes N --transported M --window S --tariff yas-net-month-25000` prints the same `Economics` JSON, disclaimer included; (g) optional Prometheus exposition written against R7 (declare text vs OpenMetrics), never copied from `pkg/api/metrics.go`.

## 7. Slices

No slice touches code in this pass; each names its acceptance check.

1. **Slice 1 — Freeze + provenance types.** Create `cgo-cost` repo (own `go.mod`, R8), port the behaviour tests from `pkg/metrics/cost_test.go` + `economics_numbers_test.go` unchanged as `0.1.0` baseline; replace bare `Tariff`/`Plan` with §6 provenance structs; keep every existing constant behind `DefaultMGATariffs("2026-11-07")` with per-row source/date from §3. Acceptance: green tests; grep proves no exported mutable var; README declares the §6 API (R9 §1).
2. **Slice 2 — De-globalise + delete ×20.** Remove `Tiers`/`DefaultTier`/`SetDefaultTier`/`CostARPerH(×20)`; in-repo `pkg/campagne`/`pkg/audit`/`pkg/api` compile via a thin adapter holding the old global until migrated. Acceptance: race-detector clean under concurrent tariff reads; `CostARPerHTier` gone (`grep`).
3. **Slice 3 — Drift repair + verification harness.** Correct `yas-month-100gb`→250 Go (800 Ar/Go) per R1, add Net Month 30000/75000 rows per R1, Airtel tier labelled provisional pending R4, FTTH tier labelled provisional pending R3; add `scripts/check-tariffs.sh` that re-fetches R1/R2/R5/R6 and fails on drift (404/JS-gated ⇒ outputs "**unverified — primary source not found**" for that row, never keeps it green). Acceptance: harness passes live; stale values from §3 table gone.
4. **Slice 4 — Measured discipline + exposition.** `(value, measured)` returns on waste functions; primary becomes caller-declared; French sentences out (codes + provenance in, presentation in CLI/README example); quoted tariffs require explicit opt-in flag; optional `RenderText` per R7 with escaping/ordering/`+Inf`/LF tests. Acceptance: boundary tests (zero transport/window, empty completions, exactly-95 %) assert non-measured, never zero-as-measured; exposition golden tests pass against an R7-conformant parser.
5. **Slice 5 — Tag + consume.** Tag `v0.1.0` (`vX.Y.Z`, immutable, R8), migrate `pkg/campagne`/`pkg/audit`/`pkg/api/server.go:417-475` to the module (or run dual-path with the Slice-2 adapter until migration lands); `GET /api/cost/tiers` + `/api/cost/enterprise` then read from `DefaultMGATariffs` output, preserving `docs/api.md:23-25` wire shape. Acceptance: `go list -m all` shows `cgo-cost`; in-repo `go test ./...` + `make verify` green; no `replace` directives left.

Each slice is independently reviewable; Slices 3–4 change behaviour-visible semantics and each should be verifier-re-run before its `VERIFIED` row is claimed (per the repo's builder/verifier split, `AGENTS.md` Roles).

## 8. Anti-toy criteria

A cost calculator is a toy if it re-describes its own constants instead of computing against an external check. It is **not** a toy only if all hold:

1. **Independent arithmetic per family.** Wholesale 81.38 vs hand division 2 500 000/30 720 (R2 inputs); emergency 5 556 vs 25 000/4.5 (R1); GOLD hourly vs 0.1 × 1 000 000/720; ×20-equivalence at 180 s window and 90× at 40 s (existing `pkg/metrics/cost_test.go:35-51` vectors, re-based on confirmed R1/R2 inputs, not self-reference).
2. **Provenance on every number.** Tariff + source URL + relevé date + quoted flag, re-fetchable per §7 Slice 3. Any Airtel/Orange Be Connect/APN/FTTH figure failing that bar is printed as "**unverified — primary source not found**" — the exact phrase required — not omitted silently and not defaulted into a primary.
3. **No silent zero.** Absent ratio/sample/window yields `measured:false`, never a 0 Ar/h that dashboards will read as "no waste" (§5.9; mirrors `BuildEconomics` refusal `pkg/metrics/cost.go:201-203`).
4. **No global mutable pricing** (§5.2). Two importers must be unable to reprice each other; the API surface is values and functions only.
5. **A real consumer + real scrape.** At least one out-of-tree importer of `cgo-cost`, plus cgo itself serving `/api/cost/enterprise` from the module (Slice 5) with the wire shape of `docs/api.md:25` intact; if Prometheus exposition is claimed, one R7-conformant scrape succeeding (HELP/TYPE, escaping, `+Inf` bucket, LF termination) — otherwise the module is a restated constant table.
6. **Version discipline.** Declared API (R9 §1), `vX.Y.Z` tags immutable (R8), changelog in MAJOR/MINOR/PATCH terms (R9 §§6–8). A repo that edits a released tag — e.g. quietly fixing the 100 Go tier in place — is a toy regardless of code quality.

---
*Teams note: slices 1–5 are proposals in this research file only. No `pkg/metrics` source, `go.mod`, or API handler was modified in this pass; implementation waits for human approval per `AGENTS.md` ("Stop and ask" — new dependency, scope growth).*
