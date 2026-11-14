import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('liveHero',()=>{
  it('hero 300 full-width — CHARGE markArea driven by live.phase, not estimated quartile',()=>{
    const s=read('web/frontend/src/views/LiveView.tsx')
    expect(s).toContain('live-hero-card')
    expect(read('web/frontend/src/styles/index.css')).toContain('height: 300px')
    expect(s).toContain('live.phase')
    // l'anneau live porte la phase des frames SSE (couture __CGO_LIVE {ringsLen,max,phase})
    const l=read('web/frontend/src/lib/live.ts')
    expect(l).toContain('phase')
  })
})