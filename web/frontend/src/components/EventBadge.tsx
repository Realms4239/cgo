import { useUIStore } from '../store/ui'

// Badge d'event — la même vérité partout (popup, bannière live, rail épinglé) :
// cellule testée, progression, shaping appliqué. Un seul composant, un seul store.
export function EventBadge({ compact }: { compact?: boolean }) {
  const live = useUIStore((s: any) => s.live)
  if (!live?.running && !live?.event_id) return null
  const ev = live.event_id || 0
  const total = live.total_events || 0
  const cell = [live.profile, live.qdisc, live.cc].filter(Boolean).join('/')
  const rep = live.repetition ? ` rep ${live.repetition}` : ''
  const cap = live.profile_capacity_mbps ? ` · ${live.profile_capacity_mbps}Mbit` : ''
  const delay = live.profile_delay_ms ? `/${live.profile_delay_ms}ms` : ''
  const loss = live.profile_loss_pct ? `/${live.profile_loss_pct}%` : ''
  return (
    <span
      data-testid="event-badge"
      className="mono"
      title={`cellule en cours — façonnage appliqué au bord : ${live.profile ?? ''} ${live.profile_capacity_mbps ?? ''} Mbit/s, délai ${live.profile_delay_ms ?? ''} ms, gigue ${live.profile_jitter_ms ?? ''} ms, perte ${live.profile_loss_pct ?? ''} %`}
      style={{
        fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.06em',
        color: 'var(--t-live, #5ad3e3)', border: '1px solid rgba(90,211,227,0.35)',
        padding: '3px 8px', whiteSpace: 'nowrap', textTransform: 'uppercase',
      }}
    >
      {ev}/{total} {cell}{rep}{!compact && (cap || delay || loss) ? `${cap}${delay}${loss}` : ''}
    </span>
  )
}

export default EventBadge
