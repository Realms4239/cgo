import { pushFrame, clearLive } from './live'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null

export function startReplay(runId: string) {
  stopReplay()
  clearLive()
  useUIStore.getState().setReplay(true, runId)
  es = new EventSource(`/api/replay/stream?run=${encodeURIComponent(runId)}`)
  es.onmessage = (ev) => {
    try {
      const d = JSON.parse(ev.data)
      const ts = d.ts || Date.now()
      // push real archived metrics into live rings
      pushFrame(ts, {
        rtt_p50_ms: parseFloat(d.rtt_p50_ms) || 0,
        rtt_p95_ms: parseFloat(d.rtt_p95_ms) || 0,
        small_p95_ms: parseFloat(d.small_p95_ms) || 0,
        bulk_goodput_mbps: parseFloat(d.bulk_goodput_mbps) || 0,
      } as any)
      // also update store so banner/gates reflect replay
      useUIStore.getState().setLive({
        ts, phase: 'replay', profile: d.profile, qdisc: d.qdisc, cc: d.cc,
        repetition: parseInt(d.repetition)||0, event_id: parseInt(d.event_id)||0,
        rtt_p50_ms: parseFloat(d.rtt_p50_ms)||0, rtt_p95_ms: parseFloat(d.rtt_p95_ms)||0,
        small_p95_ms: parseFloat(d.small_p95_ms)||0, bulk_goodput_mbps: parseFloat(d.bulk_goodput_mbps)||0,
        drops: parseInt(d.drops)||0, gates: null, running: true, load_status: 'replay',
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
