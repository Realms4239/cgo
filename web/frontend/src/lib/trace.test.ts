// La trace ne relie jamais deux sondes séparées par un silence : un trou est
// une absence de réponse, pas un 0 ms.
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { buildTrace, TRACE_HEIGHT } from './trace'
import { TraceCanvas } from '../components/TraceCanvas'
import type { AuditTimelinePoint } from './types'

const pt = (
  t: number,
  small: number | null,
  phase = 'repos',
): AuditTimelinePoint => ({
  t,
  phase,
  n: small == null ? 0 : 2,
  rtt_p50: small == null ? null : small - 2,
  rtt_p95: small == null ? null : small,
  small_p95: small,
})

describe('buildTrace', () => {
  it('ne relie jamais deux points séparés par une seconde muette', () => {
    const g = buildTrace([pt(0, 30), pt(1, null), pt(2, 220)], 220)
    expect(g.segments).toHaveLength(2)
    expect(g.segments[0].points.map(p => p.t)).toEqual([0])
    expect(g.segments[1].points.map(p => p.t)).toEqual([2])
    expect(g.yMax).toBeGreaterThanOrEqual(220)
    expect(g.objectiveY).toBeGreaterThan(0)
    expect(g.objectiveY).toBeLessThanOrEqual(TRACE_HEIGHT)
  })

  it('un segment continu reste un seul tracé', () => {
    const g = buildTrace([pt(0, 30), pt(1, 80), pt(2, 210)], 220)
    expect(g.segments).toHaveLength(1)
    expect(g.segments[0].points).toHaveLength(3)
  })

  it('des sondes absentes ne fabriquent aucune ligne', () => {
    const g = buildTrace([pt(0, null), pt(1, null)], 220)
    expect(g.segments).toEqual([])
    expect(g.yMax).toBe(220)
  })
})

describe('TraceCanvas', () => {
  it('nomme le canvas et annonce le silence des sondes', () => {
    const raw = renderToStaticMarkup(createElement(TraceCanvas, {
      points: [pt(0, 30), pt(1, null, 'charge')],
      totalSec: 3,
      deadlineMs: 220,
      responseMs: 28,
      loadResponseMs: 31,
    }))
    const html = raw
      .replace(/&#x([0-9a-f]+);/gi, (_, code) => String.fromCodePoint(parseInt(code, 16)))
      .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(Number(code)))
    expect(html).toContain('role="img"')
    expect(html).toContain('Trace des sondes')
    expect(html).toContain('temps de réponse 28 ms')
    expect(html).toContain('réponse sous charge 31 ms')
    expect(html).toContain('objectif 220 ms')
    expect(html).toContain('sonde sans réponse')
  })
})
