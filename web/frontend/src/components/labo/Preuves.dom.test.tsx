// TDD — le sceau n'est dit vérifié qu'après un acte de vérification réel :
// réponse disponible ET empreinte complète de 64 caractères hexadécimaux.
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import SealInspector, { integrityVerified, SealBadge } from './SealInspector'
import Preuves from './Preuves'
import type { AuditReport } from '../../lib/types'
import type { ArchiveRow } from '../../lib/archive'

const HASH = 'a'.repeat(64)
const mount = (node: ReactNode): HTMLElement => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}
const reportWith = (patch: Record<string, unknown>): AuditReport =>
  ({ audit_id: 'a-1', measured: {}, ...patch } as AuditReport)

describe('sceau — VÉRIFIÉ seulement après une vraie vérification', () => {
  it('une empreinte incomplète n’est jamais vérifiée', () => {
    expect(integrityVerified({ available: true, sha256: 'abc' })).toBe(false)
  })

  it('une source indisponible, même avec empreinte complète, n’est pas vérifiée', () => {
    expect(integrityVerified({ available: false, sha256: HASH })).toBe(false)
  })

  it('seule une réponse disponible avec empreinte complète est vérifiée', () => {
    expect(integrityVerified({ available: true, sha256: HASH })).toBe(true)
    expect(integrityVerified({ available: true, sha256: 'A'.repeat(64) })).toBe(true)
  })

  it('le badge ne dit VÉRIFIÉ que quand l’acte a réussi', () => {
    expect(renderToStaticMarkup(<SealBadge verified />)).toContain('VÉRIFIÉ')
    const pending = renderToStaticMarkup(<SealBadge verified={false} />)
    expect(pending).not.toContain('VÉRIFIÉ')
    expect(pending).toContain('À VÉRIFIER')
  })

  it('l’inspecteur ne se dit pas vérifié au repos', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<SealInspector />)
    expect(host.textContent).not.toContain('VÉRIFIÉ')
  })

  it('preuve absente : le vide enseigne ce qui viendra, sans chiffre inventé', () => {
    const host = mount(<Preuves />)
    const wait = host.querySelector('.archive-wait')?.textContent ?? ''
    expect(wait).toContain('Preuve manquante')
    expect(wait).toContain("s'afficheront ici")
    expect(host.textContent).toContain('—')
  })

  it('la ligne de corpus vient de la source réelle, sinon elle dit son absence', () => {
    const real = mount(<Preuves report={reportWith({ provenance: { source_file: 'corpus-2026.csv' } })} archiveRows={[]} integrity={null} />)
    expect(real.querySelector('.corpus-line')?.textContent).toContain('corpus-2026.csv')
    const absent = mount(<Preuves archiveRows={[]} integrity={null} />)
    const line = absent.querySelector('.corpus-line')?.textContent ?? ''
    expect(line).toContain('—')
    expect(line).toContain('Preuve manquante')
  })

  it('la table d’archive n’offre les actions qu’aux lignes scellées', () => {
    const sealed: ArchiveRow = { audit_id: 'a-9', has_report: 'true', verdict_level: 'green', verdict_word: 'Lien sain', timestamp: '2026-10-29T08:00:00Z', site: 'X', rtt_loaded_p95_ms: '220' }
    const unsealed: ArchiveRow = { audit_id: 'a-1', has_report: 'false', timestamp: '2026-10-29T08:00:00Z', site: 'Y' }
    const rows = mount(<Preuves archiveRows={[sealed, unsealed]} integrity={null} />).querySelectorAll('.archive-row')
    expect(rows).toHaveLength(2)
    expect(rows[0].textContent).toContain('Ouvrir')
    expect(rows[0].textContent).toContain('RAPPORT MD')
    expect(rows[0].textContent).toContain('CITER')
    expect(rows[1].textContent).toContain('Preuve manquante')
    expect(rows[1].textContent).not.toContain('CITER')
  })

  it('quatre divulgations documentent quarantaine, figures, matériel et rejeu', () => {
    const labels = Array.from(mount(<Preuves archiveRows={[]} integrity={null} />)
      .querySelectorAll('.inline-disclosure-trigger')).map(b => b.textContent ?? '').join(' ')
    expect(labels).toContain('Quarantaine')
    expect(labels).toContain('Figures')
    expect(labels).toContain('Matériel')
    expect(labels).toContain('Rejeu')
  })

  it('un acte de vérification est proposé, le sceau n’est pas dit vérifié au repos', () => {
    const html = mount(<Preuves archiveRows={[]} integrity={null} />).textContent ?? ''
    expect(html).toContain('Vérifier le sceau')
    expect(html).toContain('À VÉRIFIER')
    expect(html).not.toContain('VÉRIFIÉ')
  })
})
