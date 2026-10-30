// Rail des quatre destinations — une seule source, jamais de sous-onglets.
import { MODES, useUIStore } from '../store/ui'
import { preloadBench } from '../lib/preload'
import { NavIcon } from './NavIcon'

export default function ModeRail() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)
  const pinned = useUIStore(s => s.railPinned)
  const setPinned = useUIStore(s => s.setRailPinned)

  return (
    <aside className={'mode-rail' + (pinned ? ' pinned' : '')} aria-label="Navigation par destination">
      {MODES.map(m => (
        <button
          key={m.id}
          type="button"
          className={'mode-btn' + (mode === m.id ? ' on' : '')}
          data-mode={m.id}
          aria-pressed={mode === m.id}
          aria-current={mode === m.id ? 'page' : undefined}
          aria-label={m.label}
          title={`${m.label} — touche ${m.key}`}
          onClick={() => setMode(m.id)}
          onMouseEnter={m.id === 'bench' ? () => { void preloadBench() } : undefined}
          onFocus={m.id === 'bench' ? () => { void preloadBench() } : undefined}
        >
          <span className="mode-icon" aria-hidden="true"><NavIcon id={m.id} /></span>
          {pinned ? <span className="mode-lbl">{m.label}</span> : <span className="sr-only">{m.label}</span>}
          {pinned ? <span className="mode-key">{m.key}</span> : null}
        </button>
      ))}
      <button
        type="button"
        className="mode-pin"
        aria-label={pinned ? 'Réduire le rail' : 'Épingler le rail'}
        aria-pressed={pinned}
        onClick={() => setPinned(!pinned)}
      >
        {pinned ? '◀' : '▶'}
      </button>
      {pinned ? <div className="mode-foot mono">1–4</div> : null}
    </aside>
  )
}
