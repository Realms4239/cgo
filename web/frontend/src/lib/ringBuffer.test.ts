// Les anneaux portent la mesure, pas un zéro de fortune : une sonde muette est
// NaN, un zéro mesuré reste 0. Les témoins ci-dessous lisent la série réelle.
import { describe, expect, it } from 'vitest'
import { TypedMetricRing, DomainEventRing } from './ringBuffer'

describe('TypedMetricRing — honnêteté de l’absence', () => {
  it('push(null) garde NaN, jamais 0', () => {
    const r = new TypedMetricRing(8)
    r.push(1000, null)
    expect(r.size).toBe(1)
    expect(r.series()[0]).toBeNaN()
    // hostile : un 0 fabriqué ferait passer ce témoin
    expect(r.series()[0]).not.toBe(0)
  })

  it('push(0) garde 0 — un zéro mesuré n’est pas une absence', () => {
    const r = new TypedMetricRing(8)
    r.push(1000, 0)
    expect(r.series()[0]).toBe(0)
    expect(Number.isNaN(r.series()[0])).toBe(false)
  })

  it('un nombre non fini est une absence, pas un 0', () => {
    const r = new TypedMetricRing(8)
    r.push(1000, Number.POSITIVE_INFINITY)
    expect(r.series()[0]).toBeNaN()
  })

  it('l’anneau déborde en conservant les capacity dernières valeurs, dans l’ordre', () => {
    const r = new TypedMetricRing(3)
    for (let i = 1; i <= 5; i++) r.push(i * 100, i)
    expect(r.size).toBe(3)
    expect(Array.from(r.series())).toEqual([3, 4, 5])
    expect(Array.from(r.timestamps())).toEqual([300, 400, 500])
  })

  it('le drapeau dirty est posé à chaque écriture', () => {
    const r = new TypedMetricRing(4)
    expect(r.dirty).toBe(false)
    r.push(1, 10)
    expect(r.dirty).toBe(true)
    r.dirty = false
    r.push(2, null)
    expect(r.dirty).toBe(true)
  })

  it('pairs traduit NaN en null pour que connectNulls:false casse le tracé', () => {
    const r = new TypedMetricRing(8)
    r.push(100, 24)
    r.push(200, null)
    r.push(300, 0)
    expect(r.pairs()).toEqual([[100, 24], [200, null], [300, 0]])
  })

  it('clear rend l’anneau vide et marque le changement', () => {
    const r = new TypedMetricRing(4)
    r.push(1, 5)
    r.dirty = false
    r.clear()
    expect(r.size).toBe(0)
    expect(r.series()).toHaveLength(0)
    expect(r.dirty).toBe(true)
  })
})

describe('DomainEventRing — journal des faits', () => {
  it('list renvoie les événements dans l’ordre, capacité bornée', () => {
    const ring = new DomainEventRing(2)
    ring.push({ ts: 1, kind: 'phase', label: 'Charge' })
    ring.push({ ts: 2, kind: 'phase', label: 'Récupération' })
    ring.push({ ts: 3, kind: 'gates', label: 'Portes fermées' })
    expect(ring.list().map(e => e.ts)).toEqual([2, 3])
    expect(ring.size).toBe(2)
  })
})
