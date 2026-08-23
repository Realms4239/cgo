import type { LiveFrame } from './types'
import { pushFrame } from './live'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null
let last: Partial<LiveFrame> = {}
let frameCount = 0

export function connectSSE() {
  if (es) return
  es = new EventSource('/api/stream')
  const store = useUIStore

  es.onmessage = (ev) => {
    try {
      const data = JSON.parse(ev.data) as Partial<LiveFrame>
      // delta: retain last structural values (SPEC §2.4)
      let structural = false
      for (const k of ['profile','qdisc','cc','phase','load_status','running'] as const) {
        if (data[k] === undefined && last[k] !== undefined) (data as any)[k] = last[k]
        if (data[k] !== undefined && data[k] !== last[k]) structural = true
        if (data[k] !== undefined) (last as any)[k] = data[k]
      }
      if (data.ts) pushFrame(data.ts, data as any)
      // ponytail: React sees ~2 Hz snapshots (charts read rings at full 10 Hz);
      // structural changes flush immediately.
      frameCount++
      if (structural || frameCount % 5 === 0) {
        store.getState().setLive(data as LiveFrame)
      }
      store.getState().setConnected(true)
    } catch {}
  }
  es.addEventListener('backpressure', () => {
    store.getState().setSseStatus('backpressure')
  })
  es.onerror = () => {
    store.getState().setConnected(false)
  }
}

export function disconnectSSE() {
  es?.close(); es = null
}
