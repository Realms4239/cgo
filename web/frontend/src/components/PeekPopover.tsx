import { createPortal } from 'react-dom'
// ponytail: single popover not tour — 160×60 fixed portal, no lib
export function PeekPopover({ rect, children }: { rect: DOMRect; children: React.ReactNode }) {
  return createPortal(
    <div
      style={{
        position: 'fixed',
        left: rect.right + 8,
        top: rect.top,
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
