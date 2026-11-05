// Tableau de bord — cockpit 2–3–3 : santé, alertes, coût, puis une action.
// Aucune donnée personnelle n’est inventée : l’absence garde ses mots, le constat reste consultable.
import { useEffect, useState } from 'react'
import { useUIStore } from '../store/ui'
import { useAuditStore } from '../store/audit'
import { VerdictHero } from '../components/VerdictHero'
import { evaluateLoadedHealth } from '../lib/health'
import { retailBound } from '../lib/dashboard'
import { formatCostAR, groupDigits, relativeTime, asArray } from '../lib/format'
import { loadSettings } from '../lib/settings'

const signed = (v: number) => `${v > 0 ? '+' : ''}${v}`

const HEALTH_LABEL = {
  ok: 'Lien prêt',
  attention: 'À surveiller',
  critical: 'Critique sous charge',
  absent: 'Mesure absente',
  unconfigured: 'Seuil non configuré',
} as const

type ArchiveRecord = {
  audit_id?: string
  site?: string
  link_type?: string
  timestamp?: string
  has_report?: string
  site_name?: string
}

function healthBlock(report: import('../lib/types').AuditReport): { title: string; detail: string } {
  const h = evaluateLoadedHealth({
    linkType: report.link_type ?? '',
    idleP50Ms: report.measured?.rtt_idle_p50_ms,
    loadedP95Ms: report.measured?.rtt_loaded_p95_ms,
  })
  if (h.kind === 'absent') return { title: HEALTH_LABEL.absent, detail: 'la mesure manque' }
  if (h.kind === 'unconfigured') return { title: HEALTH_LABEL.unconfigured, detail: `Δ ${Math.round(h.bloatMs)} ms — aucun seuil de classe fourni` }
  return { title: HEALTH_LABEL[h.kind], detail: `Δ ${signed(Math.round(h.bloatMs))} ms sous charge` }
}

function CapCard({ report }: { report: import('../lib/types').AuditReport | null }) {
  const cost = report?.economics?.primary
  const parsed = formatCostAR(cost?.ar_per_hour)
  const retail = retailBound(report)
  return (
    <section className="card dash-card layout-span-4 dash-capacity" data-testid="dash-capacity">
      <p className="dash-card-kicker">Capacité gaspillée</p>
      <p className="dash-card-value">{cost ? <>{parsed.num}<small> {parsed.unit}</small></> : 'Non chiffré'}</p>
      <p className="dash-card-detail">{cost ? `${cost.name} · estimation contractuelle` : 'le scénario contractuel manque'}</p>
      {retail != null ? <p className="dash-card-detail">Borne secours : {groupDigits(retail, 2)} Ar/h — pas un modèle fixe</p> : null}
    </section>
  )
}

export default function DashboardView() {
  const setMode = useUIStore(s => s.setMode)
  const report = useAuditStore(s => s.report)
  const showReport = useAuditStore(s => s.showReport)
  const settings = loadSettings()
  const [archives, setArchives] = useState<ArchiveRecord[] | null>(null)

  useEffect(() => {
    fetch('/api/audit/list')
      .then(r => r.json())
      .then(j => setArchives(asArray<ArchiveRecord>(j.audits).slice(0, 5)))
      .catch(() => setArchives([]))
  }, [])

  if (!report) {
    return (
      <div className="syn dash">
        <section className="layout-grid">
          <div className="layout-span-12 state-card state-absent" data-testid="dash-absent">
            <p className="state-kicker">Site — mesure en attente</p>
            <h1>Ce site n’a pas encore été mesuré</h1>
            <p>L’audit de 60 secondes mesure la latence au repos puis sous charge; il prouve ou échoue, rien entre les deux.</p>
            <div className="layout-action">
              <button className="btn btn-primary" onClick={() => setMode('audit')}>Lancer l’audit de 60 secondes</button>
            </div>
          </div>
        </section>
      </div>
    )
  }

  const health = healthBlock(report)
  const cost = formatCostAR(report.economics?.primary?.ar_per_hour)
  const target = report.cap_sla?.target_pct
  const okPct = report.cap_sla?.measured ? report.cap_sla.ok_pct : undefined
  const actionIsSafe = !!report.remediation && report.remediation_state === 'available' && report.remediation_available !== false
  const rows = archives ?? [{ audit_id: report.audit_id, site: report.site, link_type: report.link_type, timestamp: report.timestamp, has_report: 'true' }]

  return (
    <div className="syn dash">
      <section className="layout-grid">
        <div className="layout-span-12 dash-overview-head">
          <p className="dash-iden mono">
            {[report.site || 'Site', report.link_type || 'lien —', report.timestamp ? report.timestamp.slice(0, 16).replace('T', ' ') : ''].filter(Boolean).join(' · ')}
            {report.timestamp ? <span className="dash-ago"> · {relativeTime(report.timestamp)}</span> : null}
          </p>
          <VerdictHero report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
        </div>

        <section className="card dash-card layout-span-4 dash-gauge" data-testid="dash-gauge">
          <p className="dash-card-kicker">Santé du lien</p>
          <p className="dash-card-value">{health.title}</p>
          <p className="dash-card-detail">{health.detail}</p>
        </section>

        <section className="card dash-card layout-span-4 dash-alerts" data-testid="dash-alerts">
          <p className="dash-card-kicker">Alertes à l’heure</p>
          <p className="dash-card-value">{okPct != null ? <>{okPct.toFixed(1)}<small> %</small></> : 'Mesure absente'}</p>
          <p className="dash-card-detail">{target != null ? `objectif ${target} %` : 'l’objectif manque'}</p>
        </section>

        <CapCard report={report} />

        <section className="card dash-action layout-span-8" data-testid="dash-action">
          <p className="dash-card-kicker">Action recommandée</p>
          <p className="dash-action-title">{actionIsSafe ? 'Appliquer le réglage recommandé.' : 'Observer d’abord — aucun réglage n’est prêt.'}</p>
          <p className="dash-card-detail">{actionIsSafe ? 'La mesure supporte le réglage, pas une promesse.' : 'La charge ne permet pas de transformer cette mesure en commande réseau.'}</p>
          <button className="btn btn-primary" onClick={() => { showReport(); setMode('audit') }}>Voir le constat</button>
        </section>

        <section className="card dash-recent layout-span-4" data-testid="dash-recent">
          <p className="dash-card-kicker">Derniers constats</p>
          <ul className="dash-recent-list">
            {rows.slice(0, 4).map((row, index) => (
              <li key={row.audit_id || `${row.site}-${index}`}>
                <span>{row.site || 'Site'} · {row.timestamp?.slice(0, 10) || '—'}</span>
                <button className="btn btn-mini" onClick={() => { showReport(); setMode('audit') }}>Ouvrir</button>
              </li>
            ))}
          </ul>
        </section>

        <section className="card dash-card layout-span-12" data-testid="dash-cost">
          <p className="dash-card-kicker">Détail contractuel</p>
          <p className="dash-card-detail">{report?.economics?.disclaimer || 'Scénarios contractuels, pas des mesures.'} {report.economics?.primary ? `Scénario principal : ${cost.num} ${cost.unit}.` : ''}</p>
        </section>
      </section>
    </div>
  )
}
