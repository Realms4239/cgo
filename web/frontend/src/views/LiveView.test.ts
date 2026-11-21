// Porte « LiveView metrics » retargetée T6 : les métriques ne sont plus dans
// une vue écran mais dans les anneaux typés et le champ qui les peint.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('instrument — mesures', () => {
  it('les anneaux portent le délai, le débit et l’interface', () => {
    const s = read('lib/ringBuffer.ts')
    for (const ring of ['small:', 'rtt50:', 'rtt95:', 'goodput:', 'ifaceRx:', 'ifaceTx:', 'events:']) {
      expect(s, ring).toContain(ring)
    }
  })

  it('le champ affiche délai sous charge, débit utile et interface', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain("rawLineSeries('Délai sous charge'")
    expect(t).toContain("rawLineSeries('Aller-retour au repos'")
    expect(t).toContain("rawLineSeries('Aller-retour sous charge'")
    expect(t).toContain('Débit utile')
    expect(t).toContain('Interface RX')
    expect(t).not.toContain('p95')
  })

  it('au repos, le champ annonce l’absence sans zéro fabriqué', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain("empty ? 'Aucun signal' : 'En direct'")
    expect(t).toContain('La trace apparaîtra ici dès la première mesure.')
    expect(read('lib/ringBuffer.ts')).toContain('NaN = absence (jamais 0)')
  })
})
