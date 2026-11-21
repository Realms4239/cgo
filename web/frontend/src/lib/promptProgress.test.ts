import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
function read(p:string){ for(const c of [p, p.replace(/^web\/frontend\//,''), `../${p}`, `../../${p}`, `web/frontend/src/components/${p.split('/').pop()}`]){ try{ return readFileSync(c,'utf8') }catch{} } return readFileSync(p,'utf8') }
describe('promptProgress', () => {
  // T6-FIX — la barre de 6 s décorative a disparu avec PromptProgressLine.
  // La garantie est désormais portée par le vrai compte à rebours SSE de
  // QuickActionsPrompt : une durée réelle, pas une animation codée en dur.
  it('compte à rebours réel issu de la frame, pas un drain décoratif', () => {
    const s=read('web/frontend/src/components/QuickActionsPrompt.tsx')
    expect(s).toContain('phaseTotal')
    expect(s).toContain('setInterval')
    expect(s).toContain('var(--t-threshold)')
    expect(s).not.toContain('6000ms')
  })
  it('quick prompt shows progress not choices when running', () => {
    const s=read('web/frontend/src/components/QuickActionsPrompt.tsx')
    expect(s).toContain('Event')
    expect(s).not.toContain('Actions rapides')
  })
})
