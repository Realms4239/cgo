// Le titre adopté est vérifié là où il vit : PRODUCT.md (ignoré par git) et
// la page de titre du livre (arbre de travail). Un fichier absent est sauté,
// jamais remplacé par une chaîne inventée.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const MAIN = 'Optimisation du transport réseau et garantie d’échéances temps réel pour la diffusion d’alertes hydrométéorologiques sur des liens d’accès contraints'
const SUB = 'Régulation du bufferbloat par gestion active de file (AQM), contrôle de congestion BBR et application au système d’information de Météo Madagascar'

const repo = join(process.cwd(), '..', '..')
const read = (p: string): string => readFileSync(join(repo, p), 'utf8')

describe('titre académique adopté', () => {
  it('PRODUCT.md porte le titre et le sous-titre', () => {
    const p = join(repo, 'PRODUCT.md')
    if (!existsSync(p)) return
    const s = read('PRODUCT.md')
    expect(s).toContain(MAIN)
    expect(s).toContain(SUB)
  })

  it('la page de titre du livre porte le même titre', () => {
    const p = join(repo, 'book', 'content', '00-front.js')
    if (!existsSync(p)) return
    const s = read('book/content/00-front.js')
    expect(s).toContain(MAIN)
    expect(s).not.toContain('Architecture d’ordonnancement')
  })
})
