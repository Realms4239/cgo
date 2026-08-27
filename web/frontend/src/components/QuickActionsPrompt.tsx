import { useEffect, useRef, useState } from 'react'
import { useUIStore } from '../store/ui'
import { animatePromptEnter, animatePromptExit } from '../lib/anime'
import { PromptProgressLine } from './PromptProgressLine'

export default function QuickActionsPrompt(){
  const live = useUIStore((s:any)=>s.live)
  const [open,setOpen]=useState(true)
  const [isHoverPaused,setIsHoverPaused]=useState(false)
  const ref=useRef<HTMLDivElement>(null)
  const reappearRef=useRef<ReturnType<typeof setTimeout>|null>(null)
  const startRef=useRef<number>(0)
  const remainRef=useRef<number>(6000)
  const timeoutRef=useRef<ReturnType<typeof setTimeout>|null>(null)

  useEffect(()=>{ if(!ref.current) return; if(open) animatePromptEnter(ref.current) },[open])

  // reset drain when re-opened or event advances
  useEffect(()=>{ if(open) remainRef.current=6000 },[open, live?.event_id, live?.phase])

  // ponytail: track remaining so CSS pause (animationPlayState) and JS timeout stay in sync
  useEffect(()=>{
    if(!open || !live?.running) return
    if(isHoverPaused){
      if(timeoutRef.current) clearTimeout(timeoutRef.current)
      remainRef.current = Math.max(0, remainRef.current - (Date.now() - startRef.current))
      return
    }
    startRef.current = Date.now()
    timeoutRef.current = setTimeout(()=>{ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)); else setOpen(false) }, remainRef.current)
    return()=>{ if(timeoutRef.current) clearTimeout(timeoutRef.current) }
  },[open, isHoverPaused, live?.running])

  useEffect(()=>{
    if(open) return
    reappearRef.current=setTimeout(()=>setOpen(true),8000)
    return()=>{ if(reappearRef.current) clearTimeout(reappearRef.current) }
  },[open])

  useEffect(()=>{
    if(!open || !live?.running) return
    const onKey=(e:KeyboardEvent)=>{ if(e.key==='Escape'){ if(ref.current) animatePromptExit(ref.current).then(()=>setOpen(false)); else setOpen(false) } }
    document.addEventListener('keydown', onKey)
    return()=>document.removeEventListener('keydown', onKey)
  },[open, live?.running])

  if (!live?.running) return null
  if (!open) return null
  const eventId = live.event_id ?? 1
  const phase = live.phase ?? 'charge'
  return (
    <div
      ref={ref}
      role="status"
      aria-live="polite"
      aria-label="Progression campagne"
      onMouseEnter={()=>setIsHoverPaused(true)}
      onMouseLeave={()=>setIsHoverPaused(false)}
      style={{
        position:'fixed',
        bottom:40,
        left:'50%',
        transform:'translateX(-50%)',
        display:'flex',
        alignItems:'center',
        gap:12,
        padding:'10px 16px 11px',
        background:'rgba(16,16,18,0.96)',
        border:'1px solid rgba(255,255,255,0.08)',
        backdropFilter:'blur(12px)',
        borderRadius:0,
        zIndex:400,
        boxShadow:'0 8px 32px rgba(0,0,0,0.45), 0 0 0 1px rgba(255,255,255,0.02) inset',
        overflow:'hidden',
        minWidth:320,
        justifyContent:'center',
      }}
    >
      <span className="mono" style={{fontFamily:'JetBrains Mono, ui-monospace, monospace', fontSize:11, fontWeight:700, color:'#f2f2f4', letterSpacing:'0.06em', whiteSpace:'nowrap'}}>Event {eventId}/6 — {phase} 4/10s</span>
      <span className="mono" style={{fontFamily:'JetBrains Mono, ui-monospace, monospace', fontSize:10, color:'#9aa3ad', padding:'2px 6px', border:'1px solid #26262a', background:'rgba(244,180,0,0.10)', lineHeight:1}}>6s</span>
      <PromptProgressLine paused={isHoverPaused} />
    </div>
  )
}
