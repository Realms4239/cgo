// L'instrument se lit d'un coup d'œil : une trace primaire haute portant la
// vraie ligne 220 ms, deux panels causal partageant la MÊME étendue x, une
// règle unique, et un journal qui vit pendant la mesure.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField, { timeFieldOption } from './TimeField'
import JournalRail, { journalEntries } from './JournalRail'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

function filledRings(): InstrumentRings {
  const r: InstrumentRings = {
    small: new TypedMetricRing(20), rtt50: new TypedMetricRing(20),
    rtt95: new TypedMetricRing(20), goodput: new TypedMetricRing(20),
    ifaceRx: new TypedMetricRing(20), ifaceTx: new TypedMetricRing(20),
  }
  for (let i = 0; i < 6; i++) {
    const ts = 1000 + i * 100
    r.small.push(ts, i === 3 ? null : 30 + i)
    r.rtt50.push(ts, 20 + i)
    r.rtt95.push(ts, 25 + i)
    r.goodput.push(ts, 90 + i)
    r.ifaceRx.push(ts, 150 + i)
    r.ifaceTx.push(ts, 40 + i)
  }
  return r
}

function html(node: React.ReactElement): string {
  return renderToStaticMarkup(node as never)
    .replace(/&#x([0-9a-f]+);/gi, (_, c) => String.fromCodePoint(parseInt(c, 16)))
    .replace(/&#(\d+);/g, (_, c) => String.fromCodePoint(Number(c)))
}

describe('timeFieldOption — axe partagé et ligne d’objectif', () => {
  const opt: any = timeFieldOption(filledRings(), 220)
  const grids = opt.grid as any[]

  it('la trace primaire est la plus haute et porte la vraie ligne 220 ms', () => {
    expect(grids).toHaveLength(3)
    const px = (v: unknown) => (typeof v === 'number' ? v : parseFloat(String(v)))
    expect(px(grids[0].height)).toBeGreaterThan(px(grids[1].height))
    expect(px(grids[0].height)).toBeGreaterThan(px(grids[2].height))
    const primary: any = (opt.series as any[])[0]
    expect(primary.markLine.data).toEqual([{ yAxis: 220 }])
    expect(String(primary.markLine.label.formatter)).toContain('220')
  })

  it('goodput et interface partagent exactement la même étendue x', () => {
    const axes: any[] = opt.xAxis as any[]
    expect(axes).toHaveLength(3)
    for (const ax of axes) {
      expect(ax.type).toBe('time')
      expect(ax.min).toBe(1000)
      expect(ax.max).toBe(1500)
    }
    const yAxes: any[] = opt.yAxis as any[]
    expect(yAxes[1].max).toBeGreaterThan(1)
    expect(yAxes[2].max).toBeGreaterThan(1)
  })

  it('la règle est le seul axe étiqueté et le curseur est lié à tous les axes', () => {
    const axes: any[] = opt.xAxis as any[]
    expect(axes[0].axisLabel.show).toBe(false)
    expect(axes[1].axisLabel.show).toBe(false)
    expect(axes[2].axisLabel.show).not.toBe(false)
    expect((opt.axisPointer as any).link).toEqual([{ xAxisIndex: 'all' }])
  })

  it('le tracé casse les absences et reste animation:false', () => {
    expect(opt.animation).toBe(false)
    const primary: any = (opt.series as any[])[0]
    expect(primary.connectNulls).toBe(false)
    expect(primary.data[3]).toEqual([1300, null])
    expect(primary.data[0]).toEqual([1000, 30])
  })

  it('les séries causal sont indexées sur les axes partagés', () => {
    const series: any[] = opt.series as any[]
    expect(series.find((s: any) => s.name === 'Débit utile')?.xAxisIndex).toBe(1)
    expect(series.find((s: any) => s.name === 'Interface RX')?.xAxisIndex).toBe(2)
  })
})

describe('TimeField — lecture au repos', () => {
  it('affiche l’objectif et le mot d’absence sans tracer de faux zéro', () => {
    const out = html(createElement(TimeField, { rings: filledRings() }))
    expect(out).toContain('objectif 220 ms')
    expect(out).toContain('Axe temps partagé')
    expect(out).toContain('En direct')
    expect(out).toContain('points')
    expect(out).toContain('role="img"')
  })
})

describe('JournalRail — vivant pendant la mesure', () => {
  it('montre les événements même running=true, du plus récent au plus ancien', () => {
    const events = new DomainEventRing()
    events.push({ ts: 1000, kind: 'phase', label: 'Repos' })
    events.push({ ts: 2000, kind: 'phase', label: 'Charge' })
    const out = html(createElement(JournalRail, { events, running: true }))
    expect(out).toContain('Charge')
    expect(out).toContain('Repos')
    expect(out.indexOf('Charge')).toBeLessThan(out.indexOf('Repos'))
    expect(out).toContain('JOURNAL')
    expect(out).not.toContain('Journal figé')
  })

  it('journalEntries reflète un événement poussé après le rendu (pas figé)', () => {
    const events = new DomainEventRing()
    events.push({ ts: 1000, kind: 'phase', label: 'Repos' })
    expect(journalEntries(events).map(e => e.label)).toEqual(['Repos'])
    events.push({ ts: 2000, kind: 'gates', label: 'Portes fermées' })
    expect(journalEntries(events).map(e => e.label)).toEqual(['Portes fermées', 'Repos'])
  })

  it('un journal vide dit l’absence, jamais 0', () => {
    const out = html(createElement(JournalRail, { events: new DomainEventRing() }))
    expect(out).toContain('Aucun événement')
    expect(out).not.toMatch(/0 événement/)
  })
})

describe('TimeField — garde-fou de source', () => {
  it('ne repeint que sur anneau dirty et n’utilise pas setState par trame', () => {
    const s = readFileSync(join(process.cwd(), 'src/components/cockpit/TimeField.tsx'), 'utf8')
    expect(s).toContain('.dirty')
    expect(s).toContain('useRafLoop')
    expect(s).toMatch(/animation:\s*false|light:\s*true/)
  })
})
