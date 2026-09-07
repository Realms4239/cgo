import { describe, it, expect } from 'vitest'
import { paginate, pageList, clampPage, totalPagesOf } from './Paginate'

describe('paginate', () => {
  it('slice zéro-index', () => {
    const a = [1, 2, 3, 4, 5, 6, 7]
    expect(paginate(a, 1, 3)).toEqual([1, 2, 3])
    expect(paginate(a, 2, 3)).toEqual([4, 5, 6])
    expect(paginate(a, 3, 3)).toEqual([7])
    // page hors bornes = CLAMPÉE sur la dernière page existante (jamais de
    // slice vide derrière un total qui a rétréci)
    expect(paginate(a, 9, 3)).toEqual([7])
    expect(paginate(a, 0, 3)).toEqual([1, 2, 3])
  })
  it('pageList smart ellipsis — 1 … 4 5 6 … 42', () => {
    expect(pageList(5, 42)).toEqual([1, '…', 4, 5, 6, '…', 42])
    expect(pageList(1, 42)).toEqual([1, 2, '…', 42])
    expect(pageList(42, 42)).toEqual([1, '…', 41, 42])
    expect(pageList(2, 4)).toEqual([1, 2, 3, 4])
    expect(pageList(1, 1)).toEqual([1])
  })
  it('clampPage — borne la page sur l\'existant', () => {
    expect(totalPagesOf(7, 3)).toBe(3)
    expect(clampPage(9, 7, 3)).toBe(3)
    expect(clampPage(0, 7, 3)).toBe(1)
    expect(clampPage(2, 7, 3)).toBe(2)
  })
})
