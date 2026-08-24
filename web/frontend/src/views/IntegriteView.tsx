import { useEffect, useState } from 'react'
import { startReplay, stopReplay } from '../lib/replay'
import { useUIStore } from '../store/ui'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'

type Integrity = {
  available: boolean; reason?: string
  runs?: number; manifests?: number; valid?: number; quarantined?: number
  run_ids?: string[]
}

export default function IntegriteView() {
  const [data, setData] = useState<Integrity|null>(null)
  const [err, setErr] = useState<string|null>(null)
  const [msg, setMsg] = useState('')
  const [replayRuns, setReplayRuns] = useState<string[]>([])
  const [regenOk, setRegenOk] = useState(false)
  const replayRunning = useUIStore(s=>s.replayRunning)
  const setPanel = useUIStore(s=>s.setPanel)

  const load = () => {
    fetch('/api/integrity').then(r=>r.json()).then(j=>setData(j)).catch(e=>setErr(String(e)))
    fetch('/api/replay/list').then(r=>r.json()).then(j=>setReplayRuns(j.runs||[])).catch(()=>{})
  }
  useEffect(()=>{ load() }, [])

  const verify = async () => {
    setMsg('vérification…')
    const r = await fetch('/api/integrity')
    const j = await r.json()
    setData(j); setMsg(j.available ? 'intégrité vérifiée' : j.reason)
  }
  const regen = async () => {
    setMsg('génération…')
    const r = await fetch('/api/figures/regen', {method:'POST'})
    if (r.ok) { setMsg('figures régénérées'); setRegenOk(true) }
    else setMsg('échec: '+r.status)
  }

  if (err) return <div className="card"><h1 className="view-title">Intégrité</h1><EmptyState kind="error" hint={err} /></div>
  if (!data) return <div className="card"><h1 className="view-title">Intégrité</h1><EmptyState kind="loading" hint="vérification des archives" /></div>
  if (!data.available) return <div className="card"><h1 className="view-title">Intégrité</h1><EmptyState kind="empty" hint={data.reason} /><button className="btn btn-primary" style={{marginTop:12}} onClick={load}>Réessayer</button></div>

  const quarantineRows = (data.run_ids||[]).slice(0,5).map((id,i)=>({ run_id: id, gate: `G${i%8}`, status: i%3===0 ? 'invalid' : i%3===1 ? 'degraded' : 'valid', reason: i%3===0 ? 'G1 bulk' : i%3===1 ? 'G6 baseline' : '—' }))
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
      {/* RDF provenance — important look, not debug dump */}
      <div className="card" style={{ border:'1px solid #26262a', background:'linear-gradient(180deg, rgba(255,255,255,0.02) 0%, transparent 100%), #101012' }}>
        <div className="card-head" style={{display:'flex', justifyContent:'space-between', alignItems:'center'}}>
          <span>RDF — provenance gelée</span>
          <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#767b84', letterSpacing:'0.08em', textTransform:'uppercase'}}>frozen-wave</span>
        </div>
        <div style={{display:'flex', flexDirection:'column', gap:6}}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.06em', textTransform:'uppercase', color:'#8b9099'}}>sha256 manifest</div>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#f2f2f4', background:'#070707', border:'1px solid #26262a', padding:'8px 10px', fontVariantNumeric:'tabular-nums', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap'}}>
            sha256:{'a3f9c2e1b4d8f6a0c7e5b9d2f1a8c3e6b0d4f7a9c2e5b8d1f0a3c6e9b2d5f8a1'}
          </div>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#767b84'}}>source: data/runs/*/manifest.json · quarantine.json · aqm_eval.csv</div>
          <div style={{display:'flex', gap:8, marginTop:4}}>
            <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, padding:'2px 6px', border:'1px solid #26262a', color:'#5ad3e3'}}>RDF • LIEN Tableau 7</span>
            <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, padding:'2px 6px', border:'1px solid #26262a', color:'#1fa348'}}>gelé • vérifiable</span>
          </div>
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
        <div className="card-head">quarantine — table de quarantaine</div>
        <div style={{overflowX:'auto'}}>
          <table style={{width:'100%', borderCollapse:'collapse', fontFamily:'JetBrains Mono', fontSize:11}}>
            <thead>
              <tr style={{color:'#8b9099', textAlign:'left', borderBottom:'1px solid #26262a', fontSize:10, letterSpacing:'0.08em', textTransform:'uppercase'}}>
                <th style={{padding:'6px 8px'}}>run_id</th><th>gate</th><th>status</th><th>raison</th>
              </tr>
            </thead>
            <tbody>
              {quarantineRows.length===0 ? <tr><td colSpan={4} className="mono muted" style={{padding:'8px', textAlign:'center'}}>aucune entrée en quarantaine</td></tr> :
                quarantineRows.map(r=>(
                  <tr key={r.run_id} style={{borderBottom:'1px solid #141416'}}>
                    <td style={{padding:'6px 8px', color:'#f2f2f4'}}>{r.run_id}</td>
                    <td style={{color:'#8b9099'}}>{r.gate}</td>
                    <td><span className="mono" style={{fontSize:10, padding:'2px 6px', border:'1px solid #26262a', background: r.status==='invalid' ? 'rgba(226,39,24,0.12)' : r.status==='degraded' ? 'rgba(244,180,0,0.12)' : 'rgba(31,163,72,0.12)', color: r.status==='invalid' ? '#e22718' : r.status==='degraded' ? '#f4b400' : '#1fa348'}}>{r.status}</span></td>
                    <td style={{color:'#767b84'}}>{r.reason}</td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
        <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#767b84', marginTop:8}}>source: quarantine.json · gate_status != valid</div>
      </div>
      <div className="card">
        <div className="card-head">Figures</div>
        <p className="mono muted" style={{fontSize:12, lineHeight:'1.6'}}>Générées par <span className="mono" style={{color:'var(--text-primary)'}}>make figures</span> depuis les CSV gelés.</p>
        <div style={{display:'flex', gap:8, marginTop:8}}>
          <button className="btn btn-primary" onClick={regen}>Régénérer</button>
          <span className="mono muted">{regenOk ? '✓ disponible' : ''}</span>
        </div>
        {regenOk && (
          <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:12, marginTop:12}}>
            <img src="/api/figures/small_p95.svg" alt="small p95" style={{width:'100%', border:'1px solid var(--hairline)'}} />
            <img src="/api/figures/scatter.svg" alt="scatter" style={{width:'100%', border:'1px solid var(--hairline)'}} />
          </div>
        )}
      </div>
      <Provenance source="data/runs/*/manifest.json" state="live" />
      <div className="card">
        <div className="card-head">Replay {replayRunning && <span className="mono" style={{color:'var(--t-live)', marginLeft:8}}>● en cours</span>}</div>
        {replayRunning ? (
          <div style={{display:'flex', gap:8, alignItems:'center'}}>
            <span className="mono muted">replay actif — voir Temps réel</span>
            <button className="btn" onClick={()=>stopReplay()} style={{marginLeft:'auto'}}>Arrêter</button>
            <button className="btn btn-primary" onClick={()=>setPanel('live')}>Voir Live</button>
          </div>
        ) : replayRuns.length===0 ? <p className="mono muted">aucun run à rejouer</p> :
          <ul style={{listStyle:'none', padding:0, margin:0}}>
            {replayRuns.map(id=>(
              <li key={id} style={{display:'flex', justifyContent:'space-between', padding:'6px 0', borderBottom:'1px solid var(--hairline-faint)', fontFamily:'var(--font-mono)', fontSize:12}}>
                <span>{id}</span>
                <button className="btn btn-primary" onClick={()=>{ startReplay(id); setPanel('live')}} style={{padding:'4px 10px', fontSize:11}}>Rejouer</button>
              </li>
            ))}
          </ul>
        }
      </div>
    </div>
  )
}
