import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import App from './App'
import { useUIStore } from './store/ui'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

// EventSource est une API de navigateur, pas le module sous test : le stub ne
// remplace que le transport SSE, jamais App ni le store.
class StubEventSource {
  onmessage: ((ev: MessageEvent) => void) | null = null
  addEventListener() { /* pas d'événements pendant ce test */ }
  close() { /* rien à fermer */ }
}

const setEnv = (globalThis as unknown as { EventSource: typeof EventSource })
const originalEventSource = setEnv.EventSource
const originalFetch = globalThis.fetch
const originalMatchMedia = window.matchMedia

// Ces trois stubs sont des API de navigateur absentes de happy-dom : ils ne
// remplacent ni App, ni le store, ni le composant sous test.
const quietFetch = (async () => ({
  ok: false, status: 404, json: async () => ({}), text: async () => '',
})) as unknown as typeof fetch
const reducedMotion = ((query: string) => ({
  matches: true, media: query, onchange: null,
  addEventListener() {}, removeEventListener() {},
  addListener() {}, removeListener() {}, dispatchEvent: () => false,
})) as unknown as typeof window.matchMedia

beforeEach(() => {
  setEnv.EventSource = StubEventSource as unknown as typeof EventSource
  globalThis.fetch = quietFetch
  window.matchMedia = reducedMotion
  act(() => {
    useUIStore.setState({
      mode: 'cockpit', live: null, connected: false, replayRunning: false,
      benchStage: 'config', proof: null,
    })
  })
})
afterEach(() => {
  setEnv.EventSource = originalEventSource
  globalThis.fetch = originalFetch
  window.matchMedia = originalMatchMedia
  act(() => {
    useUIStore.setState({
      mode: 'cockpit', live: null, connected: false, replayRunning: false,
      benchStage: 'config', proof: null,
    })
  })
})

describe('App — montage des deux vues', () => {
  it('Labo reste différé au premier rendu, puis devient visible à la bascule', async () => {
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    await act(async () => { root.render(<App />) })

    const cockpit = host.querySelector('#v-cockpit')
    expect(cockpit, 'le cockpit est monté').not.toBeNull()
    expect(cockpit!.hasAttribute('hidden'), 'le cockpit est visible').toBe(false)
    expect(host.querySelector('#v-labo'), 'le laboratoire n’est pas monté avant la visite').toBeNull()
    expect(host.querySelector('#v-bench'), 'aucune vue historique').toBeNull()

    const button = Array.from(host.querySelectorAll<HTMLButtonElement>('.modeswitch-btn'))
      .find(b => b.getAttribute('data-mode') === 'labo')
    expect(button, 'le bouton Labo existe').toBeDefined()
    await act(async () => { button!.dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    await act(async () => { /* laisse le chunk différé de Labo se résoudre */ })

    const labo = host.querySelector('#v-labo')
    expect(labo, 'la vue Labo doit être montée').not.toBeNull()
    expect(labo!.hasAttribute('hidden'), 'Labo visible').toBe(false)
    expect(host.querySelector('#v-cockpit')!.hasAttribute('hidden'), 'Cockpit masqué').toBe(true)

    await act(async () => { root.unmount() })
    host.remove()
  })
})
