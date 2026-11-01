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
  it('tokens.css — canevas plat : plus de halo radial sous la page', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    const m = css.match(/^\s*--canvas:\s*([^;]+);/m)
    expect(m, '--canvas déclaré dans la feuille').not.toBeNull()
    // contrat 20B : une surface unie, aucune couche de dégradé dans le fond
    expect(m![1]).not.toContain('gradient')
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

// Slice 20B — l'échelle se lit à bout de bras, une seule face sans serif.
describe('échelle opérateur', () => {
  const sheet = read('web/frontend/src/styles/tokens.css')
  it('porte l’échelle-body et bannit la serif d’interface', () => {
    for (const [tok, px] of [['--fs-hero', '40px'], ['--fs-metric', '32px'], ['--fs-title', '28px'], ['--fs-body', '16px'], ['--fs-label', '14px'], ['--fs-fine', '13px']]) {
      expect(sheet, tok).toContain(`${tok}: ${px}`)
    }
    expect(sheet).not.toContain('Cormorant')
    expect(sheet).not.toContain('--font-display')
  })
  it('les surfaces de console sont figées', () => {
    expect(sheet).toContain('--canvas-solid: #0a0b0e')
    expect(sheet).toContain('--surface-card: #12141a')
    expect(sheet).toContain('--surface-elevated: #181b22')
  })
  it('une seule face : --font-ui existe et retombe sur Inter via --font-body', () => {
    // Le contrat positif de la face. Sans lui, une coquille sur la déclaration
    // --font-ui renverrait les 12 règles d’interface vers la police du navigateur
    // avec la suite entière au vert : rien ne vérifiait jusqu’ici que le nom
    // résout bien, et vers Inter.
    expect(sheet).toMatch(/--font-ui:\s*var\(--font-body\)/)
    expect(sheet).toMatch(/--font-body:\s*'Inter'/)
  })
})

// Un gris « faible » n'a jamais été mesuré : il valait 4,52:1 sur la surface la
// plus claire de la console, soit la marge d'un cheveu au-dessus du seuil AA. La
// feuille de style calcule maintenant le contraste qu'elle rend, sinon
// l'éclaircissement d'une surface le ferait passer sous le seuil sans qu'aucun
// test ne s'en aperçoive.
describe('contraste des surfaces de console', () => {
  const sheet = read('web/frontend/src/styles/tokens.css')
  const hexToken = (name: string): string => {
    const m = sheet.match(new RegExp(`^\\s*${name}:\\s*(#[0-9a-fA-F]{6});`, 'm'))
    expect(m, `${name} absent de tokens.css ou non littéral`).not.toBeNull()
    return m![1]
  }
  const luminance = (hex: string): number =>
    hex
      .replace('#', '')
      .match(/../g)!
      .map(x => parseInt(x, 16) / 255)
      .map(v => (v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)))
      .reduce((sum, c, i) => sum + [0.2126, 0.7152, 0.0722][i] * c, 0)
  const ratio = (a: string, b: string): number => {
    const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x)
    return (hi + 0.05) / (lo + 0.05)
  }
  const surfaces = ['--canvas-solid', '--surface-card', '--surface-elevated']
  const faint = hexToken('--text-faint')
  const body = hexToken('--text-body')

  for (const surface of surfaces) {
    it(`le gris faible passe AA sur ${surface}`, () => {
      // 13 px / 400 est du texte courant : le seuil est 4,5:1, et la marge est la
      // seule protection contre un éclaircissement futur de la surface.
      expect(ratio(faint, hexToken(surface)), surface).toBeGreaterThanOrEqual(4.5)
    })
  }
  it('la hiérarchie faint < body reste visible sur chaque surface', () => {
    for (const surface of surfaces) {
      const bg = hexToken(surface)
      expect(ratio(faint, bg), surface).toBeLessThan(ratio(body, bg))
    }
  })
})
