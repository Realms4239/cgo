import { describe, expect, it } from 'vitest'
import { b2bARPerHour, capTone, gradeTone, primaryARPerHour } from './AuditReportCard'
import type { AuditReport } from '../lib/types'

function report(patch: Partial<AuditReport> = {}): AuditReport {
  return { audit_id: 'audit-test', measured: {}, ...patch }
}

describe('SLA CAP — jamais un vert par défaut', () => {
  it('non mesuré ⇒ steel', () => {
    expect(capTone(undefined)).toBe('steel')
    expect(capTone({ measured: false, ok_pct: 0 })).toBe('steel')
  })
  it('manquement ⇒ danger, conforme ⇒ success', () => {
    expect(capTone({ measured: true, breach: true })).toBe('danger')
    expect(capTone({ measured: true, breach: false })).toBe('success')
  })
})

describe('note bufferbloat', () => {
  it('A/B/C/D mappent vers les tons du Badge', () => {
    expect(gradeTone('A')).toBe('success')
    expect(gradeTone('B')).toBe('live')
    expect(gradeTone('C')).toBe('warning')
    expect(gradeTone('D')).toBe('danger')
    expect(gradeTone(undefined)).toBe('steel')
  })
})

describe('économie liée à economics.* (P0)', () => {
  it('le primaire lit economics.primary.ar_per_hour', () => {
    const r = report({
      economics: {
        primary_link_type: '5g',
        primary: { model: 'B', name: 'b2b-m2m-apn', ar_per_hour: 1234.5, basis: 'devis' },
        scenarios: [{ model: 'B', name: 'b2b-m2m-apn', ar_per_hour: 1234.5, basis: 'devis' }],
      },
    })
    expect(primaryARPerHour(r)).toBe(1234.5)
    expect(b2bARPerHour(r)).toBe(1234.5)
  })

  it('sans économie ⇒ null, jamais un zéro inventé', () => {
    expect(primaryARPerHour(report())).toBeNull()
    expect(b2bARPerHour(report())).toBeNull()
  })
})
