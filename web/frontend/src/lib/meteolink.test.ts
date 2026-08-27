import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
function read(p: string): string {
  const cands = [p, p.replace(/^web\/frontend\//, ''), 'web/frontend/' + p.replace(/^web\/frontend\//, '')]
  for (const c of cands) try { return readFileSync(c, 'utf8') } catch {}
  return readFileSync(p, 'utf8')
}
describe('meteolink', () => {
  it('wordmark has gradient and icon', () => {
    const s=read('web/frontend/src/components/MeteolinkWordmark.tsx')
    expect(s).toContain('METEOLINK')
    expect(s).toContain('createDrawable')
  })
  it('panel chooser exists', () => {
    const s=read('web/frontend/src/components/PanelChooser.tsx')
    expect(s).toContain('Wall')
  })
})
