import { useEffect, useRef, useState } from 'react'
import { useUIStore, PANELS } from './store/ui'
import { connectSSE, disconnectSSE } from './lib/sse'
import { animateViewEnter } from './lib/anime'
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
import CommandPalette from './components/CommandPalette'
import MeteolinkWordmark from './components/MeteolinkWordmark'
import PanelChooser from './components/PanelChooser'
// En-tête METEOLINK — dégradé wordmark + icône 16×16 par MeteolinkWordmark

export default function App() {
  const panel = useUIStore((s) => s.panel)
  const setPanel = useUIStore((s) => s.setPanel)
  const live = useUIStore((s) => s.live)
  const connected = useUIStore((s) => s.connected)
  const sseStatus = useUIStore((s) => s.sseStatus)
  const railPinned = useUIStore((s) => s.railPinned)
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [mode, setMode] = useState<string>('')

  useEffect(() => {
    connectSSE()
    return () => disconnectSSE()
  }, [])
  // badge capacité (Q10) — rechargé à chaque (re)connexion, un redéploiement ou
  // un changement de mode serveur est pris en compte sans rechargement manuel
  useEffect(() => {
    if (!connected) return
    fetch('/api/health').then(r => r.json()).then(j => setMode(j.mode ?? '')).catch(() => {})
  }, [connected])

  useEffect(() => {
    // entrée légère : fondu CSS seulement — plus de blur JS ni de re-stagger des cartes
    animateViewEnter()
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
      <div id="shell" className={railPinned ? '' : 'rail-min'}>
        <a className="skip-link" href="#main">Aller au contenu</a>
        <header style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 12, padding: '0 16px', borderBottom: '1px solid var(--hairline, #26262a)', background: 'var(--surface-soft, #0b0b0c)', minWidth: 0 }}>
          <MeteolinkWordmark />
          <div style={{ flex: 1, display: 'flex', justifyContent: 'center', minWidth: 0 }}>
            <PanelChooser />
          </div>
          <div className="hd-right">
            {mode === 'observe' && <span className="mono" title="Audit et consultation uniquement — campagne et façonnage vivent sur l'hôte Linux (docs/deploy.md)" style={{ color: 'var(--t-warn, #f4b400)', border: '1px solid currentColor', padding: '2px 8px', fontSize: 10, letterSpacing: '0.08em' }}>OBSERVATION</span>}
            <button onClick={() => setSettingsOpen(true)} aria-label="Réglages" title="Réglages">⚙</button>
            <span className="mono" style={{ color: connected ? 'var(--t-ok)' : 'var(--t-danger)' }}>{connected ? '● connecté' : '○ déconnecté'}</span>
            <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
          </div>
        </header>

        <Rail />
        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        {/* quick actions are running-state chrome — never block an idle view */}
        {live?.running && <QuickActionsPrompt />}
        <main id="main">
          {/* vues toutes montées — retour instantané, pas de refetch/réinit ECharts à chaque switch */}
          <section id="v-campagne" className="view" hidden={panel !== 'campagne'}><CampagneView /></section>
          <section id="v-live" className="view" hidden={panel !== 'live'}><LiveView /></section>
          <section id="v-resultats" className="view" hidden={panel !== 'resultats'}><ResultatsView /></section>
          <section id="v-integrite" className="view" hidden={panel !== 'integrite'}><IntegriteView /></section>
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
