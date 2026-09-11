import { useId, useMemo, useRef, useEffect, useState } from 'react'
import { lttb } from '../../lib/lttb'
import { LEVEL_COLOR, type Level } from '../../lib/settings'
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
  verdict,
  verdictLevel,
  threshold,
}: {
  label: string
  value: string
  unit?: string
  trend?: Trend
  spark?: number[]
  color?: string
  term?: string
  // grammaire Stat Grafana : le mot de verdict porte le jugement (jamais la
  // couleur seule) ; quand il est là, il remplace le glyphe de tendance
  verdict?: string
  verdictLevel?: Level
  // barre de seuil — deadline seule (0–100, repère 95), jamais ailleurs
  threshold?: { marker: number; max: number }
}) {
  const showVerdict = verdict != null && verdictLevel != null
  const verdictColor = showVerdict ? LEVEL_COLOR[verdictLevel as Level] : trendColorOf(trend)
  const verdictSym = showVerdict ? verdict as string : trendSymOf(trend)
  // sévérité réelle seulement : la valeur prend la couleur du verdict, jamais
  // celle du flash (le flash transitoire reste un flash, pas une sévérité)
  const baseColor = showVerdict ? LEVEL_COLOR[verdictLevel as Level] : '#f2f2f4'
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
    <div data-metric={term ?? label} data-testid={`metric-${term ?? label}`} className="card" style={{ padding: 12, border: '1px solid #26262a', background: 'var(--surface-card)', minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
          {term ? <Explain term={term}>{label}</Explain> : label}
        </span>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, padding: '2px 6px', border: '1px solid #26262a', background: verdictColor + '14', color: verdictColor, lineHeight: 1 }}>{verdictSym}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
        <span className={'mono' + (flash ? (flash === 'up' ? ' val-flash-up' : ' val-flash-down') : '')} style={{ fontFamily: 'JetBrains Mono', fontSize: 20, fontWeight: 700, color: flash === 'up' ? 'var(--t-ok)' : flash === 'down' ? 'var(--t-danger-text, #e84a3a)' : baseColor, fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{value}</span>
        {unit ? <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: 'var(--text-faint, #7e838c)' }}>{unit}</span> : null}
      </div>
      {threshold && value !== '—' ? (
        <div role="img" aria-label={`seuil ${threshold.marker} %`} style={{ height: 4, background: 'rgba(255,255,255,0.06)', borderRadius: 2, position: 'relative', marginTop: 2 }}>
          <div style={{ width: `${Math.max(0, Math.min(100, (parseFloat(value) / threshold.max) * 100))}%`, height: '100%', background: baseColor, borderRadius: 2, transition: 'width 0.4s ease' }} />
          <div style={{ position: 'absolute', left: `${(threshold.marker / threshold.max) * 100}%`, top: -2, bottom: -2, width: 1, background: 'var(--text-faint, #7e838c)' }} />
        </div>
      ) : null}
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

function trendColorOf(trend: Trend): string {
  return trend === 'up' ? 'var(--t-danger-text, #e84a3a)' : trend === 'down' ? '#1fa348' : 'var(--text-faint, #7e838c)'
}
function trendSymOf(trend: Trend): string {
  return trend === 'up' ? '↗' : trend === 'down' ? '↘' : '—'
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
