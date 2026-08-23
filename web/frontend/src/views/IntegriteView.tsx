import { useEffect, useState } from 'react'

type Integrity = {
  available: boolean; reason?: string
  runs?: number; manifests?: number; valid?: number; quarantined?: number
  run_ids?: string[]
}

export default function IntegriteView() {
  const [data, setData] = useState<Integrity|null>(null)
  const [err, setErr] = useState<string|null>(null)
  const [msg, setMsg] = useState('')

  const load = () => {
    fetch('/api/integrity').then(r=>r.json()).then(j=>setData(j)).catch(e=>setErr(String(e)))
  }
  useEffect(()=>{ load() }, [])

  const verify = async () => {
    setMsg('vérification…')
    const r = await fetch('/api/integrity')
    const j = await r.json()
    setData(j); setMsg(j.available ? 'intégrité vérifiée' : j.reason)
  }

  if (err) return <div className="card"><h1 className="view-title">Intégrité</h1><p className="mono muted">{err}</p></div>
  if (!data) return <div className="card"><h1 className="view-title">Intégrité</h1><p className="mono muted">chargement…</p></div>
  if (!data.available) return <div className="card"><h1 className="view-title">Intégrité</h1><p className="mono muted">{data.reason}</p><button className="btn btn-primary" style={{marginTop:12}} onClick={load}>Réessayer</button></div>

  return (
    <div className="panel-stack">
      <h1 className="view-title">Intégrité — archives gelées</h1>
      <div className="card">
        <div className="kv"><span>runs</span><b className="mono">{data.runs}</b></div>
        <div className="kv"><span>manifests</span><b className="mono">{data.manifests}</b></div>
        <div className="kv"><span>événements valides</span><b className="mono" style={{color:'var(--t-ok)'}}>{data.valid}</b></div>
        <div className="kv"><span>quarantaine</span><b className="mono" style={{color: (data.quarantined||0)>0?'var(--t-danger)':'var(--text-muted)'}}>{data.quarantined}</b></div>
        <div className="form-row" style={{gap:8, marginTop:12}}>
          <button className="btn btn-primary" onClick={verify}>Vérifier manifestes</button>
          <a className="btn" href="/api/report/export?format=md" download style={{border:'1px solid var(--hairline)', padding:'7px 16px'}}>Rapport MD</a>
          <span className="mono muted">{msg}</span>
        </div>
      </div>
      <div className="card">
        <div className="card-head">Runs archivés</div>
        {(data.run_ids||[]).length===0 ? <p className="mono muted">aucun run</p> :
          <ul style={{listStyle:'none', padding:0, margin:0}}>
            {(data.run_ids||[]).map(id=>(
              <li key={id} style={{display:'flex', justifyContent:'space-between', padding:'6px 0', borderBottom:'1px solid var(--hairline-faint)', fontFamily:'var(--font-mono)', fontSize:12}}>
                <span>{id}</span>
                <a href={`/api/results?run=${id}`} target="_blank" rel="noreferrer" style={{color:'var(--t-live)'}}>résultats</a>
              </li>
            ))}
          </ul>
        }
      </div>
      <div className="card">
        <div className="card-head">Figures</div>
        <p className="mono muted" style={{fontSize:12, lineHeight:'1.6'}}>Générées par <span className="mono" style={{color:'var(--text-primary)'}}>make figures</span> depuis les CSV gelés. Re-génération via API prévue jalon M2.3 (SVG).</p>
        <button className="btn" disabled style={{marginTop:8, opacity:.45}}>Régénérer (M2.3)</button>
      </div>
    </div>
  )
}
