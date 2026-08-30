import { useEffect, useRef, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { explain } from '../lib/explain'

// Bulle d'aide : survol affiche, clic fige, clic dehors referme.
export default function Explain({ term, children, style }: { term: string; children: ReactNode; style?: React.CSSProperties }) {
  const ref = useRef<HTMLSpanElement>(null)
  const [hover, setHover] = useState(false)
  const [pin, setPin] = useState(false)
  const [pos, setPos] = useState<{ left: number; top: number }>({ left: 8, top: 8 })
  const text = explain(term)
  const open = (hover || pin) && !!text

  const updatePos = () => {
    const r = ref.current?.getBoundingClientRect()
    if (!r) return
    setPos({
      left: Math.max(8, Math.min(r.left, window.innerWidth - 300)),
      top: Math.max(8, r.top - 8),
    })
  }

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
      data-explain={term}
      style={{ cursor: 'help', borderBottom: '1px dotted rgba(138,150,160,0.55)', padding: '2px 1px', ...style }}
      onMouseEnter={() => { updatePos(); setHover(true) }}
      onMouseLeave={() => setHover(false)}
      onClick={(e) => { e.stopPropagation(); updatePos(); setPin(p => !p) }}
    >
      {children}
      {open &&
        createPortal(
          <span
            role="tooltip"
            className="explain-pop"
            style={{
              position: 'fixed', left: pos.left, top: pos.top, transform: 'translateY(-100%)',
              width: 280, padding: '10px 12px', zIndex: 9999,
              background: 'rgba(12,12,14,0.97)', border: '1px solid #2c2c31', borderRadius: 8,
              color: '#d6dade', fontFamily: 'Inter, sans-serif', fontSize: 12, lineHeight: 1.45,
              boxShadow: '0 12px 32px rgba(0,0,0,0.55)', display: 'block', textAlign: 'left',
              whiteSpace: 'normal', pointerEvents: 'none',
            }}
          >
            {text}
          </span>,
          document.body,
        )}
    </span>
  )
}
