// InlineDisclosure — le détail pousse le contenu dans le flux, il ne
// recouvre rien. L'état est dit par aria-expanded et par la forme du signe.
import { useState, type ReactNode } from 'react'

export function InlineDisclosure({ label, children }: { label: string; children?: ReactNode }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="inline-disclosure">
      <button
        type="button"
        className="inline-disclosure-trigger"
        aria-expanded={open}
        onClick={() => setOpen(v => !v)}
      >
        <span className="inline-disclosure-mark" aria-hidden="true">{open ? '▾' : '▸'}</span>
        <span>{label}</span>
      </button>
      {open ? <div className="inline-disclosure-body">{children}</div> : null}
    </div>
  )
}
