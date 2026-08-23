import { useEffect, useState } from 'react'

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

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><p className="mono muted">{err}</p></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><p className="mono muted">chargement…</p></div>

  return (
    <div className="panel-stack">
      <h1 className="view-title">Résultats — comparaison AQM/BBR</h1>
      <div className="card" style={{overflowX:'auto'}}>
        <table style={{width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12}}>
          <thead>
            <tr style={{color:'var(--text-muted)', textAlign:'left', borderBottom:'1px solid var(--hairline)'}}>
              <th style={{padding:'6px 8px'}}>profil</th><th>qdisc</th><th>cc</th><th>n</th><th>small p95</th><th>rtt p95</th><th>goodput</th><th>quar.</th><th>best</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((g,i)=>(
              <tr key={i} style={{borderBottom:'1px solid var(--hairline-faint)', background: g.best ? 'rgba(31,163,72,0.08)' : 'transparent'}}>
                <td style={{padding:'6px 8px'}}>{g.profile}</td>
                <td>{g.qdisc}</td><td>{g.cc}</td><td>{g.count}</td>
                <td>{g.small_p95_median.toFixed(1)}</td>
                <td>{g.rtt_p95_median.toFixed(1)}</td>
                <td>{g.goodput_median.toFixed(1)}</td>
                <td>{g.quarantined}</td>
                <td>{g.best ? '★' : ''}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="form-row" style={{gap:8}}>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{border:'1px solid var(--hairline)', padding:'7px 16px'}}>Exporter MD</a>
        <span className="mono muted" style={{marginLeft:8}}>médianes sur réplications — ★ = meilleur small p95 par profil</span>
      </div>
    </div>
  )
}
