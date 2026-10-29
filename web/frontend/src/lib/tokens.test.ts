// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* try next */ }
  }
  // la dernière tente laisse l'erreur remonter avec le chemin d'origine
  return readFileSync(p, 'utf8')
}

describe('tokens', () => {
  it('METEOLINK wordmark is rendered by App through its component', () => {
    const app = read('web/frontend/src/App.tsx')
    expect(app).toContain('MeteolinkWordmark')
    const wordmark = read('web/frontend/src/components/MeteolinkWordmark.tsx')
    expect(wordmark).toContain('METEOLINK')
  })
  it('METEOLINK wordmark has glow', () => {
    const s = read('web/frontend/src/components/MeteolinkWordmark.tsx')
    // canonical METEOLINK uses filter drop-shadow cyan + gradient, fallback #f2f2f4 — exhaustive per DESIGN §3
    expect(s).toMatch(/drop-shadow.*var\(--t-live\)/)
    expect(s).toContain('linear-gradient')
  })
  it('tokens.css has radial glow', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('radial-gradient')
  })
  it('tokens.css has noise overlay', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('noise.png')
  })
  it('tokens.css — cartes opaques : le blur vit seulement dans les overlays (index.css)', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('linear-gradient')
    // plus de backdrop-filter sur .card — la cause n°1 de jank a été retirée
    expect(css).not.toContain('.card')
    const idx = read('web/frontend/src/styles/index.css')
    // aucune carte ne porte de backdrop-filter (le blur reste aux overlays flottants)
    expect(idx).not.toMatch(/\.card\s*\{[^}]*backdrop-filter/)
  })
  it('tokens.css has hairline fading', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    // le dégradé hairline est un linear-gradient à bords transparents
    expect(css).toMatch(/linear-gradient.*transparent.*#26262a/)
  })
  it('index.css has fadeIn and card hover and nav-btn inset', () => {
    const css = read('web/frontend/src/styles/index.css')
    expect(css).toContain('@keyframes fadeIn')
    expect(css).toContain('.card:hover')
    expect(css).toContain('box-shadow: inset')
  })
})

describe('mobile 390 — navigation', () => {
  it('le rail des modes devient horizontal sous 899px', () => {
    const css = read('web/frontend/src/styles/index.css')
    const m = css.match(/@media \(max-width: 899px\) \{[\s\S]*?\n\}/)
    expect(m, 'bloc media max-width:899px présent').not.toBeNull()
    expect(m![0]).toMatch(/\.mode-rail \{[^}]*flex-direction: row/)
  })

  it('sous 700px, les quatre destinations passent en barre basse', () => {
    const css = read('web/frontend/src/styles/index.css')
    const m = css.match(/@media \(max-width: 700px\) \{[\s\S]*?\n\}/)
    expect(m, 'bloc media max-width:700px présent').not.toBeNull()
    expect(m![0]).toMatch(/\.mode-rail \{[^}]*position: fixed/)
  })
})
