import { useEffect, useRef, useState } from 'react'
import { useUIStore, PANELS } from './store/ui'
import { connectSSE, disconnectSSE } from './lib/sse'
import { animateGrid, animateViewEnter } from './lib/anime'
import CampagneView from './views/CampagneView'
import LiveView from './views/LiveView'
import ResultatsView from './views/ResultatsView'
import IntegriteView from './views/IntegriteView'
import Toasts from './components/Toasts'
import FlashBanner from './components/FlashBanner'
import SettingsDrawer from './components/SettingsDrawer'
import ErrorBoundary from './components/ErrorBoundary'
import Rail from './components/Rail'
import WebGLMesh from './components/WebGLMesh'
import QuickActionsPrompt from './components/QuickActionsPrompt'
import OnboardingNudge from './components/OnboardingNudge'
import CommandPalette from './components/CommandPalette'
import MeteolinkWordmark from './components/MeteolinkWordmark'
import PanelChooser from './components/PanelChooser'
// METEOLINK header lockup — wordmark gradient + 16×16 NOC icon via MeteolinkWordmark

export default function App() {
  const panel = useUIStore((s) => s.panel)
  const setPanel = useUIStore((s) => s.setPanel)
  const live = useUIStore((s) => s.live)
  const connected = useUIStore((s) => s.connected)
  const sseStatus = useUIStore((s) => s.sseStatus)
  const density = useUIStore((s) => s.density)
  const setDensity = useUIStore((s) => s.setDensity)
  const railPinned = useUIStore((s) => s.railPinned)
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    connectSSE()
    return () => disconnectSSE()
  }, [])

  useEffect(() => {
    animateViewEnter()
    const els = document.querySelectorAll('.panel-stack .card, .view .card')
    if (els.length) animateGrid(Array.from(els) as Element[])
  }, [panel])

  const wasRunningRef = useRef(!!live?.running)
  useEffect(() => {
    const running = !!live?.running
    if (wasRunningRef.current && !running) {
      setPanel('resultats')
      useUIStore.getState().setFlash({ type: 'success', msg: 'Campagne terminée' })
      useUIStore.getState().pushToast('Campagne terminée', 'ok')
    }
    if (!wasRunningRef.current && running) {
      const cur = useUIStore.getState().panel
      if (cur === 'campagne') {
        setPanel('live')
        useUIStore.getState().setFlash({ type: 'success', msg: 'CHARGE — campagne lancée' })
      }
    }
    wasRunningRef.current = running
  }, [live?.running, setPanel])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); setPaletteOpen((o: boolean) => !o); return }
      const tag = (e.target as HTMLElement)?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA') return
      const hit = PANELS.find((p) => p.key === e.key)
      if (hit) { setPanel(hit.id); e.preventDefault() }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [setPanel])

  const run8 = live?.event_id != null ? String(live.event_id).padStart(8, '0').slice(-8) : '────────'
  const hashSrc = `${live?.profile ?? ''}${live?.qdisc ?? ''}${live?.cc ?? ''}`
  const hash8 = hashSrc ? hashSrc.slice(0, 8).padEnd(8, '·').slice(0, 8) : '────────'

  return (
    <ErrorBoundary>
      <WebGLMesh />
      <div id="shell" data-density={density} className={railPinned ? '' : 'rail-min'}>
        <a className="skip-link" href="#main">Aller au contenu</a>
        <header style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '0 16px', borderBottom: '1px solid var(--hairline, #26262a)', background: 'var(--surface-soft, #0b0b0c)', minWidth: 0 }}>
          <MeteolinkWordmark />
          <div style={{ flex: 1, display: 'flex', justifyContent: 'center', minWidth: 0 }}>
            <PanelChooser />
          </div>
          <div className="hd-right">
            <button onClick={() => setSettingsOpen(true)} aria-label="Réglages" title="Réglages">⚙</button>
            <button onClick={() => setDensity(density === 'airy' ? 'dense' : 'airy')} aria-label="Densité">{density}</button>
            <span className="mono" style={{ color: connected ? 'var(--t-ok)' : 'var(--t-danger)' }}>{connected ? '● connecté' : '○ déconnecté'}</span>
            <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
          </div>
        </header>

        <Rail />
        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        {/* quick actions are running-state chrome — never block an idle view */}
        {live?.running && <QuickActionsPrompt />}
        <main id="main">
          {panel === 'campagne' && <section id="v-campagne" className="view on"><OnboardingNudge /><CampagneView /></section>}
          {panel === 'live' && <section id="v-live" className="view on"><LiveView /></section>}
          {panel === 'resultats' && <section id="v-resultats" className="view on"><ResultatsView /></section>}
          {panel === 'integrite' && <section id="v-integrite" className="view on"><IntegriteView /></section>}
        </main>

        <footer className="foot-ticker" style={{ height: 28, display: 'flex', alignItems: 'center', gap: 12, padding: '0 16px', borderTop: '1px solid var(--hairline)', background: 'var(--surface-soft)', fontFamily: 'var(--font-mono)', fontSize: 10, fontVariantNumeric: 'tabular-nums', letterSpacing: '0.06em', textTransform: 'uppercase' as const, color: 'var(--text-faint)' }}>
          <span id="ft-prov">source: {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
          <span aria-hidden="true" style={{ opacity: 0.4 }}>|</span>
          <span id="ft-run" className="mono" style={{ fontVariantNumeric: 'tabular-nums' }}>run {run8} · hash {hash8} · {live?.phase ?? 'idle'}</span>
          <span className="fill" style={{ marginLeft: 'auto' }} />
          <span id="ft-sse" className="mono" style={{ fontVariantNumeric: 'tabular-nums' }}>SSE {sseStatus}</span>
        </footer>
        {settingsOpen && <SettingsDrawer onClose={() => setSettingsOpen(false)} />}
        <FlashBanner />
        <Toasts />
      </div>
    </ErrorBoundary>
  )
}
