import { useEffect, useRef, useState } from 'react'
import { useUIStore } from '../store/ui'
import { animatePromptEnter, animatePromptExit } from '../lib/anime'

// Popup de progression — un seul but : dire ce qui est testé, où l'on en est,
// combien de temps dure la phase. Re-apparaît à chaque nouvel event;
// Esc la referme jusqu'au prochain event.
export default function QuickActionsPrompt() {
  const live = useUIStore((s: any) => s.live)
  const [open, setOpen] = useState(true)
  const [closedForEvent, setClosedForEvent] = useState<number>(-1)
  const [elapsed, setElapsed] = useState(0)
  // réduite = pastille compacte (ne couvre plus les cartes/lignes) ;
  // sur étroit (< 700 px) on démarre réduit — le plein couvre les cartes
  // (le bas de main réserve 64 px : la pastille y tient, le plein non)
  const narrow = () => typeof window !== 'undefined' && window.innerWidth < 700
  const [min, setMin] = useState(() => narrow())
  const ref = useRef<HTMLDivElement>(null)

  const eventId = live?.event_id ?? 0
  const total = live?.total_events ?? 0
  const phase = live?.phase ?? ''
  // phase_total_s voyage dans la frame SSE (campagne.go:128) — absent = elapsed seul
  const phaseTotal = live?.phase_total_s ?? 0
  const running = !!live?.running
  // retenue : l'opérateur qui a réduit/fermé une fois ne veut plus du plein
  // écran — les events suivants rouvrent en pastille, jamais en plein
  const shrankOnce = useRef(false)

  // re-pop à chaque nouvel event — en pastille si l'opérateur a déjà réduit
  // (ou si l'écran est étroit : le plein y couvre les cartes)
  useEffect(() => {
    if (!running) return
    if (closedForEvent === eventId) return
    setOpen(true)
    setMin(shrankOnce.current || narrow())
  }, [eventId, running, closedForEvent])

  useEffect(() => { if (!ref.current) return; if (open) animatePromptEnter(ref.current) }, [open])

  // Esc referme jusqu'au prochain event
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') close(true) }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, eventId])

  // countdown réel — phaseSince ponté de l'anneau vers le store (sse.ts) ;
  // calcul immédiat aussi (pas d'attente du premier tick à 0 s)
  const readElapsed = () => {
    const since = (live as any)?.phaseSince?.[phase]
    return since ? Math.max(0, Math.round((Date.now() - since) / 1000)) : 0
  }
  useEffect(() => {
    if (!open || !running) return
    setElapsed(readElapsed())
    const t = window.setInterval(() => setElapsed(readElapsed()), 1000)
    return () => window.clearInterval(t)
  }, [open, running, phase, live])

  const close = (shrank = false) => {
    if (shrank) shrankOnce.current = true
    setClosedForEvent(eventId)
    if (ref.current) animatePromptExit(ref.current).then(() => setOpen(false))
    else setOpen(false)
  }

  if (!open || !running) return null

  const evt = total > 0 ? `Event ${eventId}/${total}` : `Event ${eventId}`
  const clock = phaseTotal > 0 ? `${elapsed}/${phaseTotal} s` : `${elapsed} s`
  const label = `${evt} — ${live?.profile}/${live?.qdisc}/${live?.cc} rep ${live?.repetition} — ${phase} ${clock}`

  // pastille réduite : rappel d'état seul, rouvre au clic, Esc la referme
  // comme la pleine (cohérence Échap partout)
  if (min) {
    return (
      <div
        ref={ref}
        role="status"
        aria-label={label}
        onClick={() => setMin(false)}
        title="cliquer pour rouvrir"
        style={{
          position: 'fixed',
          bottom: 44,
          left: '50%',
          transform: 'translateX(-50%)',
          padding: '6px 12px',
          background: 'rgba(16,16,18,0.92)',
          border: '1px solid rgba(255,255,255,0.08)',
          zIndex: 400,
          cursor: 'pointer',
          maxWidth: 'calc(100vw - 24px)',
        }}
      >
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#f4b400', letterSpacing: '0.06em', whiteSpace: 'nowrap' }}>● {evt} — {phase}</span>
      </div>
    )
  }

  return (
    <div
      ref={ref}
      role="status"
      aria-label={label}
      onMouseEnter={undefined}
      style={{
        position: 'fixed',
        bottom: 44,
        left: '50%',
        transform: 'translateX(-50%)',
        display: 'flex',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: 8,
        padding: '10px 14px 11px',
        background: 'rgba(16,16,18,0.92)',
        border: '1px solid rgba(255,255,255,0.08)',
        backdropFilter: 'blur(12px)',
        zIndex: 400,
        boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
        overflow: 'hidden',
        maxWidth: 'calc(100vw - 24px)',
        minWidth: 0,
      }}
    >
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', letterSpacing: '0.02em', overflowWrap: 'anywhere', minWidth: 0 }}>{label}</span>
      <button aria-label="Passer la cellule" title="passer la cellule courante (reprise possible)" onClick={async () => { await fetch('/api/run/skip', { method: 'POST' }).catch(() => {}); close() }} style={{ padding: '6px 12px', background: '#161618', color: '#f4b400', border: '1px solid #26262a', fontFamily: 'JetBrains Mono', fontSize: 11, letterSpacing: '0.06em', textTransform: 'uppercase', cursor: 'pointer' }}>Passer</button>
      <button aria-label="Réduire" title="réduire en pastille (rouvre au clic)" onClick={() => { shrankOnce.current = true; setMin(true) }} style={{ padding: '6px 8px', background: 'transparent', border: '1px solid #26262a', color: '#9aa3ad', fontSize: 11, cursor: 'pointer' }}>−</button>
      <button aria-label="Fermer" onClick={() => close(true)} style={{ padding: '6px 8px', background: 'transparent', border: '1px solid #26262a', color: '#9aa3ad', fontSize: 11, cursor: 'pointer' }}>Esc</button>
    </div>
  )
}
