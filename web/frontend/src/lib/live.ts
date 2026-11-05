// live — mutable 10 Hz ring buffers (SPEC §2.4). Not in Zustand.
import type { LiveFrame } from './types'

export type Pt = [number, number]
export type Ring = Pt[]
export type IfacePt = [number, number | null]

export const live = {
  rtt50: [] as Ring,
  rtt95: [] as Ring,
  small: [] as Ring,
  goodput: [] as Ring,
  // sparse SSE series — own buffer per metric (stat sparklines never borrow)
  deadline: [] as Ring,
  wasted: [] as Ring,
  cost: [] as Ring,
  drops: [] as Ring,
  // anneau interface séparé — 60 s à 10 Hz. clearLive ne le touche pas :
  // l'observation continue survive aux transitions de campagne.
  iface: {
    rx: [] as IfacePt[],
    tx: [] as IfacePt[],
    name: '',
    source: '',
    valid: false,
    gap: '',
    windowMs: 0,
    rxDrops: 0,
    txDrops: 0,
    ts: 0,
    seq: 0,
    max: 600,
  },
  max: 1800, // 180 s, un événement à 10 Hz
  ts: 0,
  phase: '', // baseline|charge|recup — from SSE delta, drives CHARGE markArea (not estimated)
  phaseSince: {} as Record<string, number>, // first ts seen per phase — drives Timeline 48 phase bands
  seq: 0, // dirty-check — incrémenté à chaque frame mesurée, les rendus ne suivent que le changement
}

function push(r: Ring, ts: number, v: number | null) {
  if (v == null || !Number.isFinite(v)) r.push([ts, 0])
  else r.push([ts, v])
  if (r.length > live.max) r.shift()
}

function pushSparse(r: Ring, ts: number, v: number | null | undefined) {
  if (v == null || !Number.isFinite(v)) return
  r.push([ts, v])
  if (r.length > live.max) r.shift()
}

export function pushIface(f: Partial<LiveFrame>) {
  const ts = f.iface_sample_ts
  if (!ts) return
  const r = live.iface
  if ((f.iface_name ?? '') !== r.name) {
    r.rx.length = 0
    r.tx.length = 0
    r.ts = 0
    r.name = f.iface_name ?? ''
  }
  if (r.ts) {
    if (ts === r.ts) return
    if (ts < r.ts) {
      if (r.ts - ts <= 1000) return
      // saut d'horloge arrière > 1 s : la série repart, sans pont
      r.rx.length = 0
      r.tx.length = 0
    }
  }
  const valid = !!f.iface_valid
  const mbps = (v?: number) => (valid && v != null && Number.isFinite(v) ? v : null)
  r.rx.push([ts, mbps(f.iface_rx_mbps)])
  r.tx.push([ts, mbps(f.iface_tx_mbps)])
  if (r.rx.length > r.max) {
    r.rx.shift()
    r.tx.shift()
  }
  r.ts = ts
  r.valid = valid
  r.gap = f.iface_gap ?? ''
  r.windowMs = f.iface_window_ms ?? 0
  r.source = f.iface_source ?? r.source
  if (valid) {
    if (typeof f.iface_rx_drops === 'number') r.rxDrops = f.iface_rx_drops
    if (typeof f.iface_tx_drops === 'number') r.txDrops = f.iface_tx_drops
  }
  r.seq++
}

export function pushFrame(ts: number, f: { rtt_p50_ms?: number; rtt_p95_ms?: number; small_p95_ms?: number; bulk_goodput_mbps?: number; phase?: string; deadline_ok_pct?: number; wasted_bytes?: number; cost_ar_per_h?: number; drops?: number }) {
  live.ts = ts
  live.seq++
  if (f.phase && f.phase !== live.phase) {
    live.phase = f.phase
    if (!live.phaseSince[f.phase]) live.phaseSince[f.phase] = ts
  }
  push(live.rtt50, ts, f.rtt_p50_ms ?? null)
  push(live.rtt95, ts, f.rtt_p95_ms ?? null)
  push(live.small, ts, f.small_p95_ms ?? null)
  // les artefacts de compteur (reset de qdisc) n'entrent jamais dans l'anneau goodput
  const g = f.bulk_goodput_mbps
  push(live.goodput, ts, g != null && g >= 0 && g <= 2500 ? g : null)
  // séries propres par métrique — les sparklines du bandeau ne partagent jamais
  pushSparse(live.deadline, ts, f.deadline_ok_pct)
  pushSparse(live.wasted, ts, f.wasted_bytes)
  pushSparse(live.cost, ts, f.cost_ar_per_h)
  if (f.drops != null && Number.isFinite(f.drops)) pushSparse(live.drops, ts, f.drops)
}

export function clearLive() {
  live.rtt50.length = 0
  live.rtt95.length = 0
  live.small.length = 0
  live.goodput.length = 0
  live.deadline.length = 0
  live.wasted.length = 0
  live.cost.length = 0
  live.drops.length = 0
  // live.iface reste : l'observation d'interface continue hors campagne
  live.phase = ''
  live.phaseSince = {}
}

// instrumentation seam 3 — surgical logs + embed analysis + playwright clip
export const __CGO_LIVE = live
if (typeof window !== 'undefined') (window as any).__CGO_LIVE = live
