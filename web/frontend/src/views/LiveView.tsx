import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import { baseOption, lineSeries } from '../lib/chartGrammar'
import { live } from '../lib/live'
import { useUIStore } from '../store/ui'

function useChart(title:string, unit:string) {
  const ref = useRef<HTMLDivElement>(null)
  const chart = useRef<echarts.ECharts | null>(null)
  useEffect(() => {
    if(!ref.current) return
    const c = echarts.init(ref.current, undefined, { renderer: 'canvas' })
    chart.current = c
    const ro = new ResizeObserver(()=>c.resize())
    ro.observe(ref.current)
    return () => { ro.disconnect(); c.dispose() }
  }, [])
  const setData = (series:any[]) => {
    if(!chart.current) return
    chart.current.setOption({ ...baseOption(title, unit), series } as any)
  }
  return { ref, setData }
}

export default function LiveView() {
  const rtt = useChart('RTT (ms)', 'ms')
  const small = useChart('Petits objets p95 (ms)', 'ms')
  const goodput = useChart('Bulk goodput (Mbit/s)', 'Mbit/s')
  const liveSnap = useUIStore(s=>s.live)

  useEffect(() => {
    let raf=0
    const tick = () => {
      rtt.setData([
        lineSeries('p50', live.rtt50 as any, '#5ad3e3'),
        lineSeries('p95', live.rtt95 as any, '#1fa348'),
      ])
      small.setData([ lineSeries('small p95', live.small as any, '#f4b400') ])
      goodput.setData([ lineSeries('goodput', live.goodput as any, '#b48ae0', true) ])
      raf = requestAnimationFrame(tick)
    }
    const id = setInterval(tick, 250)
    return () => { clearInterval(id); cancelAnimationFrame(raf) }
  }, [])

  const banner = !liveSnap ? 'OFFLINE — en attente du flux'
    : liveSnap.load_status === 'bulk-on' ? 'CHARGE — bulk actif'
    : liveSnap.phase === 'baseline' ? 'BASELINE'
    : liveSnap.phase === 'recup' ? 'RÉCUPÉRATION'
    : 'IDLE'

  return (
    <div className="panel-stack">
      <div className="banner mono">{banner} · SSE 10 Hz</div>
      <div className="card"><div ref={rtt.ref} style={{height:220}} /></div>
      <div className="card"><div ref={small.ref} style={{height:180}} /></div>
      <div className="card"><div ref={goodput.ref} style={{height:180}} /></div>
      <div className="kv"><span>drops</span><b className="mono">{liveSnap?.drops ?? '—'}</b></div>
    </div>
  )
}
