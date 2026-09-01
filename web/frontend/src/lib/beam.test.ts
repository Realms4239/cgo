import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/chart-grammar/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any).toString("utf8")}catch{}
  return readFileSync(p as any).toString("utf8")
}
describe('beam/donut + clean triple strip',()=>{
  it('Beam functional — gradient + dash animation, guarded by running',()=>{
    const s=read('web/frontend/src/components/Beam.tsx')
    expect(s).toContain('linearGradient')
    expect(s).toContain('strokeDashoffset')
    expect(s).toContain('live?.running')
  })
  it('Beam hover-sync — hovered prop dims/brightens, animation guarded prefers-reduced-motion',()=>{
    const s=read('web/frontend/src/components/Beam.tsx')
    expect(s).toContain('hovered')
    const css=read('web/frontend/src/styles/index.css')
    expect(css).toContain('@media (prefers-reduced-motion: no-preference)')
    expect(css.split('beamDash').length).toBeGreaterThanOrEqual(2)
  })
  it('DonutJFI honest — when idle',()=>{
    const s=read('web/frontend/src/components/DonutJFI.tsx')
    expect(s).toContain('—')
  })
  it('clean triple strip — slider handle chrome + brush toolbox removed, semantic visualMap kept',()=>{
    const g=read('web/frontend/src/lib/chartGrammar.ts')
    expect(g).not.toContain("type: 'slider'")
    expect(g).toContain("type: 'inside'")
    expect(g).not.toContain("type: 'piecewise'")
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).not.toContain('toolbox:')
  })
})

describe('chart grammar parameterized — GoAccess/AA craft (plans §3)',()=>{
  it('grammar exposes line + bar + area + scatter series craft from one factory',()=>{
    const g=read('web/frontend/src/lib/chartGrammar.ts')
    expect(g).toContain('export function lineSeries')
    expect(g).toContain('export function barSeries')
    expect(g).toContain('export function scatterSeries')
    // markPoint à l'unité, pas de ms codé en dur
    expect(g).toContain('unit')
    expect(g).not.toContain("formatter: 'max {c} ms'")
  })
  it('charge markArea lives in grammar — single, phase-driven (no quartile estimate)',()=>{
    const g=read('web/frontend/src/lib/chartGrammar.ts')
    expect(g).toContain('chargeMarkArea')
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).not.toContain('Math.floor(live.rtt95.length * 0.25)')
    // appliqué une fois par graphique, pas par série
    expect((lv.match(/markArea: ma/g)||[]).length).toBeLessThanOrEqual(3)
  })
  it('PanelChooser chart craft is line|bar|area and dispatches consumable tri event',()=>{
    const p=read('web/frontend/src/components/PanelChooser.tsx')
    expect(p).toContain("'line'")
    expect(p).toContain("'bar'")
    expect(p).toContain("'area'")
    expect(p).toContain('data-chart-craft')
  })
  it('LiveView consumes tri-toggle — metric visibility, chart craft, source overlay',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).toContain('panel-chooser-tri')
    expect(lv).toContain('craft')
    expect(lv).toContain('source')
    expect(lv).toContain("data-wall-cards")
  })
  it('Résultats scatter goes through the grammar — no bespoke toolbox',()=>{
    const r=read('web/frontend/src/views/ResultatsView.tsx')
    expect(r).toContain('baseOption')
    expect(r).toContain('scatterSeries')
    expect(r).not.toContain('toolbox')
  })
  it('Timeline 48 — bandes figées, curseur sans reconstruction',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).toContain('<Timeline')
    // borne figée côté appelant : la fenêtre ne grandit pas avec now
    expect(lv).toContain('tRecup + 30000')
    const t=read('web/frontend/src/components/Timeline.tsx')
    // curseur de progression par interval, jamais de rebuild svg
    expect(t).toContain('setInterval')
    expect(t).not.toContain('brushX')
    expect(t).not.toContain('Math.max(tRecup, now)')
  })
  it('Wall provenance hash8 = sha8 from /api/integrity, run-id only as fallback',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).toContain('j?.hash8')
  })
})

describe('Couche B — primitives + grammaire',()=>{
  it('primitives exist and speak tokens, not hex',()=>{
    for(const f of ['Card','CardHead','Stat','Pill','EmptyChart']){
      const s=read(`web/frontend/src/components/ui/${f}.tsx`)
      expect(s).toContain('export function')
      expect(s).toContain('var(--')
    }
  })
  it('grammar — watermark only when !idle + craft palette exported',()=>{
    const g=read('web/frontend/src/lib/chartGrammar.ts')
    expect(g).toContain('export const CRAFT')
    expect(g).toContain('idle')
    expect(g).not.toContain("graphic: [\n      // 28px")
  })
  it('Wall — no internal spec labels in UI, hero head via CardHead',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).not.toContain('hero 300px')
    expect(lv).toContain('CardHead')
  })
})
