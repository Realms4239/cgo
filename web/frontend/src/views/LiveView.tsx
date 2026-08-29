import { memo, useCallback, useEffect, useRef, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { echarts } from '../lib/echarts'
import { baseOption, lineSeries, barSeries, chargeMarkArea, CRAFT } from '../lib/chartGrammar'
import { live, clearLive } from '../lib/live'
import { computeQDI } from '../lib/qdi'
import { computeJFI } from '../lib/jfi'
import { useUIStore } from '../store/ui'
import { lttb } from '../lib/lttb'
import { useRafLoop } from '../lib/hooks'
import { animateBannerPulse, animateLiveEnter } from '../lib/anime'
import { MetricCard } from '../components/ui/MetricCard'
import { EmptyState } from '../components/ui/EmptyState'
import { Card } from '../components/ui/Card'
import { EmptyChart } from '../components/ui/EmptyChart'
import { PeekPopover } from '../components/PeekPopover'
import { Beam } from '../components/Beam'
import { DonutJFI } from '../components/DonutJFI'
import { Timeline } from '../components/Timeline'
import { CardHead } from '../components/ui/CardHead'
import { Pill } from '../components/ui/Pill'
import { loadSettings, latencyLevel, dropsLevel, deadlineLevel, goodputLevel, jfiLevel, LEVEL_COLOR } from '../lib/settings'

type Craft = 'line' | 'bar' | 'area'
type Tri = { metric: boolean; chart: Craft; source: 'live' | 'frozen' | 'both' }
const DEFAULT_TRI: Tri = { metric: true, chart: 'line', source: 'both' }

// craft adapter — one grammar, three chart crafts (Q4 chart pill: line|bar|area)
function craftSeries(craft: Craft, name: string, data: [number, number][], color: string) {
  if (craft === 'bar') return barSeries(name, data as any, color)
  return lineSeries(name, data, color, craft === 'area')
}

// useChart — callback ref: the ECharts instance follows whatever div React
// mounts (re-renders at 2 Hz may replace nodes; useEffect([]) orphaned them).
// useChart — the surface component owns the instance lifecycle (created after
// layout, disposed on ITS unmount). Parent re-renders at 2 Hz never touch it;
// React 19 ref-ordering can no longer orphan or dispose the chart.
function useChart(_title: string, _unit: string) {
  const chart = useRef<echarts.ECharts | null>(null)
  const onReady = useCallback((c: echarts.ECharts | null) => { chart.current = c }, [])
  const setData = (series: ReturnType<typeof lineSeries>[], extra?: Record<string, unknown>) => {
    if (!chart.current || chart.current.getWidth() < 10) return
    const empty = !series.some(s => ((s.data as unknown[]) ?? []).length > 1)
    const base = baseOption(_title, _unit, { idle: empty })
    const opt = { animation: false, ...base, ...(empty ? { dataZoom: [] } : {}), ...extra, series } as unknown as EChartsOption
    chart.current.setOption(opt)
  }
  return { onReady, setData, chart }
}

// ChartSurface — memoized owner: init after layout, dispose on unmount.
const ChartSurface = memo(function ChartSurface({ title, unit, domId, height, empty, hint, onReady }: {
  title: string
  unit: string
  domId?: string
  height: number | string
  empty: boolean
  hint: string
  onReady: (c: echarts.ECharts | null) => void
}) {
  const boxRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!boxRef.current) return
    const c = echarts.init(boxRef.current, undefined, { renderer: 'canvas', useDirtyRect: true, devicePixelRatio: Math.min(window.devicePixelRatio, 2) } as any)
    // prime the coordinate system BEFORE any data arrives — the first data
    // setOption on a never-painted chart crashed LineView ('coord') and
    // poisoned zrender's shared flush, freezing the other charts with it.
    // The primed base is also the visible idle grid (§3 honest emptiness).
    try { c.setOption(baseOption(title, unit, { idle: true })) } catch (e) { console.error('[chart] prime failed', e) }
    const ro = new ResizeObserver(() => { try { c.resize() } catch { } })
    ro.observe(boxRef.current)
    onReady(c)
    return () => { ro.disconnect(); try { c.dispose() } catch { }; onReady(null) }
  }, [])
  return (
    <div style={{ position: 'relative', height }}>
      <div ref={boxRef} id={domId} style={{ width: '100%', height: '100%' }} />
      {empty && <EmptyChart hint={hint} />}
    </div>
  )
})

type WallGroup = { qdisc: string; profile: string; small_p95_median: number; best?: boolean; hardware_recommendation?: string }

export default function LiveView() {
  const settings = loadSettings()
  const rtt = useChart('RTT', 'ms')
  const small = useChart('Petits objets p95', 'ms')
  const goodput = useChart('Bulk goodput', 'Mbit/s')
  const liveSnap = useUIStore((s: any) => s.live)
  const replayRunning = useUIStore((s: any) => s.replayRunning)
  const replayRunId = useUIStore((s: any) => s.replayRunId)
  const bannerRef = useRef<HTMLDivElement>(null)
  const [peek, setPeek] = useState<{ rect: DOMRect; value: number } | null>(null)
  const [wallGroups, setWallGroups] = useState<WallGroup[] | null>(null)
  const [wallHash, setWallHash] = useState<string>('')
  const [hovered, setHovered] = useState(false)
  const [tri, setTri] = useState<Tri>(DEFAULT_TRI)
  // edge control — baseline lock + shaping lever (ARG.md: control and shape)
  const [locked, setLocked] = useState<{ ms: number; at: string } | null>(() => {
    try { const r = localStorage.getItem('wall-baseline'); return r ? JSON.parse(r) : null } catch { return null }
  })
  const lockedRingRef = useRef<[number, number][]>([])
  const [shape, setShape] = useState<{ applied: boolean; qdisc?: string; capacity_mbps?: number } | null>(null)
  const [shapeCap, setShapeCap] = useState(20)
  const [shapeMsg, setShapeMsg] = useState('')
  const [journal, setJournal] = useState<{ ts: string; kind: string; msg: string }[]>([])
  const [linkOpen, setLinkOpen] = useState(false)
  const settingsRef = useRef(loadSettings())
  const [watching, setWatching] = useState(false)

  // Q4 tri-toggle — PanelChooser dispatches {metric, chart: craft, source}
  useEffect(() => {
    const onTri = (e: Event) => {
      const d = (e as CustomEvent).detail
      if (!d) return
      setTri({ metric: !!d.metric, chart: (d.chart as Craft) || 'line', source: (d.source as Tri['source']) || 'both' })
    }
    window.addEventListener('panel-chooser-tri', onTri)
    return () => window.removeEventListener('panel-chooser-tri', onTri)
  }, [])

  // honest idle: a new run starts from empty rings — no stale series under a fresh banner
  useEffect(() => { if (liveSnap?.running) clearLive() }, [liveSnap?.running])

  // frozen history for A/B baseline vs CAKE when idle
  useEffect(() => {
    if (liveSnap?.running || replayRunning) return
    let cancelled = false
    fetch('/api/results').then(r => r.json()).then(j => { if (!cancelled && j.available) setWallGroups(j.groups) }).catch(() => {})
    fetch('/api/integrity').then(r => r.json()).then(j => {
      if (cancelled) return
      // triple-provenance: hash8 = sha256(latest aqm_eval.csv)[:8]; run-id only as fallback
      const id = j?.hash8 ?? String(j?.run_ids?.[0] ?? '').slice(0, 8)
      if (id) setWallHash(String(id).slice(0, 8))
    }).catch(() => {})
    fetch('/api/shape').then(r => r.json()).then(j => { if (!cancelled) setShape(j) }).catch(() => {})
    fetch('/api/events').then(r => r.json()).then(j => { if (!cancelled && j?.events) setJournal(j.events.slice(-12).reverse()) }).catch(() => {})
    return () => { cancelled = true }
  }, [liveSnap?.running, replayRunning])

  const hash8 = wallHash || (() => {
    const src = `${liveSnap?.profile ?? ''}${liveSnap?.qdisc ?? ''}${liveSnap?.cc ?? ''}${wallGroups?.[0]?.profile ?? ''}`
    return src ? src.slice(0, 8).padEnd(8, '·').slice(0, 8) : '────────'
  })()

  const lastRef = useRef(0)
  useRafLoop((ts) => {
    if (ts - lastRef.current < 250) return
    lastRef.current = ts
    const d = (r: [number, number][]) => r.length > 400 ? lttb(r, 400) : r
    // CHARGE markArea — single per chart, phase-driven only. No quartile estimate.
    const charging = live.phase === 'charge'
    const cs = charging ? live.phaseSince['charge'] ?? live.rtt95[0]?.[0] ?? Date.now() - 1000 : 0
    const ce = charging ? live.rtt95.at(-1)?.[0] ?? Date.now() : 0
    const ma = chargeMarkArea(cs, ce, charging)
    rtt.setData([
      { ...craftSeries(tri.chart, 'p50', d(live.rtt50 as any), CRAFT.live), markArea: ma } as any,
      { ...craftSeries(tri.chart, 'p95', d(live.rtt95 as any), CRAFT.ok) } as any,
    ])
    const series: ReturnType<typeof lineSeries>[] = [{ ...craftSeries(tri.chart, 'small p95', d(live.small as any), CRAFT.threshold), markArea: ma } as any]
    if (locked && lockedRingRef.current.length > 1) {
      series.unshift({
        name: 'baseline verrouillée', type: 'line', showSymbol: false, smooth: 0.4, smoothMonotone: 'x', sampling: 'lttb' as const,
        lineStyle: { width: 1.5, type: 'dashed' as const, color: CRAFT.steel },
        emphasis: { focus: 'series' }, blur: { lineStyle: { opacity: 0.2 } },
        data: d(lockedRingRef.current),
      } as any)
    }
    small.setData(series)
    goodput.setData([{ ...craftSeries(tri.chart, 'goodput', d(live.goodput as any), CRAFT.bbr), markArea: ma } as any])
  })

  const banner = replayRunning ? `REPLAY — ${replayRunId}`
    : !liveSnap ? 'OFFLINE — en attente du flux'
    : liveSnap.load_status === 'bulk-on' ? 'CHARGE — bulk actif'
    : liveSnap.phase === 'surveil' ? 'SURVEILLANCE — sondes légères'
    : liveSnap.phase === 'baseline' ? 'BASELINE'
    : liveSnap.phase === 'recup' ? 'RÉCUPÉRATION'
    : 'IDLE'
  const bannerColor = replayRunning ? 'var(--t-live)'
    : !liveSnap ? 'var(--t-danger)'
    : banner.startsWith('CHARGE') ? 'var(--t-threshold)'
    : banner.startsWith('SURVEILLANCE') ? 'var(--t-live)'
    : banner === 'BASELINE' ? 'var(--t-live)'
    : banner === 'RÉCUPÉRATION' ? 'var(--t-ok)'
    : 'var(--text-faint)'

  const qdiVal = (() => {
    const p95 = live.rtt95.at(-1)?.[1] ?? liveSnap?.rtt_p95_ms ?? 0
    const p50 = live.rtt50.at(-1)?.[1] ?? liveSnap?.rtt_p50_ms ?? 0
    return computeQDI(p95, p50)
  })()
  const jfiVal: number | null = (() => {
    const gVals = live.goodput.slice(-20).map(([, v]) => v).filter(v => v > 0.01)
    if (gVals.length >= 2) {
      const allEqual = gVals.every(v => v === gVals[0])
      if (!allEqual) return computeJFI(gVals)
    }
    const sVals = live.small.slice(-12).map(([, v]) => v).filter(v => Number.isFinite(v) && v > 0)
    if (sVals.length >= 2) {
      const allEqual = sVals.every(v => v === sVals[0])
      if (!allEqual) return computeJFI(sVals)
    }
    return null
  })()

  const rttP95 = liveSnap?.rtt_p95_ms ?? live.rtt95.at(-1)?.[1] ?? 0
  const rttP50 = liveSnap?.rtt_p50_ms ?? live.rtt50.at(-1)?.[1] ?? 0
  const smallP95 = liveSnap?.small_p95_ms ?? live.small.at(-1)?.[1] ?? 0
  const goodputVal = liveSnap?.bulk_goodput_mbps ?? live.goodput.at(-1)?.[1] ?? 0
  const drops = liveSnap?.drops ?? 0
  const idle = !liveSnap?.running && live.rtt95.length === 0 && live.small.length === 0
  const wasted: number | null = idle ? null : liveSnap?.wasted_bytes ?? null
  const costAr: number | null = idle ? null : liveSnap?.cost_ar_per_h ?? null
  const deadlineOk: number | null = idle ? null : liveSnap?.deadline_ok_pct ?? null
  const spark = (r: [number, number][]) => r.map(([, v]) => v).slice(-20)
  const trendOf = (arr: number[]): 'up' | 'down' | 'flat' => {
    if (arr.length < 2) return 'flat'
    const a = arr[arr.length - 2], b = arr[arr.length - 1]
    if (b > a * 1.05) return 'up'
    if (b < a * 0.95) return 'down'
    return 'flat'
  }
  const qdiSpark = live.rtt95.slice(-20).map(([, v], i) => Math.max(0, v - (live.rtt50[i]?.[1] ?? v)))

  // live median of the current window — the number the comparison drives on
  const liveSmallMedian = (() => {
    const vals = live.small.slice(-120).map(([, v]) => v).filter(v => v > 0)
    if (vals.length < 5) return null
    return vals.reduce((a, b) => a + b, 0) / vals.length
  })()
  const liveDiff = locked && liveSmallMedian ? Math.round(((locked.ms - liveSmallMedian) / locked.ms) * 100) : null

  const lockBaseline = () => {
    if (liveSmallMedian == null) return
    const b = { ms: Math.round(liveSmallMedian * 10) / 10, at: new Date().toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' }) }
    lockedRingRef.current = [...live.small] // freeze the ring — the hero draws it grey dashed
    setLocked(b)
    try { localStorage.setItem('wall-baseline', JSON.stringify(b)) } catch { }
    useUIStore.getState().pushToast?.(`baseline verrouillée — ${b.ms} ms`, 'ok')
  }
  const applyShape = async (q: string) => {
    setShapeMsg('')
    try {
      const st = loadSettings()
      const r = await fetch('/api/shape', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ qdisc: q, capacity_mbps: shapeCap, delay_ms: st.linkDelayMs, jitter_ms: st.linkJitterMs, loss_pct: st.linkLossPct }) })
      const j = await r.json()
      if (!r.ok) { setShapeMsg(j?.error ?? `erreur ${r.status}`); useUIStore.getState().pushToast?.(j?.error ?? 'échec du façonnage', 'err') }
      else {
        setShape({ applied: q !== 'none', qdisc: q === 'none' ? undefined : q, capacity_mbps: shapeCap })
        setShapeMsg(q === 'none' ? 'façonnage retiré' : `${q} appliqué au bord @ ${shapeCap} Mbit/s`)
        useUIStore.getState().pushToast?.(q === 'none' ? 'façonnage retiré' : `bord façonné — ${q}`, 'ok')
      }
    } catch (e) { setShapeMsg(String(e)) }
  }
  const toggleWatch = async () => {
    try {
      const r = await fetch('/api/watch', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ on: !watching }) })
      const j = await r.json().catch(() => ({}))
      if (!r.ok) { useUIStore.getState().pushToast?.(j?.error ?? 'échec surveillance', 'err'); return }
      setWatching(!watching)
      if (!watching) { setLocked(null); lockedRingRef.current = [] }
      useUIStore.getState().pushToast?.(!watching ? 'surveillance active — le mur est vivant' : 'surveillance arrêtée', 'ok')
    } catch (e) { useUIStore.getState().pushToast?.(String(e), 'err') }
  }
  const exportConstat = () => {
    if (locked == null || liveSmallMedian == null) return
    const diff = Math.round(((locked.ms - liveSmallMedian) / locked.ms) * 100)
    const md = [
      '# Constat bufferbloat — Meteolink', '',
      `- baseline verrouillée : **${locked.ms} ms** (${locked.at})`,
      `- fenêtre courante : **${liveSmallMedian.toFixed(1)} ms**`,
      `- écart : **${diff > 0 ? '-' : ''}${diff} %**`,
      `- façonnage du bord : **${shape?.applied ? `${shape.qdisc} @ ${shape.capacity_mbps} Mbit/s` : 'aucun (pfifo)'}**`,
      `- empreinte campagne : ${hash8}`,
      `- généré : ${new Date().toLocaleString('fr-FR')}`, '',
      '_Kit de diagnostic portable — métrique small p95, fenêtre 180 s._',
    ].join('\n')
    const url = URL.createObjectURL(new Blob([md], { type: 'text/markdown' }))
    const a = document.createElement('a')
    a.href = url; a.download = `constat-bufferbloat-${Date.now()}.md`; a.click()
    URL.revokeObjectURL(url)
  }

  useEffect(() => { if (bannerRef.current) animateBannerPulse(bannerRef.current) }, [banner])
  useEffect(() => { animateLiveEnter() }, [])

  // A/B overlay — baseline pfifo_fast median vs CAKE best median same scale, diff badge, hash 8-char
  const baseline = wallGroups?.find(g => g.qdisc === 'pfifo_fast')
  const cakeBest = wallGroups?.find(g => g.qdisc === 'cake' && g.best) ?? wallGroups?.find(g => g.qdisc === 'cake') ?? wallGroups?.find(g => g.best) ?? null
  const maxAB = Math.max(baseline?.small_p95_median ?? 0, cakeBest?.small_p95_median ?? 0, smallP95, 1)
  const diffAB = baseline && cakeBest && baseline.small_p95_median > 0 ? Math.round(((baseline.small_p95_median - cakeBest.small_p95_median) / baseline.small_p95_median) * 100) : null
  const showLiveSrc = tri.source !== 'frozen'
  const showFrozenSrc = tri.source !== 'live'
  const heroEmpty = live.small.length === 0
  const rttEmpty = live.rtt95.length === 0
  const goodputEmpty = live.goodput.length === 0
  // Timeline 48 — phase bands from live.phaseSince, hidden idle (honest)
  const hasData = live.rtt95.length > 0
  const now = live.rtt95.at(-1)?.[0] ?? Date.now()
  const t0 = live.phaseSince['baseline'] ?? live.rtt95[0]?.[0] ?? now - 1000
  const tCharge = Math.max(t0 + 1, live.phaseSince['charge'] ?? (live.phase === 'charge' ? t0 + 1 : now))
  const tRecup = Math.max(tCharge + 1, live.phaseSince['recup'] ?? (live.phase === 'recup' ? now : tCharge + 1000))

  // instrumentation seam — surgical: read live option/pixels from probes
  if (typeof window !== 'undefined') (window as any).__CGO_CHARTS = { rtt: rtt.chart.current, small: small.chart.current, goodput: goodput.chart.current }
  return (
    <div id="wall" className="panel-stack" style={{ position: 'relative' }}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--text-muted)' }}>live p95 — point</div>
          <div className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 14, color: 'var(--t-live)' }}>{peek.value.toFixed(1)} ms</div>
          <div style={{ display: 'flex', alignItems: 'end', gap: 1, height: 20, marginTop: 4 }}>
            {live.rtt95.slice(-20).map(([, v]: [number, number], i: number) => <i key={i} style={{ flex: 1, height: `${Math.max(2, Math.min(20, (v / 300) * 20))}px`, background: CRAFT.live, borderRadius: 1, opacity: 0.6 + i * 0.02 }} />)}
          </div>
        </PeekPopover>
      )}
      <div ref={bannerRef} className="banner mono" style={{ color: bannerColor, borderColor: bannerColor + '55' }}>{banner}</div>
      <Card
        head="Petits objets p95"
        sub="p95 · fenêtre 180 s"
        style={{ height: 300, gridColumn: '1 / -1' }}
        onMouseEnter={e => { setHovered(true); const v = live.small.at(-1)?.[1] ?? smallP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); small.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.small.length - 1) }) }}
        onMouseLeave={() => { setHovered(false); setPeek(null) }}
      >
        <div style={{ flex: 1, minHeight: 0 }}>
          <ChartSurface title="Petits objets p95" unit="ms" domId="chart-small" height="100%" empty={heroEmpty} hint="en attente — démarrez une campagne" onReady={small.onReady} />
        </div>
      </Card>
      {!liveSnap && <div className="card" style={{ border: '1px dashed var(--hairline)', background: 'rgba(255,255,255,0.02)', textAlign: 'center' }}><EmptyState kind="empty" hint="en attente — Démarrer depuis Campagne pour alimenter le Live" /></div>}
      {/* Q4 metric pill — toggles the metric groups; one 5-col bento, 10 cells, no misaligned rows */}
      <div data-wall-cards="metric-groups" className="bento-5 wall-span" style={{ display: tri.metric ? 'grid' : 'none', gridTemplateColumns: 'repeat(5, minmax(0, 1fr))', gap: 'var(--gap, 24px)' }}>
        <MetricCard label="rtt_p95" value={rttP95 ? rttP95.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(rttP95, settings)]} spark={spark(live.rtt95)} trend={trendOf(spark(live.rtt95))} />
        <MetricCard label="rtt_p50" value={rttP50 ? rttP50.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(rttP50, settings)]} spark={spark(live.rtt50)} trend={trendOf(spark(live.rtt50))} />
        <MetricCard label="small_p95" value={smallP95 ? smallP95.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(smallP95, settings)]} spark={spark(live.small)} trend={trendOf(spark(live.small))} />
        <MetricCard label="bulk_goodput" value={goodputVal ? goodputVal.toFixed(1) : '—'} unit="Mbit/s" color={LEVEL_COLOR[goodputLevel(goodputVal, settings.shapeCap)]} spark={spark(live.goodput)} trend={trendOf(spark(live.goodput))} />
        <MetricCard label="drops" value={String(drops)} unit="" color={LEVEL_COLOR[dropsLevel(drops)]} trend={drops > 0 ? 'up' : 'flat'} />
        <MetricCard label="wasted" value={wasted == null ? '—' : wasted ? (wasted > 1024 * 1024 ? (wasted / 1024 / 1024).toFixed(1) + ' MiB' : String(wasted)) : '0'} unit="bytes" color={wasted == null ? 'var(--text-faint)' : CRAFT.threshold} trend={wasted != null && wasted > 0 ? 'up' : 'flat'} spark={spark(live.goodput)} />
        <MetricCard label="cost_ar_per_h" value={costAr == null ? '—' : costAr ? costAr.toFixed(0) : '0'} unit="Ar/h" color={costAr == null ? 'var(--text-faint)' : CRAFT.threshold} trend={costAr != null && costAr > 0 ? 'up' : 'flat'} spark={spark(live.goodput)} />
        <MetricCard label="deadline_ok" value={deadlineOk === null ? '—' : deadlineOk.toFixed(0)} unit={deadlineOk === null ? '' : '%'} color={deadlineOk === null ? 'var(--text-faint)' : LEVEL_COLOR[deadlineLevel(deadlineOk)]} trend={deadlineOk === null ? 'flat' : deadlineOk >= 95 ? 'down' : 'up'} spark={spark(live.small)} />
        <div data-testid="qdi-sparkline"><MetricCard label="QDI" value={!liveSnap || live.rtt95.length === 0 ? '—' : qdiVal.toFixed(1)} unit="ms" color={CRAFT.threshold} spark={live.rtt95.length === 0 ? undefined : qdiSpark} trend={trendOf(qdiSpark)} /></div>
        <Card head="JFI" sub="fairness 0–1" testid="jfi-card">
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: jfiVal === null ? 'var(--text-body)' : LEVEL_COLOR[jfiLevel(jfiVal)], fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{jfiVal === null ? '—' : jfiVal.toFixed(2)}</span>
            <DonutJFI value={jfiVal} />
          </div>
        </Card>
      </div>
      <div className="wall-span duo" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--gap, 24px)' }}>
        <Card
          head="RTT"
          sub="p50 · p95"
          onMouseEnter={e => { setHovered(true); const v = live.rtt95.at(-1)?.[1] ?? rttP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); rtt.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.rtt95.length - 1) }) }}
          onMouseLeave={() => { setHovered(false); setPeek(null) }}
        >
          <ChartSurface title="RTT" unit="ms" domId="chart-rtt" height={180} empty={rttEmpty} hint="rtt — en attente de flux" onReady={rtt.onReady} />
        </Card>
        <Card
          head="Bulk goodput"
          sub="mesuré au récepteur (compteur noyau)"
          onMouseEnter={e => { setHovered(true); const v = live.goodput.at(-1)?.[1] ?? goodputVal; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); goodput.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.goodput.length - 1) }) }}
          onMouseLeave={() => { setHovered(false); setPeek(null) }}
        >
          <ChartSurface title="Bulk goodput" unit="Mbit/s" domId="chart-goodput" height={180} empty={goodputEmpty} hint="goodput — en attente de flux" onReady={goodput.onReady} />
        </Card>
      </div>
      {hasData && <Timeline baselineStart={t0} chargeStart={tCharge} chargeEnd={tRecup} recupEnd={Math.max(tRecup, now)} currentPhase={live.phase || 'idle'} />}

      {/* live-wall-overlay: baseline grey dashed vs CAKE cyan solid same scale; source pill gates live|frozen|both */}
      <div className="live-wall-overlay card" data-testid="live-wall-overlay" style={{ gridColumn: '1 / -1', border: '1px solid var(--hairline)', background: 'var(--surface-card)', padding: 16, display: 'flex', flexDirection: 'column', gap: 10, width: '100%' }}>
        <CardHead
          label="Comparer — pfifo vs CAKE"
          sub={`source ${tri.source} · hash ${hash8} · ${wallGroups ? `${wallGroups.length} groupes gelés` : 'en attente'}`}
          right={diffAB != null ? (
            <span className="diff-badge mono" style={{ background: diffAB > 0 ? 'rgba(31,163,72,0.12)' : 'rgba(226,39,24,0.12)', border: '1px solid ' + (diffAB > 0 ? CRAFT.ok : CRAFT.danger), color: diffAB > 0 ? CRAFT.ok : CRAFT.danger, padding: '4px 10px', fontSize: 14, fontWeight: 700, fontVariantNumeric: 'tabular-nums' }}>{diffAB > 0 ? `-${diffAB}%` : `${diffAB}%`}</span>
          ) : undefined}
        />
        {/* edge control — verrouiller la baseline, façonner le bord, constat exportable */}
        <div style={{ border: '1px solid var(--hairline)', background: 'rgba(90,211,227,0.03)', padding: 10, display: 'flex', flexDirection: 'column', gap: 8 }}>
          <CardHead
            label="Contrôle du bord — façonner et mesurer"
            sub={shape?.applied ? `bord façonné : ${shape.qdisc} @ ${shape.capacity_mbps} Mbit/s` : 'bord non façonné (file simple)'}
            right={
              <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
                <Pill label="surveillance" value={watching ? 'on' : 'off'} on={watching} onClick={toggleWatch} title="sondes légères en continu (sans bulk) — rend l'effet du façonnage visible" />
                <Pill label="capacité" value={`${shapeCap} Mbit/s`} on onClick={() => setShapeCap(shapeCap >= 80 ? 5 : shapeCap * 2)} title="capacité du bord — clic pour doubler (boucle 5→80)" />
                <Pill label="conditions" value={linkOpen ? 'masquer' : `${settingsRef.current.linkDelayMs} ms`} on={linkOpen} onClick={() => setLinkOpen(!linkOpen)} title="conditions du lien — délai/gigue/perte appliqués au bord" />
                {['none', 'fq_codel', 'cake'].map(q => (
                  <Pill key={q} label={q === 'none' ? 'sans' : q} on={shape?.applied && shape.qdisc === q} onClick={() => applyShape(q)} title={`appliquer ${q} au bord`} />
                ))}
              </div>
            }
          />
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
            <button className="btn" onClick={lockBaseline} disabled={liveSmallMedian == null} style={{ padding: '6px 10px', fontSize: 10 }}>FIGER L'AVANT{liveSmallMedian != null ? ` — ${liveSmallMedian.toFixed(1)} ms` : ''}</button>
            {locked && <Pill label="avant" value={`${locked.ms} ms @ ${locked.at}`} on title="avant figé — cliquer pour libérer" onClick={() => { setLocked(null); try { localStorage.removeItem('wall-baseline') } catch { } }} />}
            <button className="btn btn-primary" onClick={exportConstat} disabled={locked == null || liveSmallMedian == null} style={{ padding: '6px 10px', fontSize: 10, marginLeft: 'auto' }}>TÉLÉCHARGER LE CONSTAT</button>
          </div>
          {shapeMsg && <div className="mono" style={{ fontSize: 10, color: shapeMsg.includes('erreur') || shapeMsg.includes('échec') ? CRAFT.danger : CRAFT.ok }}>{shapeMsg}</div>}
          {locked && liveSmallMedian != null && (
            <div>
              <div className="mono" style={{ fontSize: 10, color: 'var(--text-muted)', display: 'flex', justifyContent: 'space-between' }}>
                <span>avant → maintenant, même fenêtre</span>
                <span style={{ fontVariantNumeric: 'tabular-nums', color: liveDiff != null && liveDiff > 0 ? CRAFT.ok : CRAFT.danger }}>{liveDiff != null ? `${liveDiff > 0 ? '-' : '+'}${Math.abs(liveDiff)} %` : '—'}</span>
              </div>
              <div style={{ height: 10, background: 'rgba(118,123,132,0.08)', border: '1px dashed #767b84', borderRadius: 2, overflow: 'hidden', marginTop: 4, position: 'relative' }}>
                <div style={{ width: `${Math.min(100, (locked.ms / Math.max(locked.ms, liveSmallMedian)) * 100)}%`, height: '100%', background: '#767b84', opacity: 0.9, transition: 'width 0.4s ease' }} />
              </div>
              <div style={{ height: 10, background: 'rgba(90,211,227,0.08)', border: '1px solid ' + CRAFT.live, borderRadius: 2, overflow: 'hidden', marginTop: 4 }}>
                <div style={{ width: `${Math.min(100, (liveSmallMedian / Math.max(locked.ms, liveSmallMedian)) * 100)}%`, height: '100%', background: CRAFT.live, boxShadow: '0 0 6px rgba(90,211,227,0.5)', transition: 'width 0.4s ease' }} />
              </div>
            </div>
          )}
        </div>
        {showLiveSrc && (
          <div>
            <div className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: CRAFT.live, display: 'flex', justifyContent: 'space-between' }}>
              <span>live — cyan solide</span>
              <span style={{ fontVariantNumeric: 'tabular-nums', color: 'var(--text-body)' }}>{smallP95 ? `${smallP95.toFixed(1)} ms` : '—'}</span>
            </div>
            <div style={{ height: 10, background: 'rgba(90,211,227,0.08)', border: '1px solid ' + CRAFT.live, borderRadius: 2, overflow: 'hidden', marginTop: 4 }}>
              <div style={{ width: `${Math.min(100, (smallP95 / maxAB) * 100)}%`, height: '100%', background: CRAFT.live, boxShadow: '0 0 6px rgba(90,211,227,0.5)', transition: 'width 0.4s ease' }} />
            </div>
          </div>
        )}
        {showFrozenSrc && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <div>
              <div className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--text-faint)', display: 'flex', justifyContent: 'space-between' }}>
                <span>baseline pfifo_fast — gris pointillé</span>
                <span style={{ fontVariantNumeric: 'tabular-nums', color: 'var(--text-body)' }}>{baseline ? `${baseline.small_p95_median.toFixed(1)} ms` : '—'}</span>
              </div>
              <div style={{ height: 10, background: 'rgba(118,123,132,0.08)', border: '1px dashed #767b84', borderRadius: 2, overflow: 'hidden', marginTop: 4 }}>
                <div style={{ width: `${baseline ? (baseline.small_p95_median / maxAB) * 100 : 0}%`, height: '100%', background: '#767b84', opacity: 0.9, transition: 'width 0.4s ease' }} />
              </div>
            </div>
            <div>
              <div className="mono" style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: CRAFT.live, display: 'flex', justifyContent: 'space-between' }}>
                <span>CAKE — cyan solide{cakeBest?.best ? ' ★' : ''}</span>
                <span style={{ fontVariantNumeric: 'tabular-nums', color: 'var(--text-body)' }}>{cakeBest ? `${cakeBest.small_p95_median.toFixed(1)} ms` : '—'}</span>
              </div>
              <div style={{ height: 10, background: 'rgba(90,211,227,0.08)', border: '1px solid ' + CRAFT.live, borderRadius: 2, overflow: 'hidden', marginTop: 4 }}>
                <div style={{ width: `${cakeBest ? (cakeBest.small_p95_median / maxAB) * 100 : 0}%`, height: '100%', background: CRAFT.live, boxShadow: '0 0 6px rgba(90,211,227,0.5)', transition: 'width 0.4s ease' }} />
              </div>
            </div>
          </div>
        )}
      </div>

      {journal.length > 0 && (
        <div className="card" data-testid="journal" style={{ padding: '8px 12px', display: 'flex', flexDirection: 'column', gap: 4 }}>
          <CardHead label="Journal" sub="actions et événements — 50 derniers" />
          {journal.slice(0, 6).map((e, i) => (
            <div key={i} className="mono" style={{ fontSize: 10, color: 'var(--text-muted)', display: 'flex', gap: 8 }}>
              <span style={{ color: 'var(--text-faint)' }}>{e.ts.slice(11, 19)}</span>
              <span style={{ color: 'var(--t-live)', minWidth: 84 }}>{e.kind}</span>
              <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{e.msg}</span>
            </div>
          ))}
        </div>
      )}
      {liveSnap?.running && <Beam hovered={hovered} />}
    </div>
  )
}
