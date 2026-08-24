import { useEffect, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'

type Group = {
  profile: string; qdisc: string; cc: string
  count: number; quarantined: number
  rtt_p95_median: number; small_p95_median: number
  goodput_median: number; best?: boolean
}

export default function ResultatsView() {
  const [groups, setGroups] = useState<Group[]|null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/results').then(r=>r.json()).then(j=>{
      if (j.available) setGroups(j.groups)
      else setErr(j.reason || 'pas de résultats')
    }).catch(e=>setErr(String(e)))
  }, [])

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="error" hint={err} /></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="loading" hint="agrégation des réplications" /></div>

  const maxSmall = Math.max(...groups.map(g=>g.small_p95_median), 1)
  return (
    <div className="panel-stack">
      <h1 className="view-title">Résultats — comparaison AQM/BBR</h1>
      <div className="card" style={{overflowX:'auto'}}>
        <table style={{width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12}}>
          <thead>
            <tr style={{color:'var(--text-muted)', textAlign:'left', borderBottom:'1px solid var(--hairline)'}}>
              <th style={{padding:'6px 8px'}}>profil</th><th>qdisc</th><th>cc</th><th>n</th><th style={{minWidth:140}}>small p95</th><th>rtt p95</th><th>goodput</th><th>quar.</th><th>best</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((g,i)=>{
              const pct = (g.small_p95_median / maxSmall) * 100
              const barColor = g.best ? 'var(--t-ok)' : g.qdisc==='cake' ? 'var(--t-bbr)' : g.qdisc==='fq_codel' ? 'var(--t-live)' : 'var(--text-faint)'
              return (
                <tr key={i} style={{borderBottom:'1px solid var(--hairline-faint)', background: g.best ? 'rgba(31,163,72,0.08)' : 'transparent'}}>
                  <td style={{padding:'6px 8px', fontWeight:g.best?700:400}}>{g.profile}</td>
                  <td>{g.qdisc}</td><td>{g.cc}</td><td>{g.count}</td>
                  <td>
                    <div style={{display:'flex', alignItems:'center', gap:8}}>
                      <div style={{flex:1, height:6, background:'var(--hairline-faint)', position:'relative', minWidth:80}}>
                        <div style={{position:'absolute', left:0, top:0, bottom:0, width:`${pct}%`, background:barColor, boxShadow: g.best ? `0 0 6px ${barColor}` : 'none', transition:'width 0.6s ease'}} />
                      </div>
                      <span style={{minWidth:45, textAlign:'right', fontVariantNumeric:'tabular-nums'}}>{g.small_p95_median.toFixed(1)}</span>
                    </div>
                  </td>
                  <td>{g.rtt_p95_median.toFixed(1)}</td>
                  <td>{g.goodput_median.toFixed(1)}</td>
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
