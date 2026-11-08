import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import { ArchiveTable } from './ArchiveTable'
import type { ArchiveRow } from '../lib/archive'

const makeRows = (n: number): ArchiveRow[] =>
  Array.from({ length: n }, (_, i) => ({
    audit_id: `a-${i}`,
    site: `site-${i}`,
    link_type: 'fiber',
    timestamp: '2026-10-29T08:00:00Z',
    has_report: i % 3 === 0 ? 'false' : 'true',
    verdict_level: i % 4 === 0 ? 'neutral' : i % 4 === 1 ? 'red' : 'green',
    verdict_word: i % 4 === 1 ? 'Alertes en retard' : 'Lien sain',
    rtt_loaded_p95_ms: '328',
  }))

function markup(rows: ArchiveRow[]) {
  return renderToStaticMarkup(
    <ArchiveTable rows={rows} onOpen={vi.fn()} onCopyCitation={vi.fn()} />,
  )
}

describe('ArchiveTable', () => {
  it('ne rend jamais les 179 lignes dans le DOM initial', () => {
    const html = markup(makeRows(179))
    expect(html).toContain('aria-rowcount="179"')
    const rows = (html.match(/<tr/g) ?? []).length
    expect(rows).toBeLessThanOrEqual(30)
  })

  it('annonce le nombre total et le statut scellé des visibles', () => {
    const html = markup(makeRows(20))
    expect(html).toContain('archive : 20 lignes')
    expect(html).toContain('Valide')
  })

  it('une ligne sans dossier ne propose que des actions désactivées', () => {
    const html = markup([
      { audit_id: 'a-1', site: 'Site test', has_report: 'false', timestamp: '2026-10-29T08:00:00Z' },
    ])
    expect(html).toContain('Preuve manquante')
    expect(html).not.toContain('OUVRIR</button>')
  })

  it('une ligne scellée expose les trois actions', () => {
    const html = markup([
      {
        audit_id: 'a-9',
        site: 'Département X',
        has_report: 'true',
        verdict_level: 'red',
        verdict_word: 'Alertes en retard',
        timestamp: '2026-10-29T08:00:00Z',
        rtt_loaded_p95_ms: '328',
      },
    ])
    expect(html).toContain('Dégradé')
    expect(html).toContain('Alertes en retard')
    expect(html).toContain('Ouvrir')
    expect(html).toContain('RAPPORT MD')
    expect(html).toContain('CITER')
    expect(html).toContain('a-9')
  })

  it('l’absence de date est dite, jamais remplacée par zéro', () => {
    const html = markup([{ audit_id: 'a-2', has_report: 'false' }])
    expect(html).toContain('—')
  })

  it('rend l’état vide, pas une table fantôme', () => {
    const html = renderToStaticMarkup(
      <ArchiveTable rows={[]} onOpen={vi.fn()} onCopyCitation={vi.fn()} />,
    )
    expect(html).toContain('Aucun rapport')
    expect(html).not.toContain('<table')
  })
})
