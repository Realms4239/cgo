import { beforeEach, describe, expect, it } from 'vitest'
import { eventRing, metricRing, pushLiveEvent, pushLiveFrame, resetLiveRings } from './liveRings'

describe('liveRings', () => {
  beforeEach(() => resetLiveRings())

  it('maps live ticks into the typed ring without inventing zeros', () => {
    pushLiveFrame({ ts: 1, rtt_p95_ms: 120, bulk_goodput_mbps: 9.4, drops: 0 })
    const last = metricRing.at(-1)
    expect(last?.rttMs).toBe(120)
    expect(last?.throughputMbps).toBe(9.4)
    expect(last?.dropsBytes).toBe(0)
    expect(metricRing.dirty).toBe(true)
    pushLiveFrame({ ts: 2 })
    expect(Number.isNaN(metricRing.at(-1)?.rttMs)).toBe(true)
  })

  it('stores milestone events in the domain ring', () => {
    pushLiveEvent('SLA_BREACH_220MS', { rtt_ms: 260 })
    expect(eventRing.at(-1)?.kind).toBe('SLA_BREACH_220MS')
    expect(eventRing.at(-1)?.title).toContain('220')
  })
})
