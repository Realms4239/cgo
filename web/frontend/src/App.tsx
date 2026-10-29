import { useEffect, useRef, useState } from 'react'
import { useUIStore, MODES, type ModeId } from './store/ui'
import { useAuditStore } from './store/audit'
import { connectSSE, disconnectSSE } from './lib/sse'
import { clearLive } from './lib/live'
import { animateViewEnter } from './lib/anime'
import { stageAfterCampaign, stageAfterReplay } from './lib/bench'
import DashboardView from './views/DashboardView'
import AuditView from './views/AuditView'
import BenchView from './views/BenchView'
import EvidenceView from './views/EvidenceView'
import Toasts from './components/Toasts'
import FlashBanner from './components/FlashBanner'
import SettingsDrawer from './components/SettingsDrawer'
import ErrorBoundary from './components/ErrorBoundary'
import ModeRail from './components/ModeRail'
import StatusRibbon from './components/StatusRibbon'
import QuickActionsPrompt from './components/QuickActionsPrompt'
import CommandPalette from './components/CommandPalette'
import MeteolinkWordmark from './components/MeteolinkWordmark'

export default function App() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)
  const live = useUIStore(s => s.live)
  const replayRunning = useUIStore(s => s.replayRunning)
  const railPinned = useUIStore(s => s.railPinned)
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
      <div id="shell" className={railPinned ? '' : 'rail-min'}>
        <a className="skip-link" href="#main">Aller au contenu</a>
        <header className="app-header">
          <MeteolinkWordmark />
          <div className="app-header-center">
            <StatusRibbon />
          </div>
          <div className="hd-right">
            <button onClick={() => setSettingsOpen(true)} aria-label="Réglages" title="Réglages">⚙</button>
            <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
          </div>
        </header>

        <ModeRail />
        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        {live?.running && mode === 'bench' ? <QuickActionsPrompt /> : null}

        <main id="main">
          <section id="v-dashboard" className="view" hidden={mode !== 'dashboard'}><DashboardView /></section>
          {mounted.audit ? <section id="v-audit" className="view" hidden={mode !== 'audit'}><AuditView /></section> : null}
          {mounted.bench ? <section id="v-bench" className="view" hidden={mode !== 'bench'}><BenchView /></section> : null}
          {mounted.evidence ? <section id="v-evidence" className="view" hidden={mode !== 'evidence'}><EvidenceView /></section> : null}
        </main>

        <footer className="foot-ticker">
          <span id="ft-prov">source : {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
          <span aria-hidden="true" className="foot-sep">|</span>
          <span id="ft-run" className="mono">événement #{live?.event_id ?? '—'} · {live?.phase ?? 'idle'}</span>
          <span className="fill" />
          <span className="mono">Meteolink</span>
        </footer>

        {settingsOpen ? <SettingsDrawer onClose={() => setSettingsOpen(false)} /> : null}
        <FlashBanner />
        <Toasts />
      </div>
    </ErrorBoundary>
  )
}
