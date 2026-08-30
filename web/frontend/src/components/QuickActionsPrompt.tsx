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
  const ref = useRef<HTMLDivElement>(null)

  const eventId = live?.event_id ?? 0
  const total = live?.total_events ?? 0
  const phase = live?.phase ?? ''
  const phaseTotal = live?.phase_total_s ?? 0
  const running = !!live?.running

  // re-pop à chaque nouvel event — pas de re-pop intempestif en cours d'event
  useEffect(() => {
    if (!running) return
    if (closedForEvent === eventId) return
    setOpen(true)
  }, [eventId, running, closedForEvent])

  useEffect(() => { if (!ref.current) return; if (open) animatePromptEnter(ref.current) }, [open])

  // Esc referme jusqu'au prochain event
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') close() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, eventId])

  // countdown réel — depuis phaseSince porté par le store live
  useEffect(() => {
    if (!open || !running) return
    const t = window.setInterval(() => {
      const since = (live as any)?.phaseSince?.[phase]
      if (since) setElapsed(Math.max(0, Math.round((Date.now() - since) / 1000)))
    }, 1000)
    return () => window.clearInterval(t)
  }, [open, running, phase, live])

  const close = () => {
    setClosedForEvent(eventId)
    if (ref.current) animatePromptExit(ref.current).then(() => setOpen(false))
    else setOpen(false)
  }

  if (!open || !running) return null

  const label = `Event ${eventId}/${total} — ${live?.profile}/${live?.qdisc}/${live?.cc} rep ${live?.repetition} — ${phase} ${elapsed}/${phaseTotal}s`

  return (
    <div
      ref={ref}
      role="status"
      aria-label={label}
      onMouseEnter={undefined}
      style={{
        position: 'fixed',
        bottom: 40,
        left: '50%',
        transform: 'translateX(-50%)',
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '10px 14px 11px',
        background: 'rgba(16,16,18,0.92)',
        border: '1px solid rgba(255,255,255,0.08)',
        backdropFilter: 'blur(12px)',
        zIndex: 400,
        boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
        overflow: 'hidden',
      }}
    >
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', letterSpacing: '0.06em', whiteSpace: 'nowrap' }}>{label}</span>
      <button aria-label="Passer la cellule" title="passer la cellule courante (reprise possible)" onClick={async () => { await fetch('/api/run/skip', { method: 'POST' }).catch(() => {}); close() }} style={{ padding: '6px 12px', background: '#161618', color: '#f4b400', border: '1px solid #26262a', fontFamily: 'JetBrains Mono', fontSize: 11, letterSpacing: '0.06em', textTransform: 'uppercase', cursor: 'pointer' }}>Passer</button>
      <button aria-label="Fermer" onClick={close} style={{ padding: '6px 8px', background: 'transparent', border: '1px solid #26262a', color: '#9aa3ad', fontSize: 11, cursor: 'pointer' }}>Esc</button>
    </div>
  )
}
