// TDD — la bande de provenance ne dit que ce que les sources réelles disent.
// Une absence garde son tiret et son mot français : aucun chiffre, sceau ou
// « PASS » ne vient remplir une ligne vide, et nulle mesure n'est déduite des
// trames du banc en mode observe.
import type { ComponentProps } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import ProvenanceStrip from './ProvenanceStrip'

type Props = ComponentProps<typeof ProvenanceStrip>

const mount = (props: Props = {}): HTMLElement => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<ProvenanceStrip {...props} />)
  return host
}
const textOf = (props: Props = {}): string => mount(props).textContent ?? ''

describe('ProvenanceStrip', () => {
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
