import { describe, expect, it } from 'vitest'
import { auditFacts, corpusFacts, ribbonMode, sealFacts } from './ribbon'

describe('ribbon', () => {
  it('ne fabrique pas un compte quand l’intégrité est absente', () => {
    expect(corpusFacts(null)).toBe('corpus —')
    expect(corpusFacts({ available: false, runs: 159 })).toBe('corpus —')
  })

  it('rend les quatre comptes gelés', () => {
    expect(corpusFacts({ available: true, runs: 159, valid: 208, degraded: 74, quarantined: 246 }))
      .toBe('159 runs · 208 valides · 74 dégradées · 246 écartées')
  })

  it('sceau absent = tiret, jamais un faux hash', () => {
    expect(sealFacts()).toBe('sceau —')
    expect(sealFacts('a1b2c3d4')).toBe('sceau a1b2c3d4')
  })

  it('dernier audit nommé par son site et sa note', () => {
    expect(auditFacts({ audit_id: 'a', site: 'Département X', measured: {}, bloat_grade: 'D' }))
      .toBe('dernier audit · Département X · D')
    expect(auditFacts(null)).toBe('dernier audit —')
  })

  it('mode observation contre mode complet', () => {
    expect(ribbonMode('observe')).toEqual({ text: 'observation', observation: true })
    expect(ribbonMode('full').observation).toBe(false)
  })
})
