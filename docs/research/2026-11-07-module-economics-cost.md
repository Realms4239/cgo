# Economics & cost module as standalone tool (cgo-cost) — primary-source research

*Date: 2026-11-07 · Scope: `pkg/metrics/cost.go` + cost tiers (`pkg/metrics/metrics.go:60-156`) + `docs/data-prices.md` + `docs/api.md` cost endpoints. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status 2026-11-07 | Used for |
|---|---|---|---|---|
| R1 | Guide to the WMO Information System (WMO-No. 1061), Vol. II — WIS 2.0 (APPROVED v1.3.0, 2026-06-15) | https://wmo-im.github.io/wis2-guide/guide/wis2-guide-APPROVED.html | fetched | WIS 2.0 "real time" definition; CAP framing |
| R2 | Yas Madagascar — Forfaits mobiles → Internet (officiel) | https://www.yas.mg/particulier/forfaits-mobiles/internet/ | fetched | Retail grid: Net Month, Ye'low |
| R3 | Yas Madagascar — Internet et connectivité, Grandes Entreprises (officiel) | https://www.yas.mg/entreprise/large-business/internet-et-connectivite/ | fetched | GOLD / PLATINE / TELMA NET enterprise grid |
| R4 | Yas Madagascar — La Fibre (officiel) | https://www.yas.mg/particulier/internet-fixe/fibre/ | fetched | FTTH FUP grid |
| R5 | Airtel Madagascar — databundle (officiel) | https://www.airtel.mg/internetservice/databundle | fetched, body JS-gated (title only) → **unverified — primary source not found** for price values this pass | Airtel grid (doc §2 remains unre-verified) |
| R6 | Orange Madagascar — Be (officiel) | https://www.orange.mg/fr/be.html | fetched (volumes + "Crédit d'appel X Ar" detail amounts visible; "Prix" card JS) | Be mixed-bundle volumes/amounts |
| R7 | Orange Madagascar — Akama et Be Connect (officiel) | https://www.orange.mg/fr/akama-et-be-connect.html | fetched (product names only; prices/contents JS-gated) → **unverified — primary source not found** for Be Connect price values this pass | Be Connect names (doc §3/§3ter prices rest on presse) |
| R8 | Go Modules Reference | https://go.dev/ref/mod | fetched | Module path, `go` directive, version/tag discipline |
| R9 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | not re-fetched this pass (companion research fetched) | Declared public API, MAJOR/MINOR/PATCH |
| R10 | Malagasy ariary (ISO 4217 MGA, numeric 969; 1 Ar = 5 iraimbilanja; symbol Ar) | https://en.wikipedia.org/wiki/Malagasy_ariary | fetched | Ariary subunit / non-decimal status |
| R11 | XE USD→MGA mid-market | https://www.xe.com/currencyconverter/convert?Amount=1&From=USD&To=MGA | fetched (4 525,18 MGA, Oct. 2026) | FX dated comparison vs doc's 4 335 (31/08/2026) |

Local sources read in full: `pkg/metrics/cost.go`, `pkg/metrics/cost_test.go`, `pkg/metrics/metrics.go`, `pkg/api/server.go:410-475`, `docs/data-prices.md`, `docs/api.md`, `go.mod`.

## 1. Responsibility

`pkg/metrics/cost.go:1-251` answers one question: **what does wasted bandwidth cost, in Ariary, under which contract?** Its header (`pkg/metrics/cost.go:8-9`) states the doctrine: tariffs are contractual scenarios sourced in `docs/data-prices.md`, not measurements. Concretely it does three things:

- **Model A — dedicated CIR depreciation.** Two Yas Grande Entreprise plans (`pkg/metrics/cost.go:20-24`) price capacity waste hourly (`pkg/metrics/cost.go:36-41`) and per-second (`pkg/metrics/cost.go:28-30`). A dedicated link does not bill the byte; it bills the capacity that bufferbloat makes unusable (`pkg/metrics/cost.go:32-35`).
- **Model B — metered per-GB waste.** Two B2B tariffs (`pkg/metrics/cost.go:52-57`) price wasted bytes against a measured load window (`pkg/metrics/cost.go:67-73`), plus a retail emergency penalty (`pkg/metrics/cost.go:62`).
- **Model C — CAP SLA.** 16 KiB loaded completions judged against a 220 ms / 95 % DGM operational objective (`pkg/metrics/cost.go:80-112`), explicitly **not** a WMO mandate (`pkg/metrics/cost.go:75-79`; R1 confirms WIS 2.0 "real time" is "a few seconds to a few minutes", §23, with no numeric per-bulletin threshold).
- **Whole-audit assembly.** `BuildEconomics` (`pkg/metrics/cost.go:192-238`) emits all five scenarios with a primary selected by declared link type, and refuses to fabricate any number without a measured/estimated retransmit ratio (`pkg/metrics/cost.go:201-203`).

The retail-tier half of the cost surface lives in `pkg/metrics/metrics.go:60-156` (six `PriceTier`s, `CostARPerH/Tier` with a hardcoded ×20) and is exposed beside the enterprise catalogue through `docs/api.md` endpoints `GET /api/cost/tiers`, `POST /api/cost/tier`, `GET /api/cost/enterprise` (`pkg/api/server.go:417-475`). What the module deliberately does not do: probe, shape, store, or convert currency — there is no FX code in `cost.go`, and that absence is a feature (§4.4).

## 2. Exported API (file:line)

### 2.1 `pkg/metrics/cost.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `EnterprisePlan` | struct | `pkg/metrics/cost.go:12-17` | `Name`, `MonthlyOpExMGA`, `CIRMbps`, `FUPBytes` (JSON). CIR and FUP are exclusive by construction. |
| `PlanGOLD`, `PlanPLATINE` | vars | `pkg/metrics/cost.go:20-24` | 1 000 000 Ar/mo 200 Mbps CIR; 2 500 000 Ar/mo FUP 30 TiB. Both **re-confirmed live** R3 this pass (§3). |
| `AmortizedARPerSecond(monthlyOpEx)` | func | `pkg/metrics/cost.go:28-30` | Contractual month = 720 h (`pkg/metrics/cost.go:26-27`). |
| `CapacityWasteARPerHour(monthlyOpEx, wasted, transported)` | func | `pkg/metrics/cost.go:36-41` | Model A. `transported <= 0` → 0. |
| `B2BTariff` | struct | `pkg/metrics/cost.go:44-50` | `Name`, `ARPerGB`, `Quoted` (operator quote, not public grid, `pkg/metrics/cost.go:47-49`). |
| `TariffWholesale`, `TariffCorporateAPN` | vars | `pkg/metrics/cost.go:52-57` | 81,38 Ar/Go derived from PLATINE; 150 Ar/Go `Quoted:true` (devis oct. 2026 per comment — **unverified — primary source not found**: no public grid, no quote document in repo). |
| `RetailEmergencyARPerGB` | const | `pkg/metrics/cost.go:62` | 5 556 = 25 000 / 4,5 (Yas Net Month 4,5 Go, R2-confirmed, §3). Penalty-only per comment (`pkg/metrics/cost.go:59-61`). |
| `MeteredWasteARPerHour(wasted, tariff, windowSeconds)` | func | `pkg/metrics/cost.go:67-73` | Model B. `windowSeconds <= 0` → 0. ×20 is the 180 s-window case (`pkg/metrics/cost.go:64-66`). |
| `CAPDeadlineMs`, `CAPBreachThresholdPct` | consts | `pkg/metrics/cost.go:80-83` | 220 / 95.0. DGM objective, not WMO (R1 §23; `pkg/metrics/cost.go:75-79`). |
| `CAPSLA`, `AssessCAPSLA(completions)` | struct+func | `pkg/metrics/cost.go:88-112` | Empty → `Measured:false` (`pkg/metrics/cost.go:99-102`). Breach iff pct strictly < 95. |
| `RetransRatio` | struct | `pkg/metrics/cost.go:117-120` | `Ratio` + `Basis` (`measured`/`estimated`/`""`=unmeasured). |
| `EconomicsScenario`, `EconomicsNumbers` | structs | `pkg/metrics/cost.go:123-143` | Per-model numbers; `omitempty` keeps unused fields out of JSON (`pkg/metrics/cost.go:136-147`). |
| `Economics`, `PrimaryName(linkType)`, `EconomicsDisclaimer`, `BuildEconomics(...)` | structs+funcs+const | `pkg/metrics/cost.go:159-238` | Primary map: 4g/5g→APN, vsat→wholesale, else→GOLD (`pkg/metrics/cost.go:170-183`). No ratio/transport → zero scenarios (`pkg/metrics/cost.go:201-203`). |

### 2.2 Retail tiers — `pkg/metrics/metrics.go` + `docs/api.md`

| Export / endpoint | Location | Notes |
|---|---|---|
| `PriceTier`, `Tiers` (6 entries), `DefaultTier`, `SetDefaultTier`, `ActiveTierName/ActiveTier/TierByName`, `CostARPerH`, `CostARPerHTier`, `MB`, `GB` | `pkg/metrics/metrics.go:72-156` | Six tiers: yas-day-1gb 1 000, yas-month-4.5gb 5 556, airtel-month-4.5gb 5 556, yas-month-100gb 2 000, yas-ftth-100gb 490, orange-month-5gb 2 000. Default = `Tiers[1]` (`pkg/metrics/metrics.go:95`). `CostARPerHTier` hardcodes ×20 (`pkg/metrics/metrics.go:155`). |
| `GET /api/cost/tiers` | `pkg/api/server.go:417-419`, `docs/api.md:23` | Serialises `metrics.Tiers` + active name verbatim — single source, no recopy (`pkg/api/server.go:420-422`). |
| `GET /api/cost/enterprise` | `pkg/api/server.go:423-452`, `docs/api.md:25` | Plans (CIR xor FUP, never zero), tariffs (`quoted:true` surfaced), `emergency_ar_per_gb`, `primary_mapping` via `PrimaryName`, disclaimer. |
| `POST /api/cost/tier` | `pkg/api/server.go:455-475`, `docs/api.md:24` | 409 while a campaign runs (tier frozen with the run); unknown tier 400. |

## 3. Dependencies (incl. tariff sources with dates)

### 3.1 Code deps — stdlib only

`cost.go` imports only `fmt`, `strconv` (`pkg/metrics/cost.go:3-6`); `metrics.go` tiers import `fmt`, `math`, `sort`, `sync` (`pkg/metrics/metrics.go:5-10`). No internal `pkg/*` import. Extraction needs no `require` — `go.mod:5-10` (Fyne, Bubbletea, Lipgloss, x/sys) is untouched by this module. Module identity today: `github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1-3`); a standalone needs its own `go.mod` and path (R8: `go` directive is mandatory since Go 1.21).

Incoming callers (must keep compiling or be shimmed): `pkg/campagne` (tier freeze + `CostARPerH`, frozen-tier tests), `pkg/audit` (`BuildEconomics`, `AssessCAPSLA`, verdict), `pkg/api` (catalogue endpoints), `cmd/cgo` (TUI CAP constants).

### 3.2 Tariff sources — original relevé + live re-check this pass

| Value in code | Doc source | Original date | Live re-check 07/11/2026 |
|---|---|---|---|
| GOLD 1 000 000 Ar, 200 Mbps; PLATINE 1 2 500 000 Ar, FUP 30 To | `docs/data-prices.md:140-142` (yas.mg Grande Entreprise) | 31/08/2026 | **CONFIRMED** — R3 lists GOLD 1 000 000 Ar 30 jours "Jusqu'à 200 MPBS NATIONAL", PLATINE 1 2 500 000 Ar "FUP 30 To", PLATINE 2 4 000 000 Ar "Sans FUP", TELMA NET 20 833 Ar/4,5 Go. |
| `TariffWholesale` 81,38 Ar/Go | derived PLATINE: 2 500 000 / 30 720 GiB (`pkg/metrics/cost.go:22,53`) | derived 31/08/2026 | Arithmetic holds against confirmed inputs (2 500 000/30 720 = 81,3802…). "Go" = 1024³ binary convention (§4.4). |
| `RetailEmergencyARPerGB` 5 556 | Yas Net Month 25000: 25 000 Ar / 4,5 Go (`docs/data-prices.md:23`) | 31/08/2026 | **CONFIRMED** — R2 lists Net Month 25000 at 25 000 Ar, 4,5 Go, 30 jours. 25 000/4,5 = 5 555,6. |
| Tier yas-day-1gb 1 000 | Ye'low One 1 000 Ar / 1 Go / 24 h (`docs/data-prices.md:18`) | 31/08/2026 | **CONFIRMED** — R2 lists Ye'low One 1 000 Ar, 1 Go, 24H. |
| Tieryas-ftth-100gb 490 | Fibre IN 49 000 Ar / FUP 100 Go (`docs/data-prices.md:37`) | 05/10/2026 | **CONFIRMED** — R4 lists Fibre IN 49 000 Ar, FUP 100 Go (= 490 Ar/Go). |
| Tier yas-month-100gb 2 000 | Net Month 200000: 200 000 Ar / 100 Go (`docs/data-prices.md:27`) | 31/08/2026 | **STALE** — R2 now lists Net Month 200000 at 200 000 Ar / **250 Go** → 800 Ar/Go. Tier name and rate wrong. |
| Tier airtel-month-4.5gb 5 556 | Airtel Net M'Lay 25000 (4,5 Go) 2024 grid (`docs/data-prices.md:67`) | 16/03/2024 blog | **STALE** — doc's own updated grid (`docs/data-prices.md:46-54`, 05/10/2026) gives Net Mlay-25000 = 5 Go + 2,5 Go bonus = 3 333 Ar/Go effective. Live Airtel page JS-gated (R5) → re-verification impossible this pass: **unverified — primary source not found** for current Airtel values. |
| Tier orange-month-5gb 2 000 | Be Connect 5 Go ~10 000 Ar (presse + internet-mg.org, `docs/data-prices.md:87,114`) | 07/2026 presse | **Unverified** — R7 confirms product names only; prices JS-gated. Official mixed Be bundles (R6) are voice+data, not pure data (§5.3). |
| `TariffCorporateAPN` 150 Ar/Go `Quoted` | operator quote, oct. 2026 (`pkg/metrics/cost.go:55`; spec 2026-10-25 §4) | oct. 2026 | **Unverified — primary source not found** (no document in repo, no public grid; code correctly labels it `quoted:true`). |
| Yas/Orange/Airtel doc §6 tiers (daily/month pals) | `docs/data-prices.md:146-157` | 31/08/2026 | Partially stale: Yas FTTH GO/Family/Home FUPs grew (R4: GO 400 Go, Family 1 To, Home 2,5 To vs doc's 150/300/900 Go base) → doc Ar/Go 660/497/277 now 247/145/97. New offers unseen by doc: Net Month 30000 30 Go, Net Month 75000 75 Go (doc: 15 Go), Fibre ON 69 000 250 Go. |

Licence: MIT (`LICENSE`, "Copyright (c) 2026 Meteolink contributors") — extractable as-is.

## 4. Primary-source requirements

### 4.1 WMO WIS 2.0 (R1)

- The Guide (WMO-No. 1061 Vol. II, APPROVED v1.3.0, 2026-06-15) defines WIS 2.0 "real time" as *"anything from a few seconds to a few minutes — not the milliseconds required by some applications"* (R1 §23). It sets **no numeric per-bulletin latency threshold** and no 220 ms / 95 % figure. The code's framing (`pkg/metrics/cost.go:75-79`: DGM operational objective, WMO context only) is therefore the only legitimate use of R1 for Model C.
- A standalone cgo-cost **must** take deadline and target as caller parameters (e.g. `DeadlineTarget{220, 95}`), ship the DGM pair only as an explicitly labelled opt-in default ("operational target — not a WMO mandate"), and keep the Targets-on-the-verdict discipline (`pkg/metrics/cost.go:85-87`) so a sealed report states what it was judged against. Presenting 220 ms as a WMO requirement is citation theatre and fails the anti-toy bar (§8).

### 4.2 Yas / Orange / Airtel tariff pages (R2–R7)

- **Yas is a publishable grid** (R2–R4 fetched with full price+volume+validity). Every Yas-derived constant must carry operator + offer name + relevé date, and cgo-cost must be re-fetchable against these URLs: the 2026-11-07 live check already invalidated two tier values (Net Month 200000, Fibre FUPs). Tarif staleness is not hypothetical — it happened within this pass.
- **Airtel and Orange Be Connect are JS-gated** (R5, R7: fetch returns shell/names without prices). Until a values-bearing fetch exists, any Airtel/Orange price in cgo-cost is **unverified — primary source not found** and must be excluded from defaults. Orange mixed *Be* bundles (R6) are partially verifiable (volumes + detail amounts) but are **voice+data bundles**: their "Ar/Go effectifs" per doc §3bis (20 000–40 000 Ar/Go) divide the full bundle price by data volume alone. A pure-data calculator must not ingest them as data tariffs without a voice/SMS valuation — doing so overstates waste cost 4–7× versus the pure-data reference palier.
- The APN 150 Ar/Go corporate tariff, being a negotiated quote, must stay `Quoted:true` with source/date fields, and must never be presented as a public grille. A 404/JS change on the operator page must downgrade affected tiers to unverified, never silently keep them green.

### 4.3 Go modules / semver (R8, R9)

- A standalone module path is the prefix of every package path and is declared in its own `go.mod` (R8 §"module"). The monorepo path `github.com/Realms4239/cgo` (`go.mod:1`) must be replaced by a controlled path chosen **before** first publish; v2+ paths must end `/v2` (R8 §"Major version suffixes"). The `go` directive is a mandatory minimum toolchain since Go 1.21 (R8 §"go directive").
- Software using SemVer **must declare a public API** in code or documentation, precisely and comprehensively (R9 spec §1, independently verified in companion research). MAJOR = incompatible API change, MINOR = backward-compatible feature, PATCH = bug fix (R9 specs §§7–9); a released version's contents must not be modified (R9 spec §3). Go versions are `vX.Y.Z` **git tags**, not SemVer strings, and tags are immutable (R8 §"retract"). Before `v1.0.0` the API may break but must say so — the §6 table becomes a compatibility promise only at 1.0.0.

### 4.4 Ariary units (R10, R11)

- Malagasy ariary: ISO 4217 code **MGA**, numeric **969**, symbol **Ar**; notionally subdivided into **5 iraimbilanja** (1 iraimbilanja = 1/5 Ar = 20 centimes equivalent) and one of only two non-decimal circulating currencies (R10). Practically, the iraimbilanja is obsolete in circulation; the code's `float64` Ariary amounts (Ar/Go, Ar/h, transport bytes) are therefore representationally sound — no minor-unit integer handling is required, but sub-Ariary values must not be rounded to whole Ariary in scenarios.
- **Byte units are the real unit bug surface.** Code defines `GB = 1024³` (`pkg/metrics/metrics.go:72-75`) and `FUPBytes = 30·1024⁴` (`pkg/metrics/cost.go:23`), recovers the wholesale rate as "2 500 000 / 30 720 GiB" (`pkg/metrics/cost.go:22`) and calls the result "Ar/Go". The operators' "Go" is convention-loaded: the 81,38 derivation assumes operator Go = 1024 Mo. If an operator's Go were 1000 Mo, every Ar/Go shifts 2,4 %. cgo-cost must state its convention per field ("Go = 1024³, as used by the PLATINE derivation") and never mix the decimal `wastedBytes/1e6` basis-string usage (`pkg/metrics/cost.go:218`) with binary divisons without labelling.
- **No FX in the calculator.** The doc's USD/MGA reference (4 335 Ar, 31/08/2026, `docs/data-prices.md:4`) is already superseded — XE mid-market reads 4 525,18 MGA (Oct. 2026, R11). cgo-cost must price in Ar only; any USD view is a dated, optional, caller-side conversion, never a hardcoded constant.

## 5. Gap analysis (current → releasable)

Ordered by severity. Local file:line for code; R# for primaries.

1. **Live tariff drift already present.** `yas-month-100gb` (2 000 Ar/Go) contradicts R2 (250 Go → 800 Ar/Go); `airtel-month-4.5gb` (5 556) contradicts the doc's own 05/10/2026 grid (3 333 effective) and cannot be re-checked (R5 JS-gated → **unverified — primary source not found**); Fibre GO/Family/Home FUPs outgrew doc §FTTH (R4). Any tier shipped without a dated provenance string is a latent lie.
2. **Global mutable exported pricing.** `Tiers`/`DefaultTier` (`pkg/metrics/metrics.go:85-95`) and `Plan*`/`Tariff*` (`pkg/metrics/cost.go:20-57`) are exported vars any importer can rewrite, racing the `tierMu` guard (`pkg/metrics/metrics.go:99`). A library must not let one importer reprice another's scenarios.
3. **Two cost paths, two window conventions.** Legacy `CostARPerHTier` hardcodes ×20 / 3-min window (`pkg/metrics/metrics.go:153-156`) while `MeteredWasteARPerHour` takes an explicit measured `windowSeconds` (`pkg/metrics/cost.go:67-73`). The campaign keeps ×20 for historical continuity; the audit passes its measured window — but a standalone tool presenting both as canonical invites a 20× error. One path must go.
4. **Quoted vs public conflation risk.** APN 150 is a quote (**unverified — primary source not found**, §3) yet `BuildEconomics` ranks it as the primary for all 4g/5g links (`pkg/metrics/cost.go:170-179`). The API does surface `quoted:true` (`pkg/api/server.go:440-443`), but a calculator must force the caller to opt into quoted tariffs explicitly and print the disclaimer on every output.
5. **Primary mapping is a contract guess, not a measurement.** fiber/other→GOLD, vsat→wholesale, 4g/5g→APN (`pkg/metrics/cost.go:170-183`) is a correspondence of scenarios, as the disclaimer admits (`pkg/metrics/cost.go:187`). Standalone, the primary must be caller-declared, with the mapping offered only as a Madagascar default preset.
6. **French operator strings are the API.** `EconomicsDisclaimer` (`pkg/metrics/cost.go:187`, dated Yas+APN provenance in French), CAP breach label (`pkg/metrics/cost.go:109`), and basis strings (`pkg/metrics/cost.go:209,218,226`) embed language and numbers. A library must return machine codes + structured provenance (`source`, `as_of`, `quoted`, `model`) and let the caller localise presentation.
7. **No provenance fields on tariffs.** `B2BTariff` has `Name`/`ARPerGB`/`Quoted` only (`pkg/metrics/cost.go:44-50`); `PriceTier` has `Name`/`ARGB`/`Bytes` (`pkg/metrics/metrics.go:78-82`). Neither can carry offer URL, relevé date, HT/TTC, or bonus-inclusive flag — the four facts that distinguish a verified row (§3) from stale ones. `EnterprisePlan` likewise lacks `source`/`as_of`.
8. **Mixed-bundle contamination (Orange).** `orange-month-5gb` (2 000 Ar/Go) derives from presse Be Connect (R7 JS-gated, §3), while verifiable Orange data (R6) is voice+data at 20 000–40 000 Ar/Go effectifs. Ingesting either without bundle-composition metadata is a category error.
9. **Absent-data policy is sound in `cost.go`, ragged at the edge.** `BuildEconomics` emits no scenarios and `AssessCAPSLA` returns `Measured:false` without a ratio/sample (`pkg/metrics/cost.go:99-102,201-203`) — the correct invariant (never a zero that reads "measured"). But `CapacityWasteARPerHour` and `MeteredWasteARPerHour` return bare `0` on zero-transport/zero-window (`pkg/metrics/cost.go:37-40,68-70`): a caller cannot distinguish "no waste" from "no measurement". cgo-cost needs `(value, measured bool)` or error returns and must keep the §2.1 no-fabricated-zero tests.

## 6. Target tool: `cgo-cost` as independent calculator repo

Proposed contract (shape is the deliverable; names may rename at 0.x). R8: new module path + own `go.mod`; R9: this table *is* the declared public API at 1.0.0.

```go
// Provenance is mandatory, not a comment.
type Tariff struct {
    Name string  `json:"name"`            // e.g. "yas-net-month-25000"
    ARPerGB float64 `json:"ar_per_gb"`    // 25000/4.5 = 5555.56
    Quoted bool  `json:"quoted,omitempty"` // true iff devis opérateur
    Source string `json:"source"`          // operator URL (R2-…) — empty => reject
    AsOf string  `json:"as_of"`            // relevé date YYYY-MM-DD
    Model string  `json:"model"`           // "A" | "B" | "emergency" — no cross-model reuse
}
type Plan struct {
    Name string  `json:"name"`
    MonthlyOpExMGA float64 `json:"monthly_opex_mga"`
    CIRMbps, FUPBytes float64 `json:"cir_mbps,omitempty"` // xor
    Source string `json:"source"`; AsOf string `json:"as_of"`
}
type EconomicsNumbers struct { /* keep pkg/metrics/cost.go:134-143 semantics, binary-Go convention stated */ }
type Economics struct { /* scenarios + primary + retrans ratio + disclaimer; PRIMARY IS CALLER-DECLARED */ }

func CapacityWasteARPerHour(monthlyOpEx, wasted, transported float64) (float64, bool)  // bool = measured
func MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) (float64, bool)
func AmortizedARPerSecond(monthlyOpEx float64) float64                                  // 720 h convention, documented
func BuildEconomics(primary TariffOrPlan, linkType string, ratio RetransRatio, transported, wasted, loadSeconds float64) Economics

type DeadlineTarget struct{ DeadlineMs, TargetPct float64 }
func AssessDeadline(completions []float64, t DeadlineTarget) DeadlineResult             // strict-< documented
var CAPOperationalTarget = DeadlineTarget{220, 95} // "DGM operational target — not a WMO mandate (R1 §23)"

func DefaultMGATariffs(asOf string) (plans []Plan, tariffs []Tariff, err error) // dated Madagascar preset; Airtel/Orange Be Connect excluded as unverified until R5/R7 fetch succeeds
```

Design notes: (a) tariffs are data, never globals — no `Tiers`/`DefaultTier`/`tierMu`; (b) one metered path with explicit window (delete ×20 legacy); (c) machine codes out, presentation in the caller; (d) every refused/unknown input errors — no silent tier fallback (`pkg/api/server.go` 400/409 semantics preserved as errors); (e) JSON catalogue endpoint mirrors `GET /api/cost/enterprise` shape (`pkg/api/server.go:445-451`) so cgo can consume it unchanged; (f) CLI `cgo-cost --wasted-bytes N --transported M --window S --tariff yas-net-month-25000` prints the same `Economics` JSON, disclaimer included.

## 7. Slices

No slice touches code in this pass; each names its acceptance check.

1. **Slice 1 — Freeze + provenance types.** Create `cgo-cost` repo (own `go.mod`, R8), port the behaviour tests from `pkg/metrics/cost_test.go` + `economics_numbers_test.go` unchanged as `0.1.0` baseline; replace bare `Tariff`/`Plan` with §6 provenance structs; keep every existing constant behind `DefaultMGATariffs("2026-11-07")` with per-row source/date from §3. Acceptance: green tests; grep proves no exported mutable var; README declares the §6 API (R9 §1).
2. **Slice 2 — De-globalise + delete ×20.** Remove `Tiers`/`DefaultTier`/`SetDefaultTier`/`CostARPerH(×20)`; in-repo `pkg/campagne`/`pkg/audit`/`pkg/api` compile via a thin adapter holding the old global until migrated. Acceptance: race-detector clean under concurrent tariff reads; `CostARPerHTier` gone (`grep`).
3. **Slice 3 — Drift repair + verification harness.** Correct `yas-month-100gb`→250 Go (800 Ar/Go) per R2, Fibre FUPs per R4, Airtel tier per doc grid 05/10/2026 labelled provisional pending R5; add `scripts/check-tariffs.sh` that re-fetches R2–R4/R6/R7 and fails on drift (404/JS-gated ⇒ outputs "**unverified — primary source not found**" for that row, never keeps it green). Acceptance: harness passes live; stale values from §3 table gone.
4. **Slice 4 — Measured discipline.** `(value, measured)` returns on waste functions; primary becomes caller-declared; French sentences out (codes + provenance in, presentation in CLI/README example); quoted tariffs require explicit opt-in flag. Acceptance: boundary tests (zero transport/window, empty completions, exactly-95 %) assert non-measured, never zero-as-measured.
5. **Slice 5 — Tag + consume.** Tag `v0.1.0` (`vX.Y.Z`, immutable, R8), migrate `pkg/campagne`/`pkg/audit`/`pkg/api/server.go:417-475` to the module (or run dual-path with the Slice-2 adapter until migration lands); `GET /api/cost/tiers` + `/api/cost/enterprise` then read from `DefaultMGATariffs` output, preserving `docs/api.md` wire shape. Acceptance: `go list -m all` shows `cgo-cost`; in-repo `go test ./...` + `make verify` green; no `replace` directives left.

## 8. Anti-toy criteria

A cost calculator is a toy if it re-describes its own constants instead of computing against an external check. It is **not** a toy only if all hold:

1. **Independent arithmetic per family.** Wholesale 81,38 vs hand division 2 500 000/30 720 (R3 inputs); emergency 5 556 vs 25 000/4,5 (R2); GOLD hourly vs 0,1 × 1 000 000/720; ×20-equivalence at 180 s window and 90× at 40 s (existing `pkg/metrics/cost_test.go:35-51` vectors, re-based on confirmed R2/R3 inputs, not self-reference).
2. **Provenance on every number.** Tariff + source URL + relevé date + quoted flag, re-fetchable per §7 Slice 3. Any Airtel/Orange Be Connect/APN figure failing that bar is printed as "**unverified — primary source not found**" — the exact phrase required — not omitted silently and not defaulted into a primary.
3. **No silent zero.** Absent ratio/sample/window yields `measured:false`, never a 0 Ar/h that dashboards will read as "no waste" (§5.9; mirrors `BuildEconomics` refusal `pkg/metrics/cost.go:201-203`).
4. **No global mutable pricing** (§5.2). Two importers must be unable to reprice each other; the API surface is values and functions only.
5. **A real consumer.** At least one out-of-tree importer of `cgo-cost`, plus cgo itself serving `/api/cost/enterprise` from the module (Slice 5) with the wire shape of `docs/api.md:25` intact — otherwise the module is a restated constant table.
6. **Version discipline.** Declared API (R9 §1), `vX.Y.Z` tags immutable (R8), changelog in MAJOR/MINOR/PATCH terms. A repo that edits a released tag — e.g. quietly fixing the 100 Go tier in place — is a toy regardless of code quality.

---
*Research file only: no `pkg/metrics` source, `go.mod`, or API handler was modified. Slices await human approval per `AGENTS.md` ("Stop and ask" — new dependency, scope growth).*
