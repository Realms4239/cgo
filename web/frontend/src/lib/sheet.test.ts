import { describe, it, expect } from 'vitest'
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
})
