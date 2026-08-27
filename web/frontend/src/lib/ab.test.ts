// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
declare const process: any
function readLive(): string {
  const cands = ['web/frontend/src/views/LiveView.tsx','src/views/LiveView.tsx','web/frontend/src/lib/../views/LiveView.tsx']
  for (const p of cands) try { return readFileSync(p,'utf8') } catch {}
  try { return readFileSync('C:/cgo/.worktrees/m9e/web/frontend/src/views/LiveView.tsx','utf8') } catch {}
  return readFileSync('web/frontend/src/views/LiveView.tsx','utf8')
}
describe('ab', () => {
  it('live wall overlay', () => {
    const s=readLive()
    expect(s).toContain('baseline')
    expect(s).toContain('CAKE')
  })
})
