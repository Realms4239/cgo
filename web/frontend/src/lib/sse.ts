import type { LiveFrame } from './types'
import { pushIface, live as liveRing, resetEpoch } from './live'
import { instrument, type InstrumentRings } from './ringBuffer'
import { useUIStore } from '../store/ui'

let es: EventSource | null = null
let last: Partial<LiveFrame> = {}
let frameCount = 0
let lastStoreSync = 0
let lastLoggedPhase = ''
export const __CGO_SSE = { lastID: 0, dropped: 0, ringLen: 0, frameCount: 0 }
if (typeof window !== 'undefined') (window as any).__CGO_SSE = __CGO_SSE

const gatesEqual = (a: unknown, b: unknown): boolean => {
  if (a === b) return true
  if (!Array.isArray(a) || !Array.isArray(b)) return false
  if (a.length !== (b as unknown[]).length) return false
  return (a as unknown[]).every((v, i) => v === (b as unknown[])[i])
}

const PHASE_FR: Record<string, string> = {
  baseline: 'Repos', charge: 'Charge', recup: 'Récupération',
  surveil: 'Surveillance', burst: 'Rafale',
}

/** EventSource natif ne peut pas poser d'en-tête : la génération voyage en query. */
function streamURL(): string {
  return liveRing.epoch ? `/api/stream?epoch=${encodeURIComponent(liveRing.epoch)}` : '/api/stream'
}

/** Rouvrir le flux pour que les reconnexions suivantes portent la nouvelle génération. */
function reopenStream() {
  const old = es
  es = null
  old?.close()
  connectSSE()
}

export function connectSSE() {
  if (es) return
  // le serveur limite le replay aux 5 dernières frames sans Last-Event-ID ;
  // le navigateur renvoie Last-Event-ID à la reconnexion (retry:2000)
  es = new EventSource(streamURL())
  const store = useUIStore

  es.onmessage = (ev) => {
    try {
      if (store.getState().replayRunning) return // pause live during replay
      const data = JSON.parse(ev.data) as Partial<LiveFrame>
      let structural = false
      const frameEpoch = data.epoch
      if (typeof frameEpoch === 'string' && frameEpoch !== liveRing.epoch) {
        // Redémarrage du hub : le snapshot complet a été émis AVANT les deltas.
        // Vider l'état, appliquer le snapshot, rouvrir avec la nouvelle génération.
        resetEpoch(frameEpoch)
        last = {}
        lastLoggedPhase = ''
        lastStoreSync = 0
        for (const r of [instrument.small, instrument.rtt50, instrument.rtt95,
                         instrument.goodput, instrument.ifaceRx, instrument.ifaceTx,
                         instrument.events]) r.clear()
        structural = true
        reopenStream()
      }
      // l'interface se mesure hors campagne : pousser AVANT la porte "measured"
      if (data.ts) pushIface(data)
      // delta: retain last structural + sparse values (SPEC §2.4)
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
      // Écriture directe dans les anneaux typés : aucun rendu React par trame.
      const pushed = ingestFrame(data)
      __CGO_SSE.ringLen = instrument.small.size
      if (pushed && data.phase && data.phase !== lastLoggedPhase) {
        lastLoggedPhase = data.phase
        instrument.events.push({
          ts: data.ts!, kind: 'phase', label: PHASE_FR[data.phase] ?? data.phase,
        })
      }
      frameCount++
      __CGO_SSE.frameCount = frameCount
      if (typeof window !== 'undefined') (window as any).__CGO_SSE = __CGO_SSE
      // pont phaseSince — l'anneau porte les bornes de phase (lib/live.ts:13),
      // le store ne les voyait jamais : le compte à rebours restait à 0/120 s.
      // Sync plafonnée à 4 Hz (immédiate sur changement structurel).
      const nowMs = typeof performance !== 'undefined' ? performance.now() : Date.now()
      if (shouldSyncStore(nowMs, lastStoreSync, structural)) {
        lastStoreSync = nowMs
        store.getState().setLive({ ...(data as LiveFrame), phaseSince: { ...liveRing.phaseSince } })
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

/** Sync du store au plus à 4 Hz, immédiate sur un fait structurel. */
export function shouldSyncStore(nowMs: number, lastMs: number, structural: boolean): boolean {
  return structural || nowMs - lastMs >= 250
}

/**
 * Frontière de vérité : n'écrit que des trames qui mesurent réellement quelque
 * chose. Un champ absent devient NaN (jamais 0) dans son anneau. Fonction pure
 * à part des anneaux injectés — le test l'appelle sans EventSource ni store.
 */
export function ingestFrame(data: Partial<LiveFrame>, r: InstrumentRings = instrument): boolean {
  const ts = data.ts
  if (!ts) return false
  const has = (v: unknown) => v !== undefined && v !== null
  // frontière de vérité: les frames idle ne mesurent rien — jamais de 0 fabriqué.
  // En course, les frames tout-zéro sont des trous de pompe, pas des mesures — ignorer.
  // Les frames surveillance (surveil) et burst sont des mesures avec
  // running=false — elles doivent atteindre les anneaux, sinon le mur meurt.
  const measured = has(data.rtt_p50_ms) || has(data.rtt_p95_ms) || has(data.small_p95_ms)
    || has(data.bulk_goodput_mbps) || has(data.drops) || has(data.deadline_ok_pct)
    || has(data.wasted_bytes) || has(data.cost_ar_per_h)
  if (!measured) return false
  if (!(data.running || data.phase === 'surveil' || data.phase === 'burst')) return false
  const values = [data.rtt_p50_ms, data.rtt_p95_ms, data.small_p95_ms,
                  data.bulk_goodput_mbps, data.drops, data.deadline_ok_pct,
                  data.wasted_bytes, data.cost_ar_per_h].filter(has)
  if (values.length > 0 && values.every(v => v === 0)) return false

  r.small.push(ts, data.small_p95_ms ?? null)
  r.rtt50.push(ts, data.rtt_p50_ms ?? null)
  r.rtt95.push(ts, data.rtt_p95_ms ?? null)
  // les artefacts de compteur (reset de qdisc) n'entrent jamais dans l'anneau goodput
  const g = data.bulk_goodput_mbps
  r.goodput.push(ts, g != null && g >= 0 && g <= 2500 ? g : null)
  if (data.iface_sample_ts) {
    r.ifaceRx.push(data.iface_sample_ts, data.iface_valid ? (data.iface_rx_mbps ?? null) : null)
    r.ifaceTx.push(data.iface_sample_ts, data.iface_valid ? (data.iface_tx_mbps ?? null) : null)
  }
  return true
}
