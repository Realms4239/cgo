# Meteolink M8 — Edge Shaping + NOC HD Observatory

**Date:** 2026-08-25
**Status:** Approved (Sections 1–7)
**Base:** `observatory2@68291cd` (rail icons, bento scoped, scrubber hidden, watermark 28px, idle rings gate, timeline ring gate, import inline, integrite peek) + `main@f09eb36` observatory 2.0 spec/plan
**Source:** ARG.md strict verdict 70% real (Traduction Matérielle, Kit Portable, Pilote, Why Go), second strict verdict (lab vs edge, client-side observation, baseline vs algorithm A/B, real audit CSV no templates), user M8 grouping (finish pending polish → decomposed layers), DESIGN.md §3.3 sharp NOC, snippets 30+ (animated-beam, donut-chart, snap-effect, style.css), screenshots 1920 live/campagne/resultats
**Approach:** Decomposed Layers — Layer 0 pending polish → Layer 1 backend (ARG 100%) → Layer 2 shell (METEOLINK full lockup, rail/foot/arm) → Layer 3 NOC (prompt progress line + HD sparkline every MetricCard + anime svg/text/animatable/layout via context7 + snippets + neat tidy+breathing placement)

## 1. Architecture

**Goal:** Meteolink is a Kit de Diagnostic Portable DSI for Edge Shaping in a zero-trust environment (not a network controller). Lab testbed proves CAKE/BBR, client-side audit observes the real Yas/Telma 4G access link, Traduction Matérielle makes it actionable on MikroTik/ISP, and the NOC visualizes baseline vs algorithm before/after as the jury's proof. Operation look wins.

**Stack:** Go 1.25 single binary `cgo` `embed dist` + React 19 Vite 6 Zustand low-freq + ECharts 5 tree-shaken `useDirtyRect` + D3 `scale/shape/axis/brush` + `animejs@4.5.0` latest `createTimeline/stagger/utils svg/text/animatable` verified `context7 1983 snippets` + woff2 `Cormorant/Inter/JetBrains Mono` + `animated-beam/donut/snap/style.css` curated (one SVG beam, one donut gauge, no new dep) + woff2 self-hosted.

**Structure:**
- `pkg/model` 17+14 cols `AQMEvalHeader` `LinkAuditHeader`, `campagne` matrix 36 `P1/P2×pfifo_fast/fq_codel/cake×cubic/bbr×3` `TCRunner handle1: tbf+fq_codel` stacked, `probe` `Ping 5Hz Small 4–32K BulkReceive + TCP_CONGESTION 0x0d`, `metrics` `Percentile/CostAR JFI/QDI`, `results Scan` `profile|qdisc|cc median/IQR + Best + hardware_recommendation` (new), `api` `Hub` delta + `/api/hardware/translate`, `qdisc` `SumDrops/SumBytes`
- `web/frontend` `App shell 48 header + rail 56↔232 · main bento` `store/ui railPinned/density/panel/live`, `lib/live rings 600 lttb400 hooks`, `components/Rail WebGLMesh Prompt Palette Nudge FlashBanner PeekPopover Timeline Beam DonutJFI PromptProgressLine MeteolinkWordmark` + `views Campagne Live Resultats Integrite/Recommandations`
- **Grouping:** Layer 0 is Task 0 (pending polish committed as foundation), Layers 1–3 are independent specs that share the same `go vet/build/vitest/playwright` gates.

**Why this split:** Lab vs edge isolates fiction; overlay Live wall + frozen Résultats A/B table isolate live 10 Hz vs stored median; beam/donut reuse `d3 path` + CSS; one mesh canvas + one beam SVG keeps bundle `315→~325 KB gz`.

## 2. Data & Metrics

**Primary Tableau 4 + secondary Tableau 5 — but every metric has baseline vs optimized delta:**

| Metric | Baseline unmanaged `pfifo_fast/cubic` | Optimized `fq_codel/bbr` or `cake` | Viz (Both with brush) |
|---|---|---|---|
| `rtt_p95` | `415 ms` P1 | `76 ms` `→ -82%` | Overlay grey dashed vs cyan solid + amber QDI fill 60px |
| `small_p95` | `738 ms` | `20 ms` | Bar grey→green `★ best` |
| `cost Ar/h` | `167968 Ar/h` | `0 Ar/h` `→ —` when idle | Mono `f4b400→767b84` honest |
| `JFI` | `—` idle | `0.98` | Donut arc steel→green 56px |
| `wasted/cost drops` | `0` honest when idle | `null → —` | MetricCard large 20px mono + sparkline + trend |

**Storage & truth:**
- Two origins: **Lab VM** `tc netem 100M/20M bulk GRIB` reproducible vs **Edge audit** `laptop → Yas/Telma 4G → 1.1.1.1 ping 5Hz + Small + bulk wget GRIB` via `POST /api/audit/start {site,link_type,duration 10–600 30}` → `link_audit.csv 14 cols` + `aqm_eval.csv 17 cols` `manifest.json sha256 + quarantine.json` for both; `data/frozen-wave3` holds real CSV (no template) — protects `ARG.md` fiction; wording `depuis un poste client situé dans les locaux` vs `Les mesures effectuées sur les sites révèlent`
- `results.Scan` groups `profile|qdisc|cc → median/IQR + Best` + `hardware_recommendation: string` (`Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+; Si ISP: mini-PC gateway CAKE transparent bridge` ) computed from `Best qdisc` (no new probe)
- **What we leave:** `TTB backlog/s`, `per-flow p99`, `BBR states TCP_INFO per flow`, `e1_onsets 356` — per-flow store churn for secondary story.

**Gates G0–G7 tiny mono `PASS/FAIL/—` in rail, banners carry words `CHARGE/OFFLINE/BASELINE`.**

## 3. Visual System

**Tokens:** `DESIGN.md` sharp `0 radius #070707 #101012 #26262a hairlines` + observatory 2.0 `radial cyan/violet + noise 1% + WebGLMesh 1%` + M8 `METEOLINK full lockup`: `Cormorant 600 Meteolink + LIEN subtitle + 16×16 NOC icon satellite→wave stroke1.5 round + 4px tricolor stripe + favicon` `+0.04→0.08em hover gradient text-clip #f2f2f4→#a9aeb6 drop-shadow cyan + shimmer 3s` + telemetry `oklch` `color-mix` + `Rail --rail-w 56→232`, `--gap 24 airy/16 dense`, `--dur 80/150 cubicBezier`

**Layout:** `header 48 glass sticky + rail 56↔232 animate width 400 + main bento` `view-transition-name:main` + `.card container-type:inline-size @container>600px pad24` + `[data-density]` `airy 24 dense 16` + **M8 bento scoped** `#v-campagne #v-live 2-col airy @1100px` (`Résultats/Intégrité` full width for wide tables) `panel-stack max 1280 width100 margin0auto` **Both tidy+breathing**: `panel-stack grid gap24 + subgrid kv rows` `outer 24 padding cards 16→24 @container` `view-title 20 Cormorant uppercase centered` `provenance 10 tabular right-aligned` `footer 28 ticker 10 tabular hash 8-char hairline top`

**Charts HD:** ECharts `canvas useDirtyRect min(dpr,2) at init`, `grid inset 64/32/48/40 rgba 0.005`, `axis cap round minorTick shadow pointer blur12`, `tooltip blur16 inset`, `line lttb shadow12 LinearGradient 26→0, blur`, `markArea per-series amber`, `dataZoom inside+slider 24 handleIcon patch`, `visualMap piecewise`, `graphic Meteolink·LIEN 28px 3%` + `animation false on Live` + `dataZoom:[]` when empty; D3 `Timeline 48 scaleTime area + circle/star isCurrent + brushX hidden idle via rings-empty` `QDI 60 clipPath rx4 brush` `Scatter brush rect` `Beam 1 SVG Campagne→Live when running` `Donut JFI 0–1 arc`

**Motion dense `animejs@4.5.0` 1983 snippets `svg/text/animatable/layout` verified 2026-08-19 trust 9.7:** `svg.createDrawable nav-icon path draw 800` on rail pin, `text.splitText chars stagger 30 from:center grid[4,2]` on `Meteolink` hover, `animatable` on `rail width`, `layout stagger grid [2,3] from:first` on `panel-stack` bento `stagger40`, plus every component `view y[8,0] blur stagger40`, `bars scaleX left`, `prompt y[12,0] + progress line width 100%→0% linear 6s pause on hover`, `toast y[16,0] stagger20`, `shake x[-4,4,0] elastic`, `banner scale`, `clipPath inset`, `Flash slide down 300` all `prefersReducedMotion` return

**Assets Best of both:** `noise 1% + meteo station + topography + waveform` + `gate pictos + bbr/cubic badges` + `luminous-button glow` on `best ★` + `snap-effect 16px handle` + `style.css` tokens + `animated-beam 27567b` + `donut-chart 92259b` (adapted as CSS/SVG, not new deps)

## 4. Components & Views

**4.1 Carried:** `Rail WebGLMesh Prompt Palette Nudge FlashBanner PeekPopover Timeline QDI/JFI MetricCard DataTable Provenance EmptyState StatusPip Toasts ErrorBoundary Skeleton PhaseStep lttb hooks anime`.

**4.2 Enhanced views — A/B + neat placement:**

| View | M8 — multi-layered | Connection |
|---|---|---|
| **Campagne** `1` | **Diagnose edge:** `P1/P2 reps × {pfifo, fq, cake} × {cubic,bbr}×3` + `Timeline 48 hidden idle via hasData rings-empty` (fixes empty phase '') + `G0–G7 StatusPip + side peek 32×12` + `Audit Site/Link/Durée InlineField validation + Lancer` + `Import INLINE Id/cap>0/delay>0 validate + Arm + Annuler` (no `prompt()`) + `bento 2-col airy` `Campagne|État` `Portes|Audit` `Profil full-width` + `État #0·rép 0` fix `event_id != null` | Hover `event row → Live spark 20pts Peek`, `Démarrer → auto Live` `Flash CHARGE` + `beam Campagne→Live` when running |
| **Temps réel** `2` | **Monitor A/B wall:** `8 MetricCards Every card HD sparkline 60×12 clipPath brush + trend` + `3 Charts overlay baseline grey dashed vs CAKE cyan/green solid + QDI amber fill 60px + small visualMap + markArea per-series + dataZoom hidden when empty + graphic 28px + animation false` + `Beam SVG Campagne→Live when running` + `bento 2-col` `MetricCards auto-fit 160 → QDI/JFI 1fr1fr → Charts 2-col` + `banner IDLE/OFFLINE/CHARGE/REPLAY` + `idle rings→EmptyState + — not red 0%` | Hover `Live point → Résultats row` `Peek`, `idle EmptyState` honest, `beam` between views |
| **Résultats** `3` | **Interpret A/B:** `5 groups table full-width (bento scoped)` `n small p95 bar cyan→green best★` `rtt p95 goodput` `JFI Donut 0–1 steel→green` `deadline_ok —/50/99/100%` `wasted/cost toggle` `quar best` + `Scatter goodput vs small brush rect` `xAxis 0–180` + `bento single col` `table full width` `scatter full` + `Recommandations card per group hardware_recommendation` + `why Go: zero-dependency vs Flent` | Hover `quar row → Audit site` `★ → topology`, `Rejouer → Live`, `BEFORE/AFTER diff badge -82%` |
| **Intégrité** `4` (+**Recommandations R7.1**) | **Impact + translate:** `runs/manifests/valid/quar kv` `RDF provenance hash 10px tabular` `EmptyState` `runs archivés ul hover peek groups medians 4 rows` `quarantine yellow banner` `Figures 2 SVG + Régénérer` `Provenance` `Replay` **new** `Recommandations — Traduction Matérielle` `table Linux principle → MikroTik PCQ/CAKE RouterOS v7+ Queue Tree / ISP mini-PC gateway + pilote isolé` card + `KIT portable 5-min branch without tc` copy + `Why Go` vs Flent + `pilote isolé département 4G` | Hover `run_id → groups medians Peek` |

**4.3 New M8 primitives:**
- `MeteolinkWordmark.tsx` `svg.createDrawable` `splitText chars stagger 30 from:center grid[4,2]` shimmer `letterSpacing 0.04→0.08`
- `PromptProgressLine.tsx` bottom `1px #f4b400 width 100%→0% linear 6s pause on hover` `animation-play-state:paused` + time `6s` mono `10px`
- `Beam.tsx` `animated-beam` 1 SVG `path linearGradient strokeDasharray 4 animate dashOffset -40 linear infinite` between `Campagne` and `Live` when `live.running` + `ResizeObserver`
- `DonutJFI.tsx` `donut-chart` `arc 0–360 56px stroke 8 steel→green` `value JFI` `—` when idle
- `Every MetricCard HD` `60×12 svg path lttb40 clipPath` already from M8 `Every MetricCard`
- `Rail 56 icons + sr-only + Foot 28 ticker 10px tabular hash 8-char + Arm spring + 5s timeout progress line` already from M8 Layer 2

**4.4 Safely fetch/leave:** `hardware_recommendation` computed from `Best` (no probe), `Beam/Donut/snap` reuse `d3 path` + CSS/SVG, `QDI/JFI` pure, leave `TTB/per-flow p99/BBR states/e1_onsets`.

## 5. Data Flow & Real-time

```
Lab VM reproducible                         Edge audit client-side (Yas/Telma 4G)
  tc netem 100M/20ms + bulk                     laptop → 1.1.1.1 ping 5Hz + Small 4–32K + bulk wget GRIB
     ↓ probe collect 300ms loop                ↓ metrics.Summarize median p95 + DeadlineOK + CostAR
     ↓ metrics.Summarize                       ↓ no tc write → writer.Append link_audit.csv 14 cols
     ↓ campagne Gates G0–G7                    → Freeze manifest.json sha256
     → writer Freeze lab manifest + quarantine  ↓ Sync to data/frozen-wave3 real CSV (no template)
     → results.Scan profile|qdisc|cc → median/IQR + Best + hardware_recommendation + why Go copy + pilote
     → figures.Generate bar+scatter+donut+RDF
                         ↓ both converge Live+Results A/B
Live 10 Hz: OnSnap Snapshot{phase,…,rtt_*,small,goodput,drops,gates,wasted,cost,deadline,running,event_id}
  → Live.Set → sse.Hub delta ts ring 2048 fanout → EventSource /api/stream 10 Hz → live.ts rings 600 lttb400
  → useRafLoop 4Hz → ECharts lazy + anime clipPath + HD sparkline every card + Beam when running + Prompt progress line
  → HD: lttb40 on every sparkline, clipPath rx4 brushX on QDI/Timeline/Scatter, Donut JFI, Beam Campagne→Live
Replay: /api/replay/list → /api/replay/stream?run → 400ms/row → replayRunning guard pause Live → banner REPLAY
Smart: POST /api/run/start smoke pilot P2×fq/bbr×2 reps now → Live wall overlay grey vs cyan + cost —; POST /api/audit/start → Client observes real RTT spike → link_audit.csv
A/B: Live wall overlay (baseline vs CAKE same scales) + Résultats bento A/B table 5 groups scroll before/after diff badge -82% + JFI donut
Translate: GET /api/hardware/translate?profile=P2 → Traduction Matérielle text MikroTik (from Scan hardware_recommendation)
Test: smoke pilot now visual, full 36 overnight for Part IV real CSV — SSE watched live via `curl -N /api/stream 10 frames/s` + `go test -tags=real` + playwright 7 passed
```

**Invariants:** browser never derives (Go `metrics` pure); rings mutable not Zustand (`frameCount%5 + gatesEqual` keeps React 2Hz, charts 4Hz, SSE 10Hz); `Last-Event-ID` resume; `Hub 16 default→dropped backpressure`; `useDirtyRect at init`; every `anime` `prefersReducedMotion`; `prompt progress line animation-play-state:paused` on hover; `idle` via `rings-empty` not `phase string`; `panel-stack tidy+breathing` scoped `Campagne/Live` bento, `Résultats/Intégrité` full.

## 6. Error Handling & Empty States

| State | Where | Operator sees | Handling |
|---|---|---|---|
| Loading | Résultats/Intégrité/Live 2s | `EmptyState loading + Skeleton 3-line shimmer + Provenance wait` | `fetch null → loading` |
| Empty | Résultats !groups, Intégrité runs 0, Live rings 0 | `EmptyState empty "disponible après gel (M2) / Démarrer pour alimenter Live" + Provenance n=0` + `Live idle rings→— not red 0%` | `available:false` never fakes; `idle rings-empty → null` |
| Offline | Live/Campagne connected false | `EmptyState offline "backend injoignable — vérifiez VM 192.168.174.128:9090" + banner OFFLINE red + Toasts err + Flash danger` | `EventSource.onerror → Flash` |
| Stale | Live 5s no frame | `EmptyState stale + banner STALE amber` | `useInterval 5s ts` |
| Error | POST run/audit/import 409/400 | `EmptyState error + ErrorPanel + Toasts err + Flash danger 2.5s + shake + Arm retry` | `!r.ok throw→catch setErr+pushToast` |
| Invalid/Degraded | Résultats quar + quarantine.json | `StatusPip err/warn + DataTable gate_status + Provenance n + Flash warn when quar>0` | `Scan gate_status!="valid"` |
| Validation | Audit Site/link/dur 10–600 + Import Id/cap>0/delay>0 + Campagne profiles | `InlineField gap2 shake + Arm disabled until valid + helper 'requis / min' + focus ring` | `validate.ts required/min/max string trim (fixes 0) + animateShake` All+optimistic |
| Prompt | Quick 6s bottom bar + progress line | `Prompt bottom 1px amber width 100%→0% linear 6s pause on hover + count label 6s + Esc` | `setTimeout 6s + setInterval 8s + animation-play-state` |
| Rail collapsed | 56px | `icon 16 + sr-only + miniSparkline 32×12` not `C…` truncate | `ICONS map + nav-icon` |
| Chunk | any lazy | `ErrorBoundary CHUNK_ERR → sessionStorage guard auto-reload` | already ported |
| Backpressure | Live Hub drop | `banner BACKPRESSURE amber + Toasts blue backpressure(N)` | `Hub 16 default→dropped` |
| Hardware translate | Recommandations empty | `EmptyState "Traduction disponible après Scan — lancez campagne"` | computed from `Best` |
| Download | Figures empty | `EmptyState loading during POST /api/figures/regen + Download ArmButton` | `regenOk` |

Copy `idle en attente / loading chargement… / empty aucune donnée / stale données périmées / offline backend injoignable / error erreur de source` fr.

## 7. Testing & Verification

| Layer | What we test (M8 adds) | How | Must pass |
|---|---|---|---|
| Go unit | qdisc handle1 stacked, metrics Percentile/CostAR JFI/QDI, campagne HappyPath/Writer G5, probe, results Best + hardware_recommendation, translate, Why Go copy | `go vet && go test -timeout 60s` | Every commit |
| Go VM | real tc netem/tbf/fq/cake + BBR 0x0d + testbed.sh up + qdisc stats drops/wasted delta | `make test-real` on `192.168.174.128` via `deploy/engine.sh ssh` | Before smoke pilot |
| Frontend unit | lttb, qdi/jfi, chartGrammar, tokens rail/gap, anime prefersReducedMotion svg/text, validation, Rail icons 56→232, prompt 6s bottom bar progress line, beam, donut, peek flip, timeline hide | `bunx vitest run jsdom` | Every commit |
| Type | strict noUnusedLocals `LiveFrame event_id maybeUndef` | `bun run typecheck` | Every commit |
| Build | `tsc+vite+check-bundle 450KB + echarts 250KB` `~325 KB gz` with `Meteolink` lockup + mesh + beam | `bun run build` | Every commit |
| E2E smoke | `shell Meteolink icons rail 56↔232 + Foot ticker + Arm progress + palette Cmd-K + prompt line drains 6s + validation shake + smart redirect Campagne→Live→Résultats` | `npx playwright test baseURL http://192.168.174.128:9090` host→VM | After deploy |
| Screenshots | `1920/1366/390 ×4 panels 12 shots + rail collapsed/expanded + prompt line + donut + beam + bento tidy` `animations disabled` | `SHOT_DIR=... npx playwright test screenshots` | Every deploy |
| Live HD | `GET /api/stream 10 frames/s ±20% + delta once + rings empty→— not 0% + scrubber hidden when empty + watermark 28px + every sparkline 60×12` | `curl -N /api/stream \| head -n 30` while pilot | During pilot |
| Smoke pilot | `P2×fq_codel/bbr ×2 reps` now visual: baseline vs CAKE overlay grey vs cyan + cost — | `POST /api/run/start {profiles:["P2"],reps:2}` + watch Live wall + `curl -N /api/stream` | Now |
| Full 36 | `P1/P2×3qdisc×2cc×3 reps` overnight for Part IV real CSV (no template) + `link_audit.csv` | `POST /api/run/start …` poll `/api/state` `running:false` | Overnight |
| Audit Réel | `laptop→Yas/Telma 4G → ./cgo audit --site DptX --duration 30 → link_audit.csv 14 cols real RTT spike` | manual + `GET /api/audit/list` | Before soutenance |
| Integrity | `manifest.json sha256 + quarantine.json + POST /api/figures/regen → svgs RDF` | `./bin/cgo verify` | After 36 |
| Hardware translate | `GET /api/hardware/translate?profile=P2 → Traduction Matérielle text MikroTik` | `curl /api/hardware/translate` | After Scan |
| a11y | `skip-link, nav-btn icon+aria, gate aria, dialog aria-modal + inert trap, toast aria-live, focus ring, prompt line` | `playwright + manual` | Every deploy |

Ledger:
```
host: go vet && go test && cd web/frontend && bun run typecheck && bun run build && bunx vitest run
VM:   bash deploy/engine.sh --action ensure && bash deploy/engine.sh --action deploy --config deploy/cgo-vm.yaml && ssh 'make test-real' && npx playwright test
live: curl http://192.168.174.128:9090/api/health && curl -N .../api/stream | head -n 30  # watch 10 Hz + rings-empty gate
pilot: curl -X POST http://192.168.174.128:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P2"],"reps":2}'
audit: ./cgo audit --site "Département X" --link_type 5g --duration 30  # generate real link_audit.csv
```

What we don't test: `TTB backlog/s`, `per-flow p99`, `BBR states TCP_INFO per flow` leave.
