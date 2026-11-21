// Frontière de vérité du flux : une trame mesurée sans rtt_p50 laisse NaN,
// jamais un 0 fabriqué. Les trames idle ne poussent rien.
import { describe, expect, it } from 'vitest'
import { ingestFrame, shouldSyncStore } from './sse'
import { TypedMetricRing, type InstrumentRings } from './ringBuffer'

function rings(): InstrumentRings {
  return {
    small: new TypedMetricRing(8),
    rtt50: new TypedMetricRing(8),
    rtt95: new TypedMetricRing(8),
    goodput: new TypedMetricRing(8),
    ifaceRx: new TypedMetricRing(8),
    ifaceTx: new TypedMetricRing(8),
  }
}

describe('ingestFrame — NaN, jamais zéro', () => {
  it('une trame mesurée sans rtt_p50 laisse NaN dans rtt50', () => {
    const r = rings()
    const pushed = ingestFrame({
      ts: 1000, running: true, phase: 'baseline',
      small_p95_ms: 28, rtt_p95_ms: 30, bulk_goodput_mbps: 90,
    } as any, r)
    expect(pushed).toBe(true)
    expect(r.small.series()[0]).toBe(28)
    expect(r.rtt50.size).toBe(1)
    expect(r.rtt50.series()[0]).toBeNaN()
    expect(r.rtt50.series()[0]).not.toBe(0)
  })

  it('un rtt_p50 mesuré à zéro reste 0 (aucune confusion avec une absence)', () => {
    const r = rings()
    ingestFrame({ ts: 1000, running: true, phase: 'baseline', rtt_p50_ms: 0, small_p95_ms: 28 } as any, r)
    expect(r.rtt50.series()[0]).toBe(0)
  })

  it('une trame idle (rien de mesuré) n’est pas poussée', () => {
    const r = rings()
    const pushed = ingestFrame({ ts: 1000, running: false, phase: 'idle', drops: null } as any, r)
    expect(pushed).toBe(false)
    expect(r.small.size).toBe(0)
    expect(r.rtt50.size).toBe(0)
    expect(r.goodput.size).toBe(0)
  })

  it('une trame tout-zéro en course est un trou de pompe, pas une mesure', () => {
    const r = rings()
    const pushed = ingestFrame({
      ts: 1000, running: true, phase: 'baseline',
      rtt_p50_ms: 0, rtt_p95_ms: 0, small_p95_ms: 0, bulk_goodput_mbps: 0, drops: 0,
    } as any, r)
    expect(pushed).toBe(false)
    expect(r.small.size).toBe(0)
  })

  it('un bon débit hors bornes n’entre pas dans l’anneau goodput', () => {
    const r = rings()
    ingestFrame({ ts: 1000, running: true, phase: 'charge', bulk_goodput_mbps: 99999 } as any, r)
    expect(r.goodput.series()[0]).toBeNaN()
  })
})

describe('shouldSyncStore — 4 Hz', () => {
  it('sync immédiate sur un changement structurel', () => {
    expect(shouldSyncStore(100, 90, true)).toBe(true)
  })
  it('sync au plus toutes les 250 ms', () => {
    expect(shouldSyncStore(1000, 700, false)).toBe(true)
    expect(shouldSyncStore(949, 700, false)).toBe(false)
    expect(shouldSyncStore(1200, 700, false)).toBe(true)
  })
})
