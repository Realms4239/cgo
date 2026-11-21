// Porte « clean triple » retargetée T6 : l'instrument remplace le mur. Aucun
// chrome de brush, aucune ligne d'objectif estimée par quartile.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('clean triple', () => {
  it('le champ temporel n’a ni brush résiduel ni quartile inventé', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).not.toContain('brushType')
    expect(t).toContain('markLine: objectiveMarkLine(goal)')
    expect(t).not.toContain('Math.floor(')
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('data: [{ yAxis: goal }]')
  })
})
