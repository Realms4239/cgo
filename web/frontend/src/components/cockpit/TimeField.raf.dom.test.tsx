// Régression revue T5 : la boucle de peinture doit démarrer à froid. Une valeur
// poussée dans l'anneau APRÈS le montage, sans aucun re-render parent, doit
// tout de même appeler setOption. On contrôle rAF et on remplace la seule
// dépendance canvas (happy-dom n'a pas de contexte 2D) ; le composant sous test
// n'est jamais mocké.
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

describe('TimeField — peinture à froid', () => {
  it('une écriture anneau sans re-render parent déclenche setOption', async () => {
    const rings = freshRings()
    await mount(rings)
    // L'amorçage de grille au montage n'est pas la peinture sous test.
    h.chart.setOption.mockClear()
    expect(queue.length).toBeGreaterThan(0)

    // Le flux SSE écrit l'anneau ; aucun parent ne re-render.
    rings.small.push(1000, 42)
    expect(rings.small.dirty).toBe(true)

    await act(async () => { frames() })

    expect(h.chart.setOption).toHaveBeenCalled()
    const option: any = h.chart.setOption.mock.calls[0][0]
    expect(option.series[0].data[0]).toEqual([1000, 42])
  })
})
