// Le champ temporel : une seule règle x pour la trace primaire, le débit utile
// et l'interface. La ligne 220 ms est un vrai yAxis étiqueté ; les absences
// cassent le tracé. La peinture vit dans une boucle rAF, jamais dans React.
// La règle et le pied portent les faits réels (événements, phases, interface) :
// absents, ils disent « — » + un mot français, jamais un zéro ni un vert.
import { useEffect, useRef, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { echarts } from '../../lib/echarts'
import { useInterval, useRafLoop } from '../../lib/hooks'
import { baseOption, objectiveMarkLine, rawLineSeries, LIGHT, CANVAS_FS } from '../../lib/chartGrammar'
import { ringTrace } from '../../lib/trace'
import { instrument, type DomainEvent, type DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'
import { live } from '../../lib/live'

export interface TimeFieldProps {
  rings: InstrumentRings
  events?: DomainEventRing
  iface?: IfaceHonesty
  phaseSince?: Record<string, number>
  jfi?: number | null
  objectiveMs?: number
}

/** Métadonnées d'interface réellement portées par le flux (lib/live.ts). */
export interface IfaceHonesty {
  name: string
  source: string
  valid: boolean
  gap: string
  windowMs: number
  rxDrops: number
  txDrops: number
  rxLast: number | null
  txLast: number | null
}

const ABSENT = '—'
const PHASE_FR: Record<string, string> = {
  baseline: 'Repos', charge: 'Charge', recup: 'Récupération',
  surveil: 'Surveillance', burst: 'Rafale',
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

function bounds(rings: InstrumentRings): { min: number; max: number } {
  const all = [rings.small, rings.goodput, rings.ifaceRx, rings.ifaceTx]
  const stamps = all.flatMap(r => Array.from(r.timestamps()))
  return {
    min: stamps.length ? Math.min(...stamps) : 0,
    max: stamps.length ? Math.max(...stamps) : 0,
  }
}

function clock(ts: number): string {
  const d = new Date(ts)
  if (!Number.isFinite(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}`
}

/** Dernière mesure valide d'une série sparse, sinon null (jamais 0). */
function lastValue(points: [number, number | null][]): number | null {
  for (let i = points.length - 1; i >= 0; i--) {
    const v = points[i][1]
    if (v != null && Number.isFinite(v)) return v
  }
  return null
}

function readIface(): IfaceHonesty {
  const f = live.iface
  return {
    name: f.name, source: f.source, valid: f.valid, gap: f.gap,
    windowMs: f.windowMs, rxDrops: f.rxDrops, txDrops: f.txDrops,
    rxLast: lastValue(f.rx as [number, number | null][]),
    txLast: lastValue(f.tx as [number, number | null][]),
  }
}

function equity(jfi: number | null | undefined) {
  if (jfi == null || !Number.isFinite(jfi) || jfi <= 0) {
    return { value: ABSENT, word: 'équité non mesurée', gauge: ABSENT }
  }
  const pct = Math.min(100, Math.max(0, jfi <= 1 ? jfi * 100 : jfi))
  const level = pct >= 80 ? 5 : pct >= 60 ? 4 : pct >= 40 ? 3 : pct >= 20 ? 2 : 1
  const word = pct >= 80 ? 'équité forte' : pct >= 40 ? 'équité moyenne' : 'équité faible'
  return { value: `${Math.round(pct)} %`, word, gauge: '▰'.repeat(level) + '▱'.repeat(5 - level) }
}

function ifaceField(label: string, value: string, kind: 'measure' | 'meta' = 'meta') {
  const missing = value === ABSENT
  return (
    <span className="tf-ifact">
      <span className="tf-ifact-k">{label}</span>
      <span className={missing ? 'tf-ifact-v tf-absent' : 'tf-ifact-v'}>
        {missing ? `${ABSENT} ${kind === 'measure' ? 'non mesuré' : 'non renseigné'}` : value}
      </span>
    </span>
  )
}

/** Option pure de l'instrument — testée sans DOM ni instance ECharts. */
export function timeFieldOption(rings: InstrumentRings, objectiveMs = 220): EChartsOption {
  const goal = Number.isFinite(objectiveMs) && objectiveMs > 0 ? objectiveMs : 220
  const { min, max } = bounds(rings)
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

export default function TimeField({
  rings, events = instrument.events, iface = readIface(), phaseSince = live.phaseSince,
  jfi = null, objectiveMs = 220,
}: TimeFieldProps) {
  const boxRef = useRef<HTMLDivElement>(null)
  const chart = useRef<echarts.ECharts | null>(null)
  const lastPaint = useRef(0)
  const [, setMetaTick] = useState(0)
  const lastMeta = useRef('')

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

  // La règle et le pied sont du texte DOM : ils suivent le flux à 2 Hz sans
  // entrer dans la boucle de peinture. On ne re-rend que sur changement réel.
  useInterval(() => {
    const evts = events.list()
    const sig = [
      evts.length, evts[evts.length - 1]?.ts ?? '', rings.small.size,
      Object.keys(phaseSince).join(','),
      iface.name, iface.rxLast, iface.txLast, iface.windowMs,
      iface.rxDrops, iface.txDrops, iface.source, iface.gap, iface.valid,
    ].join('|')
    if (sig !== lastMeta.current) { lastMeta.current = sig; setMetaTick(t => t + 1) }
  }, 500)

  const empty = rings.small.size === 0
  const { min, max } = bounds(rings)
  const span = max - min
  const inAxis = (ts: number) => Number.isFinite(ts) && span > 0 && ts >= min && ts <= max
  const pct = (ts: number) => (span > 0 ? ((ts - min) / span) * 100 : 0)

  const list: DomainEvent[] = events.list()
  const timedEvents = list.filter(e => inAxis(e.ts))
  const phaseEntries = Object.entries(phaseSince)
    .filter(([, ts]) => inAxis(ts))
    .sort((a, b) => a[1] - b[1])
  const bands = phaseEntries.map(([key, ts], i) => {
    const start = Math.max(ts, min)
    const end = i + 1 < phaseEntries.length ? Math.min(phaseEntries[i + 1][1], max) : max
    return { key, word: PHASE_FR[key] ?? key, left: pct(start), width: Math.max(0, pct(end) - pct(start)) }
  }).filter(b => b.width > 0)

  const eq = equity(jfi)
  const trace = rings.small.size > 0 ? 'flux SSE · trace vivante' : `${ABSENT} aucune trace`

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

      <div className="tf-ruler mono" role="group" aria-label="Axe temps partagé">
        <div className="tf-ruler-top">
          <span className="tf-ruler-time">Temps</span>
          <span className="tf-ruler-goal">objectif {objectiveMs} ms</span>
          <span className="tf-ruler-legend">
            <span className="tf-legend-title">Repères</span>
            <span><span className="tf-glyph" aria-hidden="true">▲</span> fait horodaté</span>
            <span><span className="tf-glyph tf-glyph-band" aria-hidden="true">▬</span> phase mesurée</span>
            <span><span className="tf-glyph" aria-hidden="true">{ABSENT}</span> absence</span>
          </span>
        </div>
        <div className="tf-ruler-track">
          {bands.length === 0 ? (
            <span className="tf-ruler-absent">{ABSENT} phase non mesurée</span>
          ) : bands.map(b => (
            <span key={b.key} className="tf-band" role="img" aria-label={`phase ${b.word}`}
              style={{ left: `${b.left}%`, width: `${b.width}%` }}>
              <span className="tf-band-word">{b.word}</span>
            </span>
          ))}
          {timedEvents.map((e, i) => (
            <span key={`${e.ts}-${e.kind}-${i}`} className="tf-tick"
              style={{ left: `${pct(e.ts)}%` }} aria-hidden="true" />
          ))}
        </div>
        {timedEvents.length === 0 ? (
          <p className="tf-repere-absent">{ABSENT} aucun repère — les faits enregistrés pendant la mesure s'afficheront ici.</p>
        ) : (
          <ol className="tf-reperes">
            {timedEvents.map((e, i) => (
              <li key={`${e.ts}-${e.kind}-${i}`} className="tf-repere">
                <span className="tf-repere-mark" aria-hidden="true">▲</span>
                <time dateTime={new Date(e.ts).toISOString()}>{clock(e.ts)}</time>
                <span className="tf-repere-kind">{e.kind}</span>
                <span className="tf-repere-label">{e.label}</span>
              </li>
            ))}
          </ol>
        )}
      </div>

      <footer className="tf-footer mono" aria-label="Métadonnées du champ temporel">
        <div className="tf-footer-group">
          <span className="tf-footer-title">Interface</span>
          <div className="tf-footer-grid">
            {ifaceField('nom', iface.name || ABSENT)}
            {ifaceField('RX', iface.valid && iface.rxLast != null ? `${iface.rxLast} Mbit/s` : ABSENT, 'measure')}
            {ifaceField('TX', iface.valid && iface.txLast != null ? `${iface.txLast} Mbit/s` : ABSENT, 'measure')}
            {ifaceField('fenêtre', iface.windowMs > 0 ? `${iface.windowMs} ms` : ABSENT)}
            {ifaceField('pertes RX', iface.valid ? String(iface.rxDrops) : ABSENT, 'measure')}
            {ifaceField('pertes TX', iface.valid ? String(iface.txDrops) : ABSENT, 'measure')}
            {ifaceField('source', iface.source || ABSENT)}
            {ifaceField('écart', iface.gap || ABSENT)}
          </div>
        </div>
        <div className="tf-footer-group">
          <span className="tf-footer-title">Équité</span>
          <div className="tf-footer-grid">
            <span className="tf-ifact">
              <span className="tf-ifact-k">indice</span>
              <span className={eq.value === ABSENT ? 'tf-ifact-v tf-absent' : 'tf-ifact-v'}>{eq.value} {eq.word}</span>
            </span>
            <span className="tf-ifact" aria-hidden="true">
              <span className="tf-ifact-k">jauge</span>
              <span className="tf-ifact-v">{eq.gauge}</span>
            </span>
          </div>
        </div>
        <div className="tf-footer-group">
          <span className="tf-footer-title">Provenance</span>
          <div className="tf-footer-grid">
            <span className="tf-ifact">
              <span className="tf-ifact-k">trace</span>
              <span className={rings.small.size > 0 ? 'tf-ifact-v' : 'tf-ifact-v tf-absent'}>{trace}</span>
            </span>
          </div>
        </div>
      </footer>
    </section>
  )
}
