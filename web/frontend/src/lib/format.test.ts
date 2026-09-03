import { fmtIQR } from './format'
import { describe, expect, it } from 'vitest'

describe('fmtIQR', () => {
  it('rend med [low–high] à 1 décimale', () => {
    expect(fmtIQR(209.25, [207.1, 212.4])).toBe('209.3 [207.1–212.4]')
  })
  it('sans IQR rend la médiane seule', () => {
    expect(fmtIQR(209.25, undefined)).toBe('209.3')
  })
})
