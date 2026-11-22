import { pushFrame, clearLive } from './live'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null

// results.Field rend "" pour une colonne absente d'une ligne courte (ReadAQM
// tolère ces lignes) : `parseFloat(x) || 0` en faisait une mesure de 0 ms.
// NaN porte l'absence jusqu'aux anneaux, où pushFrame la garde comme absence.
const archived = (v: unknown): number =>
  v == null || (typeof v === 'string' && v.trim() === '') ? NaN : Number(v)
// Dans le corpus gelé, small_p95_ms = 0 est une sonde vide (api/server.go:385).
const small = (v: unknown): number => {
  const n = archived(v)
  return n > 0 ? n : NaN
}
// Compteurs et identifiants : comme archived() mais pour un entier — absent,
// blanc ou malformé reste NaN (jamais un 0 mesuré) ; un vrai 0 reste 0.
const archivedInt = (v: unknown): number => {
  const n = archived(v)
  return Number.isInteger(n) ? n : NaN
}

export function startReplay(runId: string) {
  stopReplay()
  clearLive()
  useUIStore.getState().setReplay(true, runId)
  es = new EventSource(`/api/replay/stream?run=${encodeURIComponent(runId)}`)
  es.onmessage = (ev) => {
    try {
      const d = JSON.parse(ev.data)
      const ts = d.ts || Date.now()
      const rtt50 = archived(d.rtt_p50_ms)
      const rtt95 = archived(d.rtt_p95_ms)
      const smallP95 = small(d.small_p95_ms)
      const goodput = archived(d.bulk_goodput_mbps)
      // pousse les métriques archivées réelles dans les anneaux live
      pushFrame(ts, {
        rtt_p50_ms: rtt50, rtt_p95_ms: rtt95,
        small_p95_ms: smallP95, bulk_goodput_mbps: goodput,
      } as any)
      // also update store so banner/gates reflect replay
      useUIStore.getState().setLive({
        ts, phase: 'replay', profile: d.profile, qdisc: d.qdisc, cc: d.cc,
        repetition: archivedInt(d.repetition), event_id: archivedInt(d.event_id),
        rtt_p50_ms: rtt50, rtt_p95_ms: rtt95,
        small_p95_ms: smallP95, bulk_goodput_mbps: goodput,
        drops: archivedInt(d.drops), gates: null, running: true, load_status: 'replay',
      } as any)
    } catch {}
  }
  es.onerror = () => stopReplay()
  es.addEventListener('error', stopReplay)
}

export function stopReplay() {
  if (es) { es.close(); es = null }
  useUIStore.getState().setReplay(false, null)
}
