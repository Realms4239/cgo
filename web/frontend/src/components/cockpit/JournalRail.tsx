// Journal d'événements — témoin de la campagne en cours. Le mur précédent
// gelait le journal pendant la mesure ; ici il se rafraîchit dès que l'anneau
// d'événements est dirty, au plus à 4 Hz, sans rendu React par trame SSE.
import { useRef, useState } from 'react'
import { useRafLoop } from '../../lib/hooks'
import type { DomainEvent, DomainEventRing } from '../../lib/ringBuffer'
import { EmptyState } from '../ui/EmptyState'
import { StatusPip } from '../ui/StatusPip'

export function journalEntries(events: DomainEventRing, limit = 12): DomainEvent[] {
  return events.list().reverse().slice(0, limit)
}

function clock(ts: number): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}`
}

export interface JournalRailProps {
  events: DomainEventRing
  running?: boolean
}

export default function JournalRail({ events, running = false }: JournalRailProps) {
  const [tick, setTick] = useState(0)
  const last = useRef(0)
  useRafLoop((ts) => {
    if (!events.dirty || ts - last.current < 250) return
    last.current = ts
    events.dirty = false
    setTick(t => t + 1)
  }, true)

  const entries = journalEntries(events)
  return (
    <aside className="journal-rail" aria-label="Journal des événements" data-tick={tick}>
      <div className="jr-head mono">
        <span>JOURNAL</span>
        <span className="jr-live">
          <StatusPip state={running ? 'on' : 'idle'} label={running ? 'en direct' : 'aucun fait'} />
          {events.size > 0 ? ` · ${events.size}` : ''}
        </span>
      </div>
      {entries.length === 0 ? (
        <EmptyState kind="empty" className="jr-empty" hint="Aucun événement — le journal se remplira aux faits réels de la campagne." />
      ) : (
        <ol className="jr-list">
          {entries.map((e, i) => (
            <li key={`${e.ts}-${i}`} className="jr-row mono">
              <span className="jr-ts">{clock(e.ts)}</span>
              <span className="jr-kind">{e.kind}</span>
              <span className="jr-label" title={e.detail}>{e.label}</span>
            </li>
          ))}
        </ol>
      )}
    </aside>
  )
}
