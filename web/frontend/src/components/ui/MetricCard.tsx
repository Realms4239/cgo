import { useMemo } from 'react'

type Trend = 'up' | 'down' | 'flat'

export function MetricCard({
  label,
  value,
  unit,
  trend = 'flat',
  spark,
  color = 'var(--t-live)',
}: {
  label: string
  value: string
  unit?: string
  trend?: Trend
  spark?: number[]
  color?: string
}) {
  const trendColor = trend === 'up' ? '#e22718' : trend === 'down' ? '#1fa348' : '#767b84'
  const trendSym = trend === 'up' ? '↗' : trend === 'down' ? '↘' : '—'
  // ponytail: sparkline is div polyline, D3 upgrade if thesis needs detail
  const path = useMemo(() => {
    if (!spark || spark.length < 2) return ''
    const w = 96, h = 24
    const min = Math.min(...spark), max = Math.max(...spark)
    const rng = max - min || 1
    return spark.map((v, i) => `${(i / (spark.length - 1)) * w},${h - ((v - min) / rng) * h}`).join(' ')
  }, [spark])
  return (
    <div className="card" style={{ padding: 12, border: '1px solid #26262a', background: 'var(--surface-card)', minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>{label}</span>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, padding: '2px 6px', border: '1px solid #26262a', background: trendColor + '14', color: trendColor, lineHeight: 1 }}>{trendSym}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 20, fontWeight: 700, color: '#f2f2f4', fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{value}</span>
        {unit ? <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#767b84' }}>{unit}</span> : null}
      </div>
      {path ? (
        <svg width={96} height={24} style={{ display: 'block', marginTop: 2 }} aria-hidden>
          <polyline fill="none" stroke={color} strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" points={path} style={{ filter: `drop-shadow(0 0 4px ${color}66)` }} />
        </svg>
      ) : (
        <div style={{ height: 24, borderTop: '1px solid #1a1a1e', marginTop: 2, opacity: 0.4 }} />
      )}
    </div>
  )
}

// ponytail: small variant reuses same shell, no extra abstraction
export function SmallMetric({ label, value, color = '#9aa3ad' }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ height: 40, border: '1px solid #26262a', background: 'rgba(154,163,173,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 12px' }}>
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', color, textTransform: 'uppercase' }}>{label}</span>
      <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', fontVariantNumeric: 'tabular-nums' }}>{value}</span>
    </div>
  )
}
