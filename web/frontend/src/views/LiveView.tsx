import { useEffect, useRef, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { echarts } from '../lib/echarts'
import { baseOption, lineSeries } from '../lib/chartGrammar'
import { live } from '../lib/live'
import { computeQDI } from '../lib/qdi'
import { computeJFI } from '../lib/jfi'
import { useUIStore } from '../store/ui'
import { lttb } from '../lib/lttb'
import { useRafLoop } from '../lib/hooks'
import { animateBannerPulse, animateLiveEnter } from '../lib/anime'
import { MetricCard } from '../components/ui/MetricCard'
import { EmptyState } from '../components/ui/EmptyState'
import { PeekPopover } from '../components/PeekPopover'
import { Beam } from '../components/Beam'
import { DonutJFI } from '../components/DonutJFI'

function chargeMarkArea(s: number, e: number) {
  return {
    itemStyle: { color: 'rgba(244,180,0,0.04)', borderColor: 'rgba(244,180,0,0.12)', borderWidth: 1, borderType: 'dashed' as const },
    label: { show: true, color: '#f4b400', fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop' as const, padding: [4, 8] as unknown as number[], backgroundColor: 'rgba(244,180,0,0.08)', formatter: 'CHARGE' },
    data: [[{ xAxis: s }, { xAxis: e }]] as any,
  }
}

function useChart(title: string, unit: string) {
  const ref = useRef<HTMLDivElement>(null)
  const chart = useRef<echarts.ECharts | null>(null)
  useEffect(() => {
    if (!ref.current) return
    const c = echarts.init(ref.current, undefined, { renderer: 'canvas', useDirtyRect: true, devicePixelRatio: Math.min(window.devicePixelRatio, 2) } as any)
    chart.current = c
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(ref.current)
    return () => { ro.disconnect(); c.dispose() }
  }, [])
  const setData = (series: ReturnType<typeof lineSeries>[], extra?: Record<string, unknown>) => {
    if (!chart.current) return
    const brush = { toolbox: ['rect'], brushType: 'rect' as const, xAxisIndex: 'all' as const, brushMode: 'single' as const }
    const empty = !series.some(s => ((s.data as unknown[]) ?? []).length > 1)
    const base = baseOption(title, unit)
    const opt = { animation: false, ...base, ...(empty ? { dataZoom: [] } : {}), brush, ...extra, series } as unknown as EChartsOption
    chart.current.setOption(opt)
  }
  return { ref, setData, chart }
}

type WallGroup = { qdisc: string; profile: string; small_p95_median: number; best?: boolean; hardware_recommendation?: string }

export default function LiveView() {
  const rtt = useChart('RTT (ms)', 'ms')
  const small = useChart('Petits objets p95 (ms)', 'ms')
  const goodput = useChart('Bulk goodput (Mbit/s)', 'Mbit/s')
  const liveSnap = useUIStore((s: any) => s.live)
  const replayRunning = useUIStore((s: any) => s.replayRunning)
  const replayRunId = useUIStore((s: any) => s.replayRunId)
  const bannerRef = useRef<HTMLDivElement>(null)
  const [peek, setPeek] = useState<{ rect: DOMRect; value: number } | null>(null)
  const [wallGroups, setWallGroups] = useState<WallGroup[] | null>(null)
  const [wallHash, setWallHash] = useState<string>('')

  // exhaustive: wallGroups fetch from /api/results when !live.running and !replayRunning — frozen history for A/B baseline vs CAKE
  useEffect(() => {
    if (liveSnap?.running || replayRunning) return
    let cancelled = false
    fetch('/api/results').then(r => r.json()).then(j => { if (!cancelled && j.available) setWallGroups(j.groups) }).catch(() => {})
    fetch('/api/integrity').then(r => r.json()).then(j => {
      if (cancelled) return
      const id = j?.run_ids?.[0] ?? j?.runs?.[0] ?? ''
      if (id) setWallHash(String(id).slice(0, 8))
      else if (j?.available && j?.run_ids?.length) setWallHash(String(j.run_ids[0]).slice(0, 8))
    }).catch(() => {})
    return () => { cancelled = true }
  }, [liveSnap?.running, replayRunning])

  // also prime hash from live meta when idle (fallback if integrity empty)
  const hash8 = wallHash || (() => {
    const src = `${liveSnap?.profile ?? ''}${liveSnap?.qdisc ?? ''}${liveSnap?.cc ?? ''}${wallGroups?.[0]?.profile ?? ''}`
    return src ? src.slice(0, 8).padEnd(8, '·').slice(0, 8) : '────────'
  })()

  const lastRef = useRef(0)
  useRafLoop((ts) => {
    if (ts - lastRef.current < 250) return
    lastRef.current = ts
    const d = (r: [number, number][]) => r.length > 400 ? lttb(r, 400) : r
    const cs = live.rtt95.length ? live.rtt95[Math.floor(live.rtt95.length * 0.25)]?.[0] ?? Date.now() - 45000 : Date.now() - 45000
    const ce = live.rtt95.length ? live.rtt95[Math.floor(live.rtt95.length * 0.75)]?.[0] ?? Date.now() - 10000 : Date.now() - 10000
    const ma = chargeMarkArea(cs, ce)
    rtt.setData([
      { ...lineSeries('p50', d(live.rtt50 as any), '#5ad3e3'), markArea: ma } as any,
      { ...lineSeries('p95', d(live.rtt95 as any), '#1fa348'), markArea: ma } as any,
    ])
    small.setData(
      [{ ...lineSeries('small p95', d(live.small as any), '#f4b400'), markArea: ma } as any],
      { visualMap: { show: false, type: 'piecewise' as const, dimension: 1, pieces: [{ gt: 100, color: '#e22718' }, { gt: 40, color: '#f4b400' }, { lte: 40, color: '#5ad3e3' }], outOfRange: { color: '#9aa3ad' } } } as any,
    )
    goodput.setData([{ ...lineSeries('goodput', d(live.goodput as any), '#b48ae0', true), markArea: ma } as any])
  })

  const banner = replayRunning ? `REPLAY — ${replayRunId}`
    : !liveSnap ? 'OFFLINE — en attente du flux'
    : liveSnap.load_status === 'bulk-on' ? 'CHARGE — bulk actif'
    : liveSnap.phase === 'baseline' ? 'BASELINE'
    : liveSnap.phase === 'recup' ? 'RÉCUPÉRATION'
    : 'IDLE'
  const bannerColor = replayRunning ? 'var(--t-live)'
    : !liveSnap ? 'var(--t-danger)'
    : banner.startsWith('CHARGE') ? 'var(--t-threshold)'
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

  useEffect(() => { if (bannerRef.current) animateBannerPulse(bannerRef.current) }, [banner])
  useEffect(() => { animateLiveEnter() }, [])

  // exhaustive A/B overlay — baseline pfifo_fast median vs CAKE best median same scale, diff badge, hash 8-char
  const baseline = wallGroups?.find(g => g.qdisc === 'pfifo_fast')
  const cakeBest = wallGroups?.find(g => g.qdisc === 'cake' && g.best) ?? wallGroups?.find(g => g.qdisc === 'cake') ?? wallGroups?.find(g => g.best) ?? null
  const maxAB = Math.max(baseline?.small_p95_median ?? 0, cakeBest?.small_p95_median ?? 0, 1)
  const diffAB = baseline && cakeBest && baseline.small_p95_median > 0 ? Math.round(((baseline.small_p95_median - cakeBest.small_p95_median) / baseline.small_p95_median) * 100) : null

  return (
    <div id="wall" className="panel-stack" style={{ position: 'relative' }}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099' }}>live p95 — point</div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 14, color: '#5ad3e3' }}>{peek.value.toFixed(1)} ms</div>
          <div style={{ display: 'flex', alignItems: 'end', gap: 1, height: 20, marginTop: 4 }}>
            {live.rtt95.slice(-20).map(([, v]: [number, number], i: number) => <i key={i} style={{ flex: 1, height: `${Math.max(2, Math.min(20, (v / 300) * 20))}px`, background: '#5ad3e3', borderRadius: 1, opacity: 0.6 + i * 0.02 }} />)}
          </div>
        </PeekPopover>
      )}
      <div ref={bannerRef} className="banner mono" style={{ color: bannerColor, borderColor: bannerColor + '55' }}>{banner} · {replayRunning ? 'replay' : 'SSE 10 Hz'}</div>
      {/* HERO: small_p95 300px full-width — thesis: wall hero, not bottom */}
      <div className="card" style={{ height: '300px', gridColumn: '1/-1', width: '100%', display: 'flex', flexDirection: 'column' }} onMouseEnter={e => { const v = live.small.at(-1)?.[1] ?? smallP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); small.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.small.length - 1) }) }} onMouseLeave={() => setPeek(null)}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8, flexWrap: 'wrap', gap: 8, minWidth: 0 }}>
          <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099', whiteSpace: 'nowrap' }}>Petits objets p95 — hero 300px</span>
          <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#767b84', display: 'inline-flex', alignItems: 'center', gap: 6, whiteSpace: 'nowrap', fontVariantNumeric: 'tabular-nums' }}>
            Last updated {new Date().toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}
            <span aria-hidden="true" style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--t-live, #5ad3e3)', boxShadow: '0 0 6px rgba(90,211,227,0.6)', display: 'inline-block', opacity: liveSnap?.running ? 1 : 0.35 }} />
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: liveSnap?.running ? '#5ad3e3' : '#767b84', display: 'inline-block' }} />
            live
          </span>
        </div>
        <div ref={small.ref} style={{ flex: 1, minHeight: 0 }} />
      </div>
      {!liveSnap && live.rtt95.length === 0 && <div className="card" style={{ border: '1px dashed var(--hairline)', background: 'rgba(255,255,255,0.02)', textAlign: 'center' }}><EmptyState kind="empty" hint="en attente — Démarrer depuis Campagne pour alimenter le Live" /></div>}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, minmax(0, 1fr))', gap: 12 }}>
        <MetricCard label="rtt_p95" value={rttP95 ? rttP95.toFixed(1) : '—'} unit="ms" color="#5ad3e3" spark={spark(live.rtt95)} trend={trendOf(spark(live.rtt95))} />
        <MetricCard label="rtt_p50" value={rttP50 ? rttP50.toFixed(1) : '—'} unit="ms" color="#5ad3e3" spark={spark(live.rtt50)} trend={trendOf(spark(live.rtt50))} />
        <MetricCard label="small_p95" value={smallP95 ? smallP95.toFixed(1) : '—'} unit="ms" color="#1fa348" spark={spark(live.small)} trend={trendOf(spark(live.small))} />
        <MetricCard label="bulk_goodput" value={goodputVal ? goodputVal.toFixed(1) : '—'} unit="Mbit/s" color="#b48ae0" spark={spark(live.goodput)} trend={trendOf(spark(live.goodput))} />
        <MetricCard label="drops" value={String(drops)} unit="" color={drops > 0 ? '#e22718' : '#767b84'} trend={drops > 0 ? 'up' : 'flat'} />
        <MetricCard label="wasted" value={wasted == null ? '—' : wasted ? (wasted > 1024 * 1024 ? (wasted / 1024 / 1024).toFixed(1) + ' MiB' : String(wasted)) : '0'} unit="bytes" color={wasted == null ? '#767b84' : '#f4b400'} trend={wasted != null && wasted > 0 ? 'up' : 'flat'} spark={spark(live.goodput)} />
        <MetricCard label="cost_ar_per_h" value={costAr == null ? '—' : costAr ? costAr.toFixed(0) : '0'} unit="Ar/h" color={costAr == null ? '#767b84' : '#f4b400'} trend={costAr != null && costAr > 0 ? 'up' : 'flat'} spark={spark(live.goodput)} />
        <MetricCard label="deadline_ok" value={deadlineOk === null ? '—' : deadlineOk.toFixed(0)} unit={deadlineOk === null ? '' : '%'} color={deadlineOk === null ? '#767b84' : deadlineOk >= 95 ? '#1fa348' : deadlineOk >= 80 ? '#f4b400' : '#e22718'} trend={deadlineOk === null ? 'flat' : deadlineOk >= 95 ? 'down' : 'up'} spark={spark(live.small)} />
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        <div data-testid="qdi-sparkline"><MetricCard label="QDI" value={!liveSnap || live.rtt95.length === 0 ? '—' : qdiVal.toFixed(1)} unit="ms" color="#f4b400" spark={live.rtt95.length === 0 ? undefined : qdiSpark} trend={trendOf(qdiSpark)} /></div>
        <div title={jfiVal === null ? 'JFI requiert détail par répétition (detail=1)' : undefined} style={{ border: '1px solid #26262a', background: 'var(--surface-card)', padding: 12, display: 'flex', flexDirection: 'column', gap: 6 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>JFI</span>
            <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, padding: '2px 6px', border: '1px solid #26262a', background: (jfiVal === null ? '#767b84' : jfiVal > 0.95 ? '#1fa348' : '#9aa3ad') + '14', color: jfiVal === null ? '#767b84' : jfiVal > 0.95 ? '#1fa348' : '#9aa3ad', lineHeight: 1 }}>{jfiVal === null ? '—' : jfiVal > 0.95 ? '—' : '↘'}</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 20, fontWeight: 700, color: '#f2f2f4', fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{jfiVal === null ? '—' : jfiVal.toFixed(2)}</span>
            <DonutJFI value={jfiVal} />
          </div>
        </div>
      </div>
      <div className="card" onMouseEnter={e => { const v = live.rtt95.at(-1)?.[1] ?? rttP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); rtt.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.rtt95.length - 1) }) }} onMouseLeave={() => setPeek(null)}><div ref={rtt.ref} style={{ height: 180 }} /></div>
      <div className="card" onMouseEnter={e => { const v = live.goodput.at(-1)?.[1] ?? goodputVal; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); goodput.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.goodput.length - 1) }) }} onMouseLeave={() => setPeek(null)}><div ref={goodput.ref} style={{ height: 180 }} /></div>
      <div className="kv" style={{ border: '1px solid #26262a', padding: '8px 12px' }}><span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>drops detail</span><b className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 12, color: drops > 0 ? '#e22718' : '#f2f2f4' }}>{drops}</b></div>

      {/* live-wall-overlay: baseline grey dashed vs CAKE cyan solid at same scale, diff badge -(baseline-best)/baseline*100% + hash 8-char */}
      <div className="live-wall-overlay card" data-testid="live-wall-overlay" style={{ gridColumn: '1 / -1', border: '1px solid #26262a', background: 'var(--surface-card)', padding: 16, display: 'flex', flexDirection: 'column', gap: 10, width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>Live wall — baseline vs CAKE</span>
          <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#767b84', fontVariantNumeric: 'tabular-nums' }}>hash {hash8} · provenance {wallGroups ? `${wallGroups.length} groupes` : 'en attente'}</span>
          {diffAB != null && (
            <span className="diff-badge mono" style={{ background: diffAB > 0 ? 'rgba(31,163,72,0.12)' : 'rgba(226,39,24,0.12)', border: '1px solid ' + (diffAB > 0 ? '#1fa348' : '#e22718'), color: diffAB > 0 ? '#1fa348' : '#e22718', padding: '4px 10px', fontSize: 14, fontWeight: 700, fontVariantNumeric: 'tabular-nums' }}>{diffAB > 0 ? `-${diffAB}%` : `${diffAB}%`}</span>
          )}
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div>
            <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#767b84', display: 'flex', justifyContent: 'space-between' }}>
              <span>baseline pfifo_fast — gris pointillé</span>
              <span style={{ fontVariantNumeric: 'tabular-nums', color: '#f2f2f4' }}>{baseline ? `${baseline.small_p95_median.toFixed(1)} ms` : '—'}</span>
            </div>
            <div style={{ height: 10, background: 'rgba(118,123,132,0.08)', border: '1px dashed #767b84', borderRadius: 2, overflow: 'hidden', marginTop: 4, position: 'relative' }}>
              <div style={{ width: `${baseline ? (baseline.small_p95_median / maxAB) * 100 : 0}%`, height: '100%', background: '#767b84', opacity: 0.9, transition: 'width 0.4s ease' }} />
            </div>
            <svg width="100%" height={2} style={{ display: 'block', marginTop: 2 }} aria-hidden><line x1={0} y1={1} x2="100%" y2={1} stroke="#767b84" strokeWidth={1} strokeDasharray="6 4" opacity={0.6} /></svg>
            <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#8b9099', marginTop: 2 }}>baseline · pfifo_fast median small_p95 — échelle commune</div>
          </div>
          <div>
            <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#5ad3e3', display: 'flex', justifyContent: 'space-between' }}>
              <span>CAKE — cyan solide</span>
              <span style={{ fontVariantNumeric: 'tabular-nums', color: '#f2f2f4' }}>{cakeBest ? `${cakeBest.small_p95_median.toFixed(1)} ms · ${cakeBest.qdisc}${cakeBest.best ? ' ★' : ''}` : '—'}</span>
            </div>
            <div style={{ height: 10, background: 'rgba(90,211,227,0.08)', border: '1px solid #5ad3e3', borderRadius: 2, overflow: 'hidden', marginTop: 4, position: 'relative' }}>
              <div style={{ width: `${cakeBest ? (cakeBest.small_p95_median / maxAB) * 100 : 0}%`, height: '100%', background: '#5ad3e3', boxShadow: '0 0 6px rgba(90,211,227,0.5)', transition: 'width 0.4s ease' }} />
            </div>
            <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#5ad3e3', marginTop: 2 }}>CAKE · best median small_p95 — même échelle que baseline</div>
          </div>
        </div>
        <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#767b84', borderTop: '1px solid var(--hairline-faint)', paddingTop: 8, display: 'flex', justifyContent: 'space-between', gap: 8, flexWrap: 'wrap' }}>
          <span>overlay: baseline gris pointillé vs CAKE cyan continu — même échelle</span>
          <span style={{ color: '#8b9099' }}>{wallGroups ? `données gelées · ${wallGroups.length} groupes · hash ${hash8}` : 'en attente — lancez campagne'}</span>
        </div>
      </div>

      {liveSnap?.running && <Beam />}
    </div>
  )
}
