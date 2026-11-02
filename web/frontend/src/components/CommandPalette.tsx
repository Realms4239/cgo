import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useModalFocus } from '../lib/modalFocus'
import { MODES, useUIStore } from '../store/ui'
import { createTimeline, stagger } from 'animejs'
import { prefersReducedMotion, EASE } from '../lib/anime'

// Cmd-K palette — une entrée par destination, pas deux chemins vers le même
// écran. Les QUICK duplicaient audit/banc/preuves sous une formulation verbe,
// ce qui faisait sept lignes dont trois menaient exactement là où les quatre
// premières menaient déjà — la navigation dupliquée que la console supprime.
// Les verbes restent trouvable : ils vivent dans un alias de recherche caché,
// pas dans une seconde ligne cliquable.
const ALIAS: Record<string, string> = {
  dashboard: 'tableau de bord etat du lien reel derniere mesure',
  audit: 'auditer auditer le lien reel lancer un audit mesure du lien',
  bench: "banc d'essai ouvrir le banc campagne matrix simuler simulation",
  evidence: 'preuves rapports voir les preuves et rapports corpus manifestes replay figures',
}

// fold — uniquement pour la RECHERCHE : « auditer le lien reel » et
// « auditer le lien réel » doivent trouver la même ligne, et un opérateur au
// clavier AZERTY ne doit pas décrire les accents pour naviguer. Les libellés
// affichés ne passent jamais par là : ce texte-ci est un index, pas un rendu.
const fold = (s: string): string => s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()

// paletteItems — la liste, en pur : une entrée par destination, filtrée sur le
// libellé et ses alias. Séparée du composant parce que le composant se rend dans
// un portail et que « pas de doublon » doit être vérifié, pas espéré.
export function paletteItems(query: string): { label: string; mode: string }[] {
  const q = fold(query.trim())
  return MODES.map((m) => ({
    label: m.label as string,
    mode: m.id as string,
    search: fold(`${m.label} ${ALIAS[m.id] ?? ''}`),
  }))
    .filter((i) => i.search.includes(q))
    .map(({ label, mode }) => ({ label, mode }))
}

export default function CommandPalette({ open, onClose }: { open: boolean; onClose: ()=>void }){
  const [q,setQ]=useState('')
  const inputRef=useRef<HTMLInputElement>(null)
  const listRef=useRef<HTMLUListElement>(null)
  const setMode=useUIStore((s:any)=>s.setMode)

  const filtered = paletteItems(q)

  const panelRef=useRef<HTMLDivElement>(null)
  useModalFocus(open, onClose, panelRef)
  useEffect(()=>{ if(open) setQ('') },[open])

  useEffect(()=>{
    if(!open || !listRef.current) return
    if(prefersReducedMotion()) return
    const els=listRef.current.querySelectorAll('li')
    if(!els.length) return
    const tl=createTimeline()
    tl.add(els, { translateY:[8,0], opacity:[0,1], delay: stagger(20, {start:60}), duration:300, ease: EASE} as any, 0)
  },[open, q])


  if(!open) return null
  // Portail hors de #shell : la palette rend #shell inerte pour bloquer
  // l'arrière-plan ; si elle vivait dedans, elle s'inertait elle-même et
  // devenait infocusable et inclickable (constaté : activeElement = body).
  return createPortal(<div id="palette" role="dialog" aria-modal="true" aria-label="Palette de commandes" onClick={onClose} className="palette-backdrop">
    <div ref={panelRef} onClick={e=>e.stopPropagation()} className="palette-panel">
      <div className="palette-head">
        <span className="palette-key">Cmd-K</span>
        <input ref={inputRef} value={q} onChange={e=>setQ(e.target.value)} placeholder="Rechercher une destination…" className="palette-input" />
        <button onClick={onClose} aria-label="Fermer" className="palette-close">Esc</button>
      </div>
      <ul ref={listRef} className="palette-list">
        {filtered.map(it=> <li key={`${it.mode}-${it.label}`} role="button" tabIndex={0} onClick={()=>{ setMode(it.mode); onClose() }} onKeyDown={e=>{ if(e.key==='Enter'||e.key===' '){ e.preventDefault(); setMode(it.mode); onClose() } }} className="palette-item"><span>{it.label}</span><span className="palette-item-panel">{it.mode}</span></li>)}
        {filtered.length===0 && <li className="palette-empty">Aucun résultat</li>}
      </ul>
    </div>
  </div>, document.body)
}
