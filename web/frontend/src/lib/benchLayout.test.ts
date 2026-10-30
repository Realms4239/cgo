// Portes de structure du banc : trois sections linéaires, aucun sous-onglet,
// aucun panneau fixe de 380 px. Elles lisent les sources réelles.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const SRC = join(process.cwd(), 'src')
const read = (p: string): string => readFileSync(join(SRC, p), 'utf8')

describe('banc d’essai linéaire', () => {
  it('trois sections nommées, aucune taxonomie imbriquée', () => {
    const b = read('views/BenchView.tsx')
    for (const id of ['bench-config', 'bench-observe', 'bench-results']) {
      expect(b, `section ${id}`).toContain(`id="${id}"`)
    }
    expect(b).toContain('<BenchConfig />')
    expect(b).toContain('<BenchProgress active={active} />')
    expect(b).toContain('<LiveView active={active} />')
    expect(b).toContain('<ResultatsView embedded active={active} />')
    expect(b, 'aucun sous-onglet').not.toContain('role="tablist"')
  })

  it('Configurer garde chaque décision de campagne', () => {
    const c = read('components/BenchConfig.tsx')
    for (const token of ['Profils', 'Flux de charge', 'Canal', 'Deadline', 'DÉMARRER', "ARRÊTER", 'AUDIT → PROFIL', 'IMPORTER']) {
      expect(c, token).toContain(token)
    }
    expect(read('store/bench.ts'), 'quatre flux par défaut').toContain('flows: 4')
  })

  it('Observer garde les portes et la matrice', () => {
    const p = read('components/BenchProgress.tsx')
    expect(p).toContain('G0–G7')
    expect(p).toContain('Matrice — état par cellule')
    expect(p).toContain('data-testid="matrix-table"')
  })

  it('l’ancien cockpit fixe a disparu', () => {
    expect(existsSync(join(SRC, 'views', 'CampagneView.tsx'))).toBe(false)
    const css = read('styles/index.css')
    expect(css).not.toContain('.cockpit')
  })

  it('l’affichage vit dans Observer, pas dans le chrome global', () => {
    const b = read('views/BenchView.tsx')
    expect(b).toContain('<DisplayControls />')
    const app = read('App.tsx')
    expect(app).not.toContain('DisplayControls')
  })

  it('hors destination, le mur ne se re-rend plus au rythme du SSE', () => {
    expect(read('views/BenchView.tsx')).toContain('active={active}')
    expect(read('views/LiveView.tsx')).toContain('rafActive && active')
    expect(read('views/LiveView.tsx')).toContain('active ? s.live : null')
    expect(read('views/ResultatsView.tsx')).toContain('active && !!s.live?.running')
    expect(read('views/IntegriteView.tsx')).toContain('if (!active) return')
  })

  it('le banc se charge à la demande, sans ECharts dans le premier rendu', () => {
    const app = read('App.tsx')
    expect(app).toContain("lazy(() => import('./views/BenchView'))")
    expect(app).toContain('<Suspense')
    const rail = read('components/ModeRail.tsx')
    expect(rail).toContain('preloadBench')
  })
})
