import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom ponytail minimal
import { readFileSync } from 'node:fs'
function read(p:string){
  for(const q of [p, p.replace('web/frontend/',''), `../${p}`, `../../${p}`, `C:/cgo/.worktrees/wall-kit-reunite/${p}`, `C:/cgo/${p}`]) try{ return readFileSync(q as any,'utf8' as any)}catch{}
  return readFileSync(p as any,'utf8' as any)
}
describe('instrumentation', ()=>{
  it('live exposes __CGO_LIVE', ()=>{
    const s=read('web/frontend/src/lib/live.ts')
    expect(s).toContain('__CGO_LIVE')
  })
  it('sse exposes __CGO_SSE', ()=>{
    const s=read('web/frontend/src/lib/sse.ts')
    expect(s).toContain('__CGO_SSE')
  })
})
