import { useEffect, useRef } from 'react'
import { useUIStore } from '../store/ui'
import { animateToasts } from '../lib/anime'

export default function Toasts() {
  const toasts = useUIStore((s: any) => s.toasts)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    const els = Array.from(ref.current.querySelectorAll('.toast'))
    if (els.length) animateToasts(els as Element[])
  }, [toasts.length])
  return (
    <div ref={ref} id="toasts" aria-live="polite" style={{position:'fixed', bottom:16, right:16, zIndex:500, display:'flex', flexDirection:'column', gap:6}}>
      {toasts.map((t: any) => (
        <div key={t.id} className={'toast ' + (t.cls || '')} style={{
          background: 'rgba(12,15,20,.96)', border: '1px solid var(--hairline)', padding:'9px 13px',
          fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-body)', minWidth:240,
          borderColor: t.cls==='ok' ? 'rgba(31,163,72,.55)' : t.cls==='err' ? 'rgba(226,39,24,.55)' : t.cls==='blue' ? 'rgba(90,211,227,.55)' : 'var(--hairline)'
        }}>
          {t.msg}
        </div>
      ))}
    </div>
  )
}
