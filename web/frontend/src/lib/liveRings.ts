// liveRings — anneaux singletons partagés entre le HUD et le flux live.
// Une valeur absente devient NaN (jamais un 0 fabriqué, SPEC §2.2).
import { DomainEventRing, TypedMetricRing } from './ringBuffer'
import type { LiveFrame } from './types'

export const MILESTONE_KINDS = [
  'CALIBRATION_BASELINE', 'BULK_SATURATION_START', 'SLA_BREACH_220MS',
  'QDISC_DROP_BURST', 'AQM_STABILIZED', 'DUAL_SEAL_COMMITTED',
] as const

export const metricRing = new TypedMetricRing(1800)
export const eventRing = new DomainEventRing()
let eventID = 0

export function pushLiveFrame(f: Partial<LiveFrame>): void {
  metricRing.push({
    ts: f.ts ?? 0,
    throughputMbps: f.bulk_goodput_mbps ?? null,
    rttMs: f.rtt_p95_ms ?? null,
    dropsBytes: f.drops ?? null,
  })
}

export function pushLiveEvent(kind: string, detail: Record<string, unknown>): void {
  eventID += 1
  eventRing.push({
    id: eventID,
    tsMonoMs: typeof performance !== 'undefined' ? performance.now() : 0,
    kind,
    severity: kind === 'SLA_BREACH_220MS' || kind === 'QDISC_DROP_BURST' ? 'error' : 'info',
    title: kind.replaceAll('_', ' '),
    detail: JSON.stringify(detail),
  })
}

export function resetLiveRings(): void {
  metricRing.reset()
  eventRing.reset()
  eventID = 0
}
