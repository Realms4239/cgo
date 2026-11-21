// Porte « beam/donut » retargetée T6 : le signal vivant n'est plus un ruban
// animé mais les anneaux typés ; le donut d'absence devient le StateSlot.
// Les garanties sont conservées ou renforcées : painted only when dirty, trous
// interrompus, absence dite « — », aucun chrome de graphe superflu.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('instrument — signal vivant, absence honnête', () => {
  it('la peinture est pilotée par les anneaux et les trous coupent le tracé', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain('rings.small.dirty')
    expect(t).toContain('rawLineSeries')
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('connectNulls: false')
  })

  it('grammaire propre — pas de slider ni de visualMap piecewise', () => {
    const g = read('lib/chartGrammar.ts')
    expect(g).not.toContain("type: 'slider'")
    expect(g).toContain("type: 'inside'")
    expect(g).not.toContain("type: 'piecewise'")
    const o = read('components/labo/Observer.tsx')
    expect(o).not.toContain('toolbox:')
    expect(o).toContain('<TimeField')
  })

  it('mouvement réduit respecté et gestes épinglés à 44 px', () => {
    expect(read('styles/index.css')).toContain('@media (prefers-reduced-motion: no-preference)')
    const css = read('styles/components.css')
    expect(css).toContain('.command-header { position: sticky')
    expect(css).toContain('.ch-actions .btn { min-height: 44px')
    expect(css).toContain('.pr-step { min-height: 44px')
  })

  it('un créneau absent dit « — » et garde sa hauteur', () => {
    const s = read('components/ui/StateSlot.tsx')
    expect(s).toContain('—')
    expect(s).toContain("data-state={absent ? 'absent' : 'measured'}")
    expect(s).toContain("minHeight: 'var(--fs-metric)'")
  })

  it('charge markArea pilotée par la phase, jamais par un quartile', () => {
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('chargeMarkArea')
    expect(g).toContain("formatter: 'CHARGE'")
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain('objectiveMarkLine(goal)')
    expect(t).not.toContain('Math.floor(')
  })
})

describe('grammaire parameterisée — craft (plans §3)', () => {
  it('une fabrique expose line, bar, area et scatter', () => {
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('export function lineSeries')
    expect(g).toContain('export function barSeries')
    expect(g).toContain('export function scatterSeries')
    expect(g).toContain('unit')
    expect(g).not.toContain("formatter: 'max {c} ms'")
  })

  it('DisplayControls reste une bascule de craft, pas un second rail', () => {
    const p = read('components/DisplayControls.tsx')
    expect(p).toContain("'line'")
    expect(p).toContain("'bar'")
    expect(p).toContain("'area'")
    expect(p).toContain('data-chart-craft')
    expect(p).toContain('panel-chooser-tri')
  })

  it('Observer consomme les anneaux réels, pas un état copié', () => {
    const o = read('components/labo/Observer.tsx')
    expect(o).toContain('<TimeField rings={instrument}')
    expect(o).toContain('<JournalRail events={instrument.events}')
  })

  it('Comparer passe par lagrammaire de compromis, sans boîte à outils', () => {
    const c = read('components/labo/Comparer.tsx')
    expect(c).toContain('paretoFrontier')
    expect(c).not.toContain('toolbox')
  })

  it('Timeline 48 garde des bornes figées et un curseur sans reconstruction', () => {
    const t = read('components/Timeline.tsx')
    expect(t).toContain('setInterval')
    expect(t).toContain('recupEnd - baselineStart')
    expect(t).not.toContain('brushX')
    expect(t).not.toContain('Math.max(tRecup, now)')
  })

  it('provenance : hash8 vient de /api/integrity, sceau vérifié par acte', () => {
    const p = read('components/labo/Preuves.tsx')
    expect(p).toContain("fetch('/api/integrity')")
    expect(p).toContain('hash8')
    const s = read('components/labo/SealInspector.tsx')
    expect(s).toContain('available === true')
    expect(s).toContain('/^[0-9a-f]{64}$/i')
  })
})

describe('Couche B — primitives + grammaire', () => {
  it('les primitives parlent jetons, pas hex', () => {
    for (const f of ['Card', 'CardHead', 'Stat', 'Pill', 'EmptyChart']) {
      const s = read(`components/ui/${f}.tsx`)
      expect(s).toContain('export function')
      expect(s).toContain('var(--')
    }
  })

  it('filigrane hors idle seulement, palette craft exportée', () => {
    const g = read('lib/chartGrammar.ts')
    expect(g).toContain('export const CRAFT')
    expect(g).toContain('idle')
    expect(g).not.toContain("graphic: [\n      // 28px")
  })

  it('Observer — aucune étiquette de spécification, tête via CardHead', () => {
    const o = read('components/labo/Observer.tsx')
    expect(o).not.toContain('hero 300px')
    expect(o).toContain('CardHead')
  })
})
