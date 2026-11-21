// Le champ temporel : une seule règle x pour la trace primaire, le débit utile
// et l'interface. La ligne 220 ms est un vrai yAxis étiqueté ; les absences
// cassent le tracé. La peinture vit dans une boucle rAF, jamais dans React.
import { useEffect, useRef } from 'react'
import type { EChartsOption } from 'echarts'
import { echarts } from '../../lib/echarts'
import { useRafLoop } from '../../lib/hooks'
import { baseOption, objectiveMarkLine, rawLineSeries, LIGHT, CANVAS_FS } from '../../lib/chartGrammar'
import { ringTrace } from '../../lib/trace'
import type { InstrumentRings } from '../../lib/ringBuffer'

export interface TimeFieldProps {
  rings: InstrumentRings
  objectiveMs?: number
}

const TIME_AXIS = {
  type: 'time' as const,
  axisLine: { lineStyle: { color: LIGHT.grid } },
  axisTick: { show: false },
  splitLine: { show: false },
}

function yAxis(name: string, max: number, gridIndex: number) {
  return {
    type: 'value' as const, gridIndex, name,
    min: 0, max,
    nameTextStyle: { color: LIGHT.textMuted, fontFamily: 'JetBrains Mono', fontSize: CANVAS_FS.axis },
    axisLine: { show: false }, axisTick: { show: false },
    axisLabel: { show: true, color: LIGHT.textMuted, fontFamily: 'JetBrains Mono', fontSize: CANVAS_FS.axis },
    splitLine: { lineStyle: { color: LIGHT.grid, type: 'dashed' as const } },
  }
}

/** Option pure de l'instrument — testée sans DOM ni instance ECharts. */
export function timeFieldOption(rings: InstrumentRings, objectiveMs = 220): EChartsOption {
  const goal = Number.isFinite(objectiveMs) && objectiveMs > 0 ? objectiveMs : 220
  const all = [rings.small, rings.goodput, rings.ifaceRx, rings.ifaceTx]
  const stamps = all.flatMap(r => Array.from(r.timestamps()))
  const min = stamps.length ? Math.min(...stamps) : 0
  const max = stamps.length ? Math.max(...stamps) : 0
  const axisMax = (ring: InstrumentRings['goodput']) => {
    let m = 1
    for (const v of ring.series()) if (Number.isFinite(v) && v > m) m = v
    return m * 1.1
  }
  const primary = ringTrace(rings.small, goal)

  return {
    backgroundColor: 'transparent',
    animation: false,
    animationDuration: 0,
    animationDurationUpdate: 0,
    textStyle: { color: LIGHT.textMuted, fontFamily: 'JetBrains Mono', fontSize: CANVAS_FS.axis },
    tooltip: {
      trigger: 'axis', backgroundColor: LIGHT.surface, borderColor: LIGHT.grid,
      textStyle: { color: LIGHT.text, fontFamily: 'JetBrains Mono', fontSize: CANVAS_FS.tooltip },
    } as EChartsOption['tooltip'],
    axisPointer: { link: [{ xAxisIndex: 'all' }] },
    grid: [
      { left: 64, right: 16, top: 6, height: '54%', containLabel: true },
      { left: 64, right: 16, top: '63%', height: '16%', containLabel: true },
      { left: 64, right: 16, top: '82%', height: '16%', containLabel: true },
    ],
    xAxis: [
      { ...TIME_AXIS, gridIndex: 0, min, max, axisLabel: { show: false } },
      { ...TIME_AXIS, gridIndex: 1, min, max, axisLabel: { show: false } },
      {
        ...TIME_AXIS, gridIndex: 2, min, max,
        axisLine: { lineStyle: { color: LIGHT.neutral } },
        axisLabel: {
          show: true, color: LIGHT.textMuted, fontFamily: 'JetBrains Mono', fontSize: CANVAS_FS.axis,
          hideOverlap: true,
          formatter: (v: number) => {
            const d = new Date(v)
            return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
          },
        },
      },
    ],
    yAxis: [
      yAxis('ms', Math.max(primary.yMax * 1.1, goal * 1.1), 0),
      yAxis('Mbit/s', axisMax(rings.goodput), 1),
      yAxis('Mbit/s', axisMax(rings.ifaceRx), 2),
    ],
    series: [
      {
        ...rawLineSeries('Délai sous charge', primary.data, LIGHT.live),
        xAxisIndex: 0, yAxisIndex: 0, markLine: objectiveMarkLine(goal),
      },
      { ...rawLineSeries('Aller-retour au repos', rings.rtt50.pairs(), LIGHT.neutral), xAxisIndex: 0, yAxisIndex: 0 },
      { ...rawLineSeries('Aller-retour sous charge', rings.rtt95.pairs(), LIGHT.ok), xAxisIndex: 0, yAxisIndex: 0 },
      { ...rawLineSeries('Débit utile', rings.goodput.pairs(), LIGHT.ok), xAxisIndex: 1, yAxisIndex: 1 },
      { ...rawLineSeries('Interface RX', rings.ifaceRx.pairs(), LIGHT.live), xAxisIndex: 2, yAxisIndex: 2 },
      { ...rawLineSeries('Interface TX', rings.ifaceTx.pairs(), LIGHT.neutral), xAxisIndex: 2, yAxisIndex: 2 },
    ],
  } as unknown as EChartsOption
}

export default function TimeField({ rings, objectiveMs = 220 }: TimeFieldProps) {
  const boxRef = useRef<HTMLDivElement>(null)
  const chart = useRef<echarts.ECharts | null>(null)
  const lastPaint = useRef(0)

  useEffect(() => {
    if (!boxRef.current) return
    const c = echarts.init(boxRef.current, undefined, {
      renderer: 'canvas', useDirtyRect: true, devicePixelRatio: Math.min(window.devicePixelRatio, 2),
    } as any)
    // amorcer la grille avant toute donnée : une première setOption sur un
    // graphe jamais peint casse LineView ('coord') et gèle les autres graphes.
    try { c.setOption({ ...baseOption('', '', { light: true, idle: true }), series: [] } as EChartsOption) } catch { }
    const ro = new ResizeObserver(() => { try { c.resize() } catch { } })
    ro.observe(boxRef.current)
    chart.current = c
    return () => { ro.disconnect(); try { c.dispose() } catch { }; chart.current = null }
  }, [])

  // Toujours actif : `active` dérivait d'un état MUTABLE (les anneaux) que le
  // flux SSE écrit sans re-render — la boucle ne démarrait jamais à froid.
  // Le callback porte déjà les deux portes utiles : dirty et throttle 4 Hz.
  useRafLoop((ts) => {
    if (!chart.current || chart.current.getWidth() < 10) return
    const dirty = rings.small.dirty || rings.goodput.dirty || rings.ifaceRx.dirty || rings.ifaceTx.dirty
    if (!dirty || ts - lastPaint.current < 250) return
    lastPaint.current = ts
    // une seule lecture par cycle : le peintre rend l'anneau propre.
    rings.small.dirty = false
    rings.goodput.dirty = false
    rings.ifaceRx.dirty = false
    rings.ifaceTx.dirty = false
    chart.current.setOption(timeFieldOption(rings, objectiveMs), { notMerge: true })
  }, true)

  const empty = rings.small.size === 0
  return (
    <section className="time-field" aria-label="Champ temporel">
      <div className="tf-header mono" role="status">
        <span className="tf-status">{empty ? 'Aucun signal' : 'En direct'}</span>
        <span className="tf-fill">{rings.small.size} / {rings.small.capacity} points</span>
        <span className="tf-source">direct</span>
      </div>
      <div className="tf-plot">
        <div
          ref={boxRef} className="tf-chart" role="img"
          aria-label="Trace primaire, débit utile et interface — axe temporel partagé"
        />
        {empty && <p className="tf-empty">La trace apparaîtra ici dès la première mesure.</p>}
      </div>
      <div className="tf-ruler mono" role="img" aria-label="Axe temps partagé">
        <span>Temps</span>
        <span>objectif {objectiveMs} ms</span>
      </div>
    </section>
  )
}
