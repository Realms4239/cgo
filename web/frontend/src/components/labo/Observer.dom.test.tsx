// TDD — la bande de repères ne montre QUE des événements réels. Anneau vide :
// « aucun repère » en gris (aucun chiffre, aucun vert). Un fait réel : une
// marque avec son heure réelle et son mot humain, jamais une gravité inventée.
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it } from 'vitest'
import { DomainEventRing } from '../../lib/ringBuffer'
import { EventTimeline } from './Observer'

function mount(events: DomainEventRing): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<EventTimeline events={events} />)
  return host
}

describe('EventTimeline — repères honnêtes', () => {
  let ring: DomainEventRing
  beforeEach(() => { ring = new DomainEventRing() })

  it('anneau vide : aucun repère, état no-data, pas de chiffre ni de vert', () => {
    const host = mount(ring)
    expect(host.querySelector('.event-timeline')?.getAttribute('data-state')).toBe('no-data')
    expect(host.textContent).toContain('aucun repère')
    expect(host.querySelector('.et-mark-empty')).not.toBeNull()
    expect(host.querySelector('.et-event')).toBeNull()
    expect(host.querySelector('[data-state="ok"], [data-state="pass"], [data-state="healthy"]')).toBeNull()
    expect(host.textContent).not.toMatch(/\d/)
  })

  it('un événement réel : une marque, l’heure réelle et le mot humain du fait', () => {
    ring.push({ ts: Date.UTC(2026, 10, 21, 13, 45, 2), kind: 'phase', label: 'Charge' })
    const host = mount(ring)
    expect(host.querySelector('.event-timeline')?.getAttribute('data-state')).toBe('events')
    const marks = host.querySelectorAll('.et-event')
    expect(marks).toHaveLength(1)
    expect(marks[0].textContent).toContain('13:45:02')
    expect(marks[0].textContent).toContain('Charge')
    expect(marks[0].textContent).not.toContain('phase')
    expect(host.textContent).not.toContain('aucun repère')
  })

  it('deux faits réels : deux marques, ordre chronologique, sans compte fabriqué', () => {
    ring.push({ ts: Date.UTC(2026, 10, 21, 13, 45, 2), kind: 'phase', label: 'Charge' })
    ring.push({ ts: Date.UTC(2026, 10, 21, 13, 46, 7), kind: 'gates', label: 'Portes' })
    const host = mount(ring)
    const marks = host.querySelectorAll('.et-event')
    expect(marks).toHaveLength(2)
    expect(marks[0].textContent).toContain('13:45:02')
    expect(marks[1].textContent).toContain('13:46:07')
    expect(marks[1].textContent).toContain('Portes')
    expect(host.querySelector('[data-state="ok"], [data-state="pass"]')).toBeNull()
  })

  it('libellé absent : repli honnête sur le kind, sans mot fabriqué', () => {
    ring.push({ ts: Date.UTC(2026, 10, 21, 13, 45, 2), kind: 'gates', label: '' })
    const mark = mount(ring).querySelector('.et-event')
    expect(mark?.textContent).toContain('gates')
  })

  it('horodatage invalide : aucun jetage, ni heure ni dateTime fabriquée', () => {
    ring.push({ ts: Number.NaN, kind: 'phase', label: 'Charge' })
    let host!: HTMLElement
    expect(() => { host = mount(ring) }).not.toThrow()
    const t = host.querySelector('.et-ts')
    expect(t?.textContent).toBe('')
    expect(t?.hasAttribute('datetime')).toBe(false)
    expect(host.textContent).not.toMatch(/Invalid Date|NaN/)
    expect(host.textContent).toContain('Charge')
  })

  it('deux faits identiques : deux marques distinctes (clés stables)', () => {
    const e = { ts: Date.UTC(2026, 10, 21, 13, 45, 2), kind: 'phase', label: 'Charge' }
    ring.push(e)
    ring.push({ ...e })
    expect(mount(ring).querySelectorAll('.et-event')).toHaveLength(2)
  })

  it('anneau vide : le vide enseigne ce qui viendra et pourquoi, sans chiffre ni vert', () => {
    const host = mount(ring)
    const empty = host.querySelector('.et-empty')
    expect(empty?.getAttribute('role')).toBe('status')
    expect(empty?.textContent).toContain('aucun repère')
    expect(empty?.textContent).toContain("s'afficheront ici")
    expect(empty?.textContent).toContain('heure')
    expect(host.querySelector('.et-mark-empty')).not.toBeNull()
    expect(host.textContent).not.toMatch(/\d/)
  })
})
