import { useEffect, useRef, useState } from 'react'
import { useUIStore } from '../store/ui'
import { animatePromptEnter, animatePromptExit } from '../lib/anime'

export default function QuickActionsPrompt(){
  const live=useUIStore((s:any)=>s.live), setPanel=useUIStore((s:any)=>s.setPanel)
  const [open,setOpen]=useState(true)
  const ref=useRef<HTMLDivElement>(null)

  useEffect(()=>{ if(!ref.current) return; if(open) animatePromptEnter(ref.current) },[open])

  useEffect(()=>{
    if(!open) return
    let id = setTimeout(()=>{ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)); else setOpen(false) },6000)
    const el=ref.current
    const onEnter=()=>clearTimeout(id)
    const onLeave=()=>{
      clearTimeout(id)
      id=setTimeout(()=>{ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)); else setOpen(false) },6000)
    }
    el?.addEventListener('mouseenter',onEnter)
    el?.addEventListener('mouseleave',onLeave)
    return()=>{ clearTimeout(id); el?.removeEventListener('mouseenter',onEnter); el?.removeEventListener('mouseleave',onLeave) }
  },[open])

  useEffect(()=>{
    // ponytail: idle 8s interval not rAF loop
    const id=setInterval(()=>{ if(!open) setOpen(true) },8000)
    return()=>clearInterval(id)
  },[open])

  const actions = !live ? [{label:'Démarrer',panel:'campagne'}, {label:'Audit',panel:'campagne'}] : live.running ? [{label:'Temps réel',panel:'live'}] : [{label:'Résultats',panel:'resultats'}, {label:'Rejouer',panel:'integrite'}]
  if(!open) return null
  return <div ref={ref} role="dialog" aria-modal="true" aria-label="Actions rapides" style={{position:'fixed', bottom:24, left:'50%', transform:'translateX(-50%)', display:'flex', gap:8, padding:'10px 14px', background:'rgba(16,16,18,0.92)', border:'1px solid rgba(255,255,255,0.08)', backdropFilter:'blur(12px)', borderRadius:10, zIndex:400, boxShadow:'0 8px 32px rgba(0,0,0,0.4)'}}>
    <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#9aa3ad', alignSelf:'center'}}>Actions rapides</span>
    {actions.map(a=><button key={a.label} onClick={()=>setPanel(a.panel as any)} style={{padding:'6px 12px', background:'#1c69d4', color:'#fff', border:'none', borderRadius:6, fontFamily:'Inter var', fontSize:12, cursor:'pointer'}}>{a.label}</button>)}
    <button aria-label="Fermer" onClick={()=>{ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)); else setOpen(false) }} style={{padding:'6px 8px', background:'transparent', border:'1px solid #26262a', borderRadius:6, color:'#9aa3ad', fontSize:11, cursor:'pointer'}}>Esc</button>
  </div>
}
