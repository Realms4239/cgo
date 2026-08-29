// live — mutable 10 Hz ring buffers (SPEC §2.4). Not in Zustand.
export type Pt = [number, number]
export type Ring = Pt[]

export const live = {
  rtt50: [] as Ring,
  rtt95: [] as Ring,
  small: [] as Ring,
  goodput: [] as Ring,
  max: 1800, // 180s one event @10Hz (Q53 B expand)
  ts: 0,
  phase: '', // baseline|charge|recup — from SSE delta, drives CHARGE markArea (not estimated)
  phaseSince: {} as Record<string, number>, // first ts seen per phase — drives Timeline 48 phase bands
}

function push(r: Ring, ts: number, v: number | null) {
  if (v == null || !Number.isFinite(v)) r.push([ts, 0])
  else r.push([ts, v])
  if (r.length > live.max) r.shift()
}

export function pushFrame(ts: number, f: { rtt_p50_ms?: number; rtt_p95_ms?: number; small_p95_ms?: number; bulk_goodput_mbps?: number; phase?: string }) {
  live.ts = ts
  if (f.phase && f.phase !== live.phase) {
    live.phase = f.phase
    if (!live.phaseSince[f.phase]) live.phaseSince[f.phase] = ts
  }
  push(live.rtt50, ts, f.rtt_p50_ms ?? null)
  push(live.rtt95, ts, f.rtt_p95_ms ?? null)
  push(live.small, ts, f.small_p95_ms ?? null)
  // counter artifacts (qdisc replacement resets) never enter the goodput ring
  const g = f.bulk_goodput_mbps
  push(live.goodput, ts, g != null && g >= 0 && g <= 2500 ? g : null)
}

export function clearLive() {
  live.rtt50.length = 0
  live.rtt95.length = 0
  live.small.length = 0
  live.goodput.length = 0
  live.phaseSince = {}
}

// instrumentation seam 3 — surgical logs + embed analysis + playwright clip
export const __CGO_LIVE = live
if (typeof window !== 'undefined') (window as any).__CGO_LIVE = live
