# MadaLink Perfect Frontend — Design

**Date:** 2026-08-24  
**Status:** Approved (Sections 1–7)  
**Source:** LIEN.md, docs/SPEC.md, docs/PLAN.md, DESIGN.md, donor thesis-cgo, user decisions 2026-08-24

## 1. Architecture

**Goal:** A 4-panel observatory (Campagne / Temps réel / Résultats / Intégrité) that is both a jury-grade defense artifact and a daily-use instrument for the Météo Madagascar operator, running as a self-contained Go binary with an embedded React SPA. Success is **B**: the operator can diagnose a link, run a campaign, and export a report without reading docs.

**Stack:** React 19 + Vite 6 + Zustand (low-freq store only) + ECharts 5 tree-shaken (`LineChart`/`Bar`/`Scatter` + `Grid`/`Tooltip`/`MarkArea` + `CanvasRenderer`) + D3 (`d3-scale`/`d3-shape`/`d3-axis` for campaign timeline + QDI sparkline) + anime.js 4.5 (`createTimeline` + `stagger` + `utils.set`) + JetBrains Mono / Inter var / Cormorant Garamond woff2 self-hosted (CDN fallback allowed since offline is flexible per Q2).

**Structure:**
- `web/frontend/src/App.tsx` — shell grid `48px header / 232px sidebar · main / 28px footer`, `PANELS` 4, `connectSSE` on mount, `1–4` keyboard nav, `ErrorBoundary` + `Toasts` at root.
- `web/frontend/src/store/ui.ts` — `panel`, `live: LiveFrame|null`, `connected`, `sseStatus`, `toasts`, `replayRunning` — *only* low-freq; high-freq stays in `lib/live.ts` mutable rings (`max 600` ≈60s @10 Hz).
- `pkg/*` — `model` (17-col `aqm_eval.csv` + 14-col `link_audit.csv` + `G0–G7`), `campagne` (matrix 36, gates, `writer.go` freeze), `qdisc` (`TCRunner` seam `handle 1:` stacked), `probe` (`BulkSendTo` per-cell `TCP_CONGESTION`), `metrics` (pure `Percentile`/`CostAR`), `results`/`figures`/`api` — backend may change (new metrics additive, per Q2).

**Why this split:** ECharts stays where 10 Hz performance is critical (Temps réel), D3 where bespoke shape is needed (timeline, QDI), anime where motion is meaningful (view fade, ArmButton armed, banner pulse). One binary, one `make build` (`tsc --noEmit && vite build && node scripts/check-bundle.mjs`).

## 2. Data & Metrics

**Primary metrics (LIEN Tableau 4 — stay primary, drive the story):**

| Metric | Source | Panel | Card look |
|---|---|---|---|
| `rtt_p95` / `rtt_p50` | `ping` 5 Hz `probe.Ping`, `metrics.Summarize` | Temps réel (RTT chart) + Résultats (rtt p95 column) | Large number + sparkline, cyan `#5ad3e3` |
| `small_p95` | `SmallObject` 4–32 KiB HTTP `probe.SmallObject` | Temps réel (small chart) + Résultats **leaderboard bar** (sorted by this, `best ★`) | Large number + bar length = value/max, green `#1fa348` when `best` with `shadowBlur:6` glow |
| `deadline_ok_pct` | `metrics.DeadlineOKPct(small, 1000)` | Résultats (deadline column, hidden by default, toggle) | Small badge `100%` |
| `bulk_goodput_mbps` | `tc -s` delta `SumBytes` / `chargeDur` (receiver-side) | Temps réel (goodput area) + Résultats (goodput column) + scatter X | Violet `#b48ae0` area, scatter X axis |
| `wasted_bytes` / `cost_ar_per_h` | `drops × 1448` (`qdisc.SumDrops` delta), `CostAR = wasted/4.5GiB×30000` | Résultats (cost column, toggle) + Intégrité (quarantine) | Mono 11px, amber when >0 |
| `drops` | `qdisc.SumDrops` delta | Temps réel (drops `kv` + `Gate` strip if shown) + Résultats (quar.) | Red `#e22718` when >0 |

**Secondary metrics (added from old CGO per Q5 B, as you approved):**

| Metric | Formula | Panel | Look |
|---|---|---|---|
| `QDI` | `rtt_p95 - rtt_p50` (or `loaded - baseline`) — queue delay increase | Temps réel **under RTT chart** as 60px `QDI` sparkline (`d3.line` + `area` `#f4b400` 12% opacity) — when `small_p95` spikes *and* `QDI` spikes, it's bufferbloat; when `small_p95` spikes but `QDI` flat, it's base RTT. | Small card, amber, collapsed by default, expands on "Diagnostic details" |
| `JFI` | Jain's fairness ` (Σx)² / (n·Σx²)` across `small_p95` per rep or per-flow goodput | Résultats **JFI column** (0.00–1.00) with subtle tint (green `>0.95`) | Mono 11px, shown always but de-emphasized |

**Storage:** `aqm_eval.csv` 17 cols `AQMEvalHeader` + `link_audit.csv` 14 cols `LinkAuditHeader` — column order enforced in `pkg/model/types.go`, frozen at `Freeze` with `manifest.json` SHA-256 + `quarantine.json` for `degraded`/`invalid` rows. `results.Scan` groups `profile|qdisc|cc` → `median`/`IQR` + `Best` per profile.

**What we intentionally do not bring back:** `TTB` (needs per-second `backlog` time series we don't store), `per-flow p99` (needs per-flow state we dropped), `BBR states` (`TCP_INFO` per flow) — each would add a probe+store+chart for a secondary story.

**Gates G0–G7:** Shown but **de-emphasized** per your call — tiny mono `G0–G7` `PASS/FAIL/—` in the sidebar status block and in Intégrité counts, no large colored strip, no banner. The operator sees `quarantined` in Résultats instead.

## 3. Visual System

**Product truth:** MeteoLink is an experimental observatory — **Operate** mode: scanability and truthful state outrank expression. One memorable element: the **Résultats leaderboard bar** (the `artificialanalysis.ai` moment). Everything else stays quiet.

**Tokens (from `DESIGN.md` observatory, now `MeteoLink`):**

| Token | Value | Use | Fidelity push |
|---|---|---|---|
| `--canvas` | `radial-gradient(ellipse at 50% -10%, rgba(90,211,227,0.04) 0%, transparent 50%), radial-gradient(ellipse at 100% 100%, rgba(180,138,224,0.03) 0%, transparent 40%), #070707` | page floor | Two subtle glows (cyan top, violet bottom-right) — depth without banding, `fixed` so it doesn't scroll. Adds 1% `noise.png` overlay `mix-blend-mode: overlay` `opacity: 0.015` for physical paper. |
| `--surface-card` | `linear-gradient(180deg, rgba(255,255,255,0.02) 0%, transparent 100%), #101012` + `backdrop-filter: blur(12px) saturate(1.15)` on hover | panels | Glass is now *layered* — gradient highlight + blur + `box-shadow: 0 0 0 1px rgba(255,255,255,0.04) inset, 0 8px 32px rgba(0,0,0,0.6), 0 1px 0 rgba(255,255,255,0.06) inset` — craft-floor "shadows carry offset and soft blur" at the limit. |
| `--hairline` | `linear-gradient(90deg, transparent, #26262a 20%, #26262a 80%, transparent)` 1px | separation | Hairline now fades at edges — less "box," more "instrument." |
| Telemetry | Same 6 hues, but now with **OKLCH**-derived stops for perceptual uniformity: `cyan oklch(0.85 0.12 200)` → `oklch(0.75 0.14 200)` for gradients. `color-mix` in CSS for hover states. | Data | Modern color science — `artificialanalysis.ai` uses OKLCH for its leaderboard tints. |

**Typography:** Self-hosted woff2 (with CDN fallback allowed): `Cormorant Garamond 600` uppercase tracked `+0.04em` → `+0.08em` on `MeteoLink` hover + `background: linear-gradient(90deg, #f2f2f4 0%, #a9aeb6 50%, #f2f2f4 100%)` `background-clip: text` + `filter: drop-shadow(0 0 12px rgba(90,211,227,0.5))` + `animate: shimmer 3s ease-in-out infinite` (`background-position` shift). On hover, `anime` `letterSpacing` `0.06→0.08em`.
- Chart titles: `Cormorant 600 13px` + `border-left: 3px solid var(--t-live)` + `padding-left: 12px` + `position: relative` + `::after` `content: ''` `position: absolute` `left: -3px` `top: 50%` `width: 3px` `height: 16px` `background: var(--t-live)` `box-shadow: 0 0 8px var(--t-live)` — the title itself glows.
- Data: `JetBrains Mono 11px tabular` + `font-feature-settings: 'tnum' 1, 'ss02' 1` + `font-variant-ligatures: none` — every `0` slashed, every `1` distinct.

**Layout:** CSS grid shell `48px header` `backdrop-filter: blur(16px) saturate(1.5)` sticky, `232px sidebar` `position: sticky` `top: 48px` `height: calc(100vh - 48px - 28px)` + `scrollbar-width: thin` `scrollbar-color: #26262a transparent`, `28px footer` `font-variant-numeric: tabular-nums`.
- `view` switch: `view-transition-name: main` + `::view-transition-old(root)` `animation: fadeOut 0.3s` + `::view-transition-new(root)` `animation: fadeIn 0.5s cubic-bezier(0.16,1,0.3,1)` — **View Transitions API** (Chrome 111+) for native-feeling page transitions, fallback to `fadeIn` keyframes.
- `card` `container-type: inline-size` + `@container (min-width: 600px) { .card { padding: 24px } }` — container queries, not just media queries.
- `panel-stack` `display: grid` `gap: 16px` + `subgrid` for `kv` rows — `grid-template-columns: subgrid` so labels align across cards.

**Chart grammar — ECharts pushed to greatest (every prop used):**

| Feature | Pushed value | Why |
|---|---|---|
| **Renderer** | `renderer: 'canvas'` `useDirtyRect: true` `devicePixelRatio: Math.min(window.devicePixelRatio, 2)` | `useDirtyRect` + capped DPR keeps 10 Hz at <3 ms on 2 vCPU. |
| **Grid** | `left: 64 right: 32 top: 48 bottom: 40 containLabel: true backgroundColor: 'rgba(255,255,255,0.005)' borderWidth: 1 borderColor: 'rgba(255,255,255,0.02)'` | Inset grid with hairline border — `artificialanalysis.ai`'s chart container. |
| **Axis** | `axisLine lineStyle: {width: 1.5, cap: 'round', color: '#2a2a30'}` `axisTick: {show: true, length: 4, lineStyle: {color: '#3a3a40'}}` `minorTick: {show: true, splitNumber: 4}` `splitLine: {lineStyle: {type: [4,4], color: '#1a1a1e', cap: 'round'}}` `axisLabel: {color: '#8b9099', fontSize: 10, fontFamily: 'JetBrains Mono', margin: 12}` `axisPointer: {type: 'shadow', shadowStyle: {color: 'rgba(90,211,227,0.04)', shadowBlur: 12}}` | Every `artificialanalysis.ai` axis detail — `cap: round` on lines, `minorTick` for precision, `shadow` pointer with blur. |
| **Tooltip** | `trigger: 'axis'` `backgroundColor: 'rgba(16,16,18,0.92)'` `backdropFilter: 'blur(16px)'` `borderWidth: 1` `borderColor: 'rgba(255,255,255,0.08)'` `padding: [16,20]` `extraCssText: 'backdrop-filter: blur(16px); box-shadow: 0 12px 40px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.06) inset; border-radius: 0;'` `textStyle: {fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', lineHeight: 18}` `formatter: (params) => \`<div style="min-width:180px">\${params.map(p=>\`<div style="display:flex; justify-content:space-between; gap:24px;"><span style="color:\${p.color}">● \${p.seriesName}</span><span style="font-variant-numeric:tabular-nums">\${p.value[1].toFixed(1)} ms</span></div>\`).join('')}<div style="margin-top:8px; padding-top:8px; border-top:1px solid #26262a; font-size:10px; color:#767b84;">QDI: \${(p95-p50).toFixed(1)} ms · JFI: \${jfi.toFixed(3)}</div></div>\`` | Rich tooltip with `QDI` + `JFI` + `★ best` + tabular numbers, `backdrop-filter` blur, `inset` border — `artificialanalysis.ai`'s tooltip is the benchmark. |
| **Series Line** | `type: 'line'` `smooth: 0.4` `smoothMonotone: 'x'` `sampling: 'lttb'` `lineStyle: {width: 2, cap: 'round', join: 'round', shadowBlur: 12, shadowColor: color+'66', shadowOffsetY: 2}` `areaStyle: {color: new echarts.graphic.LinearGradient(0,0,0,1, [{offset:0, color: color+'26'}, {offset:1, color: color+'00'}]), opacity: 0.8}` `emphasis: {focus: 'series', lineStyle: {width: 3}, itemStyle: {borderWidth: 2}}` `blur: {lineStyle: {opacity: 0.2}}` `markPoint: {data: [{type: 'max', label: {formatter: 'max {c} ms'}}]}` | `lttb` sampling built-in, `shadowBlur` 12 + `shadowOffsetY` 2 for lifted line, `LinearGradient` area, `blur` for focus, `markPoint` for max. |
| **MarkArea** | `markArea: {itemStyle: {color: 'rgba(244,180,0,0.04)', borderColor: 'rgba(244,180,0,0.12)', borderWidth: 1, borderType: 'dashed'}, label: {color: '#f4b400', fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop', padding: [4,8], backgroundColor: 'rgba(244,180,0,0.08)', borderRadius: 0}, data: [[{xAxis: chargeStart}, {xAxis: chargeEnd}]]}` | `charge` zone with `dashed` border + label, `backgroundColor` on label — `artificialanalysis.ai`'s zone highlight. |
| **DataZoom** | `type: 'inside'` `filterMode: 'none'` `zoomOnMouseWheel: true` `moveOnMouseMove: true` `preventDefaultMouseMove: true` + `type: 'slider'` `height: 24` `handleIcon: 'path://M-...'` `handleSize: '100%'` `handleStyle: {color: '#f2f2f4', borderColor: '#26262a'}` `backgroundColor: '#0b0b0c'` `fillerColor: 'rgba(90,211,227,0.12)'` `borderColor: 'transparent'` `selectedDataBackground: {lineStyle: {color: '#5ad3e3'}, areaStyle: {color: 'rgba(90,211,227,0.15)'}}` `emphasis.handleStyle: {borderColor: '#5ad3e3'}` | Full `inside` + `slider` with custom handle, `fillerColor` + `selectedDataBackground` — `artificialanalysis.ai`'s scrubber. |
| **VisualMap** | `show: false` `type: 'piecewise'` `dimension: 1` `pieces: [{gt: 100, color: '#e22718'}, {gt: 40, color: '#f4b400'}, {lte: 40, color: '#5ad3e3'}]` `outOfRange: {color: '#9aa3ad'}` | `small_p95` bar color encodes severity without extra legend — `piecewise` with `gt`/`lte`. |
| **Graphic** | `graphic: [{type: 'text', left: 'center', top: 12, style: {text: 'MeteoLink · LIEN', fill: 'rgba(255,255,255,0.025)', font: '600 56px Cormorant', textAlign: 'center'}, silent: true}, {type: 'image', left: 'center', top: 'center', style: {image: 'data:image/svg+xml;base64,...', width: 400, height: 400, opacity: 0.015}, silent: true}]` | Watermark `56px` + `noise.png` 1% overlay — `artificialanalysis.ai`'s subtle background. |
| **Animation** | `animationDuration: 900` `animationDurationUpdate: 600` `animationEasing: 'cubicInOut'` `animationDelay: (idx) => idx*15` `animationThreshold: 2000` | Staggered `lttb` reveal, `cubicInOut` not `linear` — `prefers-reduced-motion` → `0`. |

**D3 — enhanced (where ECharts can't):**
- **Campaign timeline:** `d3.scaleTime` `domain: [baselineStart, recupEnd]` `range: [0, width]` + `d3.axisBottom` `tickFormat` `d3.timeFormat('%H:%M')` + `d3.area` for `baseline` (faint cyan `0 0 12px`), `charge` (amber 12% + `pattern` diagonal `4px` `rgba(244,180,0,0.08)`), `récup` (green 8%) + `d3.line` for `event_id` ticks `r: 3` `fill: var(--t-live)` + `d3.symbol` `star` for `best` events.
- **QDI sparkline:** `d3.line` `curveMonotoneX` + `d3.area` `curveMonotoneX` `fill: url(#qdi-gradient)` (`linearGradient` `amber` `0% 26% → 100% 0%`) + `clipPath` `rx:4` + `d3.brushX` `extent: [[0,0],[width,60]]` `on('brush end')` updates `live.max` (time window) — the `artificialanalysis.ai` time scrubber.

**Motion — enhanced (still 3 moments, but richer):**
- View: `View Transitions API` + `anime.timeline` — `anime({targets: '.view', translateY: [8,0], opacity: [0,1], filter: ['blur(4px)','blur(0)'], duration: 500, easing: 'cubicBezier(0.16,1,0.3,1)'})` + `anime({targets: '.card', translateY: [12,0], opacity: [0,1], delay: anime.stagger(40), duration: 600})` staggered card entrance.
- `ArmButton` armed: `anime({targets: '.btn-danger', scale: [0.96,1], rotate: [0.5,-0.5,0], duration: 400, easing: 'easeOutElastic(1, .6)'})` + `box-shadow` pulse `0 0 0 → 0 0 16px`.
- Live banner: `anime({targets: '.banner', scale: [1,1.015,1], backgroundColor: ['#161618','rgba(90,211,227,0.08)'], duration: 700})` throttled 2 Hz, `prefers-reduced-motion` → `opacity` pulse only.

**Assets:** `public/textures/noise.png` `public/figures/meteo-station.jpg` (real DGM, to be replaced) `public/fonts/*` already, plus `public/icons/meteo-*.svg` custom `16×16` stroke `1.5` `round`/`round` for `rtt`/`small`/`goodput`/`drops`.

## 4. Components & Views

**4.1 — Design tokens already shipped (no change, just carried):** `--canvas` radial glow + `#070707`, `--surface-*`, `--hairline`, `--text-*`, `MeteoLink` tricolor identity-only, 6 telemetry hues, `Cormorant`/`Inter`/`JetBrains Mono` woff2, `48px header`/`232px sidebar`/`28px footer` grid, `card` one level, `ArmButton` double-confirm, `banner` state line.

**4.2 — Enhanced views (what changes for "10000× perfect usable"):**

| View | Current (4 panels, 780 LOC) | Perfected (MeteoLink, 4 panels, ~1.2k LOC) — what the operator actually does |
|---|---|---|
| **Campagne** `1` | Selectors `P1/P2` `reps` + `ArmButton` `DÉMARRER` + `kv` `phase/profile/qdisc/cc/event` + `G0–G7` `gate-row` + `Audit` form + `Import` | **Diagnosis control:** Keep selectors, but add **D3 timeline** (`scaleTime` `area` `baseline` faint cyan, `charge` amber 12% + diagonal pattern, `récup` green 8%, `circle` current phase pulse) at top — the operator sees "where in the 3-minute event are we?" without reading numbers. `G0–G7` `gate-row` gets `StatusPip` dot + `Provenance` line (`source: tc -s` `rafraîchi 10s`). `Audit` form gets `EmptyState` `offline` when `GET /api/audit/list` fails, and `Provenance` `n` count. |
| **Temps réel** `2` | 3 ECharts `Canvas` (`RTT` `small` `goodput`) + `banner` `IDLE/BASELINE/CHARGE` + `drops` `kv` | **Monitoring:** Keep 3 charts, but each `card` now has `chartGrammar` **pushed** (`markArea` charge zone, `dataZoom` `slider` 24px `handleIcon`, `visualMap` for `small_p95` severity, `graphic` watermark) + **D3 `QDI` sparkline 60px** under RTT (`rtt_p95 - rtt_p50`, amber `area` `clipPath rx:4`, `brushX` to scrub `live.max`). `banner` now state-colored + `REPLAY` cyan. `drops` `kv` becomes `StatusPip` `live`/`err` + `Provenance` `source: /api/stream · 10 Hz`. All charts `animationDuration: 0` (anime drives container `clipPath`). |
| **Résultats** `3` | Table `profil/qdisc/cc/n/small p95/rtt p95/goodput/quar/best ★` + bar `width pct` + CSV/MD export + `Provenance` | **Interpreting:** Keep table, but **bar** now `anime` `scaleX` + `shadowBlur` + `rx:2` `drop-shadow` (the `artificialanalysis.ai` bar). Add **Scatter** `goodput` vs `small_p95` (ECharts `Scatter` `symbolSize: 8` `emphasis` `shadowBlur: 12`, `best` green `★` larger `12`) with `brush` `toolbox` "compare 2 configs" + `DataTable` sortable (`aria-sort`) for `quarantine` drill-down. `EmptyState` `empty` when `!groups` + `Provenance` `source: data/runs/*/aqm_eval.csv` `n=groups.length`. |
| **Intégrité** `4` | `valid/quarantined` `kv` + `runs` `ul` + `Figures` 2 SVGs `regenOk` + `Replay` `stream` link | **Impact:** Keep `valid`/`quarantined` `kv` but as `StatusPip` `ok`/`err` + `Provenance` `source: manifest.json` `n` + `EmptyState` `offline` when `!available`. `Runs` `ul` becomes `DataTable` sortable `run_id`/`manifest`/`quarantine` + `Provenance` per run. `Figures` now `2` SVGs with `RDF` provenance `sha256` + `Download` `ArmButton` + `EmptyState` `loading` during `POST /api/figures/regen`. `Replay` now `REPLAY` banner + `DataTable` `run-smoke` + `ArmButton` `Rejouer` → `Live` (already done) + `Provenance` `source: aqm_eval.csv` `extra: 200ms` |

**4.3 — New shared primitives (from donor, now ported):**

- `lttb.ts` + `hooks.ts` `useRafLoop`/`useInterval` — already done, now used in `LiveView`.
- `EmptyState` (`idle`/`loading`/`empty`/`stale`/`offline`/`error` + `hint`) + `Provenance` (`source` `rafraîchi` `n` `extra` + `state` `live`/`err`/`wait`) — already copied, now wired in `Resultats`/`Intégrité`.
- `StatusPip` (`ok`/`warn`/`err`/`on`/`idle` via `PIP_MAP`) + `Badge`/`Chip` — for `G0–G7` and `quarantined` counts.
- `DataTable` (`sortable` `aria-sort` `render` hooks) — for `Résultats` quarantine drill-down and `Intégrité` runs.
- `ErrorBoundary` (chunk-load `Failed to fetch` auto-reload `sessionStorage` guard) + `Toasts` (`aria-live` `3.2s` auto-drop) — at `App` root.
- `Skeleton` (N-line shimmer) + `PhaseStep` (numbered `todo`/`blocked`/`running`/`done` + `Lamp` + `Provenance` footer) — for `Campagne`'s `G0–G7` as `PhaseStep` list.

**4.4 — What we safely fetch from old CGO (and what we leave):**

| Old CGO did | Can we safely fetch? | Verdict for MeteoLink |
|---|---|---|
| `e1_onsets.csv` detector bank (5 detectors, KS, EWMA) + `e3_events.csv` adaptive | No — `356` events, `KS`/`EWMA` stats lib, `oracle` at 10 Hz | **Leave** — LIEN's `QDI`/`JFI` cover the diagnostic need with 2 lines of math, no `KS` lib. |
| `tc -s` `PollStats` per-qdisc `drops`/`overlimits`/`backlog` + `PollLeafStats` isolation | **Yes, safely** — `qdisc/stats.go` already ported, `SumDrops`/`SumBytes` used for `goodput` delta + `wasted_bytes` | **Fetch** — already done, now `goodput` is `tc -s` delta, not sender `Write`. |
| `WastedBytes` via `TCP_INFO Total_retrans × Snd_mss` | **Yes, safely** — `flows/bbr_native.go` `WastedBytes()` pattern, but needs `TCP_INFO` per flow | **Fetch as simplified** — `wasted = drops × 1448` (MSS) per `campagne.go:213` is the `old CGO`'s fallback `drops×1200` idea, honest and 1 line. Full `TCP_INFO` `Total_retrans` can come later if `wasted_bytes` becomes a thesis claim. |
| `QDI`, `TTB`, `JFI`/`LFI` | **QDI yes, JFI yes, TTB no** — `QDI` is `rtt_p95 - rtt_p50`, `JFI` is `(Σx)²/(n·Σx²)` across reps (both pure, no new probe). `TTB` needs per-second `backlog` time series we don't store. | **Fetch `QDI` + `JFI` as secondary** (as you approved). Leave `TTB`/`per-flow p99`/`BBR states`. |
| `Section` + `Provenance` + `EmptyState` + `StatusPip` + `DataTable` + `Toasts` | **Yes, safely** — pure UI, no backend coupling | **Fetched** — 6 files, `components.css` already copied. |
| `lttb` + `useRafLoop` + `useLiveEchart` | **Yes** — already fetched | Done. |
| `check-bundle.mjs` budget guard | **Yes** — already `450 KB` guard, `echarts` watchdog | Done. |

## 5. Data Flow & Real-time

**How a measurement becomes a pixel:**

```
Probe (ping / SmallObject / BulkReceive)
  → collect(secs) loop per phase (300 ms: Ping×5 + Small×1)
  → metrics.Summarize (median, p95, IQR) + metrics.DeadlineOK + metrics.CostAR
  → campagne.RunEvent → gates G0–G7 → GateStatus valid/degraded/invalid
  → writer.Append (G5 dedup via seen map) → Freeze → manifest.json (SHA-256) + quarantine.json
  → results.Scan (group profile|qdisc|cc → median/IQR + Best per profile)
  → figures.Generate (bar + scatter SVG + RDF provenance)
  ↓
Live path (10 Hz, not stored):
  Deps.OnSnap(Snapshot{phase,profile,qdisc,cc,repetition,event_id,load_status,rtt_p50/95,small_p95,goodput,drops,gates,running})
  → Live.Set → sse.Hub.Publish (delta-encode profile/qdisc/cc/repetition/event_id/phase, add ts, JSON marshal, ring 2048, fanout id+data, backpressure named event)
  → EventSource /api/stream (onmessage JSON parse, delta-retain last, pushFrame if running, frameCount%5 → setLive 2 Hz, backpressure → sseStatus, onerror → connected false)
  → live.ts rings (max 600, 60s @10 Hz, lttb(400) when >400) → useRafLoop (4 Hz) → ECharts setOption lazyUpdate
  ↓
Replay path (archived):
  GET /api/replay/list → run_ids
  GET /api/replay/stream?run= → ReadAll aqm_eval.csv rows[1:], for each row: JSON marshal 11 cols + ts + running:true + phase:replay, id++, data:, flush, 400 ms sleep
  → EventSource /api/replay/stream → replay.ts onmessage → pushFrame + setLive (replay:true) → LiveView banner REPLAY — live SSE paused via replayRunning guard in sse.ts
```

**Key invariants (ponytail):**

- Browser never derives a metric — Go is the sole calculator (`pkg/metrics` pure, `campagne.go` gates).
- Rings are mutable, not in Zustand — React re-renders at 2 Hz (store throttle), charts repaint at 4 Hz (rAF), SSE ingests at 10 Hz — no 10 Hz React churn.
- `Last-Event-ID` replay + `gate_status` in CSV → `results.Scan` + `quarantine.json` at freeze — the run can be resumed via `seen` map (`matrix.go`).

**Backpressure & offline:**

- `Hub` `sub.ch` buffered 16, `select default` → `dropped++` + `event: backpressure` named event → `sseStatus = 'backpressure'` + `Toasts` `blue`.
- `EventSource.onerror` → `connected false` → `banner OFFLINE` + `EmptyState offline` in `Campagne`/`Live`.
- `Writer` `seen` prevents G5 dupes even if `POST /api/run/start` is called twice.

## 6. Error Handling & Empty States

**Every failure is a designed state, not a blank panel:**

| State | Where | What the operator sees | How we handle it |
|---|---|---|---|
| **Loading** | `Resultats` `Integrite` `Live` (first 2 s before SSE) | `EmptyState kind="loading" hint="agrégation des réplications"` + `Skeleton` 3-line shimmer (matches table row shape) | `useEffect` `fetch` → `setGroups(null)` → `loading`; `Skeleton` in `card` until `groups` arrives |
| **Empty** | `Resultats` when `available:false` (no frozen CSV) + `Integrite` when `runs:0` | `EmptyState kind="empty" hint="disponible après gel (jalon M2)"` + `Provenance source: data/runs/*/aqm_eval.csv state:wait` + `ArmButton` `DÉMARRER` in `Campagne` | `GET /api/results` returns `{"available":false,"reason":...}` — frontend never fabricates a row |
| **Offline** | `Live` `Campagne` when `connected:false` or `GET /api/state` fails | `EmptyState kind="offline" hint="backend injoignable — vérifiez VM 192.168.174.128:9090"` + `banner OFFLINE` red + `Toasts` `err` | `EventSource.onerror` → `setConnected(false)` + `setSseStatus('déconnecté')` + `ErrorPanel` with `Réessayer` → `connectSSE()` |
| **Stale** | `Live` when `Date.now() - live.ts > 5000` (no frame for 5 s) | `EmptyState kind="stale" hint="données périmées — dernier év. il y a 5 s"` + `banner` `STALE` amber + `Provenance` `state:live` → `state:err` | `useInterval` 5 s staleness check in `App` (already `live.ts` `ts` + `dirty` flags) |
| **Error** | `Campagne` `POST /api/run/start` 409, `POST /api/audit/start` 409, `POST /api/profile/import` 400 | `EmptyState kind="error" hint={error.message}` + `ErrorPanel` `msg` + `Toasts` `err` + `retry` `ArmButton` | `fetch` `!r.ok` → `throw` → `catch` → `setErr` + `pushToast(msg, 'err')` |
| **Invalid / Degraded** | `Resultats` `quarantined` column + `Integrite` `valid/quarantined` `kv` + `quarantine.json` table | `StatusPip` `err` for `invalid` `warn` for `degraded` + `DataTable` `gate_status` sortable + `Provenance` `n` | `results.Scan` counts `gate_status!="valid"` as `quarantined`; `writer.go` `quarantine.json` persists the reason |
| **Chunk load failure** | Any `lazy` view (12 views in donor, 4 in MeteoLink) over flaky tunnel | `ErrorBoundary` `CHUNK_ERR` `Failed to fetch dynamically imported module` → `sessionStorage` one-shot auto-reload guard + manual `Recharger` button | `components/ErrorBoundary.tsx` already ported |
| **Backpressure** | `Live` when `Hub` drops frames | `banner` `BACKPRESSURE` amber + `Toasts` `blue` `backpressure (N)` + `s` | `Hub` `sub.ch` buffered 16, `default` → `dropped++` + named event |

**Copy discipline (from donor, fixed encoding):** `EmptyState` `COPY` map `idle: 'en attente'` `loading: 'chargement…'` `empty: 'aucune donnée'` `stale: 'données périmées'` `offline: 'backend injoignable'` `error: 'erreur de source'` — French, sentence case, no filler.

**Form validation:** `Campagne` `Audit` `site` `required` + `link_type` `select` + `duration` `min:10 max:600` + `Import profil` `id` `required` `capacity>0` `delay>0` — inline `helper` + `error` below input, `gap-2`, `label` above input (craft-floor).

## 7. Testing & Verification

**The run must be green before any claim is made (ponytail: one runnable check per branch/loop/parser):**

| Layer | What we test | How | When it must pass |
|---|---|---|---|
| **Go unit** | `qdisc` arg vectors `handle 1:` stacked, `metrics` `Percentile`/`CostAR` golden `4.5 GiB==30000`, `campagne` `HappyPath` `BulkFails→invalid` `Writer G5+Freeze`, `matrix` order+resume, `probe` `SmallObject` httptest, `results` best `fq_codel`, `audit` `Run` | `go vet ./... && go test ./... -timeout 60s` (FakeRunner, no `tc` on host) | Every commit |
| **Go VM** | Real `tc` `netem`/`tbf`/`fq_codel`/`cake` + `BBR` `TCP_CONGESTION` `0x0d` + `testbed.sh up` | `make test-real` (`-tags=real` `go test ./pkg/qdisc -run TestReal` + `pkg/campagne -run TestReal`) on `192.168.174.128` via `deploy/engine.sh` `ssh` | Before any campaign |
| **Frontend unit** | `lttb` spike preservation, `CampagneView` `ArmButton` double-confirm, `gates` strip | `bunx vitest run` `jsdom` `src/**/*.test.ts` | Every commit |
| **Frontend type** | `noUnusedLocals` `strict` | `bun run typecheck` `tsc --noEmit` | Every commit |
| **Build** | `tsc` + `vite` + `node scripts/check-bundle.mjs` 450 KB guard + `echarts` 250 KB watchdog | `bun run build` | Every commit |
| **E2E smoke** | `shell renders four MeteoLink panels` + `live panel receives 10 Hz frames` (gates 8, banner not OFFLINE) | `npx playwright test` `baseURL http://192.168.174.128:9090` host→VM | After every deploy |
| **E2E screenshots** | 3 viewports `1920×1080`/`1366×768`/`390×844` ×4 panels =12 shots | `SHOT_DIR=C:/cgo/shots npx playwright test screenshots` `animations: 'disabled'` | Every deploy |
| **Live** | `GET /api/state` `phase`/`event_id`/`running`, `curl -N /api/stream` 10 frames/s ±20% + delta `profile` once, `GET /api/results` `available:true` + 12 groups, `GET /api/integrity` `manifests` `quarantined` | `deploy/engine.sh --action deploy` → `health` `curl /api/health` `{"ok":true}` → `POST /api/run/start` → poll `state` | After every deploy + after 36 finishes |
| **Integrity** | `manifest.json` SHA-256 chain, `quarantine.json`, `make figures` from frozen CSV | `./bin/cgo verify` + `POST /api/figures/regen` → `data/figures/small_p95.svg` `scatter.svg` with `RDF` provenance | After 36 `running:false` |

**Verification ledger (per milestone close, as in `docs/PLAN.md`):**

```text
host:   go vet ./... && go test ./... && cd web/frontend && bun run typecheck && bun run build && bunx vitest run
VM:     bash deploy/engine.sh --action ensure && bash deploy/engine.sh --action deploy && ssh 'cd ~/cgo && make test-real' && npx playwright test
live:   curl http://192.168.174.128:9090/api/health && curl -N http://192.168.174.128:9090/api/stream | head -n 30
```

**What we intentionally do not test:** `TTB` (needs per-second `backlog`), `per-flow p99` (needs per-flow state), `BBR states` (`TCP_INFO` per flow) — each would add a probe+store+chart for a secondary story.
