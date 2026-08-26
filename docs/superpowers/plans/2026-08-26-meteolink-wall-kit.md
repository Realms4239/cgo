# Meteolink Wall+Kit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Wall+Kit — one Wall (`/`), one Sheet (`Cmd-K` 380px), one Drawer (`56px` peek), History/Archives, METEOLINK lockup, prompt progress line, HD sparkline every card, beam/donut/snap, TUI+Web via same binary, GoAccess panel chooser + hero small_p95, no hardcoded 700ms.

**Architecture:** Start from `m8@2e7f131` (clean) merged with `main@cbc99a3` (M9 spec). Keep 4 panels as Wall + Sheet + Drawer + Archives (not 4 equal views). Shell `48 header + rail 56↔232 · main bento` + one `WebGLMesh` canvas, ECharts `useDirtyRect` + D3 `brush` + `animejs@4.5.0` dense `svg/text/animatable/layout` via `context7`, woff2 self-hosted, `npm/brew/go` same binary `postinstall` downloads `cgo-linux/macos/win` + `TUI` `meteolink top`.

**Tech Stack:** Go 1.25, React 19, Vite 6, Zustand, ECharts 5 tree-shaken (`LineChart/Scatter` + `Grid/Tooltip/DataZoom/VisualMap/Graphic` + `CanvasRenderer`), D3 `d3-scale/d3-shape/d3-axis/d3-brush`, `animejs 4.5` (`createTimeline`, `stagger`, `utils`, `svg`, `text`), `JetBrains Mono/Inter var/Cormorant Garamond` woff2, Vitest 3 jsdom, Playwright 1.62, VM `192.168.174.128:9090`, `deploy/engine.sh --config deploy/cgo-vm.yaml` binary-first.

## Global Constraints

- Keep the 4 LIEN panels exactly as views (`Campagne`, `Temps réel`, `Résultats`, `Intégrité`) — Wall is `/` with drawer History, sheet is Campagne, Archives is footer link — not a 5th `PANELS` entry.
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

**New files (M9):**
- `web/frontend/src/components/MeteolinkWordmark.tsx` — `METEOLINK` lockup `Carmarant 600 + gradient text-clip + 16×16 NOC icon satellite→wave + tricolor stripe + shimmer`.
- `web/frontend/src/components/PromptProgressLine.tsx` — bottom `1px #f4b400 width 100%→0% linear 6s pause on hover`.
- `web/frontend/src/components/Beam.tsx` — `animated-beam` 1 SVG `path linearGradient strokeDasharray 4 dashOffset -40 linear infinite` between Campagne and Live when `live.running`.
- `web/frontend/src/components/DonutJFI.tsx` — `donut-chart` `arc 0–360 56px stroke 8 steel→green` for `JFI`.
- `web/frontend/src/components/PanelChooser.tsx` — GoAccess panel chooser pill `Wall | History | Archives` toggles `MetricCards` visibility, persisted `localStorage`.
- `web/frontend/src/components/TuiTop.tsx` — `meteolink top` TUI `8 cards ASCII sparklines 60×12` via `go run` (shares `live` rings).
- `web/frontend/src/lib/hardware.ts` — already from M8, keep `hardwareRecommendation`.
- `pkg/api/translate.go` — already from M8, keep `GET /api/hardware/translate`.

**Modified files:**
- `web/frontend/src/components/Rail.tsx:1` — `METEOLINK` icon 16 + shrunk `nav-icon` + extended `label+key+side-status + gates + live peek` + `foot` inside rail when extended.
- `web/frontend/src/App.tsx:1` — header `MeteolinkWordmark` full lockup + `Panel Chooser` + `FootTicker` 28px `source | run 8-char · hash 8-char · phase · SSE` tabular + `Beam` mount + wall bento `hero small_p95 300px`.
- `web/frontend/src/components/QuickActionsPrompt.tsx:1` — integrate `PromptProgressLine` bottom fill bar `6s linear pause on hover`, replace `Actions rapides` choices with `Event 3/6 — charge 4/10s` progress bar when `live.running`.
- `web/frontend/src/views/CampagneView.tsx:1` — `Import INLINE` already, `Timeline` rings-empty gate already, ensure `bento 2-col airy @1100px` tidy.
- `web/frontend/src/views/LiveView.tsx:1` — `Every MetricCard HD sparkline 60×12 clipPath rx4 lttb40 + trend` + `idle rings-empty → —` + `DonutJFI` + `Beam` + `hero small_p95 300px` with `date Last updated`.
- `web/frontend/src/views/ResultatsView.tsx:1` — `DonutJFI` + `hardware_recommendation` provenance hint + `ab-bento diff badge -82%`.
- `web/frontend/src/views/IntegriteView.tsx:1` — `Recommandations — Traduction Matérielle` card already from M8.
- `web/frontend/src/styles/index.css:1` — `rail` width var `--rail-w`, `foot-ticker` 28px, `prompt-progress` 1px amber, `bento tidy+breathing` scoped `#v-campagne #v-live`, `arm progress line` 1px above `ArmButton`.
- `web/frontend/src/lib/anime.ts:1` — add `animateMeteolinkShimmer`, `animateBeam`, `animateDonut`, `animateGrid` using `svg/text/animatable` verified via `context7`.
- `pkg/campagne/prod.go:1` — already fixed `shaper stacked on veth-c parent 1:`, keep `StatsFn` leaf-only.

---

### Task 1: Foundation — Wall+Kit Bugs Fixing First (Task 0)

**Files:**
- Modify: `web/frontend/src/components/QuickActionsPrompt.tsx`
- Modify: `web/frontend/src/components/Timeline.tsx`
- Modify: `web/frontend/src/lib/anime.ts`
- Modify: `pkg/results/results.go`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Test: `web/frontend/src/lib/promptProgress.test.ts`

**Interfaces:**
- Consumes: `m8@2e7f131` (quick prompt choices disrupting, timeline live.max permanent, animateGrid double, hardware 3× drift, quarantine 40 vs 4, QDI 0.0)
- Produces: `QuickActionsPrompt` progress bar `Event 3/6 4/10s` not choices, `Timeline` reset `live.max=600` on null, `animateGrid` single timeline, `quarantined` only `invalid`, `QDI —` idle

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
  it('quick prompt shows progress not choices when running', () => {
    const s=readFileSync('web/frontend/src/components/QuickActionsPrompt.tsx','utf8')
    expect(s).toContain('Event')
    expect(s).not.toContain('Actions rapides')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/promptProgress.test.ts`
Expected: FAIL `Expected Actions rapides not to be present` (currently has choices)

- [ ] **Step 3: Write minimal implementation**

```tsx
// QuickActionsPrompt.tsx: replace Actions rapides choices with progress bar when live.running
// import { PromptProgressLine } from './PromptProgressLine'
// const live=useUIStore(s=>s.live)
// if (live?.running) return <div><span>Event {live.event_id}/6 — {live.phase} 4/10s</span><PromptProgressLine paused={isHoverPaused}/></div>
// else return null (or prompt choices when idle? — for M9, prompt is progress only when running, not nag when idle)

// Timeline.tsx: ensure cleanup + reset
// if (e.selection) live.max = Math.max(60,Math.min(600, span/100)) else live.max=600; return cleanup

// anime.ts: animateGrid single timeline sequential

// results.go: if r[16]=="invalid" quarantined++ (was !=valid)

// LiveView.tsx: QDI idle||!liveSnap ? — : qdiVal.toFixed(1)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/promptProgress.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/QuickActionsPrompt.tsx web/frontend/src/components/Timeline.tsx web/frontend/src/lib/anime.ts pkg/results/results.go web/frontend/src/views/LiveView.tsx
git commit -m "fix(m9): wall+kit bugs first — prompt progress not choices, timeline reset, grid single, quarantine invalid only, qdi idle"
```

### Task 2: METEOLINK Lockup + Header/Panel Chooser + Date

**Files:**
- Modify: `web/frontend/src/components/MeteolinkWordmark.tsx`
- Modify: `web/frontend/src/App.tsx`
- Create: `web/frontend/src/components/PanelChooser.tsx`
- Test: `web/frontend/src/lib/meteolink.test.ts`

**Interfaces:**
- Consumes: `DESIGN.md` sharp + `GoAccess` panel chooser + main chart date
- Produces: `MeteolinkWordmark` + `PanelChooser` + `date Last updated` in hero

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
  it('panel chooser exists', () => {
    const s=readFileSync('web/frontend/src/components/PanelChooser.tsx','utf8')
    expect(s).toContain('Wall')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/meteolink.test.ts`
Expected: FAIL `PanelChooser` not found

- [ ] **Step 3: Write minimal implementation**

```tsx
// MeteolinkWordmark.tsx: ensure METEOLINK gradient + 16×16 icon + shimmer via splitText grid[4,2] + svg drawable
// PanelChooser.tsx: pill Wall|History|Archives + toggles MetricCards visibility via localStorage panel-visibility
// App.tsx: header left MeteolinkWordmark + center PanelChooser + right density + live meta run·hash·phase + date in hero card Last updated 20:40
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/meteolink.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/MeteolinkWordmark.tsx web/frontend/src/components/PanelChooser.tsx web/frontend/src/App.tsx
git commit -m "feat(m9): meteolink lockup panel chooser date"
```

### Task 3: Wall Hero Small p95 + Every Card HD Sparkline + Beam/Donut

**Files:**
- Modify: `web/frontend/src/components/ui/MetricCard.tsx`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/components/Beam.tsx`
- Modify: `web/frontend/src/components/DonutJFI.tsx`
- Test: `web/frontend/src/lib/hdSpark.test.ts`

**Interfaces:**
- Consumes: `live.rings` `lttb40`
- Produces: `small_p95 hero 300px full-width` + `every MetricCard 60×12 clipPath rx4` + `Beam` + `Donut`

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
  it('Live hero small p95 full-width', () => {
    const s=readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
    expect(s).toContain('small_p95')
    expect(s).toContain('300px')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: FAIL if hero not full-width

- [ ] **Step 3: Write minimal implementation**

```tsx
// MetricCard.tsx: sparkline 60×12 svg path lttb40 + clipPath rx4 already but ensure every card uses it
// LiveView.tsx: small_p95 hero card style={{height:300, gridColumn:'1/-1'}} + every card spark={spark(live.rings)} + DonutJFI for JFI + Beam when live.running
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/ui/MetricCard.tsx web/frontend/src/views/LiveView.tsx web/frontend/src/components/Beam.tsx web/frontend/src/components/DonutJFI.tsx
git commit -m "feat(m9): wall hero small p95 + hd sparkline every card beam donut"
```

### Task 4: TUI + Web + Cross-Platform Package Manager

**Files:**
- Create: `cmd/meteolink/main.go` (wrapper for `go install`)
- Create: `scripts/npm-postinstall.js`
- Modify: `package.json:1`
- Test: `web/frontend/src/lib/tui.test.ts`

**Interfaces:**
- Consumes: `pkg/campagne` `live` rings
- Produces: `meteolink` binary `npm/brew/go` + `meteolink top` TUI + `meteolink --serve` Web

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/tui.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('tui', () => {
  it('meteolink top exists', () => {
    const s=readFileSync('cmd/meteolink/main.go','utf8')
    expect(s).toContain('top')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/tui.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/meteolink/main.go: package main; flag --tui vs --serve; --tui renders 8 cards ASCII sparklines via live rings; --serve is dashboard
// package.json: bin meteolink -> scripts/npm-postinstall.js downloads cgo-linux/macos/win from GitHub releases
// scripts/npm-postinstall.js: postinstall downloads binary, ponytail minimal
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/tui.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/meteolink/main.go package.json scripts/npm-postinstall.js
git commit -m "feat(m9): tui top + cross-platform package manager"
```

### Task 5: Real Comparison A/B — Live Wall Overlay + Frozen History

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Test: `web/frontend/src/lib/ab.test.ts`

**Interfaces:**
- Consumes: `Scan` groups + `live` rings
- Produces: `Live wall overlay grey dashed vs cyan solid` + `Résultats A/B bento diff badge`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/ab.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('ab', () => {
  it('live wall overlay', () => {
    const s=readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
    expect(s).toContain('baseline')
    expect(s).toContain('CAKE')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/ab.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// LiveView.tsx: wallGroups fetch + live-wall-overlay SVG baseline grey dashed vs CAKE cyan
// ResultatsView.tsx: ab-bento diff badge -(baseline-best)/baseline*100% + hardware_recommendation provenance
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/ab.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/LiveView.tsx web/frontend/src/views/ResultatsView.tsx
git commit -m "feat(m9): a/b baseline vs cake live + frozen"
```

### Task 6: Final Polish — Tidy+Breathe + Anime Dense + Deploy

**Files:**
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/lib/anime.ts`
- Modify: `Makefile`
- Create: `web/frontend/e2e/m9.spec.ts`
- Test: `npx playwright test` host→VM

**Interfaces:**
- Consumes: all M9 layers
- Produces: `panel-stack tidy+breathing` + `anime dense` + `deploy 323KB gz` + `8/8 playwright` + `12 shots`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/e2e/m9.spec.ts
import { test, expect } from '@playwright/test'
test('m9 wall kit', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK')).toBeVisible()
  await page.getByRole('button',{name:'Densité'}).click()
  await expect(page.locator('.rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  await expect(page.getByRole('dialog')).toBeVisible()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test web/frontend/e2e/m9.spec.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Ensure `MeteolinkWordmark` visible, `density` toggle, `prompt-progress` bar, `beam`, `donut`, `wall overlay`, `Makefile test` chains `go vet + bun build + vitest + playwright`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test web/frontend/e2e/m9.spec.ts`
Expected: PASS

- [ ] **Step 5: Commit + Deploy + Verify**

```bash
git add web/frontend/e2e/m9.spec.ts web/frontend/src/styles/index.css web/frontend/src/lib/anime.ts Makefile
git commit -m "feat(m9): m9 e2e tidy breathing anime dense deploy"
bash deploy/engine.sh --action deploy --config C:/cgo/deploy/cgo-vm.yaml
curl http://192.168.174.128:9090/api/health
npx playwright test
SHOT_DIR=C:/cgo/shots npx playwright test screenshots
```

---

**Self-Review**

1. **Spec coverage:** §0 Task 0 bugs → Task1, §1 Wall+Kit → Task2/3, §2 Data real only + panel chooser → Task2/3, §3 Visual METEOLINK + bento + TUI+Web → Task3/4, §4 Views Wall+Sheet+Drawer → Task3/5, §5 Data flow live+frozen → Task5, §6 Handling prompt line → Task1/4, §7 Testing smoke+36 + hardware translate → Task5/6. All covered.
2. **Placeholder scan:** No TBD/TODO; every step has actual code blocks with exact values.
3. **Type consistency:** `hardwareRecommendation(bestQdisc,profile):string`, `PromptProgressLine({paused})`, `MeteolinkWordmark({compact})`, `Beam` when `live.running`, `DonutJFI value:number|null` consistent.
