import { formatBytesFR, formatCostAR, groupDigits, truncateWords, verdictWord } from './format'
import { describe, expect, it } from 'vitest'

describe('groupDigits', () => {
  it('groupe par trois en espaces fines', () => {
    expect(groupDigits(12400)).toBe('12 400')
    expect(groupDigits(68.5, 2)).toBe('68,50')
  })
  it('non-fini → —, jamais inventé', () => {
    expect(groupDigits(NaN)).toBe('—')
  })
})

describe('formatBytesFR', () => {
  it('octets français, jamais bytes/MiB', () => {
    expect(formatBytesFR(512)).toEqual({ num: '512', unit: 'o' })
    expect(formatBytesFR(2048).unit).toBe('Kio')
    expect(formatBytesFR(12 * 1024 * 1024 + 400 * 1024).unit).toBe('Mio')
  })
  it('null → — sans unité', () => {
    expect(formatBytesFR(null)).toEqual({ num: '—', unit: '' })
    expect(formatBytesFR(undefined)).toEqual({ num: '—', unit: '' })
  })
})

describe('formatCostAR', () => {
  it('Ariary groupés à 2 décimales', () => {
    expect(formatCostAR(68.5)).toEqual({ num: '68,50', unit: 'Ar/h' })
  })
  it('null → — sans unité', () => {
    expect(formatCostAR(null)).toEqual({ num: '—', unit: '' })
  })
})

describe('verdictWord', () => {
  it('un mot par niveau, jamais la couleur seule', () => {
    expect(verdictWord('ok')).toBe('bon')
    expect(verdictWord('warn')).toBe('fragile')
    expect(verdictWord('crit')).toBe('mauvais')
  })
})

describe('truncateWords', () => {
  it('coupe à la frontière de mot, jamais en plein mot', () => {
    expect(truncateWords('alpha beta gamma delta', 12)).toBe('alpha beta…')
  })
  it('court → inchangé', () => {
    expect(truncateWords('court', 120)).toBe('court')
  })
})
