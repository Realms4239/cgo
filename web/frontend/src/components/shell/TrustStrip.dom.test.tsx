import type { ComponentProps } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TrustStrip from './TrustStrip'

function mount(props: Partial<ComponentProps<typeof TrustStrip>> = {}): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<TrustStrip connection="online" {...props} />)
  return host
}

const ALL_GOOD = { source: 'ebpf', freshness: 'frais', seal: 'scellé', backpressure: 0 }

describe('TrustStrip — ligne de statut', () => {
  it('pose identité et version à gauche, sans doubler les commandes du chrome', () => {
    const host = mount(ALL_GOOD)
    const id = host.querySelector('.trust-id')
    expect(id).not.toBeNull()
    expect(id?.textContent).toContain('CGO')
    expect(id?.textContent).toMatch(/2\.0\.0/)
    expect(host.querySelector('.chrome-trigger')).toBeNull()
  })

  it('dit le mot HORS LIGNE, pas seulement une pastille', () => {
    const host = mount({ connection: 'offline' })
    expect(host.querySelector('[data-slot="connection"]')?.textContent).toContain('HORS LIGNE')
    expect(host.querySelector('[data-slot="connection"]')?.textContent).not.toContain('EN LIGNE')
  })

  it('compte les trames perdues en toutes lettres', () => {
    const host = mount({ connection: 'online', backpressure: 3 })
    expect(host.querySelector('[data-slot="backpressure"]')?.textContent).toContain('3 trames perdues')
  })

  it('nomme ABSENT un sceau manquant plutôt d’écrire zéro, en état discret', () => {
    const host = mount({ connection: 'online' })
    const seal = host.querySelector('[data-slot="seal"]')
    expect(seal?.textContent).toContain('ABSENT')
    expect(seal?.textContent).not.toMatch(/\d/)
    expect(seal?.getAttribute('data-state')).toBe('absent')
  })

  it('dit AUCUNE PERTE quand la mesure vaut zéro', () => {
    const host = mount({ connection: 'online', backpressure: 0 })
    expect(host.querySelector('[data-slot="backpressure"]')?.textContent).toContain('AUCUNE PERTE')
    expect(host.querySelector('[data-slot="backpressure"]')?.getAttribute('data-state')).toBe('ok')
  })

  it('garde les cinq créneaux dans l’ordre, chacun mot + forme + état', () => {
    const host = mount(ALL_GOOD)
    const slots = Array.from(host.querySelectorAll('[data-slot]'))
    expect(slots.map(s => s.getAttribute('data-slot'))).toEqual([
      'connection', 'source', 'freshness', 'seal', 'backpressure',
    ])
    for (const slot of slots) {
      expect(slot.querySelector('.trust-shape'), slot.getAttribute('data-slot') || '').not.toBeNull()
      expect((slot.querySelector('.trust-word')?.textContent || '').trim().length).toBeGreaterThan(0)
      expect(['ok', 'absent', 'alarm']).toContain(slot.getAttribute('data-state'))
    }
  })

  it('ne hurle pas : aucun bandeau ANOMALIE, une seule pastille allumée', () => {
    const bad = mount({ connection: 'offline', source: null, freshness: null, seal: null, backpressure: 4 })
    expect(bad.querySelector('.trust-alert')).toBeNull()
    expect(bad.textContent).not.toContain('ANOMALIE')
    expect(bad.querySelector('.trust-strip')?.getAttribute('data-abnormal')).toBe('1')
    const lit = bad.querySelectorAll('.trust-slot[data-lit="1"]')
    expect(lit).toHaveLength(1)
    expect(lit[0]?.getAttribute('data-slot')).toBe('connection')
  })

  it('reste calme quand tout va bien : aucune pastille allumée', () => {
    const good = mount(ALL_GOOD)
    expect(good.querySelector('.trust-strip')?.getAttribute('data-abnormal')).toBe('0')
    expect(good.querySelectorAll('.trust-slot[data-lit="1"]')).toHaveLength(0)
    expect(good.textContent).not.toContain('ANOMALIE')
  })

  it('allume la seule pastille en faute quand le lien tombe', () => {
    const bad = mount({ connection: 'offline', ...ALL_GOOD })
    const lit = bad.querySelectorAll('.trust-slot[data-lit="1"]')
    expect(lit).toHaveLength(1)
    expect(lit[0]?.getAttribute('data-slot')).toBe('connection')
  })

  it('allume la pastille des pertes quand elles sont la seule faute', () => {
    const bad = mount({ connection: 'online', ...ALL_GOOD, backpressure: 2 })
    const lit = bad.querySelectorAll('.trust-slot[data-lit="1"]')
    expect(lit).toHaveLength(1)
    expect(lit[0]?.getAttribute('data-slot')).toBe('backpressure')
  })
})
