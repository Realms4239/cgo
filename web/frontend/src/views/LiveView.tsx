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

// ponytail: per-series CHARGE markArea amber 0.04 dashed insideTop Mono — never at option root
function chargeMarkArea(chargeStart: number, chargeEnd: number) {
  return {
    itemStyle: { color: 'rgba(244,180,0,0.04)', borderColor: 'rgba(244,180,0,0.12)', borderWidth: 1, borderType: 'dashed' as const },
    label: { show: true, color: '#f4b400', fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop' as const, padding: [4, 8] as unknown as number[], backgroundColor: 'rgba(244,180,0,0.08)', formatter: 'CHARGE' },
    data: [[{ xAxis: chargeStart }, { xAxis: chargeEnd }]] as any,
  }
}

function useChart(title:string, unit:string) {
  const ref = useRef<HTMLDivElement>(null)
  const chart = useRef<echarts.ECharts | null>(null)
  useEffect(() => {
    if(!ref.current) return
    const c = echarts.init(ref.current, undefined, { renderer: 'canvas', useDirtyRect: true, devicePixelRatio: Math.min(window.devicePixelRatio, 2) } as any)
    chart.current = c
    const ro = new ResizeObserver(()=>c.resize())
    ro.observe(ref.current)
    return () => { ro.disconnect(); c.dispose() }
  }, [])
  const setData = (series:ReturnType<typeof lineSeries>[], extra?: Record<string, unknown>) => {
    if(!chart.current) return
    // ponytail: brush toolbox rect minimal — full toolbox if UX needs
    const brush = { toolbox: ['rect'], brushType: 'rect' as const, xAxisIndex: 'all' as const, brushMode: 'single' as const }
    // hide scrubber when empty — dataZoom lives in baseOption, clear it explicitly (merge semantics)
    const empty = !series.some(s => ((s.data as unknown[]) ?? []).length > 1)
    const base = baseOption(title, unit)
    const opt = { animation: false, ...base, ...(empty ? { dataZoom: [] } : {}), brush, ...extra, series } as unknown as EChartsOption
    chart.current.setOption(opt)
  }
  return { ref, setData, chart }
}

export default function LiveView() {
  const rtt = useChart('RTT (ms)', 'ms')
  const small = useChart('Petits objets p95 (ms)', 'ms')
  const goodput = useChart('Bulk goodput (Mbit/s)', 'Mbit/s')
  const liveSnap = useUIStore(s=>s.live)
  const replayRunning = useUIStore(s=>s.replayRunning)
  const replayRunId = useUIStore(s=>s.replayRunId)
  const bannerRef = useRef<HTMLDivElement>(null)
  const [peek, setPeek] = useState<{ rect: DOMRect; value: number } | null>(null)

  const lastRef = useRef(0)
  useRafLoop((ts) => {
    if (ts - lastRef.current < 250) return
    lastRef.current = ts
    const d = (r: [number,number][]) => r.length > 400 ? lttb(r, 400) : r
    const chargeStart = live.rtt95.length ? live.rtt95[Math.floor(live.rtt95.length * 0.25)]?.[0] ?? Date.now() - 45000 : Date.now() - 45000
    const chargeEnd = live.rtt95.length ? live.rtt95[Math.floor(live.rtt95.length * 0.75)]?.[0] ?? Date.now() - 10000 : Date.now() - 10000
    const ma = chargeMarkArea(chargeStart, chargeEnd)
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
    return null // ponytail: JFI null until windowed detail with variance — no synthetic 1.00
  })()

  // Tableau 4/5 — LIEN primary + secondary for important look
  const rttP95 = liveSnap?.rtt_p95_ms ?? live.rtt95.at(-1)?.[1] ?? 0
  const rttP50 = liveSnap?.rtt_p50_ms ?? live.rtt50.at(-1)?.[1] ?? 0
  const smallP95 = liveSnap?.small_p95_ms ?? live.small.at(-1)?.[1] ?? 0
  const goodputVal = liveSnap?.bulk_goodput_mbps ?? live.goodput.at(-1)?.[1] ?? 0
  const drops = liveSnap?.drops ?? 0
  // idle snapshots carry engine zeros — deadline 0% red reads as SLA breach; '—' is honest when no data has flowed
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
  const jfiS = live.goodput.slice(-20).map(([, v]) => v).filter(v => v > 0)

  useEffect(() => {
    if (bannerRef.current) animateBannerPulse(bannerRef.current)
  }, [banner])

  useEffect(() => {
    animateLiveEnter()
  }, [])

  return (
    <div className="panel-stack" style={{position:'relative'}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#8b9099'}}>live p95 — point</div>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:14, color:'#5ad3e3'}}>{peek.value.toFixed(1)} ms</div>
          <div style={{display:'flex', alignItems:'end', gap:1, height:20, marginTop:4}}>
            {live.rtt95.slice(-20).map(([,v]:[number,number],i:number)=><i key={i} style={{flex:1, height:`${Math.max(2, Math.min(20,(v/300)*20))}px`, background:'#5ad3e3', borderRadius:1, opacity:0.6+i*0.02}} />)}
          </div>
        </PeekPopover>
      )}
      <div ref={bannerRef} className="banner mono" style={{ color: bannerColor, borderColor: bannerColor + '55' }}>{banner} · {replayRunning ? 'replay' : 'SSE 10 Hz'}</div>
      {!liveSnap && live.rtt95.length===0 && <div className="card" style={{border:'1px dashed var(--hairline)', background:'rgba(255,255,255,0.02)', textAlign:'center'}}><EmptyState kind="empty" hint="en attente — Démarrer depuis Campagne pour alimenter le Live" /></div>}
      {/* important look — all LIEN Tableau 4/5 metrics */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: 12 }}>
        <MetricCard label="rtt_p95" value={rttP95 ? rttP95.toFixed(1) : '—'} unit="ms" color="#5ad3e3" spark={spark(live.rtt95)} trend={trendOf(spark(live.rtt95))} />
        <MetricCard label="rtt_p50" value={rttP50 ? rttP50.toFixed(1) : '—'} unit="ms" color="#5ad3e3" spark={spark(live.rtt50)} trend={trendOf(spark(live.rtt50))} />
        <MetricCard label="small_p95" value={smallP95 ? smallP95.toFixed(1) : '—'} unit="ms" color="#1fa348" spark={spark(live.small)} trend={trendOf(spark(live.small))} />
        <MetricCard label="bulk_goodput" value={goodputVal ? goodputVal.toFixed(1) : '—'} unit="Mbit/s" color="#b48ae0" spark={spark(live.goodput)} trend={trendOf(spark(live.goodput))} />
        <MetricCard label="drops" value={String(drops)} unit="" color={drops > 0 ? '#e22718' : '#767b84'} trend={drops > 0 ? 'up' : 'flat'} />
        <MetricCard label="wasted" value={wasted == null ? '—' : wasted ? (wasted > 1024 * 1024 ? (wasted / 1024 / 1024).toFixed(1) + ' MiB' : String(wasted)) : '0'} unit="bytes" color={wasted == null ? '#767b84' : '#f4b400'} trend={wasted != null && wasted > 0 ? 'up' : 'flat'} />
        <MetricCard label="cost_ar_per_h" value={costAr == null ? '—' : costAr ? costAr.toFixed(0) : '0'} unit="Ar/h" color={costAr == null ? '#767b84' : '#f4b400'} trend={costAr != null && costAr > 0 ? 'up' : 'flat'} />
        <MetricCard label="deadline_ok" value={deadlineOk === null ? '—' : deadlineOk.toFixed(0)} unit={deadlineOk === null ? '' : '%'} color={deadlineOk === null ? '#767b84' : deadlineOk >= 95 ? '#1fa348' : deadlineOk >= 80 ? '#f4b400' : '#e22718'} trend={deadlineOk === null ? 'flat' : deadlineOk >= 95 ? 'down' : 'up'} />
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        <div data-testid="qdi-sparkline"><MetricCard label="QDI" value={!liveSnap ? '—' : qdiVal.toFixed(1)} unit="ms" color="#f4b400" spark={qdiSpark} trend={trendOf(qdiSpark)} /></div>
        <div title={jfiVal === null ? "JFI requiert détail par répétition (detail=1)" : undefined}>
          <MetricCard label="JFI" value={jfiVal === null ? '—' : jfiVal.toFixed(2)} unit="" color={jfiVal === null ? '#767b84' : '#9aa3ad'} spark={jfiS.length >= 2 ? jfiS : undefined} trend={jfiVal === null ? 'flat' : jfiVal > 0.95 ? 'flat' : 'down'} />
        </div>
      </div>
      {/* ponytail: hidden stubs removed — real sparkline in Task 3 card */}
      <div className="card" onMouseEnter={e=>{ const v=live.rtt95.at(-1)?.[1] ?? rttP95; setPeek({rect:e.currentTarget.getBoundingClientRect(), value:v}); rtt.chart.current?.dispatchAction({type:'showTip', seriesIndex:0, dataIndex: Math.max(0, (live.rtt95.length-1))}) }} onMouseLeave={()=>setPeek(null)}><div ref={rtt.ref} style={{ height: 220 }} /></div>
      <div className="card" onMouseEnter={e=>{ const v=live.small.at(-1)?.[1] ?? smallP95; setPeek({rect:e.currentTarget.getBoundingClientRect(), value:v}); small.chart.current?.dispatchAction({type:'showTip', seriesIndex:0, dataIndex: Math.max(0, (live.small.length-1))}) }} onMouseLeave={()=>setPeek(null)}><div ref={small.ref} style={{ height: 180 }} /></div>
      <div className="card" onMouseEnter={e=>{ const v=live.goodput.at(-1)?.[1] ?? goodputVal; setPeek({rect:e.currentTarget.getBoundingClientRect(), value:v}); goodput.chart.current?.dispatchAction({type:'showTip', seriesIndex:0, dataIndex: Math.max(0, (live.goodput.length-1))}) }} onMouseLeave={()=>setPeek(null)}><div ref={goodput.ref} style={{ height: 180 }} /></div>
      <div className="kv" style={{ border: '1px solid #26262a', padding: '8px 12px' }}><span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>drops detail</span><b className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 12, color: drops > 0 ? '#e22718' : '#f2f2f4' }}>{drops}</b></div>
    </div>
  )
}
