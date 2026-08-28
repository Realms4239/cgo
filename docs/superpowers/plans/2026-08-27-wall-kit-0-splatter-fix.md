# Wall-kit 0 — Splatter Fix + 3 Instrumentation Seams Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix splattered furniture before new layers — rail jank C→bento breathing B→scroll lock A, honest emptiness, validation, plus 3 instrumentation seams (Go --tags=instrumentation, frontend data-testid, embed trace) so debug-after can prove fixes.

**Architecture:** Layer 0 is debug-before heavy surgery: prove `nothing renders` via `kit/logs` + `playwright clip + trace + __CGO_*` before any new furniture, keeping `go vet + bun typecheck strict` green.

**Tech Stack:** Go 1.25, React 19 Vite 6, Zustand low-freq + live.ts mutable 1800, ECharts useDirtyRect, Vitest jsdom, Playwright 1.62, kit/logs gitignored

## Global Constraints

- Build-then-embed TOTAL gz 600KB watchdog echarts 350KB (decently portable) every commit
- Go `go vet && go test -tags=instrumentation -timeout 60s -race` every commit
- Frontend `bun run typecheck strict noUnusedLocals` every commit
- SSE 10Hz delta Last-Event-ID 2048 backpressure 16→dropped, live.ts rings 1800 phase-driven, useRafLoop 4Hz lttb400
- No auth token, ArmButton 5s progress line, double-confirm ≤5s blur resets
- Display METEOLINK, binary cgo, CGO_DASHBOARD__ADDR env
- Metrics Go pure, browser never derives
- Gates quarantined only invalid (G0/G1/G3/G4/G5), tiny mono PASS/FAIL/— rail only
- DESIGN sharp 0 #070707 hairline #26262a, tokens var(--t-*), bento tidy+breathing, guarded anime prefersReducedMotion
- kit/engine.sh --action build always before deploy, VMware detect-fallback VBox, cloudflared $CLOUDFLARE_TUNNEL_TOKEN not committed
- Inline not subagents, 3 seams for heavy debugging 70%

---

## File Structure

**New files:**
- `kit/logs/.gitkeep` — gitignored logs dir (build.log tc.log sse.log embed.log e2e.log)
- `web/frontend/src/lib/instrumentation.ts` — `export const INSTRUMENT = {live, sse}` for `window.__CGO_*`
- `web/frontend/e2e/wall-splatter.spec.ts` — component-clip surgical specs for rail/bento/scroll

**Modified files:**
- `web/frontend/src/components/Rail.tsx:1` — var(--rail-w) + ICONS + sr-only + foot ticker
- `web/frontend/src/styles/index.css:1` — bento wall scoped #wall @1100px + @900px rail row + calc(100vh-48-28) overflow
- `web/frontend/src/views/IntegriteView.tsx:1` — max-height calc + ordered sections + RDF sha placeholder fix
- `pkg/campagne/campagne.go:216` — already leaf-only, add instrumentation tag log
- `pkg/api/sse.go:48` — Hub delta vs worker throttling guard

---

### Task 0.1: Instrumentation Seams (Go tag + frontend attrs + embed)

**Files:**
- Create: `web/frontend/src/lib/instrumentation.ts`
- Modify: `pkg/campagne/campagne.go:1` (add `//go:build instrumentation` file `pkg/campagne/instrument.go`)
- Modify: `web/frontend/src/lib/live.ts:1`
- Test: `web/frontend/src/lib/instrumentation.test.ts`

**Interfaces:**
- Consumes: `live.ts rings 1800` + `sse.ts Hub`
- Produces: `window.__CGO_LIVE {ringsLen,max,phase} + window.__CGO_SSE {lastID,dropped,ringLen,frameCount} + kit/logs/tc.log`

- [ ] **Step 1: Write failing test**

```ts
// web/frontend/src/lib/instrumentation.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('instrumentation', ()=>{
  it('live exposes __CGO', ()=>{
    const s=readFileSync('web/frontend/src/lib/live.ts','utf8')
    expect(s).toContain('__CGO_LIVE')
  })
  it('sse exposes __CGO_SSE', ()=>{
    const s=readFileSync('web/frontend/src/lib/sse.ts','utf8')
    expect(s).toContain('__CGO_SSE')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/instrumentation.test.ts`
Expected: FAIL `__CGO_LIVE not found`

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/live.ts add
export const live = {rtt50:[],rtt95:[],small:[],goodput:[],max:600}
if(typeof window!=='undefined') (window as any).__CGO_LIVE = live
// web/frontend/src/lib/sse.ts add
let frameCount=0
export const __CGO_SSE={lastID:0,dropped:0,ringLen:0,frameCount}
if(typeof window!=='undefined') (window as any).__CGO_SSE=__CGO_SSE
// pkg/campagne/instrument.go //go:build instrumentation
package campagne
import "log"
func logTC(args ...string){ log.Printf("tc %v",args)}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/instrumentation.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/live.ts web/frontend/src/lib/sse.ts pkg/campagne/instrument.go web/frontend/src/lib/instrumentation.test.ts
git commit -m "feat(kit): instrumentation 3 seams live/sse/tc go tag"
```

### Task 0.2: Rail jank C Fix (var(--rail-w) + animatable)

**Files:**
- Modify: `web/frontend/src/components/Rail.tsx:1`
- Modify: `web/frontend/src/styles/index.css:1`
- Test: `web/frontend/src/components/Rail.test.ts`

**Interfaces:**
- Consumes: `tokens.css --rail-w`
- Produces: `Rail 56↔232 var(--rail-w) animate 400 + ICONS nav-icon 16 + sr-only + live peek 32×12`

- [ ] **Step 1: Write failing test**

```ts
// Rail.test.ts
import {readFileSync} from 'node:fs'
import {describe,it,expect} from 'vitest'
describe('Rail',()=>{
  it('uses css var not inline',()=>{
    const s=readFileSync('web/frontend/src/components/Rail.tsx','utf8')
    expect(s).toContain('--rail-w')
    expect(s).not.toContain('style={{width: pinned?')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: FAIL inline width found

- [ ] **Step 3: Write minimal implementation**

```tsx
// Rail.tsx
<aside ref={ref} className={'rail '+(pinned?'pinned':'')} data-pinned={pinned?'1':'0'} style={{'--rail-w': pinned?232:56} as any}>
// index.css
.rail{width:var(--rail-w); transition:width 400ms var(--ease)}
@media(max-width:900px){.rail{width:100% !important; height:48px; flex-direction:row}}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/Rail.tsx web/frontend/src/styles/index.css
git commit -m "fix(wall): rail var(--rail-w) animatable 400 + mobile row 48"
```

### Task 0.3: Bento breathing B Fix (wall bento scoped)

**Files:**
- Modify: `web/frontend/src/styles/index.css:1`
- Test: `web/frontend/src/lib/bento.test.ts`

- [ ] **Step 1: Write failing test**

```ts
import {readFileSync} from 'node:fs'
import {describe,it,expect} from 'vitest'
describe('bento',()=>{
  it('wall bento scoped',()=>{
    const s=readFileSync('web/frontend/src/styles/index.css','utf8')
    expect(s).toContain('#wall')
    expect(s).toContain('gap: var(--gap)')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/bento.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```css
#wall .panel-stack{gap:var(--gap); max-width:1280; container-type:inline-size}
#wall[data-density="airy"]{--gap:24px} #wall[data-density="dense"]{--gap:16px}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/bento.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/styles/index.css
git commit -m "fix(wall): bento wall scoped airy 24 dense 16"
```

### Task 0.4: Integrité scroll lock A Fix + honest EmptyState

**Files:**
- Modify: `web/frontend/src/views/IntegriteView.tsx:1`
- Modify: `web/frontend/src/views/LiveView.tsx:1` (honest —)
- Test: `web/frontend/e2e/wall-splatter.spec.ts`

- [ ] **Step 1: Write failing test**

```ts
import {test,expect} from '@playwright/test'
test('integrite scroll', async({page})=>{
  await page.goto('/archives')
  const box=await page.locator('main').boundingBox()
  expect(box?.height).toBeGreaterThan(0)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test wall-splatter -v`
Expected: FAIL scroll lock

- [ ] **Step 3: Write minimal implementation**

```tsx
// IntegriteView.tsx wrap panel-stack in <div style={{maxHeight:'calc(100vh - 48px - 28px)', overflowY:'auto'}}>
// LiveView.tsx MetricCard value idle? '—' : value
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test wall-splatter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/IntegriteView.tsx web/frontend/src/views/LiveView.tsx
git commit -m "fix(wall): integrite calc(100vh-48-28) scroll + honest —"
```
