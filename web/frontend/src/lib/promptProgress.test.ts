import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
function read(p:string){ for(const c of [p, p.replace(/^web\/frontend\//,''), `../${p}`, `../../${p}`, `web/frontend/src/components/${p.split('/').pop()}`]){ try{ return readFileSync(c,'utf8') }catch{} } return readFileSync(p,'utf8') }
describe('promptProgress', () => {
  it('progress line 6s bottom amber', () => {
    const s=read('web/frontend/src/components/PromptProgressLine.tsx')
    expect(s).toContain('6000')
    expect(s).toContain('#f4b400')
    expect(s).toContain('paused')
  })
  it('quick prompt shows progress not choices when running', () => {
    const s=read('web/frontend/src/components/QuickActionsPrompt.tsx')
    expect(s).toContain('Event')
    expect(s).not.toContain('Actions rapides')
  })
})
