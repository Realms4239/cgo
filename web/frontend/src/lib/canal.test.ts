import { describe, it, expect } from 'vitest'
import { canalGate, effectiveChannelModel } from './canal'

// La règle UI doit être exactement celle du serveur (400 avant mutation) :
// bursty n'est légal que si TOUS les profils sont P2/P3.
describe('canalGate', () => {
  it('autorise bursty sur P2 et P2/P3', () => {
    expect(canalGate(['P2'])).toEqual({ allowed: true, reason: '' })
    expect(canalGate(['P2', 'P3']).allowed).toBe(true)
  })

  it('refuse hors P2/P3 en nommant les fautifs et la règle', () => {
    const g = canalGate(['P1', 'P4'])
    expect(g.allowed).toBe(false)
    expect(g.reason).toContain('P1')
    expect(g.reason).toContain('P4')
    expect(g.reason).toContain('P2/P3')
    expect(g.reason).toContain('clean')
  })

  it('refuse un mélange P2 + P4 (le profil fautif suffit)', () => {
    const g = canalGate(['P2', 'P4'])
    expect(g.allowed).toBe(false)
    expect(g.reason).toContain('P4')
    expect(g.reason).not.toContain('P2,')
  })

  it('refuse sans profil — rien à mesurer', () => {
    expect(canalGate([]).allowed).toBe(false)
    expect(canalGate([]).reason).not.toBe('')
  })
})

describe('effectiveChannelModel', () => {
  it('retombe sur clean quand bursty est refusé', () => {
    expect(effectiveChannelModel('bursty', canalGate(['P1']))).toBe('clean')
    expect(effectiveChannelModel('bursty', canalGate(['P2']))).toBe('bursty')
    expect(effectiveChannelModel('clean', canalGate(['P1']))).toBe('clean')
  })
})
