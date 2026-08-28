import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom — same pattern as meteolink/tui tests
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/wall-kit-reunite/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('beam/donut + clean triple strip',()=>{
  it('Beam functional — gradient + dash animation, guarded by running',()=>{
    const s=read('web/frontend/src/components/Beam.tsx')
    expect(s).toContain('linearGradient')
    expect(s).toContain('strokeDashoffset')
    expect(s).toContain('live?.running')
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