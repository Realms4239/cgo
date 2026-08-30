import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
function read(p:string){ for(const c of [p, p.replace(/^web\/frontend\//,''), `../${p}`, `../../${p}`, `web/frontend/src/components/${p.split('/').pop()}`]){ try{ return readFileSync(c,'utf8') }catch{} } return readFileSync(p,'utf8') }
describe('prompt', () => {
  it('popup de progression — countdown réel, re-pop par event, pas de timer en dur', () => {
    const s=read('web/frontend/src/components/QuickActionsPrompt.tsx')
    // le total et le countdown viennent du snapshot serveur, jamais en dur
    expect(s).toContain('total_events')
    expect(s).toContain('phase_total_s')
    // re-pop seulement au changement d'event, pas toutes les 8 s
    expect(s).toContain('closedForEvent')
    expect(s).not.toContain('8000')
    // skip exposé dans la popup
    expect(s).toContain('/api/run/skip')
  })
  it('palette has Cmd-K', () => {
    const s=read('web/frontend/src/components/CommandPalette.tsx')
    expect(s).toContain('Cmd-K')
    expect(s).toContain('createTimeline')
  })
})
