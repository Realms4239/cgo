import { useEffect, useRef, useState } from 'react'
import { createTimeline, stagger } from 'animejs'
import { prefersReducedMotion } from '../lib/anime'

const STEPS = [
  { k: 'Audit', desc: 'Auditez votre lien avant de lancer' },
  { k: 'Démarrer', desc: 'Lancez la campagne depuis Campagne' },
  { k: 'Live', desc: 'Suivez en Temps réel' },
]

export default function OnboardingNudge(){
  const [open,setOpen]=useState(false)
  const ref=useRef<HTMLDivElement>(null)

  useEffect(()=>{
    try{ if(localStorage.getItem('nudge_seen')==='2026-08-26') return }catch{}
    setOpen(true)
    const id=setTimeout(()=>{ setOpen(false); try{ localStorage.setItem('nudge_seen','2026-08-26') }catch{} },8000)
    return()=>clearTimeout(id)
  },[])

  useEffect(()=>{
    if(!open || !ref.current) return
    if(prefersReducedMotion()) return
    const tl=createTimeline()
    tl.add(ref.current.querySelectorAll('.nudge-step'), { translateY:[8,0], opacity:[0,1], delay: stagger(40, {start:60}), duration:400, ease:'cubicBezier(0.16,1,0.3,1)'} as any, 0)
  },[open])

  if(!open) return null
  return <div ref={ref} role="dialog" aria-label="Bienvenue" style={{position:'static', background:'rgba(16,16,18,0.96)', border:'1px solid rgba(255,255,255,0.08)', borderRadius:0, padding:'10px 14px', display:'flex', gap:12, alignItems:'center', flexWrap:'wrap', width:'100%', boxSizing:'border-box'}}>
    {STEPS.map(s=> <div key={s.k} className="nudge-step" style={{display:'flex', flexDirection:'column', gap:2, minWidth:90, borderRight:'1px solid #1e1e22', paddingRight:12}}>
      <span style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.08em', textTransform:'uppercase', color:'#5ad3e3'}}>{s.k}</span>
      <span style={{fontFamily:'Inter var', fontSize:11, color:'#9aa3ad'}}>{s.desc}</span>
    </div>)}
    <button onClick={()=>{ setOpen(false); try{ localStorage.setItem('nudge_seen','2026-08-26') }catch{} }} aria-label="Fermer" style={{marginLeft:4, padding:'6px 8px', background:'transparent', border:'1px solid #26262a', borderRadius:0, color:'#9aa3ad', fontSize:11, cursor:'pointer'}}>Fermer</button>
  </div>
}
