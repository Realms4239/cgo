import { createPortal } from 'react-dom'
// Un seul popover, pas de visite guidée — 160×60, sans lib.
export function PeekPopover({ rect, children }: { rect: DOMRect; children: React.ReactNode }) {
  const left = rect.right + 168 > window.innerWidth ? rect.left - 168 : rect.right + 8
  const top = Math.min(rect.top, window.innerHeight - 68)
  return createPortal(
    <div
      style={{
        position: 'fixed',
        left,
        top,
        width: 160,
        height: 60,
        background: 'var(--surface-card)',
        border: '1px solid var(--hairline)',
        padding: 8,
        opacity: 0.98,
        backdropFilter: 'blur(8px)',
        pointerEvents: 'none',
        zIndex: 700,
        overflow: 'hidden',
        borderRadius: 6,
        boxShadow: '0 8px 24px rgba(0,0,0,0.25)',
      }}
    >
      {children}
    </div>,
    document.body,
  )
}
