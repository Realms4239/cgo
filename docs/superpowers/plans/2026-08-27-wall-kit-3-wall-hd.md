# Wall-kit 3 — Wall HD (MetricCard + Live Hero + Beam/Donut + Clean ECharts) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver hybrid lively Wall — selective HD sparklines + hero 300 phase-driven 1800 + Timeline 48 + functional Beam/Donut + clean triple strip, mix B+C extra panels.

**Architecture:** LiveView 3 charts clean-lively (hero 300 + RTT 180 + goodput 180 + Timeline 48 D3) + 8 MetricCards selective + drawer bar/scatter D3 pareto, all lttb + useDirtyRect + useRafLoop 4Hz.

**Tech Stack:** ECharts 5 tree-shaken useDirtyRect, D3 scale/shape/axis/brush, lttb, anime guarded, var(--t-*)

## Global Constraints

- Selective HD: continuous rtt/small/goodput/QDI spark 60×12 lttb40 clipPath rx4, drops/wasted/cost numeric + trend
- Hero 300 full-width date Last updated pulse + hash, RTT/goodput 180 share charge markArea from live.phase, rings 1800 one event
- Clean triple strip: dataZoom slider/visualMap watermark blur/LinearGradient removed, markArea CHARGE single
- Functional Beam hover-sync when running, DonutJFI — when idle, anime guarded prefersReducedMotion state changes only
- Build-then-embed TOTAL 600 echarts 350

---

### Task 3.1: MetricCard Selective HD

**Files:**
- Modify: `web/frontend/src/components/ui/MetricCard.tsx:1`
- Test: `web/frontend/src/lib/hdSpark.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('hdSpark',()=>{
  it('MetricCard clipPath',()=>{
    const s=readFileSync('web/frontend/src/components/ui/MetricCard.tsx','utf8')
    expect(s).toContain('clipPath')
    expect(s).toContain('60')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
export function MetricCard({label,value,unit,color,spark,trend}){
  const path = spark? lttb(spark,40).map(([x,y])=>`${x},${y}`).join(' ') : ''
  return <div data-metric={label}><svg width={60} height={12}><clipPath id={`c-${label}`}><rect rx={4}/></clipPath><path d={path} stroke={color} clipPath={`url(#c-${label})`}/></svg>{value}{unit}{trend}</div>
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/hdSpark.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/ui/MetricCard.tsx
git commit -m "feat(wall): MetricCard selective HD 60×12 lttb40 clipPath"
```

### Task 3.2: LiveView Hero 300 Phase-driven 1800

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx:1`
- Modify: `web/frontend/src/lib/live.ts:1` rings 600→1800
- Test: `web/frontend/src/lib/liveHero.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('liveHero',()=>{
  it('hero 300 phase-driven',()=>{
    const s=readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
    expect(s).toContain('300px')
    expect(s).toContain('live.phase')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/liveHero.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// live.ts
export const live={rtt50:[],rtt95:[],small:[],goodput:[],max:1800}
// LiveView.tsx
<div style={{height:300, gridColumn:'1/-1'}}><div ref={small.ref} style={{flex:1}}/><span>Last updated {new Date().toLocaleTimeString('fr-FR',{hour:'2-digit',minute:'2-digit'})} · hash {hash8}</span></div>
const cs = live.phase==='charge'? Date.now()-120000 : 0
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/liveHero.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/LiveView.tsx web/frontend/src/lib/live.ts
git commit -m "feat(wall): hero 300 phase-driven rings 1800 + date pulse"
```

### Task 3.3: Beam/Donut + Clean Triple Strip

**Files:**
- Create: `web/frontend/src/components/Beam.tsx`
- Create: `web/frontend/src/components/DonutJFI.tsx`
- Modify: `web/frontend/src/lib/anime.ts:1`
- Test: `web/frontend/src/lib/beam.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('beam',()=>{
  it('beam functional',()=>{
    const s=readFileSync('web/frontend/src/components/Beam.tsx','utf8')
    expect(s).toContain('linearGradient')
    expect(s).toContain('dashOffset')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/beam.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
export function Beam(){ const live=useUIStore(s=>s.live); if(!live?.running) return null; return <svg style={{position:'fixed', height:2}}><path stroke="url(#grad)" strokeDasharray={4} style={{animation:'dash -40 linear infinite'}}/></svg>}
export function DonutJFI({value}:{value:number|null}){ if(value==null) return <span>—</span>; return <svg width={56} height={56}><circle stroke={value>0.95?'#1fa348':'#9aa3ad'} strokeDasharray={`${value*176} 176`}/></svg>}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/beam.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/Beam.tsx web/frontend/src/components/DonutJFI.tsx
git commit -m "feat(wall): Beam hover-sync + DonutJFI — when idle + clean triple strip"
```
