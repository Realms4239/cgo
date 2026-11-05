// AuditLive — l'étape « Observer » : phases, restant, la trace des sondes et
// deux faits face à l'objectif. Une sonde perdue brise le tracé ; elle n'est
// jamais affichée comme 0 ms.
import { ArmButton } from './ArmButton'
import { Badge, StatusPip } from './ui/StatusPip'
import { TraceCanvas } from './TraceCanvas'
import { auditPhases, phaseLabel, remainingSeconds } from '../lib/auditFlow'
import type { AuditDirection, AuditLive as AuditLiveState } from '../lib/types'

function fmtMs(v: number | null | undefined): string {
  return v != null && Number.isFinite(v) && v > 0 ? v.toFixed(1) : '—'
}

function finiteDeadline(v: number | null | undefined): number | null {
  return v != null && Number.isFinite(v) && v > 0 ? v : null
}

export function AuditLive({ live, direction, onStop }: {
  live: AuditLiveState | null
  direction: AuditDirection
  onStop: () => void
}) {
  const phases = auditPhases(direction)
  const remain = remainingSeconds(live)
  const running = !!live?.running
  const deadline = finiteDeadline(live?.deadline_ms)
  return (
    <div data-testid="audit-live" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
        <Badge tone={running ? 'live' : 'steel'}>{running ? 'en cours' : 'attente'}</Badge>
        <span className="mono" role="status" aria-live="polite" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-muted)' }}>
          {phaseLabel(live?.phase)} · reste {remain ?? '—'} s
        </span>
        <span className="mono" style={{ marginLeft: 'auto', fontSize: 'var(--fs-micro)', color: 'var(--text-faint)', fontVariantNumeric: 'tabular-nums' }}>
          {live?.elapsed_sec ?? '—'}/{live?.total_sec ?? '—'} s
        </span>
      </div>

      {/* Chaque phase est un état ; seule la phase active est allumée — pas de
          « terminé » inventé quand le « both » repasse par un mini-drain. */}
      <div role="list" aria-label="phases de l'audit" style={{ display: 'flex', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
        {phases.map(p => (
          <span key={p} role="listitem" aria-current={live?.phase === p ? 'step' : undefined}>
            <StatusPip state={live?.phase === p ? 'live' : 'idle'} label={phaseLabel(p)} />
          </span>
        ))}
      </div>

      {deadline != null ? (
        <TraceCanvas
          points={live?.timeline ?? []}
          totalSec={live?.total_sec}
          deadlineMs={deadline}
          responseMs={live?.last_rtt_ms}
          loadResponseMs={live?.last_small_ms}
        />
      ) : (
        <p className="mono muted">Trace indisponible — objectif non scellé par la réponse.</p>
      )}

      <div className="live-brief">
        <div className="live-fact">
          <span>temps de réponse</span>
          <b>{fmtMs(live?.last_rtt_ms)}<small> ms</small></b>
        </div>
        <div className="live-fact">
          <span>réponse sous charge</span>
          <b>{fmtMs(live?.last_small_ms)}<small> ms</small></b>
          <em>{deadline != null ? `objectif ${Math.round(deadline)} ms` : 'objectif non disponible'}</em>
        </div>
      </div>

      <div className="mono" style={{ fontSize: 'var(--fs-micro)', lineHeight: 1.5, color: 'var(--text-faint)' }}>
        {live?.phase_remain_sec != null ? `cette phase : ${live.phase_remain_sec} s restantes · ` : ''}
        une sonde perdue brise la ligne — elle ne vaut jamais 0 ms
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
        <ArmButton label="ANNULER" confirmLabel="CONFIRMER L'ANNULATION" onConfirm={onStop} />
        <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>
          aucun CSV partiel n'est conservé
        </span>
      </div>
    </div>
  )
}
