// Bandeau compact (§1.4) : la même source de vérité que le spine, repliée.
// Un seul mot de verdict, une absence jamais chiffrée, le chiffre-clé et le
// mot d'état dans le bandeau, les quatre faits derrière un disclosure dont
// l'état est dit par aria-expanded et par la forme du signe.
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it } from 'vitest'
import AnswerBand from './AnswerBand'
import DecisionSpine from './DecisionSpine'
import type { AuditReport } from '../../lib/types'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const REPORT = {
  audit_id: 'audit-t4',
  site: 'Siège',
  link_type: 'fiber',
  measured: {
    rtt_idle_p50_ms: 20,
    rtt_loaded_p95_ms: 78,
    small_measured: true,
    http_small_p95_ms: 64,
    load_measured: true,
  },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: "Sous charge, la latence reste sous l'objectif." },
  cap_sla: { measured: true, ok_pct: 99.1, target_pct: 95, deadline_ms: 220 },
  bloat_grade: 'A',
  economics: { primary: { model: 'A', name: 'ligne dédiée', ar_per_hour: 1388.89 } },
  remediation: { qdisc: 'cake', cc: 'bbr' },
  remediation_state: 'available',
  remediation_available: true,
} as unknown as AuditReport

let cleanup: (() => void) | null = null

function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(node) })
  cleanup = () => {
    act(() => root.unmount())
    host.remove()
  }
  return host
}

afterEach(() => {
  cleanup?.()
  cleanup = null
})

describe('AnswerBand compact — une seule source de vérité', () => {
  it('un seul mot de verdict, rendu mot pour mot, même dans le bandeau replié', () => {
    const host = mount(<AnswerBand><DecisionSpine report={REPORT} /></AnswerBand>)
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
    expect(host.querySelector('#verdict-word')?.textContent).toContain('Lien sain')
  })

  it('absence : le mot pinned reste, aucun chiffre fabriqué', () => {
    const host = mount(<AnswerBand><DecisionSpine report={null} /></AnswerBand>)
    const word = host.querySelector('#verdict-word')!
    expect(word.textContent).toContain('MESURE ABSENTE')
    expect(word.getAttribute('data-state')).toBe('absent')
    expect(host.textContent).not.toMatch(/\d+\s*ms\b/)
    for (const slot of host.querySelectorAll('.state-slot[data-state="absent"]')) {
      expect(slot.textContent).toContain('—')
      expect(slot.textContent).not.toMatch(/^0\b/)
    }
  })

  it('le bandeau porte PREUVE et le révélateur 4 faits, fermé par défaut', () => {
    const host = mount(<AnswerBand><DecisionSpine report={REPORT} /></AnswerBand>)
    expect(host.querySelector('.band-proof')?.textContent?.trim()).toBe('PREUVE')
    const trigger = host.querySelector('.facts-disclosure-trigger')!
    expect(trigger.textContent).toContain('4 faits')
    expect(trigger.getAttribute('aria-expanded')).toBe('false')
    expect(trigger.getAttribute('aria-controls')).toBe('answer-facts')
    const facts = host.querySelector('.answer-facts')!
    expect(facts.id).toBe('answer-facts')
    expect(facts.getAttribute('data-open')).toBe('false')
  })

  it('4 faits bascule aria-expanded, la forme du signe et révèle les faits dans le flux', () => {
    const host = mount(<AnswerBand><DecisionSpine report={REPORT} /></AnswerBand>)
    const trigger = host.querySelector<HTMLButtonElement>('.facts-disclosure-trigger')!
    expect(trigger.querySelector('.facts-disclosure-mark')?.textContent).toBe('▸')

    act(() => { trigger.dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    expect(trigger.getAttribute('aria-expanded')).toBe('true')
    expect(trigger.querySelector('.facts-disclosure-mark')?.textContent).toBe('▾')
    const facts = host.querySelector<HTMLElement>('.answer-facts')!
    expect(facts.getAttribute('data-open')).toBe('true')
    expect(facts.querySelectorAll('.state-slot')).toHaveLength(4)

    act(() => { trigger.dispatchEvent(new MouseEvent('click', { bubbles: true })) })
    expect(trigger.getAttribute('aria-expanded')).toBe('false')
    expect(host.querySelector('.answer-facts')?.getAttribute('data-open')).toBe('false')
  })

  it('le chiffre-clé et le mot d’état viennent du jugement, pas d’un chiffre inventé', () => {
    const measured = mount(<AnswerBand><DecisionSpine report={REPORT} /></AnswerBand>)
    // REPORT : p95 repos 20, p95 charge 78 ⇒ Δ +58 ms, mot « Lien prêt ».
    expect(measured.querySelector('.band-number')?.textContent).toContain('+58 ms')
    // fiber : Δ +58 ms > seuil 30 ⇒ « Critique sous charge », jamais un vert inventé.
    expect(measured.querySelector('.band-status')?.textContent).toBe('Critique sous charge')

    const absent = mount(<AnswerBand><DecisionSpine report={null} /></AnswerBand>)
    expect(absent.querySelector('.band-number')?.textContent).toMatch(/—/)
    expect(absent.querySelector('.band-number')?.textContent).not.toMatch(/\d+\s*ms\b/)
    expect(absent.querySelector('.band-status')?.textContent).toBe('Non mesurée')
  })
})
