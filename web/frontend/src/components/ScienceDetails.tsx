// Le second niveau de lecture : tout le chiffre scientifique, replié par
// défaut. L'opérateur lit l'écran ; le jury ouvre ce tiroir.
import { useState, type ReactNode } from 'react'

export function ScienceDetails({
  children,
  label = 'Détails scientifiques',
  defaultOpen = false,
  testid,
}: {
  children: ReactNode
  label?: string
  defaultOpen?: boolean
  testid?: string
}) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <div className="sci-details" data-testid={testid}>
      <button
        type="button"
        className="btn sci-details-toggle"
        aria-expanded={open}
        onClick={() => setOpen(o => !o)}
      >
        {open ? 'MASQUER' : 'AFFICHER'} — {label.toUpperCase()}
      </button>
      {open ? <div className="sci-details-body">{children}</div> : null}
    </div>
  )
}
