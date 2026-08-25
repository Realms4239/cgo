import { useEffect, useRef } from 'react'
import { useUIStore, PANELS } from '../store/ui'
import { animateRail } from '../lib/anime'

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
  return <aside ref={ref} className={'rail '+(pinned?'pinned':'')} style={{width: pinned?232:56}} aria-label="Navigation">
    {PANELS.map(p=><button key={p.id} className={'nav-btn'+(panel===p.id?' on':'')} data-panel={p.id} onClick={()=>setPanel(p.id)}><span className="nav-lbl">{p.label}</span>{pinned&&<span className="nav-key">{p.key}</span>}</button>)}
    <button onClick={()=>setPinned(!pinned)} aria-label="Épingler" className="nav-btn" style={{marginTop:8, borderTop:'1px solid var(--hairline)', justifyContent:'center'}}>{pinned?'◀':'▶'}</button>
    <canvas ref={canvasRef} className="miniSparkline" width={32} height={12} aria-hidden="true" />
    <div className="gates" role="img" aria-label="Portes G0 à G7">
      {(live?.gates ?? Array(8).fill(null)).map((g: boolean|null, i:number) => {
        const titles=['G0 cible','G1 bulk','G2 sondes','G3 latence','G4 débit','G5 doublon','G6 baseline','G7 CPU']
        return <span key={i} className={'gate '+(g===null?'':g?'ok':'fail')} data-gate={i} title={titles[i]+': '+(g===null?'—':g?'PASS':'FAIL')} />
      })}
    </div>
    {pinned && <div className="side-status">
      <div><span>phase</span><b className="mono">{live?.phase || '—'}</b></div>
      <div><span>événement</span><b className="mono">{live?.event_id ? '#'+live.event_id : '—'}</b></div>
      <div><span>SSE</span><b className="mono">{sseStatus}</b></div>
    </div>}
  </aside>
}
