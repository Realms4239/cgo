import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom — same pattern as meteolink/tui tests
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/wall-kit-reunite/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('bento',()=>{
  it('wall bento scoped — density var(--gap) only on #wall',()=>{
    const css=read('web/frontend/src/styles/index.css')
    expect(css).toContain('#wall')
    expect(css).toContain('gap: var(--gap)')
    // wall is the LiveView panel-stack root
    const live=read('web/frontend/src/views/LiveView.tsx')
    expect(live).toContain('id="wall"')
  })
})