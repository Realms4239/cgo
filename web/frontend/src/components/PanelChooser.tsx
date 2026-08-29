import { useEffect, useState } from 'react'
import { useUIStore } from '../store/ui'
import { Pill } from './ui/Pill'

type Vis = Record<string, boolean>
const KEYS = ['Tableau live', 'Résultats', 'Provenance'] as const
const LS_KEY = 'panel-visibility'

function readVis(): Vis {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return { Wall: true, History: true, Archives: true }
}

export default function PanelChooser() {
  const panel = useUIStore((s: any) => s.panel)
  const setPanel = useUIStore((s: any) => s.setPanel)
  const [vis] = useState<Vis>(() => {
    if (typeof window === 'undefined') return { Wall: true, History: true, Archives: true }
    return readVis()
  })

  // vis persisted via PanelChooser mount; no toggle on nav — wall kit keeps hero always visible
  useEffect(() => {
    try { localStorage.setItem(LS_KEY, JSON.stringify(vis)) } catch {}
  }, [vis])

  const visAttr = KEYS.filter((k) => vis[k]).join(',')
  const nav = (label: string) => {
    if (label === 'Tableau live') setPanel('live')
    else if (label === 'Résultats') setPanel('resultats')
    else if (label === 'Provenance') setPanel('integrite')
  }
  const isActive = (label: string) => {
    if (label === 'Tableau live') return panel === 'live'
    if (label === 'Résultats') return panel === 'resultats'
    if (label === 'Provenance') return panel === 'integrite'
    return false
  }

  const [tri, setTri] = useState<{metric:boolean; chart:'line'|'bar'|'area'; source:'live'|'frozen'|'both'}>(()=>{
    try {
      const r=localStorage.getItem('panel-chooser-tri')
      if(r){ const p=JSON.parse(r)
        // legacy boolean chart → craft mapping (true=line, false=bar)
        const chart = typeof p.chart==='boolean' ? (p.chart?'line':'bar') : (p.chart||'line')
        return {metric:!!p.metric, chart, source:p.source||'both'}
      }
    }catch{}
    return {metric:true, chart:'line', source:'both'}
  })
  useEffect(()=>{ try{localStorage.setItem('panel-chooser-tri', JSON.stringify(tri))}catch{}; window.dispatchEvent(new CustomEvent('panel-chooser-tri',{detail:tri})) },[tri])

  return (
    <div
      className="panel-chooser"
      data-panel-visibility={visAttr}
      data-tri={`${tri.metric?'m':''}c${tri.source}`}
      data-chart-craft={tri.chart}
      role="group"
      aria-label="Choix du panneau"
      style={{
        display: 'flex',
        gap: 8,
        padding: '10px 14px',
        background: 'rgba(16,16,18,0.92)',
        border: '1px solid var(--hairline, #26262a)',
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
        borderRadius: 0,
        alignItems: 'center',
        flexWrap:'wrap' as const,
      }}
    >
      {KEYS.map((k) => (
        <button
          key={k}
          onClick={() => nav(k)}
          aria-pressed={isActive(k)}
          aria-label={k}
          style={{
            fontFamily: 'JetBrains Mono, ui-monospace, monospace',
            fontSize: 11,
            fontVariantNumeric: 'tabular-nums',
            letterSpacing: '0.06em',
            textTransform: 'uppercase' as const,
            padding: '6px 10px',
            border: '1px solid var(--hairline, #26262a)',
            background: isActive(k) ? 'rgba(90,211,227,0.12)' : 'transparent',
            color: isActive(k) ? 'var(--t-live, #5ad3e3)' : 'var(--text-muted, #8b9099)',
            lineHeight: 1,
            cursor: 'pointer',
            opacity: vis[k] ? 1 : 0.5,
          }}
        >
          {k}
        </button>
      ))}
      <span style={{width:1, height:18, background:'var(--hairline)', margin:'0 4px'}} aria-hidden />
      {/* labeled controls — the control name is always visible, no cryptic one-word chrome */}
      <Pill label="groupes" value={tri.metric ? 'on' : 'off'} on={tri.metric} onClick={() => setTri(t => ({ ...t, metric: !t.metric }))} title="groupes de métriques on/off" />
      <Pill label="craft" value={tri.chart} on={tri.chart !== 'line'} onClick={() => setTri(t => ({ ...t, chart: t.chart === 'line' ? 'bar' : t.chart === 'bar' ? 'area' : 'line' }))} title="chart craft line|bar|area" />
      <Pill label="source" value={tri.source} on={tri.source !== 'live'} onClick={() => setTri(t => ({ ...t, source: t.source === 'live' ? 'frozen' : t.source === 'frozen' ? 'both' : 'live' }))} title="source live|frozen|both" />
      <select value={tri.source} onChange={e=>setTri(t=>({...t, source:e.target.value as 'live'|'frozen'|'both'}))} aria-label="source live|frozen|both" className="sr-only" tabIndex={-1}>
        <option value="live">live</option><option value="frozen">frozen</option><option value="both">both</option>
      </select>
    </div>
  )
}
