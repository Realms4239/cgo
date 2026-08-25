import type { LiveFrame } from './types'
import { pushFrame } from './live'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null
let last: Partial<LiveFrame> = {}
let frameCount = 0

const gatesEqual = (a: unknown, b: unknown): boolean => {
  if (a === b) return true
  if (!Array.isArray(a) || !Array.isArray(b)) return false
  if (a.length !== (b as unknown[]).length) return false
  return (a as unknown[]).every((v, i) => v === (b as unknown[])[i])
}

export function connectSSE() {
  if (es) return
  es = new EventSource('/api/stream')
  const store = useUIStore

  es.onmessage = (ev) => {
    try {
      if (store.getState().replayRunning) return // pause live during replay
      const data = JSON.parse(ev.data) as Partial<LiveFrame>
      // delta: retain last structural + sparse values (SPEC §2.4)
      let structural = false
      for (const k of ['profile','qdisc','cc','phase','load_status','running',
                       'event_id','repetition','drops','gates'] as const) {
        if (data[k] === undefined && last[k] !== undefined) (data as any)[k] = last[k]
        if (data[k] !== undefined) {
          if (k === 'gates') {
            if (!gatesEqual(data[k], last[k])) structural = true
          } else if (data[k] !== last[k]) structural = true
        }
        if (data[k] !== undefined) (last as any)[k] = data[k]
      }
      // truth boundary: idle frames carry no measurement — never fabricate 0s
      if (data.ts && data.running) pushFrame(data.ts, data as any)
      frameCount++
      if (structural || frameCount % 5 === 0) {
        store.getState().setLive(data as LiveFrame)
      }
      store.getState().setConnected(true)
      store.getState().setSseStatus('connecté')
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
