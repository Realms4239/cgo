# Module TUI + wall — standalone display tool (research)

Date: 2026-11-07. Scope: `cmd/cgo/tui*.go`, `web/frontend/src/App.tsx`,
`web/frontend/src/components/VerdictHero.tsx`, the wall inside
`web/frontend/src/views/LiveView.tsx` (the rendered `#wall`), and
`docs/architecture.md`. Sibling: `docs/research/2026-11-07-module-tui-kit-vm.md`
covers the kit/VM TUIs; this file covers the operator TUI + web wall only, as
one display tool. No code, no commit.

Method: read source, inventory exported surface by `file:line`, check against
the primary sources below. Every claim cites `file:line` or `(R#)`; anything
else is labelled **unverified — primary source not found**.

## Primary sources (fetched 2026-11-07)

- [B1] Bubble Tea — charmbracelet/bubbletea (GitHub, `main` = v2) —
  https://github.com/charmbracelet/bubbletea : "A Go framework based on The
  Elm Architecture" ; "Bubble Tea programs are comprised of a **model** that
  describes the application state and three simple methods" — `Init`
  (initial `Cmd`), `Update` (events → updated model + `Cmd`), `View` (renders
  UI from model) ; `tea.NewProgram(initialModel())` then `p.Run()` ;
  "You can't really log to stdout with Bubble Tea because your TUI is busy
  occupying that" → log to file via `tea.LogToFile`. NOTE: [B1] `main`/v2
  returns `tea.View`, `tea.KeyPressMsg`, import `charm.land/bubbletea/v2` —
  that is **not** the pinned API; the pinned line is v1.3.10, see [B2].
- [B2] `pkg.go.dev` — `github.com/charmbracelet/bubbletea v1.3.10` (the
  pinned version, `go.mod:7`) — https://pkg.go.dev/github.com/charmbracelet/bubbletea@v1.3.10 :
  `type Cmd func() Msg` ("Cmd is an IO operation that returns a message when
  it's complete") ; `Tick(d, fn)` / `Every(d, fn)` — "Tick sends a single
  message and won't automatically dispatch messages at an interval. To do
  that, you'll want to return another Tick command after receiving your tick
  message" ; `Batch`, `Quit`, `WithAltScreen`, `WithFPS`, `WindowSize()`
  (auto-delivered `WindowSizeMsg` on start/resize) ; `type KeyMsg` /
  `KeyMsg.String()` ("It's safe (and encouraged) for use in key comparison") ;
  `func (m model) View() string` ; `LogToFile(path, prefix)`.
- [L1] Lip Gloss — charmbracelet/lipgloss (GitHub) —
  https://github.com/charmbracelet/lipgloss : "Style definitions for nice
  terminal layouts" ; `lipgloss.NewStyle().Bold().Foreground().Background().
  Padding().Border().Render(...)` (CSS-like, v1 import
  `github.com/charmbracelet/lipgloss`); automatic downsampling of color
  profiles (TrueColor → 256 → 16); `lipgloss.Wrap`, `lipgloss.Width/Height/Size`,
  `JoinHorizontal/JoinVertical`, `Place`, table/list/tree sub-packages.
- [G1] Go `go.mod` file reference — https://go.dev/doc/modules/gomod-ref :
  module path is the unique identifier + import prefix, v2+ must end `/v2` ;
  `go` directive is the mandatory minimum Go version since Go 1.21
  ("Go toolchains refuse to use modules declaring newer Go versions") ;
  `require`/`replace`/`exclude`/`retract` semantics; `replace` only applies in
  the main module and may be used transiently for local development.
- [G2] Go 1.4 release notes, "Internal packages" — https://go.dev/doc/go1.4 :
  "place it in a directory named `internal` … may not be imported by packages
  outside the source subtree".
- [R1] Grafana — Dashboards overview — https://grafana.com/docs/grafana/latest/dashboards/ :
  "A Grafana dashboard is a set of one or more panels, organized and arranged
  into one or more rows or tabs, that provide an at-a-glance view" ; panels
  query/transform raw data from a data source into visualizations.
- [R2] Grafana — Panel overview — https://grafana.com/docs/grafana/latest/panels-visualizations/panel-overview/ :
  "A Grafana panel is a visual representation of data composed of a query and
  a visualization" (plus transformations, legend, tooltips); panel menu
  offers View/Edit/Share/Explore/Inspect (Data/Query/Panel JSON).
- [R3] Grafana — Variables — https://grafana.com/docs/grafana/latest/visualizations/dashboards/variables/ :
  "A variable is a placeholder for a value that you can use in dashboard
  queries, panel titles, links" ; query/custom/text-box/constant/data-source/
  interval types; values sync to URL as `var-<varname>=value` unless
  `skipUrlSync`.
- [R4] Grafana — Dashboard JSON model — https://grafana.com/docs/grafana/latest/visualizations/dashboards/view-dashboard-json-model/ :
  "Grafana dashboards are represented as JSON objects that store metadata,
  panels, variables, and settings" ; three schema models (V2 Resource current,
  V1 Resource, Classic) ; Classic panel `gridPos` `{w: 1-24, h, x, y}`,
  `templating.list`, `refresh`, `schemaVersion`, `version`.
- [P1] Prometheus — Exposition formats — https://prometheus.io/docs/instrumenting/exposition_formats/ :
  "As of Prometheus version 2.0, all processes that expose metrics to
  Prometheus must use a **text** format, by default" ; text format is
  line-oriented, LF-separated, last line must end with LF, `Content-Type:
  text/plain; version=0.0.4` ; only one `HELP` and one `TYPE` per metric
  name, `TYPE` must appear before the first sample (else `untyped`);
  label-value escaping of `\`, `"`, LF required; unique metric+labels per line;
  histograms need `{le="+Inf"}` == count, summaries `x_sum`/`x_count`.
- [P2] Prometheus prior local pass — `docs/research/2026-11-07-module-prometheus-metrics.md`
  and `docs/research/2026-11-07-module-slo-runbook.md` record that
  `pkg/api/metrics.go` emits six static `"0"` samples and `rules.yml` names
  metrics (`cgo_cap_deadline_ok_pct`, `cgo_cap_sla_breach`, …) that are not
  emitted. Inherited as local evidence, not re-fetched here.

Local claims cite `path:line`. Where no primary source exists the claim is
marked `unverified — primary source not found`.

## 1. Responsibility

### 1.1 What the tool is

- **Display only.** The operator TUI (`cmd/cgo/tui.go:1-405` +
  `cmd/cgo/tui_audit.go:1-577`) and the web wall (`App.tsx` shell +
  `views/LiveView.tsx` `#wall` + `components/VerdictHero.tsx`) render state
  produced elsewhere: campaign frames via polling/SSE, sealed audit reports
  via `GET /api/audit/status`, frozen results via `GET /api/results`. Neither
  surface measures, shapes, or scores: `tui.go:395-397` ("Honest : sans API,
  Live affiche l'état déconnecté, jamais de données fabriquées"),
  `VerdictHero.tsx:1-4` ("Le verdict vient du moteur. L'interface ne classe
  rien"), `tui_audit.go:133-134` ("Jamais un chiffre recalculé : la cascade
  vit dans pkg/metrics et son résultat est gelé"). That is exactly [R1]'s
  dashboard shape (panels that query a data source and visualize it) and
  [P1]'s scrape-target shape (a process that exposes what it measured, never
  invents it).
- **Two renderers, one contract.** The terminal and the browser are two
  renderers over the same three contracts: (a) live frame (`frameTUI`
  `tui.go:37-52` ≈ `LiveFrame` `types.ts:1-43`), (b) sealed audit report
  (`auditView` `tui_audit.go:24-39` ≈ `AuditReport` `types.ts:236-256`),
  (c) frozen groups (`resultRow` `tui.go:54-66` ≈ `WallGroup`
  `LiveView.tsx:98`). Keeping the two renderers on one contract is the whole
  extraction problem (§5.1).
- **No `docs/architecture.md` Wall section exists.** `docs/architecture.md:1-103`
  has sections Composants / Flux SSE / Couche tc / Portes G0–G7 / Gel, manifest
  / Modes d'exécution / Pourquoi Go; `grep Wall docs/architecture.md` = 0 hits.
  The nearest wall-relevant text is the SSE hub description
  (`docs/architecture.md:23-41`: 10 Hz `time.Ticker`, delta compression,
  2048-frame ring + `Last-Event-ID`, 64-subscriber cap) and the manifest
  provenance paragraph (`docs/architecture.md:76-83`). Any "Wall section"
  requirement is therefore **unverified — primary source not found** beyond
  those two paragraphs; this report treats them as the architecture source.

### 1.2 `cmd/cgo` operator TUI (Bubble Tea, pinned v1.3.10)

- 5 tabs `Setup|Kit|Campagne|Live|Résultats` (`tui.go:33,296-302`); `main`
  entry `runTUI` (`tui.go:398-405`): `tea.NewProgram(initialModelTUI(),
  tea.WithAltScreen())` then `p.Run()` — the [B2] `NewProgram`+`Run`+`WithAltScreen`
  shape for v1.3.10 (not the [B1] v2 `tea.View`/`KeyPressMsg` shape).
  Dispatched as `cgo tui|top` (`cmd/cgo/main.go:53-54`).
- [B1]/[B2] Elm loop preserved: `modelTUI.Init` (`tui.go:176-178`,
  `tea.Batch(pollFrames, pollAudit)`), `Update` (`tui.go:180-258`,
  [B2] `KeyMsg.String()` comparison `msg.String()`), `View`
  (`tui.go:282-393`, returns `string` per [B2] v1). Polling is [B2] `Cmd`
  ("IO operation that returns a message"): `pollFrames` 1 s tick on
  `/api/state` (`tui.go:127-139`), `pollResults` on `/api/results`
  (`tui.go:141-159`), `pollAudit` on `/api/audit/status`
  (`tui_audit.go:105-129`), `httpPost` (`tui.go:260-278`). Every re-arm is
  explicit (tick returns another tick, [B2] "Beginners' note") — e.g.
  `resultsMsg` → 5 s `Tick` (`tui.go:249`), `auditMsg` → 10 s `Tick`
  (`tui.go:252`).
- Styling is [L1] `lipgloss.NewStyle().Padding.Foreground.Border().Render()`
  (`tui.go:20-31`: `stTab`, `stCell`/`stCellOn` with `RoundedBorder`).
  [L1] guarantees color downsampling, not the project's 80-column budget:
  the budget is `paneWidth = 78` (`tui_audit.go:98`) + `wrapWidth`
  hard-cut (`tui_audit.go:204-249`) + `TestRenderedViewsStayWithinEightyColumns`
  (`tui_audit_test.go:250-273` per sibling report).
- Audit mirror (`tui_audit.go:18-23,511-543`): verdict word+sentence verbatim
  (`verdictLine` `tui_audit.go:135-146`) + exactly four facts
  (`auditFactLines` `tui_audit.go:152-188`, zero never stands in for absent;
  `LossPct *float64` pointer so true 0 % renders, nil renders "non mesuré").
  Panes `deploymentLines`/`proofLines` (`tui_audit.go:343-413`), script
  writing `safeAuditID`+`scriptFor`+`writeRemediation` under `data/audits/<id>/`
  (`tui_audit.go:424-506`, never `/tmp`), OSC 52 "tentée" never "copiée"
  (`tui_audit.go:548-577`). All of this is display honesty, not measurement.
- Coupling: `tui.go` imports only the network, `pkg/metrics` (one type + one
  const, `tui_audit.go:15,33,173`), the filesystem, [B2]/[L1]. It never
  imports `internal/kit`/`internal/vm` except via subprocess `kitExec`
  (`tui.go:161-167`, `exec.Command(os.Args[0], "kit", action)`). It is
  already a standalone display tool in a `package main` costume (sibling
  report §5.2.1).

### 1.3 Web wall (`App.tsx` shell, `LiveView.tsx` `#wall`, `VerdictHero.tsx`)

- Shell: `App.tsx:35-178` owns mode routing (`MODES` dashboard/audit/bench/evidence,
  `store/ui.ts:7-12`), SSE connect (`connectSSE()` `App.tsx:51-55`,
  `lib/sse.ts:18-68`), campaign end/start transitions (`App.tsx:73-98`),
  keyboard nav (`App.tsx:106-122`), and the header/footer chrome. Only the
  landing view renders first; Bench/Audit/Evidence are `lazy()` + `Suspense`
  with mount-once caching (`App.tsx:23-25,44-49,147-164`). The wall lives
  inside BenchView → LiveView; `DashboardView.tsx:1-156` is the sealed-audit
  cockpit (VerdictHero at `DashboardView.tsx:113`).
- Wall (`LiveView.tsx:100-667`, root `<div id="wall">` at `LiveView.tsx:430`):
  ECharts surfaces via `useChart`/`ChartSurface` (`LiveView.tsx:47-96`,
  canvas renderer, primed idle option before first data `LiveView.tsx:80-84`);
  4 charts (RTT p50/p95, small p95 hero, goodput, iface RX/TX
  `LiveView.tsx:109-112`); 10-cell metric bento (`LiveView.tsx:477-498`);
  A/B overlay pfifo vs CAKE with `hash8` seal (`LiveView.tsx:532-641`);
  edge control (Surveillance / Façonnage / Essai `LiveView.tsx:548-618`);
  journal (`LiveView.tsx:643-663`). Data path is SSE `GET /api/stream`
  (10 Hz hub per `docs/architecture.md:23-41`) → delta merge + truth gate
  (`sse.ts:32-49`: structural fields retained, frames pushed only when
  `running|surveil|burst` + `measured`) → mutable rings outside Zustand
  (`lib/live.ts:8-39`, `max: 1800` = 180 s at 10 Hz; iface ring 600 = 60 s)
  → rAF paint loop with dirty-check (`LiveView.tsx:185-225`, iface ≥100 ms
  gate, stats ≥250 ms gate, LTTB downsample >400 pts `LiveView.tsx:204`).
  That loop shape (query → transform → visualize per [R1]/[R2]) is the web
  analogue of the TUI's [B2] Tick loop; the two must not be conflated.
- `VerdictHero.tsx:1-71`: renders `operator_verdict.{word,sentence,level}`
  verbatim (`VerdictHero.tsx:40-52`), level vocabulary frozen to
  `neutral|amber|red|green` with fallback `neutral` (`VerdictHero.tsx:24-26`),
  four facts from `operatorFacts` (`lib/dashboard.ts:83-124`: alerts %, delay
  ms, quality grade, wasted-capacity cost), absent → words not numbers
  (`'Pas encore mesuré'`/`'Non chiffré'`, `VerdictHero.tsx:19-21,30`).
  It never recomputes a verdict — the same rule as the TUI mirror (§1.2).
- Wall extras with no TUI equivalent: baseline lock (`LiveView.tsx:306-313`,
  `localStorage 'wall-baseline'`), shape POST (`LiveView.tsx:314-327`,
  `POST /api/shape`, 409-greyed during campaign `LiveView.tsx:587-602`),
  burst POST (`LiveView.tsx:328-336`, `POST /api/burst`), watch toggle
  (`LiveView.tsx:337-348`, `POST /api/watch`), export constat markdown
  (`LiveView.tsx:353-370`). The TUI has none of these (it starts/stops a
  campaign via `POST /api/run/start|stop` `tui.go:230-239` instead).

## 2. Exported surface (file:line)

### 2.1 Rule: nothing is exported

`tui.go:1`, `tui_audit.go:1`, `kittui.go:1` are `package main`. The web app
exports React components (`export default function App()`, `VerdictHero`,
`LiveView`) inside a `private: true` npm package (`package.json:2-4`,
`name: cgo-console`) with no published component API. Per [G2]-adjacent
reasoning (and [G1]'s module-identity rule), **no symbol is importable
outside the binary / the bundle**. The real exported surface is therefore
process + HTTP + filesystem + JSON-model contract:

| Surface | Evidence |
|---|---|
| `cgo tui` / `cgo top` → `runTUI` (5 tabs, `WithAltScreen`) | `cmd/cgo/main.go:53-54`, `tui.go:398-405`, `tui.go:33` |
| `CGO_API` default `https://meteolink.dev:9090` (HTTPS via `localCertPool`, `tui.go:99-119`) | `tui.go:89-95` |
| `GET /api/state` (1 s), `GET /api/results` (5 s), `GET /api/audit/status` (10 s) | `tui.go:127-159`, `tui_audit.go:105-129` |
| `POST /api/run/start {profiles:["P2"],reps:1}`, `POST /api/run/stop` | `tui.go:230-239,260-278` |
| Script files `data/audits/<audit_id>/remediation.routeros.rsc` / `remediation.linux.sh` | `tui_audit.go:424-506` |
| `GET /api/stream` (SSE, 10 Hz hub, delta, 64-sub cap, `retry: 2000`) | `docs/architecture.md:23-41`, `lib/sse.ts:18-68` |
| `GET /api/results`, `GET /api/integrity` (hash8), `GET /api/shape`, `GET /api/events`, `GET /api/audit/list` | `LiveView.tsx:159-172`, `DashboardView.tsx:63-68` |
| `POST /api/shape`, `POST /api/burst`, `POST /api/watch` | `LiveView.tsx:314-348` |
| `GET /metrics` (Prometheus text, static six zeros) | `pkg/api/server.go:397`, `pkg/api/metrics.go:43-64` |
| Sealed-audit JSON (`AuditReport`, `types.ts:236-256`) + live frame (`LiveFrame`, `types.ts:1-43`) | `lib/types.ts` |
| No Grafana dashboard JSON exists in repo (`deploy/` has zero `*.json`; `rules.yml` only) | glob `deploy/**/*.json` = 0 |
| Wall #wall layout: 4 charts + 10-cell bento + A/B overlay + 3 edge sections + journal | `LiveView.tsx:430,445-663` |

Unexported test seams (not API): `initialModelTUI` (`tui.go:89-95`),
`modelTUI.Init/Update/View` (`tui.go:176-393`), `auditFactLines/verdictLine/
deploymentLines/proofLines/scriptFor/writeRemediation/copyScript/osc52`
(`tui_audit.go:135-577`); web: `connectSSE/disconnectSSE` (`sse.ts:18-72`),
`pushFrame/pushIface/clearLive` (`live.ts:53-124`), `operatorFacts/
operatorSummary` (`dashboard.ts:58-124`).

## 3. Dependencies

- Go module `github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1-3`); direct
  requires (`go.mod:5-10`): `fyne.io/fyne/v2 v2.8.1`,
  `github.com/charmbracelet/bubbletea v1.3.10`,
  `github.com/charmbracelet/lipgloss v1.1.0`, `golang.org/x/sys v0.36.0`.
  Per [G1] that `go` floor is mandatory since Go 1.21 and the module path is
  the unique identifier — a split repo needs its own `go.mod` + floor.
- TUI imports: `tea` + `lipgloss` (`tui.go:14-15`, `tui_audit.go:13`) plus
  `github.com/Realms4239/cgo/pkg/metrics` (`tui_audit.go:15`). No Bubbles,
  no Harmonica, no ntcharts — [B1]'s "Libraries we use" list (Bubbles, Lip
  Gloss, Harmonica, BubbleZone, ntcharts) is a pointer, not a requirement.
  [B2] notes `WithFPS`, `WindowSize`, `LogToFile` exist in v1.3.10 but the
  TUI uses none of them (blind 1/5/10 s ticks, no `WindowSizeMsg` handling,
  `fmt.Fprintln(os.Stderr, ...)` only on exit `tui.go:401`).
- Web: `react ^19`, `react-dom ^19`, `echarts ^5.5`, `zustand ^5`,
  `animejs ^4` (`web/frontend/package.json:15-21`); dev: `vite ^6`,
  `vitest ^3`, `@playwright/test ^1.62`, `typescript ^5` (`package.json:22-33`).
  Charting is ECharts (not Grafana panels, not [L1] tables) — [R1]/[R2] are
  architectural analogues, not runtime dependencies.
- Server side (consumed, not owned): `pkg/api` routes `GET /api/health`,
  `GET /metrics`, `GET /api/state`, `GET /api/results`, `POST /api/run/start|stop`,
  `POST /api/shape`, `POST /api/burst`, `POST /api/watch`
  (`pkg/api/server.go:394-779`); SSE hub 10 Hz (`docs/architecture.md:23-41`).
  `GET /metrics` handler emits six hard-zero samples (`metrics.go:46-53`) —
  see §5.3.

## 4. Primary-source requirements

- [B2] Bubble Tea v1.3.10 contract: program = `Init`/`Update`/`View`
  ([B1] Elm Architecture, [B2] `View() string`); `Cmd func() Msg` is the
  only IO channel; `Tick`/`Every` send one message and must be re-armed by
  returning them from `Update` ([B2] "Beginners' note"); `WithAltScreen`
  owns the alternate screen; `KeyMsg.String()` is the sanctioned key compare;
  stdout belongs to the renderer ("can't really log to stdout" [B1] → file
  logging). A standalone `cgo-wall` TUI must preserve this loop exactly —
  direct HTTP in `View` violates the architecture. NOTE: [B1] `main` is now
  v2 (`View() tea.View`, `KeyPressMsg`, `charm.land/bubbletea/v2`); upgrading
  the pin is a breaking import change, not a flag flip.
- [L1] Lip Gloss is style-only ("Style definitions for nice terminal
  layouts"): `NewStyle().Border().Render()`, color downsampling, `Wrap/
  Width/Size`. It does not guarantee the 80-column budget, the four-fact
  shape, or truth-gating — those are project rules (`paneWidth = 78`
  `tui_audit.go:98`, `auditFactLines` `tui_audit.go:152-188`).
- [R1]/[R2] Grafana dashboards: a dashboard = panels (query + visualization +
  transformations) arranged in rows/tabs over a data source; a panel must be
  inspectable (Inspect → Data/Query/JSON) ([R2]). [R4] dashboards are JSON
  objects with `panels`, `templating.list`, `refresh`, `schemaVersion`,
  `gridPos {w:1-24}`. A `cgo-wall` claiming "Grafana-ready" must therefore
  ship versioned dashboard JSON whose panel queries hit `/api/state`/
  `/api/results`/`/api/audit/status` (or Prometheus series via [P1]) — a
  screenshot or a React-only wall is not a Grafana dashboard.
- [R3] Variables: one variable drives queries + titles + links and syncs to
  `var-<name>=` URL params. The wall's `profile/qdisc/cc` (`LiveView.tsx:98`,
  `frameTUI` `tui.go:37-52`) and `wall-baseline` localStorage state
  (`LiveView.tsx:125-127,311`) are the Grafana-variable equivalent; a
  `cgo-wall` Grafana export needs `templating.list` entries for them or
  duplication per dashboard.
- [P1] Prometheus exposition: since 2.0 all scraped processes must emit text
  format by default; `Content-Type: text/plain; version=0.0.4`; one `HELP` +
  one `TYPE` per metric, `TYPE` before first sample; escape `\`/`"`/LF in
  label values; unique metric+labels; histogram `{le="+Inf"}` == count.
  The current `pkg/api/metrics.go:17-41` builder deviates (prior pass
  `2026-11-07-module-results-figures.md:115`): no timestamps, no label
  escaping, always emits `{...}` even without labels, type inferred from
  `_total` suffix. A standalone wall/exporter must be written against [P1],
  not by copying that builder.
- [G1]/[G2] Module split: each independent repo needs its own `go.mod` +
  module path + `go` floor ([G1]); `replace` may point at a sibling during
  migration but must not survive a tag ([G1]); `internal/` may not be
  imported outside its subtree ([G2]) — `pkg/metrics` (imported by
  `tui_audit.go:15`) must be vendored, published, or duplicated as a local
  `Verdict` type in the new repo.
- [G1] Versioning: "The module path becomes the import prefix for all
  packages in the module" and v2+ needs `/v2`. Start `cgo-wall` at v0.x
  (no compatibility promise) with its own module path; the frontend's
  `version: 1.3.5` (`package.json:4`) and Go `KitVersion` are two copies of
  one number (sibling report §5.2.5) that must be single-sourced in the new
  repo.

## 5. Gap analysis

### 5.1 Tool vs library (structural)

| Fact | Evidence |
|---|---|
| TUI lives in `package main` → zero importable symbols | `tui.go:1`, `tui_audit.go:1` |
| Web app is `private: true` npm package → zero importable components | `package.json:2-4` |
| TUI hardcodes `CGO_API` default + 1/5/10 s ticks, no backoff, no `WithFPS` | `tui.go:89-95,127-159,249-252` |
| TUI execs `os.Args[0] kit <action>` (assumes binary contains kit); wall POSTs shape/burst/watch directly | `tui.go:161-167`, `LiveView.tsx:314-348` |
| Wall state lives in mutable module singletons (`live` rings, `liveRing.phaseSince`), not in Zustand store | `live.ts:8-39`, `store/ui.ts:26-36` |
| Wall depends on companion HTTP endpoints beyond TUI's four (shape/burst/watch/events/integrity/audit-list) | `LiveView.tsx:159-172,314-348`, `DashboardView.tsx:63-68` |
| `GET /metrics` is a static six-zero demo, not an exporter | `pkg/api/metrics.go:46-53` |
| No Grafana dashboard JSON ships (0 `*.json` under `deploy/`); `rules.yml` names unemitted metrics | glob = 0; P2 prior pass |

### 5.2 Including the web wall changes the split (vs sibling cgo-tui/cgo-kit)

1. **The wall is bigger than the TUI and less extractable.** The TUI is
   ~982 lines (`tui.go` 405 + `tui_audit.go` 577) over four endpoints. The
   wall slice (`App.tsx` 178 + `LiveView.tsx` 667 + `VerdictHero.tsx` 71 +
   `sse.ts` 73 + `live.ts` 128 + `dashboard.ts` 153 + `types.ts` 256 + stores)
   is ~1700+ lines over ten endpoints with charting, animation, and
   localStorage state. Extracting "the wall" without the bench/audit config
   screens leaves dead controls (shape/burst/watch POSTs have no home once
   BenchView is gone unless re-homed).
2. **The wall is not a Grafana dashboard.** It is a bespoke React + ECharts
   app with anime.js entrances (`App.tsx:69`, `LiveView.tsx:372-373`) and
   imperative chart instances (`LiveView.tsx:47-96`). Per [R1]/[R2] a Grafana
   dashboard is panel JSON over a data source; per [R4] it is versioned
   JSON with `panels`/`templating`/`schemaVersion`. None of that exists in
   this repo. A `cgo-wall` repo must author it from scratch (or explicitly
   de-scope Grafana and say so).
3. **Prometheus surface is decorative.** `GET /metrics` emits six static
   zeros (`metrics.go:46-53`, help text `metrics.go:54-61`); the rule inputs
   behind `rules.yml` (`cgo_cap_deadline_ok_pct`, `cgo_cap_sla_breach`) are
   not emitted (P2). A wall that claims "Prometheus-ready" on this endpoint
   fails [P1]'s "processes that expose metrics" contract (must serve real
   samples, `TYPE` before sample, escaped labels, unique metric+labels).
4. **Two honesty implementations must stay identical.** Verdict rendering is
   implemented twice: Go `verdictLine`/`auditFactLines` (`tui_audit.go:135-188`)
   and TS `VerdictHero` + `operatorFacts` (`VerdictHero.tsx:32-71`,
   `dashboard.ts:83-124`). Any drift (e.g. one side rounding, one side
   re-scoring) is a silent divergence; extraction must share a golden fixture
   across both repos.
5. **Scheduling honesty differs.** TUI polls blindly every 1/5/10 s with no
   `WindowSizeMsg`, no backoff, no `WithFPS` ([B2] all exist in v1.3.10).
   Wall paints only on dirty-check + ≥100/250 ms gates (`LiveView.tsx:185-225`)
   and pauses when inactive (`LiveView.tsx:114,225`). A unified `cgo-wall`
   must reconcile both without breaking either contract.
6. **`docs/architecture.md` has no Wall section** (`docs/architecture.md:1-103`,
   grep Wall = 0). Do not cite it as wall authority; cite SSE hub
   (`docs/architecture.md:23-41`) and provenance (`docs/architecture.md:76-83`)
   instead.

### 5.3 Live-blocker rows (`docs/VERIFY.md`) still blocking wall slices

- Interface chart paint ≥8/s REFUTED (`docs/VERIFY.md:39`: 7.1–8.0 paints/s
  headless/headed, SA-rAF 46–52 Hz) → the ifr TX/RX chart cannot claim 10 Hz
  painting; data advances 10.3/s, repaint is compositor-bound.
- Linux live, eBPF, IFB/BBRv3, `.dist`, CHR rows BLOCKED → the TUI mirror
  and the wall's Live tab have no live acceptance; fixtures only (VERIFY row
  20C: "Never observed against a live audit").
- No new wall claim may promote those rows; per `AGENTS.md` the builder may
  not certify its own number and the verifier re-runs the command.

## 6. Target tool — `cgo-wall` as independent display repo

### 6.1 Verdict

- **Extract one `cgo-wall` repo (binary + static bundle + dashboard JSON),
  not a Go/TS library.** [G2]/`package main` forbid importing the TUI;
  `private: true` forbids importing the web app. A published library would
  freeze side-effecting signatures (poll ticks, POST bodies, `TargetReachable`
  shape caps) as public API — the wrong commitment. The operator contract is
  process behaviour: exit code, `CGO_API` env, polling endpoints, script
  files, SSE delta/ring semantics, and dashboard JSON.
- **Keep `cgo-wall` display-only.** No measurement, shaping, or scoring code
  moves (the verdict cascade stays in `pkg/metrics` per
  `tui_audit.go:133-134`). `POST /api/shape|burst|watch` and
  `POST /api/run/start|stop` are forwarded user actions, not new logic.

### 6.2 Target shape

```
cgo-wall   (independent repo, own go.mod + go floor [G1], own package.json)
  cmd/cgo-wall          — main.go: `cgo-wall tui` (today cmd/cgo/tui*.go,
                          promoted out of package main per [G2]);
                          `cgo-wall serve` serves the static wall bundle
                          (today web/frontend build output) OR --api-url
                          points at a running cgo server
  internal/tui          — modelTUI (Init/Update/View per [B1]/[B2] v1.3.10);
                          internal/auditview (auditView, verdictLine,
                          auditFactLines, scriptFor, writeRemediation)
                          OR fold into internal/tui if single consumer
  web/                  — today web/frontend/src verbatim, minus Bench/Audit/
                          Evidence config screens (display only: DashboardView,
                          LiveView #wall, VerdictHero, sse/live/dashboard libs,
                          ui stores). Charts stay ECharts ([R1] analogue).
  dashboards/           — grafana/ JSON model per [R4] (Classic or V1/V2,
                          author + pin schemaVersion), panels per [R2]
                          (RTT, small p95, goodput, iface, A/B overlay,
                          provenance hash8), templating variables per [R3]
                          (profile/qdisc/cc/site) — NEW, does not exist today
  exporter/             — Prometheus text exposition per [P1] (HELP/TYPE
                          before sample, escaped labels/chars, LF-terminated),
                          replacing metrics.go static zeros — NEW
  deps: bubbletea v1.3.10 [B2], lipgloss v1.1.0 [L1], stdlib net/http,
        react 19 + echarts 5 + zustand 5 (pinned by package.json)
        NO internal/kit, NO internal/vm, NO pkg/campagne|qdisc|probe —
        talks to a running cgo via HTTP/SSE + writes ONLY
        data/audits/<id>/ remediation scripts
```

Why not "one library `cgo/display`" or "keep it in `cmd/cgo`": the first
freezes exit-code/POST semantics as a Go API; the second keeps the operator
monitor coupled to host-admin actions (kit/vm) that `docs/deploy.md:159-177`
deliberately separates, and keeps the Prometheus/Grafana surfaces decorative.

### 6.3 Module-path notes ([G1])

- Pick the path once (`github.com/<org>/cgo-wall`); [G1] warns to use the
  repository location as the import prefix and not to rename after publish.
  v2+ needs `/v2` — start v0.1.0 (no compatibility promise), do not claim
  1.0.0 until the CLI + HTTP + JSON-model contract is golden-tested.
- `replace github.com/Realms4239/cgo/pkg/metrics => ../cgo-metrics` may be
  used transiently ([G1] `replace` — main-module only) and must be dropped
  before tagging; prefer vendoring the single `Verdict` type (the TUI needs
  only `metrics.Verdict` + `CAPDeadlineMs`, `tui_audit.go:15,33,173`).

## 7. Proposed slices (one logical group each, explicit paths)

| Slice | Scope | Acceptance | Gate |
|---|---|---|---|
| S1 — Freeze contracts | Golden-test TUI keymap/tab order/four facts (`tui_audit.go:152-188`), wall `#wall` panel list (`LiveView.tsx:430-663`), `VerdictHero` absent-words (`VerdictHero.tsx:19-21`), SSE delta+truth gate (`sse.ts:32-49`, `docs/architecture.md:29-41`), `/metrics` static-zero shape (`metrics.go:46-53`). Declare in `docs/cgo-wall-api.md` per [G1] (path = ID) | `go test ./cmd/cgo -run TestContent` + `vitest run` green; `--help` + dashboard JSON absent-list committed as golden; new repo tests pass without sibling repos | Full `make verify` six gates; `VERIFY.md` row moves only by independent verifier run |
| S2 — `cgo-wall` TUI extraction (operator first) | Move `cmd/cgo/tui.go` + `tui_audit.go` → `cgo-wall/internal/tui` (promoted out of `package main` per [G2]); `runTUI` → `cmd/cgo-wall/main.go`; `CGO_API` + 1/5/10 s ticks preserved exactly ([B2] re-arm); keep 80-col suite (`tui_audit_test.go:250-273`); vendor `metrics.Verdict` | Binary renders 5 tabs + both panes from fixtures; `TestRenderedViews*`, `TestWriteRemediation*`, verdict/facts tests green in new repo; no `internal/kit/vm` import | `go test ./...` in new repo; live API row stays BLOCKED (fixtures only) |
| S3 — `cgo-wall` web extraction (display only) | Move `App.tsx` shell + `views/DashboardView.tsx` + `views/LiveView.tsx` + `components/VerdictHero.tsx` + `lib/sse|live|dashboard|types|format` + `store/ui|audit` → `cgo-wall/web/`; delete Bench/Audit/Evidence config screens; keep rings-outside-Zustand (`live.ts:8-39`), dirty-check paint (`LiveView.tsx:185-225`), absent-words contract | `vitest run` + `npm run build` + bundle budget (`check-bundle.mjs`) green in new repo; wall renders idle-empty + fixture live data; no fetch to deleted config endpoints remains | `npm run build` + e2e fixture run; paint-rate row stays REFUTED (no 10 Hz claim) |
| S4 — Grafana dashboard authoring | New `dashboards/grafana/` JSON models per [R4] (Classic or V1/V2) + `templating.list` variables per [R3] (profile/qdisc/cc/site) + panels per [R2] (RTT/small/goodput/iface/A-B/provenance) querying `/api/state|results|audit/status` | JSON validates against [R4] schema; one out-of-tree Grafana import renders; variable change re-queries all panels; Inspect → JSON shows provenance hash8 | JSON schema lint + import screenshot in checkpoint; no live campaign claimed |
| S5 — Prometheus exporter | New `exporter/` implementing [P1] text format (HELP/TYPE before sample, escaped `\`/`"`/LF, unique metric+labels, LF final line, `version=0.0.4`) emitting last frame + sealed-audit gauges; replace `metrics.go` static zeros (`metrics.go:46-53`) | Golden exposition tests pass against a [P1]-conformant parser; one scrape against a running `cgo-wall` returns real (non-zero) samples; absent stays absent (never emitted) | `go test ./exporter` + live scrape log; rules-input metrics (`cgo_cap_*`) remain `unverified — primary source not found` until measured |
| S6 — TUI hardening | Add [B2] `WindowSize()` handling + `WithFPS` cap + backoff replacing 1/5/10 s blind ticks; wire [B2] `LogToFile` ([B1] stdout rule); keep 80-col hard cut (`tui_audit.go:204-249`) | Fixture re-render at 40/80/120 cols never exceeds 80; disconnect stays "déconnecté" (never fabricated) | `go test ./...` in `cgo-wall`; no live claim |

Live runs need human approval per repo `AGENTS.md` ("The live bench without
explicit human approval" is never-touch) and `VERIFY.md` how-to-move rules
(rows move to VERIFIED only after the named command runs on the named host).

## 8. Anti-toy criteria (failing any one = toy)

1. **No fake live data.** Tabs 4–5 render only polled `/api/state` +
   `/api/results` + `/api/audit/status`, or the explicit "aucun audit /
   API injoignable" strings (`tui.go:371`, `tui_audit.go:110-121`);
   wall renders idle-empty (`LiveView.tsx:458`, `EmptyState`/`EmptyChart`)
   or SSE-measured frames only (`sse.ts:43-49`). A cell showing a number on
   transport error (`frameMsg{}` zero frame `tui.go:130-138`) without saying
   "non mesuré" is a toy ([B1] `View` renders the model, never the network).
2. **Four facts, never invented.** Go `auditFactLines`
   (`tui_audit.go:152-188`) and TS `operatorFacts` (`dashboard.ts:83-124`)
   return "non mesuré / Pas encore mesuré / Non chiffré", never `0.0` from
   a zero value. A renderer printing `0` for an absent latency is a toy.
3. **Verdict never recomputed.** `verdictLine` (`tui_audit.go:135-146`) and
   `VerdictHero` (`VerdictHero.tsx:40-52`) render the frozen
   `Verdict.Word/Sentence/Level` verbatim. Re-scoring in either surface is
   a toy.
4. **80 columns hard (TUI).** `wrap78`/`wrapWidth` (`tui_audit.go:204-249`)
   pin every line ≤80. A pane scrolling horizontally or splitting a `tc`
   command across two commands is a toy. ([L1] gives `Wrap`, not the budget.)
5. **Grafana JSON is real or absent.** Shipped dashboards are valid [R4]
   JSON (`panels`, `templating.list`, `schemaVersion`) with [R2]-inspectable
   panels and [R3] variables syncing to `var-*` URLs. A "Grafana" claim with
   only a screenshot, or panels with no data-source query, is a toy ([R1]
   "panels … query and transform raw data").
6. **Prometheus text is [P1]-exact.** `Content-Type: text/plain; version=0.0.4`,
   one `HELP`+`TYPE` per metric, `TYPE` before first sample, escaped
   `\`/`"`/LF, unique metric+labels, LF-terminated. Copying
   `pkg/api/metrics.go:17-41` (no escaping, no timestamps, `_total` type
   inference, static zeros `metrics.go:46-53`) is a toy.
7. **Bubble Tea contract intact.** `Init`/`Update`/`View` ([B1]),
   `Cmd func() Msg` + tick re-arm ([B2]), stdout owned by renderer ([B1] →
   file log). A `View` doing HTTP or a logger printing to stdout is a toy.
8. **SSE delta contract intact.** Structural fields retained by client
   (`sse.ts:32-42`), 2048-frame ring + `Last-Event-ID` replay, 5-frame fresh
   replay, backpressure never blocks measurement (`docs/architecture.md:29-41`).
   Dropping a field because the delta omitted it is a toy.
9. **Scripts never empty, paths never escape.** `scriptFor` honest errors,
   `safeAuditID` + prefix check, `data/audits` only
   (`tui_audit.go:424-506`). A `c` key writing an empty file or a file
   under `/tmp` is a toy.
10. **OSC 52 never claimed.** `copyScript` says "tentée" (`tui_audit.go:569-577`).
    A status asserting delivery is a toy (no terminal ACK exists).
11. **Paint-rate honesty.** Wall may claim "data advances 10/s" but not
    "paints 10 Hz": VERIFY row REFUTED at 7.1–8.0 paints/s
    (`docs/VERIFY.md:39`). A README claiming 10 Hz rendering is a toy
    (compositor-bound, not JS-bound).
12. **Live-bench claims stay BLOCKED until the command runs.** Per
    `VERIFY.md`, fixture-green never promotes Linux live, eBPF, IFB, BBRv3,
    `.dist`, CHR, or PC2/nemo rows. Builder ≠ verifier.
13. **One version identity.** `package.json:4` (`1.3.5`), `KitVersion`, git
    tag, and `cgo version` must agree; a release where they diverge is a toy
    release ([G1] one version per module — the three-copy drift is the gap,
    not the model).

## 9. Caveats

- `docs/architecture.md` has no Wall section (`docs/architecture.md:1-103`,
  grep Wall = 0); §1.1 cites its SSE + provenance paragraphs instead. Any
  "Wall section" quote beyond that is **unverified — primary source not found**.
- [B1] `main` is v2 (`tea.View`, `KeyPressMsg`, `charm.land/bubbletea/v2`):
  above-quoted [B1] tutorial details are v1-vs-v2 mismatched by design;
  [B2] v1.3.10 is the pin and wins on conflict.
- [P1] OpenMetrics vs text-format choice for S5 is open: text format
  (`version=0.0.4`) is chosen here because `metrics.go:45` already serves it;
  OpenMetrics (`# EOF` trailer) is the alternative target per [P1] — pick one
  and declare it. `unverified — primary source not found` for any claim that
  rule inputs `cgo_cap_deadline_ok_pct`/`cgo_cap_sla_breach` are currently
  emitted (no scrape executed in this pass; P2 records them as non-emitted).
- React/ECharts/zustand/anime.js versions and their APIs are local evidence
  only (`package.json:15-33`); per-component docs were not fetched, so any
  claim about React 19 concurrent rendering, ECharts option internals, or
  Zustand v5 store semantics beyond `store/ui.ts:26-56` is **unverified —
  primary source not found**.
- `docs/deploy.md`/`docs/VERIFY.md` rows are project-local claims, not
  external primary sources; they are cited by line as evidence, and BLOCKED /
  REFUTED rows remain so until the named host condition clears.

---
Skipped: kit/VM TUIs (sibling `2026-11-07-module-tui-kit-vm.md`), measurement/
shaping/audit engines, `cmd/cgo-gui` (fyne), TUI v2 migration, Grafana
provisioning YAML (fetched JSON model only). Risk: `GET /metrics` static zeros
and absent Grafana JSON make "Prometheus/Grafana-ready" the easiest claim to
fake — S4/S5 golden tests are the gate.
