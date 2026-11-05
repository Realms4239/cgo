import { describe, it, expect } from 'vitest'
import { cellEncoding, chipGroups, fmtHalf, halfInterval } from './results'

const distinctOf = (vals: Record<string, string[]>) => (k: 'profile' | 'qdisc' | 'cc') => vals[k] ?? []

describe('chipGroups', () => {
  it('rend trois groupes étiquetés PROFIL / FILE / CC avec « tous » en tête', () => {
    const gs = chipGroups(distinctOf({ profile: ['P1', 'P2'], qdisc: ['cake', 'pfifo_fast'], cc: ['bbr', 'cubic'] }))
    expect(gs.map(g => g.label)).toEqual(['PROFIL', 'FILE', 'CC'])
    expect(gs.map(g => g.key)).toEqual(['profile', 'qdisc', 'cc'])
    for (const g of gs) expect(g.values[0]).toBe('tous')
    expect(gs[1].values).toEqual(['tous', 'cake', 'pfifo_fast'])
  })

  it('un groupe vide reste « tous » seul — jamais une rangée muette', () => {
    const gs = chipGroups(distinctOf({}))
    expect(gs.every(g => g.values.length === 1 && g.values[0] === 'tous')).toBe(true)
  })
})

describe('cellEncoding', () => {
  it('la primaire garde la barre, la secondaire reste nue', () => {
    expect(cellEncoding(true)).toBe('bar')
    expect(cellEncoding(false)).toBe('plain')
  })
})

describe('dispersion', () => {
  it('demi-intervalle réel ou rien', () => {
    expect(halfInterval(10, 14)).toBe(2)
    expect(halfInterval(10, 10)).toBeNull()
    expect(halfInterval(14, 10)).toBeNull()
  })
  it('format court au-delà de 10', () => {
    expect(fmtHalf(0.42)).toBe('±0.4')
    expect(fmtHalf(12.7)).toBe('±13')
  })
})
