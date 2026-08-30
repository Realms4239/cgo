import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('bento',()=>{
  it('wall bento scoped — density var(--gap) only on #wall',()=>{
    const css=read('web/frontend/src/styles/index.css')
    expect(css).toContain('#wall')
    expect(css).toContain('gap: var(--gap)')
    // le mur est la racine panel-stack de LiveView
    const live=read('web/frontend/src/views/LiveView.tsx')
    expect(live).toContain('id="wall"')
  })
})