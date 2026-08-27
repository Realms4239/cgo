import { describe, it, expect } from 'vitest'
import { computeJFI } from './jfi'
describe('jfi', () => {
  it('computes jains fairness', () => {
    expect(computeJFI([1,1,1])).toBeCloseTo(1)
    expect(computeJFI([1,0,0])).toBeCloseTo(0.33,1)
  })
})
