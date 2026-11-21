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
  it('expose les quatre profils sans mode recherche', async () => {
    const html = await renderWithResearch(false)
    expect(html).toContain('data-research')
    expect(html).toContain('Mode recherche')
    expect(html).toContain('Profils de campagne')
    expect(html).toContain('P1 — Fibre dédiée')
    expect(html).toContain('P4 — Orbite basse')
    expect(html).not.toContain('Flux de charge')
    expect(html).not.toContain('Canal')
    expect(html).not.toContain('Deadline')
  })

  it('révèle les réglages en mode recherche', async () => {
    const html = await renderWithResearch(true)
    expect(html).toContain('Profils de campagne')
    expect(html).toContain('Répétitions')
    expect(html).toContain('Flux de charge')
    expect(html).toContain('Canal')
    expect(html).toContain('Deadline')
    expect(html).toContain('aria-pressed="true"')
  })
})