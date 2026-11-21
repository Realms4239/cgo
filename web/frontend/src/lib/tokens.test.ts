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
  it('METEOLINK wordmark is rendered by Chrome through its component', () => {
    const chrome = read('web/frontend/src/components/shell/Chrome.tsx')
    expect(chrome).toContain('MeteolinkWordmark')
    const wordmark = read('web/frontend/src/components/MeteolinkWordmark.tsx')
    expect(wordmark).toContain('METEOLINK')
  })
  it('METEOLINK wordmark has glow', () => {
    const s = read('web/frontend/src/components/MeteolinkWordmark.tsx')
    // canonical METEOLINK uses filter drop-shadow teal + gradient, fallback #333333 — exhaustive per DESIGN §3
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
    expect(css).toMatch(/linear-gradient.*transparent.*#e8e8e8/)
  })
  it('index.css has fadeIn and card hover and nav-btn inset', () => {
    const css = read('web/frontend/src/styles/index.css')
    expect(css).toContain('@keyframes fadeIn')
    expect(css).toContain('.card:hover')
    expect(css).toContain('box-shadow: inset')
  })
})

describe('mobile 390 — navigation', () => {
  it('sous 899px, la navigation reste dans l’en-tête sans indices clavier', () => {
    const css = read('web/frontend/src/styles/index.css')
    const m = css.match(/@media \(max-width: 899px\) \{[\s\S]*?\n\}/)
    expect(m, 'bloc media max-width:899px présent').not.toBeNull()
    expect(m![0]).toMatch(/\.modeswitch-key \{[^}]*display: none/)
  })

  it('sous 700px, les quatre destinations passent en barre basse', () => {
    const css = read('web/frontend/src/styles/index.css')
    const m = css.match(/@media \(max-width: 700px\) \{[\s\S]*?\n\}/)
    expect(m, 'bloc media max-width:700px présent').not.toBeNull()
    expect(m![0]).toMatch(/\.modeswitch \{[^}]*position: fixed/)
    expect(m![0]).toMatch(/\.modeswitch-btn \{[^}]*min-height: 48px/)
  })
})

describe('desktop grid 890–1920', () => {
  const index = read('web/frontend/src/styles/index.css')
  const tokens = read('web/frontend/src/styles/tokens.css')
  it('declares the shared 12-column grid and wrap contract', () => {
    expect(tokens).toContain('--grid-gap:')
    expect(index).toContain('.layout-grid')
    expect(index).toContain('flex-wrap: wrap')
  })
  it('keeps the operator canvas dark but no longer uses the old narrow panel max-width', () => {
    expect(index).toContain('.layout-shell')
    expect(index).not.toContain('.panel-stack { display: flex; flex-direction: column; gap: var(--sp-4); max-width: 1180px;')
  })
})

describe('échelle opérateur', () => {
  const sheet = read('web/frontend/src/styles/tokens.css')
  it('porte l’échelle-body et bannit la serif d’interface', () => {
    for (const [tok, px] of [['--fs-hero', '40px'], ['--fs-metric', '32px'], ['--fs-title', '28px'], ['--fs-body', '16px'], ['--fs-label', '14px'], ['--fs-fine', '13px']]) {
      expect(sheet, tok).toContain(`${tok}: ${px}`)
    }
    expect(sheet).not.toContain('Cormorant')
    expect(sheet).not.toContain('--font-display')
  })
  it('les surfaces de console sont figées (papier clair ratifié)', () => {
    expect(sheet).toContain('--canvas-solid: #fbfbfb')
    expect(sheet).toContain('--surface-card: #ffffff')
    expect(sheet).toContain('--surface-elevated: #ffffff')
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

// T9 — palette « papier clair » ratifiée (plan 2026-11-20, contrainte globale
// ligne 25). Chaque valeur est verrouillée : le test échoue en rouge contre
// l'ancienne feuille sombre, puis doit rester vert sans assouplir l'AA.
describe('palette claire ratifiée — valeurs verrouillées', () => {
  const sheet = read('web/frontend/src/styles/tokens.css')
  const token = (name: string): string => {
    const m = sheet.match(new RegExp(`^\\s*${name}:\\s*(#[0-9a-fA-F]{6})\\s*;`, 'm'))
    expect(m, `${name} absent de tokens.css ou non littéral hex`).not.toBeNull()
    return m![1].toLowerCase()
  }

  it('mappe chaque valeur du plan sur son jeton existant', () => {
    const map: Record<string, string> = {
      // surfaces
      '--canvas-solid': '#fbfbfb',
      '--surface-card': '#ffffff',
      '--surface-soft': '#f3f3f3',
      '--hairline': '#e8e8e8',
      '--hairline-strong': '#c8c8c8',
      // texte
      '--text-primary': '#333333',
      '--text-body': '#5a5a5a',
      '--text-muted': '#6d6d6d',
      // identité (teal / emerald / gold / slate)
      '--brand-1': '#17938a',
      '--brand-2': '#2dcc70',
      '--brand-3': '#f4c910',
      '--brand-teal-ink': '#0f6b62',
      '--brand-slate': '#34495e',
      // sémantique
      '--t-live': '#17938a',
      '--t-ok': '#1e7a3c',
      '--t-ok-fill': '#229b4e',
      '--t-threshold': '#806400',
      '--t-warn-mark': '#9a7400',
      '--t-danger': '#d51600',
      '--t-neutral': '#6d6d6d',
    }
    for (const [name, hex] of Object.entries(map)) expect(token(name), name).toBe(hex)
  })

  it('ne laisse aucune valeur de la palette sombre dans la feuille', () => {
    for (const dark of ['#0a0b0e', '#0e1015', '#12141a', '#181b22', '#26262a', '#3a3a40', '#f2f2f4', '#5ad3e3', '#e22718', '#1fa348', '#f4b400']) {
      expect(sheet, dark).not.toContain(dark)
    }
  })
})
