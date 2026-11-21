// TDD — trois faits que le document Labo doit tenir : arrêt épinglé,
// absence classée en dernier, duel vide refusé avec des mots.
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LaboView from '../../views/LaboView'
import Comparer, { rankRows, sortRows, type RankRow, type SortKey, type SortDir } from './Comparer'

let uiState: Record<string, unknown>
let auditState: Record<string, unknown>

vi.mock('../../store/ui', () => ({
  useUIStore: (selector: (s: Record<string, unknown>) => unknown) => selector(uiState),
}))
vi.mock('../../store/audit', () => ({
  useAuditStore: (selector: (s: Record<string, unknown>) => unknown) => selector(auditState),
}))

function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}

const MEASURED: RankRow = { id: 'a', label: 'Fibre', delayMs: 64, goodputMbps: 88 }
const ABSENT: RankRow = { id: 'b', label: 'Satellite', delayMs: null, goodputMbps: null }

beforeEach(() => {
  uiState = { live: { running: true }, connected: true, setMode: vi.fn(), openProof: vi.fn() }
  auditState = { report: null, stage: 'observing', start: vi.fn(), stop: vi.fn() }
})

describe('Comparer / document Labo — TDD', () => {
  it('ARRÊTER reste épinglé et visible sans défilement pendant la mesure', () => {
    const host = mount(<LaboView />)
    const header = host.querySelector('header.command-header[data-pinned="true"]')
    expect(header, 'en-tête de commande épinglé').not.toBeNull()
    expect(header!.textContent).toContain('ARRÊTER')
  })

  it('une mesure absente trie en dernier et affiche « — »', () => {
    expect(rankRows([ABSENT, MEASURED]).map(r => r.id)).toEqual(['a', 'b'])
    const host = mount(<Comparer rows={[ABSENT, MEASURED]} />)
    const rows = Array.from(host.querySelectorAll('.rank-row'))
    expect(rows).toHaveLength(2)
    expect(rows[1].textContent).toContain('—')
    expect(rows[1].textContent).toContain('Satellite')
  })

  it('une valeur nulle ou négative fournie en props est traitée comme absente', () => {
    const zero: RankRow = { id: 'z', label: 'Zéro', delayMs: 0, goodputMbps: 0 }
    const negative: RankRow = { id: 'n', label: 'Négatif', delayMs: -5, goodputMbps: 88 }
    expect(rankRows([zero, negative, MEASURED]).map(r => r.id)).toEqual(['a', 'n', 'z'])
    const host = mount(<Comparer rows={[zero, negative, MEASURED]} />)
    const rows = Array.from(host.querySelectorAll('.rank-row'))
    expect(rows[1].textContent).toContain('—')
    expect(rows[2].textContent).toContain('—')
    expect(rows[1].textContent).toContain('non mesuré')
  })

  it('un duel à une mesure vide refuse avec des mots, jamais « égalité »', () => {
    const host = mount(<Comparer rows={[MEASURED]} a={MEASURED} b={null} />)
    expect(host.textContent).toContain('Sélectionnez deux mesures')
    expect(host.textContent).not.toContain('égalité')
  })

  it('classement vide : le vide enseigne ce qui viendra et pourquoi, sans zéro', () => {
    const text = mount(<Comparer rows={[]} />).textContent ?? ''
    expect(text).toContain('Aucune mesure')
    expect(text).toContain('apparaîtra ici')
    expect(text).not.toMatch(/\d/)
  })

  it('duel refusé : le vide nomme ce qui apparaîtra dès deux mesures', () => {
    const duel = mount(<Comparer rows={[]} />).querySelector('.duel-refused')?.textContent ?? ''
    expect(duel).toContain('Sélectionnez deux mesures')
    expect(duel).toContain("s'affichera ici")
  })

  it('un duel mesuré nomme le gagnant, porte un Δ signé et une échelle partagée', () => {
    const slow: RankRow = { id: 's', label: 'Satellite', delayMs: 120, goodputMbps: 20 }
    const duel = mount(<Comparer rows={[MEASURED, slow]} a={MEASURED} b={slow} />)
      .querySelector('.duel[data-state="compared"]')!
    const text = duel.textContent ?? ''
    expect(text).toContain('Fibre')
    expect(text).toContain('plus rapide')
    expect(text).toMatch(/[+−-]\s?\d/)
    expect(duel.querySelector('.duel-scale')).not.toBeNull()
    expect(duel.querySelectorAll('.duel-bar')).toHaveLength(2)
  })

  it('les en-têtes du classement sont triables', () => {
    const host = mount(<Comparer rows={[MEASURED]} />)
    expect(host.querySelectorAll('.rank-table thead th button').length).toBeGreaterThanOrEqual(3)
    expect(host.querySelector('.rank-table thead th[aria-sort]')).not.toBeNull()
  })

  it('un tri sur toute colonne et tout sens garde les non mesurés en dernier', () => {
    const slow: RankRow = { id: 's', label: 'Satellite', delayMs: 120, goodputMbps: 20 }
    const rows = [ABSENT, slow, MEASURED]
    const keys: SortKey[] = ['delay', 'goodput', 'label']
    const dirs: SortDir[] = ['asc', 'desc']
    for (const key of keys) {
      for (const dir of dirs) {
        const ordered = sortRows(rows, key, dir)
        expect(ordered[ordered.length - 1].id, `${key}/${dir}`).toBe('b')
      }
    }
  })

  it('les épingles A/B et PREUVE restent visibles même sans mesure', () => {
    const host = mount(<Comparer rows={[]} />)
    const pins = host.querySelector('.duel-pins')?.textContent ?? ''
    expect(pins).toContain('A')
    expect(pins).toContain('B')
    expect(host.textContent).toContain('PREUVE')
  })
})
