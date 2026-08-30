import type { LiveFrame } from './types'
import { pushFrame } from './live'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null
let last: Partial<LiveFrame> = {}
let frameCount = 0
export const __CGO_SSE = { lastID: 0, dropped: 0, ringLen: 0, frameCount: 0 }
if (typeof window !== 'undefined') (window as any).__CGO_SSE = __CGO_SSE

const gatesEqual = (a: unknown, b: unknown): boolean => {
  if (a === b) return true
  if (!Array.isArray(a) || !Array.isArray(b)) return false
  if (a.length !== (b as unknown[]).length) return false
  return (a as unknown[]).every((v, i) => v === (b as unknown[])[i])
}

export function connectSSE() {
  if (es) return
  // le serveur limite le replay aux 5 dernières frames sans Last-Event-ID;
  // le navigateur renvoie Last-Event-ID à la reconnexion (retry:2000)
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
      // frontière de vérité: les frames idle ne mesurent rien — jamais de 0 fabriqué.
      // En course, les frames tout-zéro sont des trous de pompe, pas des mesures — ignorer.
      // Les frames surveillance (surveil) et burst sont des mesures avec
      // running=false — elles doivent atteindre les anneaux, sinon le mur meurt.
      const measured = !!(data.rtt_p50_ms || data.rtt_p95_ms || data.small_p95_ms || data.bulk_goodput_mbps || data.drops)
      if (data.ts && (data.running || data.phase === 'surveil' || data.phase === 'burst') && measured) pushFrame(data.ts, data as any)
      frameCount++
      __CGO_SSE.frameCount = frameCount
      if (typeof window !== 'undefined') (window as any).__CGO_SSE = __CGO_SSE
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
export const _frameCount = () => frameCount
