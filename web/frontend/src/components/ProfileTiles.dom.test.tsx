// @vitest-environment happy-dom
import { describe, expect, it, vi } from 'vitest'
import { createRoot } from 'react-dom/client'
import { act } from 'react'
import ProfileTiles from './ProfileTiles'

const profiles = [
  { id: 'P1', imported: false },
  { id: 'P2', imported: false },
  { id: 'P3', imported: false },
  { id: 'P4', imported: false },
]

describe('ProfileTiles — clic réel', () => {
  it('quatre tuiles, aria-pressed sélectionné sur P2', () => {
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    act(() => {
      root.render(<ProfileTiles profiles={profiles} selected={['P2']} onToggle={() => undefined} />)
    })
    expect(host.textContent).toContain('Fibre dédiée')
    expect(host.querySelector('[aria-pressed="true"]')?.textContent).toContain('P2')
  })

  it('onToggle est appelé avec le bon id au clic', () => {
    const onToggle = vi.fn()
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    act(() => {
      root.render(<ProfileTiles profiles={profiles} selected={['P2']} onToggle={onToggle} />)
    })
    const p4 = [...host.querySelectorAll('button.profile-tile')].find(b => b.textContent?.includes('P4'))!
    act(() => { (p4 as HTMLButtonElement).click() })
    expect(onToggle).toHaveBeenCalledWith('P4')
  })
})
