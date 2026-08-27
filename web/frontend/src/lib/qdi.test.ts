import { describe, it, expect } from 'vitest'
import { computeQDI } from './qdi'
describe('qdi', () => {
  it('computes queue delay increase', () => {
    expect(computeQDI(50, 20)).toBe(30)
    expect(computeQDI(20, 20)).toBe(0)
  })
})
