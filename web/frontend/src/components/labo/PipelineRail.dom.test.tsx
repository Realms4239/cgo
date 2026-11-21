// TDD — le rail reste un stepper (aria-current="step"), jamais un tablist.
// Le mot de progression et le modèle clavier W3C sont vérifiés sur le DOM rendu.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PipelineRail from './PipelineRail'

function mount(active: number): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<PipelineRail active={active} onSelect={() => {}} />)
  return host
}

describe('PipelineRail — stepper et progression', () => {
  it('le mot de progression est visible et suit l’étape active', () => {
    expect(mount(0).querySelector('.pr-progress')?.textContent).toBe('Étape 1 sur 4')
    expect(mount(2).querySelector('.pr-progress')?.textContent).toBe('Étape 3 sur 4')
    expect(mount(3).querySelector('.pr-progress')?.textContent).toBe('Étape 4 sur 4')
  })

  it('le stepper garde aria-current="step" et refuse la sémantique tablist', () => {
    const host = mount(2)
    const steps = host.querySelectorAll('.pr-step')
    expect(steps).toHaveLength(4)
    expect(host.querySelector('[role="tablist"]')).toBeNull()
    expect(host.querySelector('[role="tab"]')).toBeNull()
    expect(steps[2]?.getAttribute('aria-current')).toBe('step')
    for (const i of [0, 1, 3]) expect(steps[i]?.hasAttribute('aria-current')).toBe(false)
  })

  it('un seul arrêt dans l’ordre de tabulation : le focus entre sur l’étape active', () => {
    const steps = mount(2).querySelectorAll('.pr-step')
    expect(steps[2]?.getAttribute('tabindex')).toBe('0')
    for (const i of [0, 1, 3]) expect(steps[i]?.getAttribute('tabindex')).toBe('-1')
  })

  it('les libellés se replient sans être tronqués ; cibles ≥44 px conservées', () => {
    const sheet = readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')
    expect(sheet).toMatch(/\.pr-label\s*\{[^}]*overflow-wrap\s*:\s*(anywhere|break-word)/)
    expect(sheet).not.toMatch(/\.pr-label\s*\{[^}]*text-overflow\s*:\s*ellipsis/)
    expect(sheet).toMatch(/\.pr-step\s*\{[^}]*min-height\s*:\s*44px/)
  })

  it('l’étape courante se dit par un mot et une forme, jamais par la couleur seule', () => {
    const steps = mount(2).querySelectorAll('.pr-step')
    expect(steps[2]?.textContent).toContain('en cours')
    for (const i of [0, 1, 3]) expect(steps[i]?.textContent).not.toContain('en cours')
    expect(steps[2]?.querySelector('.pr-marker')?.getAttribute('data-current')).toBe('true')
    for (const i of [0, 1, 3]) {
      expect(steps[i]?.querySelector('.pr-marker')?.getAttribute('data-current')).toBe('false')
    }
  })

  it('aucun état « fait / terminé » n’est fabriqué sur les étapes précédentes', () => {
    const text = mount(2).textContent ?? ''
    expect(text).not.toMatch(/terminée?|effectuée?|déjà fait|✓|✔/)
  })
})
