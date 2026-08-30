import { useEffect, useLayoutEffect, useRef, useState, type ReactNode } from 'react'
import { explain } from '../lib/explain'

// Micro-explication d'un élément essentiel: survol montre,
// clic épingle (projection-friendly — pas de curseur tremblant), clic dehors
// referme. Un seul composant, un seul dictionnaire (lib/explain.ts).
export default function Explain({ term, children, style }: { term: string; children: ReactNode; style?: React.CSSProperties }) {
  const ref = useRef<HTMLSpanElement>(null)
  const [hover, setHover] = useState(false)
  const [pin, setPin] = useState(false)
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null)
  const text = explain(term)
  const open = (hover || pin) && !!text

  useLayoutEffect(() => {
    if (!open || !ref.current) { setPos(null); return }
    const r = ref.current.getBoundingClientRect()
    setPos({
      left: Math.max(8, Math.min(r.left, window.innerWidth - 300)),
      top: Math.max(8, r.top - 8),
    })
  }, [open])

  useEffect(() => {
    if (!pin) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setPin(false)
    }
    const onEsc = (e: KeyboardEvent) => { if (e.key === 'Escape') setPin(false) }
    document.addEventListener('click', onDoc)
    document.addEventListener('keydown', onEsc)
    return () => { document.removeEventListener('click', onDoc); document.removeEventListener('keydown', onEsc) }
  }, [pin])

  if (!text) return <span style={style}>{children}</span>
  return (
    <span
      ref={ref}
      style={{ cursor: 'help', borderBottom: '1px dotted rgba(138,150,160,0.55)', ...style }}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      onClick={(e) => { e.stopPropagation(); setPin(p => !p) }}
    >
      {children}
      {open && pos && (
        <span
          role="tooltip"
          className="explain-pop"
          style={{
            position: 'fixed', left: pos.left, top: pos.top, transform: 'translateY(-100%)',
            width: 280, padding: '10px 12px', zIndex: 900,
            background: 'rgba(12,12,14,0.97)', border: '1px solid #2c2c31', borderRadius: 8,
            color: '#d6dade', fontFamily: 'Inter var, sans-serif', fontSize: 12, lineHeight: 1.45,
            boxShadow: '0 12px 32px rgba(0,0,0,0.55)', display: 'block', textAlign: 'left',
            whiteSpace: 'normal', pointerEvents: 'none',
          }}
        >
          {text}
        </span>
      )}
    </span>
  )
}
