// L'instrument T5 doit être visible dans le vrai Cockpit, pas seulement isolé
// dans son test de composant. Ces assertions lisent le DOM rendu de CockpitView.
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CockpitView from '../../views/CockpitView'
import { objectiveMarkLine } from '../../lib/chartGrammar'
import { instrument } from '../../lib/ringBuffer'

let auditState: Record<string, unknown>
let uiState: Record<string, unknown>

vi.mock('../../store/audit', () => ({
  useAuditStore: (selector: (s: Record<string, unknown>) => unknown) => selector(auditState),
}))
vi.mock('../../store/ui', () => ({
  useUIStore: (selector: (s: Record<string, unknown>) => unknown) => selector(uiState),
}))

function html(): string {
  return renderToStaticMarkup(createElement(CockpitView) as never)
    .replace(/&#x([0-9a-f]+);/gi, (_, c) => String.fromCodePoint(parseInt(c, 16)))
    .replace(/&#(\d+);/g, (_, c) => String.fromCodePoint(Number(c)))
}

beforeEach(() => {
  auditState = { report: null, stage: 'idle', failure: null, open: vi.fn(), start: vi.fn(), stop: vi.fn() }
  uiState = { setMode: vi.fn(), live: null }
  instrument.events.clear()
  instrument.small.clear()
})

describe('CockpitView — instrument T5 monté', () => {
  it('monte le champ à axe temporel partagé', () => {
    const out = html()
    expect(out).toContain('aria-label="Champ temporel"')
    expect(out).toContain('Règle temporelle partagée')
    expect(out).toContain('Trace primaire, débit utile et interface')
  })

  it('affiche la vraie étiquette de la ligne 220 ms de objectiveMarkLine', () => {
    const out = html()
    const label = String(objectiveMarkLine(220).label.formatter)
    expect(label).toContain('220')
    expect(out).toContain(label)
  })

  it('monte le journal réel avec une absence honnête, pas un compteur fabriqué', () => {
    const out = html()
    expect(out).toContain('aria-label="Journal des événements"')
    expect(out).toContain('Aucun événement')
    expect(out).not.toMatch(/0 événement/)
    expect(instrument.events.list()).toEqual([])
  })

  it('le journal reflète l’anneau d’événements réel quand il contient un fait', () => {
    instrument.events.push({ ts: 1000, kind: 'phase', label: 'Charge' })
    const out = html()
    expect(out).toContain('Charge')
    expect(out).toContain('JOURNAL')
  })
})
