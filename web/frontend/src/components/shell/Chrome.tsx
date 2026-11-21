import MeteolinkWordmark from '../MeteolinkWordmark'
import ModeSwitch from './ModeSwitch'

// Coque 56 px : identité, bascule des deux modes, palette et réglages.
export default function Chrome({
  onPalette, onSettings,
}: { onPalette: () => void; onSettings: () => void }) {
  return (
    <header className="app-chrome">
      <MeteolinkWordmark />
      <ModeSwitch />
      <div className="hd-right">
        <button className="chrome-trigger" onClick={onPalette} aria-label="Palette de commandes" title="Palette de commandes (Cmd-K)">⌘K</button>
        <button className="chrome-trigger" onClick={onSettings} aria-label="Réglages" title="Réglages">⚙</button>
      </div>
    </header>
  )
}
