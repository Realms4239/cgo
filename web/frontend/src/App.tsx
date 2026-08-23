import { useEffect } from 'react';
import { useUIStore, PANELS } from './store/ui';

export default function App() {
  const panel = useUIStore((s) => s.panel);
  const setPanel = useUIStore((s) => s.setPanel);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA') return;
      const hit = PANELS.find((p) => p.key === e.key);
      if (hit) {
        setPanel(hit.id);
        e.preventDefault();
      }
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
        <div className="hd-right"><span id="hd-state">au repos</span></div>
      </header>

      <aside className="sidebar" aria-label="Navigation">
        {PANELS.map((p) => (
          <button
            key={p.id}
            className={'nav-btn' + (panel === p.id ? ' on' : '')}
            data-panel={p.id}
            onClick={() => setPanel(p.id)}
          >
            <span className="nav-lbl">{p.label}</span>
            <span className="nav-key">{p.key}</span>
          </button>
        ))}
        <div className="side-status" aria-label="État campagne et portes G0–G7">
          <div><span>campagne</span><b id="st-campagne">—</b></div>
          <div><span>événement</span><b id="st-event">—</b></div>
          <div className="gates" role="img" aria-label="Portes G0 à G7">
            {Array.from({ length: 8 }, (_, i) => (
              <span key={i} className="gate" data-gate={i} title={'G' + i} />
            ))}
          </div>
        </div>
      </aside>

      <main id="main">
        {PANELS.map((p) => (
          <section key={p.id} id={'v-' + p.id} className={'view' + (panel === p.id ? ' on' : '')}>
            <h1 className="view-title">{p.label}</h1>
            <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-cap)' }}>
              panneau en construction — jalon M{p.milestone}
            </p>
          </section>
        ))}
      </main>

      <footer>
        <span id="ft-prov">source : —</span>
        <span className="fill" />
        <span id="ft-sse">SSE : déconnecté</span>
      </footer>
    </div>
  );
}
