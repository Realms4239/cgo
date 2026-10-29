// Auditer — une seule destination : déclarer, observer, lire. La session vit
// dans le store racine ; quitter cette vue ne coupe ni le suivi ni le rapport.
import { useEffect, useState } from 'react'
import { AuditDeclare } from '../components/AuditDeclare'
import { AuditLive } from '../components/AuditLive'
import { AuditReportCard, gradeTone } from '../components/AuditReportCard'
import AuditHistoryModal from '../components/AuditHistoryModal'
import { ScienceDetails } from '../components/ScienceDetails'
import { Badge } from '../components/ui/StatusPip'
import { CardHead } from '../components/ui/CardHead'
import { auditStep } from '../lib/auditFlow'
import { alertsBand, delayBand, qualityWord } from '../lib/dashboard'
import { formatCostAR } from '../lib/format'
import { loadSettings } from '../lib/settings'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'

const fmtMs = (v?: number) => (v != null && v > 0 ? v.toFixed(0) : '—')

export default function AuditView() {
  const s = useAuditStore()
  const [historyOpen, setHistoryOpen] = useState(false)
  const settings = loadSettings()

  useEffect(() => {
    useAuditStore.getState().open()
  }, [])

  const copyText = (text: string, label: string) => {
    navigator.clipboard?.writeText(text).then(
      () => useUIStore.getState().pushToast(`${label} copié`, 'ok'),
      () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
    )
  }

  const report = s.report
  const m = report?.measured
  const small = m?.small_measured ? m?.http_small_p95_ms : m?.rtt_loaded_p95_ms
  const delay = delayBand(small, settings.warnMs, settings.critMs)
  const alerts = report?.cap_sla?.measured ? alertsBand(report.cap_sla.ok_pct) : null
  const primary = report?.economics?.primary
  const cost = formatCostAR(primary?.ar_per_hour)
  const remediation = report?.remediation

  return (
    <div className="panel-stack audit-page">
      <header className="mode-head">
        <div>
          <h1>Auditer le lien réel</h1>
          <p>Mesuré sur le lien réel, sans modification — aucun droit administrateur, aucun CSV partiel.</p>
        </div>
        <div className="mode-head-actions">
          <button className="btn" onClick={() => setHistoryOpen(true)} title="audits archivés : rapprochement au référentiel, import par ligne">
            HISTORIQUE{s.historyN !== null ? ` (${s.historyN})` : ''}
          </button>
          {s.stage === 'idle' ? (
            <button className="btn" onClick={s.open}>OUVRIR LA DÉCLARATION</button>
          ) : null}
        </div>
      </header>

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
          <div className="audit-summary">
            <div className="audit-summary-cell">
              <span>Délai sous charge</span>
              <b>{fmtMs(small)}<small> ms</small></b>
              <em data-level={delay?.level ?? 'none'}>{delay?.word ?? 'non mesuré'}</em>
            </div>
            <div className="audit-summary-cell">
              <span>Alertes à l'heure</span>
              <b>{report.cap_sla?.measured ? `${fmtMs(report.cap_sla.ok_pct)} %` : '—'}</b>
              <em data-level={alerts?.level ?? 'none'}>{alerts?.word ?? 'non mesuré'}</em>
            </div>
            <div className="audit-summary-cell">
              <span>Qualité du lien</span>
              <b>{report.bloat_grade || '—'}</b>
              <em>{qualityWord(report.bloat_grade)}</em>
            </div>
            <div className="audit-summary-cell">
              <span>Perte financière</span>
              <b>{cost.num}<small> {cost.unit}</small></b>
              <em>scénario contractuel</em>
            </div>
          </div>
          {report.bloat_verdict ? <p className="audit-verdict">{report.bloat_verdict}</p> : null}

          {remediation ? (
            <div className="audit-remedy">
              <span className="mono muted">Configuration à appliquer au bord — copiée du rapport scellé</span>
              <div className="audit-remedy-actions">
                {remediation.routeros ? (
                  <button className="btn btn-primary" onClick={() => copyText(remediation.routeros ?? '', 'Configuration MikroTik')}>
                    COPIER LA CONFIGURATION MIKROTIK
                  </button>
                ) : null}
                {remediation.linux_tc?.length ? (
                  <button className="btn" onClick={() => copyText((remediation.linux_tc ?? []).join('\n'), 'Configuration Linux')}>
                    COPIER LA CONFIGURATION LINUX
                  </button>
                ) : null}
              </div>
            </div>
          ) : (
            <p className="mono muted">Aucune remédiation émise — capacité ou RTT repos non mesurés.</p>
          )}

          <ScienceDetails label="Détails scientifiques du rapport" testid="audit-report-details">
            <AuditReportCard report={report} onCopy={copyText} />
          </ScienceDetails>

          <div className="audit-launch">
            <button className="btn" onClick={s.restart}>NOUVEL AUDIT</button>
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
