// Une seule navigation — quatre destinations, jamais de sous-onglets.
import { MODES, useUIStore } from '../store/ui'
import { preloadMode } from '../lib/preload'

export default function TopNav() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)

  return (
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
  )
}
