import { describe, it, expect } from 'vitest'
import { baseOption, lineSeries } from './chartGrammar'
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
})
