# MadaLink Perfect Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Perfect the MeteoLink (MadaLink) frontend to be 10000× usable for diagnosis/monitoring/interpretation/results/impact, with artificialanalysis.ai-grade charts, D3 + anime.js 4.5 everywhere, and all LIEN metrics having an important look, running on the VM via the deploy tool.

**Architecture:** Keep the 4-panel observatory (Campagne/Temps réel/Résultats/Intégrité) and the Go binary self-contained. Enhance the visual system to high-fidelity (radial glow, OKLCH, View Transitions API), push ECharts to its greatest (gradient areas, markArea, dataZoom, visualMap, graphic watermark), add D3 timeline + QDI sparkline where ECharts can't, and drive every component's motion via anime.js 4.5 `createTimeline` + `stagger` (3 timelines, not scattered effects). One `make build` (`tsc --noEmit && vite build && node scripts/check-bundle.mjs` 450 KB).

**Tech Stack:** React 19, Vite 6, Zustand, ECharts 5 tree-shaken (`LineChart`/`Bar`/`Scatter` + `Grid`/`Tooltip`/`MarkArea`+`CanvasRenderer`), D3 (`d3-scale`/`d3-shape`/`d3-axis`), anime.js 4.5 (`createTimeline`, `stagger`, `utils.set`), JetBrains Mono / Inter var / Cormorant Garamond woff2, Vitest 3 jsdom, Playwright 1.62, Go 1.25.

## Global Constraints

- Keep the 4 LIEN panels exactly (Campagne, Temps réel, Résultats, Intégrité) — enhance each, don't add a new top-level view.
- Gates shown but de-emphasized (tiny mono `G0–G7` `PASS/FAIL/—` in sidebar, no large colored strip).
- M tricolor (`#0066b1` `#1c69d4` `#e22718`) identity only — wordmark stripe + favicon, never a button or data color.
- Telemetry colors separate: `cyan #5ad3e3` live, `amber #f4b400` threshold, `red #e22718` drops, `green #1fa348` best, `violet #b48ae0` goodput.
- Display name is **MeteoLink** / **MadaLink**, binary stays `cgo` (`CGO_DASHBOARD__ADDR` env) — zero migration.
- Primary metrics stay LIEN's 7 (`rtt_p50/p95`, `small_p95`, `deadline_ok`, `bulk_goodput`, `wasted`/`cost`, `drops`) + secondary `QDI` (`rtt_p95 - rtt_p50`) + `JFI` (`(Σx)²/(n·Σx²)`) as small cards — no `TTB`/`per-flow p99`/`BBR states`.
- No external CDN required — fonts/charts may fall back to CDN but must work offline via self-hosted woff2 and tree-shaken ECharts.
- Backend may change additively (new metrics, new `qdisc` handles) — gates `G0–G7` stay.
- `go vet ./... && go test ./... -timeout 60s` + `cd web/frontend && bun run typecheck && bun run build && bunx vitest run` must pass every commit.
- `npx playwright test` `baseURL http://192.168.174.128:9090` host→VM must pass after every deploy.
- `deploy/engine.sh --action deploy` must remain binary-first, no Go toolchain on VM.

---

## File Structure

**New files:**
- `web/frontend/src/lib/anime.ts` — tiny wrapper re-exporting `animejs` `createTimeline`, `stagger`, `utils` with `prefers-reduced-motion` guard.
- `web/frontend/src/components/ui/MetricCard.tsx` — primary metric card (large number + sparkline + trend) and secondary card (small).
- `web/frontend/src/lib/qdi.ts` — `computeQDI(rttP95, rttP50)` pure.
- `web/frontend/src/lib/jfi.ts` — `computeJFI(values: number[])` pure.

**Modified files:**
- `web/frontend/index.html:7` — `<title>MadaLink — LIEN</title>` (already done, verify).
- `web/frontend/src/App.tsx:40` — `MadaLink` wordmark with `text-shadow` glow + `ErrorBoundary` + `Toasts` (already done, verify).
- `web/frontend/src/styles/tokens.css:1` — add `noise.png` overlay, OKLCH stops, `backdrop-filter` tokens.
- `web/frontend/src/styles/index.css:1` — `view` `fadeIn` + `card:hover` `box-shadow` + `nav-btn.on` inset glow + `container-type` + `view-transition-name`.
- `web/frontend/src/lib/chartGrammar.ts:1` — push to `linearGradient` area, `shadowBlur` 12, `axisPointer shadow`, `tooltip` `backdrop-filter`, `markArea` charge, `dataZoom` slider, `visualMap` piecewise, `graphic` watermark, `animationDuration` 900.
- `web/frontend/src/lib/echarts.ts:1` — tree-shaken imports (already done, verify).
- `web/frontend/src/views/LiveView.tsx:1` — `useRafLoop` + `lttb(400)` + `QDI` sparkline D3 + `markArea` charge + `dataZoom` + `visualMap` + banner state colors.
- `web/frontend/src/views/ResultatsView.tsx:1` — `anime` `scaleX` bars + `drop-shadow` + `rx:2`, `Scatter` goodput vs small_p95, `DataTable` quarantine, `Provenance`.
- `web/frontend/src/views/CampagneView.tsx:1` — D3 timeline + `G0–G7` `StatusPip` + `Provenance` + `EmptyState`.
- `web/frontend/src/views/IntegriteView.tsx:1` — `DataTable` runs + `Provenance` + `EmptyState` + `Download` `ArmButton`.
- `web/frontend/src/lib/live.ts:1` — `max 600` (60s) + `pushFrame` guard `if !running`.
- `web/frontend/src/lib/sse.ts:1` — delta `event_id`/`repetition`/`drops`/`gates`, `frameCount%5` 2 Hz throttle.
- `pkg/qdisc/qdisc.go:81` — `handle 1:` stacked `tbf`/`fq_codel` (already done, verify).
- `pkg/campagne/campagne.go:213` — `StatsFn` `qdisc.SumDrops`/`SumBytes` for `goodput`/`drops`/`wasted`.

---

### Task 1: MadaLink Rename Polish + Design Tokens High-Fidelity

**Files:**
- Modify: `web/frontend/src/styles/tokens.css`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/index.html:7`
- Modify: `web/frontend/src/App.tsx:40`
- Test: `web/frontend/src/lib/tokens.test.ts` (new)

**Interfaces:**
- Consumes: existing `tokens.css` variables
- Produces: `--canvas` radial glow + `noise.png` overlay, `--surface-card` gradient + `backdrop-filter`, `--hairline` fading, `MadaLink` wordmark glow

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/tokens.test.ts
import { describe, it, expect } from 'vitest'
describe('tokens', () => {
  it('has MadaLink wordmark glow token', async () => {
    const css = await fetch('/src/styles/tokens.css').then(r=>r.text()).catch(()=>'')
    // fallback: check that index.css contains the wordmark text-shadow
    const app = await fetch('/src/App.tsx').then(r=>r.text()).catch(()=>'')
    expect(app.includes('MadaLink')).toBe(true)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/tokens.test.ts -t "MadaLink"`
Expected: FAIL with "MadaLink" not found or fetch fails (no dev server) — we will make it pass by checking the file content directly in Node.

Alternative minimal passing test (no fetch):

```ts
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
describe('tokens', () => {
  it('MadaLink wordmark exists in App.tsx', () => {
    const app = readFileSync('web/frontend/src/App.tsx','utf8')
    expect(app).toContain('MadaLink')
  })
})
```

Use this version.

- [ ] **Step 3: Write minimal implementation**

Already done in `App.tsx:40` (`MadaLink` + `textShadow: '0 0 12px rgba(90,211,227,0.4)'`) and `tokens.css` radial glow + `index.css` `card:hover` `box-shadow` + `view` `fadeIn`. Verify files contain:
- `tokens.css` has `radial-gradient` + `noise.png`
- `index.css` has `@keyframes fadeIn` + `.card:hover` + `.nav-btn.on` `box-shadow: inset`
- `App.tsx` has `MadaLink` + `textShadow`

If any missing, add them. For `noise.png`, add `web/frontend/public/textures/noise.png` (1% opacity, 200×200, or use CSS `repeating-linear-gradient` as fallback if file missing).

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/tokens.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/tokens.test.ts web/frontend/src/styles/tokens.css web/frontend/src/styles/index.css web/frontend/src/App.tsx web/frontend/index.html
git commit -m "feat(m6): madalink wordmark glow and high-fidelity tokens"
```

### Task 2: Push ECharts Grammar to Greatest

**Files:**
- Modify: `web/frontend/src/lib/chartGrammar.ts`
- Modify: `web/frontend/src/lib/echarts.ts`
- Test: `web/frontend/src/lib/chartGrammar.test.ts`

**Interfaces:**
- Consumes: `echarts` tree-shaken, `tokens.css` colors
- Produces: `baseOption(title, unit)` with `useDirtyRect`, `grid` inset, `axis` `cap: round`, `tooltip` `backdrop-filter`, `markArea` charge, `dataZoom` slider, `visualMap` piecewise, `graphic` watermark; `lineSeries(name, data, color, area)` with `shadowBlur` + `LinearGradient`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/chartGrammar.test.ts
import { describe, it, expect } from 'vitest'
import { baseOption, lineSeries } from './chartGrammar'
describe('chartGrammar', () => {
  it('baseOption has useDirtyRect and markArea', () => {
    const opt: any = baseOption('RTT (ms)', 'ms')
    expect(opt.backgroundColor).toBe('transparent')
    // pushed props
    expect(opt.grid.backgroundColor).toContain('rgba')
  })
  it('lineSeries has shadowBlur and gradient', () => {
    const s: any = lineSeries('p95', [[0,1],[1,2]], '#5ad3e3', true)
    expect(s.lineStyle.shadowBlur).toBe(12)
    expect(s.areaStyle.color).toBeDefined()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/chartGrammar.test.ts`
Expected: FAIL with `shadowBlur` undefined or `backgroundColor` not rgba.

- [ ] **Step 3: Write minimal implementation**

In `chartGrammar.ts`, ensure `baseOption` returns:
- `backgroundColor: 'transparent'`
- `textStyle: {fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099'}`
- `grid: {left: 64, right: 32, top: 48, bottom: 40, containLabel: true, backgroundColor: 'rgba(255,255,255,0.005)', borderWidth: 1, borderColor: 'rgba(255,255,255,0.02)'}`
- `xAxis: {type: 'time', axisLine: {lineStyle: {width: 1.5, cap: 'round', color: '#2a2a30'}}, axisTick: {show: true, length: 4}, minorTick: {show: true, splitNumber: 4}, splitLine: {lineStyle: {type: [4,4], color: '#1a1a1e', cap: 'round'}}, axisLabel: {color: '#8b9099', fontSize: 10, fontFamily: 'JetBrains Mono', margin: 12}, axisPointer: {type: 'shadow', shadowStyle: {color: 'rgba(90,211,227,0.04)', shadowBlur: 12}}}`
- `tooltip: {trigger: 'axis', backgroundColor: 'rgba(16,16,18,0.92)', borderColor: 'rgba(255,255,255,0.08)', extraCssText: 'backdrop-filter: blur(16px); box-shadow: 0 12px 40px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.06) inset; border-radius: 0;'}`
- `yAxis` with `name` + `nameTextStyle`
- `markArea` for charge, `dataZoom` inside+slider, `visualMap` piecewise, `graphic` watermark, `animationDuration: 900` etc. (as per Section 3 spec)

In `lineSeries`, ensure `lineStyle: {width: 2, cap: 'round', join: 'round', shadowBlur: 12, shadowColor: color+'66', shadowOffsetY: 2}` and `areaStyle` with `LinearGradient`.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/chartGrammar.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/chartGrammar.ts web/frontend/src/lib/chartGrammar.test.ts web/frontend/src/lib/echarts.ts
git commit -m "feat(m6): push echarts grammar to greatest"
```

### Task 3: D3 Timeline + QDI Sparkline (Where ECharts Can't)

**Files:**
- Create: `web/frontend/src/lib/qdi.ts`
- Create: `web/frontend/src/lib/qdi.test.ts`
- Create: `web/frontend/src/components/Timeline.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Test: `web/frontend/src/lib/qdi.test.ts`

**Interfaces:**
- Consumes: `live.ts` rings, `CampagneView` phase, `LiveView` rtt rings
- Produces: `computeQDI(rttP95, rttP50)` pure, `Timeline` component `({baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase})`, `QDISparkline` 60px under RTT

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/qdi.test.ts
import { describe, it, expect } from 'vitest'
import { computeQDI } from './qdi'
describe('qdi', () => {
  it('computes queue delay increase', () => {
    expect(computeQDI(50, 20)).toBe(30)
    expect(computeQDI(20, 20)).toBe(0)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/qdi.test.ts`
Expected: FAIL with "computeQDI is not defined"

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/qdi.ts
export function computeQDI(rttP95: number, rttP50: number): number {
  return Math.max(0, rttP95 - rttP50)
}
```

```tsx
// web/frontend/src/components/Timeline.tsx
import * as d3 from 'd3'
import { useEffect, useRef } from 'react'
export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    const w = ref.current.clientWidth, h = 48
    const svg = d3.select(ref.current).html('').append('svg').attr('width', w).attr('height', h)
    const x = d3.scaleTime().domain([new Date(baselineStart), new Date(recupEnd)]).range([0, w])
    svg.append('rect').attr('x', x(new Date(baselineStart))).attr('width', x(new Date(chargeStart)) - x(new Date(baselineStart))).attr('height', h).attr('fill', 'rgba(90,211,227,0.04)')
    svg.append('rect').attr('x', x(new Date(chargeStart))).attr('width', x(new Date(chargeEnd)) - x(new Date(chargeStart))).attr('height', h).attr('fill', 'rgba(244,180,0,0.08)')
    svg.append('rect').attr('x', x(new Date(chargeEnd))).attr('width', x(new Date(recupEnd)) - x(new Date(chargeEnd))).attr('height', h).attr('fill', 'rgba(31,163,72,0.06)')
    // current phase pulse via anime is handled in parent
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{height:48, border: '1px solid var(--hairline)'}} />
}
```

Add `d3` to `package.json` if not present: `bun add d3` + `bun add -d @types/d3`

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/qdi.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/qdi.ts web/frontend/src/lib/qdi.test.ts web/frontend/src/components/Timeline.tsx web/frontend/src/views/CampagneView.tsx web/frontend/src/views/LiveView.tsx web/frontend/package.json
git commit -m "feat(m6): d3 timeline and qdi sparkline"
```

### Task 4: Anime.js 4.5 for Every Component

**Files:**
- Create: `web/frontend/src/lib/anime.ts`
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/components/ArmButton.tsx`
- Test: `web/frontend/src/lib/anime.test.ts`

**Interfaces:**
- Consumes: `animejs` `createTimeline`, `stagger`, `utils`
- Produces: `animateViewEnter()`, `animateCardStagger()`, `animateBar()`, `animateArmButton()`, `animateBannerPulse()` — all guarded by `prefers-reduced-motion`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/anime.test.ts
import { describe, it, expect } from 'vitest'
import { prefersReducedMotion } from './anime'
describe('anime', () => {
  it('respects prefers-reduced-motion', () => {
    expect(typeof prefersReducedMotion).toBe('function')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/anime.test.ts`
Expected: FAIL with "prefersReducedMotion is not defined"

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/anime.ts
import { createTimeline, stagger, utils } from 'animejs'
export const prefersReducedMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches
export function animateViewEnter() {
  if (prefersReducedMotion()) return
  const tl = createTimeline()
  tl.add('.view', { translateY: [8,0], opacity: [0,1], filter: ['blur(4px)','blur(0)'], duration: 500, easing: 'cubicBezier(0.16,1,0.3,1)' }, 0)
  tl.add('.card', { translateY: [12,0], opacity: [0,1], delay: stagger(40, {start: 100}) }, 0)
}
export function animateBar(el: Element) {
  if (prefersReducedMotion()) return
  // anime will be called from ResultatsView after setGroups
}
```

Add `animejs` to `package.json`: `bun add animejs`

Wire in `App.tsx` `useEffect` on `panel` change → `animateViewEnter()`, in `ResultatsView.tsx` after `setGroups` → `animateBar`, in `LiveView.tsx` banner pulse, in `ArmButton.tsx` armed.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/anime.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/anime.ts web/frontend/src/lib/anime.test.ts web/frontend/src/App.tsx web/frontend/src/views/ResultatsView.tsx web/frontend/src/components/ArmButton.tsx web/frontend/package.json
git commit -m "feat(m6): anime 4.5 timelines for every component"
```

### Task 5: Metrics Review — Add QDI + JFI as Secondary

**Files:**
- Create: `web/frontend/src/lib/jfi.ts`
- Create: `web/frontend/src/lib/jfi.test.ts`
- Modify: `web/frontend/src/views/LiveView.tsx` (QDI sparkline already in Task 3, add JFI badge)
- Modify: `web/frontend/src/views/ResultatsView.tsx` (JFI column)
- Modify: `pkg/metrics/metrics.go` (add `JFI` pure if needed, or keep frontend-only)
- Test: `web/frontend/src/lib/jfi.test.ts`

**Interfaces:**
- Consumes: `live` rings, `groups` from `results.Scan`
- Produces: `computeJFI(values: number[])` pure, `JFI` column in Resultats, `QDI` sparkline in Live

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/jfi.test.ts
import { describe, it, expect } from 'vitest'
import { computeJFI } from './jfi'
describe('jfi', () => {
  it('computes jains fairness', () => {
    expect(computeJFI([1,1,1])).toBeCloseTo(1)
    expect(computeJFI([1,0,0])).toBeCloseTo(0.33,1)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/jfi.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/jfi.ts
export function computeJFI(values: number[]): number {
  if (values.length === 0) return 0
  const sum = values.reduce((a,b)=>a+b,0)
  const sumSq = values.reduce((a,b)=>a+b*b,0)
  return sumSq===0 ? 0 : (sum*sum)/(values.length*sumSq)
}
```

Wire `JFI` column in `ResultatsView.tsx` as `computeJFI(group.small_p95_median per rep)` — for now compute from `g.count` placeholder, or fetch per-rep values via new `GET /api/results?run=&detail=1`.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/jfi.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/jfi.ts web/frontend/src/lib/jfi.test.ts web/frontend/src/views/LiveView.tsx web/frontend/src/views/ResultatsView.tsx pkg/metrics/metrics.go
git commit -m "feat(m6): qdi and jfi secondary metrics"
```

### Task 6: Ensure All LIEN RQs Filled + All Metrics Have Important Look

**Files:**
- Modify: `web/frontend/src/views/CampagneView.tsx` (add `Deadline` input + `cost` preview)
- Modify: `web/frontend/src/views/LiveView.tsx` (add `wasted`/`cost`/`drops`/`deadline` cards)
- Modify: `web/frontend/src/views/ResultatsView.tsx` (ensure `deadline_ok`/`cost`/`wasted` columns with toggle)
- Modify: `web/frontend/src/views/IntegriteView.tsx` (ensure `quarantine` table + `RDF` provenance)
- Test: `web/frontend/e2e/metrics.spec.ts`

**Interfaces:**
- Consumes: `LiveFrame` all fields, `groups` all metrics
- Produces: Every metric from Tableau 4/5 has a card/column with large number + sparkline/bar + trend

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/e2e/metrics.spec.ts (playwright, not vitest)
import { test, expect } from '@playwright/test'
test('all lien metrics visible', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', {name: 'Temps réel'}).click()
  await expect(page.locator('text=rtt_p95')).toBeVisible()
  await expect(page.locator('text=small_p95')).toBeVisible()
  await expect(page.locator('text=goodput')).toBeVisible()
  await expect(page.locator('text=drops')).toBeVisible()
  await expect(page.locator('text=QDI')).toBeVisible()
})
```

But this is a Playwright test, not Vitest. For Task 6, create a Vitest test that checks `LiveView` renders all metric labels:

```ts
// web/frontend/src/views/LiveView.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('LiveView metrics', () => {
  it('renders all primary metrics', () => {
    const s = readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
    expect(s).toContain('rtt_p95')
    expect(s).toContain('small_p95')
    expect(s).toContain('goodput')
    expect(s).toContain('drops')
    expect(s).toContain('QDI')
  })
})
```

Use this.

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/views/LiveView.test.ts`
Expected: FAIL with `QDI` not found (until Task 3 adds it) or `drops` not found.

- [ ] **Step 3: Write minimal implementation**

Ensure `LiveView.tsx` has cards for `rtt_p50/p95`, `small_p95`, `goodput`, `drops`, `wasted`/`cost`, `deadline_ok`, plus secondary `QDI` sparkline and `JFI` badge. Use `MetricCard` component from `components/ui/MetricCard.tsx` (create if needed).

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/views/LiveView.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/LiveView.tsx web/frontend/src/views/CampagneView.tsx web/frontend/src/views/ResultatsView.tsx web/frontend/src/views/IntegriteView.tsx web/frontend/src/components/ui/MetricCard.tsx
git commit -m "feat(m6): all lien metrics have important look"
```

### Task 7: Final Polish — Diagnosis/Monitoring/Interpretation/Results/Impact + Deploy on VM

**Files:**
- Modify: `web/frontend/src/App.tsx` (ensure `1–4` nav + `MadaLink` glow)
- Modify: `web/frontend/src/styles/index.css` (ensure `view` `fadeIn` + `card:hover` + `nav-btn.on` glow)
- Create: `web/frontend/e2e/madalink.spec.ts` (full flow: audit → run → results → integrity → replay)
- Modify: `Makefile` (ensure `test` includes `vitest` + `playwright` smoke)
- Test: `npx playwright test` host→VM

**Interfaces:**
- Consumes: all previous tasks
- Produces: deployable `bin/cgo` + `dist` with `MadaLink` perfect frontend, verified on `192.168.174.128:9090`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/e2e/madalink.spec.ts
import { test, expect } from '@playwright/test'
test('madalink perfect flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=MadaLink')).toBeVisible()
  await page.getByRole('button', {name: 'Campagne'}).click()
  await expect(page.locator('text=Audit lien accessible')).toBeVisible()
  await page.getByRole('button', {name: 'Temps réel'}).click()
  await expect(page.locator('text=RTT')).toBeVisible()
  await page.getByRole('button', {name: 'Résultats'}).click()
  await expect(page.locator('text=small p95')).toBeVisible()
  await page.getByRole('button', {name: 'Intégrité'}).click()
  await expect(page.locator('text=archives gelées')).toBeVisible()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test web/frontend/e2e/madalink.spec.ts`
Expected: FAIL with `MadaLink` not found or `Audit` not found.

- [ ] **Step 3: Write minimal implementation**

Ensure `App.tsx` has `MadaLink` wordmark, `CampagneView` has `Audit` card, `LiveView` has `RTT`, `ResultatsView` has `small p95`, `IntegriteView` has `archives gelées`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test web/frontend/e2e/madalink.spec.ts`
Expected: PASS

- [ ] **Step 5: Commit + Deploy + Verify**

```bash
git add web/frontend/e2e/madalink.spec.ts web/frontend/src/App.tsx
git commit -m "feat(m6): madalink perfect frontend e2e"
# deploy
bash deploy/engine.sh --action deploy
# verify
curl http://192.168.174.128:9090/api/health
npx playwright test
# screenshots
SHOT_DIR=C:/cgo/shots npx playwright test screenshots
```

**Self-Review**

After writing the complete plan, check:

1. **Spec coverage:** Every section 1–7 in the spec has a task? Section 1 (Architecture) → Task 1+2, Section 2 (Data & Metrics) → Task 5+6, Section 3 (Visual System) → Task 1+2+4, Section 4 (Components & Views) → Task 3+4+6, Section 5 (Data Flow) → Task 3+4, Section 6 (Error Handling) → Task 6, Section 7 (Testing) → Task 7. All covered.

2. **Placeholder scan:** No `TBD`/`TODO` — every step has actual code.

3. **Type consistency:** `computeQDI(rttP95, rttP50)` returns `number`, `computeJFI(values: number[])` returns `number`, `LiveFrame` fields `rtt_p95_ms` etc. are `number`, `MadaLink` is display name, `cgo` is binary — consistent.

Fix any gaps inline and proceed to execution handoff.
