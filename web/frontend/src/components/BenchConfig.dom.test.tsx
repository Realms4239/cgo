import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'

const RESEARCH_KEY = 'meteolink-bench-research'
const renderWithResearch = async (value: boolean) => {
  localStorage.setItem(RESEARCH_KEY, String(value))
  vi.resetModules()
  const { default: Config } = await import('./BenchConfig')
  return renderToStaticMarkup(<Config />)
}

afterEach(() => {
  localStorage.removeItem(RESEARCH_KEY)
  vi.resetModules()
})

describe('BenchConfig', () => {
  it('masque les réglages de campagne hors mode recherche', async () => {
    const html = await renderWithResearch(false)
    expect(html).toContain('data-research')
    expect(html).toContain('Mode recherche')
    expect(html).toContain('Le duel P2')
    expect(html).not.toContain('Profils')
    expect(html).not.toContain('Flux de charge')
    expect(html).not.toContain('Canal')
    expect(html).not.toContain('Deadline')
  })

  it('révèle les réglages en mode recherche', async () => {
    const html = await renderWithResearch(true)
    expect(html).toContain('Profils')
    expect(html).toContain('Répétitions')
    expect(html).toContain('Flux de charge')
    expect(html).toContain('Canal')
    expect(html).toContain('Deadline')
    expect(html).toContain('aria-pressed="true"')
  })
})