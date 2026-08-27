import { useEffect, useRef } from 'react'
import { useUIStore, PANELS } from '../store/ui'
import { animateRail } from '../lib/anime'
import { StatusPip } from './ui/StatusPip'

export default function Rail(){
  const panel=useUIStore(s=>s.panel), setPanel=useUIStore(s=>s.setPanel)
  const pinned=useUIStore(s=>s.railPinned), setPinned=useUIStore(s=>s.setRailPinned)
  const live=useUIStore(s=>s.live)
  const sseStatus=useUIStore(s=>s.sseStatus)
  const ref=useRef<HTMLElement>(null)
  const canvasRef=useRef<HTMLCanvasElement>(null)
  useEffect(()=>{ if(ref.current) animateRail(ref.current, pinned) },[pinned])
  useEffect(()=>{
    const c=canvasRef.current
    if(!c) return
    const ctx=c.getContext('2d')
    if(!ctx) return
    ctx.clearRect(0,0,32,12)
    // ponytail: static faint cyan sparkline, live data if throughput matters
    ctx.strokeStyle='#5ad3e3'
    ctx.globalAlpha=0.6
    ctx.lineWidth=1
    ctx.beginPath()
    const pts=[4,8,5,6,9,3,12,7,16,5,20,9,24,4,28,6]
    for(let i=0;i<pts.length;i+=2){ const x=pts[i], y=pts[i+1]; if(i===0) ctx.moveTo(x,y); else ctx.lineTo(x,y) }
    ctx.stroke()
  },[pinned, live])
  const ICONS:Record<string,string>={campagne:'◉', live:'∼', resultats:'▦', integrite:'⬢'}
  return <aside ref={ref} className={'rail '+(pinned?'pinned':'')} data-pinned={pinned?'1':'0'} aria-label="Navigation">
    {PANELS.map(p=><button key={p.id} className={'nav-btn'+(panel===p.id?' on':'')} data-panel={p.id} onClick={()=>setPanel(p.id)} title={p.label} aria-label={p.label}><span className="nav-icon" aria-hidden="true" style={{width:16,height:16,display:'inline-flex',alignItems:'center',justifyContent:'center',fontSize:14,lineHeight:1}}>{ICONS[p.id]||'•'}</span>{pinned?<span className="nav-lbl">{p.label}</span>:<span className="sr-only">{p.label}</span>}{pinned&&<span className="nav-key">{p.key}</span>}</button>)}
    <button onClick={()=>setPinned(!pinned)} aria-label="Épingler" className="nav-btn" style={{marginTop:8, borderTop:'1px solid var(--hairline)', justifyContent:'center'}}>{pinned?'◀':'▶'}</button>
    <canvas ref={canvasRef} className="miniSparkline" width={32} height={12} aria-hidden="true" style={{opacity:0.6}} />
    {pinned && live && (
      <div className="live-peek" aria-hidden="true" style={{display:'flex',alignItems:'end',gap:1,height:12,margin:'4px 14px',opacity:0.7}}>
        {(live?.gates ?? []).length ? null : null}
        {/* sparkline peek 32×12 cyan 0.6 when extended — ponytail: static preview, live data if needed */}
        <span className="mono" style={{fontSize:10,color:'var(--t-live)'}}>∼ live</span>
      </div>
    )}
    <div className="gates" role="img" aria-label="Portes G0 à G7" style={{display:'flex',gap:4,margin:'6px 8px',flexWrap:'wrap'}}>
      {(live?.gates ?? Array(8).fill(null)).map((g: boolean|null, i:number) => {
        const titles=['G0 cible','G1 bulk','G2 sondes','G3 latence','G4 débit','G5 doublon','G6 baseline','G7 CPU']
        const state=g===null?'idle':g?'ok':'err'
        if(pinned){
          return <span key={i} data-gate={i} title={titles[i]+': '+(g===null?'—':g?'PASS':'FAIL')} style={{display:'inline-flex'}}><StatusPip state={state as any} label={`G${i}`} title={titles[i]+': '+(g===null?'—':g?'PASS':'FAIL')} /></span>
        }
        return <span key={i} className={'gate '+(g===null?'':g?'ok':'fail')} data-gate={i} title={titles[i]+': '+(g===null?'—':g?'PASS':'FAIL')} />
      })}
    </div>
    {pinned && <div className="side-status">
      <div><span>phase</span><b className="mono">{live?.phase || '—'}</b></div>
      <div><span>événement</span><b className="mono">{live?.event_id ? '#'+live.event_id : '—'}</b></div>
      <div><span>SSE</span><b className="mono">{sseStatus}</b></div>
    </div>}
    {pinned && <div className="rail-foot mono" style={{marginTop:'auto',padding:'8px 14px',borderTop:'1px solid var(--hairline)',fontSize:10,color:'var(--text-faint)',letterSpacing:'0.06em',textTransform:'uppercase' as const}}>Meteolink • LIEN</div>}
  </aside>
}
