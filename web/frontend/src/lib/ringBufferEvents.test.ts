import { describe, expect, it } from 'vitest'
import { DomainEventRing } from './ringBuffer'

const event = (id: number) => ({ id, tsMonoMs: id, kind: 'SLA_BREACH_220MS', severity: 'warn', title: `e${id}`, detail: '' })

describe('DomainEventRing', () => {
  it('wraps at 100 slots, oldest first', () => {
    const r = new DomainEventRing()
    for (let i = 1; i <= 102; i += 1) r.push(event(i))
    expect(r.length).toBe(100)
    expect(r.at(0)?.id).toBe(3)
    expect(r.at(-1)?.id).toBe(102)
  })

  it('returns the last n events oldest-first and resolves negative indices from the end', () => {
    const r = new DomainEventRing()
    for (let i = 1; i <= 5; i += 1) r.push(event(i))
    expect(r.tail(3).map((e) => e.id)).toEqual([3, 4, 5])
    expect(r.at(-2)?.id).toBe(4)
    expect(r.at(99)).toBeUndefined()
  })

  it('reset empties the ring so the next push starts at index 0', () => {
    const r = new DomainEventRing()
    r.push(event(1))
    r.reset()
    expect(r.length).toBe(0)
    expect(r.at(0)).toBeUndefined()
    r.push(event(2))
    expect(r.at(0)?.id).toBe(2)
  })
})
