// Observer — le vrai instrument : une seule règle de temps pour le délai, le
// débit utile et l'interface, plus le journal des faits. Les anneaux sont la
// source ; aucune valeur n'est recopiée dans React à chaque trame.
import { useState } from 'react'
import { instrument, type DomainEvent, type DomainEventRing } from '../../lib/ringBuffer'
import { useInterval } from '../../lib/hooks'
import TimeField from '../cockpit/TimeField'
import JournalRail from '../cockpit/JournalRail'
import { CardHead } from '../ui/CardHead'

// Heure réelle du fait (UTC), jamais un compteur ni une gravité fabriquée.
// Une date invalide donne un rendu honnête : ni heure ni attribut dateTime,
// pas de RangeError ni de « Invalid Date » affichée.
function validDate(ts: number): Date | null {
  const d = new Date(ts)
  return Number.isFinite(d.getTime()) ? d : null
}

function clock(d: Date | null): string {
  return d ? d.toISOString().slice(11, 19) : ''
}

/** Bande de repères — une marque par événement réel issu de `events.list()`. */
export function EventTimeline({ events }: { events: DomainEventRing }) {
  const [entries, setEntries] = useState<DomainEvent[]>(() => events.list())
  // Se rafraîchit sans toucher au drapeau `dirty` déjà consommé par JournalRail.
  useInterval(() => {
    const next = events.list()
    const changed = next.length !== entries.length
      || next[next.length - 1]?.ts !== entries[entries.length - 1]?.ts
    if (changed) setEntries(next)
  }, 500)

  return (
    <div className="event-timeline" data-state={entries.length ? 'events' : 'no-data'}>
      <p className="et-label mono">Repères</p>
      {entries.length === 0 ? (
        <p className="et-empty" role="status">
          <span className="et-mark et-mark-empty" aria-hidden="true" />
          aucun repère
        </p>
      ) : (
        <ol className="et-list">
          {entries.map(e => {
            const d = validDate(e.ts)
            return (
              <li key={`${e.ts}-${e.kind}-${e.label}`} className="et-event">
                <span className="et-mark" aria-hidden="true" />
                <time className="et-ts mono" dateTime={d?.toISOString()}>{clock(d)}</time>
                <span className="et-kind">{e.label || e.kind}</span>
              </li>
            )
          })}
        </ol>
      )}
    </div>
  )
}

export default function Observer({ running = false }: { running?: boolean }) {
  return (
    <div className="observer">
      <CardHead label="Instrument" sub="Latence, débit et état du lien au fil du temps." />
      <EventTimeline events={instrument.events} />
      <div className="observer-field">
        <TimeField rings={instrument} objectiveMs={220} />
        <JournalRail events={instrument.events} running={running} />
      </div>
    </div>
  )
}
