// Le titre adopté est vérifié là où il vit : PRODUCT.md (ignoré par git) et
// la page de titre du livre (arbre de travail). Un fichier absent est sauté,
// jamais remplacé par une chaîne inventée.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const MAIN = 'Conception et évaluation d’un banc d’essai reproductible pour la comparaison de politiques AQM et BBR sur des liens d’accès contraints'
const SUB = 'Le cas des liens d’accès du service national de prévision météorologique de Madagascar'

const repo = join(process.cwd(), '..', '..')
const read = (p: string): string => readFileSync(join(repo, p), 'utf8')

describe('titre académique adopté', () => {
  it('PRODUCT.md porte le titre et le sous-titre', () => {
    const p = join(repo, 'PRODUCT.md')
    if (!existsSync(p)) return
    const s = read('PRODUCT.md')
    expect(s).toContain(MAIN)
    expect(s).toContain(SUB)
    expect(s).not.toContain('garantie d’échéances')
  })

  it('la page de titre du livre porte le même titre', () => {
    const p = join(repo, 'book', 'content', '00-front.js')
    if (!existsSync(p)) return
    const s = read('book/content/00-front.js')
    expect(s).toContain(MAIN)
    expect(s).not.toContain('garantie d’échéances')
    expect(s).not.toContain('Architecture d’ordonnancement')
  })
})
