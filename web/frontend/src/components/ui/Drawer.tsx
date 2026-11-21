// Drawer — profondeur à droite : le panneau porte le focus, Échap ne ferme
// que le tiroir du dessus, l'arrière-plan reste inerte jusqu'au dernier.
// createPortal hors de #shell : sinon le panneau s'inertait lui-même.
import { useRef, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { useModalFocus } from '../../lib/modalFocus'

export interface DrawerProps {
  open: boolean
  onClose: () => void
  title: string
  width?: string
  children?: ReactNode
}

export function Drawer({ open, onClose, title, width = '420px', children }: DrawerProps) {
  const panelRef = useRef<HTMLDivElement>(null)
  useModalFocus(open, onClose, panelRef)
  if (!open) return null

  return createPortal(
    <div className="ui-drawer">
      <div
        ref={panelRef}
        className="ui-drawer-panel"
        role="dialog"
        aria-modal="true"
        aria-label={title}
        style={{ width, maxWidth: 'calc(100vw - var(--ui-spine-w, 320px))' }}
      >
        <div className="ui-drawer-head">
          <h2 className="ui-drawer-title">{title}</h2>
          <button type="button" className="ui-drawer-close" onClick={onClose} aria-label="Fermer">FERMER</button>
        </div>
        <div className="ui-drawer-body">{children}</div>
      </div>
    </div>,
    document.body,
  )
}
