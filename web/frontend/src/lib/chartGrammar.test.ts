import { describe, it, expect } from 'vitest'
import { baseOption, lineSeries, rawLineSeries, paretoFrontier, objectiveMarkLine, LIGHT } from './chartGrammar'

// Contraste WCAG réel (aucun nombre inventé) : les traits du graphique doivent
// tenir 3:1 sur le canevas clair ratifié.
function contrast(hex: string, bg: string): number {
  const lin = (c: number) => { const x = c / 255; return x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4 }
  const lum = (h: string) => {
    const n = parseInt(h.slice(1), 16)
    return 0.2126 * lin((n >> 16) & 255) + 0.7152 * lin((n >> 8) & 255) + 0.0722 * lin(n & 255)
  }
  const a = lum(hex), b = lum(bg)
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05)
}
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
  it('rawLineSeries ne lisse pas, n’échantillonne pas et ne comble pas les trous', () => {
    const s: any = rawLineSeries('RX', [[0, 1], [1, null]], '#5ad3e3')
    expect(s.smooth).toBe(false)
    expect(s.connectNulls).toBe(false)
    expect(s.sampling).toBeUndefined()
  })
})

describe('chartGrammar — instrument clair et ligne d’objectif réelle', () => {
  it('objectiveMarkLine(220) porte un vrai yAxis: 220 et une étiquette', () => {
    const ml: any = objectiveMarkLine(220)
    expect(ml.data).toEqual([{ yAxis: 220 }])
    expect(ml.label?.show).toBe(true)
    expect(String(ml.label?.formatter)).toContain('220')
    expect(ml.silent).toBe(true)
  })

  it('la palette claire est sémantique et tient 3:1 sur le canevas ratifié', () => {
    expect(LIGHT.canvas).toBe('#fbfbfb')
    for (const hue of [LIGHT.live, LIGHT.ok, LIGHT.warn, LIGHT.danger, LIGHT.neutral]) {
      expect(contrast(hue, LIGHT.canvas)).toBeGreaterThanOrEqual(3)
    }
  })

  it('baseOption en mode clair reste animation:false et utilise les jetons clairs', () => {
    const opt: any = baseOption('Délai sous charge', 'ms', { light: true })
    expect(opt.animation).toBe(false)
    expect(opt.textStyle.color).toBe(LIGHT.textMuted)
    expect(opt.title.textStyle.color).toBe(LIGHT.text)
    expect(opt.xAxis.axisLine.lineStyle.color).toBe(LIGHT.grid)
  })
})
