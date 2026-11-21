import type { ComponentProps } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TrustStrip from './TrustStrip'

function mount(props: Partial<ComponentProps<typeof TrustStrip>> = {}): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<TrustStrip connection="online" {...props} />)
  return host
}

describe('TrustStrip', () => {
  it('dit le mot HORS LIGNE, pas seulement une pastille', () => {
    const host = mount({ connection: 'offline' })
    expect(host.querySelector('[data-slot="connection"]')?.textContent).toContain('HORS LIGNE')
  })

  it('compte les trames perdues en toutes lettres', () => {
    const host = mount({ connection: 'online', backpressure: 3 })
    expect(host.querySelector('[data-slot="backpressure"]')?.textContent).toContain('3 trames perdues')
  })

  it('nomme ABSENT un sceau manquant plutôt d’écrire zéro', () => {
    const host = mount({ connection: 'online' })
    const seal = host.querySelector('[data-slot="seal"]')?.textContent || ''
    expect(seal).toContain('ABSENT')
    expect(seal).not.toMatch(/\d/)
  })

  it('garde les cinq créneaux dans l’ordre, chacun mot + forme', () => {
    const host = mount({ connection: 'online', source: 'ebpf', freshness: 'frais', seal: 'scellé', backpressure: 0 })
    const slots = Array.from(host.querySelectorAll('[data-slot]'))
    expect(slots.map(s => s.getAttribute('data-slot'))).toEqual([
      'connection', 'source', 'freshness', 'seal', 'backpressure',
    ])
    for (const slot of slots) {
      expect(slot.querySelector('.trust-shape'), slot.getAttribute('data-slot') || '').not.toBeNull()
      expect((slot.querySelector('.trust-word')?.textContent || '').trim().length).toBeGreaterThan(0)
    }
  })

  it('ajoute un bloc saturé et le mot ANOMALIE dès qu’un fait est anormal', () => {
    const bad = mount({ connection: 'offline' })
    expect(bad.querySelector('.trust-strip')?.getAttribute('data-abnormal')).toBe('1')
    expect(bad.querySelector('.trust-flag')).not.toBeNull()
    expect(bad.textContent).toContain('ANOMALIE')

    const good = mount({ connection: 'online', source: 'ebpf', freshness: 'frais', seal: 'scellé', backpressure: 0 })
    expect(good.querySelector('.trust-strip')?.getAttribute('data-abnormal')).toBe('0')
    expect(good.querySelector('.trust-flag')).toBeNull()
  })
})
