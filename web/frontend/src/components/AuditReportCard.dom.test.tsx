import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { AuditReportCard } from './AuditReportCard'
import type { AuditReport } from '../lib/types'

const SEAL = 'b'.repeat(64)

function report(patch: Partial<AuditReport> = {}): AuditReport {
  return {
    audit_id: 'audit-dom',
    site: 'site-dom',
    link_type: '4g',
    provider: 'yas',
    measured: {
      direction: 'down',
      load_measured: true,
      throughput_mbps: 18.2,
      http_small_p95_ms: 210.5,
      small_measured: true,
    },
    provenance: { source_file: 'audit_report.json', row_count: 1, source_sha256: SEAL },
    ...patch,
  }
}

function html(r: AuditReport): string {
  return renderToStaticMarkup(<AuditReportCard report={r} onCopy={() => {}} />)
}

describe('AuditReportCard surface', () => {
  it('renders ingress scripts only when the report carries them', () => {
    const withIngress = html(report({
      remediation: {
        qdisc: 'cake',
        linux_tc: ['tc qdisc add dev eth0 root cake'],
        routeros: '/queue type add name=cake',
        linux_tc_ingress: ['tc qdisc add dev ifb0 root cake'],
        routeros_ingress: '/queue tree add name=ingress',
      },
    }))
    expect(withIngress).toContain('audit-remediation-ingress')
    expect(withIngress).toContain('COPIER LINUX TC (INGRESS)')
    expect(withIngress).toContain('COPIER ROUTEROS (INGRESS)')

    const without = html(report({ remediation: { qdisc: 'cake', linux_tc: ['tc qdisc add dev eth0 root cake'] } }))
    expect(without).not.toContain('audit-remediation-ingress')
  })

  it('renders the audit window only when phases are present', () => {
    const withPhases = html(report({ phases: { baseline_seconds: 30, load_seconds: 120, recovery_seconds: 30 } }))
    expect(withPhases).toContain('audit-phases')
    expect(withPhases).toContain('repos')
    expect(withPhases).toContain('récupération')

    expect(html(report())).not.toContain('audit-phases')
  })

  it('renders one provenance idiom and keeps the citation action', () => {
    const h = html(report())
    expect((h.match(/data-testid="provenance"/g) ?? []).length).toBe(1)
    expect(h).toContain('CITER CE RAPPORT')
    expect(h).toContain(SEAL)
  })
})

describe('AuditReportCard — validation RouterOS', () => {
  const remediation = { qdisc: 'cake', linux_tc: ['tc qdisc add dev eth0 root cake'], routeros: '/queue type add name=cgo-cake kind=cake' }

  it('dit la validation avec la version, le niveau et l’empreinte', () => {
    const h = html(report({
      remediation,
      routeros_validation: {
        version: '7.15.3', level: 'windows-tcg', export_sha256: 'abcdef1234567890',
        checks: [{ name: 'queue type cgo-cake.kind', expected: 'cake', found: 'cake', pass: true }],
      },
    }))
    expect(h).toContain('Configuration RouterOS validée')
    expect(h).toContain('7.15.3')
    expect(h).toContain('abcdef12')
    expect(h).not.toContain('appliquée')
  })

  it('un écart dit non validée', () => {
    const h = html(report({
      remediation,
      routeros_validation: {
        version: '7.15.3', level: 'windows-tcg', export_sha256: 'abcdef1234567890',
        checks: [{ name: 'queue type cgo-cake.kind', expected: 'cake', found: 'none', pass: false }],
      },
    }))
    expect(h).toContain('Configuration RouterOS non validée')
    expect(h).not.toContain('Configuration RouterOS validée (')
  })

  it('un script sans preuve le dit', () => {
    const h = html(report({ remediation }))
    expect(h).toContain('non validée')
    expect(h).toContain('aucune preuve CHR')
  })

  it('zéro vérification n’est pas une validation', () => {
    const h = html(report({
      remediation,
      routeros_validation: { version: '7.15.3', level: 'windows-tcg', export_sha256: 'abcdef1234567890', checks: [] },
    }))
    expect(h).toContain('non validée — aucune vérification')
    expect(h).not.toContain('Configuration RouterOS validée (')
  })
})
