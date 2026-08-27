import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest file check
import { readFileSync, existsSync } from 'node:fs'
// @ts-ignore node:path
import { resolve } from 'node:path'
function readAny(candidates: string[]): string {
  for (const p of candidates) {
    if (existsSync(p)) return readFileSync(p, 'utf8')
    const alt = resolve(p)
    if (existsSync(alt)) return readFileSync(alt, 'utf8')
  }
  // last try: fallback to repo-root relative
  return readFileSync(candidates[0], 'utf8')
}
describe('Rail', () => {
  it('Rail.tsx exists with mini sparkline and pin', () => {
    const s = readAny(['web/frontend/src/components/Rail.tsx','src/components/Rail.tsx','C:/cgo/.worktrees/observatory2/web/frontend/src/components/Rail.tsx'])
    expect(s).toContain('railPinned')
    expect(s).toContain('miniSparkline')
  })
  it('store adds railPinned and density', () => {
    const s = readAny(['web/frontend/src/store/ui.ts','src/store/ui.ts','C:/cgo/.worktrees/observatory2/web/frontend/src/store/ui.ts'])
    expect(s).toContain('railPinned')
    expect(s).toContain('density')
  })
})
describe('Rail polish', () => {
  it('Rail has icons and bento scoped', () => {
    const r=readAny(['web/frontend/src/components/Rail.tsx','src/components/Rail.tsx'])
    expect(r).toContain('nav-icon')
    expect(r).toContain('ICONS')
    const css=readAny(['web/frontend/src/styles/index.css','src/styles/index.css'])
    expect(css).toContain('#v-campagne .panel-stack')
    expect(css).toMatch(/max-width:\s*1280/)
  })
})
