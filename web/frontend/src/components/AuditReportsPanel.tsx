// Preuves — les rapports d'audit archivés, citables, ouverts sans nouvel
// endpoint : /api/audit/list puis /api/audit/report?audit_id=.
import { useEffect, useState } from 'react'
import { AuditReportCard } from './AuditReportCard'
import { CardHead } from './ui/CardHead'
import { EmptyState } from './ui/EmptyState'
import { useUIStore } from '../store/ui'
import { asArray } from '../lib/format'
import type { AuditReport } from '../lib/types'

type Row = Record<string, string>

export function AuditReportTable({ rows, onOpen }: { rows: Row[]; onOpen: (id: string) => void }) {
  if (rows.length === 0) {
    return <EmptyState kind="empty" hint="Aucun rapport d'audit archivé." />
  }
  return (
    <div className="bench-matrix-scroll">
      <table className="data-table audit-reports-table">
        <thead>
          <tr><th>site</th><th>lien</th><th>date</th><th>p95 chargé</th><th>qualité</th><th /></tr>
        </thead>
        <tbody>
          {rows.map((a, i) => (
            <tr key={a.audit_id || `${a.site}-${i}`}>
              <td>{a.site || '—'}</td>
              <td>{a.link_type || '—'}{a.provider ? ` · ${a.provider}` : ''}</td>
              <td>{(a.timestamp || '').slice(0, 10)}</td>
              <td>{a.rtt_loaded_p95_ms ? `${a.rtt_loaded_p95_ms} ms` : '—'}</td>
              <td>{a.bloat_grade || '—'}</td>
              <td>
                <button
                  className="btn audit-report-open"
                  disabled={!a.audit_id}
                  onClick={() => a.audit_id && onOpen(a.audit_id)}
                >
                  OUVRIR
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default function AuditReportsPanel() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [report, setReport] = useState<AuditReport | null>(null)
  const [msg, setMsg] = useState('')

  useEffect(() => {
    fetch('/api/audit/list')
      .then(r => r.json())
      .then(j => setRows(asArray<Row>(j.audits)))
      .catch(() => { setRows([]); setMsg('liste des rapports indisponible') })
  }, [])

  const open = (id: string) => {
    setMsg('chargement du rapport…')
    fetch(`/api/audit/report?audit_id=${encodeURIComponent(id)}`)
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(j => { setReport(j as AuditReport); setMsg('') })
      .catch(e => setMsg('rapport introuvable : ' + e.message))
  }

  const copyText = (text: string, label: string) => {
    navigator.clipboard?.writeText(text).then(
      () => useUIStore.getState().pushToast(`${label} copié`, 'ok'),
      () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
    )
  }

  return (
    <section className="card audit-reports" data-testid="audit-reports">
      <CardHead label="Rapports d'audit" sub="archives citables — chaque rapport porte son sceau" />
      {report ? (
        <>
          <button className="btn audit-report-back" onClick={() => setReport(null)}>← RETOUR À LA LISTE</button>
          <AuditReportCard report={report} onCopy={copyText} />
        </>
      ) : rows === null ? (
        <p className="mono muted">chargement…</p>
      ) : (
        <>
          <AuditReportTable rows={rows} onOpen={open} />
          {msg ? <p className="mono muted audit-reports-msg">{msg}</p> : null}
        </>
      )}
    </section>
  )
}
