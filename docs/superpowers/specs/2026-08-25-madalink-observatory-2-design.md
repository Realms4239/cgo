# MadaLink Observatory 2.0 — High-End NOC Overhaul

**Date:** 2026-08-25
**Status:** Approved (Sections 1–7)
**Base:** `agent/m6-perfect@da7fadd` (7 tasks: tokens glow, grammar push, D3 timeline, anime 4.5, QDI/JFI, MetricCard, css alias + truth fixes)
**Source:** User discovery 2026-08-25 — cramped, sidebar dense, views disparate, missing anime/assets, toast/flash/validation, ephemeral prompt, usability; LIEN.md, docs/SPEC.md, docs/PLAN.md, DESIGN.md, donor thesis-cgo
**Approach:** 1 Hybrid Observatory 2.0 — rail 56↔232 animated + bento airy/density + lightweight WebGL mesh + ECharts max + D3 brush + every component anime dense (all guarded)

## 1. Architecture

**Goal:** One instrument where diagnose→run and monitor→interpret are each 2 clicks, every view feels connected yet has a reason to be separate, every pixel NOC-grade. Both loops, operator wins on conflict.

**Stack on top of da7fadd:** React 19 + Vite 6 + Zustand low-freq only (`panel`, `railPinned`, `density`, `live`, `connected`, `promptOpen`, `paletteOpen`, `flash`) + ECharts 5 tree-shaken (LineChart/Bar/Scatter + Grid/Tooltip/DataZoom/VisualMap/Graphic + CanvasRenderer `useDirtyRect`) + D3 partials `d3-scale/d3-shape/d3-axis/d3-brush` (timeline, QDI sparkline, scatter brush) + `anime.js 4.5` `createTimeline/stagger/utils` dense + lightweight WebGL mesh `1×` under shell (radial glows + `noise.png` overlay) + `JetBrains Mono / Inter var / Cormorant Garamond` woff2 self-hosted (CDN fallback allowed, offline flexible).

**Shell — innovative yet ponytail (fewest deps, one mesh, one rail):**
- `web/frontend/src/App.tsx` grid `48px header glass + rail 56px ↔ 232px (animated width, not display) · main bento · 28px footer`. `railPinned` persisted `localStorage`. `header` `backdrop-filter blur16 saturate1.5` sticky; `main` `gap var(--gap) 24px airy / 16px dense` toggle (`localStorage density`), `max-width 1280` centered; `view-transition-name: main` for morph; `container-type:inline-size` on cards.
- `web/frontend/src/store/ui.ts` small low-freq; high-freq `lib/live.ts` rings `max 600 (60s @10Hz)` + `lttb400`, `lib/sse.ts` 10 Hz ingest → `frameCount%5 + gatesEqual → 2 Hz setLive` (already fixed 10 Hz churn).
- `pkg/*` additive only (`Scan` wasted/cost real already in `95028ae`); `cmd/cgo` binary stays `cgo`, display `MadaLink`, `CGO_DASHBOARD__ADDR`.

**Why this split:** ECharts where 10 Hz matters (Live), D3 where brush/clip/scrub matters (Timeline, QDI, scatter compare), anime where meaning matters (every component but guarded), one WebGL mesh under `body::before` (not per chart) so `289 → ~299 KB gz`.

## 2. Data & Metrics

**Primary (LIEN Tableau 4 — stay primary):** `rtt_p50/p95` ping 5 Hz `metrics.Summarize`, `small_p95` 4–32 KiB `probe.SmallObject`, `deadline_ok_pct` 1000 ms `metrics.DeadlineOK`, `bulk_goodput_mbps` `tc -s SumBytes/dur` receiver, `wasted_bytes`/`cost_ar_per_h` `SumDrops×1448` real via `pkg/results.Scan` median (`95028ae` honest, no `as any` fallback), `drops` `SumDrops`. Each shown as `MetricCard` large `20px mono + sparkline + trend` amber/cyan/red/green/violet.

**Secondary (Tableau 5):** `QDI = rtt_p95 - rtt_p50` amber 60px brushable sparkline under RTT (`d3-line/area clipPath rx4 + brushX → live.max`), `JFI = (Σx)²/(n·Σx²)` steel badge `0.00–1.00` (`—` when `count<2` or all-equal, ponytail until `detail=1`). No `TTB`/`per-flow p99`/`BBR states` (needs store/probe churn).

**Storage:** `aqm_eval.csv 17 cols AQMEvalHeader + link_audit.csv 14 cols LinkAuditHeader` order enforced; `writer.go` `Freeze → manifest.json sha256 + quarantine.json`; `results.Scan` groups `profile|qdisc|cc → median/IQR + Best` per profile; Live `Snapshot{phase,profile,qdisc,cc,rep,event_id,rtt_*,small_*,goodput,drops,gates,wasted_bytes,cost_ar,deadline_ok_pct,gates,running} → Live.Set → sse.Hub delta + ts → rings`.

**Open panels:** 4 panels stay the views (`Campagne / Temps réel / Résultats / Intégrité`), openness = ephemeral 5th layer: `Command palette Cmd-K` + `Quick-actions prompt 6s center` + `hover peek popovers` + `replay stream` — no new `PANELS` entry, no migration. `gates G0–G7` tiny mono `PASS/FAIL/—` in rail, not a strip. `M tricolor #0066b1 #1c69d4 #e22718` identity stripe only.

## 3. Visual System

**Tokens (from DESIGN.md observatory, pushed + rail/density/webgl):**

| Token | Value | Fidelity |
|---|---|---|
| `--canvas` | `radial cyan 4% top + violet 3% bottom-right + #070707` + `noise.png mix-blend overlay 0.015 fixed` + lightweight WebGL mesh `<canvas fixed inset -1 opacity 0.015>` single gradient mesh | physical paper + depth, one canvas |
| `--surface-card` | `linear white 2%→transparent + #101012 + backdrop-filter blur12 saturate1.15 + inset 1px white 4% + outer 8×32 soft` | layered glass |
| `--hairline` | `linear transparent→#26262a 20%→#26262a 80%→transparent 1px` | faded instrument |
| `--rail-w` | `56px ↔ 232px` `dur 400 easeOutCubic` animated width | not display toggle |
| `--gap` | `24px airy / 16px dense` bento `localStorage` | breathing toggle |
| Telemetry | `cyan #5ad3e3 live, amber #f4b400 QDI, red #e22718 drops, green #1fa348 best, violet #b48ae0 goodput, steel #9aa3ad JFI` with `oklch` stops via `color-mix` | perceptual |
| Type | `Cormorant 600 +0.04→0.08em hover + gradient text-clip + drop-shadow 12px + shimmer 3s` / `JetBrains Mono tnum ss02` / `Inter var` | artificialanalysis.ai titles |
| Spacing | `--sp-1 4 --sp-2 8 --sp-3 12 --sp-4 16 --sp-5 24` base 4px |  |
| Motion | `--dur-fast 80 --dur-base 150 --ease cubicBezier(0.4,0,0.2,1)` | guarded |

**Layout:** `header 48px glass sticky + rail flex col thin scrollbar + main max 1280 centered + footer 28px tabular`; `view view-transition-name:main ::view-transition-old/new fade 0.3/0.5 cubicBezier(0.16,1,0.3,1)`; card `container-type:inline-size @container>600px pad24`; `panel-stack grid gap var(--gap) subgrid` for kv alignment.

**Chart grammar — ECharts max + D3 brush focus (pushed):**

| Feature | Pushed value |
|---|---|
| Renderer | `canvas useDirtyRect true devicePixelRatio min(dpr,2)` at `echarts.init` (not setOption) |
| Grid | `left64 right32 top48 bottom40 containLabel true backgroundColor rgba255 0.005 border 0.02` inset hairline |
| Axis | `axisLine width1.5 cap round #2a2a30 tick length4 #3a3a40 minorTick splitNumber4 splitLine [4,4] #1a1a1e cap round label #8b9099 Mono 10 margin12 axisPointer shadow rgba cyan 0.04 shadowBlur12` |
| Tooltip | `axis trigger background rgba16 0.92 blur16 border 0.08 padding [16,20] extraCssText backdrop-filter blur16 shadow inset border-radius0 text Mono11 lineHeight18 formatter min-width180 flex justify-between gap24 tabular + QDI/JFI footer border #26262a hint` |
| Line | `smooth0.4 monotoneX sampling lttb lineStyle width2 cap round join round shadowBlur12 offsetY2 area LinearGradient 26→0% opacity .8 focus series width3 blur opacity0.2 markPoint max` |
| MarkArea | `per-series itemStyle rgba amber 0.04 border amber 0.12 dashed label amber Mono10 insideTop padding [4,8] bg amber 0.08` data `[[{xAxis:chargeStart},{xAxis:chargeEnd}]]` (never at option root) |
| DataZoom | `inside filter none wheel/move + slider height24 handleIcon patch handleStyle #f2f2f4 border #26262a bg #0b0b0c filler rgba cyan 0.12 selectedDataBackground line #5ad3e3 area 0.15 emphasis border #5ad3e3` |
| VisualMap | `show false piecewise dimension1 pieces gt100 red gt40 amber lte40 cyan outOfRange steel` severity |
| Graphic | `text center top12 MeteoLink·LIEN fill 2.5% 56px Cormorant + image noise 1% 400×400 0.015 silent` |
| Animation | `animation false on Live (anime clipPath drives container inset 0 100%→0 600), elsewhere animationDuration 900 delay idx*15` + `prefers-reduced-motion → 0` |

**D3 enhanced (where ECharts can't):** Timeline `scaleTime domain [baselineStart,recupEnd] range[0,w] + axisBottom timeFormat + area cyan/amber/green + circle ticks r3 + star best + isCurrent stroke + brushX when real phase times available + hidden when idle`; QDI sparkline `line monotoneX + area amber gradient + clipPath rx4 + brushX → live.max`; Scatter `brush toolbox compare`.

**Motion — rich dense every component, context7 /websites/animejs verified (1983 snippets):** `createTimeline {defaults:{duration500 ease cubicBezier(0.16,1,0.3,1)}} → add('.view',{y[8,0],opacity[0,1],filter[blur4→0]},0) + add('.card',{y[12,0],opacity}, stagger40 start100)` bars `scaleX left stagger30 rx2 drop-shadow`, rail `width 56↔232 400`, prompt `y[12,0] blur stagger40` enter `y[0,12] opacity` exit `6s auto`, toast `y[16,0] slide stack stagger20 3.2s`, validation `x[-4,4,0] 400 elastic + borderColor`, ArmButton `scale[0.96,1] rotate[0.5,-0.5,0] 400 easeOutElastic + shadow pulse`, banner `scale[1,1.015,1] bg rgba 2 Hz`, chart container `clipPath inset(0 100%→0) 600`. Every export `if(prefersReducedMotion()) return` (`matchMedia('(prefers-reduced-motion: reduce)')`).

**Assets both rich (lazy):** `public/textures/noise.png 200×200 1% + meteo-station.jpg + topography.svg + waveform mesh + icons/meteo-*.svg 16×16 stroke1.5 round/round` for rtt/small/goodput/drops + `gate pictos + bbr/cubic badges + quarantine stamps` + `fonts woff2` + `gradient orbs` glass reflection.

## 4. Components & Views

**4.1 Carried tokens:** `--canvas` mesh+noise, `--surface-*`, `--hairline-fade`, `--rail-w`, `--gap` airy/dense, `MadaLink` glow `0 0 12px rgba cyan 0.4`, 6 telemetry `oklch`, fonts, `48/232↔56/28` grid, `card` glass, `ArmButton` double-confirm, `banner`.

**4.2 Enhanced views — interconnected yet separated for a reason:**

| View | Perfected 2.0 — why separate | Connection + preview |
|---|---|---|
| **Campagne** `1` | **Diagnose:** selectors `P1/P2 reps` + `Deadline disabled hint 1000 ms` + `cost preview —→value` + D3 Timeline `48px hairline scaleTime 3 areas + ticks + isCurrent pulse + hidden idle` + `G0–G7 StatusPip + Provenance` + `Audit form required/link_type/duration 10–600` inline `helper/error + focus ring + shake` + `Import profil` validation + `PhaseStep` + `EmptyState offline` | Hover `event row → Live sparkline peek popover 160×60` no nav. Démarrer success → `toast + FlashBanner CHARGE` → **auto Live** |
| **Temps réel** `2` | **Monitor:** 3 ECharts `Canvas init useDirtyRect` (`RTT small goodput` max grammar per-series markArea + graphic watermark + dataZoom slider + visualMap on small) + D3 QDI `60px brush rx4 amber` under RTT + `banner CHARGE/REPLAY 2 Hz pulse` + `drops MetricCard large 20px mono + sparkline + trend` + `StatusPip live/err + Provenance 10 Hz` `animation false` | Hover `Live point → Résultats row` value popover. `idle 8s → prompt Voir Résultats 6s` |
| **Résultats** `3` | **Interpret:** Table `profile|qdisc|cc|n|small p95|rtt p95|goodput|JFI —|quar|best ★ + cost/wasted toggle` + bar `anime scaleX left stagger30 rx2 shadowBlur6 green best` + Scatter `goodput vs small symbol8 best12 star brush toolbox compare` + DataTable quarantine drill sortable `aria-sort` + EmptyState + Provenance | Hover `quarantined row → Audit site tooltip` + `★ best → topology badge`. `running:false → auto Résultats + toast Campagne terminée + prompt Rejouer 6s` |
| **Intégrité** `4` | **Impact:** `valid/quarantined StatusPip ok/err + Provenance manifest n + EmptyState offline` + `Runs DataTable run_id/manifest/quarantine Provenance` + `Figures 2 SVG RDF sha256 + Download ArmButton` + `Replay stream → Live` | Hover `run_id → sparkline of its group medians`. Quick-actions anywhere can `Rejouer` without visiting. |

**4.3 New shared primitives (dense motion, validation, connection):**
- `CommandPalette.tsx` `Cmd-K` `createTimeline y[8,0] opacity stagger20 filter blur` dialog `aria-modal` trap `Esc` spring `400`, filters 4 panels + 3 quick actions + `sseStatus`.
- `QuickActionsPrompt.tsx` `center bottom fixed card y[12,0] blur4→0 stagger40 enter + y[0,12] opacity exit 6s auto` (hover pauses via `clearTimeout`), contextual actions `idle→[Démarrer,Audit] running→[Live,pause peek] done→[Résultats,Rejouer]`, reappears `idle 8s` `useInterval`, `localStorage nudge_seen` guard for onboarding.
- `OnboardingNudge.tsx` once `localStorage 2026-08-26` steps `Audit→Démarrer→Live` `PhaseStep+Lamp` `8s` auto-dismiss if untouched `stagger 40`.
- `FlashBanner.tsx` top `slide down 300 easeOutCubic state success/danger/info 2.5s auto`.
- `useValidation.ts` + `InlineField.tsx` `label→input→helper/error gap2 borderColor var(--t-danger) on invalid shake x[-4,4,0] 400 elastic + focus ring`, `All + optimistic`: `ArmButton disabled until valid`, `toast optimistic + FlashBanner success` on submit, rollback `err` if `!ok`, `Esc` clears.
- `Rail.tsx` `56↔232 anime width 400` + `miniSparkline 32×12 canvas` when collapsed + `nav-btn on inset shadow` + `pin` button.
- Carried: `lttb, useRafLoop/useInterval, EmptyState (idle/loading/empty/stale/offline/error) COPY fr, Provenance live/err/wait, StatusPip, DataTable sortable, ErrorBoundary chunk-reload guard, Toasts bottom-right slide 3.2s blue backpressure, Skeleton PhaseStep, MetricCard large+secondary, Timeline, QDI, anime qdi/jfi, gatesEqual` all guarded.

**4.4 Safely fetch vs leave:** QDI/JFI yes, TTB/per-flow p99/BBR states/e1_onsets 356 events leave (store/probe churn); wasted/cost already real via `Scan+Snapshot` (no new `TCP_INFO` yet); `lttb/hooks/EmptyState/Provenance/StatusPip/DataTable/Toasts/ErrorBoundary/Skeleton` already ported — validation uses native `required min max` + `anime` shake, no new lib.

## 5. Data Flow & Real-time

```
Probe ping×5 + Small×1 + BulkReceive per 300 ms collect
  → metrics.Summarize median/p95/IQR + DeadlineOK + CostAR
  → campagne.RunEvent → gates G0–G7 → Snapshot{phase,profile,qdisc,cc,rep,event_id,
      rtt_p50/95,small_p95,goodput,drops,wasted_bytes,cost_ar,deadline_ok_pct,gates,running}
  → writer.Append seen-G5 dedup → Freeze manifest.json sha256 + quarantine.json
  → results.Scan profile|qdisc|cc → median/IQR + Best per profile (wasted_median/cost_median real)
  → figures.Generate bar+scatter SVG + RDF provenance
  ↓ Live 10 Hz (stored 1s, streamed 10Hz):
  Deps.OnSnap(Snapshot) → Live.Set → sse.Hub.Publish delta(profile/qdisc/cc/rep/phase)+ts+ring 2048 fanout
  → EventSource /api/stream parse delta-retain → pushFrame if running → frameCount%5 + gatesEqual → setLive 2 Hz → Toasts/Flash if phase change
  → live.ts rings max600 lttb400 → useRafLoop 4 Hz → ECharts setOption lazy + anime clipPath inset(0 100%→0)
  ↓ Smart connect (2.0):
  POST /api/run/start 200 → FlashBanner success slide-down + Toasts + QuickActionsPrompt dismiss + store setPanel('live') auto → banner CHARGE pulse 2 Hz
  running:false → FlashBanner 'Campagne terminée' + Toasts + auto setPanel('resultats') + Prompt 'Rejouer 6s'
  Cmd-K → CommandPalette dialog filter 4 panels + 3 quick actions → setPanel + anime y[8,0] stagger20
  hover Campagne row / Live point → portal PeekPopover no setPanel (data live.ts rings / results groups) 160×60 sparkline
  ↓ Replay: GET /api/replay/list → GET /api/replay/stream?run → 400 ms per row → sse pause replayRunning guard → banner REPLAY cyan
```

**Invariants:** Browser never derives (Go `pkg/metrics` pure); rings mutable not Zustand (2 Hz React, 4 Hz charts, 10 Hz ingest); `Last-Event-ID` + `quarantine.json` resume via `seen`; `Hub sub.ch 16 default→dropped++ backpressure`; `useDirtyRect` at `echarts.init`; every `anime` `prefersReducedMotion` early; rail width animates not display; prompt hover pauses timer.

## 6. Error Handling & Empty States

| State | Where | Operator sees | Handling |
|---|---|---|---|
| Loading | Resultats Intégrité Live first 2s | `EmptyState loading 'chargement…' + Skeleton 3-line shimmer + Provenance wait` | fetch null → loading |
| Empty | Resultats !groups Intégrité runs 0 | `EmptyState empty 'disponible après gel (M2)' + Provenance n=0 wait + ArmButton Démarrer` | `available:false` never fakes |
| Offline | Live Campagne connected false | `EmptyState offline 'backend injoignable — vérifiez VM 192.168.174.128:9090' + banner OFFLINE red + Toasts err + FlashBanner danger` | `EventSource.onerror → false → Flash` |
| Stale | Live 5s no frame | `EmptyState stale 'données périmées — dernier év. il y a 5 s' + banner STALE amber` | `useInterval 5s ts compare` |
| Error | POST /api/run/start 409 /api/audit/start 409 /api/profile/import 400 | `EmptyState error hint={message} + ErrorPanel + Toasts err + FlashBanner danger 2.5s + ArmButton retry + inline field shake` | `!r.ok throw→catch setErr+pushToast` |
| Invalid/Degraded | Résultats quar + Intégrité valid/quar + quarantine.json | `StatusPip err/warn + DataTable gate_status sortable + Provenance n + FlashBanner warn when quar>0` | `Scan gate_status!="valid"` |
| Chunk | any lazy view | `ErrorBoundary CHUNK_ERR → sessionStorage guard auto-reload + Recharger` | already ported |
| Backpressure | Live Hub drop | `banner BACKPRESSURE amber + Toasts blue backpressure(N) + FlashBanner info` | `sub.ch 16 default→dropped event` |
| Validation | Audit site required + link_type select + duration 10–600 + Campagne profiles/reps + Import id/cap>0/delay>0 | `InlineField gap2 label above + focus ring + shake x[-4,4,0] 400 elastic + ArmButton disabled until valid + helper 'requis'` | `useValidation.ts native required min max + anime shake`, All + optimistic: optimistic Toast+Flash success on submit, rollback err if !ok, Esc clears |
| Prompt | Quick-actions 6s + Onboarding once | `center bottom card y[12,0] blur4→0 stagger40 + 4 buttons + Esc / hover pauses + auto slide-down dismiss` | `setTimeout 6s + idle 8s reappear + localStorage nudge_seen` |

Copy: `COPY idle en attente / loading chargement… / empty aucune donnée / stale données périmées / offline backend injoignable / error erreur de source` fr.

## 7. Testing & Verification

| Layer | What we test (2.0 adds) | How | Must pass |
|---|---|---|---|
| Go unit | qdisc handle1: stacked, metrics Percentile/CostAR 4.5GiB==30000, campagne HappyPath/BulkFails→invalid/Writer G5+Freeze, matrix order+resume, probe httptest, results best+median, audit Run | `go vet ./... && go test ./... -timeout 60s` | Every commit |
| Go VM | real tc netem/tbf/fq_codel/cake + BBR 0x0d + testbed.sh up | `make test-real` on `192.168.174.128` via `deploy/engine.sh ssh` | Before campaign |
| Frontend unit | lttb, qdi/jfi, chartGrammar pushed, tokens rail/gap, anime prefersReducedMotion, LiveView.test rtt/small/goodput/drops/QDI/wasted/cost, **new: CommandPalette Cmd-K filter, QuickActionsPrompt 6s dismiss, useValidation shake, Rail width, FlashBanner** | `bunx vitest run` jsdom | Every commit |
| Type | strict noUnusedLocals | `bun run typecheck tsc --noEmit` | Every commit |
| Build | tsc+vite+check-bundle 450 KB + echarts 250 KB + WebGL 1 canvas | `bun run build` `~299 KB gz` | Every commit |
| E2E smoke | `shell rail 56↔232 + 4 panels + SSE 10 Hz + prompt 6s + palette Cmd-K + validation shake + smart redirect Campagne→Live→Résultats` | `npx playwright test baseURL http://192.168.174.128:9090` host→VM | After deploy |
| Screenshots | `1920/1366/390 ×4 panels 12 shots + rail collapsed/expanded + prompt + flash + validation + hover peek` animations disabled | `SHOT_DIR=C:/cgo/shots npx playwright test screenshots` | Every deploy |
| Live | `/api/state phase/event/running, curl -N /api/stream 10 frames/s ±20% delta once, /api/results 12 groups, /api/integrity manifests` | `deploy deploy → health curl → POST /api/run/start poll` | After deploy + 36 end |
| Integrity | `manifest.json sha256 chain + quarantine.json + POST /api/figures/regen → svg RDF` | `./bin/cgo verify` | After 36 running:false |
| a11y | skip-link, nav-btn aria, gate aria-label, dialog aria-modal, toast aria-live, focus ring, palette trap | `playwright + manual` | Every deploy |

Ledger:
```
host: go vet && go test && cd web/frontend && bun run typecheck && bun run build && bunx vitest run
VM:   bash deploy/engine.sh --action ensure && bash deploy/engine.sh --action deploy --config deploy/cgo-vm.yaml && ssh 'make test-real' && npx playwright test
live: curl http://192.168.174.128:9090/api/health && curl -N .../api/stream | head -n 30
```

What we don't test: TTB/per-flow p99/BBR states (still leave).

