import { describe, expect, it } from 'vitest'
import { TypedMetricRing } from './ringBuffer'

describe('TypedMetricRing', () => {
  it('stores absent metrics as NaN, never zero', () => {
    const r = new TypedMetricRing(4)
    r.push({ ts: 1, throughputMbps: NaN, rttMs: 42, dropsBytes: null })
    expect(Number.isNaN(r.at(0)?.throughputMbps)).toBe(true)
    expect(Number.isNaN(r.at(0)?.dropsBytes)).toBe(true)
    expect(r.at(0)?.rttMs).toBe(42)
    expect(r.measured(0, 'rttMs')).toBe(true)
    expect(r.measured(0, 'throughputMbps')).toBe(false)
  })

  it('wraps in O(1) and keeps the last capacity samples', () => {
    const r = new TypedMetricRing(3)
    for (let i = 1; i <= 5; i += 1) r.push({ ts: i, throughputMbps: i, rttMs: i, dropsBytes: i })
    expect(r.length).toBe(3)
    expect(r.at(0)?.ts).toBe(3)
    expect(r.at(-1)?.ts).toBe(5)
  })

  it('tracks dirty until the painter acknowledges', () => {
    const r = new TypedMetricRing(2)
    expect(r.dirty).toBe(false)
    r.push({ ts: 1, throughputMbps: 1, rttMs: 2, dropsBytes: 0 })
    expect(r.dirty).toBe(true)
    r.markPainted()
    expect(r.dirty).toBe(false)
  })
})
