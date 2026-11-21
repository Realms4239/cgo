// R1 — le lexique doit aussi couvrir le canvas. Les noms de séries vivent dans
// l'option ECharts passée à setOption(), jamais dans renderToStaticMarkup :
// on capture donc la vraie option, en ne remplaçant que la dépendance canvas.
import { act, createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import TimeField from './TimeField'
import { TypedMetricRing, type InstrumentRings } from '../../lib/ringBuffer'

const h = vi.hoisted(() => ({
  chart: {
    setOption: vi.fn(),
    getWidth: () => 800,
    resize: vi.fn(),
    dispose: vi.fn(),
  },
}))

vi.mock('../../lib/echarts', () => ({ echarts: { init: () => h.chart } }))

let queue: FrameRequestCallback[]
let clock: number
let root: Root | null = null

function freshRings(): InstrumentRings {
  return {
    small: new TypedMetricRing(8), rtt50: new TypedMetricRing(8),
    rtt95: new TypedMetricRing(8), goodput: new TypedMetricRing(8),
    ifaceRx: new TypedMetricRing(8), ifaceTx: new TypedMetricRing(8),
  }
}

beforeEach(() => {
  ;(globalThis as any).IS_REACT_ACT_ENVIRONMENT = true
  queue = []
  clock = 1000
  h.chart.setOption.mockClear()
  vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => {
    queue.push(cb)
    return queue.length
  })
  vi.stubGlobal('cancelAnimationFrame', () => {})
  document.body.innerHTML = ''
})

afterEach(() => {
  if (root) { act(() => root!.unmount()); root = null }
  vi.unstubAllGlobals()
})

async function mount(rings: InstrumentRings) {
  const host = document.createElement('div')
  document.body.appendChild(host)
  root = createRoot(host)
  await act(async () => { root!.render(createElement(TimeField, { rings })) })
}

function frames(n = 3) {
  for (let i = 0; i < n; i++) {
    const pending = queue.splice(0)
    for (const cb of pending) cb((clock += 16))
  }
}

describe('TimeField — lexique du canvas', () => {
  it('aucun titre ni nom de série ECharts ne parle le lexique de laboratoire', async () => {
    const rings = freshRings()
    await mount(rings)
    // L'amorçage de grille au montage n'est pas l'option sous test.
    h.chart.setOption.mockClear()

    rings.small.push(1000, 42)
    await act(async () => { frames() })

    expect(h.chart.setOption, 'la vraie peinture a bien été capturée').toHaveBeenCalled()
    const option: any = h.chart.setOption.mock.calls[0][0]
    const labels: string[] = [
      ...(option.title?.text ? [String(option.title.text)] : []),
      ...(Array.isArray(option.series) ? option.series.map((s: any) => String(s.name ?? '')) : []),
    ]

    expect(labels.length, 'des libellés réels sont bien capturés').toBeGreaterThan(0)
    const offenders = labels.filter(l => /p95|qdisc|cake/i.test(l))
    expect(offenders, `libellés capturés : ${labels.join(' | ')}`).toEqual([])
    expect(labels, 'le libellé opérateur attendu est présent').toContain('Délai sous charge')
  })
})
