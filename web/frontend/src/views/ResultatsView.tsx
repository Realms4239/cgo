import { useEffect, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar } from '../lib/anime'
import { computeJFI } from '../lib/jfi'
import { PeekPopover } from '../components/PeekPopover'
import { echarts } from '../lib/echarts'

type Group = {
  profile: string; qdisc: string; cc: string
  count: number; quarantined: number
  rtt_p95_median: number; small_p95_median: number
  goodput_median: number
  deadline_median?: number; deadline_ok_pct?: number
  wasted_median?: number; cost_median?: number; wasted_bytes?: number; cost_ar_per_h?: number
  best?: boolean
}

export default function ResultatsView() {
  const [groups, setGroups] = useState<Group[]|null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [showCosts, setShowCosts] = useState(true)
  const [peek, setPeek] = useState<{ rect: DOMRect; g: Group } | null>(null)
  const scatterRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetch('/api/results').then(r=>r.json()).then(j=>{
      if (j.available) setGroups(j.groups)
      else setErr(j.reason || 'pas de résultats')
    }).catch(e=>setErr(String(e)))
  }, [])

  useEffect(() => {
    if (!groups) return
    const id = requestAnimationFrame(() => {
      document.querySelectorAll('.leader-bar').forEach((el) => animateBar(el))
    })
    return () => cancelAnimationFrame(id)
  }, [groups])

  // ponytail: scatter brush toolbox rect minimal — star 12 for best, clipPath deferred if perf matters
  useEffect(() => {
    if (!groups || !scatterRef.current) return
    const c = echarts.init(scatterRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(scatterRef.current)
    c.setOption({
      animation: false,
      grid: { left: 48, right: 16, top: 24, bottom: 32, containLabel: true },
      brush: { toolbox: ['rect'], brushType: 'rect' as const, xAxisIndex: 'all', yAxisIndex: 'all', brushMode: 'single' as const },
      toolbox: { feature: { brush: { type: ['rect'] } } },
      xAxis: { type: 'value' as const, name: 'goodput (Mbit/s)', axisLabel: { fontSize: 10, fontFamily: 'JetBrains Mono' } },
      yAxis: { type: 'value' as const, name: 'small p95 (ms)', axisLabel: { fontSize: 10, fontFamily: 'JetBrains Mono' } },
      tooltip: { trigger: 'item' as const },
      series: [{
        type: 'scatter' as const,
        data: groups.map(g => [g.goodput_median, g.small_p95_median]),
        symbolSize: (_val: any, params: any) => {
          const g = groups[(params as any).dataIndex]
          return g?.best ? 12 : 8
        },
        symbol: 'circle' as const,
        itemStyle: { color: '#5ad3e3' },
        emphasis: { itemStyle: { color: '#f4b400' } },
      }],
    } as any)
    return () => { ro.disconnect(); c.dispose() }
  }, [groups])

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="error" hint={err} /></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="loading" hint="agrégation des réplications" /></div>

  const maxSmall = Math.max(...groups.map(g=>g.small_p95_median), 1)
  return (
    <div className="panel-stack" style={{position:'relative'}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#8b9099'}}>{peek.g.profile} · {peek.g.qdisc} · {peek.g.cc}</div>
          <div style={{display:'flex', gap:6, marginTop:4, alignItems:'end'}}>
            <span className="mono" style={{fontSize:10, color:'#5ad3e3'}}>{peek.g.small_p95_median.toFixed(1)} ms</span>
            <span className="mono" style={{fontSize:10, color:'#767b84'}}>n={peek.g.count}</span>
            <div style={{flex:1, height:4, background:'var(--hairline-faint)', borderRadius:2, overflow:'hidden'}}>
              <div style={{width:`${(peek.g.small_p95_median/maxSmall)*100}%`, height:'100%', background: peek.g.best ? 'var(--t-ok)' : '#5ad3e3'}} />
            </div>
          </div>
        </PeekPopover>
      )}
      <h1 className="view-title">Résultats — comparaison AQM/BBR</h1>
      <div className="form-row" style={{justifyContent:'flex-end', gap:8, marginBottom:8}}>
        <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.08em', textTransform:'uppercase', color:'#8b9099'}}>affichage</span>
        <button className="btn" onClick={()=>setShowCosts(v=>!v)} style={{border:'1px solid #26262a', padding:'4px 10px', font:'700 10px JetBrains Mono', letterSpacing:'0.08em', textTransform:'uppercase', background: showCosts ? 'rgba(90,211,227,0.08)' : 'transparent', color: showCosts ? '#5ad3e3' : '#8b9099'}}>
          {showCosts ? 'masquer coûts' : 'afficher coûts'}
        </button>
        <span className="mono muted" style={{fontFamily:'JetBrains Mono', fontSize:10}}>deadline_ok · wasted · cost</span>
      </div>
      <div className="card" style={{overflowX:'auto'}}>
        <table style={{width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12}}>
          <thead>
            <tr style={{color:'var(--text-muted)', textAlign:'left', borderBottom:'1px solid var(--hairline)'}}>
              <th style={{padding:'6px 8px'}}>profil</th><th>qdisc</th><th>cc</th><th>n</th><th style={{minWidth:140}}>small p95</th><th>rtt p95</th><th>goodput</th><th title="Jain's fairness 0–1" style={{width:52, fontSize:11, fontFamily:'JetBrains Mono', color:'#9aa3ad'}}>JFI</th>
              {showCosts && <><th style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#8b9099'}}>deadline_ok</th><th style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#8b9099'}}>wasted</th><th style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#f4b400'}}>cost</th></>}
              <th>quar.</th><th>best</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((g,i)=>{
              const pct = (g.small_p95_median / maxSmall) * 100
              const barColor = g.best ? 'var(--t-ok)' : g.qdisc==='cake' ? 'var(--t-bbr)' : g.qdisc==='fq_codel' ? 'var(--t-live)' : 'var(--text-faint)'
              // ponytail: JFI degenerate guard — identical medians or n<2 → — until detail=1 provides per-rep values
              const jfiVals = Array.from({ length: g.count }, () => g.small_p95_median)
              const jfiDegenerate = g.count < 2 || jfiVals.every(v => v === jfiVals[0])
              const jfi: number | null = jfiDegenerate ? null : computeJFI(jfiVals)
              const wasted: number | null = g.wasted_median ?? g.wasted_bytes ?? null
              const cost: number | null = g.cost_median ?? g.cost_ar_per_h ?? null
              const deadlineOk: number | null = g.deadline_median ?? g.deadline_ok_pct ?? null
              return (
                <tr key={i} style={{borderBottom:'1px solid var(--hairline-faint)', background: g.best ? 'rgba(31,163,72,0.08)' : 'transparent'}} onMouseEnter={e=>setPeek({rect:e.currentTarget.getBoundingClientRect(), g})} onMouseLeave={()=>setPeek(null)}>
                  <td style={{padding:'6px 8px', fontWeight:g.best?700:400}}>{g.profile}</td>
                  <td>{g.qdisc}</td><td>{g.cc}</td><td>{g.count}</td>
                  <td>
                    <div style={{display:'flex', alignItems:'center', gap:8}}>
                      <div style={{flex:1, height:6, background:'var(--hairline-faint)', position:'relative', minWidth:80, borderRadius:2, overflow:'hidden'}}>
                        <div className="leader-bar" style={{position:'absolute', left:0, top:0, bottom:0, width:`${pct}%`, background:barColor, boxShadow: g.best ? `0 0 6px ${barColor}` : 'none', transformOrigin:'left center', borderRadius:2, filter: g.best ? `drop-shadow(0 0 4px ${barColor})` : 'none'}} />
                      </div>
                      <span style={{minWidth:45, textAlign:'right', fontVariantNumeric:'tabular-nums'}}>{g.small_p95_median.toFixed(1)}</span>
                    </div>
                  </td>
                  <td>{g.rtt_p95_median.toFixed(1)}</td>
                  <td>{g.goodput_median.toFixed(1)}</td>
                  <td title={jfi===null ? "JFI requiert détail par répétition (detail=1)" : undefined} style={{width:52, fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: jfi===null ? '#767b84' : jfi > 0.95 ? '#1fa348' : '#9aa3ad', textAlign:'right'}}>{jfi===null ? '—' : jfi.toFixed(2)}</td>
                  {showCosts && <>
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: deadlineOk == null ? '#767b84' : deadlineOk >=95 ? '#1fa348' : '#f4b400', textAlign:'right'}}>{deadlineOk == null ? '—' : deadlineOk.toFixed(0)+'%'}</td>
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: wasted == null ? '#767b84' : wasted>0 ? '#e22718' : '#767b84', textAlign:'right'}}>{wasted == null ? '—' : wasted ? (wasted>1024*1024 ? (wasted/1024/1024).toFixed(1)+'M' : String(wasted)) : '0'}</td>
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: cost == null ? '#767b84' : cost>0 ? '#f4b400' : '#767b84', textAlign:'right'}}>{cost == null ? '—' : cost ? cost.toFixed(0) : '0'}</td>
                  </>}
                  <td>{g.quarantined}</td>
                  <td>{g.best ? '★' : ''}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      <div className="card" style={{ padding: 12 }}>
        <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099', marginBottom: 6 }}>goodput vs small — brush rect pour comparer</div>
        <div ref={scatterRef} style={{ height: 220 }} />
      </div>
      <div className="form-row" style={{gap:8}}>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{border:'1px solid var(--hairline)', padding:'7px 16px'}}>Exporter MD</a>
        <span className="mono muted" style={{marginLeft:8}}>médianes — ★ meilleur small p95 par profil · barres relatives au max</span>
      </div>
      <Provenance source="data/runs/*/aqm_eval.csv" state="live" extra={`${groups.length} groupes · max small ${maxSmall.toFixed(1)} ms`} />
    </div>
  )
}
