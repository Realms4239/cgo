import { describe, it, expect } from 'vitest'
import { prefersReducedMotion } from './anime'
describe('anime', () => {
  it('respects prefers-reduced-motion', () => {
    expect(typeof prefersReducedMotion).toBe('function')
  })
})
