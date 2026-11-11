import { describe, expect, it } from 'vitest'
import { auditCitation } from './citation'
import type { AuditReport } from './types'
import { measured, provenance, report as sealedReport } from './auditFixtures'

const HEX = 'a'.repeat(64)

function report(patch: Partial<AuditReport> = {}): AuditReport {
  return sealedReport({ audit_id: 'audit-1', measured: measured({ direction: 'down' }), ...patch })
}

// Le fil arrive non typé : les rapports anciens omettent des champs que le
// moteur d'aujourd'hui scelle toujours. La citation doit les ABSORBER sans
// inventer — c'est exactement le cas que ce cast franchit une seule fois.
const wire = (json: string): AuditReport => JSON.parse(json) as unknown as AuditReport

describe('auditCitation', () => {
  it('carries the id, site, link/provider, timestamp, direction and the full seal', () => {
    const s = auditCitation(report({
      site: 'site-a',
      link_type: '4g',
      provider: 'yas',
      timestamp: '2026-10-28T10:00:00Z',
      provenance: provenance({ source_sha256: HEX }),
    }))
    expect(s).toContain('audit-1')
    expect(s).toContain('site-a')
    expect(s).toContain('4g/yas')
    expect(s).toContain('direction descendant (down)')
    expect(s).toContain(HEX)
  })

  it('omits missing fields and the seal sentence when unsealed', () => {
    const s = auditCitation(wire(`{"audit_id":"audit-1","measured":{},"provenance":{"source_sha256":"${HEX}"}}`))
    expect(s).toContain('Rapport scellé')
  })
  it('returns empty when the seal is missing, never a partial citation', () => {
    expect(auditCitation(report({ provenance: undefined }))).toBe('')
    expect(auditCitation(report({ provenance: provenance({ source_sha256: '' }) }))).toBe('')
  })

  it('returns an empty string for an empty id, never a partial citation', () => {
    expect(auditCitation(report({ audit_id: '' }))).toBe('')
  })
})
