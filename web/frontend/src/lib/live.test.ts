import { describe, it, expect, beforeEach } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { live, pushIface, clearLive } from './live'
import type { LiveFrame } from './types'

function frame(ts: number, over: Partial<LiveFrame> = {}): Partial<LiveFrame> {
  return {
    iface_name: 'c-cli',
    iface_source: 'twin',
    iface_sample_ts: ts,
    iface_valid: true,
    iface_rx_mbps: 10,
    iface_tx_mbps: 2,
    iface_rx_drops: 1,
    iface_tx_drops: 2,
    iface_window_ms: 100,
    iface_gap: '',
    ...over,
  }
}

beforeEach(() => {
  const r = live.iface
  r.rx.length = 0
  r.tx.length = 0
  r.name = ''
  r.source = ''
  r.valid = false
  r.gap = ''
  r.windowMs = 0
  r.rxDrops = 0
  r.txDrops = 0
  r.ts = 0
  r.seq = 0
})

describe('pushIface', () => {
  it('ignore un timestamp identique (frame répétée)', () => {
    pushIface(frame(1000))
    pushIface(frame(1000))
    expect(live.iface.rx.length).toBe(1)
    expect(live.iface.seq).toBe(1)
  })

  it('un échantillon invalide casse la série avec null', () => {
    pushIface(frame(1000))
    pushIface(frame(1100, { iface_valid: false, iface_gap: 'absent' }))
    expect(live.iface.rx.at(-1)).toEqual([1100, null])
    expect(live.iface.tx.at(-1)).toEqual([1100, null])
    expect(live.iface.gap).toBe('absent')
  })

  it('un changement de nom vide l’anneau', () => {
    pushIface(frame(1000))
    pushIface(frame(1100, { iface_name: 'e-srv' }))
    expect(live.iface.rx.length).toBe(1)
    expect(live.iface.name).toBe('e-srv')
  })

  it('un saut d’horloge arrière > 1 s vide l’anneau', () => {
    pushIface(frame(5000))
    pushIface(frame(3000))
    expect(live.iface.rx.length).toBe(1)
  })

  it('un échantillon en retard ≤ 1 s est ignoré', () => {
    pushIface(frame(5000))
    pushIface(frame(4500))
    expect(live.iface.rx.length).toBe(1)
  })

  it('clearLive ne vide pas l’interface', () => {
    pushIface(frame(1000))
    clearLive()
    expect(live.iface.rx.length).toBe(1)
  })

  it('plafonne à 600 points', () => {
    for (let i = 0; i < 601; i++) pushIface(frame(1000 + i * 100))
    expect(live.iface.rx.length).toBe(600)
  })

  it('pousse un frame idle sans running ni phase', () => {
    pushIface(frame(1000))
    expect(live.iface.rx.length).toBe(1)
  })
})

describe('sse', () => {
  it('pousse l’interface avant la porte measured', () => {
    const cands = ['src/lib/sse.ts', 'web/frontend/src/lib/sse.ts']
    let s = ''
    for (const p of cands) {
      try { s = readFileSync(p, 'utf8'); break } catch {}
    }
    expect(s).toContain('pushIface(data)')
    expect(s.indexOf('pushIface(data)')).toBeLessThan(s.indexOf('const measured'))
  })
})
