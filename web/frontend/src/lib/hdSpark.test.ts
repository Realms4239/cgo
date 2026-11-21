// Porte « hdSpark » retargetée T6 : le délai souscharge occupe le champ complet
// de l'instrument, avec un objectif nommé et une hauteur mesurée.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('hdSpark', () => {
  it('MetricCard garde son sparkline 60×12', () => {
    const s = read('components/ui/MetricCard.tsx')
    expect(s).toContain('60')
    expect(s).toContain('clipPath')
  })
  it('le délai sous charge occupe le champ, objectif inclus', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain("rawLineSeries('Délai sous charge'")
    expect(t).toContain("rawLineSeries('Aller-retour sous charge'")
    expect(t).not.toContain('p95')
    expect(t).toContain('objectif {objectiveMs} ms')
    const css = read('styles/components.css')
    expect(css).toContain('.observer-field .tf-plot { height: 320px; }')
  })
})
