// Le titre adopté est vérifié là où il vit : PRODUCT.md (ignoré par git) et
// la page de titre du livre (arbre de travail). Un fichier absent est sauté,
// jamais remplacé par une chaîne inventée.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

// Title 2 (verrouillé 2026-11-14, aligné ISGEI MSTI-STR) :
// « Optimisation temps réel de la qualité de service des liens d'accès contraints :
//   une architecture d'observabilité eBPF et de régulation AQM/BBR à preuves
//   scellées pour les flux critiques de Météo Madagascar »
const TITLE2 = 'Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar'

const repo = join(process.cwd(), '..', '..')
const read = (p: string): string => readFileSync(join(repo, p), 'utf8')

describe('titre académique adopté', () => {
  it('PRODUCT.md porte le titre verrouillé', () => {
    const p = join(repo, 'PRODUCT.md')
    if (!existsSync(p)) return
    const s = read('PRODUCT.md')
    expect(s).toContain(TITLE2)
    expect(s).not.toContain('Conception et évaluation')
  })

  it('la page de titre du livre porte le même titre', () => {
    const p = join(repo, 'book', 'content', '00-front.js')
    if (!existsSync(p)) return
    const s = read('book/content/00-front.js')
    expect(s).toContain(TITLE2)
    expect(s).not.toContain('garantie d’échéances')
    expect(s).not.toContain('Architecture d’ordonnancement')
    expect(s).not.toContain('Conception et évaluation')
  })
})
