// Porte de performance des graphes : canvas + dirty-rect partout, DPR borné,
// animations coupées. Le seul initialiseur restant est le champ de l'instrument.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('graphes — performance', () => {
  it('chaque init est canvas, dirty-rect, DPR ≤ 2', () => {
    for (const f of ['components/cockpit/TimeField.tsx']) {
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
