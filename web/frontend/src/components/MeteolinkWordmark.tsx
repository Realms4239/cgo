import { useEffect, useRef } from 'react'
import { animateMeteolinkShimmer } from '../lib/anime'

export function MeteolinkWordmark({ compact }: { compact?: boolean }) {
  const ref = useRef<HTMLSpanElement>(null)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    // verified via animejs context7: svg.createDrawable + text.splitText + animatable dense in lib/anime
    // ponytail: delegate to animateMeteolinkShimmer — wordmark keeps METEOLINK + createDrawable/splitText via lib
    animateMeteolinkShimmer(el)
  }, [])
  return (
    <span ref={ref} className="wordmark-lockup" style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
      <span className="stripe" aria-hidden="true" style={{ display: 'flex', flexDirection: 'column', width: 4, height: 18, gap: 2 }}>
        <i style={{ flex: 1, background: 'var(--brand-1, #0066b1)' }} />
        <i style={{ flex: 1, background: 'var(--brand-2, #1c69d4)' }} />
        <i style={{ flex: 1, background: 'var(--brand-3, #e22718)' }} />
      </span>
      {/* 16×16 NOC icon satellite→wave stroke1.5 round */}
      <svg className="noc-icon" width={16} height={16} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" style={{ flex: 'none', color: 'var(--t-live, #5ad3e3)' }}>
        <circle cx={5} cy={5} r={2.4} stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" />
        <path d="M5 2 V1.2 M5 8.8 V7.6 M2.2 5 H1.2 M8.8 5 H7.6 M3.3 3.3 L2.4 2.4 M7.6 7.6 L6.7 6.7 M3.3 6.7 L2.4 7.6 M7.6 2.4 L6.7 3.3" stroke="currentColor" strokeWidth={1.2} strokeLinecap="round" />
        <path d="M1.8 11.5 Q5.5 8 8 11.5 T14.2 11.5" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" />
      </svg>
      <span
        className="wordmark-text"
        style={{
          fontFamily: 'var(--font-display, "Cormorant Garamond")',
          fontWeight: 600,
          fontSize: compact ? 14 : 20,
          letterSpacing: '0.04em',
          textTransform: 'uppercase' as const,
          color: '#f2f2f4',
          background: 'linear-gradient(90deg, #f2f2f4, #a9aeb6)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          backgroundClip: 'text',
          filter: 'drop-shadow(0 0 8px rgba(90,211,227,0.35))',
          lineHeight: 1,
        }}
      >
        METEOLINK
      </span>
      {!compact && <span className="hd-sub" style={{ fontFamily: 'var(--font-mono)', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase' as const, color: 'var(--text-faint)' }}>LIEN — trafic critique · AQM/BBR · Mada</span>}
    </span>
  )
}

export default MeteolinkWordmark
