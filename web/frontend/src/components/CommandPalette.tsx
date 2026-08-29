import { useEffect, useRef, useState } from 'react'
import { PANELS, useUIStore } from '../store/ui'
import { createTimeline, stagger } from 'animejs'
import { prefersReducedMotion } from '../lib/anime'

// Cmd-K palette — filters 4 panels + 3 quick actions
const QUICK = [
  { label: 'Démarrer campagne', panel: 'campagne' as const },
  { label: 'Voir le Tableau live', panel: 'live' as const },
  { label: 'Voir résultats', panel: 'resultats' as const },
]

export default function CommandPalette({ open, onClose }: { open: boolean; onClose: ()=>void }){
  const [q,setQ]=useState('')
  const inputRef=useRef<HTMLInputElement>(null)
  const listRef=useRef<HTMLUListElement>(null)
  const prevFocusRef=useRef<HTMLElement|null>(null)
  const setPanel=useUIStore((s:any)=>s.setPanel)

  const items = [...PANELS.map(p=>({label:p.label, panel:p.id})), ...QUICK]
  const filtered = items.filter(i=> i.label.toLowerCase().includes(q.toLowerCase()))

  useEffect(()=>{ if(open) { setQ(''); setTimeout(()=>inputRef.current?.focus(), 0) } },[open])

  useEffect(()=>{
    if(!open || !listRef.current) return
    if(prefersReducedMotion()) return
    const els=listRef.current.querySelectorAll('li')
    if(!els.length) return
    const tl=createTimeline()
    tl.add(els, { translateY:[8,0], opacity:[0,1], delay: stagger(20, {start:60}), duration:300, ease:'cubicBezier(0.16,1,0.3,1)'} as any, 0)
  },[open, q])

  useEffect(()=>{
    if(!open) return
    prevFocusRef.current=document.activeElement as HTMLElement
    const modal=document.getElementById('shell') as HTMLElement | null
    if(modal) (modal as any).inert=true
    const first=document.querySelector('#palette input') as HTMLElement | null
    first?.focus()
    const onKey=(e:KeyboardEvent)=>{
      if(e.key==='Tab'){
        const els=[...document.querySelectorAll('#palette [tabindex], #palette button, #palette input')] as HTMLElement[]
        if(!els.length) return
        const idx=els.indexOf(document.activeElement as HTMLElement)
        if(e.shiftKey){ if(idx===0){ els[els.length-1].focus(); e.preventDefault() } } else { if(idx===els.length-1){ els[0].focus(); e.preventDefault() } }
      }
      if(e.key==='Escape') onClose()
    }
    document.addEventListener('keydown',onKey)
    return()=>{
      document.removeEventListener('keydown',onKey)
      if(modal) (modal as any).inert=false
      ;(document.activeElement as HTMLElement)?.blur()
      prevFocusRef.current?.focus()
    }
  },[open, onClose])

  if(!open) return null
  return <div id="palette" role="dialog" aria-modal="true" aria-label="Palette de commandes" onClick={onClose} style={{position:'fixed', inset:0, background:'rgba(0,0,0,0.45)', display:'flex', alignItems:'flex-start', justifyContent:'center', paddingTop:120, zIndex:600}}>
    <div onClick={e=>e.stopPropagation()} style={{width:480, maxWidth:'90vw', background:'#101012', border:'1px solid #26262a', borderRadius:12, overflow:'hidden', boxShadow:'0 16px 48px rgba(0,0,0,0.6)'}}>
      <div style={{display:'flex', alignItems:'center', gap:8, padding:'10px 12px', borderBottom:'1px solid #1e1e22'}}>
        <span style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#9aa3ad', border:'1px solid #26262a', padding:'2px 6px', borderRadius:4}}>Cmd-K</span>
        <input ref={inputRef} value={q} onChange={e=>setQ(e.target.value)} placeholder="Rechercher panneau ou action…" style={{flex:1, background:'transparent', border:'none', outline:'none', color:'#f2f2f4', fontFamily:'Inter var', fontSize:13}} />
        <button onClick={onClose} aria-label="Fermer" style={{background:'transparent', border:'none', color:'#9aa3ad', cursor:'pointer'}}>Esc</button>
      </div>
      <ul ref={listRef} style={{listStyle:'none', margin:0, padding:8, maxHeight:280, overflowY:'auto'}}>
        {filtered.map(it=> <li key={it.label} onClick={()=>{ setPanel(it.panel); onClose() }} style={{padding:'8px 10px', borderRadius:6, cursor:'pointer', fontFamily:'Inter var', fontSize:13, color:'#d8dde6', display:'flex', justifyContent:'space-between'}}><span>{it.label}</span><span style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#767b84'}}>{it.panel}</span></li>)}
        {filtered.length===0 && <li style={{padding:'12px', color:'#767b84', fontSize:12}}>Aucun résultat</li>}
      </ul>
    </div>
  </div>
}
