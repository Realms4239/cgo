// Oracle indépendant : on monte les vrais composants avec des anneaux vides,
// puis on lit le CSS réel. Aucun chiffre n'est fabriqué pour une absence.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField from './TimeField'
import JournalRail from './JournalRail'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

function emptyRings(): InstrumentRings {
  const mk = () => new TypedMetricRing(1)
  return { small: mk(), rtt50: mk(), rtt95: mk(), goodput: mk(), ifaceRx: mk(), ifaceTx: mk() }
}

describe('cockpit field — placement et absence honnête', () => {
  it('le champ et le journal partagent le rythme déclaré', () => {
    const css = readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')
    expect(css).toContain('.cockpit-field { min-width: 0; display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--cockpit-rhythm); }')
    expect(css).toContain('.time-field { min-width: 0; display: flex; flex-direction: column; gap: var(--cockpit-rhythm); }')
    expect(css).toContain('.journal-rail { min-width: 0; display: flex; flex-direction: column; gap: var(--cockpit-rhythm); max-height: 100%; }')
  })
  it('sans événement, le journal dit l’absence sans chiffre ni vert', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<JournalRail events={new DomainEventRing(4)} /> as never)
    const empty = host.querySelector<HTMLElement>('.c-empty')
    expect(empty).not.toBeNull()
    expect(empty!.textContent).toContain('Aucun événement')
    expect(empty!.textContent).not.toMatch(/\d/)
    expect(empty!.className).not.toMatch(/(?:^|\s)(?:ok|success|green|pass)(?:\s|$)/i)
  })
})
