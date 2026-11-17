// Tableau de bord — cockpit 2–3–3 : santé, alertes, coût, puis une action.
// Aucune donnée personnelle n’est inventée : l’absence garde ses mots, le constat reste consultable.
import { useEffect, useState } from 'react'
import { useAuditStore } from '../store/audit'
import { VerdictHero } from '../components/VerdictHero'
import { retailBound } from '../lib/dashboard'
import { qualityWord } from '../lib/dashboard'
import { formatCostAR, groupDigits, relativeTime, asArray } from '../lib/format'
import { loadSettings } from '../lib/settings'
import { auditList } from '../lib/auditList'
import IntegriteView from './IntegriteView'
import AuditReportsPanel from '../components/AuditReportsPanel'
import AuditView from './AuditView'

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

// La santé chargée est dérivée par le moteur Go et injectée par l'API
// (measured.loaded_health) — l'interface affiche, elle ne recalcule pas.
// Un rapport historique sans projection rend l'état absent, comme hier.
function healthBlock(report: import('../lib/types').AuditReport): { title: string; detail: string } {
  const h = report.measured?.loaded_health
  if (!h || h.kind === 'absent') return { title: HEALTH_LABEL.absent, detail: 'la mesure manque' }
  const delta = `Δ ${signed(Math.round(h.bloat_ms))} ms`
  if (h.kind === 'unconfigured') return { title: HEALTH_LABEL.unconfigured, detail: `${delta} — aucun seuil de classe fourni` }
  if (h.kind === 'ok' || h.kind === 'attention' || h.kind === 'critical') {
    return { title: HEALTH_LABEL[h.kind], detail: `${delta} sous charge` }
  }
  // kind inconnu : jamais un titre inventé — l'absence parle.
  return { title: HEALTH_LABEL.absent, detail: 'la mesure manque' }
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

function PreuvesPanel() {
  return (
    <section className="card dash-card layout-span-12 preuves-panel" data-testid="cockpit-preuves">
      <p className="dash-card-kicker">Preuves &amp; Rapports</p>
      <IntegriteView embedded active />
      <AuditReportsPanel />
    </section>
  )
}

function hash8(sha?: string): string {
  return sha && sha.length >= 8 ? sha.slice(0, 8) : '—'
}

function MechanismCard({ report }: { report: import('../lib/types').AuditReport }) {
  const m = report.measured
  const idle = typeof m?.rtt_idle_p50_ms === 'number' ? m.rtt_idle_p50_ms : undefined
  const loaded = typeof m?.rtt_loaded_p95_ms === 'number' ? m.rtt_loaded_p95_ms : undefined
  const delta = (idle != null && loaded != null && loaded >= idle) ? loaded - idle : undefined
  const load = m?.load_measured
  return (
    <details className="mechanism" data-testid="cockpit-mechanism">
      <summary><span>Le mécanisme derrière ce verdict — seuls des faits scellés sont cités</span></summary>
      <ul className="mechanism-list">
        <li><b>Note du moteur.</b> {report.bloat_grade ? `Le rapport est noté ${report.bloat_grade} — ${qualityWord(report.bloat_grade)}.` : 'Pas de note sur ce rapport.'}</li>
        <li><b>Allongement sous charge.</b> {idle != null && loaded != null ? `Le délai sous charge vient à ${loaded.toFixed(0)} ms contre ${idle.toFixed(0)} ms au repos — un écart de ${(delta ?? 0).toFixed(0)} ms dû à l’attente dans les files, pas à la technologie d’accès.` : 'Charge non mesurée — aucun écart ne peut être affirmé.'}</li>
        <li><b>Pourquoi ce verdict.</b> {report.bloat_verdict || 'Le moteur ne rend pas de phrase pour ce rapport.'}</li>
        <li><b>À l’échéance.</b> {report.cap_sla?.measured ? `${report.cap_sla.ok_pct?.toFixed(0) ?? '—'} % de requêtes respectent l’objectif de ${report.cap_sla.target_pct ?? '—'} % — échéance ${report.cap_sla?.deadline_ms} ms.` : 'Échéance non mesurée.'}</li>
        {load ? null : <li><b>Prudence.</b> La charge n’a pas été mesurée ; tout delta affiché ailleurs reste indicatif.</li>}
      </ul>
    </details>
  )
}

function SealFooter({ report }: { report: import('../lib/types').AuditReport }) {
  const prov = report.provenance
  const telemetry = report.kernel_telemetry
  return (
    <footer className="dash-seal" data-testid="cockpit-seal" aria-label="Sceau du rapport">
      <span className="dash-seal-item">sceau <b>{hash8(prov?.source_sha256)}</b></span>
      <span className="dash-seal-item">rapport <b>{report.audit_id}</b></span>
      {telemetry?.source ? (
        <span className="dash-seal-item">télémétrie <b>{telemetry.source === 'ebpf' ? 'eBPF' : telemetry.source}</b>{telemetry.object_sha256 ? <> · objet <b>{hash8(telemetry.object_sha256)}</b></> : null}</span>
      ) : null}
      <span className="dash-seal-item">scellé <b>{report.timestamp?.slice(0, 16).replace('T', ' ') || '—'}</b></span>
    </footer>
  )
}

function InlineAudit({ onClose }: { onClose: () => void }) {
  return (
    <section className="inline-audit" aria-label="audit inline" data-testid="cockpit-audit-inline">
      <div className="inline-audit-bar">
        <p className="dash-card-kicker">Audit du lien</p>
        <button className="btn btn-mini" onClick={onClose}>FERMER</button>
      </div>
      <AuditView />
    </section>
  )
}

export default function CockpitView() {
  const report = useAuditStore(s => s.report)
  const showReport = useAuditStore(s => s.showReport)
  const settings = loadSettings()
  const [archives, setArchives] = useState<ArchiveRecord[] | null>(null)
  const [auditOpen, setAuditOpen] = useState(false)

  useEffect(() => {
    auditList()
      .then(j => setArchives(asArray<ArchiveRecord>(j.audits).slice(0, 5)))
      .catch(() => setArchives([]))
  }, [])

  if (!report) {
      return (
      <div className="syn dash dash-absent-page">
        <section className="layout-grid dash-absent-grid">
          <div className="layout-span-7 state-card state-absent" data-testid="dash-absent">
            <p className="state-kicker">Site — mesure en attente</p>
            <h1>Ce site n’a pas encore été mesuré</h1>
            <p>L’audit de 60 secondes mesure la latence au repos puis sous charge; il prouve ou échoue, rien entre les deux.</p>
            <div className="layout-action">
              <button className="btn btn-primary" onClick={() => setAuditOpen(true)}>Lancer l’audit de 60 secondes</button>
              <button className="btn" onClick={() => setAuditOpen(true)}>Voir un rapport archivé</button>
            </div>
          </div>
          <aside className="layout-span-5 state-card state-explain" aria-label="Ce que l’audit établit">
            <p className="state-kicker">Ce que l’audit établit</p>
            <ol className="absent-steps">
              <li><span>01</span><div><b>Repos — 10 s</b><p>latence de base, sans charge artificielle.</p></div></li>
              <li><span>02</span><div><b>Charge — 40 s</b><p>débit atteint, pertes et aggravation du délai sous charge.</p></div></li>
              <li><span>03</span><div><b>Vidange — 10 s</b><p>retour au repos, pour voir si les files se vident.</p></div></li>
            </ol>
            <p className="dash-card-detail">Aucune valeur n’est inventée avant ces mesures; les preuves archivées restent disponibles dans Preuves &amp; Rapports.</p>
          </aside>
        </section>
        <PreuvesPanel />
        {auditOpen ? <InlineAudit onClose={() => setAuditOpen(false)} /> : null}
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
          <h1 className="sr-only">Tableau de bord</h1>
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
          <p className="dash-card-kicker">Respect de l’échéance</p>
          <p className="dash-card-value">{okPct != null ? <>{okPct.toFixed(1)}<small> %</small></> : 'Mesure absente'}</p>
          <p className="dash-card-detail">{target != null ? `objectif ${target} %` : 'l’objectif manque'}</p>
        </section>

        <CapCard report={report} />

        <section className="card dash-action layout-span-8" data-testid="dash-action">
          <p className="dash-card-kicker">Action recommandée</p>
          <p className="dash-action-title">{actionIsSafe ? 'Appliquer le réglage recommandé.' : 'Observer d’abord — aucun réglage n’est prêt.'}</p>
          <p className="dash-card-detail">{actionIsSafe ? 'La mesure supporte le réglage, pas une promesse.' : 'La charge ne permet pas de transformer cette mesure en commande réseau.'}</p>
          <button className="btn btn-primary" onClick={() => { showReport(); setAuditOpen(true) }}>Voir le constat</button>
        </section>

        <section className="card dash-recent layout-span-4" data-testid="dash-recent">
          <p className="dash-card-kicker">Derniers constats</p>
          <ul className="dash-recent-list">
            {rows.slice(0, 4).map((row, index) => (
              <li key={row.audit_id || `${row.site}-${index}`}>
                <span>{row.site || 'Site'} · {row.timestamp?.slice(0, 10) || '—'}</span>
                <button className="btn btn-mini" onClick={() => { showReport(); setAuditOpen(true) }}>Ouvrir</button>
              </li>
            ))}
          </ul>
        </section>

        <section className="card dash-card layout-span-12" data-testid="dash-cost">
          <p className="dash-card-kicker">Détail contractuel</p>
          <p className="dash-card-detail">{report?.economics?.disclaimer || 'Scénarios contractuels, pas des mesures.'} {report.economics?.primary ? `Scénario principal : ${cost.num} ${cost.unit}.` : ''}</p>
        </section>
        <MechanismCard report={report} />
        <PreuvesPanel />
        <SealFooter report={report} />
        {auditOpen ? <InlineAudit onClose={() => setAuditOpen(false)} /> : null}
      </section>
    </div>
  )
}
