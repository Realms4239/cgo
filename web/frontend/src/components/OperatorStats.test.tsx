import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { OperatorStats } from './OperatorStats'
import type { AuditReport } from '../lib/types'

const measured = {
  audit_id: 'audit-x',
  measured: { direction: 'down', load_measured: true, rtt_loaded_p95_ms: 64 },
  cap_sla: { measured: true, ok_pct: 97 },
  bloat_grade: 'C',
  bloat_verdict: 'queue sous charge',
} as unknown as AuditReport

const html = (r: AuditReport | null) => renderToStaticMarkup(<OperatorStats report={r} warnMs={80} critMs={150} />)

describe('OperatorStats — une absence ne pèse pas comme une mesure', () => {
  it('sans rapport : les trois cellules sont marquées vides', () => {
    const h = html(null)
    expect((h.match(/data-empty/g) ?? []).length).toBe(3)
  })
  it('rapport mesuré : aucune cellule vide', () => {
    const h = html(measured)
    expect((h.match(/data-empty/g) ?? []).length).toBe(0)
    expect(h).toContain('64')
  })
})
