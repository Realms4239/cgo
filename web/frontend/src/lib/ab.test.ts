// Porte « ab » retargetée T6 : le duel vit dans Comparer. On ne demande plus
// un décor (baseline/CAKE) mais l'invariant honnête : deux mesures complètes,
// un refus explicite côté vide, et une vraie ligne d'objectif dans l'instrument.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('ab — duel et ligne de référence', () => {
  it('deux mesures complètes seulement ; un côté vide refuse sans égalité inventée', () => {
    const s = read('components/labo/Comparer.tsx')
    expect(s).toContain('Sélectionnez deux mesures')
    expect(s).not.toContain('égalité')
    expect(s).toContain('référence')
    expect(s).toContain('paretoFrontier')
  })

  it('la référence est une vraie ligne d’objectif portée par l’instrument', () => {
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('objectiveMarkLine')
    expect(g).toContain('objectif ${Math.round(goal)} ms')
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain('objectiveMarkLine(goal)')
    expect(t).toContain('objectif {objectiveMs} ms')
  })
})
