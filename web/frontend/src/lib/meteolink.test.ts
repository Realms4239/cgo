import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs for vitest file check
import { readFileSync, existsSync } from 'node:fs'
function readAny(cands: string[]): string {
  for (const p of cands) {
    try {
      if (existsSync(p)) return readFileSync(p, 'utf8')
    } catch {}
    try { return readFileSync(p, 'utf8') } catch {}
  }
  return readFileSync(cands[0], 'utf8')
}
describe('meteolink', () => {
  it('wordmark has gradient and icon', () => {
    const s=readAny(['web/frontend/src/components/MeteolinkWordmark.tsx','src/components/MeteolinkWordmark.tsx','web/frontend/src/components/MeteolinkWordmark.tsx','C:/cgo/.worktrees/m8/web/frontend/src/components/MeteolinkWordmark.tsx'])
    expect(s).toContain('METEOLINK')
    expect(s).toContain('createDrawable')
  })
})
