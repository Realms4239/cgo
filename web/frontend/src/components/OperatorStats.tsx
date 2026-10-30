// Les trois lectures opérateur d'un audit — un seul rendu pour le tableau de
// bord et la vue Auditer. Le coût n'apparaît que là où il est demandé.
import { operatorSummary } from '../lib/dashboard'
import { formatCostAR } from '../lib/format'
import type { AuditReport } from '../lib/types'

export function OperatorStats({
  report,
  warnMs,
  critMs,
  includeCost = false,
}: {
  report: AuditReport | null | undefined
  warnMs: number
  critMs: number
  includeCost?: boolean
}) {
  const s = operatorSummary(report, warnMs, critMs)
  const cost = formatCostAR(report?.economics?.primary?.ar_per_hour)
  return (
    <div className={includeCost ? 'audit-summary' : 'dash-stats'} data-testid="operator-stats">
      <div className="audit-summary-cell">
        <span>Délai sous charge</span>
        <b>{s.delayMs != null ? s.delayMs.toFixed(0) : '—'}<small> ms</small></b>
        <em data-level={s.delay?.level ?? 'none'}>{s.delay?.word ?? 'non mesuré'}</em>
      </div>
      <div className="audit-summary-cell">
        <span>Alertes à l'heure</span>
        <b>{s.alertsPct != null ? `${s.alertsPct.toFixed(1)} %` : '—'}</b>
        <em data-level={s.alerts?.level ?? 'none'}>{s.alerts?.word ?? 'non mesuré'}</em>
      </div>
      <div className="audit-summary-cell">
        <span>Qualité du lien</span>
        <b>{s.grade || '—'}</b>
        <em>{s.quality}</em>
      </div>
      {includeCost ? (
        <div className="audit-summary-cell">
          <span>Perte financière</span>
          <b>{cost.num}<small> {cost.unit}</small></b>
          <em>scénario contractuel</em>
        </div>
      ) : null}
    </div>
  )
}
