import { useEffect, useRef, cloneElement, isValidElement } from 'react'
import { animateShake } from '../lib/anime'

export function InlineField({ label, error, helper, children }: { label: string; error?: string; helper?: string; children: any }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => { if (error && ref.current) animateShake(ref.current) }, [error])
  const content = isValidElement(children)
    ? cloneElement(children as any, {
        style: {
          ...((children as any).props?.style || {}),
          ...(error ? { borderColor: 'var(--t-danger-text, #e84a3a)' } : {}),
        },
      })
    : children
  return (
    <div ref={ref} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <label style={{ fontSize: 10, fontFamily: 'JetBrains Mono', color: 'var(--text-muted)', letterSpacing: '0.06em', textTransform: 'uppercase' as const }}>{label}</label>
      {content}
      {error ? <span style={{ color: 'var(--t-danger-text, #e84a3a)', fontSize: 11, fontFamily: 'JetBrains Mono' }}>{error}</span> : helper ? <span style={{ color: 'var(--text-faint)', fontSize: 11, fontFamily: 'JetBrains Mono' }}>{helper}</span> : null}
    </div>
  )
}
