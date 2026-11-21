import { MODES, useUIStore } from '../../store/ui'
import { preloadMode } from '../../lib/preload'

// Deux destinations, une seule bascule : le mot et la forme portent l'état,
// la couleur ne fait que l'accompagner.
export default function ModeSwitch() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)

  return (
    <nav className="modeswitch" aria-label="Modes de la console">
      {MODES.map(m => {
        const on = mode === m.id
        return (
          <button
            key={m.id}
            type="button"
            className={'modeswitch-btn' + (on ? ' on' : '')}
            data-mode={m.id}
            aria-current={on ? 'page' : undefined}
            title={`${m.label} — touche ${m.key}`}
            style={{ minHeight: 48 }}
            onClick={() => setMode(m.id)}
            onMouseEnter={() => { preloadMode(m.id) }}
            onFocus={() => { preloadMode(m.id) }}
          >
            <span className={'modeswitch-shape modeswitch-shape-' + m.id} aria-hidden="true" />
            <span className="modeswitch-lbl">{m.label}</span>
            <kbd className="modeswitch-key">{m.key}</kbd>
            {on ? <span className="sr-only"> (actif)</span> : null}
          </button>
        )
      })}
    </nav>
  )
}
