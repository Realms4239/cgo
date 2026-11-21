import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ModeSwitch from './ModeSwitch'
import { useUIStore } from '../../store/ui'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const realSetMode = useUIStore.getState().setMode

beforeEach(() => {
  useUIStore.setState({ mode: 'cockpit', setMode: realSetMode })
})
afterEach(() => {
  useUIStore.setState({ mode: 'cockpit', setMode: realSetMode })
})

describe('ModeSwitch', () => {
  it('expose exactement deux modes nommés, avec un seul courant', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<ModeSwitch />)

    const buttons = Array.from(host.querySelectorAll('.modeswitch-btn'))
    expect(buttons).toHaveLength(2)
    expect(Array.from(host.querySelectorAll('.modeswitch-lbl')).map(n => n.textContent))
      .toEqual(['Synthèse', 'Labo'])
    expect(buttons[0].getAttribute('aria-current')).toBe('page')
    expect(buttons[1].hasAttribute('aria-current')).toBe(false)
  })

  it('donne à chaque mode une cible d’au moins 44 px', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<ModeSwitch />)

    for (const button of host.querySelectorAll<HTMLElement>('.modeswitch-btn')) {
      const height = getComputedStyle(button).minHeight || button.style.minHeight
      expect(parseInt(height, 10), button.textContent || '').toBeGreaterThanOrEqual(44)
    }
  })

  it('demande le mode labo au clic sur Labo', () => {
    const setMode = vi.fn()
    useUIStore.setState({ setMode })
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    act(() => { root.render(<ModeSwitch />) })

    const buttons = host.querySelectorAll('button')
    act(() => { buttons[1].dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    expect(setMode).toHaveBeenCalledWith('labo')

    act(() => { root.unmount() })
    host.remove()
  })
})
