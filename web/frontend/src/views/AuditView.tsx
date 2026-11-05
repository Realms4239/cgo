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
import { auditStep } from '../lib/auditFlow'
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

  const report = s.report
  const remediation = report?.remediation

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
          const active = auditStep(s.stage) === i + 1
          return (
            <li key={label} aria-current={active ? 'step' : undefined} className={active ? 'on' : ''}>
              {i + 1}. {label}
            </li>
          )
        })}
      </ol>

      {s.stage === 'idle' ? (
        <section className="card audit-idle">
          <CardHead label="Déclarer un audit" />
          <p className="mono muted">
            Quatre décisions suffisent : le site, le type de lien, la durée et le sens. Les champs
            avancés restent repliés.
          </p>
          <button className="btn btn-primary" onClick={s.open}>DÉCLARER</button>
          {report ? (
            <button className="btn" onClick={s.showReport}>VOIR LE DERNIER RAPPORT</button>
          ) : null}
        </section>
      ) : null}

      {s.stage === 'declared' ? (
        <section className="card" data-testid="audit-declare-card">
          <CardHead label="Déclarer" sub="presets, quatre décisions, avancé replié" />
          <AuditDeclare />
          {report ? (
            <button className="btn audit-last-report" onClick={s.showReport}>
              VOIR LE DERNIER RAPPORT — {report.site || 'site —'}
            </button>
          ) : null}
        </section>
      ) : null}

      {s.stage === 'observing' ? (
        <section className="card" data-testid="audit-observe-card">
          <CardHead label="Observer" sub="phases, restant, échantillons de la fenêtre" />
          <AuditLive live={s.live} direction={s.direction} onStop={s.stop} />
          {s.msg ? <span className="mono muted audit-msg">{s.msg}</span> : null}
        </section>
      ) : null}

      {s.stage === 'done' && report ? (
        <section className="card audit-read" data-testid="audit-read-card">
          <CardHead
            label="Lire"
            sub={`${report.site || 'site —'} · ${report.link_type || 'lien —'}${report.timestamp ? ` · ${report.timestamp}` : ''}`}
            right={<Badge tone={gradeTone(report.bloat_grade)}>{report.bloat_grade || 'non noté'}</Badge>}
          />
          <VerdictHero report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
          {report.bloat_verdict ? <p className="audit-verdict">{report.bloat_verdict}</p> : null}

          {remediation ? (
            <RemediationCard remediation={remediation} state={report.remediation_state ?? (remediation ? "available" : "observation_only")} />
          ) : (
            <p className="mono muted">Aucune remédiation émise — capacité ou RTT repos non mesurés.</p>
          )}

          <div className="audit-launch">
            <button className="btn" onClick={s.restart}>NOUVEL AUDIT</button>
            <button className="btn" onClick={() => useUIStore.getState().openProof('proof')}>VOIR LES PREUVES</button>
            <span className="mono muted">{s.msg}</span>
          </div>
        </section>
      ) : null}

      {s.stage === 'done' && !report ? (
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
