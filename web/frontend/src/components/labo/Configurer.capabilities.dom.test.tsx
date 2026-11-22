// TDD — les contrôles du Labo ne disent que ce que les sources réelles portent.
// Aucune absence n'invente un chiffre, un sceau ou une option fictive.
import type { ReactNode } from 'react'
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Configurer from './Configurer'
import type { Capabilities } from '../../lib/capabilities'
import type { Param } from '../../lib/schema'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const CAPS: Capabilities = {
  qdiscs: ['pfifo_fast', 'fq_codel', 'cake'],
  ccs: ['cubic', 'bbr'],
  known_ccs: ['cubic', 'bbr', 'bbr3'],
  kernel: '6.8.0-45-generic',
  bbr3: false,
}

const withEnum = (key: string, values: string[], extra: Partial<Param> = {}): Param => ({
  key, label: key, default: values[0] ?? '', desc: '', enum: values, ...extra,
})

const SCHEMA: Param[] = [
  withEnum('link_channel', ['statique', 'empirique'], { default: 'statique' }),
  withEnum('link_direction', ['up', 'down', 'both'], { default: 'up' }),
  withEnum('dist_name', [], { default: '—' }),
  { key: 'bbr3_state', label: 'État BBRv3', default: 'BBRv3 non détecté — CUBIC + BBRv1 sur noyau 6.8, détection bbr3 prête', desc: '' },
]

const CALIBS = [
  { name: 'fibre-20m', sha256: 'a'.repeat(64), samples: 120, p50_ms: 14.2, verified: true },
  { name: 'fantome', sha256: 'b'.repeat(64), samples: 5, p50_ms: 2, verified: false },
]

const mount = (node: ReactNode): HTMLElement => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}
const rendered = (props: Record<string, unknown>) => mount(<Configurer {...props} />)

// Les fixtures injectées sont la source du test : aucun hook ne doit sonder le réseau.
beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, json: async () => ({}) }))
})
afterEach(() => vi.unstubAllGlobals())

describe('Configurer — capacités et configuration réelles', () => {
  it('(a) la bande de capacités vient de /api/capabilities et nomme BBRv3 absent', () => {
    const host = rendered({ caps: CAPS, params: SCHEMA, calibrations: CALIBS })
    const strip = host.querySelector('.configurer-caps')?.textContent ?? ''
    expect(strip).toContain('pfifo_fast')
    expect(strip).toContain('cake')
    expect(strip).toContain('cubic')
    expect(host.innerHTML).toContain('6.8.0-45-generic') // noyau porté par le titre honnête
    expect(strip).toContain('BBRv3 non détecté')
  })

  it('(b) Canal n\'offre que les valeurs réelles du schéma', () => {
    const chan = rendered({ caps: CAPS, params: SCHEMA, calibrations: CALIBS })
      .querySelector('.configurer-control[data-control="link_channel"]')!
    expect(chan.textContent).toContain('Statique')
    expect(chan.textContent).toContain('Empirique .dist')
    expect(chan.querySelectorAll('button[aria-pressed]')).toHaveLength(2)
  })

  it('(b2) Canal garde le libellé empirique mais le désactive sans calibration vérifiée', () => {
    const chan = rendered({ caps: CAPS, params: SCHEMA, calibrations: [] })
      .querySelector('.configurer-control[data-control="link_channel"]')!
    expect(chan.textContent).toContain('Statique')
    expect(chan.textContent).toContain('Empirique .dist')
    expect(chan.textContent).toContain('aucune calibration vérifiée')
    const buttons = chan.querySelectorAll<HTMLButtonElement>('button')
    expect(buttons).toHaveLength(2)
    expect(buttons[0].disabled).toBe(false)
    expect(buttons[1].disabled).toBe(true)
    expect(buttons[1].getAttribute('aria-disabled')).toBe('true')
  })

  it('(c) Direction traduit up/down/both, « Les deux » seulement si le schéma le porte', () => {
    const dir = rendered({ caps: CAPS, params: SCHEMA, calibrations: CALIBS })
      .querySelector('.configurer-control[data-control="link_direction"]')!
    expect(dir.textContent).toContain('Egress WAN')
    expect(dir.textContent).toContain('Ingress IFB')
    expect(dir.textContent).toContain('Les deux')
    expect(dir.querySelectorAll('button[aria-pressed]')).toHaveLength(3)

    const absent = rendered({ caps: null, params: [withEnum('link_direction', [])], calibrations: [] })
      .querySelector('.configurer-control[data-control="link_direction"]')!
    expect(absent.textContent).toContain('—')
    expect(absent.textContent).toContain('direction indisponible')
    expect(absent.querySelector('button')).toBeNull()
  })

  it('(d) Distribution ne liste que les sidecars vérifiés, sinon elle dit son absence', () => {
    const dist = rendered({ caps: CAPS, params: SCHEMA, calibrations: CALIBS })
      .querySelector('.configurer-control[data-control="distribution"]')!
    const names = Array.from(dist.querySelectorAll('option')).map(o => o.textContent)
    expect(names).toContain('fibre-20m')
    expect(names).not.toContain('fantome')
    expect(dist.querySelector<HTMLSelectElement>('select')!.disabled).toBe(false)

    const empty = rendered({ caps: null, params: SCHEMA, calibrations: [] })
      .querySelector('.configurer-control[data-control="distribution"]')!
    expect(empty.textContent).toContain('— aucune calibration vérifiée')
    expect(empty.querySelector<HTMLSelectElement>('select')!.disabled).toBe(true)
    expect(empty.querySelectorAll('option')).toHaveLength(0)
  })

  it('(e) aucun état d\'absence ne porte de chiffre, PASS ou VÉRIFIÉ', () => {
    const absent = rendered({ caps: null, params: [withEnum('link_direction', [])], calibrations: [] })
    const zones = [
      absent.querySelector('.configurer-control[data-control="link_direction"]')?.textContent ?? '',
      absent.querySelector('.configurer-control[data-control="distribution"]')?.textContent ?? '',
    ]
    expect(zones).toHaveLength(2)
    for (const text of zones) {
      expect(text.length, 'la zone d\'absence doit exister').toBeGreaterThan(0)
      expect(text).not.toMatch(/\d/)
      expect(text).not.toContain('PASS')
      expect(text).not.toContain('VÉRIFIÉ')
    }
  })

  it('la sélection change aria-pressed sans revendiquer la moindre mesure', () => {
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    try {
      act(() => { root.render(<Configurer caps={CAPS} params={SCHEMA} calibrations={CALIBS} />) })
      const buttons = () => Array.from(host.querySelectorAll<HTMLButtonElement>('.configurer-control[data-control="link_channel"] button'))
      expect(buttons()[0].getAttribute('aria-pressed')).toBe('true')
      act(() => { buttons()[1].click() })
      expect(buttons()[0].getAttribute('aria-pressed')).toBe('false')
      expect(buttons()[1].getAttribute('aria-pressed')).toBe('true')
      expect(host.textContent).not.toContain('PASS')
    } finally {
      act(() => root.unmount())
      host.remove()
    }
  })
})
