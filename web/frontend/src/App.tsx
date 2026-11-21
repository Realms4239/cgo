import { lazy, Suspense, useEffect, useRef, useState } from 'react'
import { useUIStore, MODES, type ModeId } from './store/ui'
import { useAuditStore } from './store/audit'
import { connectSSE, disconnectSSE } from './lib/sse'
import { clearLive } from './lib/live'
import { animateViewEnter } from './lib/anime'
import { isTextEntryTarget } from './lib/modalFocus'
import { stageAfterCampaign, stageAfterReplay } from './lib/bench'
import { relativeTime } from './lib/format'
import CockpitView from './views/CockpitView'
import Toasts from './components/Toasts'
import FlashBanner from './components/FlashBanner'
import SettingsDrawer from './components/SettingsDrawer'
import ProofDrawer from './components/ProofDrawer'
import ErrorBoundary from './components/ErrorBoundary'
import CommandPalette from './components/CommandPalette'
import Chrome from './components/shell/Chrome'
import TrustStrip from './components/shell/TrustStrip'

// Seule la destination d'atterrissage est dans le premier rendu : le laboratoire
// porte l'instrument (ECharts + boucle rAF) et ne se monte qu'à la visite.
const LaboView = lazy(() => import('./views/LaboView'))

function LaboFallback() {
  return <div className="card mode-loading mono muted">Chargement du laboratoire…</div>
}

export default function App() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)
  const live = useUIStore(s => s.live)
  const connected = useUIStore(s => s.connected)
  const replayRunning = useUIStore(s => s.replayRunning)
  const report = useAuditStore(s => s.report)
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  // une destination lourde ne se monte qu'à la première visite, puis reste.
  const [mounted, setMounted] = useState<Partial<Record<ModeId, boolean>>>({ cockpit: true })

  useEffect(() => {
    setMounted(m => (m[mode] ? m : { ...m, [mode]: true }))
  }, [mode])

  useEffect(() => {
    connectSSE()
    useAuditStore.getState().restore()
    return () => disconnectSSE()
  }, [])

  const isInitialMode = useRef(true)
  useEffect(() => {
    const found = MODES.find(m => m.id === mode)
    document.title = found ? `${found.label} — Meteolink` : 'Meteolink'
    const main = document.getElementById('main')
    main?.scrollTo({ top: 0 })
    if (isInitialMode.current) {
      isInitialMode.current = false
    } else {
      // la destination est annoncée au clavier et aux lecteurs d'écran
      main?.focus({ preventScroll: true })
    }
    animateViewEnter()
  }, [mode])

  const wasRunningRef = useRef(!!live?.running)
  useEffect(() => {
    const running = !!live?.running
    const store = useUIStore.getState()
    if (wasRunningRef.current && !running) {
      // fantôme d'après-arrêt : les anneaux et la présentation retombent.
      clearLive()
      const last = store.live
      if (last) {
        store.setLive({
          ...last, phase: 'idle', running: false, load_status: '',
          profile: '', qdisc: '', cc: '', repetition: 0,
          rtt_p50_ms: 0, rtt_p95_ms: 0, small_p95_ms: 0, bulk_goodput_mbps: 0,
          drops: 0, wasted_bytes: undefined, cost_ar_per_h: undefined,
          deadline_ok_pct: undefined, gates: Array(8).fill(null),
        } as typeof last)
      }
      store.setBenchStage(stageAfterCampaign(store.benchStage, false))
      store.setFlash({ type: 'success', msg: 'Campagne terminée' })
      store.pushToast('Campagne terminée', 'ok')
    }
    if (!wasRunningRef.current && running) {
      store.setBenchStage(stageAfterCampaign(store.benchStage, true))
      store.setFlash({ type: 'success', msg: 'CHARGE — campagne lancée' })
    }
    wasRunningRef.current = running
  }, [live?.running])

  useEffect(() => {
    if (!replayRunning) return
    const store = useUIStore.getState()
    store.setBenchStage(stageAfterReplay(store.benchStage, true))
  }, [replayRunning])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        setPaletteOpen(o => !o)
        return
      }
      if (isTextEntryTarget(e.target)) return
      const hit = MODES.find(m => m.key === e.key)
      if (hit) {
        setMode(hit.id)
        e.preventDefault()
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [setMode])

  return (
    <ErrorBoundary>
      <div id="shell">
        <a className="skip-link" href="#main">Aller au contenu</a>
        <Chrome
          onPalette={() => setPaletteOpen(o => !o)}
          onSettings={() => setSettingsOpen(true)}
        />

        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />

        <main id="main" tabIndex={-1} className="layout-shell">
          <section id="v-cockpit" className="view" hidden={mode !== 'cockpit'}><CockpitView /></section>
          {mounted.labo ? (
            <section id="v-labo" className="view" hidden={mode !== 'labo'}>
              <Suspense fallback={<LaboFallback />}><LaboView /></Suspense>
            </section>
          ) : null}
        </main>

        <TrustStrip
          connection={connected ? 'online' : 'offline'}
          source={live?.profile ?? null}
          freshness={live && Number.isFinite(new Date(live.ts).getTime()) ? relativeTime(new Date(live.ts).toISOString()) : null}
          seal={report?.audit_id ?? null}
          backpressure={live ? live.drops : undefined}
        />

        {settingsOpen ? <SettingsDrawer onClose={() => setSettingsOpen(false)} /> : null}
        <ProofDrawer />
        <FlashBanner />
        <Toasts />
      </div>
    </ErrorBoundary>
  )
}
