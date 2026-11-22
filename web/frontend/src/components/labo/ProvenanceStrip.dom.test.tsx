// TDD — la bande de provenance ne dit que ce que les sources réelles disent.
// Une absence garde son tiret et son mot français : aucun chiffre, sceau ou
// « PASS » ne vient remplir une ligne vide, et nulle mesure n'est déduite des
// trames du banc en mode observe.
import type { ComponentProps } from 'react'
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ProvenanceStrip from './ProvenanceStrip'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

type Props = ComponentProps<typeof ProvenanceStrip>

const mount = (props: Props = {}): HTMLElement => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<ProvenanceStrip {...props} />)
  return host
}
const textOf = (props: Props = {}): string => mount(props).textContent ?? ''

afterEach(() => vi.unstubAllGlobals())

describe('ProvenanceStrip', () => {
  it('ne sonde aucun /api quand les deux valeurs sont injectées (pas de requête en double)', () => {
    const fetchSpy = vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) })
    vi.stubGlobal('fetch', fetchSpy)
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    try {
      act(() => {
        root.render(<ProvenanceStrip
          integrity={{ integrity: 'verified', hash8: 'a1b2c3d4' }}
          doctor={{ mode: 'observe' }}
        />)
      })
      expect(fetchSpy).not.toHaveBeenCalled()
      expect(host.textContent).toContain('a1b2c3d4')
    } finally {
      act(() => root.unmount())
      host.remove()
    }
  })

  it('sonde /api/integrity et /api/doctor seulement quand rien n\'est injecté', () => {
    const fetchSpy = vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) })
    vi.stubGlobal('fetch', fetchSpy)
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    try {
      act(() => { root.render(<ProvenanceStrip />) })
      expect(fetchSpy).toHaveBeenCalledTimes(2)
      expect(fetchSpy).toHaveBeenCalledWith('/api/integrity')
      expect(fetchSpy).toHaveBeenCalledWith('/api/doctor')
    } finally {
      act(() => root.unmount())
      host.remove()
    }
  })

  it('rend three honest lines when every source is empty', () => {
    const host = mount({ integrity: null, doctor: null })
    expect(host.querySelectorAll('.provenance-line')).toHaveLength(3)
    const t = host.textContent ?? ''
    expect(t).toContain('— trace de terrain non mesurée')
    expect(t).toContain('— confrontation eBPF absente')
    expect(t).toContain('— intégrité indisponible')
    expect(t).toContain('capture de terrain autorisée absente')
    expect(t).toContain('ebpf_vs_tc_report.json')
    expect(t).toContain('/api/integrity')
  })

  it('invents no digit, green word or seal in an absent state', () => {
    const t = textOf({ integrity: {}, doctor: {} })
    expect(/\d/.test(t), 'aucun chiffre dans une absence').toBe(false)
    expect(t.toLowerCase()).not.toMatch(/pass|vérifi|scell|sceau|vert\b/)
    expect(t).toContain('— trace de terrain non mesurée')
    expect(t).toContain('— intégrité indisponible')
  })

  it('renders integrity only from the real integrity value', () => {
    const t = textOf({
      integrity: { available: true, integrity: 'verified', ok: true, hash8: 'a1b2c3d4' },
      doctor: null,
    })
    expect(t).toContain('a1b2c3d4')
    expect(t).toContain('vérifiée')
    expect(t).not.toContain('— intégrité indisponible')
    expect(t).toContain('— trace de terrain non mesurée')
    expect(t).toContain('— confrontation eBPF absente')
  })

  it('names a real degraded check without claiming verification', () => {
    const t = textOf({
      integrity: { available: true, integrity: 'degraded', ok: false, hash8: 'deadv001' },
      doctor: null,
    })
    expect(t).toContain('deadv001')
    expect(t).toContain('dégradée')
    expect(t).not.toContain('vérifiée')
  })

  it('a hash or ok flag alone is not a status — absence stays absence', () => {
    const t = textOf({ integrity: { available: true, ok: true, hash8: 'a1b2c3d4' }, doctor: null })
    expect(t).toContain('— intégrité indisponible')
    expect(t).not.toContain('a1b2c3d4')
    expect(t).not.toContain('vérifiée')
  })

  it('names the host mode as context, never as a field measurement', () => {
    const t = textOf({
      integrity: null,
      doctor: { mode: 'observe', checks: [{ name: 'os', status: 'ok', detail: 'windows/amd64' }] },
    })
    expect(t).toContain('mode observe')
    expect(t).not.toContain('windows/amd64')
    expect(t).toContain('— trace de terrain non mesurée')
    expect(/\d/.test(t)).toBe(false)
  })
})
