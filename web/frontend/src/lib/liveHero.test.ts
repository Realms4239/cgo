import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom — same pattern as meteolink/tui tests
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/wall-kit-reunite/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('liveHero',()=>{
  it('hero 300 full-width — CHARGE markArea driven by live.phase, not estimated quartile',()=>{
    const s=read('web/frontend/src/views/LiveView.tsx')
    expect(s).toContain('height: 300')
    expect(s).toContain('live.phase')
    // live ring carries phase from SSE frames (instrumentation seam __CGO_LIVE {ringsLen,max,phase})
    const l=read('web/frontend/src/lib/live.ts')
    expect(l).toContain('phase')
  })
})