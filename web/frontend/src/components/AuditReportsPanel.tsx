// Preuves — les rapports d'audit archivés, citables, ouverts sans nouvel
// endpoint : /api/audit/list puis /api/audit/report?audit_id=.
// La vue reste présentationnelle : les chargements sont injectés par le
// conteneur, et la table demeure montée quand un rapport est ouvert.
import { useEffect, useRef, useState } from 'react'
import { ArchiveTable } from './ArchiveTable'
import type { ArchiveRow } from '../lib/archive'
import { AuditReportCard } from './AuditReportCard'
import { CardHead } from './ui/CardHead'
import { useUIStore } from '../store/ui'
import { auditCitation } from '../lib/citation'
import { asArray } from '../lib/format'
import type { AuditReport } from '../lib/types'

export function AuditReportsView({
  rows,
  error,
  msg,
  report,
  citationMsg,
  onRetry,
  onOpen,
  onCopyCitation,
  onCloseReport,
  onCopy,
}: {
  rows: ArchiveRow[] | null
  error: string | null
  msg: string
  report: AuditReport | null
  citationMsg: string | null
  onRetry: () => void
  onOpen: (id: string) => void
  onCopyCitation: (id: string) => void
  onCloseReport: () => void
  onCopy: (text: string, label: string) => void
}) {
  const proofRef = useRef<HTMLElement>(null)

  useEffect(() => {
    if (report) proofRef.current?.focus()
  }, [report])

  return (
    <section className="card audit-reports" data-testid="audit-reports">
      <CardHead label="Rapports d'audit" sub="archives citables — chaque rapport porte son sceau" />
      {error ? (
        <div className="state-card state-error" role="alert">
          <p className="state-kicker">Liste indisponible</p>
          <p>{error}</p>
          <div className="layout-action">
            <button className="btn btn-primary" onClick={onRetry}>Réessayer</button>
          </div>
        </div>
      ) : rows === null ? (
        <p className="mono muted">chargement…</p>
      ) : (
        <>
          <ArchiveTable rows={rows} onOpen={onOpen} onCopyCitation={onCopyCitation} />
          {msg ? <p className="mono muted audit-reports-msg">{msg}</p> : null}
          {report ? (
            <section
              className="archive-proof"
              role="region"
              aria-label="Rapport ouvert"
              ref={proofRef}
              tabIndex={-1}
            >
              <button className="btn audit-report-back" onClick={onCloseReport}>← RETOUR À LA LISTE</button>
              <AuditReportCard report={report} onCopy={onCopy} />
            </section>
          ) : citationMsg ? (
            <p className="mono muted audit-reports-msg">{citationMsg}</p>
          ) : null}
        </>
      )}
    </section>
  )
}

export default function AuditReportsPanel() {
  const [rows, setRows] = useState<ArchiveRow[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [report, setReport] = useState<AuditReport | null>(null)
  const [msg, setMsg] = useState('')
  const [citationMsg, setCitationMsg] = useState<string | null>(null)
  const wantedCitation = useRef<string | null>(null)

  const load = () => {
    setError(null)
    setRows(null)
    fetch('/api/audit/list')
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(j => setRows(asArray<ArchiveRow>(j.audits)))
      .catch(() => { setRows([]); setError('liste des rapports indisponible — vérifiez que le service tourne, puis réessayez') })
  }

  useEffect(load, [])

  const openReport = (id: string) => {
    setMsg('chargement du rapport…')
    fetch(`/api/audit/report?audit_id=${encodeURIComponent(id)}`)
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(j => { setReport(j as AuditReport); setMsg('') })
      .catch(() => setMsg('rapport indisponible — réessayez'))
  }

  const copyCitation = (id: string) => {
    if (wantedCitation.current === id) return
    wantedCitation.current = id
    setCitationMsg(null)
    fetch(`/api/audit/report?audit_id=${encodeURIComponent(id)}`)
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then((j: AuditReport) => {
        const text = auditCitation(j)
        if (!j.provenance?.source_sha256) {
          setCitationMsg('citation impossible — empreinte source absente')
          return
        }
        if (!text) {
          setCitationMsg('citation impossible — le rapport ne porte pas d’identifiant')
          return
        }
        return navigator.clipboard?.writeText(text).then(
          () => useUIStore.getState().pushToast('Citation copiée', 'ok'),
          () => setCitationMsg('copie impossible — sélectionnez le texte'),
        )
      })
      .catch(() => setCitationMsg('citation impossible — preuve indisponible'))
      .finally(() => { wantedCitation.current = null })
  }

  const copyText = (text: string, label: string) => {
    navigator.clipboard?.writeText(text).then(
      () => useUIStore.getState().pushToast(`${label} copié`, 'ok'),
      () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
    )
  }

  return (
    <AuditReportsView
      rows={rows}
      error={error}
      msg={msg}
      report={report}
      citationMsg={citationMsg}
      onRetry={load}
      onOpen={openReport}
      onCopyCitation={copyCitation}
      onCloseReport={() => setReport(null)}
      onCopy={copyText}
    />
  )
}
