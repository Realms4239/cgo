import { describe, expect, it } from 'vitest'
import { nextStep, pickHeadline } from './synthese'

const groups = [
  { profile: 'P2', qdisc: 'pfifo_fast', cc: 'bbr', deadline_median: 60.2, deadline_valid_n: 22 },
  { profile: 'P2', qdisc: 'cake', cc: 'bbr', deadline_median: 98.1, deadline_valid_n: 20 },
  { profile: 'P2', qdisc: 'fq_codel', cc: 'bbr', deadline_median: 96.2, deadline_valid_n: 18 },
  { profile: 'P1', qdisc: 'cake', cc: 'bbr', deadline_median: 100, deadline_valid_n: 16 },
]

describe('pickHeadline', () => {
  it('choisit la file simple et la meilleure cellule P2 exploitables', () => {
    const h = pickHeadline(groups)
    expect(h.simple?.qdisc).toBe('pfifo_fast')
    expect(h.best?.deadline_median).toBe(98.1)
  })
  it('ignore les cellules sans échantillon valide', () => {
    expect(pickHeadline([{ profile: 'P2', qdisc: 'cake', cc: 'bbr', deadline_median: 99, deadline_valid_n: 0 }])).toEqual({})
  })
})

describe('nextStep', () => {
  it('suit une campagne en cours', () => {
    expect(nextStep(true, 159).mode).toBe('bench')
  })
  it('ne prétend pas lire un corpus absent', () => {
    expect(nextStep(false, 0).title).toContain('Aucun corpus')
  })
  it('propose un audit quand le corpus existe', () => {
    expect(nextStep(false, 159).body).toContain('159 runs')
  })
})
