// Une seule navigation — deux destinations, jamais de sous-onglets.
import { MODES, useUIStore } from '../store/ui'
import { preloadMode } from '../lib/preload'

export default function TopNav() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)
  const connected = useUIStore(s => s.connected)

  return (
    <>
      {/* Bandeau utilitaire : contexte réel du poste, jamais une adresse inventée. */}
      <div className="util-strip" aria-hidden="true">
        <span>Météorologie · suivi du trafic critique · Madagascar</span>
        <span>{connected ? 'Liaison active' : 'Liaison en attente'}</span>
      </div>
      <nav className="topnav" aria-label="Destinations">
        {MODES.map(m => (
          <button
            key={m.id}
            type="button"
            className={'topnav-btn' + (mode === m.id ? ' on' : '')}
            data-mode={m.id}
            aria-current={mode === m.id ? 'page' : undefined}
            title={`${m.label} — touche ${m.key}`}
            onClick={() => setMode(m.id)}
            onMouseEnter={() => { preloadMode(m.id) }}
            onFocus={() => { preloadMode(m.id) }}
          >
            <span className="topnav-lbl">{m.label}</span>
            <kbd className="topnav-key">{m.key}</kbd>
          </button>
        ))}
      </nav>
    </>
  )
}
