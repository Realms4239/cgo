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
  it('METEOLINK wordmark exists in App.tsx', () => {
    const app = read('web/frontend/src/App.tsx')
    expect(app).toContain('METEOLINK')
  })
  it('METEOLINK wordmark has glow', () => {
    const s = read('web/frontend/src/components/MeteolinkWordmark.tsx')
    // canonical METEOLINK uses filter drop-shadow cyan + gradient, fallback #f2f2f4 — exhaustive per DESIGN §3
    expect(s).toMatch(/drop-shadow.*rgba\(90,211,227/)
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
    // le blur survit uniquement sur les overlays flottants
    expect(idx).toContain('.panel-chooser')
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
