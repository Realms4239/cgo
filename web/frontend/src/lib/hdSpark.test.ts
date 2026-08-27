import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
function read(p: string): string {
  const cands = [p, p.replace(/^web\/frontend\//, ''), 'web/frontend/' + p.replace(/^web\/frontend\//, '')]
  for (const c of cands) try { return readFileSync(c, 'utf8') } catch {}
  return readFileSync(p, 'utf8')
}
describe('hdSpark', () => {
  it('MetricCard has sparkline 60×12', () => {
    const s=read('web/frontend/src/components/ui/MetricCard.tsx')
    expect(s).toContain('60')
    expect(s).toContain('clipPath')
  })
  it('Live hero small p95 full-width', () => {
    const s=read('web/frontend/src/views/LiveView.tsx')
    expect(s).toContain('small_p95')
    expect(s).toContain('300px')
  })
})
