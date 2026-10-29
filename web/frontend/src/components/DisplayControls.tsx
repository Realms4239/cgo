// Affichage de l'observateur — groupes de métriques, forme de courbe, source.
// Les mêmes préférences qu'avant, déplacées du chrome global vers le banc.
import { useEffect, useState } from 'react'
import { Pill } from './ui/Pill'

type Tri = { metric: boolean; chart: 'line' | 'bar' | 'area'; source: 'live' | 'frozen' | 'both' }

const KEY = 'panel-chooser-tri'

const initial = (): Tri => {
  try {
    const r = localStorage.getItem(KEY)
    if (r) {
      const p = JSON.parse(r)
      const chart = typeof p.chart === 'boolean' ? (p.chart ? 'line' : 'bar') : (p.chart || 'line')
      return { metric: !!p.metric, chart, source: p.source || 'both' }
    }
  } catch {}
  return { metric: true, chart: 'line', source: 'both' }
}

export default function DisplayControls() {
  const [tri, setTri] = useState<Tri>(initial)
  useEffect(() => {
    try { localStorage.setItem(KEY, JSON.stringify(tri)) } catch {}
    window.dispatchEvent(new CustomEvent('panel-chooser-tri', { detail: tri }))
  }, [tri])

  return (
    <div className="display-controls" data-tri={`${tri.metric ? 'm' : ''}c${tri.source}`} data-chart-craft={tri.chart} role="group" aria-label="Affichage de l'observateur">
      <Pill label="groupes" value={tri.metric ? 'on' : 'off'} on={tri.metric} onClick={() => setTri(t => ({ ...t, metric: !t.metric }))} title="groupes de métriques on/off" />
      <Pill label="forme" value={tri.chart} on={tri.chart !== 'line'} onClick={() => setTri(t => ({ ...t, chart: t.chart === 'line' ? 'bar' : t.chart === 'bar' ? 'area' : 'line' }))} title="forme de courbe : ligne, barre, aire" />
      <Pill label="source" value={tri.source} on={tri.source !== 'live'} onClick={() => setTri(t => ({ ...t, source: t.source === 'live' ? 'frozen' : t.source === 'frozen' ? 'both' : 'live' }))} title="source : live = mesures en cours · archive = mesures archivées" />
    </div>
  )
}
