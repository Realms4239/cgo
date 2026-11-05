import { describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { AuditReportsView } from './AuditReportsPanel'
import type { AuditReport } from '../lib/types'

const SEAL = 'b'.repeat(64)

const sealedReport: AuditReport = {
  audit_id: 'audit-dom',
  site: 'site-dom',
  link_type: '4g',
  provider: 'yas',
  measured: { direction: 'down', load_measured: true, http_small_p95_ms: 210.5, small_measured: true },
  provenance: { source_file: 'audit_report.json', row_count: 1, source_sha256: SEAL },
}

const base = {
  error: null,
  msg: '',
  citationMsg: null,
  onRetry: vi.fn(),
  onOpen: vi.fn(),
  onCopyCitation: vi.fn(),
  onCloseReport: vi.fn(),
  onCopy: vi.fn(),
}

describe('AuditReportsView', () => {
  it('rend l’attente avant la première réponse', () => {
    const html = renderToStaticMarkup(<AuditReportsView {...base} rows={null} report={null} />)
    expect(html).toContain('chargement…')
  })

  it('rend l’erreur avec sa reprise, pas le silence', () => {
    const html = renderToStaticMarkup(
      <AuditReportsView {...base} rows={[]} error="liste des rapports indisponible" report={null} />,
    )
    expect(html).toContain('liste des rapports indisponible')
    expect(html).toContain('Réessayer')
  })

  it('rend l’archive virtualisée et compte ses lignes', () => {
    const rows = Array.from({ length: 179 }, (_, i) => ({
      audit_id: `a-${i}`,
      site: `site-${i}`,
      has_report: 'true',
      verdict_level: 'green',
      timestamp: '2026-10-29T08:00:00Z',
    }))
    const html = renderToStaticMarkup(<AuditReportsView {...base} rows={rows} report={null} />)
    expect(html).toContain('archive : 179 lignes')
    expect(html).toContain('aria-rowcount="179"')
  })

  it('ouvre le rapport sans démonter la liste', () => {
    const rows = [{
      audit_id: 'audit-dom',
      site: 'site-dom',
      has_report: 'true',
      timestamp: '2026-10-29T08:00:00Z',
    }]
    const html = renderToStaticMarkup(<AuditReportsView {...base} rows={rows} report={sealedReport} />)
    expect(html).toContain('Rapport ouvert')
    expect(html).toContain('aria-rowcount="1"')
    expect(html).toContain('RETOUR À LA LISTE')
  })
})
