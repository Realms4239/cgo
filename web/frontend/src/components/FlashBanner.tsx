import { useEffect, useRef } from 'react'
import { useUIStore } from '../store/ui'
import { animateFlash } from '../lib/anime'

export default function FlashBanner() {
  const flash = useUIStore((s: any) => s.flash)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (flash && ref.current) {
      animateFlash(ref.current)
      const id = window.setTimeout(() => useUIStore.getState().setFlash(null), 2500)
      return () => clearTimeout(id)
    }
  }, [flash])
  if (!flash) return null
  const bg =
    flash.type === 'success' ? '#1fa348' : flash.type === 'danger' ? '#e22718' : '#5ad3e3'
  const color = flash.type === 'info' ? '#000' : '#fff'
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
        fontFamily: 'JetBrains Mono',
        fontSize: 11,
        letterSpacing: '0.08em',
        textTransform: 'uppercase',
        zIndex: 600,
        border: '1px solid rgba(0,0,0,0.15)',
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
      }}
    >
      {flash.msg}
    </div>
  )
}
