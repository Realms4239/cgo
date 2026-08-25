import { useEffect, useRef, useState } from 'react';
import { useUIStore, PANELS } from './store/ui';
import { connectSSE, disconnectSSE } from './lib/sse';
import { animateViewEnter } from './lib/anime';
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

export default function App() {
  const panel = useUIStore((s) => s.panel);
  const setPanel = useUIStore((s) => s.setPanel);
  const live = useUIStore((s) => s.live);
  const connected = useUIStore((s) => s.connected);
  const sseStatus = useUIStore((s) => s.sseStatus);
  const density = useUIStore((s) => s.density);
  const railPinned = useUIStore((s) => s.railPinned);
  const [paletteOpen,setPaletteOpen]=useState(false);

  useEffect(() => {
    connectSSE();
    return () => disconnectSSE();
  }, []);

  useEffect(() => {
    animateViewEnter();
  }, [panel]);

  // smart redirect: running false after true → auto Résultats
  const wasRunningRef = useRef(!!live?.running)
  useEffect(() => {
    const running = !!live?.running
    if (wasRunningRef.current && !running) {
      setPanel('resultats')
      useUIStore.getState().setFlash({ type: 'success', msg: 'Campagne terminée' })
      useUIStore.getState().pushToast('Campagne terminée', 'ok')
    }
    wasRunningRef.current = running
  }, [live?.running, setPanel])

  // smart redirect: Démarrer success → auto Live (wasRunning false -> true)
  const wasRunning2Ref = useRef(!!live?.running)
  useEffect(() => {
    const running = !!live?.running
    if (!wasRunning2Ref.current && running) {
      const cur = useUIStore.getState().panel
      if (cur === 'campagne') {
        setPanel('live')
        useUIStore.getState().setFlash({ type: 'success', msg: 'CHARGE — campagne lancée' })
      }
    }
    wasRunning2Ref.current = running
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

  return (
    <ErrorBoundary>
    <WebGLMesh />
    <div id="shell" data-density={density} className={railPinned ? '' : 'rail-min'}>
      <a className="skip-link" href="#main">Aller au contenu</a>
      <header>
        <span className="stripe" aria-hidden="true"><i /><i /><i /></span>
        <span className="wordmark" style={{textShadow: '0 0 12px rgba(90,211,227,0.4)'}}>MadaLink</span>
        <span className="hd-sub">LIEN — trafic critique · AQM/BBR · Mada</span>
        <div className="hd-right">
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

      <footer>
        <span id="ft-prov">source : {live?.profile || '—'} · {live?.qdisc || '—'} · {live?.cc || '—'}</span>
        <span className="fill" />
        <span id="ft-sse" className="mono">SSE : {sseStatus}</span>
      </footer>
      <FlashBanner />
      <Toasts />
    </div>
    </ErrorBoundary>
  );
}
