// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { eventRing, metricRing, pushLiveEvent, pushLiveFrame, resetLiveRings } from '../lib/liveRings'
import { RingBufferHUD } from './RingBufferHUD'

// La boucle rAF est pilotée à la main : pas de timing réel, une frame = un tick.
let frames: FrameRequestCallback[] = []

function mount() {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(<RingBufferHUD />) })
  return { host, root }
}

function flushFrame() {
  const pending = frames
  frames = []
  for (const cb of pending) act(() => { cb(performance.now()) })
}

const text = (host: HTMLElement, key: string) => host.querySelector(`[data-hud="${key}"]`)?.textContent ?? ''

describe('RingBufferHUD', () => {
  beforeEach(() => {
    resetLiveRings()
    frames = []
    vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback): number => {
      frames.push(cb)
      return frames.length
    })
    vi.stubGlobal('cancelAnimationFrame', () => {})
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('sans trame, les compteurs affichent le tiret et le badge reste protégé', () => {
    const { host } = mount()
    expect(text(host, 'throughput')).toBe('—')
    expect(text(host, 'rtt')).toBe('—')
    expect(text(host, 'drops')).toBe('—')
    expect(text(host, 'event')).toBe('—')
    expect(text(host, 'status')).toBe('PROTÉGÉ')
  })

  it('une trame p95=260 ms atteint le HUD via la boucle rAF et marque la peinture', () => {
    const { host } = mount()
    pushLiveFrame({ ts: 1, rtt_p95_ms: 260, bulk_goodput_mbps: 9.4, drops: 0 })
    expect(metricRing.dirty).toBe(true)
    flushFrame()
    expect(text(host, 'rtt')).toContain('260')
    expect(text(host, 'throughput')).toContain('9.4')
    expect(text(host, 'drops')).toBe('0')
    expect(text(host, 'status')).toBe('CONGESTION > 220 ms')
    expect(metricRing.dirty).toBe(false)
  })

  it('un rtt sous le seuil repasse en PROTÉGÉ sans effacer la mesure', () => {
    const { host } = mount()
    pushLiveFrame({ ts: 1, rtt_p95_ms: 260 })
    flushFrame()
    expect(text(host, 'status')).toBe('CONGESTION > 220 ms')
    pushLiveFrame({ ts: 2, rtt_p95_ms: 120 })
    flushFrame()
    expect(text(host, 'status')).toBe('PROTÉGÉ')
    expect(text(host, 'rtt')).toContain('120')
  })

  it('le dernier événement de domaine est affiché', () => {
    const { host } = mount()
    pushLiveEvent('SLA_BREACH_220MS', { rtt_ms: 260 })
    flushFrame()
    expect(text(host, 'event')).toContain('220')
  })

  it('annule la frame en attente au démontage', () => {
    const cancel = vi.fn()
    vi.stubGlobal('cancelAnimationFrame', cancel)
    const { root } = mount()
    act(() => { root.unmount() })
    expect(cancel).toHaveBeenCalledTimes(1)
    expect(eventRing.length).toBe(0)
  })
})
