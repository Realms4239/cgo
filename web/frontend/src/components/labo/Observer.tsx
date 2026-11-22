// Observer — le vrai instrument : une seule règle de temps pour le délai, le
// débit utile et l'interface, plus le journal des faits. Les anneaux sont la
// source ; aucune valeur n'est recopiée dans React à chaque trame.
import { useState } from 'react'
import { instrument, type DomainEvent, type DomainEventRing } from '../../lib/ringBuffer'
import { live as liveRing } from '../../lib/live'
import { useCapabilities, useInterval } from '../../lib/hooks'
import { useBenchStore } from '../../store/bench'
import { useUIStore } from '../../store/ui'
import TimeField from '../cockpit/TimeField'
import JournalRail from '../cockpit/JournalRail'
import { CardHead } from '../ui/CardHead'
import { EmptyState } from '../ui/EmptyState'
import Matrix, { type MatrixPlan, type CampaignProgress } from './Matrix'
import Gates from './Gates'
import CampaignTimeline from './CampaignTimeline'

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
        <EmptyState kind="idle" className="et-empty" markClassName="et-mark-empty"
          hint="aucun repère — les faits enregistrés pendant la mesure s'afficheront ici avec leur heure." />
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

/** Un objet live présent mais vide ne doit pas masquer les bornes réelles de l'anneau. */
export function selectPhaseSince(
  livePhaseSince: Record<string, number> | undefined,
  ringPhaseSince: Record<string, number>,
): Record<string, number> {
  return livePhaseSince && Object.keys(livePhaseSince).length > 0 ? livePhaseSince : ringPhaseSince
}

export default function Observer({ running = false }: { running?: boolean }) {
  const caps = useCapabilities()
  const profiles = useBenchStore(s => s.profiles)
  const reps = useBenchStore(s => s.reps)
  const live = useUIStore(s => s.live)
  // Le plan vient des réglages réels et des capacités annoncées par l'hôte.
  const plan: MatrixPlan = { profiles, qdiscs: caps?.qdiscs ?? [], ccs: caps?.ccs ?? [], reps }
  // Une campagne a eu lieu dès qu'un identifiant d'événement réel existe —
  // l'après-campagne ne doit jamais retomber en « attente ».
  const progress: CampaignProgress | null = live && (live.event_id > 0 || (live.total_events ?? 0) > 0 || (live.done_events ?? 0) > 0)
    ? { running: !!live.running, eventId: live.event_id, doneEvents: live.done_events ?? 0, totalEvents: live.total_events ?? 0 }
    : null
  const isRunning = running || !!live?.running
  return (
    <div className="observer">
      <CardHead label="Instrument" sub="Latence, débit et état du lien au fil du temps." />
      <EventTimeline events={instrument.events} />
      <div className="observer-depth">
        <Matrix plan={plan} progress={progress} />
        <Gates gates={live?.gates ?? []} />
      </div>
      <CampaignTimeline phaseSince={selectPhaseSince(live?.phaseSince, liveRing.phaseSince)} running={isRunning} />
      <div className="observer-field">
        <TimeField rings={instrument} objectiveMs={220} />
        <JournalRail events={instrument.events} running={isRunning} />
      </div>
    </div>
  )
}
