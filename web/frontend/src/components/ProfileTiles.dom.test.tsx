import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import ProfileTiles from './ProfileTiles'

const profiles = [
  { id: 'P1', imported: false },
  { id: 'P2', imported: false },
  { id: 'P3', imported: false },
  { id: 'P4', imported: false },
]

describe('ProfileTiles', () => {
  it('shows full profile names and never clips them', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(
      <ProfileTiles profiles={profiles} selected={['P2']} onToggle={() => undefined} />,
    )

    expect(host.textContent).toContain('Fibre dédiée')
    expect(host.textContent).toContain('20 ms RTT')
    expect(host.textContent).toContain('4G dégradée')
    expect(host.textContent).toContain('Orbite basse')
    expect(host.querySelector('[aria-pressed="true"]')?.textContent).toContain('P2')
  })
})
