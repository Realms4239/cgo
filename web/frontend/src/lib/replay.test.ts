// Frontière de vérité du rejeu : une colonne absente d'un aqm_eval.csv arrive
// comme "" (results.Field) et doit rester une absence (NaN), jamais un 0 mesuré.
// EventSource est une API de navigateur, pas le module sous test : le stub ne
// remplace que le transport, jamais replay.ts, live.ts ni le store.
import { afterEach, describe, expect, it } from 'vitest'
import { startReplay, stopReplay } from './replay'
import { live } from './live'
import { useUIStore } from '../store/ui'

class CapturingEventSource {
  static instances: CapturingEventSource[] = []
  onmessage: ((ev: MessageEvent) => void) | null = null
  addEventListener() { /* aucun événement pendant ce test */ }
  close() { /* rien à fermer */ }
  constructor() { CapturingEventSource.instances.push(this) }
}

const env = globalThis as unknown as { EventSource: typeof EventSource }
const originalEventSource = env.EventSource

afterEach(() => {
  stopReplay()
  env.EventSource = originalEventSource
  CapturingEventSource.instances.length = 0
})

describe('startReplay — absence de mesure, jamais un zéro fabriqué', () => {
  it('une colonne small_p95 absente du CSV reste NaN dans l’anneau et le store', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    const es = CapturingEventSource.instances[0]
    expect(es).toBeDefined()

    es.onmessage?.({
      data: JSON.stringify({
        ts: 1000, phase: 'replay',
        rtt_p50_ms: '110.0', rtt_p95_ms: '111.0',
        small_p95_ms: '', // results.Field : champ manquant = ""
        bulk_goodput_mbps: '7.3', drops: '0',
      }),
    } as MessageEvent)

    expect(live.small.at(-1)?.[1]).toBeNaN()
    expect(live.small.at(-1)?.[1]).not.toBe(0)
    expect(useUIStore.getState().live?.small_p95_ms).toBeNaN()
    // hostile : les vraies mesures restent des nombres, pas une absence globale
    expect(live.rtt50.at(-1)?.[1]).toBe(110)
    expect(live.goodput.at(-1)?.[1]).toBe(7.3)
  })

  it('une sonde vide (small_p95_ms = 0) reste une absence, pas un 0 ms tracé', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    CapturingEventSource.instances[0].onmessage?.({
      data: JSON.stringify({ ts: 2000, phase: 'replay', small_p95_ms: '0', rtt_p50_ms: '110' }),
    } as MessageEvent)
    expect(live.small.at(-1)?.[1]).toBeNaN()
    expect(useUIStore.getState().live?.small_p95_ms).toBeNaN()
  })
})

describe('startReplay — compteurs et identifiants : absence jamais un zéro fabriqué', () => {
  it('colonnes vides (results.Field) → NaN pour drops, repetition, event_id', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    CapturingEventSource.instances[0].onmessage?.({
      data: JSON.stringify({
        ts: 3000, phase: 'replay', drops: '', repetition: '', event_id: '',
      }),
    } as MessageEvent)
    const frame = useUIStore.getState().live
    expect(frame?.drops).toBeNaN()
    expect(frame?.repetition).toBeNaN()
    expect(frame?.event_id).toBeNaN()
  })

  it('clés entièrement absentes → NaN, pas un 0 mesuré', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    CapturingEventSource.instances[0].onmessage?.({
      data: JSON.stringify({ ts: 3100, phase: 'replay', rtt_p50_ms: '110' }),
    } as MessageEvent)
    const frame = useUIStore.getState().live
    expect(frame?.drops).toBeNaN()
    expect(frame?.repetition).toBeNaN()
    expect(frame?.event_id).toBeNaN()
  })

  it('valeurs malformées → NaN (échec fermé)', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    CapturingEventSource.instances[0].onmessage?.({
      data: JSON.stringify({
        ts: 3200, phase: 'replay', drops: 'abc', repetition: 'n/a', event_id: 'x-9',
      }),
    } as MessageEvent)
    const frame = useUIStore.getState().live
    expect(frame?.drops).toBeNaN()
    expect(frame?.repetition).toBeNaN()
    expect(frame?.event_id).toBeNaN()
  })

  it('hostile : un vrai "0" archivé reste 0, pas une absence', () => {
    env.EventSource = CapturingEventSource as unknown as typeof EventSource
    startReplay('run-test')
    CapturingEventSource.instances[0].onmessage?.({
      data: JSON.stringify({
        ts: 3300, phase: 'replay', drops: '0', repetition: '0', event_id: '0',
      }),
    } as MessageEvent)
    const frame = useUIStore.getState().live
    expect(frame?.drops).toBe(0)
    expect(frame?.repetition).toBe(0)
    expect(frame?.event_id).toBe(0)
  })
})
