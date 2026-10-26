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
  const baseColor = showVerdict ? LEVEL_COLOR[verdictLevel as Level] : 'var(--text-primary)'
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
    <div data-metric={term ?? label} data-testid={`metric-${term ?? label}`} className="card mc-card">
      <div className="mc-head">
        <span className="mono mc-label">
          {term ? <Explain term={term}>{label}</Explain> : label}
        </span>
        <span className="mono mc-verdict" style={{ background: `color-mix(in srgb, ${verdictColor} 8%, transparent)`, color: verdictColor }}>{verdictSym}</span>
      </div>
      <div className="mc-value-row">
        <span className={'mono mc-value' + (flash ? (flash === 'up' ? ' val-flash-up' : ' val-flash-down') : '')} style={{ color: flash === 'up' ? 'var(--t-ok)' : flash === 'down' ? 'var(--t-danger-text)' : baseColor }}>{value}</span>
        {unit ? <span className="mono mc-unit">{unit}</span> : null}
      </div>
      {threshold && value !== '—' ? (
        <div role="img" aria-label={`seuil ${threshold.marker} %`} className="mc-track">
          <div className="mc-fill" style={{ background: baseColor, transform: `scaleX(${Math.max(0, Math.min(1, parseFloat(value) / threshold.max))})` }} />
          <div className="mc-marker" style={{ left: `${(threshold.marker / threshold.max) * 100}%` }} />
        </div>
      ) : null}
      {path ? (
        <svg width={60} height={12} className="mc-spark" aria-hidden>
          <defs><clipPath id={clipId}><rect width={60} height={12} rx={4} /></clipPath></defs>
          <g clipPath={`url(#${clipId})`}><polyline fill="none" stroke={color} strokeWidth={1.2} strokeLinecap="round" strokeLinejoin="round" points={path} style={{ filter: `drop-shadow(0 0 4px color-mix(in srgb, ${color} 40%, transparent))` }} /></g>
        </svg>
      ) : (
        <div className="mc-empty" />
      )}
    </div>
  )
}

function trendColorOf(trend: Trend): string {
  return trend === 'up' ? 'var(--t-danger-text)' : trend === 'down' ? 'var(--t-ok)' : 'var(--text-faint)'
}
function trendSymOf(trend: Trend): string {
  return trend === 'up' ? '↗' : trend === 'down' ? '↘' : '—'
}

// Variante compacte: mêmes coquilles, pas d'abstraction en plus.
export function SmallMetric({ label, value, color = 'var(--t-neutral)' }: { label: string; value: string; color?: string }) {
  return (
    <div className="sm-card">
      <span className="mono sm-label" style={{ color }}>{label}</span>
      <span className="mono sm-value">{value}</span>
    </div>
  )
}
