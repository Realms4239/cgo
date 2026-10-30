// Bandeau d'état — trois lectures ponctuelles, jamais un sondage en boucle.
// Une valeur absente s'écrit « — » et se réessaie ; aucun compte inventé.
import { useEffect, useState } from 'react'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'
import { remainingSeconds } from '../lib/auditFlow'
import { auditFacts, corpusFacts, ribbonMode, sealFacts, type IntegrityFacts } from '../lib/ribbon'

export default function StatusRibbon() {
  const connected = useUIStore(s => s.connected)
  const sseStatus = useUIStore(s => s.sseStatus)
  const setMode = useUIStore(s => s.setMode)
  const report = useAuditStore(s => s.report)
  const stage = useAuditStore(s => s.stage)
  const auditLive = useAuditStore(s => s.live)
  const [mode, setModeName] = useState('')
  const [integrity, setIntegrity] = useState<IntegrityFacts | null>(null)
  const [failed, setFailed] = useState(false)

  const load = () => {
    Promise.allSettled([
      fetch('/api/health').then(r => (r.ok ? r.json() : Promise.reject(new Error(String(r.status))))),
      fetch('/api/integrity').then(r => (r.ok ? r.json() : Promise.reject(new Error(String(r.status))))),
    ]).then(([h, i]) => {
      const hOk = h.status === 'fulfilled'
      const iOk = i.status === 'fulfilled'
      if (hOk) setModeName((h.value as any)?.mode ?? '')
      if (iOk) setIntegrity(i.value as IntegrityFacts)
      setFailed(!hOk || !iOk)
    })
  }

  useEffect(() => {
    load()
  }, [connected, report?.audit_id])

  const m = ribbonMode(mode)
  const auditing = stage === 'observing'

  return (
    <div className="status-ribbon" role="status" aria-label="État de la mesure" data-testid="status-ribbon">
      <span className={'ribbon-item ribbon-mode' + (m.observation ? ' observation' : '')}>{m.text}</span>
      <span className="ribbon-sep" aria-hidden="true" />
      <span className={'ribbon-item' + (connected ? ' ok' : ' off')}>
        SSE {sseStatus}
      </span>
      {auditing ? (
        <span className="ribbon-item live">
          Audit en cours{auditLive?.phase ? ` · ${auditLive.phase}` : ''}
          {remainingSeconds(auditLive) != null ? ` · ${remainingSeconds(auditLive)} s` : ''}
        </span>
      ) : null}
      <span className="ribbon-item">{auditFacts(report)}</span>
      <span className="ribbon-item ribbon-corpus">{corpusFacts(integrity)}</span>
      <span className="ribbon-item ribbon-seal">{sealFacts(integrity?.hash8)}</span>
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
