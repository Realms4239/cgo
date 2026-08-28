import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom — same pattern as meteolink/tui tests
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/chart-grammar/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
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
    expect(g).toContain('visualMap')
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
    // markPoint unit-aware, not hardcoded ms
    expect(g).toContain('unit')
    expect(g).not.toContain("formatter: 'max {c} ms'")
  })
  it('charge markArea lives in grammar — single, phase-driven (no quartile estimate)',()=>{
    const g=read('web/frontend/src/lib/chartGrammar.ts')
    expect(g).toContain('chargeMarkArea')
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).not.toContain('Math.floor(live.rtt95.length * 0.25)')
    // applied once per chart, not per series
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
  it('Timeline 48 on the Wall + brush window matches rings 1800',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).toContain('<Timeline')
    const t=read('web/frontend/src/components/Timeline.tsx')
    expect(t).toContain('1800')
    expect(t).not.toContain('Math.min(600')
  })
  it('Wall provenance hash8 = sha8 from /api/integrity, run-id only as fallback',()=>{
    const lv=read('web/frontend/src/views/LiveView.tsx')
    expect(lv).toContain('j?.hash8')
  })
})
