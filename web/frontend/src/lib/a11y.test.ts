import { describe, expect, it } from 'vitest'
import { contrastRatio, wcagAaNormalText } from './a11y'

describe('a11y contrast', () => {
  it('orders black and white contrast', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 1)
    expect(contrastRatio('#000000', '#000000')).toBeCloseTo(1, 5)
  })

  it('reads computed rgb() colors', () => {
    expect(contrastRatio('rgb(0, 0, 0)', 'rgb(255, 255, 255)')).toBeCloseTo(21, 1)
  })

  it('requires WCAG AA normal text contrast', () => {
    expect(wcagAaNormalText(contrastRatio('#4b5563', '#ffffff'))).toBe(true)
    expect(wcagAaNormalText(contrastRatio('#9ca3af', '#ffffff'))).toBe(false)
  })
})
