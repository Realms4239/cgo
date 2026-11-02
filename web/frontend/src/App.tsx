import { lazy, Suspense, useEffect, useRef, useState } from 'react'
import { useUIStore, MODES, type ModeId } from './store/ui'
import { useAuditStore } from './store/audit'
import { connectSSE, disconnectSSE } from './lib/sse'
import { clearLive } from './lib/live'
import { animateViewEnter } from './lib/anime'
import { stageAfterCampaign, stageAfterReplay } from './lib/bench'
import { relativeTime } from './lib/format'
import DashboardView from './views/DashboardView'
import Toasts from './components/Toasts'
import FlashBanner from './components/FlashBanner'
import SettingsDrawer from './components/SettingsDrawer'
import ErrorBoundary from './components/ErrorBoundary'
import TopNav from './components/TopNav'
import QuickActionsPrompt from './components/QuickActionsPrompt'
import CommandPalette from './components/CommandPalette'
import MeteolinkWordmark from './components/MeteolinkWordmark'

// Seule la destination d'atterrissage est dans le premier rendu : le banc
// porte ECharts, l'audit et les preuves portent leurs propres panneaux.
const BenchView = lazy(() => import('./views/BenchView'))
const AuditView = lazy(() => import('./views/AuditView'))
const EvidenceView = lazy(() => import('./views/EvidenceView'))

function ViewFallback({ label }: { label: string }) {
  return <div className="card mode-loading mono muted">{label}</div>
}

function BenchFallback() {
  return <ViewFallback label="Chargement du banc d'essai…" />
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
  const [mounted, setMounted] = useState<Partial<Record<ModeId, boolean>>>({ dashboard: true })

  useEffect(() => {
    setMounted(m => (m[mode] ? m : { ...m, [mode]: true }))
  }, [mode])

  useEffect(() => {
    connectSSE()
    useAuditStore.getState().restore()
    return () => disconnectSSE()
  }, [])

  useEffect(() => {
    const found = MODES.find(m => m.id === mode)
    document.title = found ? `${found.label} — Meteolink` : 'Meteolink'
    const main = document.getElementById('main')
    main?.scrollTo({ top: 0 })
    // la destination est annoncée au clavier et aux lecteurs d'écran
    main?.focus({ preventScroll: true })
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
      const tag = (e.target as HTMLElement)?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
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
        <header className="app-header">
          <MeteolinkWordmark />
          <TopNav />
          <div className="hd-right">
            <span
              className="hd-conn"
              data-on={connected ? '1' : '0'}
              role="img"
              aria-label={connected ? 'Connexion établie' : 'Hors ligne'}
            />
            <span id="hd-state" className="mono">{live?.phase || 'au repos'}</span>
            <span className="hd-audit">dernier audit {relativeTime(report?.timestamp) || '—'}</span>
            <button onClick={() => setSettingsOpen(true)} aria-label="Réglages" title="Réglages">⚙</button>
          </div>
        </header>

        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        {live?.running && mode === 'bench' ? <QuickActionsPrompt /> : null}

        <main id="main" tabIndex={-1}>
          <section id="v-dashboard" className="view" hidden={mode !== 'dashboard'}><DashboardView /></section>
          {mounted.audit ? (
            <section id="v-audit" className="view" hidden={mode !== 'audit'}>
              <Suspense fallback={<ViewFallback label="Chargement de l’audit…" />}><AuditView /></Suspense>
            </section>
          ) : null}
          {mounted.bench ? (
            <section id="v-bench" className="view" hidden={mode !== 'bench'}>
              <Suspense fallback={<BenchFallback />}><BenchView /></Suspense>
            </section>
          ) : null}
          {mounted.evidence ? (
            <section id="v-evidence" className="view" hidden={mode !== 'evidence'}>
              <Suspense fallback={<ViewFallback label="Chargement des preuves…" />}><EvidenceView /></Suspense>
            </section>
          ) : null}
        </main>

        <footer className="foot-ticker">
          <span id="ft-prov">source : {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
          <span className="ft-brand mono">Meteolink</span>
        </footer>

        {settingsOpen ? <SettingsDrawer onClose={() => setSettingsOpen(false)} /> : null}
        <FlashBanner />
        <Toasts />
      </div>
    </ErrorBoundary>
  )
}
