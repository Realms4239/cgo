import { useEffect, useRef, useState } from 'react'
import { PANELS, useUIStore } from '../store/ui'
import { createTimeline, stagger } from 'animejs'
import { prefersReducedMotion, EASE } from '../lib/anime'

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
    tl.add(els, { translateY:[8,0], opacity:[0,1], delay: stagger(20, {start:60}), duration:300, ease: EASE} as any, 0)
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
  return <div id="palette" role="dialog" aria-modal="true" aria-label="Palette de commandes" onClick={onClose} className="palette-backdrop">
    <div onClick={e=>e.stopPropagation()} className="palette-panel">
      <div className="palette-head">
        <span className="palette-key">Cmd-K</span>
        <input ref={inputRef} value={q} onChange={e=>setQ(e.target.value)} placeholder="Rechercher panneau ou action…" className="palette-input" />
        <button onClick={onClose} aria-label="Fermer" className="palette-close">Esc</button>
      </div>
      <ul ref={listRef} className="palette-list">
        {filtered.map(it=> <li key={it.label} role="button" tabIndex={0} onClick={()=>{ setPanel(it.panel); onClose() }} onKeyDown={e=>{ if(e.key==='Enter'||e.key===' '){ e.preventDefault(); setPanel(it.panel); onClose() } }} className="palette-item"><span>{it.label}</span><span className="palette-item-panel">{it.panel}</span></li>)}
        {filtered.length===0 && <li className="palette-empty">Aucun résultat</li>}
      </ul>
    </div>
  </div>
}
