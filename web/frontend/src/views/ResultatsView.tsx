import { useEffect, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar } from '../lib/anime'
import { computeJFI } from '../lib/jfi'

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

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="error" hint={err} /></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="loading" hint="agrégation des réplications" /></div>

  const maxSmall = Math.max(...groups.map(g=>g.small_p95_median), 1)
  return (
    <div className="panel-stack">
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
              // ponytail: wasted/cost fallback linear, non-linear if thesis needs
              const wasted = (g as any).wasted_median ?? (g as any).wasted_bytes ?? g.quarantined * 1448
              const cost = (g as any).cost_median ?? (g as any).cost_ar_per_h ?? (wasted / (4.5*1024*1024*1024))*30000
              const deadlineOk = (g as any).deadline_median ?? (g as any).deadline_ok_pct ?? 100
              return (
                <tr key={i} style={{borderBottom:'1px solid var(--hairline-faint)', background: g.best ? 'rgba(31,163,72,0.08)' : 'transparent'}}>
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
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: deadlineOk >=95 ? '#1fa348' : '#f4b400', textAlign:'right'}}>{typeof deadlineOk==='number' ? deadlineOk.toFixed(0)+'%' : '—'}</td>
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: wasted>0 ? '#e22718' : '#767b84', textAlign:'right'}}>{wasted ? (wasted>1024*1024 ? (wasted/1024/1024).toFixed(1)+'M' : String(wasted)) : '0'}</td>
                    <td style={{fontFamily:'JetBrains Mono', fontSize:11, fontVariantNumeric:'tabular-nums', color: cost>0 ? '#f4b400' : '#767b84', textAlign:'right'}}>{cost ? cost.toFixed(0) : '0'}</td>
                  </>}
                  <td>{g.quarantined}</td>
                  <td>{g.best ? '★' : ''}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
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
