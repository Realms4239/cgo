import { useEffect, useRef } from 'react'
import { useUIStore } from '../store/ui'
import { animateFlash } from '../lib/anime'

export default function FlashBanner() {
  const flash = useUIStore((s: any) => s.flash)
  const ref = useRef<HTMLDivElement>(null)
  const genRef = useRef(0)
  useEffect(() => {
    if (!flash) return
    const gen = ++genRef.current
    if (ref.current) animateFlash(ref.current)
    const id = window.setTimeout(() => { if (gen === genRef.current) useUIStore.getState().setFlash(null) }, 2500)
    return () => clearTimeout(id)
  }, [flash])
  if (!flash) return null
  const bg =
    flash.type === 'success' ? 'var(--t-ok)' : flash.type === 'danger' ? 'var(--t-danger)' : 'var(--t-live)'
  // blanc sur vert var(--t-ok) = 3.29 (FAIL) → encre sombre 5.83 ; blanc sur rouge 4.63 ok
  const color = flash.type === 'success' ? 'var(--chart-on-ok)' : flash.type === 'info' ? 'var(--canvas-solid)' : 'var(--text-primary)'
  return (
    <div
      ref={ref}
      role="status"
      style={{
        position: 'fixed',
        top: 48,
        left: '50%',
        transform: 'translateX(-50%)',
        background: bg,
        color,
        padding: '8px 16px',
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--fs-tag)',
        letterSpacing: 'var(--tracking-caps)',
        textTransform: 'none',
        zIndex: 600,
        border: '1px solid rgba(0,0,0,0.15)',
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
      }}
    >
      {flash.msg}
    </div>
  )
}
