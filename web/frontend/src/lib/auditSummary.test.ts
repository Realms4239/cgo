import { describe, expect, it, beforeEach } from 'vitest'
import { compareAudits, loadLastSummary, saveLastSummary, summaryOf, type AuditSummary } from './auditSummary'
import type { AuditReport } from './types'

const report = (over: Partial<AuditReport> = {}): AuditReport => ({
  audit_id: 'audit-1',
  site: 'Département X',
  link_type: '5g',
  timestamp: '2026-10-29T10:00:00Z',
  measured: {
    http_small_p95_ms: 328,
    small_measured: true,
    rtt_loaded_p95_ms: 410,
    rtt_idle_p95_ms: 55,
    qdi_ms: 355,
  },
  cap_sla: { measured: true, ok_pct: 60 },
  bloat_grade: 'D',
  ...over,
})

describe('summaryOf', () => {
  it('préfère le petit objet mesuré, garde le chargé en repli', () => {
    expect(summaryOf(report()).small_p95_ms).toBe(328)
    const fallback = summaryOf(report({ measured: { rtt_loaded_p95_ms: 410 } }))
    expect(fallback.small_p95_ms).toBe(410)
  })

  it('un SLA non mesuré reste absent — jamais 0 %', () => {
    expect(summaryOf(report({ cap_sla: { measured: false } })).deadline_ok_pct).toBeUndefined()
  })
})

describe('compareAudits', () => {
  const prev: AuditSummary = { audit_id: 'a0', site: 'Département X', link_type: '5g', small_p95_ms: 400, deadline_ok_pct: 50 }

  it('compare même site et même lien', () => {
    const c = compareAudits(prev, { audit_id: 'a1', site: 'Département X', link_type: '5g', small_p95_ms: 300, deadline_ok_pct: 80 })
    expect(c.comparable).toBe(true)
    expect(c.smallDeltaPct).toBe(-25)
    expect(c.deadlineDelta).toBe(30)
  })

  it('refuse un site différent', () => {
    expect(compareAudits(prev, { ...prev, site: 'Département Y' }).comparable).toBe(false)
  })

  it('refuse un type de lien différent', () => {
    expect(compareAudits(prev, { ...prev, link_type: 'vsat' }).reason).toContain('lien')
  })

  it('sans précédent, aucune comparaison', () => {
    expect(compareAudits(null, prev).comparable).toBe(false)
  })
})

describe('résumé local', () => {
  beforeEach(() => localStorage.clear())

  it('aller-retour localStorage', () => {
    const s = summaryOf(report())
    saveLastSummary(s)
    expect(loadLastSummary()?.audit_id).toBe('audit-1')
  })

  it('un localStorage corrompu ne casse pas le tableau de bord', () => {
    localStorage.setItem('cgo-audit-last', '{oops')
    expect(loadLastSummary()).toBeNull()
  })
})
