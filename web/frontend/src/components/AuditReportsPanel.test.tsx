import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { AuditReportTable } from './AuditReportsPanel'

describe('AuditReportTable', () => {
  it('liste un rapport avec son site, sa note et son action', () => {
    const html = renderToStaticMarkup(
      <AuditReportTable
        rows={[{ audit_id: 'audit-7', site: 'Département X', link_type: '5g', timestamp: '2026-10-29T08:00:00Z', rtt_loaded_p95_ms: '328', bloat_grade: 'D', has_report: 'true' }]}
        onOpen={() => {}}
      />,
    )
    expect(html).toContain('Département X')
    expect(html).toContain('2026-10-29')
    expect(html).toContain('328')
    expect(html).toContain('OUVRIR')
  })

  it('dit l’absence de rapport sans inventer de ligne', () => {
    const html = renderToStaticMarkup(<AuditReportTable rows={[]} onOpen={() => {}} />)
    expect(html).toContain('Aucun rapport d')
    expect(html).toContain('archivé.')
  })

  it('un rapport sans identifiant ne s’ouvre pas', () => {
    const html = renderToStaticMarkup(
      <AuditReportTable rows={[{ site: 'site sans id' }]} onOpen={() => {}} />,
    )
    expect(html).toContain('disabled')
  })

  it('une ligne sans dossier d’archive est dite avant le clic', () => {
    const html = renderToStaticMarkup(
      <AuditReportTable rows={[{ audit_id: 'audit-old', site: 'Site X', has_report: 'false' }]} onOpen={() => {}} />,
    )
    expect(html).toContain('non archivé')
    expect(html).not.toContain('OUVRIR')
  })
})
