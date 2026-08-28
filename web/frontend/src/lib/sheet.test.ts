import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom — same pattern as meteolink/tui tests
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/wall-kit-reunite/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('sheet',()=>{
  it('sheet slides over wall',()=>{
    const s=read('web/frontend/src/views/CampagneView.tsx')
    expect(s).toContain('380')
    expect(s).toContain('backdrop')
  })
  it('Sheet cockpit has 3-step helpers',()=>{
    const s=read('web/frontend/src/views/CampagneView.tsx')
    expect(s).toContain('Auditer')
  })
  it('Sheet is a real fixed overlay sliding over the wall',()=>{
    const s=read('web/frontend/src/views/CampagneView.tsx')
    expect(s).toContain("position:'fixed'")
    expect(s).toContain('right:0')
    expect(s).toContain('width:380')
    expect(s).toContain('overflowY')
  })
})
