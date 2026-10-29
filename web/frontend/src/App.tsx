import { useEffect, useRef, useState } from 'react'
import { useUIStore, PANELS, type PanelId } from './store/ui'
import { connectSSE, disconnectSSE } from './lib/sse'
import { clearLive } from './lib/live'
import { animateViewEnter } from './lib/anime'
import SyntheseView from './views/SyntheseView'
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
  // les vues lourdes (ECharts, tableaux) ne se montent qu'à la première visite ;
  // une fois montées elles restent en place pour un retour instantané.
  const [mounted, setMounted] = useState<Partial<Record<PanelId, boolean>>>({ synthese: true })

  useEffect(() => {
    setMounted(m => m[panel] ? m : { ...m, [panel]: true })
  }, [panel])

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
      // fantôme d'après-arrêt : les anneaux et la présentation retombent —
      // profil/charge/PASS périmés ne restent pas à côté de `phase idle`
      clearLive()
      const last = useUIStore.getState().live
      if (last) {
        useUIStore.getState().setLive({
          ...last, phase: 'idle', running: false, load_status: '',
          profile: '', qdisc: '', cc: '', repetition: 0,
          rtt_p50_ms: 0, rtt_p95_ms: 0, small_p95_ms: 0, bulk_goodput_mbps: 0,
          drops: 0, wasted_bytes: undefined, cost_ar_per_h: undefined,
          deadline_ok_pct: undefined, gates: Array(8).fill(null),
        } as typeof last)
      }
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

  // (pied de page : l'event_id s'affiche comme événement, jamais `run`)

  return (
    <ErrorBoundary>
      <WebGLMesh />
      <div id="shell" className={railPinned ? '' : 'rail-min'}>
        <a className="skip-link" href="#main">Aller au contenu</a>
        <header className="app-header">
          <MeteolinkWordmark />
          <div className="app-header-center">
            <PanelChooser />
          </div>
          <div className="hd-right">
            {mode === 'observe' && <span className="mono obs-badge" title="Audit et consultation uniquement — campagne et façonnage vivent sur l'hôte Linux (docs/deploy.md)">OBSERVATION</span>}
            <button onClick={() => setSettingsOpen(true)} aria-label="Réglages" title="Réglages">⚙</button>
            <span className="mono" style={{ color: connected ? 'var(--t-ok)' : 'var(--t-danger-text)' }}>{connected ? '● connecté' : '○ déconnecté'}</span>
            <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
          </div>
        </header>

        <Rail />
        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        {/* quick actions are running-state chrome — never block an idle view */}
        {live?.running && <QuickActionsPrompt />}
        <main id="main">
          {/* vues montées à la première visite — le socle et les graphiques
              lourds ne bloquent plus le premier rendu */}
          <section id="v-synthese" className="view" hidden={panel !== 'synthese'}><SyntheseView /></section>
          {mounted.campagne ? <section id="v-campagne" className="view" hidden={panel !== 'campagne'}><CampagneView /></section> : null}
          {mounted.live ? <section id="v-live" className="view" hidden={panel !== 'live'}><LiveView /></section> : null}
          {mounted.resultats ? <section id="v-resultats" className="view" hidden={panel !== 'resultats'}><ResultatsView /></section> : null}
          {mounted.integrite ? <section id="v-integrite" className="view" hidden={panel !== 'integrite'}><IntegriteView /></section> : null}
        </main>

        <footer className="foot-ticker">
          <span id="ft-prov">source: {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
          <span aria-hidden="true" className="foot-sep">|</span>
          <span id="ft-run" className="mono">événement #{live?.event_id ?? '—'} · {live?.phase ?? 'idle'}</span>
          <span className="fill" />
          <span id="ft-sse" className="mono">SSE {sseStatus}</span>
        </footer>
        {settingsOpen && <SettingsDrawer onClose={() => setSettingsOpen(false)} />}
        <FlashBanner />
        <Toasts />
      </div>
    </ErrorBoundary>
  )
}
