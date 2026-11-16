// @vitest-environment happy-dom
import { describe, expect, it, vi } from 'vitest'
import { createRoot } from 'react-dom/client'
import { act } from 'react'
import CommandPalette, { paletteItems } from './CommandPalette'

describe('CommandPalette — paletteItems', () => {
  it('une entrée par destination — jamais un doublon', () => {
    const all = paletteItems('')
    expect(all.map(i => i.mode).sort()).toEqual(['cockpit', 'labo'])
  })

  it('fold ignore les accents et la casse — « auditer » trouve Cockpit', () => {
    expect(paletteItems('AUDITER').map(i => i.mode)).toEqual(['cockpit'])
    expect(paletteItems('Banc').map(i => i.mode)).toEqual(['labo'])
  })
})

describe('CommandPalette — ouverture, filtrage, Échap', () => {
  it('ouvre fermée (rendu nul), ouvre sur open=true, ferme sur Échap', () => {
    const onClose = vi.fn()
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    act(() => { root.render(<CommandPalette open={false} onClose={onClose} />) })
    expect(document.getElementById('palette')).toBeNull()
    act(() => { root.render(<CommandPalette open={true} onClose={onClose} />) })
    expect(document.getElementById('palette')).not.toBeNull()
    const input = document.querySelector('.palette-input') as HTMLInputElement
    expect(input).not.toBeNull()
    act(() => { input.focus() })
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })
    act(() => { document.dispatchEvent(event) })
    expect(onClose).toHaveBeenCalled()
  })
})
