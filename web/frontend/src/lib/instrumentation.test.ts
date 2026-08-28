import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest jsdom ponytail minimal
import { readFileSync } from 'node:fs'
describe('instrumentation', ()=>{
  it('live exposes __CGO_LIVE', ()=>{
    const s=readFileSync('web/frontend/src/lib/live.ts','utf8')
    expect(s).toContain('__CGO_LIVE')
  })
  it('sse exposes __CGO_SSE', ()=>{
    const s=readFileSync('web/frontend/src/lib/sse.ts','utf8')
    expect(s).toContain('__CGO_SSE')
  })
})
