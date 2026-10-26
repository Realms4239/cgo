// AuditLive — l'étape « Observer » : phases, restant, p50/p95 de la fenêtre
// courante et ANNULER. Aucun zéro n'est affiché comme une mesure : une sonde
// perdue (0) est ignorée par le calcul des percentiles.
import { ArmButton } from './ArmButton'
import { Badge, StatusPip } from './ui/StatusPip'
import { Stat } from './ui/Stat'
import { auditPhases, liveStats, phaseLabel, remainingSeconds } from '../lib/auditFlow'
import type { AuditDirection, AuditLive as AuditLiveState } from '../lib/types'

function fmtMs(v: number | null): string {
  return v == null ? '—' : v.toFixed(1)
}

export function AuditLive({ live, direction, onStop }: {
  live: AuditLiveState | null
  direction: AuditDirection
  onStop: () => void
}) {
  const phases = auditPhases(direction)
  const stats = liveStats(live?.samples?.rtt_ms, live?.samples?.small_ms)
  const remain = remainingSeconds(live)
  const running = !!live?.running
  return (
    <div data-testid="audit-live" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
        <Badge tone={running ? 'live' : 'steel'}>{running ? 'en cours' : 'attente'}</Badge>
        <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-muted)' }}>
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

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 'var(--sp-2)' }}>
        <div className="c-kpi"><Stat label="RTT p50" value={fmtMs(stats.rttP50)} unit="ms" /></div>
        <div className="c-kpi"><Stat label="RTT p95" value={fmtMs(stats.rttP95)} unit="ms" /></div>
        <div className="c-kpi"><Stat label="petit p50" value={fmtMs(stats.smallP50)} unit="ms" /></div>
        <div className="c-kpi"><Stat label="petit p95" value={fmtMs(stats.smallP95)} unit="ms" /></div>
      </div>

      <div className="mono" style={{ fontSize: 'var(--fs-micro)', lineHeight: 1.5, color: 'var(--text-faint)' }}>
        {live?.phase_remain_sec != null ? `cette phase : ${live.phase_remain_sec} s restantes · ` : ''}
        échantillons de la fenêtre courante (≤ 32) — une sonde perdue n'entre pas dans les percentiles
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
