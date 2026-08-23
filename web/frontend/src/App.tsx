import { useEffect } from 'react';
import { useUIStore, PANELS } from './store/ui';
import { connectSSE, disconnectSSE } from './lib/sse';
import CampagneView from './views/CampagneView';
import LiveView from './views/LiveView';

export default function App() {
  const panel = useUIStore((s) => s.panel);
  const setPanel = useUIStore((s) => s.setPanel);
  const live = useUIStore((s) => s.live);
  const connected = useUIStore((s) => s.connected);
  const sseStatus = useUIStore((s) => s.sseStatus);

  useEffect(() => {
    connectSSE();
    return () => disconnectSSE();
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA') return;
      const hit = PANELS.find((p) => p.key === e.key);
      if (hit) { setPanel(hit.id); e.preventDefault(); }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [setPanel]);

  return (
    <div id="shell">
      <a className="skip-link" href="#main">Aller au contenu</a>
      <header>
        <span className="stripe" aria-hidden="true"><i /><i /><i /></span>
        <span className="wordmark">CGO</span>
        <span className="hd-sub">LIEN — trafic critique · AQM/BBR</span>
        <div className="hd-right">
          <span className="mono" style={{color: connected ? 'var(--t-ok)' : 'var(--t-danger)'}}>{connected ? '● connecté' : '○ déconnecté'}</span>
          <span id="hd-state" className="mono">{live?.phase ?? 'idle'}</span>
        </div>
      </header>

      <aside className="sidebar" aria-label="Navigation">
        {PANELS.map((p) => (
          <button key={p.id} className={'nav-btn' + (panel === p.id ? ' on' : '')} data-panel={p.id} onClick={() => setPanel(p.id)}>
            <span className="nav-lbl">{p.label}</span>
            <span className="nav-key">{p.key}</span>
          </button>
        ))}
        <div className="side-status">
          <div><span>phase</span><b className="mono">{live?.phase ?? '—'}</b></div>
          <div><span>événement</span><b className="mono">{live ? `#${live.event_id}` : '—'}</b></div>
          <div><span>SSE</span><b className="mono">{sseStatus}</b></div>
          <div className="gates" role="img" aria-label="Portes G0 à G7">
            {(live?.gates ?? Array(8).fill(null)).map((g: boolean|null, i:number) => (
              <span key={i} className={'gate ' + (g===null?'':g?'ok':'fail')} data-gate={i} title={'G'+i} />
            ))}
          </div>
        </div>
      </aside>

      <main id="main">
        <section id="v-campagne" className={'view' + (panel==='campagne'?' on':'')}><CampagneView /></section>
        <section id="v-live" className={'view' + (panel==='live'?' on':'')}><LiveView /></section>
        <section id="v-resultats" className={'view' + (panel==='resultats'?' on':'')}>
          <h1 className="view-title">Résultats</h1>
          <div className="card"><p className="mono muted">disponible après gel — jalon M2</p></div>
        </section>
        <section id="v-integrite" className={'view' + (panel==='integrite'?' on':'')}>
          <h1 className="view-title">Intégrité</h1>
          <div className="card"><p className="mono muted">disponible après gel — jalon M2</p></div>
        </section>
      </main>

      <footer>
        <span id="ft-prov">source : {live?.profile ?? '—'} · {live?.qdisc ?? '—'} · {live?.cc ?? '—'}</span>
        <span className="fill" />
        <span id="ft-sse" className="mono">SSE : {sseStatus}</span>
      </footer>
    </div>
  );
}
