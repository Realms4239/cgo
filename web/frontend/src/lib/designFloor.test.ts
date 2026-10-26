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

describe('design floor — portes source', () => {
  it('aucun hex dans les .tsx (3, 6 ou 8 chiffres)', () => {
    const bad = offending(TSX, (s, f) => lines(s, /#[0-9a-fA-F]{3,8}\b/).map((l) => `${f}:${l}`))
    expect(bad, 'les couleurs passent par var(--token) ou palette.ts').toEqual([])
  })

  it('aucune taille de police numérique — les 5 jetons seulement', () => {
    const bad = offending(TSX, (s, f) => lines(s, /fontSize:\s*\d/).map((l) => `${f}:${l}`))
    expect(bad, 'utiliser var(--fs-micro|tag|cap|body|head)').toEqual([])
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
    expect(bad, 'espacement = var(--sp-*), rien sous --fs-micro').toEqual([])
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

  it('nav de tête et rail dérivent de la même source PANELS', () => {
    const chooser = read(join(SRC, 'components', 'PanelChooser.tsx'))
    const rail = read(join(SRC, 'components', 'Rail.tsx'))
    for (const [name, src] of [['PanelChooser', chooser], ['Rail', rail]] as const) {
      expect(src, `${name} importe PANELS`).toContain("PANELS } from '../store/ui'")
      expect(src, `${name} boucle sur PANELS`).toContain('PANELS.map')
      expect(src, `${name} expose aria-pressed`).toContain('aria-pressed=')
    }
    expect(chooser).toContain('const active = panel === p.id')
    expect(rail).toContain('aria-pressed={panel===p.id}')
  })
})
