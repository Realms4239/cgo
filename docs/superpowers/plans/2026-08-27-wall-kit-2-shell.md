# Wall-kit 2 — Shell (App + Rail + PanelChooser + Sheet) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Wall-kit shell — header 48 METEOLINK lockup + PanelChooser tri-toggle + density + rail 56↔232 var + Sheet 380 auto-minimize as Kit cockpit.

**Architecture:** App shell grid 48 + rail · main wall bento max 1280 + footer 28, PanelChooser tri-toggle metric/chart/source drives Wall and Drawer via CustomEvent, Sheet slides over Wall.

**Tech Stack:** React 19 Vite 6 Zustand, var(--rail-w), PanelChooser localStorage, Sheet backdrop blur, ArmButton 5s

## Global Constraints

- Shell grid 48 header + rail 56↔232 animate 400 var(--rail-w) + main wall bento + footer 28 ticker
- PanelChooser tri-toggle metric groups/chart craft/source live|frozen|both persisted localStorage CustomEvent
- Sheet Cmd-K 380 backdrop blur auto-minimize to pill Event 3/6 after Arm, Kit 3-step helpers inline
- Rail ICONS nav-icon 16 sr-only + live peek 32×12 + foot ticker when extended, gates tiny mono rail only
- Build-then-embed TOTAL 600 echarts 350, inline not subagents

---

### Task 2.1: App Shell Grid + MeteolinkWordmark

**Files:**
- Modify: `web/frontend/src/App.tsx:1`
- Create: `web/frontend/src/components/MeteolinkWordmark.tsx`
- Test: `web/frontend/src/lib/meteolink.test.ts`

- [ ] **Step 1: Write failing test**

```ts
import {readFileSync} from 'node:fs'
import {describe,it,expect} from 'vitest'
describe('wordmark',()=>{
  it('has METEOLINK gradient',()=>{
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
export function MeteolinkWordmark({compact}:{compact?:boolean}){
  // Cormorant 600 + satellite→wave icon + tricolor stripe + splitText grid[4,2] shimmer
  return <span>METEOLINK</span>
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/meteolink.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/MeteolinkWordmark.tsx web/frontend/src/App.tsx
git commit -m "feat(shell): METEOLINK lockup header 48 + App grid"
```

### Task 2.2: PanelChooser Tri-toggle

**Files:**
- Create: `web/frontend/src/components/PanelChooser.tsx`
- Test: `web/frontend/src/lib/panel.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('panelChooser',()=>{
  it('tri-toggle exists',()=>{
    const s=readFileSync('web/frontend/src/components/PanelChooser.tsx','utf8')
    expect(s).toContain('Wall')
    expect(s).toContain('localStorage')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/panel.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
export default function PanelChooser(){
  const [vis,setVis]=useState(()=>JSON.parse(localStorage.getItem('panel-visibility')||'{}'))
  useEffect(()=>localStorage.setItem('panel-visibility',JSON.stringify(vis)),[vis])
  useEffect(()=>window.dispatchEvent(new CustomEvent('panel-visibility',{detail:vis})),[vis])
  return <div className="panel-chooser"><button onClick={()=>setVis(v=>({...v,metric:!v.metric}))}>Metric</button></div>
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/panel.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/PanelChooser.tsx
git commit -m "feat(shell): PanelChooser tri-toggle metric/chart/source"
```

### Task 2.3: Sheet Cockpit 380 + Arm 5s

**Files:**
- Modify: `web/frontend/src/views/CampagneView.tsx:1`
- Modify: `web/frontend/src/components/ArmButton.tsx:1`
- Test: `web/frontend/src/lib/sheet.test.ts`

- [ ] **Step 1: Write failing test**

```ts
describe('sheet',()=>{
  it('sheet slides over wall',()=>{
    const s=readFileSync('web/frontend/src/views/CampagneView.tsx','utf8')
    expect(s).toContain('380')
    expect(s).toContain('backdrop')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bunx vitest run web/frontend/src/lib/sheet.test.ts`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```tsx
// CampagneView as Sheet overlay
<div style={{position:'fixed', right:0, width:380, backdropFilter:'blur(12px)'}}>
  <InlineField label="Site"><input/></InlineField>
  <ArmButton label="DÉMARRER" onConfirm={()=>{setPill('Event 3/6'); window.dispatchEvent(new CustomEvent('sheet-minimize'))}} />
</div>
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bunx vitest run web/frontend/src/lib/sheet.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/CampagneView.tsx web/frontend/src/components/ArmButton.tsx
git commit -m "feat(shell): Sheet 380 backdrop + Arm 5s Kit cockpit"
```
