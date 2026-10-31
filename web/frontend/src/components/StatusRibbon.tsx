// Bandeau d'état — trois lectures ponctuelles, jamais un sondage en boucle.
// Une valeur absente s'écrit « — » et se réessaie ; aucun compte inventé.
import { useEffect, useState } from 'react'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'
import { remainingSeconds } from '../lib/auditFlow'
import { auditFacts, ribbonMode } from '../lib/ribbon'

export default function StatusRibbon() {
  const connected = useUIStore(s => s.connected)
  const sseStatus = useUIStore(s => s.sseStatus)
  const setMode = useUIStore(s => s.setMode)
  const report = useAuditStore(s => s.report)
  const stage = useAuditStore(s => s.stage)
  const auditLive = useAuditStore(s => s.live)
  const [mode, setModeName] = useState('')
  const [failed, setFailed] = useState(false)

  // Le bandeau ne porte que l'état et le dernier audit. Les comptes du corpus,
  // le sceau et le palier de télémétrie vivent dans Preuves et le rapport : un
  // technicien n'a pas à les décoder avant d'agir. Une lecture, une requête.
  const load = () => {
    fetch('/api/health')
      .then(r => (r.ok ? r.json() : Promise.reject(new Error(String(r.status)))))
      .then(j => { setModeName((j as any)?.mode ?? ''); setFailed(false) })
      .catch(() => setFailed(true))
  }

  useEffect(() => {
    load()
  }, [connected, report?.audit_id])

  const m = ribbonMode(mode)
  const auditing = stage === 'observing'

  return (
    <div className="status-ribbon" aria-label="État de la mesure" data-testid="status-ribbon">
      <span className={'ribbon-item ribbon-mode' + (m.observation ? ' observation' : '')}>{m.text}</span>
      <span className="ribbon-sep" aria-hidden="true" />
      <span className={'ribbon-item' + (connected ? ' ok' : ' off')}>
        SSE {sseStatus}
      </span>
      {auditing ? (
        // Seul fait réellement temporel du bandeau : l'audit en cours.
        // L'annoncer seul évite de relire les comptes du corpus à chaque
        // rechargement — un lecteur d'écran n'a pas à entendre le décor.
        <span className="ribbon-item live" role="status" aria-live="polite">
          Audit en cours{auditLive?.phase ? ` · ${auditLive.phase}` : ''}
          {remainingSeconds(auditLive) != null ? ` · ${remainingSeconds(auditLive)} s` : ''}
        </span>
      ) : null}
      <span className="ribbon-item">{auditFacts(report)}</span>
      {failed ? (
        <button type="button" className="btn ribbon-retry" onClick={load} title="relire l'état du serveur">
          RÉESSAYER
        </button>
      ) : null}
      <button type="button" className="btn ribbon-verify" onClick={() => setMode('evidence')}>
        VÉRIFIER
      </button>
    </div>
  )
}
