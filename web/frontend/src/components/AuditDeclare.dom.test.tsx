// Déclarer : les champs libres restent dans la maison NFC — un site saisi
// en forme décomposée est recomposé avant d'entrer dans le rapport.
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { AuditDeclare } from './AuditDeclare'
import { useAuditStore } from '../store/audit'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

function mount() {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(<AuditDeclare />) })
  return {
    host,
    unmount() {
      act(() => root.unmount())
      host.remove()
    },
  }
}

function typeInto(el: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!
  act(() => {
    setter.call(el, value)
    el.dispatchEvent(new Event('input', { bubbles: true }))
  })
}

beforeEach(() => {
  useAuditStore.setState({
    site: '', provider: '', link: 'fiber', duration: 60, direction: 'both',
    advanced: true, presets: [], presetNote: '', msg: '', report: null,
  })
})

afterEach(() => {
  document.querySelectorAll('.audit-declare').forEach(n => n.remove())
})

describe('AuditDeclare — champs NFC et bornés', () => {
  it('le site est limité à 80 caractères et recomposé en NFC', () => {
    const m = mount()
    const site = m.host.querySelector<HTMLInputElement>('input[name="site"]')!
    expect(site.maxLength, 'limite de saisie du site').toBe(80)
    typeInto(site, 'Café'.normalize('NFD'))
    expect(useAuditStore.getState().site, 'NFC stocké').toBe('Café')
    m.unmount()
  })

  it('le fournisseur est limité à 80 caractères et recomposé en NFC', () => {
    const m = mount()
    const provider = m.host.querySelector<HTMLInputElement>('input[name="audit-provider"]')!
    expect(provider.maxLength, 'limite de saisie du fournisseur').toBe(80)
    typeInto(provider, 'Télécom'.normalize('NFD'))
    expect(useAuditStore.getState().provider, 'NFC stocké').toBe('Télécom')
    m.unmount()
  })
})
