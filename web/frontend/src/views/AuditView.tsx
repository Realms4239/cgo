// Auditer — une seule destination : déclarer, observer, lire. La session vit
// dans le store racine ; quitter cette vue ne coupe ni le suivi ni le rapport.
import { useEffect, useState } from 'react'
import { AuditDeclare } from '../components/AuditDeclare'
import { AuditLive } from '../components/AuditLive'
import { gradeTone } from '../components/AuditReportCard'
import AuditHistoryModal from '../components/AuditHistoryModal'
import { ModeHead } from '../components/ModeHead'
import { RemediationCard } from '../components/RemediationCard'
import { VerdictHero } from '../components/VerdictHero'
import { Badge } from '../components/ui/StatusPip'
import { CardHead } from '../components/ui/CardHead'
import PhaseStepper from '../components/PhaseStepper'
import { selectAuditViewState } from '../lib/viewState'
import { loadSettings } from '../lib/settings'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'

export default function AuditView() {
  const s = useAuditStore()
  const [historyOpen, setHistoryOpen] = useState(false)
  const settings = loadSettings()

  useEffect(() => {
    useAuditStore.getState().open()
  }, [])

  const snapshot = selectAuditViewState({
    stage: s.stage,
    live: s.live,
    report: s.report,
    failure: s.failure ?? null,
  })
  const report = snapshot.status === 'complete' ? snapshot.report : s.report
  const remediation = report?.remediation
  const currentPhase = snapshot.status === 'running' ? snapshot.phase : 'repos'
  const stepIndex = snapshot.status === 'complete' ? 3 : snapshot.status === 'running' ? 2 : 1

  return (
    <div className="panel-stack audit-page">
      <ModeHead
        title="Auditer le lien réel"
        sub="Sans droits administrateur · aucun CSV partiel."
        actions={(
          <>
            <button className="btn" onClick={() => setHistoryOpen(true)} title="audits archivés : rapprochement au référentiel, import par ligne">
              HISTORIQUE{s.historyN !== null ? ` (${s.historyN})` : ''}
            </button>
            {s.stage === 'idle' ? <button className="btn" onClick={s.open}>OUVRIR LA DÉCLARATION</button> : null}
          </>
        )}
      />

      <ol className="audit-steps mono" aria-label="étapes de l'audit">
        {['Déclarer', 'Observer', 'Lire'].map((label, i) => {
          const active = stepIndex === i + 1
          return (
            <li key={label} aria-current={active ? 'step' : undefined} className={active ? 'on' : ''}>
              {i + 1}. {label}
            </li>
          )
        })}
      </ol>

      {snapshot.status === 'error' ? (
        <section className="state-card state-error" data-testid="audit-error">
          <p className="state-kicker">Lecture interrompue</p>
          <h2>La preuve n’a pas pu être relue</h2>
          <p>{snapshot.message}</p>
          <div className="layout-action">
            <button className="btn btn-primary" onClick={snapshot.retryAction}>Réessayer</button>
            <button className="btn" onClick={s.restart}>Retour à la déclaration</button>
          </div>
        </section>
      ) : null}

      {snapshot.status === 'loading_stale' ? (
        <section className="state-card state-stale" data-testid="audit-stale">
          <p className="state-kicker">Lecture en cours</p>
          <h2>Un constat existe, mais il n’est pas relu</h2>
          <p>Les données déjà disponibles restent historiques; elles ne sont pas lues comme une mesure active.</p>
          <div className="layout-action">
            <button className="btn btn-primary" onClick={s.showReport}>Voir le dernier constat</button>
          </div>
        </section>
      ) : null}

      {snapshot.status === 'idle' ? (
        <section className="card audit-idle">
          <CardHead label="Déclarer un audit" />
          <p>Cette destination mesure la latence au repos, puis sous charge. Elle ne fabrique pas de résultat intermédiaire.</p>
          <button className="btn btn-primary" onClick={s.open}>DÉCLARER</button>
          {snapshot.lastAudit ? (
            <button className="btn" onClick={s.showReport}>VOIR LE DERNIER CONSTAT</button>
          ) : null}
        </section>
      ) : null}

      {s.stage === 'declared' && snapshot.status !== 'error' ? (
        <section className="card" data-testid="audit-declare-card">
          <CardHead label="Déclarer" sub="site, type de lien, sens" />
          <AuditDeclare />
          {s.report ? (
            <button className="btn audit-last-report" onClick={s.showReport}>
              VOIR LE DERNIER CONSTAT — {s.report.site || 'site —'}
            </button>
          ) : null}
        </section>
      ) : null}

      {snapshot.status === 'running' ? (
        <section className="card" data-testid="audit-observe-card">
          <CardHead label="Observer" sub="phases, restant, échantillons de la fenêtre" />
          <PhaseStepper phase={currentPhase} />
          <AuditLive live={s.live} direction={s.direction} onStop={s.stop} />
          {s.msg ? <span className="mono muted audit-msg">{s.msg}</span> : null}
        </section>
      ) : null}

      {snapshot.status === 'complete' && report ? (
        <section className="card audit-read" data-testid="audit-read-card">
          <CardHead
            label="Lire"
            sub={`${report.site || 'site —'} · ${report.link_type || 'lien —'}${report.timestamp ? ` · ${report.timestamp}` : ''}`}
            right={<Badge tone={gradeTone(report.bloat_grade)}>{report.bloat_grade || 'non noté'}</Badge>}
          />
          <VerdictHero report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
          {report.bloat_verdict ? <p className="audit-verdict">{report.bloat_verdict}</p> : null}

          {remediation && report.remediation_state === 'available' ? (
            <RemediationCard remediation={remediation} state="available" />
          ) : (
            <p className="mono muted">Observation seule — aucun réglage n’est prêt à être appliqué.</p>
          )}

          <div className="audit-launch">
            <button className="btn" onClick={s.restart}>NOUVEL AUDIT</button>
            <button className="btn" onClick={() => useUIStore.getState().openProof('proof')}>VOIR LES PREUVES</button>
            <span className="mono muted">{s.msg}</span>
          </div>
        </section>
      ) : null}

      {snapshot.status === 'complete' && !report ? (
        <section className="card">
          <p className="mono muted">Rapport absent — relancez un audit.</p>
          <button className="btn" onClick={s.restart}>NOUVEL AUDIT</button>
        </section>
      ) : null}

      {historyOpen ? (
        <AuditHistoryModal
          onClose={() => {
            setHistoryOpen(false)
            useAuditStore.getState()._history()
          }}
          onImported={msg => useUIStore.getState().pushToast(msg, 'ok')}
        />
      ) : null}
    </div>
  )
}
