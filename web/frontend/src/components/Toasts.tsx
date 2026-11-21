// Notifications : le genre est dit par un mot, l'emplacement par un jeton.
// Aucune couleur n'est écrite ici — palette.ts est la seule maison.
import { useEffect, useRef } from 'react'
import { useUIStore } from '../store/ui'
import { animateToasts } from '../lib/anime'
import { KIND_TOKEN, TOKEN } from '../lib/palette'

const KIND_WORD: Record<string, string> = {
  ok: 'Succès', err: 'Échec', blue: 'Info', '': 'Info',
}

export default function Toasts() {
  const toasts = useUIStore((s: any) => s.toasts)
  const dropToast = useUIStore((s: any) => s.dropToast)
  const pauseToasts = useUIStore((s: any) => s.pauseToasts)
  const resumeToasts = useUIStore((s: any) => s.resumeToasts)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    const els = Array.from(ref.current.querySelectorAll('.toast'))
    if (els.length) animateToasts(els as Element[])
  }, [toasts.length])
  return (
    <div
      ref={ref}
      id="toasts"
      aria-live="polite"
      onMouseEnter={pauseToasts}
      onMouseLeave={resumeToasts}
      style={{
        position: 'fixed', bottom: 'var(--sp-4)', right: 'var(--sp-4)', zIndex: 500,
        display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)',
      }}
    >
      {toasts.map((t: any) => {
        const accent = TOKEN[KIND_TOKEN[t.cls as keyof typeof KIND_TOKEN] ?? ''] ?? TOKEN.textMuted
        return (
          <div
            key={t.id}
            className={'toast ' + (t.cls || '')}
            style={{
              background: TOKEN.surfaceElevated,
              border: `1px solid ${TOKEN.hairline}`,
              borderLeft: `3px solid ${accent}`,
              padding: 'var(--sp-2) var(--sp-3)',
              fontFamily: 'var(--font-mono)',
              fontSize: 'var(--fs-label)',
              color: TOKEN.textBody,
              display: 'flex', alignItems: 'center', gap: 'var(--sp-3)',
              minWidth: 260, maxWidth: 360,
            }}
          >
            <span className="toast-word" style={{ color: accent, fontWeight: 700, flexShrink: 0 }}>
              {KIND_WORD[t.cls] ?? 'Info'}
            </span>
            <span className="toast-msg" style={{ overflowWrap: 'anywhere' }}>{t.msg}</span>
            <button
              type="button"
              className="toast-dismiss"
              aria-label="Fermer la notification"
              onClick={() => dropToast(t.id)}
              style={{
                minWidth: 44, minHeight: 44, marginLeft: 'auto', flexShrink: 0,
                background: 'transparent', border: 'none', color: TOKEN.textMuted,
                fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-body)', cursor: 'pointer',
              }}
            >
              ✕
            </button>
          </div>
        )
      })}
    </div>
  )
}
