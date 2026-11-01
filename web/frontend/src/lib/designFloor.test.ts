// Portes mesurables de la passe système (T9). Elles lisent les sources réelles,
// pas un composant monté : une dérive de style doit rougir sans navigateur.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync, readdirSync } from 'node:fs'
import { join, relative, sep } from 'node:path'
import { describe, it, expect } from 'vitest'

const SRC = join(process.cwd(), 'src')

function walk(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    const p = join(dir, e.name)
    return e.isDirectory() ? walk(p) : [p]
  })
}
function read(p: string): string { return readFileSync(p, 'utf8') }
function rel(p: string): string { return relative(process.cwd(), p).split(sep).join('/') }
function offending(files: string[], pick: (src: string, file: string) => string[]): string[] {
  return files.flatMap((f) => pick(read(f), rel(f)))
}
function lines(src: string, re: RegExp): string[] {
  return src.split('\n').flatMap((line, i) => (re.test(line) ? [`${i + 1}: ${line.trim()}`] : []))
}

const TSX = walk(SRC).filter((f) => f.endsWith('.tsx') && !f.endsWith('.test.tsx'))
const CSS = walk(SRC).filter((f) => f.endsWith('.css'))
// Tout le code source, hors tests : les gardes d’arbre marchent src/ en entier
// (css, ts, tsx) pour qu’aucune feuille ou composant isolé ne serve d’angle mort.
const SRC_FILES = walk(SRC).filter((f) => /\.(tsx?|css|jsx?|mjs)$/.test(f) && !/\.test\./.test(f))

describe('design floor — portes source', () => {
  it('aucun hex dans les .tsx (3, 6 ou 8 chiffres)', () => {
    const bad = offending(TSX, (s, f) => lines(s, /#[0-9a-fA-F]{3,8}\b/).map((l) => `${f}:${l}`))
    expect(bad, 'les couleurs passent par var(--token) ou palette.ts').toEqual([])
  })

  it('aucune taille de police numérique — les six étages seulement', () => {
    const bad = offending(TSX, (s, f) => lines(s, /fontSize:\s*\d/).map((l) => `${f}:${l}`))
    expect(bad, 'utiliser var(--fs-hero|metric|title|body|label|fine)').toEqual([])
  })

  it('aucun letterSpacing littéral en em', () => {
    const bad = offending(TSX, (s, f) => lines(s, /letterSpacing:\s*['"][0-9.-]+em['"]/).map((l) => `${f}:${l}`))
    expect(bad, 'utiliser var(--tt-06|08|10|neg-02)').toEqual([])
  })

  it('aucun idiome « 6px 8px » ni taille < 10 px', () => {
    const bad = offending(TSX, (s, f) => {
      const a = lines(s, /6px 8px/)
      const b = lines(s, /fontSize:\s*['"]?[0-9](?:px)?['"]?[,\s}]/)
      return [...a, ...b].map((l) => `${f}:${l}`)
    })
    expect(bad, 'espacement = var(--sp-*), tailles = var(--fs-hero|metric|title|body|label|fine), rien sous --fs-fine').toEqual([])
  })

  it('la serif d’interface est sortie de tout l’arbre src/ (hors tests)', () => {
    // Complète le garde de tokens.test.ts, qui ne lit que tokens.css : ici on
    // marche src/ comme les portes ci-dessus, donc un var(--font-display)
    // réintroduit dans index.css ou un littéral serif dans un .tsx rougit. Le
    // descripteur font-display: swap des @font-face n’est PAS le jeton retiré et
    // ne doit pas être signalé — le confondre rendrait cette garde inutile.
    const bad = offending(SRC_FILES, (s, f) =>
      lines(s, /Cormorant|var\(--font-display\)/).map((l) => `${f}:${l}`))
    expect(bad, 'Inter est la seule face ; --font-display et la serif sont retirés').toEqual([])
  })

  it('tout var(--fs-*) consommé sous src/ est déclaré dans tokens.css', () => {
    // Les alias de transition empêchent 47 règles de perdre leur font-size ; cette
    // protection ne peut pas tenir à un seul grep ponctuel. Si une déclaration
    // change de nom ou disparaît, la règle remonte la cascade et retombe sur la
    // taille héritée — le garde doit alors rougir.
    const sheet = read(join(SRC, 'styles', 'tokens.css'))
    const declared = new Set(
      [...sheet.matchAll(/(^|[\s;])(--fs-[a-z]+)\s*:/gm)].map((m) => m[2]),
    )
    const used = new Set(
      SRC_FILES.flatMap((f) =>
        [...read(f).matchAll(/var\(\s*(--fs-[a-z]+)\s*\)/g)].map((m) => m[1])),
    )
    const missing = [...used].filter((n) => !declared.has(n)).sort()
    expect(missing, 'taille utilisée sans déclaration dans tokens.css').toEqual([])
  })

  it('aucun alias de compatibilité consommé dans les .tsx', () => {
    // Ruling HM-1 : le bloc d'alias reste une couche de compatibilité pour
    // styles/*.css ; les composants parlent les jetons canoniques.
    const alias = /var\(--(?:tw-|tt-|tl-|tx-|radius-|danger|warning|success|series-|accent|ease-standard|fs-base|fs-num|surface-raised|state-)/
    const bad = offending(TSX, (s, f) => lines(s, alias).map((l) => `${f}:${l}`))
    expect(bad, 'les .tsx utilisent var(--token) canonique').toEqual([])
  })

  it('chaque primitive ui/*.tsx a au moins un importeur hors test', () => {
    const ui = walk(join(SRC, 'components', 'ui')).filter((f) => f.endsWith('.tsx') && !f.endsWith('.test.tsx'))
    const consumers = walk(SRC).filter((f) => /\.(ts|tsx)$/.test(f) && !/\.test\.tsx?$/.test(f))
    const dead = ui.filter((f) => {
      const name = f.slice(f.lastIndexOf(sep) + 1).replace(/\.tsx$/, '')
      return !consumers.some((c) => c !== f && read(c).includes(`ui/${name}`))
    }).map(rel)
    expect(dead, 'adopter ou supprimer la primitive').toEqual([])
  })

  it('aucune transition de largeur (layout thrash)', () => {
    const bad = [
      ...offending(TSX, (s, f) => lines(s, /transition:\s*['"][^'"]*\bwidth\b/).map((l) => `${f}:${l}`)),
      ...offending(CSS, (s, f) => lines(s, /transition:\s*[^;]*\bwidth\b/).map((l) => `${f}:${l}`)),
    ]
    expect(bad, 'transform: scaleX() + transform-origin').toEqual([])
  })

  it('aucune durée littérale dans les transitions .tsx — --dur-* seulement', () => {
    const bad = offending(TSX, (s, f) =>
      lines(s, /transition:\s*['"][^'"]*\d+(?:\.\d+)?m?s\b/).map((l) => `${f}:${l}`))
    expect(bad, 'utiliser var(--dur-fast|base|slow)').toEqual([])
  })

  it('reduced-motion collapse les durées et le focus-visible reste global', () => {
    const css = read(join(SRC, 'styles', 'tokens.css'))
    expect(css).toMatch(/@media \(prefers-reduced-motion: reduce\)/)
    expect(css).toMatch(/--dur-slow:\s*1ms/)
    expect(read(join(SRC, 'styles', 'index.css'))).toMatch(/^:focus-visible\s*\{/m)
  })

  it('résultats : en-têtes triables avec aria-sort, puces de sélection avec aria-pressed', () => {
    const src = read(join(SRC, 'views', 'ResultatsView.tsx'))
    expect(src, 'aria-sort sur la colonne triée').toContain('aria-sort=')
    expect(src, 'les puces de rang/filtre sont des bascules').toContain('aria-pressed={active}')
  })

  it('le rail dérive de MODES et chaque destination est une bascule accessible', () => {
    const rail = read(join(SRC, 'components', 'ModeRail.tsx'))
    expect(rail, 'importe MODES').toContain("from '../store/ui'")
    expect(rail, 'importe MODES').toContain('MODES')
    expect(rail, 'boucle sur MODES').toContain('MODES.map')
    expect(rail, 'expose aria-pressed').toContain('aria-pressed={mode === m.id}')
    const app = read(join(SRC, 'App.tsx'))
    expect(app, 'raccourcis 1–4 depuis MODES').toContain('MODES.find(m => m.key === e.key)')
  })
})
