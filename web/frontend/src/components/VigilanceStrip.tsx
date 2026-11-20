// Bandeau de vigilance — onglet rouge et état réel de la liaison.
// Aucune alerte n'est fabriquée : sans mesure, le bandeau le dit simplement.
import type { ReactNode } from 'react'

export default function VigilanceStrip({
  status,
  detail,
  action,
}: {
  status: string
  detail: string
  action?: ReactNode
}) {
  return (
    <section className="vig-strip" aria-label="Vigilance">
      <div className="vig-tab">
        <span aria-hidden="true">⚠</span>
        <span>Vigilance</span>
      </div>
      <p className="vig-body">
        <b>{status}</b>
        <span>{detail}</span>
      </p>
      {action ? <div className="vig-link">{action}</div> : null}
    </section>
  )
}
