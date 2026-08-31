import { useId, useMemo, useRef, useEffect, useState } from 'react'
import { lttb } from '../../lib/lttb'
import Explain from '../Explain'

type Trend = 'up' | 'down' | 'flat'

export function MetricCard({
  label,
  value,
  unit,
  trend = 'flat',
  spark,
  color = 'var(--t-live)',
  term,
}: {
  label: string
  value: string
  unit?: string
  trend?: Trend
  spark?: number[]
  color?: string
  term?: string
}) {
  const trendColor = trend === 'up' ? '#e22718' : trend === 'down' ? '#1fa348' : '#767b84'
  const trendSym = trend === 'up' ? '↗' : trend === 'down' ? '↘' : '—'
  const clipId = useId().replace(/:/g, '-')
  // flash directionnel — la carte s'allume quand SA valeur change (montée = rouge,
  // descente = vert pour une métrique où bas = mieux; le parent décide via good/badIsUp)
  const prevRef = useRef(value)
  const [flash, setFlash] = useState<'' | 'up' | 'down'>('')
  useEffect(() => {
    if (value === prevRef.current || value === '—' || prevRef.current === '—') { prevRef.current = value; return }
    const pv = parseFloat(prevRef.current), nv = parseFloat(value)
    prevRef.current = value
    if (!Number.isFinite(pv) || !Number.isFinite(nv) || pv === nv) return
    // sur le mur live, une métrique qui monte en valeur = charge qui grossit (rouge);
    // qui descend = récupération (vert) — sauf trend 'up' explicite (goodput)
    setFlash(trend === 'up' ? (nv > pv ? 'up' : 'down') : (nv > pv ? 'down' : 'up'))
    const t = setTimeout(() => setFlash(''), 400)
    return () => clearTimeout(t)
  }, [value, trend])
  // Sparkline 60×12 (clipPath rx4) + lttb40 — zone D3 si remplissage utile.
  const path = useMemo(() => {
    if (!spark || spark.length < 2) return ''
    let data = spark
    if (spark.length > 40) {
      const tmp = spark.map((v, i) => [i, v] as [number, number])
      data = lttb(tmp, 40).map(([, v]) => v)
    }
    const w = 60, h = 12
    const min = Math.min(...data), max = Math.max(...data)
    const rng = max - min || 1
    return data.map((v, i) => `${(i / (data.length - 1)) * w},${h - ((v - min) / rng) * h}`).join(' ')
  }, [spark])
  return (
    <div data-metric={label} data-testid={`metric-${label}`} className="card" style={{ padding: 12, border: '1px solid #26262a', background: 'var(--surface-card)', minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
          {term ? <Explain term={term}>{label}</Explain> : label}
        </span>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, padding: '2px 6px', border: '1px solid #26262a', background: trendColor + '14', color: trendColor, lineHeight: 1 }}>{trendSym}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
        <span className={'mono' + (flash ? (flash === 'up' ? ' val-flash-up' : ' val-flash-down') : '')} style={{ fontFamily: 'JetBrains Mono', fontSize: 20, fontWeight: 700, color: flash === 'up' ? 'var(--t-ok)' : flash === 'down' ? 'var(--t-danger)' : '#f2f2f4', fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{value}</span>
        {unit ? <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#767b84' }}>{unit}</span> : null}
      </div>
      {path ? (
        <svg width={60} height={12} style={{ display: 'block', marginTop: 2 }} aria-hidden>
          <defs><clipPath id={clipId}><rect width={60} height={12} rx={4} /></clipPath></defs>
          <g clipPath={`url(#${clipId})`}><polyline fill="none" stroke={color} strokeWidth={1.2} strokeLinecap="round" strokeLinejoin="round" points={path} style={{ filter: `drop-shadow(0 0 4px ${color}66)` }} /></g>
        </svg>
      ) : (
        <div style={{ height: 12, borderTop: '1px solid #1a1a1e', marginTop: 2, opacity: 0.4 }} />
      )}
    </div>
  )
}

// Variante compacte: mêmes coquilles, pas d'abstraction en plus.
export function SmallMetric({ label, value, color = '#9aa3ad' }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ height: 40, border: '1px solid #26262a', background: 'rgba(154,163,173,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 12px' }}>
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', color, textTransform: 'uppercase' }}>{label}</span>
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', fontVariantNumeric: 'tabular-nums' }}>{value}</span>
    </div>
  )
}
