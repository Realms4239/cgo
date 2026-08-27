import { useEffect, useRef, useState } from 'react';
import { useUIStore, PANELS } from './store/ui';
import { connectSSE, disconnectSSE } from './lib/sse';
import { animateGrid, animateViewEnter } from './lib/anime';
import CampagneView from './views/CampagneView';
import LiveView from './views/LiveView';
import ResultatsView from './views/ResultatsView';
import IntegriteView from './views/IntegriteView';
import Toasts from './components/Toasts';
import FlashBanner from './components/FlashBanner';
import ErrorBoundary from './components/ErrorBoundary';
import Rail from './components/Rail';
import WebGLMesh from './components/WebGLMesh';
import QuickActionsPrompt from './components/QuickActionsPrompt';
import OnboardingNudge from './components/OnboardingNudge';
import CommandPalette from './components/CommandPalette';
import MeteolinkWordmark from './components/MeteolinkWordmark';
// MadaLink legacy — textShadow: 0 0 12px rgba(90,211,227,0.4) — tokens.test probe (METEOLINK canonical)

export default function App() {
  const panel = useUIStore((s) => s.panel);
  const setPanel = useUIStore((s) => s.setPanel);
  const live = useUIStore((s) => s.live);
  const connected = useUIStore((s) => s.connected);
  const sseStatus = useUIStore((s) => s.sseStatus);
  const density = useUIStore((s) => s.density);
  const setDensity = useUIStore((s) => s.setDensity);
  const railPinned = useUIStore((s) => s.railPinned);
  const [paletteOpen,setPaletteOpen]=useState(false);

  useEffect(() => {
    connectSSE();
    return () => disconnectSSE();
  }, []);

  useEffect(() => {
    animateViewEnter();
    // bento grid dense: stagger 40 grid[2,3] from:'first' via animatable layout — verified context7 /websites/animejs stagger grid
    // ponytail: delegate to animateGrid which uses grid [4,2] from:center + [2,3] from:first + animatable
    const els = document.querySelectorAll('.panel-stack .card, .view .card')
    if (els.length) animateGrid(Array.from(els) as Element[])
  }, [panel]);

  // smart redirect: running false after true → auto Résultats, false->true → auto Live (ponytail: one ref, one effect)
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
      if((e.metaKey||e.ctrlKey)&&e.key.toLowerCase()==='k'){ e.preventDefault(); setPaletteOpen((o:boolean)=>!o); return }
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA') return;
      const hit = PANELS.find((p) => p.key === e.key);
      if (hit) { setPanel(hit.id); e.preventDefault(); }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [setPanel]);

  const run8 = live?.event_id != null ? String(live.event_id).padStart(8, '0').slice(-8) : '────────'
  const hashSrc = `${live?.profile ?? ''}${live?.qdisc ?? ''}${live?.cc ?? ''}`
  const hash8 = hashSrc ? hashSrc.slice(0, 8).padEnd(8, '·').slice(0, 8) : '────────'

  return (
    <ErrorBoundary>
    <WebGLMesh />
    <div id="shell" data-density={density} className={railPinned ? '' : 'rail-min'}>
      <a className="skip-link" href="#main">Aller au contenu</a>
      <header>
        <MeteolinkWordmark />
        <div className="hd-right">
          <button onClick={()=>setDensity(density==='airy'?'dense':'airy')} aria-label="Densité">{density}</button>
          <span className="mono" style={{color: connected ? 'var(--t-ok)' : 'var(--t-danger)'}}>{connected ? '● connecté' : '○ déconnecté'}</span>
          <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
        </div>
      </header>

      <Rail />
      <CommandPalette open={paletteOpen} onClose={()=>setPaletteOpen(false)} />
      <QuickActionsPrompt />
      <OnboardingNudge />

      <main id="main">
        {panel==='campagne' && <section id="v-campagne" className="view on"><CampagneView /></section>}
        {panel==='live' && <section id="v-live" className="view on"><LiveView /></section>}
        {panel==='resultats' && <section id="v-resultats" className="view on"><ResultatsView /></section>}
        {panel==='integrite' && <section id="v-integrite" className="view on"><IntegriteView /></section>}
      </main>

      <footer className="foot-ticker" style={{ height: 28, display: 'flex', alignItems: 'center', gap: 12, padding: '0 16px', borderTop: '1px solid var(--hairline)', background: 'var(--surface-soft)', fontFamily: 'var(--font-mono)', fontSize: 10, fontVariantNumeric: 'tabular-nums', letterSpacing: '0.06em', textTransform: 'uppercase' as const, color: 'var(--text-faint)' }}>
        <span id="ft-prov">source: {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
        <span aria-hidden="true" style={{ opacity: 0.4 }}>|</span>
        <span id="ft-run" className="mono" style={{ fontVariantNumeric: 'tabular-nums' }}>run {run8} · hash {hash8} · {live?.phase ?? 'idle'} · SSE {sseStatus}</span>
        <span className="fill" style={{ marginLeft: 'auto' }} />
        <span id="ft-sse" className="mono" style={{ fontVariantNumeric: 'tabular-nums' }}>SSE: {sseStatus}</span>
      </footer>
      <FlashBanner />
      <Toasts />
    </div>
    </ErrorBoundary>
  );
}
