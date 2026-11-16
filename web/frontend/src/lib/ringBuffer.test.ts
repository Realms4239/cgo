import { describe, expect, it } from 'vitest'
import { FloatRing, entries } from './ringBuffer'

describe('FloatRing', () => {
  it('rejette une capacité invalide', () => {
    expect(() => new FloatRing(0)).toThrow(RangeError)
    expect(() => new FloatRing(1.5)).toThrow(RangeError)
  })

  it('pousse et lit en ordre chronologique', () => {
    const r = new FloatRing(3)
    r.push(100, 1.5)
    r.push(200, 2.5)
    expect(r.length).toBe(2)
    expect(r.at(0)).toEqual([100, 1.5])
    expect(r.at(1)).toEqual([200, 2.5])
  })

  it('évince le plus ancien à capacité pleine, O(1)', () => {
    const r = new FloatRing(3)
    r.push(1, 1)
    r.push(2, 2)
    r.push(3, 3)
    r.push(4, 4)
    expect(r.length).toBe(3)
    expect(r.at(0)).toEqual([2, 2])
    expect(r.at(2)).toEqual([4, 4])
  })

  it('NaN = non mesuré, jamais converti en zéro', () => {
    const r = new FloatRing(2)
    r.push(1, NaN)
    r.push(2, 42)
    expect(r.measured(0)).toBe(false)
    expect(r.measured(1)).toBe(true)
    expect(r.at(0)?.[1]).toBeNaN()
    expect(r.tail(2).map(([, v]) => v)).toEqual([NaN, 42])
  })

  it('tail borne le n et conserve la chronologie', () => {
    const r = new FloatRing(10)
    for (let i = 0; i < 5; i += 1) r.push(i * 100, i)
    expect(r.tail(9999).map(([, v]) => v)).toEqual([0, 1, 2, 3, 4])
    expect(r.tail(-1)).toEqual([])
    expect(r.tail(2).map(([, v]) => v)).toEqual([3, 4])
  })

  it('slice gère les indices négatifs comme un tableau', () => {
    const r = new FloatRing(10)
    for (let i = 0; i < 4; i += 1) r.push(i, i)
    expect(r.slice(-2).map(([, v]) => v)).toEqual([2, 3])
    expect(r.slice(1, 3).map(([, v]) => v)).toEqual([1, 2])
    expect(r.slice(3, 1)).toEqual([])
  })

  it('since filtre par borne basse de ts', () => {
    const r = new FloatRing(10)
    for (let i = 0; i < 6; i += 1) r.push(i * 100, i * 10)
    expect(r.since(250).map(([, v]) => v)).toEqual([30, 40, 50])
    expect(r.since(0)).toHaveLength(6)
  })

  it('reset vide l’anneau et rend NaN partout', () => {
    const r = new FloatRing(2)
    r.push(1, 7)
    r.reset()
    expect(r.length).toBe(0)
    expect(r.at(0)).toBeUndefined()
  })

  it('entries se comporte comme un itérable tableau pour .map(([, v]) => v)', () => {
    const r = new FloatRing(5)
    r.push(1, 3.5)
    r.push(2, NaN)
    expect(entries(r).map(([, v]) => v)).toEqual([3.5, NaN])
  })
})
