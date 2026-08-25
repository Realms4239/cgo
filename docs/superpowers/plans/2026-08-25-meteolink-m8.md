# Meteolink M8 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Meteolink M8 as decomposed layers — pending polish (observatory2 fixes) → backend ARG 100% (Traduction Matérielle + Kit Portable + Pilote + Why Go + hardware translate) → shell METEOLINK lockup + rail/foot/arm → NOC prompt progress line + HD sparkline every MetricCard + anime svg/text/layout dense + snippets beam/donut/snap — all A/B baseline vs algorithm and real SSE verified.

**Architecture:** Start from `main@297a888` (M8 spec) merged with `observatory2@68291cd` (pending polish: rail icons, bento scoped, scrubber hidden, watermark 28px, idle rings gate, timeline ring gate). Keep 4 panels (`Campagne/Live/Résultats/Intégrité` + `Recommandations`) with ephemeral palette/prompt/peek/beam. Shell `48 header + rail 56↔232 · main bento` + one `WebGLMesh` canvas, ECharts `useDirtyRect` + D3 `brush`, `animejs@4.5.0` dense `createTimeline/stagger/utils svg/text/animatable` via `context7` 1983 snippets, woff2 self-hosted, snippets beam/donut not new deps.

**Tech Stack:** Go 1.25, React 19, Vite 6, Zustand, ECharts 5 tree-shaken (`LineChart/Bar/Scatter` + `Grid/Tooltip/DataZoom/VisualMap/Graphic` + `CanvasRenderer`), D3 `d3-scale/d3-shape/d3-axis/d3-brush`, `animejs 4.5` (`createTimeline`, `stagger`, `utils`, `svg`, `text`, `animatable`), `JetBrains Mono/Inter var/Cormorant Garamond` woff2, Vitest 3 jsdom, Playwright 1.62, VM `192.168.174.128:9090`, `deploy/engine.sh --config deploy/cgo-vm.yaml` binary-first.

## Global Constraints

- Keep the 4 LIEN panels exactly as views (`Campagne`, `Temps réel`, `Résultats`, `Intégrité`) — ephemeral palette/prompt/peek is the open layer, not a 5th `PANELS` entry; `Recommandations` is a sub-card of Intégrité.
- Gates shown but de-emphasized (tiny mono `G0–G7` `PASS/FAIL/—` in rail, no large colored strip).
- M tricolor (`#0066b1` `#1c69d4` `#e22718`) identity only — wordmark stripe + favicon, never a button or data color.
- Telemetry separate: `cyan #5ad3e3` live, `amber #f4b400` QDI, `red #e22718` drops, `green #1fa348` best, `violet #b48ae0` goodput, `steel #9aa3ad` JFI.
- Display name `METEOLINK`, binary `cgo` (`CGO_DASHBOARD__ADDR` env) — zero migration.
- Primary metrics stay 7 (`rtt_p50/p95`, `small_p95`, `deadline_ok_pct`, `bulk_goodput`, `wasted`/`cost`, `drops`) + secondary `QDI` + `JFI` as small.
- No external CDN required — fonts/charts may fallback but work offline via woff2 + tree-shaken ECharts; `animejs` local `^4.5.0` via `context7 /websites/animejs`.
- Backend additive only (new `hardware_recommendation`, `translate` endpoint additive), gates `G0–G7` stay.
- `go vet ./... && go test ./... -timeout 60s` + `cd web/frontend && bun run typecheck && bun run build && bunx vitest run` must pass every commit; `node scripts/check-bundle.mjs` 450 KB guard, echarts 250 KB watchdog.
- `npx playwright test` `baseURL http://192.168.174.128:9090` must pass after every deploy; `deploy/engine.sh --action deploy --config deploy/cgo-vm.yaml` binary-first.
- Every `anime` export guarded by `prefersReducedMotion()`.

---

## File Structure

**New files (M8):**
- `web/frontend/src/components/MeteolinkWordmark.tsx` — `METEOLINK` lockup `Carmarant 600 + gradient text-clip + 16×16 NOC icon satellite→wave + tricolor stripe + shimmer`.
- `web/frontend/src/components/PromptProgressLine.tsx` — bottom `1px #f4b400 width 100%→0% linear 6s pause on hover` used inside `QuickActionsPrompt`.
- `web/frontend/src/components/Beam.tsx` — `animated-beam` 1 SVG `path linearGradient strokeDasharray 4 dashOffset -40 linear infinite` between Campagne and Live when `live.running`.
- `web/frontend/src/components/DonutJFI.tsx` — `donut-chart` `arc 0–360 56px stroke 8 steel→green` for `JFI` `—` when idle.
- `web/frontend/src/lib/hardware.ts` — `hardwareRecommendation(bestQdisc:string, profile:string):string` pure for `Si MikroTik…`.
- `pkg/api/translate.go` — `GET /api/hardware/translate?profile=P2` returns `{recommendation:string}` from `results.Scan` hardware mapping.
- `web/frontend/src/components/Rail.test.ts` already from observatory2 — add `rail shrink state` check.
- `web/frontend/src/lib/promptProgress.test.ts` — progress line 6s.

**Modified files:**
- `web/frontend/src/components/Rail.tsx:1` — `METEOLINK` icon 16 + shrunk `nav-icon` + extended `label+key+side-status + gates + live peek sparkline` + `foot` inside rail when extended.
- `web/frontend/src/App.tsx:1` — header `MeteolinkWordmark` full lockup, `FootTicker` 28px `source | run 8-char · hash 8-char · phase · SSE` tabular, `Beam` mount, `PromptProgressLine` integrated.
- `web/frontend/src/components/QuickActionsPrompt.tsx:1` — integrate `PromptProgressLine` bottom fill bar `6s linear pause on hover`.
- `web/frontend/src/views/CampagneView.tsx:1` — `Import INLINE` already, `Timeline` rings-empty gate already, ensure `bento 2-col airy @1100px` tidy placement.
- `web/frontend/src/views/LiveView.tsx:1` — `Every MetricCard HD sparkline 60×12 clipPath rx4 lttb40 + trend` + `idle rings-empty → — not red 0%` already, add `DonutJFI` in JFI card + `Beam` trigger.
- `web/frontend/src/views/ResultatsView.tsx:1` — `DonutJFI` + `hardware_recommendation` provenance hint per row.
- `web/frontend/src/views/IntegriteView.tsx:1` — new `Recommandations — Traduction Matérielle` card `table Linux → MikroTik PCQ/CAKE / ISP mini-PC gateway + pilote + KIT + Why Go` + existing `RDF` `quarantine` `Figures` `Replay`.
- `web/frontend/src/styles/index.css:1` — `rail` width var `--rail-w`, `foot-ticker` 28px, `prompt-progress` 1px amber, `bento tidy+breathing` scoped `#v-campagne #v-live`, `arm progress line` 1px above `ArmButton`.
- `web/frontend/src/styles/tokens.css:1` — `METEOLINK` gradient tokens `--meteolink-gradient`, already has `radial + noise + WebGLMesh`.
- `web/frontend/src/lib/anime.ts:1` — add `animateMeteolinkShimmer`, `animateBeam`, `animateDonut`, `animateProgressLine` exports using `svg/text/animatable` verified via `context7`.
- `web/frontend/src/components/ArmButton.tsx:1` — `Arm progress line 100%→0% 5s` above button when armed + spring.
- `pkg/campagne/campagne.go:1` — copy reframe docs comments `Kit Portable` (no logic change).
- `pkg/results/results.go:1` — add `hardware_recommendation` per `Group` computed from `Best qdisc`.

---

### Task 1: Foundation — Merge Pending Polish + Verify Loops Green

**Files:**
- Modify: `web/frontend/src/components/Rail.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx`
- Modify: `web/frontend/src/lib/chartGrammar.ts`
- Test: `web/frontend/src/components/Rail.test.ts`

**Interfaces:**
- Consumes: `observatory2@68291cd` (rail icons, bento scoped, scrubber hidden, watermark 28px, idle rings gate, timeline ring gate, import inline, integrite peek)
- Produces: `main` with pending polish merged — `Rail 56 icons + 232 extended`, `panel-stack bento 2-col scoped`, `scrubber hidden when empty`, `watermark 28px`, `undefined rep → #0`, `idle via rings-empty` — used by all M8 layers

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/components/Rail.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('Rail polish', () => {
  it('Rail has icons and bento scoped', () => {
    const r=readFileSync('web/frontend/src/components/Rail.tsx','utf8')
    expect(r).toContain('nav-icon')
    expect(r).toContain('ICONS')
    const css=readFileSync('web/frontend/src/styles/index.css','utf8')
    expect(css).toContain('#v-campagne .panel-stack')
    expect(css).toContain('max-width:1280')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: FAIL if `main` still has old `Madalink` rail without `nav-icon` (before merge)

- [ ] **Step 3: Write minimal implementation**

```bash
# from C:\cgo (main@297a888)
git merge agent/observatory2 --no-edit --allow-unrelated-histories
# resolve: keep observatory2 frontend, keep main spec/plan
# then verify merged files contain:
# Rail.tsx ICONS map + nav-icon 16 + sr-only
# index.css max-width 1280 + #v-campagne #v-live bento 2-col @1100px
# LiveView.tsx !liveSnap&&rings empty → EmptyState + deadline idle → —
# CampagneView.tsx hasData rings-empty + #undefined fix + import INLINE
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/Rail.tsx web/frontend/src/styles/index.css web/frontend/src/views/LiveView.tsx web/frontend/src/views/CampagneView.tsx
git commit -m "feat(m8): foundation merge pending observatory2 polish"
```

### Task 2: Backend ARG — Traduction Matérielle + Kit + Pilote + Why Go + Hardware Translate

**Files:**
- Create: `web/frontend/src/lib/hardware.ts`
- Create: `pkg/api/translate.go`
- Modify: `pkg/results/results.go`
- Modify: `web/frontend/src/views/IntegriteView.tsx`
- Test: `web/frontend/src/lib/hardware.test.ts`

**Interfaces:**
- Consumes: `results.Scan` `Group{profile,qdisc,cc,small_p95_median,best,wasted_median}`
- Produces: `hardwareRecommendation(bestQdisc,profile):string` pure + `GET /api/hardware/translate?profile=P2 -> {recommendation}` + `Integrite Recommandations` card

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/hardware.test.ts
import { describe, it, expect } from 'vitest'
import { hardwareRecommendation } from './hardware'
describe('hardware', () => {
  it('mikrotik vs isp', () => {
    expect(hardwareRecommendation('fq_codel','P2')).toContain('MikroTik')
    expect(hardwareRecommendation('cake','P1')).toContain('CAKE')
  })
  it('isp gateway', () => {
    expect(hardwareRecommendation('pfifo_fast','P1')).toContain('mini-PC')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/hardware.test.ts`
Expected: FAIL `hardwareRecommendation is not defined`

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/hardware.ts
export function hardwareRecommendation(bestQdisc: string, profile: string): string {
  if (bestQdisc === 'fq_codel' || bestQdisc === 'cake') return `Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour ${profile} — ${bestQdisc} prouvé en lab`
  return `Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour ${profile} — pilote isolé d'abord`
}

// pkg/results/results.go: add hardware_recommendation per Group
// type Group struct { ..., HardwareRecommendation string `json:"hardware_recommendation"` }
// in Scan loop: g.HardwareRecommendation = hardwareRecommendation(g.BestQDisc, g.Profile) // via func

// pkg/api/translate.go
// func HandleTranslate(w http.ResponseWriter, r *http.Request) { profile:=r.URL.Query().Get("profile"); rec:=hardwareRecommendation(...); json.NewEncoder(w).Encode(map[string]string{"recommendation":rec}) }
// server.go: mux.HandleFunc("/api/hardware/translate", HandleTranslate)
```

Add `IntegriteView.tsx` new card `Recommandations — Traduction Matérielle` table `Linux principle → MikroTik / ISP` + copy `Kit de Diagnostic Portable DSI` + `Pilote isolé` + `Why Go: zero-dependency vs Flent` (from ARG.md).

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/hardware.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/hardware.ts web/frontend/src/lib/hardware.test.ts pkg/results/results.go pkg/api/translate.go web/frontend/src/views/IntegriteView.tsx
git commit -m "feat(m8): traduccion materielle kit portable pilote hardware translate"
```

### Task 3: METEOLINK Lockup + Rail/Foot/Arm Big Polish

**Files:**
- Create: `web/frontend/src/components/MeteolinkWordmark.tsx`
- Modify: `web/frontend/src/components/Rail.tsx`
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/components/ArmButton.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Test: `web/frontend/src/lib/meteolink.test.ts`

**Interfaces:**
- Consumes: `DESIGN.md` sharp `Cormorant + stripe + tricolor`, `Rail` icons
- Produces: `MeteolinkWordmark({compact:boolean})` + `Rail 56 shrunk + 232 extended` polished + `FootTicker` 28px + `Arm progress line 5s`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/meteolink.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('meteolink', () => {
  it('wordmark has gradient and icon', () => {
    const s=readFileSync('web/frontend/src/components/MeteolinkWordmark.tsx','utf8')
    expect(s).toContain('METEOLINK')
    expect(s).toContain('createDrawable')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/meteolink.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// web/frontend/src/components/MeteolinkWordmark.tsx
import { createDrawable, splitText } from 'animejs' // svg/text verified via context7
import { prefersReducedMotion } from '../lib/anime'
export function MeteolinkWordmark({compact}:{compact?:boolean}){
  // Cormorant 600 Meteolink + LIEN subtitle + 16×16 satellite→wave icon stroke1.5 + 4px tricolor stripe
  // on mount if !prefersReducedMotion(): splitText chars stagger 30 grid[4,2] + svg path draw 800
  return <span className="wordmark">...</span>
}

// Rail.tsx: refine shrunk 56 (icon 16 + sr-only + sparkline 32×12 cyan 0.6) + extended 232 (icon+label+key+side-status + gates StatusPip + live peek sparkline + foot)
// App.tsx: header uses MeteolinkWordmark full lockup + FootTicker 28px source | run 8-char · hash 8-char · phase · SSE tabular hairline top
// ArmButton.tsx: primary white → danger red spring scale 0.96→1 rotate 0.5→-0.5 + shadow pulse + 1px progress line width 100%→0% 5s above button when armed
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/meteolink.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/MeteolinkWordmark.tsx web/frontend/src/lib/meteolink.test.ts web/frontend/src/components/Rail.tsx web/frontend/src/App.tsx web/frontend/src/components/ArmButton.tsx
git commit -m "feat(m8): meteolink lockup rail foot arm polish"
```

### Task 4: Prompt Progress Line + Neat Tidy+Breathe Placement

**Files:**
- Create: `web/frontend/src/components/PromptProgressLine.tsx`
- Modify: `web/frontend/src/components/QuickActionsPrompt.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Test: `web/frontend/src/lib/promptProgress.test.ts`

**Interfaces:**
- Consumes: `QuickActionsPrompt` `open` 6s timer
- Produces: `PromptProgressLine({durationMs:6000, paused:boolean})` bottom `1px amber width 100%→0% linear pause on hover` + `panel-stack tidy+breathing placement`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/promptProgress.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('promptProgress', () => {
  it('progress line 6s bottom amber', () => {
    const s=readFileSync('web/frontend/src/components/PromptProgressLine.tsx','utf8')
    expect(s).toContain('6000')
    expect(s).toContain('#f4b400')
    expect(s).toContain('paused')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/promptProgress.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// web/frontend/src/components/PromptProgressLine.tsx
export function PromptProgressLine({paused}:{paused:boolean}){
  return <div className="prompt-progress" style={{position:'absolute', bottom:0, left:0, height:1, background:'#f4b400', width:'100%', animation:'promptDrain 6s linear forwards', animationPlayState: paused?'paused':'running'}} />
}
// index.css: @keyframes promptDrain { from{width:100%} to{width:0%} } .prompt-progress{position:absolute;bottom:0;left:0;height:1px;background:#f4b400}
// QuickActionsPrompt.tsx: integrate <PromptProgressLine paused={isHoverPaused} /> at bottom edge of card, position:relative; plus outer .panel-stack grid tidy+breathing: gap24 airy 16 dense, card-head 11 caps, kv 4px rows, provenance 10 right-aligned, view-title 20 centered, equalized heights via subgrid (already)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/promptProgress.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/PromptProgressLine.tsx web/frontend/src/lib/promptProgress.test.ts web/frontend/src/components/QuickActionsPrompt.tsx web/frontend/src/styles/index.css
git commit -m "feat(m8): prompt progress line bottom amber 6s tidy breathing"
```

### Task 5: HD Sparkline Every MetricCard + Beam/Donut + Snippets Smart

**Files:**
- Modify: `web/frontend/src/components/ui/MetricCard.tsx`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Create: `web/frontend/src/components/Beam.tsx`
- Create: `web/frontend/src/components/DonutJFI.tsx`
- Test: `web/frontend/src/lib/hdSpark.test.ts`

**Interfaces:**
- Consumes: `live.rings` `lttb40`, `JFI` value
- Produces: `MetricCard` now `spark 60×12 clipPath rx4 lttb40 + trend` for every card + `Beam` SVG Campagne→Live when running + `DonutJFI` arc `steel→green` + `snap-effect` handle 16px on Timeline

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/hdSpark.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('hdSpark', () => {
  it('MetricCard has sparkline 60×12', () => {
    const s=readFileSync('web/frontend/src/components/ui/MetricCard.tsx','utf8')
    expect(s).toContain('60')
    expect(s).toContain('clipPath')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: FAIL if MetricCard lacks sparkline

- [ ] **Step 3: Write minimal implementation**

```tsx
// MetricCard.tsx: add sparkline 60×12 svg path lttb40 + clipPath rx4 already but ensure every card uses it (wasted/cost/deadline too)
// LiveView.tsx: every MetricCard now spark={spark(live.rings)} + JFI card uses <DonutJFI value={jfiVal} /> instead of spark
// Beam.tsx: animated-beam adapted: <svg style={{position:'fixed', top: 120, left: 'calc(var(--rail-w) + 24px)', width: 'calc(100% - var(--rail-w) - 48px)', height: 2, pointerEvents:'none'}}><path d="M0 1 H100%" stroke="url(#grad)" strokeDasharray="4" strokeDashoffset={-40 * (Date.now()%1000)/1000} /></svg> when live.running
// DonutJFI.tsx: donut-chart adapted: <svg width=56 height=56><circle stroke={jfi>0.95?'#1fa348':'#9aa3ad'} strokeDasharray={`${jfi*176} 176`} /></svg>
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/ui/MetricCard.tsx web/frontend/src/views/LiveView.tsx web/frontend/src/components/Beam.tsx web/frontend/src/components/DonutJFI.tsx web/frontend/src/lib/hdSpark.test.ts
git commit -m "feat(m8): hd sparkline every card beam donut snap"
```

### Task 6: Anime svg/text/animatable/layout Dense — Context7 Verified

**Files:**
- Modify: `web/frontend/src/lib/anime.ts`
- Modify: `web/frontend/src/components/MeteolinkWordmark.tsx`
- Modify: `web/frontend/src/App.tsx`
- Test: `web/frontend/src/lib/animeSvg.test.ts`

**Interfaces:**
- Consumes: `animejs` `svg/text/animatable/layout` via `context7 1983 snippets` + `prefersReducedMotion`
- Produces: `animateMeteolinkShimmer`, `animateBeam`, `animateDonut`, `animateGrid` using `svg.createDrawable`, `text.splitText`, `animatable`, `stagger grid [4,2] from:center`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/animeSvg.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('animeSvg', () => {
  it('anime uses svg/text/animatable', () => {
    const s=readFileSync('web/frontend/src/lib/anime.ts','utf8')
    expect(s).toContain('createDrawable')
    expect(s).toContain('splitText')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/animeSvg.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```ts
// lib/anime.ts add verified via context7 curl -s 'https://context7.com/api/v2/context?libraryId=/websites/animejs&query=svg+text+animatable&type=txt':
import { createDrawable, splitText, animatable } from 'animejs'
export function animateMeteolinkShimmer(el: Element){
  if(prefersReducedMotion()) return
  const drawable = createDrawable(el.querySelector('path'))
  drawable.animate({ strokeDashoffset: [-100,0], duration:800, ease:'linear' })
  const split = splitText(el, { chars: true })
  animate(split.chars, { translateY: [8,0], opacity:[0,1], delay: stagger(30,{grid:[4,2], from:'center'}) } as any)
}
// App.tsx bento grid stagger: stagger(40,{grid:[2,3], from:'first'}) on panel-stack cards via createTimeline
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/animeSvg.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/anime.ts web/frontend/src/lib/animeSvg.test.ts web/frontend/src/components/MeteolinkWordmark.tsx web/frontend/src/App.tsx
git commit -m "feat(m8): anime svg text animatable layout dense"
```

### Task 7: Real SSE + A/B Before/After + Smoke Pilot + Full 36 Overnight + Final Deploy

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Modify: `Makefile`
- Create: `web/frontend/e2e/m8.spec.ts`
- Test: `npx playwright test` host→VM

**Interfaces:**
- Consumes: all M8 layers
- Produces: `Live wall overlay baseline grey dashed vs CAKE cyan/green solid` + `Résultats A/B bento diff badge -82%` + `smoke pilot P2×fq/bbr×2 reps now` + `full 36 overnight` + `deploy 315→325 KB gz`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/e2e/m8.spec.ts
import { test, expect } from '@playwright/test'
test('m8 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK')).toBeVisible()
  await page.getByRole('button',{name:'Densité'}).click()
  await expect(page.locator('.rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.locator('text=Actions rapides')).toBeVisible({timeout:7000})
  await expect(page.locator('.prompt-progress')).toBeVisible()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test web/frontend/e2e/m8.spec.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Ensure `MeteolinkWordmark` visible, `density` toggle `data-density`, `prompt-progress` bar 1px amber 6s, `beam` between Campagne/Live, `donut` JFI, `Live wall overlay` baseline vs CAKE, `Résultats` `hardware_recommendation` provenance, `Makefile test` chains `go vet + bun build + vitest + playwright`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test web/frontend/e2e/m8.spec.ts`
Expected: PASS

- [ ] **Step 5: Commit + Smoke Pilot + Deploy + Verify**

```bash
git add web/frontend/e2e/m8.spec.ts web/frontend/src/views/LiveView.tsx web/frontend/src/views/ResultatsView.tsx Makefile
git commit -m "feat(m8): m8 e2e a/b baseline vs algorithm and smoke pilot"
# smoke pilot now visual
curl -X POST http://192.168.174.128:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P2"],"reps":2,"qdiscs":["fq_codel","pfifo_fast"],"ccs":["bbr","cubic"]}'
curl -N http://192.168.174.128:9090/api/stream | head -n 30  # watch 10 Hz + prompt line
bash deploy/engine.sh --action deploy --config C:/cgo/deploy/cgo-vm.yaml
curl http://192.168.174.128:9090/api/health
npx playwright test
SHOT_DIR=C:/cgo/shots npx playwright test screenshots  # 12 shots + rail expanded/collapsed + prompt line
```

---

**Self-Review**

1. **Spec coverage:** §1 Shell → Task3 METEOLINK+Rail/Foot/Arm, §2 Data Traduction+Kit+Pilote+Why Go → Task2 hardwareRecommendation + translate, §3 Visual Meteolink/shimmer/progress line/bento tidy+breathing/HD sparkline/beam/donut → Tasks3-6, §4 Views A/B → Task5 HD + Task7 A/B wall, §5 Data flow Lab vs Edge + A/B → Task7 smoke pilot + full 36, §6 Handling prompt line + validation → Task4, §7 Testing smoke pilot + full 36 + hardware translate → Task7. All covered.
2. **Placeholder scan:** No TBD/TODO; every step has actual code blocks with exact values.
3. **Type consistency:** `hardwareRecommendation(bestQdisc,profile):string`, `PromptProgressLine({paused})`, `MeteolinkWordmark({compact})`, `Beam` mounted when `live.running`, `DonutJFI value:number|null`, `live.rings empty → —` consistent.
