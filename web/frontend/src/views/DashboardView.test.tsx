import { describe, expect, it, vi, beforeEach } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import DashboardView from './DashboardView'
import type { AuditReport } from '../lib/types'

const measured = {
  audit_id: 'audit-1',
  timestamp: '2026-11-05T10:00:00Z',
  site: 'Siège du gouvernement',
  link_type: 'fiber',
  provider: 'yas',
  measured: {
    small_measured: true,
    http_small_p95_ms: 64,
    rtt_loaded_p95_ms: 78,
    rtt_idle_p50_ms: 20,
    load_measured: true,
    load_attempted: true,
  },
  bloat_grade: 'A',
  operator_verdict: { level: 'green', word: 'Lien stable', sentence: 'Le débit sous charge tient votre objectif.' },
  cap_sla: { measured: true, ok_pct: 99.1, target_pct: 95, deadline_ms: 220, breach: false },
  economics: {
    disclaimer: 'Scénarios contractuels, pas des mesures.',
    primary: { model: 'A', name: 'ligne dédiée', ar_per_hour: 1388.89 },
  },
  remediation_available: false,
  remediation_state: 'observation_only',
} as unknown as AuditReport

let auditState: Record<string, unknown>
let uiState: Record<string, unknown>

vi.mock('../store/audit', () => ({
  useAuditStore: (selector: (state: Record<string, unknown>) => unknown) => selector(auditState),
}))
vi.mock('../store/ui', () => ({
  useUIStore: (selector: (state: Record<string, unknown>) => unknown) => selector(uiState),
}))

function markup(): string {
  return renderToStaticMarkup(<DashboardView />)
}

beforeEach(() => {
  auditState = { report: null, previous: null, showReport: vi.fn() }
  uiState = { setMode: vi.fn() }
})

describe('DashboardView absent', () => {
  it('renders one CTA and no fake zero', () => {
    const html = markup()
    expect(html).toContain('data-testid="dash-absent"')
    expect(html).toContain('Lancer l’audit de 60 secondes')
    expect(html).not.toMatch(/0\s*(ms|Mbit\/s|Ar)/i)
  })
})

describe('DashboardView measured', () => {
  beforeEach(() => {
    auditState = { report: measured, previous: null, showReport: vi.fn() }
  })

  it('shows class-aware link verdict and no synthetic absent values', () => {
    const html = markup()
    expect(html).toContain('data-testid="dash-gauge"')
    expect(html).toContain('Santé du lien')
    expect(html).toContain('+58 ms')
    expect(html).not.toContain('Pas encore mesuré')
  })
})
