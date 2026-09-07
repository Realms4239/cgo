import { describe, it, expect } from 'vitest'
import { normDown, normUp, components, lienScore, type LienMetrics } from './lien'

const M = (o: Partial<LienMetrics>): LienMetrics => ({
  small: null, rtt: null, goodput: null, deadline: null, cost: null, ...o,
})

describe('lien', () => {
  it('perfect group scores 100', () => {
    const all = [
      M({ small: 10, rtt: 5, goodput: 100, deadline: 99, cost: 0 }),
      M({ small: 50, rtt: 40, goodput: 20, deadline: 10, cost: 500 }),
    ]
    expect(lienScore(all, all[0])).toBeCloseTo(100)
  })
  it('worst group scores 0', () => {
    const all = [
      M({ small: 10, rtt: 5, goodput: 100, deadline: 99, cost: 0 }),
      M({ small: 50, rtt: 40, goodput: 20, deadline: 10, cost: 500 }),
    ]
    expect(lienScore(all, all[1])).toBeCloseTo(0)
  })
  it('single valid metric still scores (mean of non-null)', () => {
    const all = [M({ goodput: 10 }), M({ goodput: 20 })]
    expect(lienScore(all, all[1])).toBeCloseTo(100)
    expect(lienScore(all, all[0])).toBeCloseTo(0)
    expect(lienScore(all, M({}))).toBeNull()
  })
  it('null handling: v null or empty column → null, all-null → null score', () => {
    expect(normDown([1, 2], null)).toBeNull()
    expect(normUp([], 5)).toBeNull()
    expect(normDown([null, null], 5)).toBeNull()
    expect(components([M({}), M({})], M({}))).toEqual({ small: null, rtt: null, goodput: null, deadline: null, cost: null })
    expect(lienScore([M({}), M({})], M({}))).toBeNull()
  })
  it('degenerate equal values → 1 per component → 100', () => {
    const all = [M({ small: 30, rtt: 10, goodput: 50, deadline: 80, cost: 100 }), M({ small: 30, rtt: 10, goodput: 50, deadline: 80, cost: 100 })]
    expect(components(all, all[0])).toEqual({ small: 1, rtt: 1, goodput: 1, deadline: 1, cost: 1 })
    expect(lienScore(all, all[0])).toBeCloseTo(100)
  })
})
