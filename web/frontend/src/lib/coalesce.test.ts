import { describe, expect, it, vi } from 'vitest'
import { createCoalescer } from './coalesce'

describe('createCoalescer', () => {
  it('lance une seule fois par fenêtre de 20 ms, avec la dernière valeur', () => {
    vi.useFakeTimers()
    const seen: number[] = []
    const c = createCoalescer((v: number) => seen.push(v))
    c.schedule(1); c.schedule(2); c.schedule(3)
    expect(seen).toEqual([])
    vi.advanceTimersByTime(20)
    expect(seen).toEqual([3])
    vi.useRealTimers()
  })

  it('une rafale sous la fenêtre crée une seule exécution', () => {
    vi.useFakeTimers()
    const seen: string[] = []
    const c = createCoalescer((v: string) => seen.push(v))
    c.schedule('a'); c.schedule('b')
    vi.advanceTimersByTime(100)
    c.schedule('c')
    vi.advanceTimersByTime(19)
    expect(seen).toEqual(['b'])
    vi.advanceTimersByTime(1)
    expect(seen).toEqual(['b', 'c'])
    vi.useRealTimers()
  })

  it('flush applique immédiatement la valeur en attente', () => {
    vi.useFakeTimers()
    const seen: number[] = []
    const c = createCoalescer((v: number) => seen.push(v))
    c.schedule(7)
    c.flush()
    expect(seen).toEqual([7])
    vi.useRealTimers()
  })

  it('cancel abandonne la valeur en attente', () => {
    vi.useFakeTimers()
    const seen: number[] = []
    const c = createCoalescer((v: number) => seen.push(v))
    c.schedule(7)
    c.cancel()
    vi.advanceTimersByTime(100)
    expect(seen).toEqual([])
    vi.useRealTimers()
  })
})
