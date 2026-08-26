# Meteolink M9 — Wall + Kit (GoAccess-inspired, Edge Shaping, TUI+Web)

**Date:** 2026-08-26
**Status:** Approved (Sections 0–7)
**Base:** `m8@2e7f131` (clean, no debug) + `observatory2@1e8edfd` (bento, scrubber, watermark, idle rings) + `main@3eab86f` M8 spec/plan
**Source:** M8 buggy diagnosis (prompt always popping, campaign terminée spam, 4 views toy, comparaison static, integrity not important), ARG.md 70% real (Traduction Matérielle, Kit Portable, Pilote, Why Go), second strict verdict (lab vs edge, client-side observation, baseline vs algorithm A/B, real audit CSV), GoAccess https://goaccess.io/ https://rt.goaccess.io/ (panel chooser, main chart date+live, TUI+Web), user grills (Small p95+QDI hero, drawer growth both, sheet hybrid, Both with brew, Both header/foot, both tidy+breathing, M8 decomposed layers)
**Approach:** Wall + Kit — one Wall (`/`), one Sheet (`Cmd-K`, `380px`), one Drawer (`56px` peek → full History), one Archives link, cross-platform `npm/brew/go` + `observe|control` SSH auth, HD sparkline every card, beam/donut/snap, anime svg/text/layout dense via context7, Task 0 bugs fixing first

## 0. Bugs Fixing First

Task 0 stabilises the toy before M9 layers: `QuickActionsPrompt` `Actions rapides` choices disrupting → replaced by **progress bar** `Event 3/6 — charge 4/10s` bottom `1px #f4b400 width 100%→0% linear 6s pause on hover` (not choices) + `remainRef` sync; `Timeline live.max` permanent shrink → `live.max=600` on null selection + cleanup; `animateGrid` double timeline jank → single `createTimeline .add grid[4,2]→[2,3]` + pause on panel change; `HardwareRecommendation` dedupe single `results/hardware.go` (was 3× drift `results.go` per row vs `translate.go` hardcoded `P1→cake`); `Scan quarantined` only `==invalid` (was `!=valid` → 40 vs 4); `Live QDI 0.0 → —` `idle||!liveSnap` + spark `idle→undefined`; `scrubber hidden when empty` `dataZoom:[]`; `watermark 56→28px METEOLINK·LIEN` (was clipped `daLink`); `peek run-aware` `fetch /api/results?run=`; `screenshots click fallback` for mobile rail; `shaper stacked on veth-c parent 1: tbf 80Mbit` (was veth-s egress unshaped → 4647 Mbps) + `StatsFn leaf-only` (was dual-hop double count → 231 Mbps) → goodput `77.7` valid G4; `sse Running` flap `phase!=""` → `true` + pump; `mobile rail 56→100% row 48` + `wordmark fallback #f2f2f4`; `drops spark` honest. All verified `go vet/test bun typecheck/build 323KB gz vitest 18×32 playwright 8/8`.

## 1. Architecture

**Goal:** 5-minute concrete job: `Auditer 30s` on Yas 4G at Dpt → `Comparer 2 configs` (pfifo vs CAKE) live on Wall → `Exporter` 1-page PDF with cost Ar/h. Jury sees Lab vs Edge vs Traduction vs Wall, not 4 toys.

**Stack:** `cgo` single binary `Go 1.25` `embed dist` + `React 19 Vite 6 Zustand low-freq + live.ts rings 600 lttb400` + `ECharts tree-shaken + D3 brush` + `animejs@4.5.0` `svg/text/animatable` via `context7` 1983 snippets + `woff2` + `npm wrapper + brew tap + go install` same binary `postinstall` downloads `cgo-linux/macos/win` + `TUI` `meteolink top` ASCII sparklines (GoAccess TUI) + one `WebGLMesh` canvas, one `Beam` SVG.

**Structure — Wall+Kit, not 4 equal views:**
- `/` **Wall** `header 48 METEOLINK lockup + pill Wall|History|Archives + Panel Chooser + density airy/dense + live meta run·hash·phase` `rail 56 icons ↔232 extended` `main wall` `footer 28 ticker` — wall is `8 HD MetricCards` (`small_p95` hero `20px` + `QDI amber 60×12` biggest) + `3 charts` `small_p95 hero 300px full-width` + `beam` + `donut JFI` + `bottom drawer` is Résultats. Cross-platform `npm/brew/go`.
- `/history` **History** full `Résultats A/B` `12 groups` `bar + scatter + diff badge`
- `/archives` **Archives** `Recommandations` `RDF` — footer link, not a view
- `Sheet` right `Cmd-K` `380px` `backdrop blur` `P1/P2 reps × qdisc × CC` + `Audit InlineField` + `Import INLINE` — slides over wall, `Démarrer` → `Arm 5s progress line` → sheet auto-minimizes to pill `Event 3/6` at top, wall takes over, drawer peeks.
- **Kit modes:** `observe` (laptop, no creds, audit only) vs `control` (mini-PC gateway, `CGO_MODE=control` `CGO_GATEWAY_SSH=ssh://altfloat@mini --key ~/.ssh/id_ed25519` `sudo tc` via `NsRunner` `cgo-srv` `veth-s`)

## 2. Data & Metrics

**Primary Tableau 4 + secondary Tableau 5 — every metric has baseline vs optimized delta, real only, both live+frozen with provenance:**

| Metric | Baseline unmanaged `pfifo_fast/cubic` | Optimized `fq_codel/bbr` or `cake` | Viz (Both with brush) |
|---|---|---|---|
| `rtt_p95` | `pfifo_fast median` `191 ms` | `cake best` `76 ms` `→ -60%` | Overlay grey dashed vs cyan solid + amber QDI fill |
| `small_p95` hero `300px` | `738 ms` | `20 ms` | Bar grey→green `★ best` |
| `cost Ar/h` | `167968` | `0` `→ —` when idle | Mono `f4b400→767b84` honest |
| `JFI` | `—` idle | `0.98` | Donut arc `steel→green 56px` |
| `wasted` | `0` honest when idle | `null → —` | MetricCard |

- **Real only, no `700` hardcoded:** Hero `small_p95` `738.4 vs 20.0` from `Scan median` (`P1 pfifo 738` vs `cake 20`), never `700` constant. If no `Scan`, hero shows `EmptyState “lancez campagne”`.
- **Both live + frozen with provenance:** Live overlay `grey dashed 73.9 vs cyan 46.9` from `live.rtt95` rings `600` `lttb40` when `running` + frozen `Résultats` `5→12 groups` from `Scan` when `!running`. Tooltip: `P1 pfifo rep2: 738.4 ms · hash a1b2c3d4 · 2026-08-25T20:13` (`manifest.json`).
- **Panel chooser (GoAccess):** Header pill toggles `MetricCards` visibility — `rtt`/`small`/`goodput`/`wasted`/`cost`/`QDI`/`JFI` on/off persisted `localStorage`, main chart `small_p95 300px` stays hero.
- **Gates honest:** `quarantined` only `gate_status=="invalid"` (not `degraded`), `hardware_recommendation` per profile best `MikroTik PCQ/CAKE vs mini-PC gateway`.

## 3. Visual System

**Tokens:** `sharp 0 radius #070707 #101012 hairlines #26262a` + `radial cyan/violet + noise 1% + WebGLMesh 1%` + `METEOLINK full lockup` `Carmarant gradient + 16×16 NOC icon satellite→wave stroke1.5 round + 4px tricolor stripe + fallback #f2f2f4` `+0.04→0.08em shimmer`.

**Layout:** `header 48 METEOLINK + Panel Chooser + density + live meta run·hash·phase` `rail 56↔232 width 400 + mobile row 48` `main wall bento` `view-title 20 centered` + `footer 28 ticker tabular` (now `Wall` bento `Campagne/Live 2-col`, `History/Archives` full). `Tidy+breathing`: `panel-stack gap24 airy/16 dense`, `card-head 11 caps`, `view-title 20 centered`, `provenance 10 right-aligned`.

**Charts HD:** `small_p95 hero 300px full-width` with `date Last updated 20:40` live pulse + `QDI 60 amber` below + `RTT/goodput 180px` secondary. `overlay grey dashed vs cyan` real only.

**TUI+Web:** `meteolink --serve` (Web wall) + `meteolink top` (TUI `8 cards ASCII sparklines 60×12` in terminal, same `live` rings).

**Motion dense `animejs@4.5.0` `svg/text/animatable/layout` (latest 4.5.0):** `svg.createDrawable nav-icon path draw 800` on rail pin, `text.splitText chars stagger 30 from:center grid[4,2]` on `METEOLINK` hover, `animatable` on `rail width`, `layout stagger grid [2,3] from:first` on `panel-stack` bento `stagger40`, plus every component `view y[8,0] blur stagger40`, `bars scaleX left`, `rail width 400`, `prompt y[12,0] + progress line width 100%→0% linear 6s pause on hover`, `toast y[16,0] stagger20`, `shake x[-4,4,0] elastic`, `banner scale`, `clipPath inset`, `Flash slide down 300` all `prefersReducedMotion` return

**Assets Best of both:** `noise 1% + meteo station + topography + waveform` + `gate pictos + bbr/cubic badges` + `luminous-button glow` on `best ★` + `snap-effect 16px handle` + `style.css` tokens + `animated-beam 27567b` + `donut-chart 92259b` (adapted as CSS/SVG, not new deps)

## 4. Components & Views

**Wall `/`** — `8 HD MetricCards Every card 60×12 clipPath rx4 lttb40 + trend` (`small_p95` hero `20px` + `QDI amber` biggest) + `3 charts` hero `small_p95 300px` + `beam` + `donut JFI` + `bottom drawer` is Résultats. Cross-platform `npm/brew/go`.
- `/history` **History** full `Résultats A/B` `12 groups` `bar + scatter + diff badge` `EXPORT MD/CSV`
- `/archives` **Archives** `Recommandations` `RDF` — footer link
- `Sheet` right `Cmd-K` `380px` `backdrop blur` `P1/P2 reps × qdisc × CC` + `Audit InlineField` + `Import INLINE` — slides over wall, `Démarrer` → `Arm 5s progress line` → sheet auto-minimizes to pill `Event 3/6` at top, wall takes over, drawer peeks.
- **Kit modes:** `observe` (laptop, no creds, audit only) vs `control` (mini-PC gateway, `CGO_MODE=control` `CGO_GATEWAY_SSH=ssh://altfloat@mini --key ~/.ssh/id_ed25519` `sudo tc` via `NsRunner` `cgo-srv` `veth-s`)

## 5. Data Flow & Real-time

```
Lab VM netem 20ms + tbf 80Mbit stacked parent 1: on veth-c (upload) 
  vs Edge audit laptop → Yas 10.200.0.1:8081/small (client-side, no sudo)
→ metrics Summarize → campagne Gates → writer Freeze manifest.json → Scan hardware_recommendation
→ Live 10 Hz OnSnap → Hub delta → sse.ts frameCount%5 →2Hz React + live.ts rings 600 lttb400 → useRafLoop 4Hz → ECharts + HD sparkline every card + beam
→ Wall hero small_p95 date Last updated + live pulse + TUI `meteolink top` same rings ASCII
→ Sheet Démarrer → pill Event 3/6 + drawer row append + scatter pop
```

## 6. Error Handling & Empty States

`Live` rings `0` → `—` not `0%` red (QDI `—`, deadline `—`, wasted `—`), `scrubber hidden` when empty, `timeline hidden idle` via `hasData rings-empty`, `prompt progress line` `6s` pause on hover + `Esc`, `rail mobile` `100% row 48`, `wordmark fallback #f2f2f4`, `quarantined` only `invalid`, `hardware empty` → `available:false`.

## 7. Testing & Verification

`go vet/test` `bun typecheck/build 323KB gz` `vitest 18×32` `playwright 8` + `screenshots 12` + `go test -tags=real` + `testbed.sh up` (netem+tbf stacked) + `smoke P1 1rep → 77 Mbps valid` `G4` + `full 36 P1/P2×3×2×3 reps →12 groups` + `audit --site DptX --duration 30` → `link_audit.csv 14 cols` real `rtt 20→375 small 52→754 0.90 Ar/h` + `GET /api/hardware/translate?profile=P2` `MikroTik` + `TUI` `meteolink top` `8 cards` + `date Last updated`.

