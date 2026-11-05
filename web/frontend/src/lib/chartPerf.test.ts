// Porte de performance des graphes : canvas + dirty-rect partout, DPR borné,
// animations coupées. Un nouveau graphe qui n'applique pas ces quatre règles
// se paie en mémoire (DPR 3 sur écran 300 %) et en saccades.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('graphes — performance', () => {
  it('chaque init est canvas, dirty-rect, DPR ≤ 2', () => {
    for (const f of ['views/LiveView.tsx', 'views/ResultatsView.tsx', 'components/CompareView.tsx']) {
      const s = read(f)
      expect(s, f).toContain("renderer: 'canvas'")
      expect(s, f).toContain('useDirtyRect: true')
      expect(s, f).toContain('Math.min(window.devicePixelRatio, 2)')
    }
  })
  it('les animations sont coupées par défaut', () => {
    expect(read('lib/chartGrammar.ts')).toContain('animation: false')
  })
})
