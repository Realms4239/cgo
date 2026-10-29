import { useEffect, useRef } from 'react'
import { useUIStore, PANELS } from '../store/ui'
import { StatusPip } from './ui/StatusPip'
import { NavIcon } from './NavIcon'
import { explain } from '../lib/explain'
import { live as liveRing } from '../lib/live'
import { paletteOnce } from '../lib/palette'

export default function Rail(){
  const panel=useUIStore(s=>s.panel), setPanel=useUIStore(s=>s.setPanel)
  const pinned=useUIStore(s=>s.railPinned), setPinned=useUIStore(s=>s.setRailPinned)
  // sélecteurs par champ — le rail ne re-rend que si SA donnée change, pas à chaque frame live
  const phase=useUIStore(s=>s.live?.phase ?? '')
  const eventId=useUIStore(s=>s.live?.event_id ?? null)
  const gates=useUIStore(s=>s.live?.gates ?? null)
  const sseStatus=useUIStore(s=>s.sseStatus)
  const liveOn=useUIStore(s=>!!s.live)
  // une valeur live suffit à redessiner : l'anneau est la source, pas un motif gelé
  const sample=useUIStore(s=>s.live?.small_p95_ms ?? null)
  const ref=useRef<HTMLElement>(null)
  const canvasRef=useRef<HTMLCanvasElement>(null)
  useEffect(()=>{
    const c=canvasRef.current
    if(!c) return
    const ctx=c.getContext('2d')
    if(!ctx) return
    ctx.clearRect(0,0,32,12)
    const vals=liveRing.small.slice(-16).map(([,v])=>v).filter(v=>Number.isFinite(v)&&v>0)
    if(vals.length<2) return
    const min=Math.min(...vals), max=Math.max(...vals), span=Math.max(1,max-min)
    ctx.strokeStyle=paletteOnce().live
    ctx.globalAlpha=0.6
    ctx.lineWidth=1
    ctx.beginPath()
    vals.forEach((v,i)=>{
      const x=(i/(vals.length-1))*32
      const y=10-((v-min)/span)*8
      if(i===0) ctx.moveTo(x,y); else ctx.lineTo(x,y)
    })
    ctx.stroke()
    ctx.globalAlpha=1
  },[pinned,sample])
  const lastSmall=liveOn?liveRing.small.at(-1)?.[1]:undefined
  return <aside ref={ref} className={'rail '+(pinned?'pinned':'')} data-pinned={pinned?'1':'0'} aria-label="Navigation" style={{'--rail-w': pinned?232:56} as any}>
    {PANELS.map(p=><button key={p.id} className={'nav-btn'+(panel===p.id?' on':'')} data-panel={p.id} onClick={()=>setPanel(p.id)} title={p.label} aria-label={p.label} aria-pressed={panel===p.id}><span className="nav-icon" aria-hidden="true"><NavIcon id={p.id} /></span>{pinned?<span className="nav-lbl">{p.label}</span>:<span className="sr-only">{p.label}</span>}{pinned&&<span className="nav-key">{p.key}</span>}</button>)}
    <button onClick={()=>setPinned(!pinned)} aria-label="Épingler" className="nav-btn rail-pin">{pinned?'◀':'▶'}</button>
    <canvas ref={canvasRef} className="miniSparkline" width={32} height={12} aria-hidden="true" />
    {pinned && liveOn && lastSmall != null && (
      <div className="live-peek" aria-hidden="true">
        <span className="mono rail-peek">∼ {lastSmall.toFixed(0)} ms</span>
      </div>
    )}
    <div className="gates" role="img" aria-label="Portes G0 à G7">
      {(gates ?? Array(8).fill(null)).map((g: boolean|null, i:number) => {
        const titles=['G0 cible','G1 bulk','G2 sondes','G3 latence','G4 débit','G5 doublon','G6 baseline','G7 CPU']
        const state=g===null?'idle':g?'ok':'err'
        const tip=`G${i} ${titles[i].split(" ").slice(1).join(" ")}: ${g===null?"—":g?"PASS":"FAIL"} — ${explain("G"+i)}`
        if(pinned){
          return <span key={i} data-gate={i} className="rail-gate"><StatusPip state={state as any} label={`G${i}`} title={tip} /></span>
        }
        return <span key={i} className={'gate '+(g===null?'':g?'ok':'fail')} data-gate={i} title={tip} />
      })}
    </div>
    {pinned && <div className="side-status">
      <div><span>phase</span><b className="mono">{phase || '—'}</b></div>
      <div><span>événement</span><b className="mono">{eventId ? '#'+eventId : '—'}</b></div>
      <div><span>SSE</span><b className="mono">{sseStatus}</b></div>
    </div>}
    {pinned && <div className="rail-foot mono">Meteolink</div>}
  </aside>
}
