// DOM/node oracle — client-side epoch reset (parked-clearance Task 1).
// Only the transport (EventSource) is faked; the real onmessage path, the
// instrument rings, the live state and the store are exercised. Absence stays
// NaN, never 0. These tests must be RED before the epoch logic exists.
import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { live, clearLive, pushFrame } from './live'
import { connectSSE, disconnectSSE, ingestFrame } from './sse'
import { instrument } from './ringBuffer'
import { useUIStore } from '../store/ui'

class FakeES {
  static last: FakeES | null = null
  onmessage: ((ev: { data: string }) => void) | null = null
  onerror: (() => void) | null = null
  private listeners: Record<string, ((ev: any) => void)[]> = {}
  closed = false
  url: string
  constructor(url: string) {
    this.url = url
    FakeES.last = this
  }
  addEventListener(type: string, fn: (ev: any) => void) {
    ;(this.listeners[type] ||= []).push(fn)
  }
  close() { this.closed = true }
}

const RINGS = [instrument.small, instrument.rtt50, instrument.rtt95,
  instrument.goodput, instrument.ifaceRx, instrument.ifaceTx, instrument.events]

beforeEach(() => {
  clearLive()
  ;(live as any).epoch = ''
  live.phase = ''
  for (const r of RINGS) r.clear()
  useUIStore.setState({ replayRunning: false, live: null, connected: false })
  ;(globalThis as any).EventSource = FakeES
  FakeES.last = null
})

afterEach(() => {
  disconnectSSE()
})

describe('SSE epoch reset (client)', () => {
  it('epoch mismatch clears stale rings and applies the full snapshot before deltas', () => {
    // Previous generation: both live and instrument rings carry measurements.
    pushFrame(1000, { phase: 'charge', rtt_p50_ms: 42, small_p95_ms: 30 } as any)
    ingestFrame({ ts: 1000, running: true, phase: 'charge', rtt_p50_ms: 42, small_p95_ms: 30 } as any)
    expect(live.rtt50.length).toBeGreaterThan(0)
    expect(instrument.rtt50.size).toBe(1)
    expect(instrument.rtt50.series()[0]).toBe(42)

    connectSSE()
    const first = FakeES.last!
    expect(first.url).toBe('/api/stream')

    // Full structural frame from a NEW hub generation (epoch B).
    first.onmessage!({ data: JSON.stringify({
      epoch: 'B', ts: 2000, running: true, phase: 'recup', profile: 'NEW',
      qdisc: 'cake', cc: 'bbr', gates: [false], rtt_p50_ms: 7, small_p95_ms: 9,
    }) })

    expect((live as any).epoch).toBe('B')
    // stale samples gone: only the new full-snapshot sample remains
    expect(instrument.rtt50.size).toBe(1)
    expect(instrument.rtt50.series()[0]).toBe(7)
    expect(instrument.small.size).toBe(1)
    expect(instrument.small.series()[0]).toBe(9)
    // the full structural snapshot is applied to the store (not stale values)
    const st = useUIStore.getState().live as any
    expect(st.phase).toBe('recup')
    expect(st.profile).toBe('NEW')
  })

  it('same epoch does NOT reset; absence stays NaN, never 0', () => {
    connectSSE()
    const first = FakeES.last!
    first.onmessage!({ data: JSON.stringify({
      epoch: 'B', ts: 2000, running: true, phase: 'recup',
      profile: 'NEW', small_p95_ms: 9,
    }) })
    expect((live as any).epoch).toBe('B')
    const reopened = FakeES.last!
    expect(reopened).not.toBe(first)
    expect(reopened.url).toBe('/api/stream?epoch=B')

    // Same generation: a delta without rtt_p50 accumulates (no reset), and the
    // missing measurement is NaN — never a fabricated 0.
    reopened.onmessage!({ data: JSON.stringify({
      epoch: 'B', ts: 2100, running: true, small_p95_ms: 10,
    }) })
    expect(instrument.small.size).toBe(2)
    expect(instrument.rtt50.size).toBe(2)
    expect(instrument.rtt50.series()[0]).toBeNaN()
    expect(instrument.rtt50.series()[1]).toBeNaN()
  })
})
