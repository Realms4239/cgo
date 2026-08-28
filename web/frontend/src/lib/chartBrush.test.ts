// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
function readLive(): string {
  const cands = ['web/frontend/src/views/LiveView.tsx', 'src/views/LiveView.tsx', 'C:/cgo/.worktrees/observatory2/web/frontend/src/views/LiveView.tsx']
  for (const p of cands) try { return readFileSync(p, 'utf8') } catch {}
  return readFileSync('web/frontend/src/views/LiveView.tsx', 'utf8')
}
describe('clean triple', () => {
  it('LiveView has per-series markArea, no vestigial brush chrome', () => {
    const s = readLive()
    expect(s).not.toContain('brushType')
    expect(s).toContain('markArea')
  })
})
