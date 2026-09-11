import { fmtIQR, improve, fmtCount, verdictHeadline, shouldLogScale } from './format'
import { describe, expect, it } from 'vitest'

describe('fmtIQR', () => {
  it('rend med [low–high] à 1 décimale', () => {
    expect(fmtIQR(209.25, [207.1, 212.4])).toBe('209.3 [207.1–212.4]')
  })
  it('sans IQR rend la médiane seule', () => {
    expect(fmtIQR(209.25, undefined)).toBe('209.3')
  })
})

describe('improve', () => {
  it('down : baisser = positif (mieux)', () => {
    expect(improve(100, 66, 'down')).toBe(34)
    expect(improve(100, 120, 'down')).toBe(-20)
  })
  it('up : monter = positif (mieux) — le bug goodput -3% vert', () => {
    expect(improve(100, 97, 'up')).toBe(-3) // pire → négatif, jamais vert
    expect(improve(100, 110, 'up')).toBe(10)
  })
  it('a=0 ou non-fini → null, jamais de % inventé', () => {
    expect(improve(0, 5, 'down')).toBeNull()
    expect(improve(NaN, 5, 'up')).toBeNull()
  })
})

describe('fmtCount', () => {
  it('entiers sans décimale — pertes 0, pas 0.0', () => {
    expect(fmtCount(0)).toBe('0')
    expect(fmtCount(12)).toBe('12')
    expect(fmtCount(12400)).toBe('12 400')
  })
  it('null/NaN → —', () => {
    expect(fmtCount(null)).toBe('—')
    expect(fmtCount(NaN)).toBe('—')
  })
})

describe('verdictHeadline', () => {
  it('nomme le gagnant, jamais le perdant seul', () => {
    expect(verdictHeadline(4, 1, 5, 'A', 'B')).toBe('4/5 métriques en faveur de A')
    expect(verdictHeadline(0, 5, 5, 'A', 'B')).toBe('5/5 métriques en faveur de B')
  })
  it('égalité dite une fois', () => {
    expect(verdictHeadline(2, 2, 5, 'A', 'B')).toBe('égalité 2–2 sur 5 métriques')
  })
})

describe('shouldLogScale', () => {
  it('738 vs 20 → log honnête', () => {
    expect(shouldLogScale([738.2, 121.4, 21.5])).toBe(true)
  })
  it('zéros ou faible dynamique → linéaire', () => {
    expect(shouldLogScale([0, 0, 0])).toBe(false)
    expect(shouldLogScale([20, 25, 30])).toBe(false)
    expect(shouldLogScale([42])).toBe(false)
  })
})
