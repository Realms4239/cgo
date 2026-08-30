import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
function read(p:string){ for(const c of [p, p.replace(/^web\/frontend\//,''), `../${p}`, `../../${p}`, `web/frontend/src/components/${p.split('/').pop()}`]){ try{ return readFileSync(c,'utf8') }catch{} } return readFileSync(p,'utf8') }
describe('prompt', () => {
  it('prompt has 6s timer and idle 8s', () => {
    const s=read('web/frontend/src/components/QuickActionsPrompt.tsx')
    expect(s).toContain('6000')
    expect(s).toContain('8000')
  })
  it('palette has Cmd-K', () => {
    const s=read('web/frontend/src/components/CommandPalette.tsx')
    expect(s).toContain('Cmd-K')
    expect(s).toContain('createTimeline')
  })
})
