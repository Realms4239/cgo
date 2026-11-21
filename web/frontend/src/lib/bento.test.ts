// Porte « bento » retargetée à l'amendement §B : le gap reste scopé à une
// racine de grille nommée (#wall), jamais aux tableaux ; le Labo est désormais
// une piste de panneaux translateX sans gap entre eux.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('bento', () => {
  it('gap var(--gap) uniquement sur les racines de grille, jamais sur les tableaux', () => {
    const css = read('styles/index.css')
    expect(css).toContain('#wall')
    expect(css).toContain('gap: var(--gap)')
    const comp = read('styles/components.css')
    expect(comp, 'le tableau dense ne prend pas le gap bento').not.toContain('.rank-table { gap:')
    expect(comp).toContain('.panel-track {')
    expect(comp, 'glissement en 180 ms sur transform seul').toContain('transition: transform 180ms ease-out')
    const labo = read('views/LaboView.tsx')
    expect(labo).toContain('className="labo-view"')
    expect(labo).toContain('className="panel-track"')
    expect(labo, 'translation d’un panneau entier').toContain('translateX(-${active * 100}%)')
  })
})
