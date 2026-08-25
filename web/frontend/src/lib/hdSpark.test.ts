import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
function readMC(): string {
  const cands = ['web/frontend/src/components/ui/MetricCard.tsx','src/components/ui/MetricCard.tsx','C:/cgo/.worktrees/m8/web/frontend/src/components/ui/MetricCard.tsx']
  for (const p of cands) try { return readFileSync(p,'utf8') } catch {}
  return readFileSync('web/frontend/src/components/ui/MetricCard.tsx','utf8')
}
describe('hdSpark', () => {
  it('MetricCard has sparkline 60×12', () => {
    const s=readMC()
    expect(s).toContain('60')
    expect(s).toContain('clipPath')
  })
})
