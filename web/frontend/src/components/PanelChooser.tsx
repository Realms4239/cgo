import { useEffect, useState } from 'react'
import { useUIStore, PANELS } from '../store/ui'
import { Pill } from './ui/Pill'

// La nav de tête et le rail lisent la même source (PANELS + store) : une
// seule vérité pour l'état actif, plus de liste partielle à maintenir.
export default function PanelChooser() {
  const panel = useUIStore((s: any) => s.panel)
  const setPanel = useUIStore((s: any) => s.setPanel)
  const showTri = panel === 'live' || panel === 'resultats'

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
      data-tri={`${tri.metric?'m':''}c${tri.source}`}
      data-chart-craft={tri.chart}
      role="group"
      aria-label="Choix du panneau"
    >
      {PANELS.map((p) => {
        const active = panel === p.id
        return (
          <button
            key={p.id}
            className="panel-chooser-btn"
            onClick={() => setPanel(p.id)}
            aria-pressed={active}
            aria-label={p.label}
            title={p.label}
          >
            {p.label}
          </button>
        )
      })}
      {showTri ? <>
      <span className="panel-chooser-sep" aria-hidden />
      {/* labeled controls — visibles seulement là où les graphiques existent */}
      <Pill label="groupes" value={tri.metric ? 'on' : 'off'} on={tri.metric} onClick={() => setTri(t => ({ ...t, metric: !t.metric }))} title="groupes de métriques on/off" />
      <Pill label="craft" value={tri.chart} on={tri.chart !== 'line'} onClick={() => setTri(t => ({ ...t, chart: t.chart === 'line' ? 'bar' : t.chart === 'bar' ? 'area' : 'line' }))} title="chart craft line|bar|area" />
      <Pill label="source" value={tri.source} on={tri.source !== 'live'} onClick={() => setTri(t => ({ ...t, source: t.source === 'live' ? 'frozen' : t.source === 'frozen' ? 'both' : 'live' }))} title="source : live = mesures en cours uniquement · gelé = archives uniquement" />
      </> : null}
    </div>
  )
}
