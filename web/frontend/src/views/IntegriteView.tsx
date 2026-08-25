import { useEffect, useState } from 'react'
import { startReplay, stopReplay } from '../lib/replay'
import { useUIStore } from '../store/ui'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { PeekPopover } from '../components/PeekPopover'

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
  const [peek, setPeek] = useState<{ rect: DOMRect; run: string } | null>(null)
  const replayRunning = useUIStore(s=>s.replayRunning)
  const setPanel = useUIStore(s=>s.setPanel)
  const [groups, setGroups] = useState<any[]|null>(null)

  const load = () => {
    fetch('/api/integrity').then(r=>r.json()).then(j=>setData(j)).catch(e=>setErr(String(e)))
    fetch('/api/replay/list').then(r=>r.json()).then(j=>setReplayRuns(j.runs||[])).catch(()=>{})
    fetch('/api/results').then(r=>r.json()).then(j=>setGroups(j.groups??null)).catch(()=>setGroups(null))
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

  // ponytail: synthetic quarantine removed — render real gate_status from Scan when available
  return (
    <div className="panel-stack" style={{position:'relative'}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#8b9099', marginBottom:4}}>run — {peek.run}</div>
          {(() => {
            const gs = (groups ?? []).filter((g: any) => peek.run.includes(g.profile ?? ''))
            if (!gs.length) return <div className="mono" style={{fontSize:10, color:'#767b84'}}>aucun groupe — gel d'abord</div>
            return gs.slice(0, 4).map((g: any, i: number) => (
              <div key={i} style={{display:'flex', justifyContent:'space-between', gap:12, fontFamily:'JetBrains Mono', fontSize:10}}>
                <span style={{color:'#9aa3ad'}}>{g.profile}·{g.qdisc}·{g.cc}</span>
                <span style={{color: g.best ? '#1fa348' : '#f2f2f4', fontVariantNumeric:'tabular-nums'}}>{g.best ? '★ ' : ''}{Number(g.small_p95_median ?? 0).toFixed(1)} ms</span>
              </div>
            ))
          })()}
        </PeekPopover>
      )}
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
          {/* ponytail: fake sha removed — wire real manifest hash when /api/integrity exposes it */}
          {(data as any).sha256 ? (
            <div data-testid="provenance-hash" className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#f2f2f4', background:'#070707', border:'1px solid #26262a', padding:'8px 10px', fontVariantNumeric:'tabular-nums', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap'}}>
              sha256:{(data as any).sha256}
            </div>
          ) : (
            <div title="empreinte SHA non exposée par /api/integrity — afficher le manifest gelé">
              <EmptyState kind="empty" hint="empreinte SHA non exposée par /api/integrity — afficher le manifest gelé" />
            </div>
          )}
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
              <li key={id} style={{display:'flex', justifyContent:'space-between', padding:'6px 0', borderBottom:'1px solid var(--hairline-faint)', fontFamily:'var(--font-mono)', fontSize:12}} onMouseEnter={e=>setPeek({rect:(e.currentTarget as HTMLElement).getBoundingClientRect(), run:id})} onMouseLeave={()=>setPeek(null)}>
                <span>{id}</span>
                <a href={`/api/results?run=${id}`} target="_blank" rel="noreferrer" style={{color:'var(--t-live)'}}>résultats</a>
              </li>
            ))}
          </ul>
        }
      </div>
      <div className="card">
        <div className="card-head">quarantine — table de quarantaine</div>
        {/* ponytail: synthetic quarantine removed — render real gate_status from Scan when available */}
        {(data.quarantined||0)===0 ? (
          <div style={{padding:'8px 0'}}><EmptyState kind="empty" hint="aucune mise en quarantaine (gate_status=valid)" /></div>
        ) : (
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#f4b400', padding:'8px 10px', border:'1px solid #26262a', background:'rgba(244,180,0,0.06)'}}>
            {data.quarantined} événement(s) en quarantaine — détail gate_status via /api/results
          </div>
        )}
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
