// HUDtemps réel : les anneaux singletons sont la source, le store React ne l'est plus.
// Une absence reste un tiret (jamais un 0 fabriqué, SPEC §2.2). La peinture écrit
// directement dans des nœuds textContent : aucun setState dans la boucle rAF.
import { useEffect, useRef } from 'react'
import { eventRing, metricRing } from '../lib/liveRings'

const ABSENT = '—'
const CONGESTION = 'CONGESTION > 220 ms'
const PROTEGE = 'PROTÉGÉ'

export interface HUDSpans {
  throughput: HTMLSpanElement | null
  rtt: HTMLSpanElement | null
  drops: HTMLSpanElement | null
  status: HTMLSpanElement | null
  event: HTMLSpanElement | null
}

function fmt(v: number | null | undefined, digits: number): string {
  if (v == null || Number.isNaN(v)) return ABSENT
  return v.toFixed(digits)
}

/** Peint l'état courant des anneaux dans les nœuds du HUD. */
export function paintHUD(spans: HUDSpans): void {
  const put = (el: HTMLSpanElement | null, value: string) => {
    if (el) el.textContent = value
  }
  const m = metricRing.at(-1)
  put(spans.throughput, fmt(m?.throughputMbps, 1))
  put(spans.rtt, fmt(m?.rttMs, 0))
  put(spans.drops, fmt(m?.dropsBytes, 0))
  const rtt = m?.rttMs
  put(spans.status, rtt != null && !Number.isNaN(rtt) && rtt > 220 ? CONGESTION : PROTEGE)
  put(spans.event, eventRing.at(-1)?.title ?? ABSENT)
  metricRing.markPainted()
}

export function RingBufferHUD() {
  const throughput = useRef<HTMLSpanElement>(null)
  const rtt = useRef<HTMLSpanElement>(null)
  const drops = useRef<HTMLSpanElement>(null)
  const status = useRef<HTMLSpanElement>(null)
  const event = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    const spans: HUDSpans = {
      throughput: throughput.current,
      rtt: rtt.current,
      drops: drops.current,
      status: status.current,
      event: event.current,
    }
    let frame = 0
    const tick = () => {
      paintHUD(spans)
      frame = requestAnimationFrame(tick)
    }
    paintHUD(spans)
    frame = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(frame)
  }, [])

  return (
    <section className="hud" aria-label="État du lien en temps réel">
      <div className="hud-metric">
        <span className="hud-label">Débit</span>
        <span ref={throughput} data-hud="throughput">{ABSENT}</span>
        <small>Mbit/s</small>
      </div>
      <div className="hud-metric">
        <span className="hud-label">RTT WIS 2.0</span>
        <span ref={rtt} data-hud="rtt">{ABSENT}</span>
        <small>ms</small>
      </div>
      <div className="hud-metric">
        <span className="hud-label">Drops eBPF</span>
        <span ref={drops} data-hud="drops">{ABSENT}</span>
      </div>
      <span ref={status} data-hud="status" className="hud-status">{PROTEGE}</span>
      <div className="hud-event">
        <span className="hud-label">Dernier événement</span>
        <span ref={event} data-hud="event">{ABSENT}</span>
      </div>
    </section>
  )
}
