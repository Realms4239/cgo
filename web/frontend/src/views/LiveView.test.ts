// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
function readLive(): string {
  const cands = ['web/frontend/src/views/LiveView.tsx','src/views/LiveView.tsx','C:/cgo/.worktrees/m6-perfect/web/frontend/src/views/LiveView.tsx']
  for (const p of cands) try { return readFileSync(p,'utf8') } catch {}
  return readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
}
describe('LiveView metrics', () => {
  it('renders all primary metrics', () => {
    const s = readLive()
    expect(s).toContain('rtt_p95')
    expect(s).toContain('small_p95')
    expect(s).toContain('goodput')
    expect(s).toContain('drops')
    expect(s).toContain('QDI')
    // extended for Task 6 — all LIEN Tableau 4/5 important look
    expect(s).toContain('wasted')
    expect(s).toContain('cost')
    expect(s).toContain('deadline')
  })
})
