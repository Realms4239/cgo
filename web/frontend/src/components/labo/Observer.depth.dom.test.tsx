// TDD — profondeur CH2. Oracle indépendant : les mots et les états sont lus
// dans le HTML rendu, pas dans le composant. Trois vérités à prouver :
// portes tri-states lues sur la valeur réelle (jamais sur l'index), cellules
// absentes en « — » + mot français (jamais un zéro), et attente vs FINALE.
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import type { ReactElement } from 'react'
import Matrix from './Matrix'
import Gates from './Gates'
import CampaignTimeline, { campaignBands } from './CampaignTimeline'

const T0 = Date.UTC(2026, 10, 21, 13, 45, 0)
const plan = { profiles: ['P1', 'P2'], qdiscs: ['pfifo_fast', 'cake'], ccs: ['cubic', 'bbr'], reps: 2 }

function mount(node: ReactElement): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node)
  return host
}

describe('Portes G0–G7 — texte tri-state lu sur la mesure', () => {
  it('PASS / FAIL / — viennent de la valeur, pas du rang (G0 nul reste —)', () => {
    const host = mount(<Gates gates={[null, true, false, null, null, null, null, null]} />)
    const rows = host.querySelectorAll('.labo-gate-row')
    expect(rows).toHaveLength(8)
    expect(rows[0].getAttribute('data-state')).toBe('na')
    expect(rows[0].querySelector('.labo-gate-token')?.textContent).toBe('—')
    expect(rows[0].textContent).toContain('non évaluée')
    expect(rows[1].getAttribute('data-state')).toBe('pass')
    expect(rows[1].querySelector('.labo-gate-token')?.textContent).toBe('PASS')
    expect(rows[2].getAttribute('data-state')).toBe('fail')
    expect(rows[2].querySelector('.labo-gate-token')?.textContent).toBe('FAIL')
    expect(host.querySelectorAll('[data-state="pass"]')).toHaveLength(1)
    expect(host.querySelectorAll('[data-state="fail"]')).toHaveLength(1)
    // Aucune pastille/badge vert-par-index : le mot seul porte l'état.
    expect(host.querySelector('[class*="badge"]')).toBeNull()
    // Divulgation science inline, pas un vert décoratif.
    expect(host.textContent).toMatch(/science/i)
  })

  it('aucune valeur : huit « — », jamais de PASS fabriqué', () => {
    const host = mount(<Gates gates={[]} />)
    expect(host.querySelectorAll('[data-state="na"]')).toHaveLength(8)
    expect(host.textContent).not.toContain('PASS')
    expect(host.textContent).not.toContain('FAIL')
  })
})

describe('Matrice — cellules absentes et ordre serveur', () => {
  it('au repos : attente, « — » par cellule, jamais un compte 0/n fabriqué', () => {
    const host = mount(<Matrix plan={plan} progress={null} />)
    const rows = host.querySelectorAll('.labo-matrix-table tbody tr')
    expect(rows).toHaveLength(8) // 2 profils × 2 files × 2 contrôles
    for (const row of rows) {
      expect(row.getAttribute('data-state')).toBe('attente')
      expect(row.textContent).toContain('—')
      expect(row.textContent).not.toMatch(/\d+\s*\/\s*\d+/)
    }
    expect(host.querySelector('.labo-matrix')?.getAttribute('data-state')).toBe('attente')
    expect(host.textContent).toMatch(/attente/i)
    expect(host.textContent).toMatch(/campagne/i) // une seule phrase
    expect(host.querySelector('[data-state="ok"], [data-state="pass"]')).toBeNull()
  })

  it('plan indisponible : aucune cellule inventée, « — » + mot d’absence', () => {
    const host = mount(<Matrix plan={{ profiles: [], qdiscs: [], ccs: [], reps: 0 }} progress={null} />)
    expect(host.querySelectorAll('.labo-matrix-table')).toHaveLength(0)
    expect(host.textContent).toContain('—')
    expect(host.textContent).toMatch(/non disponible/i)
  })

  it('campagne terminée : mots FINALE / terminée, jamais retour à attente', () => {
    const total = plan.profiles.length * plan.qdiscs.length * plan.ccs.length * plan.reps // 16 événements
    const host = mount(<Matrix plan={plan} progress={{ running: false, eventId: total, doneEvents: total, totalEvents: total }} />)
    expect(host.querySelector('.labo-matrix')?.getAttribute('data-state')).toBe('final')
    expect(host.textContent).toContain('FINALE')
    expect(host.textContent).not.toMatch(/attente/i)
    const rows = host.querySelectorAll('.labo-matrix-table tbody tr')
    expect(rows[0].getAttribute('data-state')).toBe('terminée')
    expect(rows[0].textContent).toContain('2/2')
    expect(rows[7].getAttribute('data-state')).toBe('terminée')
  })

  it('en cours : la cellule courante porte son vrai n/n dans l’ordre serveur', () => {
    const host = mount(<Matrix plan={plan} progress={{ running: true, eventId: 5, doneEvents: 4, totalEvents: 8 }} />)
    expect(host.querySelector('.labo-matrix')?.getAttribute('data-state')).toBe('live')
    const rows = host.querySelectorAll('.labo-matrix-table tbody tr')
    expect(rows[0].getAttribute('data-state')).toBe('terminée')
    expect(rows[0].textContent).toContain('2/2')
    expect(rows[2].getAttribute('data-state')).toBe('en cours')
    expect(rows[2].textContent).toContain('1/2')
    expect(rows[3].getAttribute('data-state')).toBe('attente')
  })
})

describe('Chronologie — une seule horloge réelle', () => {
  it('bornes issues des temps réels seulement, sans décalage fabriqué', () => {
    const bands = campaignBands({ baseline: T0, charge: T0 + 10000 }, T0 + 15000, false)
    expect(bands).toHaveLength(2)
    expect(bands[0]).toMatchObject({ key: 'baseline', start: T0, end: T0 + 10000 })
    expect(bands[1]).toMatchObject({ key: 'charge', start: T0 + 10000, end: T0 + 10000 })
    const live = campaignBands({ baseline: T0, charge: T0 + 10000 }, T0 + 15000, true)
    expect(live[1].end).toBe(T0 + 15000) // phase ouverte = maintenant (horloge réelle)
  })

  it('aucune phase : « — » + phrase, aucune bande fantôme', () => {
    const host = mount(<CampaignTimeline phaseSince={{}} running={false} />)
    expect(host.querySelector('.labo-timeline')?.getAttribute('data-state')).toBe('absent')
    expect(host.textContent).toContain('—')
    expect(host.textContent).toMatch(/aucune phase/i)
    expect(host.querySelectorAll('.labo-tl-band')).toHaveLength(0)
  })

  it('durées lues sur les horodatages, jamais un 30 s ou 15 s inventé', () => {
    const host = mount(<CampaignTimeline phaseSince={{ baseline: T0, charge: T0 + 10000, recup: T0 + 18000 }} running={false} />)
    expect(host.querySelectorAll('.labo-tl-band')).toHaveLength(3)
    expect(host.textContent).toContain('10 s')
    expect(host.textContent).toContain('8 s')
    expect(host.textContent).not.toMatch(/30 s|15 s/)
  })
})
