import { describe, expect, it } from 'vitest'
import { auditCitation } from './citation'
import type { AuditReport } from './types'

const HEX = 'a'.repeat(64)

function report(patch: Partial<AuditReport> = {}): AuditReport {
  return { audit_id: 'audit-1', measured: { direction: 'down' }, ...patch }
}

describe('auditCitation', () => {
  it('carries the id, site, link/provider, timestamp, direction and the full seal', () => {
    const s = auditCitation(report({
      site: 'site-a',
      link_type: '4g',
      provider: 'yas',
      timestamp: '2026-10-28T10:00:00Z',
      provenance: { source_sha256: HEX },
    }))
    expect(s).toContain('audit-1')
    expect(s).toContain('site-a')
    expect(s).toContain('4g/yas')
    expect(s).toContain('direction descendant (down)')
    expect(s).toContain(HEX)
  })

  it('omits missing fields and the seal sentence when unsealed', () => {
    const s = auditCitation(report({ site: undefined, link_type: undefined, provider: undefined, timestamp: undefined, measured: {} }))
    expect(s).toBe('Audit CGO audit-1.')
    expect(s).not.toContain('Rapport scellé')
  })

  it('returns an empty string for an empty id, never a partial citation', () => {
    expect(auditCitation(report({ audit_id: '' }))).toBe('')
  })
})
