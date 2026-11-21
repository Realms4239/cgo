// Porte « bento » retargetée T6 : le gap reste scopé à une racine nommée,
// désormais la racine du document Labo plutôt que le mur supprimé.
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
    expect(comp).toContain('.labo-chapters {')
    expect(comp).toContain('gap: var(--gap)')
    expect(comp, 'le tableau dense ne prend pas le gap bento').not.toContain('.rank-table { gap:')
    const labo = read('views/LaboView.tsx')
    expect(labo).toContain('className="labo-view"')
    expect(labo).toContain('className="labo-chapters"')
  })
})
