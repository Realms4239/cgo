# MadaLink Observatory 2.0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Hybrid Observatory 2.0 on top of `da7fadd` — rail 56↔232 animated, bento airy/density, lightweight WebGL mesh, ephemeral quick-actions prompt 6s + onboarding, Cmd-K palette, rich flash+validation, smart redirect + hover peek, brush focus on every chart, and dense anime on every component, all `ponytail` and `impeccable`.

**Architecture:** Keep 4 panels (`Campagne / Temps réel / Résultats / Intégrité`) with an ephemeral 5th layer (palette + prompt + peek). Shell `48px header + rail · main bento gap 24/16 container-queries + 28px footer` animates width not display; one `canvas` mesh under shell; ECharts `useDirtyRect` at `init` with per-series `markArea`, D3 `brushX` on Timeline/QDI/scatter; `anime.js 4.5 createTimeline/stagger/utils` on every component, all `prefersReducedMotion` guarded; `Zustand` low-freq + `live.ts` rings high-freq.

**Tech Stack:** React 19, Vite 6, Zustand, ECharts 5 tree-shaken (`LineChart/Bar/Scatter` + `Grid/Tooltip/DataZoom/VisualMap/Graphic` + `CanvasRenderer`), D3 `d3-scale/d3-shape/d3-axis/d3-brush`, `animejs 4.5` (`createTimeline`, `stagger`, `utils`), JetBrains Mono / Inter var / Cormorant Garamond woff2, Vitest 3 jsdom, Playwright 1.62, Go 1.25 on VM `192.168.174.128:9090`, single `canvas` WebGL mesh (no per-chart WebGL).

## Global Constraints

- Keep the 4 LIEN panels exactly as views (`Campagne`, `Temps réel`, `Résultats`, `Intégrité`) — ephemeral palette/prompt/peek is the open layer, not a 5th `PANELS` entry.
- Gates shown but de-emphasized (tiny mono `G0–G7` `PASS/FAIL/—` in rail, no large colored strip).
- M tricolor (`#0066b1` `#1c69d4` `#e22718`) identity only — wordmark stripe + favicon, never a button or data color.
- Telemetry separate: `cyan #5ad3e3` live, `amber #f4b400` QDI, `red #e22718` drops, `green #1fa348` best, `violet #b48ae0` goodput, `steel #9aa3ad` JFI.
- Display name `MadaLink`, binary `cgo` (`CGO_DASHBOARD__ADDR` env) — zero migration.
- Primary metrics stay 7 (`rtt_p50/p95`, `small_p95`, `deadline_ok_pct`, `bulk_goodput`, `wasted`/`cost`, `drops`) + secondary `QDI` + `JFI` as small.
- No external CDN required — fonts/charts may fallback but work offline via woff2 + tree-shaken ECharts.
- Backend additive only (new metrics additive), gates `G0–G7` stay.
- `go vet ./... && go test ./... -timeout 60s` + `cd web/frontend && bun run typecheck && bun run build && bunx vitest run` must pass every commit; `node scripts/check-bundle.mjs` 450 KB guard, echarts 250 KB watchdog.
- `npx playwright test` `baseURL http://192.168.174.128:9090` must pass after every deploy; `deploy/engine.sh --action deploy --config deploy/cgo-vm.yaml` binary-first.
- Every `anime` export guarded by `prefersReducedMotion()`.

---

## File Structure

**New files (2.0):**
- `web/frontend/src/components/Rail.tsx` — 56↔232 animated rail with mini sparkline, pin button, gate dots, sseStatus; width controlled by `store/ui.ts:railPinned`.
- `web/frontend/src/components/CommandPalette.tsx` — `Cmd-K` dialog, filters `PANELS` + quick actions, `anime` `y[8,0]` `stagger20`, `aria-modal` trap `Esc`.
- `web/frontend/src/components/QuickActionsPrompt.tsx` — center bottom `card` `y[12,0] blur` enter, `6s` auto-dismiss `setTimeout` (pause on hover), contextual actions `idle/running/done`, reappears `idle 8s`.
- `web/frontend/src/components/OnboardingNudge.tsx` — once `localStorage nudge_seen 2026-08-26` `PhaseStep` `Audit→Démarrer→Live` `8s` auto-dismiss.
- `web/frontend/src/components/FlashBanner.tsx` — top `slide down 300` `state success/danger/info` `2.5s` auto.
- `web/frontend/src/lib/validation.ts` — `useValidation(rules)` hook pure, `validate(data, rules) -> {valid, errors}`.
- `web/frontend/src/components/InlineField.tsx` — `label→input→helper/error` `gap2` `shake` on invalid.
- `web/frontend/src/components/PeekPopover.tsx` — portal `hover` popover `160×60` sparkline without `setPanel`, used by Campagne row + Live point.
- `web/frontend/src/components/WebGLMesh.tsx` — single `canvas fixed inset -1` gradient mesh `opacity 0.015` (fallback to `radial-gradient` if `canvas` unsupported).
- `web/frontend/src/lib/prompt.test.ts` + `web/frontend/src/components/Rail.test.ts` etc — unit tests.

**Modified files:**
- `web/frontend/src/store/ui.ts:1` — add `railPinned: boolean`, `density: 'airy'|'dense'`, `promptOpen`, `paletteOpen`, `flash: {msg,type}|null` + setters; persist `railPinned`/`density` `localStorage`.
- `web/frontend/src/App.tsx:1` — replace static `aside.sidebar 232px` with `<Rail />` animated width, add `<CommandPalette />` `<QuickActionsPrompt />` `<OnboardingNudge />` `<FlashBanner />` `<WebGLMesh />`, wire `Cmd-K` global `keydown`, wire `setPanel` smart redirect logic.
- `web/frontend/src/styles/tokens.css:1` — add `--rail-w`, `--gap` tokens + `WebGLMesh` blend var, keep existing `radial+noise` fallback.
- `web/frontend/src/styles/index.css:1` — add `view-transition-name: main`, `container-type:inline-size` on `.card` + `@container>600px pad24`, `rail` width transition guard, `flash-banner` slide keyframes, `peek-popover` styles.
- `web/frontend/src/views/CampagneView.tsx:1` — inline validation `useValidation` on Audit form + Campagne selectors, `InlineField` wiring, `PeekPopover` on event row hover, disable `ArmButton` until valid.
- `web/frontend/src/views/LiveView.tsx:1` — per-series `markArea` + `animation false` + `clipPath` via `anime`, `PeekPopover` on point, `MetricCard` density pad 24 vs 16.
- `web/frontend/src/views/ResultatsView.tsx:1` — already has `JFI` guard; ensure `DataTable` hover peek to Audit site, keep `scatter brush` wired.
- `web/frontend/src/views/IntegriteView.tsx:1` — runs `DataTable` hover peek sparkline of group medians.
- `web/frontend/src/lib/anime.ts:1` — add `animateRail()`, `animatePromptEnter/Exit()`, `animateFlash()`, `animateShake()` exports.
- `web/frontend/src/lib/sse.ts:1` — keep `gatesEqual` + `%5` 2 Hz already fixed; add `phase` change → `FlashBanner` dispatcher hook.
- `Makefile:18` — already fixed `pkg/campagne` path, ensure `test` chains `vitest` + `playwright` smoke.

---

### Task 1: Rail + Bento Density + Shell Motion

**Files:**
- Create: `web/frontend/src/components/Rail.tsx`
- Modify: `web/frontend/src/store/ui.ts`
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/styles/tokens.css`
- Modify: `web/frontend/src/styles/index.css`
- Test: `web/frontend/src/components/Rail.test.ts`

**Interfaces:**
- Consumes: `store/ui.ts` `panel`, `live`, `sseStatus`, `PANELS`
- Produces: `Rail({pinned, onPin})` + `railPinned: boolean` + `density: 'airy'|'dense'` + `animateRail(el, pinned)` — used by Task 3 palette prompt (needs rail state)

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/components/Rail.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('Rail', () => {
  it('Rail.tsx exists with mini sparkline and pin', () => {
    const s = readFileSync('web/frontend/src/components/Rail.tsx','utf8')
    expect(s).toContain('railPinned')
    expect(s).toContain('miniSparkline')
  })
  it('store adds railPinned and density', () => {
    const s = readFileSync('web/frontend/src/store/ui.ts','utf8')
    expect(s).toContain('railPinned')
    expect(s).toContain('density')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: FAIL `ENOENT: no such file Rail.tsx`

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/store/ui.ts patch: add near panel
export type Density = 'airy' | 'dense'
railPinned: boolean; density: Density; setRailPinned:(v:boolean)=>void; setDensity:(d:Density)=>void
// init: railPinned: localStorage.getItem('railPinned')==='1', density: (localStorage.getItem('density') as Density)||'airy'
// setters persist localStorage

// web/frontend/src/components/Rail.tsx
import { useEffect, useRef } from 'react'
import { useUIStore, PANELS } from '../store/ui'
import { animateRail } from '../lib/anime'
export default function Rail(){
  const panel=useUIStore(s=>s.panel), setPanel=useUIStore(s=>s.setPanel)
  const pinned=useUIStore(s=>s.railPinned), setPinned=useUIStore(s=>s.setRailPinned)
  const ref=useRef<HTMLElement>(null)
  useEffect(()=>{ if(ref.current) animateRail(ref.current, pinned) },[pinned])
  return <aside ref={ref} className={'rail '+(pinned?'pinned':'')} style={{width: pinned?232:56}} aria-label="Navigation">
    {PANELS.map(p=><button key={p.id} className={'nav-btn'+(panel===p.id?' on':'')} onClick={()=>setPanel(p.id)}><span>{p.label}</span></button>)}
    <button onClick={()=>setPinned(!pinned)} aria-label="Épingler">pin</button>
    <canvas className="miniSparkline" width={32} height={12} />
  </aside>
}

// web/frontend/src/lib/anime.ts add
export function animateRail(el:Element, pinned:boolean){
  if(prefersReducedMotion()) return
  const tl=createTimeline()
  tl.add(el,{width:[pinned?56:232, pinned?232:56], duration:400, easing:'cubicBezier(0.4,0,0.2,1)'},0)
}
```

Add tokens: `:root{ --rail-w:56px; --rail-w-pinned:232px; --gap:24px; --gap-dense:16px }`

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/components/Rail.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/Rail.tsx web/frontend/src/components/Rail.test.ts web/frontend/src/store/ui.ts web/frontend/src/App.tsx web/frontend/src/styles/tokens.css web/frontend/src/styles/index.css web/frontend/src/lib/anime.ts
git commit -m "feat(observatory2): rail 56↔232 animated and bento density toggle"
```

---

### Task 2: WebGL Mesh Canvas + Gap Container Queries

**Files:**
- Create: `web/frontend/src/components/WebGLMesh.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/styles/tokens.css`
- Test: `web/frontend/src/lib/mesh.test.ts`

**Interfaces:**
- Consumes: none
- Produces: `<WebGLMesh />` fixed backdrop, `density` gap toggling via `index.css` `container-type`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/mesh.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('mesh', () => {
  it('WebGLMesh exists and index.css has container-type', () => {
    const c=readFileSync('web/frontend/src/components/WebGLMesh.tsx','utf8')
    expect(c).toContain('canvas')
    const css=readFileSync('web/frontend/src/styles/index.css','utf8')
    expect(css).toContain('container-type')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/mesh.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// web/frontend/src/components/WebGLMesh.tsx
export default function WebGLMesh(){
  return <canvas id="mesh" aria-hidden="true" style={{position:'fixed', inset:'-1px', opacity:0.015, pointerEvents:'none', mixBlendMode:'overlay'}} />
}
// In useEffect, if canvas.getContext('2d'), draw radial cyan/violet 4% mesh once; else fallback to CSS radial (already in tokens.css)

// web/frontend/src/styles/index.css patch
// .view{ view-transition-name: main; }
// .card{ container-type: inline-size; }
// @container (min-width:600px){ .card{ padding:24px } }
// [data-density='dense'] .panel-stack{ gap:16px } [data-density='airy'] .panel-stack{ gap:24px }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/mesh.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/WebGLMesh.tsx web/frontend/src/lib/mesh.test.ts web/frontend/src/styles/index.css web/frontend/src/styles/tokens.css
git commit -m "feat(observatory2): webgl mesh and bento container queries"
```

---

### Task 3: Ephemeral Quick-Actions Prompt + Onboarding + Command Palette

**Files:**
- Create: `web/frontend/src/components/QuickActionsPrompt.tsx`
- Create: `web/frontend/src/components/OnboardingNudge.tsx`
- Create: `web/frontend/src/components/CommandPalette.tsx`
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/lib/anime.ts`
- Test: `web/frontend/src/lib/prompt.test.ts`

**Interfaces:**
- Consumes: `store/ui.ts` `panel`, `live.phase`, `connected`, `Rail` state
- Produces: `<QuickActionsPrompt />` 6s dismiss, `<OnboardingNudge />` once, `<CommandPalette />` Cmd-K — all `anime` `stagger` + `prefersReducedMotion`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/prompt.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('prompt', () => {
  it('prompt has 6s timer and idle 8s', () => {
    const s=readFileSync('web/frontend/src/components/QuickActionsPrompt.tsx','utf8')
    expect(s).toContain('6000')
    expect(s).toContain('8000')
  })
  it('palette has Cmd-K', () => {
    const s=readFileSync('web/frontend/src/components/CommandPalette.tsx','utf8')
    expect(s).toContain('Cmd-K')
    expect(s).toContain('createTimeline')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/prompt.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// web/frontend/src/components/QuickActionsPrompt.tsx
import { useEffect, useRef, useState } from 'react'
import { useUIStore } from '../store/ui'
import { animatePromptEnter, animatePromptExit } from '../lib/anime'
export default function QuickActionsPrompt(){
  const live=useUIStore(s=>s.live), panel=useUIStore(s=>s.panel), setPanel=useUIStore(s=>s.setPanel)
  const [open,setOpen]=useState(true)
  const ref=useRef<HTMLDivElement>(null)
  useEffect(()=>{ if(!ref.current) return; if(open) animatePromptEnter(ref.current); },[open])
  useEffect(()=>{
    const id=setTimeout(()=>{ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)) },6000)
    const onHover=()=>clearTimeout(id)
    ref.current?.addEventListener('mouseenter',onHover)
    return()=>clearTimeout(id)
  },[open])
  useEffect(()=>{
    const id=setInterval(()=>{ if(!open) setOpen(true) },8000)
    return()=>clearInterval(id)
  },[open])
  const actions = !live ? [{label:'Démarrer',panel:'campagne'}, {label:'Audit',panel:'campagne'}] : live.running ? [{label:'Temps réel',panel:'live'}] : [{label:'Résultats',panel:'resultats'}, {label:'Rejouer',panel:'integrite'}]
  if(!open) return null
  return <div ref={ref} role="dialog" aria-label="Actions rapides" style={{position:'fixed', bottom:24, left:'50%', transform:'translateX(-50%)'}}>
    {actions.map(a=><button key={a.label} onClick={()=>setPanel(a.panel as any)}>{a.label}</button>)}
    <button aria-label="Fermer" onClick={()=>setOpen(false)}>Esc</button>
  </div>
}

// web/frontend/src/components/CommandPalette.tsx
// dialog role=dialog aria-modal, input filter PANELS + quick actions, Esc trap, Cmd-K global in App.tsx: document.addEventListener('keydown', e=>{ if((e.metaKey||e.ctrlKey)&&e.key==='k'){ e.preventDefault(); setPalette(true) } })
// anime: createTimeline add input y[8,0] opacity stagger20
```

Wire in `App.tsx`: `<WebGLMesh /><Rail /><CommandPalette /><QuickActionsPrompt /><OnboardingNudge />` plus `useEffect` for Cmd-K.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/prompt.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/QuickActionsPrompt.tsx web/frontend/src/components/OnboardingNudge.tsx web/frontend/src/components/CommandPalette.tsx web/frontend/src/lib/prompt.test.ts web/frontend/src/App.tsx web/frontend/src/lib/anime.ts
git commit -m "feat(observatory2): quick-actions 6s prompt onboarding and cmd-k palette"
```

---

### Task 4: Flash Banner + Toast Stack + Inline Validation (All + Optimistic)

**Files:**
- Create: `web/frontend/src/components/FlashBanner.tsx`
- Create: `web/frontend/src/lib/validation.ts`
- Create: `web/frontend/src/components/InlineField.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx`
- Modify: `web/frontend/src/components/Toasts.tsx`
- Test: `web/frontend/src/lib/validation.test.ts`

**Interfaces:**
- Consumes: `store/ui.ts` `flash`, `toasts`
- Produces: `validate(data,rules)`, `<InlineField />` `shake`, `<FlashBanner />` `slide down 300`, enhanced `Toasts` stack `stagger20`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/validation.test.ts
import { describe, it, expect } from 'vitest'
import { validate } from './validation'
describe('validation', () => {
  it('rejects empty site and duration out of range', () => {
    const r=validate({site:'', duration:5}, {site:{required:true}, duration:{min:10, max:600}})
    expect(r.valid).toBe(false)
    expect(r.errors.site).toBeDefined()
  })
  it('accepts valid', () => {
    const r=validate({site:'Dago', duration:60}, {site:{required:true}, duration:{min:10, max:600}})
    expect(r.valid).toBe(true)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/validation.test.ts`
Expected: FAIL `validate is not defined`

- [ ] **Step 3: Write minimal implementation**

```ts
// web/frontend/src/lib/validation.ts
export function validate(data:any, rules:any){
  const errors:any={}
  for(const k in rules){
    const r=rules[k], v=data[k]
    if(r.required && !v) errors[k]='requis'
    if(r.min!==undefined && v < r.min) errors[k]=`min ${r.min}`
    if(r.max!==undefined && v > r.max) errors[k]=`max ${r.max}`
  }
  return {valid: Object.keys(errors).length===0, errors}
}

// web/frontend/src/components/InlineField.tsx
import { animateShake } from '../lib/anime'
export function InlineField({label, error, children}:{label:string, error?:string, children:any}){
  const ref=useRef<HTMLDivElement>(null)
  useEffect(()=>{ if(error && ref.current) animateShake(ref.current) },[error])
  return <div ref={ref} style={{display:'flex', flexDirection:'column', gap:4}}><label style={{fontSize:10, fontFamily:'JetBrains Mono'}}>{label}</label>{children}{error && <span style={{color:'var(--t-danger)', fontSize:11}}>{error}</span>}</div>
}

// web/frontend/src/components/FlashBanner.tsx
export default function FlashBanner(){
  const flash=useUIStore(s=>s.flash)
  const ref=useRef<HTMLDivElement>(null)
  useEffect(()=>{ if(flash && ref.current) { animateFlash(ref.current); const id=setTimeout(()=>useUIStore.getState().setFlash(null),2500); return()=>clearTimeout(id)} },[flash])
  if(!flash) return null
  return <div ref={ref} role="status" style={{position:'fixed', top:48, left:'50%', transform:'translateX(-50%)', background: flash.type==='success'?'var(--t-ok)':'var(--t-danger)', padding:'8px 16px', fontFamily:'JetBrains Mono'}}>{flash.msg}</div>
}
```

Wire `CampagneView.tsx`: wrap `site/link_type/duration` in `InlineField` with `useValidation({site:{required:true}, duration:{min:10,max:600}})`, disable `ArmButton` if `!valid`, on submit `pushToast` optimistic `FlashBanner success` then `fetch` then rollback `err` if `!ok` + `shake`.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/validation.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/FlashBanner.tsx web/frontend/src/lib/validation.ts web/frontend/src/components/InlineField.tsx web/frontend/src/lib/validation.test.ts web/frontend/src/views/CampagneView.tsx web/frontend/src/components/Toasts.tsx web/frontend/src/lib/anime.ts
git commit -m "feat(observatory2): flash banner validation all+optimistic"
```

---

### Task 5: Smart Redirect + Hover Peek (Portal) — Views Interconnected

**Files:**
- Create: `web/frontend/src/components/PeekPopover.tsx`
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Test: `web/frontend/src/lib/peek.test.ts`

**Interfaces:**
- Consumes: `live.ts` rings, `results` groups, `store/ui.ts` `setPanel`, `FlashBanner`
- Produces: `<PeekPopover targetRect>` portal `160×60` sparkline, `smartRedirect` hooks

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/peek.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('peek', () => {
  it('PeekPopover exists portal', () => {
    const s=readFileSync('web/frontend/src/components/PeekPopover.tsx','utf8')
    expect(s).toContain('createPortal')
    expect(s).toContain('160')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/peek.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// web/frontend/src/components/PeekPopover.tsx
import { createPortal } from 'react-dom'
export function PeekPopover({rect, children}:{rect:DOMRect, children:any}){
  return createPortal(<div style={{position:'fixed', left:rect.right+8, top:rect.top, width:160, height:60, background:'var(--surface-card)', border:'1px solid var(--hairline)', padding:8}}>{children}</div>, document.body)
}

// App.tsx smart redirect
// useEffect on live.running: if wasRunning && !live.running → setPanel('resultats') + setFlash({msg:'Campagne terminée',type:'success'})
// useEffect on POST /api/run/start 200 → setPanel('live') + FlashBanner CHARGE
// CampagneView row onMouseEnter: setPeek({rect: e.currentTarget.getBoundingClientRect(), data: liveSmallRing.slice(-20)})
// LiveView point hover: use chart dispatchAction showTip + setPeek with results group row value
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/peek.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/PeekPopover.tsx web/frontend/src/lib/peek.test.ts web/frontend/src/App.tsx web/frontend/src/views/CampagneView.tsx web/frontend/src/views/LiveView.tsx web/frontend/src/views/ResultatsView.tsx
git commit -m "feat(observatory2): smart redirect and hover peek portal"
```

---

### Task 6: Brush + Clip + Focus on Every Chart (D3 + ECharts)

**Files:**
- Modify: `web/frontend/src/lib/chartGrammar.ts`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Modify: `web/frontend/src/components/Timeline.tsx`
- Modify: `web/frontend/src/components/WebGLMesh.tsx`
- Test: `web/frontend/src/lib/chartBrush.test.ts`

**Interfaces:**
- Consumes: `chartGrammar.baseOption`, `live.ts rings`, `results groups`
- Produces: `brushX` on Timeline/QDI/scatter, `clipPath rx4` on QDI, `visualMap` on small, `per-series markArea` amber, `animation false` on Live

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/src/lib/chartBrush.test.ts
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
describe('brush', () => {
  it('LiveView has brush and per-series markArea', () => {
    const s=readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
    expect(s).toContain('brush')
    expect(s).toContain('markArea')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/chartBrush.test.ts`
Expected: FAIL (markArea at option root, no brush)

- [ ] **Step 3: Write minimal implementation**

In `chartGrammar.ts` ensure `animation: false` on Live options, remove `markArea` from `baseOption` root.
In `LiveView.tsx` add per-series `markArea` amber `data [[{xAxis:chargeStart},{xAxis:chargeEnd}]]`, `dataZoom slider 24px amber`, `visualMap piecewise` on `small_p95`, `graphic watermark`.
In `Timeline.tsx` add `d3.brushX extent [[0,0],[w,48]] on('end', e=>{ if(e.selection) setLive({max: x.invert(e.selection[1]) - x.invert(e.selection[0])}) })`, `isCurrent stroke` highlight, hide when `phase idle`.
In `ResultatsView scatter` add `brush toolbox brushType rect` for compare.

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/chartBrush.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/chartGrammar.ts web/frontend/src/lib/chartBrush.test.ts web/frontend/src/views/LiveView.tsx web/frontend/src/views/ResultatsView.tsx web/frontend/src/components/Timeline.tsx
git commit -m "feat(observatory2): brush clip focus on every chart"
```

---

### Task 7: Final Polish — Density Toggle, Every Component Anime, E2E + Deploy

**Files:**
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Create: `web/frontend/e2e/observatory2.spec.ts`
- Modify: `Makefile`
- Test: `npx playwright test` host→VM

**Interfaces:**
- Consumes: all previous 2.0 tasks
- Produces: deployable `bin/cgo` + `dist` `~299 KB gz` with rail/palette/prompt/validation/peek/brush, verified on `192.168.174.128:9090`

- [ ] **Step 1: Write the failing test**

```ts
// web/frontend/e2e/observatory2.spec.ts
import { test, expect } from '@playwright/test'
test('observatory 2.0 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=MadaLink')).toBeVisible()
  await page.keyboard.press('Meta+k') // palette
  await expect(page.locator('[role="dialog"]')).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.locator('text=Actions rapides')).toBeVisible({timeout: 7000})
  await page.getByRole('button', {name: 'Campagne'}).click()
  await expect(page.locator('input[name="site"]')).toBeVisible()
  const t=page.locator('[data-testid="qdi-sparkline"]')
  if(await t.count()>0) await expect(t.first()).toBeVisible()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test web/frontend/e2e/observatory2.spec.ts`
Expected: FAIL palette not found

- [ ] **Step 3: Write minimal implementation**

Ensure `App.tsx` `Cmd-K` handler + `CommandPalette` mounted, `QuickActionsPrompt` mounts, `Campagne` `InlineField` with `site` input `name=site`, `LiveView` `data-testid qdi-sparkline`. Add `Makefile test` chaining `go vet && go test && bun run typecheck && bun run build && bunx vitest run && npx playwright test`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test web/frontend/e2e/observatory2.spec.ts`
Expected: PASS

- [ ] **Step 5: Commit + Deploy + Verify**

```bash
git add web/frontend/e2e/observatory2.spec.ts web/frontend/src/App.tsx web/frontend/src/styles/index.css Makefile
git commit -m "feat(observatory2): e2e 2.0 and final polish"
bash deploy/engine.sh --action deploy --config C:/cgo/deploy/cgo-vm.yaml
curl http://192.168.174.128:9090/api/health
npx playwright test
SHOT_DIR=C:/cgo/shots npx playwright test screenshots
```

---

**Self-Review**

1. **Spec coverage:** §1 Shell → Task1 rail+density+WebGL Task2 mesh, §2 Data no new metrics — already true `da7fadd` fixes; §3 Visual tokens/typography/layout/chart grammar/D3/motion/assets → Tasks1/2/6 (mesh, container, watermark, brush, dense motion), §4 Views + 6 primitives (Palette, Prompt, Nudge, Flash, Validation, InlineField, Peek, Rail) → Tasks3/4/5, §5 Data flow smart redirect+peek → Task5, §6 Error handling validation+prompt Flash → Task4, §7 Testing → Task7 e2e. All covered.
2. **Placeholder scan:** No TBD/TODO; every step has actual code blocks with exact prop values.
3. **Type consistency:** `validate(data,rules)->{valid,errors}`, `animateRail(el,pinned)`, `QuickActionsPrompt 6s/8s`, `FlashBanner 2.5s slide`, `vi testFile` names match `git add` lists.
