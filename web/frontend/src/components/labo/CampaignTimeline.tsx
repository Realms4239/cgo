// Chronologie de campagne — une seule horloge réelle, partagée avec la règle
// du champ : les bornes viennent des horodatages réels de phase (phaseSince),
// jamais d'un décalage « maintenant − 30 s » qui grandirait tout seul. La fin
// d'une phase est le début réel de la suivante ; la phase ouverte va à
// l'heure courante uniquement pendant la mesure.
import { useState } from 'react'
import { useInterval } from '../../lib/hooks'
import { CardHead } from '../ui/CardHead'
import { EmptyState } from '../ui/EmptyState'

const PHASES = [
  { key: 'baseline', label: 'Repos' },
  { key: 'charge', label: 'Charge' },
  { key: 'recup', label: 'Récupération' },
] as const

export interface PhaseBand {
  key: string
  label: string
  start: number
  end: number
}

/** Bornes réelles seulement. Aucune phase sans horodatage n'est dessinée. */
export function campaignBands(phaseSince: Record<string, number>, now: number, running: boolean): PhaseBand[] {
  const seen = PHASES.filter(p => Number.isFinite(phaseSince[p.key]) && phaseSince[p.key] > 0)
  if (!seen.length) return []
  return seen.map((p, i) => {
    const start = phaseSince[p.key]
    const next = seen[i + 1]
    const end = next ? phaseSince[next.key] : running ? Math.max(now, start) : start
    return { key: p.key, label: p.label, start, end }
  })
}

const secs = (start: number, end: number) => Math.max(0, Math.round((end - start) / 1000))

export default function CampaignTimeline({
  phaseSince, running,
}: {
  phaseSince: Record<string, number>
  running: boolean
}) {
  const [now, setNow] = useState(() => Date.now())
  useInterval(() => setNow(Date.now()), 1000, running)
  const bands = campaignBands(phaseSince, now, running)
  const min = bands.length ? bands[0].start : 0
  const max = bands.length ? Math.max(...bands.map(b => b.end)) : 0
  const span = Math.max(1, max - min)
  const at = (t: number) => Math.max(0, Math.min(100, ((t - min) / span) * 100))
  const state = bands.length ? (running ? 'live' : 'final') : 'absent'

  return (
    <section className="labo-timeline" data-state={state}>
      <CardHead label="Chronologie de campagne" sub="Une seule horloge réelle, partagée avec la règle du champ." />
      {bands.length ? (
        <>
          <div className="labo-tl-track" aria-hidden="true">
            {bands.map(b => (
              <span key={b.key} className="labo-tl-band" data-phase={b.key}
                style={{ left: `${at(b.start)}%`, width: `${Math.max(0, at(b.end) - at(b.start))}%` }} />
            ))}
            {running ? <span className="labo-tl-cursor" style={{ left: `${at(now)}%` }} /> : null}
          </div>
          <ul className="labo-tl-legend">
            {bands.map(b => (
              <li key={b.key}>
                <span className="labo-tl-swatch" data-phase={b.key} aria-hidden="true" />
                {b.label} · {b.end > b.start ? `${secs(b.start, b.end)} s` : '—'}
              </li>
            ))}
          </ul>
        </>
      ) : (
        <EmptyState kind="idle" hint="aucune phase enregistrée — la chronologie se remplira aux temps réels de la campagne." />
      )}
    </section>
  )
}
