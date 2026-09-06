import { useEffect, useState } from 'react'
import { startReplay, stopReplay } from '../lib/replay'
import { useUIStore } from '../store/ui'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { PeekPopover } from '../components/PeekPopover'
import { hardwareRecommendation } from '../lib/hardware'
import { asArray } from '../lib/format'

type Integrity = {
  available: boolean; reason?: string
  runs?: number; manifests?: number; valid?: number; quarantined?: number
  run_ids?: string[]
  breakdown?: { run: string; rows: number; valid: number; quarantined: number }[]
  updated?: string
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
  const [peekGroups, setPeekGroups] = useState<any[]|null>(null)
  const [quar, setQuar] = useState<{run:string;event_id:number;profile:string;qdisc:string;cc:string;gate_status:string;failed_gates?:string[]}[]>([])
  const [quarSum, setQuarSum] = useState<{total_invalid?:number;g4_low?:number;g4_high?:number;empty_probes?:number;g3_implausible?:number}|null>(null)

  const load = () => {
    fetch('/api/integrity').then(r=>r.json()).then(j=>setData(j)).catch(e=>setErr(String(e)))
    fetch('/api/replay/list').then(r=>r.json()).then(j=>setReplayRuns(asArray<string>(j.runs))).catch(()=>{})
    fetch('/api/results').then(r=>r.json()).then(j=>setGroups(asArray(j.groups))).catch(()=>setGroups([]))
    fetch('/api/quarantine').then(r=>r.json()).then(j=>setQuar(asArray(j.quarantines))).catch(()=>{})
    fetch('/api/quarantine/summary').then(r=>r.json()).then(j=>setQuarSum(j)).catch(()=>{})
  }
  useEffect(()=>{ load() }, [])
  useEffect(()=>{
    if (!peek) return
    setPeekGroups(null)
    fetch(`/api/results?run=${encodeURIComponent(peek.run)}`).then(r=>r.json()).then(j=>setPeekGroups(asArray(j.groups))).catch(()=>setPeekGroups([]))
  },[peek?.run])

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

  // Quarantaine synthétique retirée — gate_status réel depuis Scan.
  return (
    <div className="panel-stack" style={{position:'relative', maxHeight:'calc(100vh - 48px - 28px)', overflowY:'auto'}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#8b9099', marginBottom:4}}>run — {peek.run}</div>
          {(() => {
            const src = asArray<any>(peekGroups).length ? peekGroups as any[] : asArray<any>(groups)
            if (src.length===0) return <div className="mono" style={{fontSize:10, color:'#767b84'}}>aucun groupe — gel d'abord</div>
            // Meilleur par profil pour ce run, repli sur les 4 premiers.
            const bests = src.filter((g:any)=>g.best)
            const show = bests.length ? bests.slice(0,4) : src.slice(0,4)
            return show.map((g: any, i: number) => (
              <div key={i} style={{display:'flex', justifyContent:'space-between', gap:12, fontFamily:'JetBrains Mono', fontSize:10}}>
                <span style={{color:'#9aa3ad'}}>{g.profile}·{g.qdisc}·{g.cc}</span>
                <span style={{color: g.best ? '#1fa348' : '#f2f2f4', fontVariantNumeric:'tabular-nums'}}>{g.best ? '★ ' : ''}{Number(g.small_p95_median ?? 0).toFixed(1)} ms</span>
              </div>
            ))
          })()}
        </PeekPopover>
      )}
      <h1 className="view-title">Intégrité — archives gelées</h1>
      {/* bandeau preuve resserré — une ligne mono, pas des cartes */}
      <div className="card" data-testid="proof-banner" style={{ display:'flex', gap:16, alignItems:'baseline', flexWrap:'wrap', padding:'10px 14px' }}>
        <span className="mono" style={{ fontSize:12, fontWeight:700 }}>Preuve gelée</span>
        <span className="mono" data-testid="proof-runs" style={{ fontSize:11 }}>{data.runs} runs · {data.manifests} manifests</span>
        <span className="mono" style={{ fontSize:11, color:'var(--t-ok)' }}>{data.valid} valides</span>
        <span className="mono" style={{ fontSize:11, color:(data.quarantined||0)>0?'var(--t-danger)':'var(--text-muted)' }}>{data.quarantined} quarantaine</span>
        <span className="mono muted" style={{ fontSize:11, marginLeft:'auto' }}>maj {data.updated || '—'}</span>
      </div>
      <div className="card">
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
          {/* SHA réel exposé par /api/integrity */}
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
      <div className="card" data-testid="runs-table">
        <div className="card-head">Runs archivés</div>
        <table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12 }}>
          <thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
            <th style={{ padding:'6px 8px' }}>run</th><th>lignes</th><th>valides</th><th>quar.</th><th></th><th></th>
          </tr></thead><tbody>
          {(Array.isArray(data.breakdown) ? data.breakdown : (asArray<string>(data.run_ids).map(id=>({run:id,rows:0,valid:0,quarantined:0})))).map(b=>(
            <tr key={b.run} style={{ borderBottom:'1px solid var(--hairline-faint)' }} onMouseEnter={e=>setPeek({rect:(e.currentTarget as HTMLElement).getBoundingClientRect(), run:b.run})} onMouseLeave={()=>setPeek(null)}>
              <td style={{ padding:'6px 8px' }}>{b.run}</td><td>{b.rows}</td><td>{b.valid}</td><td>{b.quarantined}</td>
              <td><a href={`/api/results?run=${b.run}`} target="_blank" rel="noreferrer" style={{ color:'var(--t-live)' }}>résultats</a></td>
              <td><button className="btn btn-primary" onClick={()=>{ startReplay(b.run); setPanel('live') }} style={{ padding:'4px 10px', fontSize:11 }}>Rejouer</button></td>
            </tr>
          ))}
          </tbody></table>
      </div>
      <div className="card" data-testid="quarantine-table">
        <div className="card-head">Quarantaine — lignes invalidées par les portes</div>
        {quar.length===0 ? <div style={{padding:'8px 0'}}><EmptyState kind="empty" hint="aucune mise en quarantaine (gate_status=valid)" /></div> :
        <table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12 }}>
          <thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
            <th style={{ padding:'6px 8px' }}>run</th><th>événement</th><th>cellule</th><th>statut</th><th>portes</th>
          </tr></thead><tbody>
          {quar.map((q,i)=><tr key={i} style={{ borderBottom:'1px solid var(--hairline-faint)' }}>
            <td style={{ padding:'6px 8px' }}>{q.run}</td><td>#{q.event_id}</td><td>{q.profile}·{q.qdisc}·{q.cc}</td><td style={{ color:'#f4b400' }}>{q.gate_status}</td><td style={{ color:'#8b9099' }}>{(q.failed_gates ?? []).join(' ') || '—'}</td>
          </tr>)}
          </tbody></table>}
        <div className="mono" style={{ fontSize:10, color:'#767b84', marginTop:8 }}>source: quarantine.json · gate_status != valid{quarSum && quarSum.total_invalid != null ? ` · synthèse live : ${quarSum.total_invalid} invalid — G4 bas ${quarSum.g4_low}, G4 haut ${quarSum.g4_high}, sondes vides ${quarSum.empty_probes}, G3 ${quarSum.g3_implausible} (non-exclusif, G4 = régime)` : ''}</div>
      </div>
      <div className="card" style={{ border:'1px solid #26262a' }}>
        <div className="card-head">Recommandations — Traduction Matérielle</div>
        <p className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, lineHeight:'1.6', color:'#9aa3ad', marginBottom:12}}>
          Transposition du principe Linux prouvé en lab vers matériel DSI — sans réécrire l'infra. <span style={{color:'#5ad3e3'}}>Client-side observé, pas contrôleur réseau</span>.
        </p>
        <div style={{overflowX:'auto', marginBottom:12}}>
          <table style={{width:'100%', borderCollapse:'collapse', fontFamily:'JetBrains Mono', fontSize:11}}>
            <thead>
              <tr style={{textAlign:'left', color:'#8b9099', borderBottom:'1px solid #26262a'}}>
                <th style={{padding:'6px 8px', fontWeight:600}}>Principe Linux</th>
                <th style={{padding:'6px 8px', fontWeight:600}}>MikroTik</th>
                <th style={{padding:'6px 8px', fontWeight:600}}>ISP / mini-PC gateway</th>
              </tr>
            </thead>
            <tbody style={{color:'#f2f2f4'}}>
              <tr style={{borderBottom:'1px solid #1a1a1e'}}>
                <td style={{padding:'6px 8px'}}>pfifo_fast <span style={{color:'#767b84'}}>baseline FIFO</span></td>
                <td style={{padding:'6px 8px'}}>— ne pas reproduire</td>
                <td style={{padding:'6px 8px'}}>— ne pas reproduire</td>
              </tr>
              <tr style={{borderBottom:'1px solid #1a1a1e'}}>
                <td style={{padding:'6px 8px'}}>fq_codel <span style={{color:'#1fa348'}}>prouvé lab</span></td>
                <td style={{padding:'6px 8px'}}>Queue Tree PCQ / CAKE RouterOS v7+</td>
                <td style={{padding:'6px 8px'}}>mini-PC transparent bridge CAKE en amont CPE</td>
              </tr>
              <tr>
                <td style={{padding:'6px 8px'}}>cake <span style={{color:'#1fa348'}}>prouvé lab</span></td>
                <td style={{padding:'6px 8px'}}>CAKE RouterOS v7+ Queue Tree</td>
                <td style={{padding:'6px 8px'}}>mini-PC gateway CAKE transparent</td>
              </tr>
            </tbody>
          </table>
        </div>
        {asArray<any>(groups).filter((g:any)=>g.best).length>0 && (
          <div style={{marginBottom:12, display:'flex', flexDirection:'column', gap:6}}>
            <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.06em', textTransform:'uppercase', color:'#8b9099'}}>Traduction matérielle par profil (depuis Scan ★ best)</div>
            {asArray<any>(groups).filter((g:any)=>g.best).slice(0,4).map((g:any)=>(
              <div key={`${g.profile}-${g.qdisc}`} className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, padding:'6px 8px', background:'#070707', border:'1px solid #26262a', color:'#f2f2f4'}}>
                <span style={{color:'#5ad3e3'}}>{g.profile}</span> · {g.qdisc} → {g.hardware_recommendation || hardwareRecommendation(g.qdisc, g.profile)}
              </div>
            ))}
          </div>
        )}
        <div style={{display:'grid', gap:10, fontSize:12, lineHeight:'1.6'}}>
          <div style={{padding:'8px 10px', background:'rgba(90,211,227,0.06)', border:'1px solid #26262a'}}>
            <b style={{color:'#f2f2f4'}}>Kit de Diagnostic Portable DSI</b> <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#5ad3e3', border:'1px solid #26262a', padding:'1px 4px', marginLeft:6}}>5-min branch</span>
            <div style={{color:'#9aa3ad', marginTop:4}}>Branchement sans <code style={{color:'#f2f2f4'}}>tc</code> : laptop → Yas/Telma 4G → <code style={{color:'#f2f2f4'}}>1.1.1.1</code> ping 5 Hz + Small 4–32 K + bulk GRIB. Zéro install, client-side observé depuis les locaux — pas contrôleur réseau.</div>
          </div>
          <div style={{padding:'8px 10px', background:'rgba(244,180,0,0.06)', border:'1px solid #26262a'}}>
            <b style={{color:'#f2f2f4'}}>Pilote isolé</b> <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#f4b400', border:'1px solid #26262a', padding:'1px 4px', marginLeft:6}}>1 site → 1 département 4G</span>
            <div style={{color:'#9aa3ad', marginTop:4}}>Un seul site 4G à la fois, mesure avant/après. Si CAKE prouvé en lab réduit <code style={{color:'#1fa348'}}>small_p95 738 → 20 ms</code>, alors seulement généraliser. Pas de déploiement global sans pilote.</div>
          </div>
          <div style={{padding:'8px 10px', background:'#070707', border:'1px solid #26262a'}}>
            <b style={{color:'#f2f2f4'}}>Why Go: zero-dependency vs Flent</b>
            <div style={{color:'#9aa3ad', marginTop:4}}>Un seul binaire <code style={{color:'#f2f2f4'}}>cgo</code> posable sur laptop DSI — pas de Python, pas de netperf, pas de dépendance. Flent exige stack complète et lab ; <code style={{color:'#5ad3e3'}}>Go 1.25</code> embarque <code style={{color:'#f2f2f4'}}>embed dist</code> + csv/metrics/probe campagne en un fichier. Reproductible offline, <code style={{color:'#767b84'}}>manifest.json sha256 + quarantine.json</code> vérifiable.</div>
          </div>
        </div>
        <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#767b84', marginTop:8}}>
          Traduction matérielle : <b className="mono" style={{color:'#f2f2f4'}}>{(() => { try { const g = (window as any).__CGO_GROUPS ?? null; return g?.best?.profile ?? 'P2' } catch { return 'P2' } })()}</b> — routeur Linux/pfSense : tc direct · MikroTik : Queue Tree PCQ/CAKE (RouterOS v7+) · accès ISP : passerelle dédiée en amont
        </div>
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
            <div>
              <img src="/api/figures/small_p95.svg" alt="small p95" style={{width:'100%', border:'1px solid var(--hairline)'}} />
              <div className="mono muted" style={{fontSize:10, marginTop:4}}>small p95 par configuration · hash {(data as any).hash8 ?? '—'}</div>
            </div>
            <div>
              <img src="/api/figures/scatter.svg" alt="scatter" style={{width:'100%', border:'1px solid var(--hairline)'}} />
              <div className="mono muted" style={{fontSize:10, marginTop:4}}>compromis latence / débit · hash {(data as any).hash8 ?? '—'}</div>
            </div>
          </div>
        )}
      </div>
      <Provenance source="data/runs/*/manifest.json" state="live" />
      <div className="card">
        <div className="card-head">Replay {replayRunning && <span className="mono" style={{color:'var(--t-live)', marginLeft:8}}>● en cours</span>}</div>
        {replayRunning ? (
          <div style={{display:'flex', gap:8, alignItems:'center'}}>
            <span className="mono muted">replay actif — voir le Tableau live</span>
            <button className="btn" onClick={()=>stopReplay()} style={{marginLeft:'auto'}}>Arrêter</button>
            <button className="btn btn-primary" onClick={()=>setPanel('live')}>Voir le Tableau live</button>
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
