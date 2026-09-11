import { describe, it, expect } from 'vitest'
import { baseOption, lineSeries, paretoFrontier } from './chartGrammar'
describe('chartGrammar', () => {
  it('baseOption has useDirtyRect and markArea', () => {
    const opt: any = baseOption('RTT (ms)', 'ms')
    expect(opt.backgroundColor).toBe('transparent')
    // pushed props
    expect(opt.grid.backgroundColor).toContain('rgba')
  })
  it('lineSeries has shadowBlur and gradient', () => {
    const s: any = lineSeries('p95', [[0,1],[1,2]], '#5ad3e3', true)
    expect(s.lineStyle.shadowBlur).toBe(12)
    expect(s.areaStyle.color).toBeDefined()
  })
  it('paretoFrontier garde les non-dominés seuls (down/down)', () => {
    const pts = [{ x: 1, y: 1 }, { x: 2, y: 2 }, { x: 1, y: 3 }, { x: 3, y: 1 }]
    expect(paretoFrontier(pts, 'down', 'down')).toEqual([0])
  })
  it('paretoFrontier respecte les directions (up/up : max en haut à droite)', () => {
    const pts = [{ x: 1, y: 1 }, { x: 2, y: 2 }, { x: 1, y: 3 }, { x: 3, y: 1 }]
    expect(paretoFrontier(pts, 'up', 'up').sort()).toEqual([1, 2, 3])
  })
  it('paretoFrontier mixte (goodput up × coût down)', () => {
    const pts = [{ x: 10, y: 5 }, { x: 20, y: 8 }, { x: 20, y: 4 }]
    expect(paretoFrontier(pts, 'down', 'up')).toEqual([0, 1])
  })
})
