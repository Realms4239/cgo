// Oracle statique : les specs E2E ne doivent pas viser les vues supprimées par
// la refonte deux destinations (Synthèse / Labo). Le DOM test App.mounting
// prouve déjà que ces identifiants n'existent plus ; ici on empêche qu'un spec
// E2E continue de les chercher et échoue éternellement contre l'UI réelle.
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const E2E_DIR = join(process.cwd(), 'e2e')
const SPEC_FILES = readdirSync(E2E_DIR).filter(f => f.endsWith('.spec.ts'))
const REMOVED = ['v-dashboard', 'v-audit', 'v-bench', 'v-evidence', 'Tableau de bord', 'Preuves & Rapports']

describe('E2E contract — aucune vue historique', () => {
  it('les specs existent et sont découvrables', () => {
    expect(SPEC_FILES.length).toBeGreaterThan(0)
  })
  for (const spec of SPEC_FILES) {
    it(`${spec} ne cible que les vues vivantes`, () => {
      const text = readFileSync(join(E2E_DIR, spec), 'utf8')
      for (const token of REMOVED) {
        expect(text, `${spec} contient le jeton supprimé ${token}`).not.toContain(token)
      }
    })
  }
})
