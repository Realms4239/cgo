import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { describe, expect, it } from 'vitest'
import { InlineDisclosure } from './InlineDisclosure'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

describe('InlineDisclosure — détail dans le flux', () => {
  it('bascule aria-expanded et pousse le contenu sans le superposer', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)
    const root = createRoot(el)
    act(() => {
      root.render(<InlineDisclosure label="Détails du calcul"><p>formule complète</p></InlineDisclosure>)
    })

    const trigger = el.querySelector('button')!
    expect(trigger.getAttribute('aria-expanded')).toBe('false')
    expect(el.querySelector('.inline-disclosure-body')).toBeNull()

    act(() => { trigger.dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    expect(trigger.getAttribute('aria-expanded')).toBe('true')
    const body = el.querySelector<HTMLElement>('.inline-disclosure-body')!
    expect(body.textContent).toContain('formule complète')
    expect(body.style.position).not.toBe('absolute')
    expect(body.style.position).not.toBe('fixed')

    act(() => root.unmount())
    el.remove()
  })
})
