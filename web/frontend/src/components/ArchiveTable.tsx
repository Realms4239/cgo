// Archive des preuves — fenêtre de lignes, aucun chargement par ligne.
// Seules les lignes visibles existent dans le DOM ; le dossier complet n'est
// lu que sur action de l'opérateur.
import { useState, type UIEvent } from 'react'
import {
  archiveStatus,
  archiveVerdictWord,
  computeWindow,
  formatArchiveDate,
  type ArchiveRow,
} from '../lib/archive'
import { EmptyState } from './ui/EmptyState'

export const ARCHIVE_ROW_PX = 44
export const ARCHIVE_WINDOW_ROWS = 8

export function ArchiveTable({
  rows,
  onOpen,
  onCopyCitation,
}: {
  rows: ArchiveRow[]
  onOpen: (id: string) => void
  onCopyCitation: (id: string) => void
}) {
  const [scrollTop, setScrollTop] = useState(0)

  if (rows.length === 0) {
    return <EmptyState kind="empty" hint="Aucun rapport d'audit archivé." />
  }

  const viewportHeight = ARCHIVE_ROW_PX * ARCHIVE_WINDOW_ROWS
  const { start, end } = computeWindow({
    total: rows.length,
    scrollTop,
    rowHeight: ARCHIVE_ROW_PX,
    viewportHeight,
    overscan: 3,
  })
  const visible = rows.slice(start, end)

  const onScroll = (event: UIEvent<HTMLDivElement>) => {
    setScrollTop(event.currentTarget.scrollTop)
  }

  return (
    <div className="archive-block">
      <p className="mono archive-count" aria-live="polite">archive : {rows.length} ligne{rows.length > 1 ? 's' : ''}</p>
      <div
        className="archive-window"
        onScroll={onScroll}
        style={{ height: `${viewportHeight}px` }}
        tabIndex={0}
        role="region"
        aria-label="Table de l’archive — faites défiler pour d’autres lignes"
      >
        <div
          className="archive-spacer"
          style={{ paddingTop: `${start * ARCHIVE_ROW_PX}px`, paddingBottom: `${(rows.length - end) * ARCHIVE_ROW_PX}px` }}
        >
          <table className="data-table archive-table" aria-rowcount={rows.length}>
            <caption className="archive-caption">Archive du lien — {rows.length} lignes</caption>
            <colgroup>
              <col style={{ width: '11%' }} />
              <col style={{ width: '18%' }} />
              <col style={{ width: '17%' }} />
              <col style={{ width: '13%' }} />
              <col style={{ width: '12%' }} />
              <col style={{ width: '16%' }} />
              <col style={{ width: '13%' }} />
            </colgroup>
            <thead>
              <tr>
                <th scope="col">date</th>
                <th scope="col">site</th>
                <th scope="col">lien</th>
                <th scope="col">RTT chargé</th>
                <th scope="col">état</th>
                <th scope="col">verdict</th>
                <th scope="col">actions</th>
              </tr>
            </thead>
            <tbody>
              {visible.map((row, index) => {
                const rowIndex = start + index + 1
                const id = row.audit_id?.trim()
                const sealed = row.has_report === 'true'
                const canOpen = sealed && !!id
                const date = formatArchiveDate(row.timestamp ?? '')
                const rtt = row.rtt_loaded_p95_ms?.trim()
                const status = archiveStatus(row)
                const statusMissing = status === '—'
                const site = row.site?.trim() || '—'
                const link = [row.link_type?.trim(), row.provider?.trim()].filter(Boolean).join(' · ') || '—'
                return (
                  <tr key={id ?? `${site}-${rowIndex}`} aria-rowindex={rowIndex} className="archive-row">
                    <td className="archive-cell" title={row.timestamp?.trim() || 'horodatage absent'}>{date}</td>
                    <td className="archive-cell" title={site}>{site}</td>
                    <td className="archive-cell" title={link}>{link}</td>
                    <td className="archive-cell archive-num" title={rtt ? `${rtt} ms` : 'non mesuré'}>{rtt ? `${rtt} ms` : '—'}</td>
                    <td className="archive-cell" title={statusMissing ? 'rapport non scellé — aucune source' : `verdict scellé : ${status}`}>{status}</td>
                    <td className="archive-cell" title={sealed ? (row.verdict_word ?? 'mot scellé absent') : 'rapport absent'}>{archiveVerdictWord(row)}</td>
                    <td className="archive-cell archive-actions">
                      {canOpen ? (
                        <>
                          <button type="button" className="btn btn-mini" onClick={() => onOpen(id)}>Ouvrir</button>
                          <a
                            className="btn btn-mini"
                            href={`/api/audit/report?audit_id=${encodeURIComponent(id)}&format=md`}
                            download={`${id}.md`}
                          >
                            RAPPORT MD
                          </a>
                          <button type="button" className="btn btn-mini" onClick={() => onCopyCitation(id)}>CITER</button>
                        </>
                      ) : (
                        <span className="data-missing" title="ligne archivée avant le dossier de rapport">Preuve manquante</span>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
      <p className="mono muted archive-note">Le dossier complet n’est chargé que sur « Ouvrir ».</p>
    </div>
  )
}
