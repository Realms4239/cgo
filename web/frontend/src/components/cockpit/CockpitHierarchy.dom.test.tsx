// Oracle indépendant : la hiérarchie est lue dans le CSS réel et dans le DOM
// produit par le vrai DecisionSpine — aucun mock du composant testé.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it } from 'vitest'
import DecisionSpine from './DecisionSpine'
import type { AuditReport } from '../../lib/types'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const REPORT = {
  audit_id: 'audit-t1', site: 'Siège', link_type: 'fiber',
  measured: { rtt_idle_p50_ms: 20, rtt_loaded_p95_ms: 78, small_measured: true,
    http_small_p95_ms: 64, load_measured: true },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: 'Sous charge, la latence reste sous objectif.' },
  cap_sla: { measured: true, ok_pct: 99.1, target_pct: 95, deadline_ms: 220 },
  bloat_grade: 'A', remediation_state: 'available', remediation_available: true,
} as unknown as AuditReport

let cleanup: (() => void) | null = null
function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(node) })
  cleanup = () => { act(() => root.unmount()); host.remove() }
  return host
}
afterEach(() => { cleanup?.(); cleanup = null })

const css = (name: string) => readFileSync(join(process.cwd(), 'src/styles', name), 'utf8')

describe('cockpit — hiérarchie et rythme partagés', () => {
  it('déclare une seule échelle de type et un seul rythme', () => {
    const tokens = css('tokens.css')
    expect(tokens).toContain('--cockpit-font-hero: 600 var(--fs-hero)/1.15 var(--font-ui);')
    expect(tokens).toContain('--cockpit-font-value: 400 var(--fs-body)/1.5 var(--font-ui);')
    expect(tokens).toContain('--cockpit-font-label: 500 var(--fs-fine)/1.3 var(--font-mono);')
    expect(tokens).toContain('--cockpit-rhythm: var(--sp-2);')
  })
  it('le verdict précède les faits dans le DOM rendu', () => {
    const host = mount(<DecisionSpine report={REPORT} />)
    const word = host.querySelector('#verdict-word')
    const facts = host.querySelector('.answer-facts')
    expect(word).not.toBeNull()
    expect(facts).not.toBeNull()
    expect(word!.compareDocumentPosition(facts!) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
  })
  it('les déclarations de type utilisent les alias, pas des valeurs en double', () => {
    const components = css('components.css')
    expect(components).toContain('.answer-word { font: var(--cockpit-font-hero);')
    expect(components).toContain('.answer-sentence { font: var(--cockpit-font-value);')
    expect(components).toContain('.health-label { font: var(--cockpit-font-label);')
  })
})
