import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import { ProfileTile } from './ProfileTile'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const tile = { title: 'P3 — VSAT géostationnaire', spec: '5 Mbit/s · 600 ms RTT', meta: '1 % perte · Satellite' }

function host(selected: boolean): HTMLElement {
  const el = document.createElement('div')
  el.innerHTML = renderToStaticMarkup(<ProfileTile tile={tile} selected={selected} onToggle={() => {}} />)
  return el
}

describe('ProfileTile — nom entier, état dit en mot', () => {
  it('ne tronque jamais le nom et l’affiche en entier', () => {
    const el = host(true)
    const name = el.querySelector<HTMLElement>('.profile-tile-title')!
    expect(name.textContent).toBe(tile.title)
    expect(name.style.textOverflow).not.toBe('ellipsis')
    expect(el.textContent).toContain('5 Mbit/s · 600 ms RTT')
  })

  it('sélectionné = bordure + coche + mot « sélectionné »', () => {
    const el = host(true)
    expect(el.querySelector('[aria-pressed="true"]')).not.toBeNull()
    expect(el.textContent).toContain('sélectionné')
    expect(el.textContent).toContain('✓')
  })

  it('non sélectionné ne dit pas « sélectionné » et garde une forme', () => {
    const el = host(false)
    expect(el.querySelector('[aria-pressed="false"]')).not.toBeNull()
    expect(el.textContent).not.toContain('sélectionné')
    expect(el.textContent).toContain('○')
  })

  it('appelle onToggle au clic', () => {
    const onToggle = vi.fn()
    const el = document.createElement('div')
    document.body.appendChild(el)
    const root = createRoot(el)
    act(() => { root.render(<ProfileTile tile={tile} selected={false} onToggle={onToggle} />) })
    act(() => { el.querySelector('button')!.dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    expect(onToggle).toHaveBeenCalledOnce()
    act(() => root.unmount())
    el.remove()
  })
})
