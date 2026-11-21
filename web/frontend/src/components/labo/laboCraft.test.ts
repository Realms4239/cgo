// TDD — plancher de craft du Labo lu dans les sources : mot + forme sur le rail,
// respiration de la tête de chapitre, cycle d'état des actions, une seule motion.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const sheet = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')

describe('Labo — plancher de craft', () => {
  it('le rail porte une marque de forme et un mot pour l’étape courante', () => {
    const s = sheet()
    expect(s).toMatch(/\.pr-marker\[data-current='true'\]/)
    expect(s).toMatch(/\.pr-marker\[data-current='false'\]/)
    expect(s).toMatch(/\.pr-state\s*\{/)
  })

  it('la tête de chapitre respire au-dessus du titre, pas en dessous, sans kicker', () => {
    const s = sheet()
    expect(s).toMatch(/\.chapter-head\s*\{[^}]*padding-top:/)
    expect(s).toMatch(/\.chapter-head\s*\{[^}]*margin-bottom:\s*var\(--sp-[23]\)/)
    expect(s).not.toMatch(/\.chapter-title\s*\{[^}]*margin-bottom/)
    const chapter = readFileSync(join(process.cwd(), 'src/components/labo/Chapter.tsx'), 'utf8')
    expect(chapter).not.toMatch(/kicker/i)
  })

  it('cycle d’état complet des actions de panneau en jetons ratifiés', () => {
    const s = sheet()
    // Le focus ne porte plus d'override Labo : la règle globale :focus-visible
    // (2px) est vérifiée par calcul de style dans redesign-feel.spec.ts.
    expect(s).toMatch(/\.labo-view \.btn:disabled/)
    expect(s).toMatch(/--state-disabled-opacity/)
    expect(s).toMatch(/--state-disabled-bg/)
  })

  // Plancher de craft, pas preuve d'absence d'animation : un total ≤5 ne
  // détecte qu'un dépassement grossier, pas une transition ajoutée isolément.
  it('plancher : pas de @keyframes et ≤5 transitions (majoré, non preuve)', () => {
    const s = sheet()
    expect(s).not.toMatch(/@keyframes/)
    expect((s.match(/transition:/g) ?? []).length).toBeLessThanOrEqual(5)
  })
})
