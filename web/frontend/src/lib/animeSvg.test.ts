import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync, existsSync } from 'node:fs'
function readAnime(): string {
  const cands = ['web/frontend/src/lib/anime.ts','src/lib/anime.ts','C:/cgo/.worktrees/m8/web/frontend/src/lib/anime.ts','C:/cgo/web/frontend/src/lib/anime.ts']
  for (const p of cands) { try { if (existsSync(p)) return readFileSync(p,'utf8') } catch {} try { return readFileSync(p,'utf8') } catch {} }
  return readFileSync('web/frontend/src/lib/anime.ts','utf8')
}
describe('animeSvg', () => {
  it('anime uses svg/text/animatable', () => {
    const s=readAnime()
    expect(s).toContain('createDrawable')
    expect(s).toContain('splitText')
  })
})
