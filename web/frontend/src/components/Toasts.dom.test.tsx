// Les notifications ne parlent pas seulement par la couleur : un mot dit le
// genre, un bouton les ferme, et le survol suspend la fermeture automatique.
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Toasts from './Toasts'
import { useUIStore } from '../store/ui'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const originalMatchMedia = window.matchMedia

function mount() {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(<Toasts />) })
  return {
    host,
    unmount() {
      act(() => root.unmount())
      host.remove()
    },
  }
}

beforeEach(() => {
  vi.useFakeTimers()
  window.matchMedia = ((q: string) => ({
    matches: true, media: q, onchange: null,
    addEventListener() {}, removeEventListener() {},
    addListener() {}, removeListener() {}, dispatchEvent() { return false },
  })) as unknown as typeof window.matchMedia
  useUIStore.setState({ toasts: [] })
})

afterEach(() => {
  vi.clearAllTimers()
  useUIStore.setState({ toasts: [] })
  window.matchMedia = originalMatchMedia
  vi.useRealTimers()
})

describe('Toasts — surface honnête', () => {
  it('chaque notification porte un mot de genre et un bouton de fermeture', () => {
    const s = useUIStore.getState()
    s.pushToast('Audit terminé', 'ok')
    s.pushToast('Échec audit : source absente', 'err')
    s.pushToast('Arrêt demandé', 'blue')
    s.pushToast('Message nu')
    const m = mount()
    const text = m.host.textContent ?? ''
    expect(text).toContain('Succès')
    expect(text).toContain('Échec')
    expect(text).toContain('Info')
    expect(m.host.querySelectorAll('.toast')).toHaveLength(4)
    const dismiss = m.host.querySelectorAll<HTMLButtonElement>('.toast-dismiss')
    expect(dismiss).toHaveLength(4)
    expect(dismiss[0].getAttribute('aria-label'), 'libellé du bouton').toBe('Fermer la notification')
    act(() => { dismiss[0].click() })
    expect(useUIStore.getState().toasts).toHaveLength(3)
    expect(m.host.querySelectorAll('.toast')).toHaveLength(3)
    m.unmount()
  })

  it('le survol suspend la fermeture automatique', () => {
    const m = mount()
    act(() => { useUIStore.getState().pushToast('Campagne terminée', 'ok') })
    expect(m.host.querySelectorAll('.toast')).toHaveLength(1)
    const container = m.host.querySelector<HTMLElement>('#toasts')!
    act(() => {
      container.dispatchEvent(new MouseEvent('mouseover', { bubbles: true, relatedTarget: document.body }))
    })
    act(() => { vi.advanceTimersByTime(5000) })
    expect(m.host.querySelectorAll('.toast'), 'sous le curseur, la notification reste').toHaveLength(1)
    act(() => {
      container.dispatchEvent(new MouseEvent('mouseout', { bubbles: true, relatedTarget: document.body }))
    })
    act(() => { vi.advanceTimersByTime(1400) })
    expect(m.host.querySelectorAll('.toast'), 'après le survol, la notification part').toHaveLength(0)
    m.unmount()
  })
})
