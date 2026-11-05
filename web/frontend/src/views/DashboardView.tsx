// Tableau de bord — une colonne : verdict, remédiation, coût, variation.
// Aucune étiquette de simulation n'occupe l'accueil ; les détails des rapports
// vivent dans le tiroir de preuve.
import { useUIStore } from '../store/ui'
import { useAuditStore } from '../store/audit'
import { CardHead } from '../components/ui/CardHead'
import { RemediationCard } from '../components/RemediationCard'
import { VerdictHero } from '../components/VerdictHero'
import { bothAbsent, retailBound } from '../lib/dashboard'
import { compareAudits, summaryOf } from '../lib/auditSummary'
import { formatCostAR, groupDigits, relativeTime } from '../lib/format'
import { loadSettings } from '../lib/settings'

const signed = (v: number) => `${v > 0 ? '+' : ''}${v}`

export default function DashboardView() {
  const setMode = useUIStore(s => s.setMode)
  const report = useAuditStore(s => s.report)
  const previous = useAuditStore(s => s.previous)
  const showReport = useAuditStore(s => s.showReport)
  const settings = loadSettings()

  const primary = report?.economics?.primary
  const cost = formatCostAR(primary?.ar_per_hour)
  const retail = retailBound(report)
  const currentSummary = report ? summaryOf(report) : null
  const change = compareAudits(previous, currentSummary)
  const identity = report
    ? [report.site || 'site —', report.link_type || 'lien —', report.timestamp ? report.timestamp.slice(0, 16).replace('T', ' ') : '']
      .filter(Boolean)
      .join(' · ')
    : ''

  return (
    <div className="syn dash">
      <header className="syn-hero">
        {identity ? (
          <p className="dash-iden mono">
            {identity}
            {report?.timestamp ? <span className="dash-ago"> · {relativeTime(report.timestamp)}</span> : null}
          </p>
        ) : null}
        <VerdictHero report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
        <div className="dash-hero-action">
          {report ? (
            <button className="btn btn-primary" onClick={() => { showReport(); setMode('audit') }}>Ouvrir le rapport</button>
          ) : (
            <button className="btn btn-primary" onClick={() => setMode('audit')}>Auditer un lien</button>
          )}
        </div>
      </header>

      <RemediationCard remediation={report?.remediation} state={report?.remediation_state ?? (report?.remediation ? "available" : "observation_only")} />

      {report ? (
        bothAbsent(!!primary, !!(change.comparable && currentSummary)) ? (
          <section className="card dash-quiet" data-testid="dash-quiet">
            <p className="syn-read"><b>Ce que ça coûte</b> — non mesuré : aucun coût n’est chiffré plutôt qu’un zéro fabriqué.</p>
            <p className="syn-read"><b>Ce qui a changé</b> — aucune comparaison disponible{change.reason ? ` — ${change.reason}` : ''}. Il faut deux audits du même site et du même type de lien.</p>
          </section>
        ) : (
          <>
            <section className="card dash-card" data-testid="dash-cost">
              <CardHead label="Ce que ça coûte" sub="scénario contractuel, pas une mesure" />
              {primary ? (
                <>
                  <div className="dash-cost">
                    <span>Perte financière</span>
                    <b>{cost.num}<small> {cost.unit}</small></b>
                    <em>{primary.name} · scénario primaire</em>
                  </div>
                  {retail != null ? (
                    <p className="syn-read">
                      Borne haute hors-forfait : <b>{groupDigits(retail, 2)} Ar/h</b> — une pénalité de secours,
                      pas un modèle d'infrastructure.
                    </p>
                  ) : null}
                  <p className="dash-disclaimer">{report?.economics?.disclaimer || 'Scénarios contractuels, pas des mesures.'}</p>
                </>
              ) : (
                <p className="syn-read">Non mesuré — aucun coût n'est chiffré plutôt qu'un zéro fabriqué.</p>
              )}
            </section>

            <section className="card dash-card" data-testid="dash-change">
              <CardHead label="Ce qui a changé" sub="comparaison locale avec l'audit précédent" />
              {change.comparable && currentSummary ? (
                <>
                  <div className="dash-change">
                    {change.smallDeltaPct != null ? (
                      <div>
                        <span>Délai sous charge</span>
                        <b data-dir={change.smallDeltaPct > 0 ? 'bad' : change.smallDeltaPct < 0 ? 'good' : 'same'}>
                          {signed(change.smallDeltaPct)} %
                        </b>
                      </div>
                    ) : null}
                    {change.deadlineDelta != null ? (
                      <div>
                        <span>Alertes à l'heure</span>
                        <b data-dir={change.deadlineDelta > 0 ? 'good' : change.deadlineDelta < 0 ? 'bad' : 'same'}>
                          {signed(change.deadlineDelta)} pts
                        </b>
                      </div>
                    ) : null}
                  </div>
                  <p className="syn-read">
                    {previous?.timestamp ? `Audit précédent du ${previous.timestamp.slice(0, 10)}` : 'Audit précédent'} ·{' '}
                    {previous?.grade ? `qualité ${previous.grade} → ${currentSummary.grade || '—'}` : `qualité ${currentSummary.grade || '—'}`}.
                  </p>
                </>
              ) : (
                <p className="syn-read">
                  Aucune comparaison disponible{change.reason ? ` — ${change.reason}` : ''}. Il faut deux audits du
                  même site et du même type de lien.
                </p>
              )}
            </section>
          </>
        )
      ) : null}
    </div>
  )
}
