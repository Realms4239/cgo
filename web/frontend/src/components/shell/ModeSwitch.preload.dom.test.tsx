// T6-FIX — le préchauffage est prouvé par un appel réel, pas par une lecture de
// source : la bascule appelle preloadMode('labo') au survol et au focus, et le
// store fait de même sur setMode. Le mock ne remplace que le transport de chunk.
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useUIStore } from '../../store/ui'
import ModeSwitch from './ModeSwitch'

const preload = vi.hoisted(() => ({ call: vi.fn() }))
vi.mock('../../lib/preload', () => ({ preloadMode: preload.call }))

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

let host: HTMLDivElement
let root: ReturnType<typeof createRoot>

function laboButton(): HTMLButtonElement {
  const button = Array.from(host.querySelectorAll<HTMLButtonElement>('.modeswitch-btn'))
    .find(b => b.getAttribute('data-mode') === 'labo')
  expect(button, 'bouton Labo présent').toBeDefined()
  return button!
}

beforeEach(async () => {
  preload.call.mockClear()
  act(() => { useUIStore.setState({ mode: 'cockpit' }) })
  host = document.createElement('div')
  document.body.appendChild(host)
  root = createRoot(host)
  await act(async () => { root.render(<ModeSwitch />) })
})

afterEach(async () => {
  await act(async () => { root.unmount() })
  host.remove()
  preload.call.mockClear()
})

describe('préchauffage du chunk Labo', () => {
  it('au focus du bouton Labo, appelle réellement preloadMode', async () => {
    await act(async () => { laboButton().focus() })
    expect(preload.call).toHaveBeenCalledWith('labo')
  })

  it('au survol du bouton Labo, appelle réellement preloadMode', async () => {
    await act(async () => {
      laboButton().dispatchEvent(new MouseEvent('mouseover', { bubbles: true }))
    })
    expect(preload.call).toHaveBeenCalledWith('labo')
  })

  it('le store préchauffe aussi sur un changement de mode', () => {
    act(() => { useUIStore.getState().setMode('labo') })
    expect(preload.call).toHaveBeenCalledWith('labo')
  })

  it('le mode courant ne préchauffe rien d’autre', () => {
    expect(useUIStore.getState().mode).toBe('cockpit')
    expect(preload.call).not.toHaveBeenCalled()
  })
})
