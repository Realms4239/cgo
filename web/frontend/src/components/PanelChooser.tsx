import { useEffect, useState } from 'react'
import { useUIStore } from '../store/ui'

type Vis = Record<string, boolean>
const KEYS = ['Wall', 'History', 'Archives'] as const
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
  const [vis, setVis] = useState<Vis>(() => {
    if (typeof window === 'undefined') return { Wall: true, History: true, Archives: true }
    return readVis()
  })

  useEffect(() => {
    try { localStorage.setItem(LS_KEY, JSON.stringify(vis)) } catch {}
  }, [vis])

  const visAttr = KEYS.filter((k) => vis[k]).join(',')
  const nav = (label: string) => {
    if (label === 'Wall') setPanel('campagne')
    else if (label === 'History') setPanel('resultats')
    else if (label === 'Archives') setPanel('integrite')
  }
  const isActive = (label: string) => {
    if (label === 'Wall') return panel === 'campagne' || panel === 'live'
    if (label === 'History') return panel === 'resultats'
    if (label === 'Archives') return panel === 'integrite'
    return false
  }

  return (
    <div
      className="panel-chooser"
      data-panel-visibility={visAttr}
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
      }}
    >
      {KEYS.map((k) => (
        <button
          key={k}
          onClick={() => {
            nav(k)
            setVis((v) => ({ ...v, [k]: !v[k] }))
            // toggle visibility, but navigation takes precedence — re-enable on nav
            setVis((v) => ({ ...v, [k]: true }))
          }}
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
    </div>
  )
}
