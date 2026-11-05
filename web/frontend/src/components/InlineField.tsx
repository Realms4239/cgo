import { useEffect, useId, useRef, cloneElement, isValidElement } from 'react'
import { animateShake } from '../lib/anime'

export function InlineField({ label, error, helper, children }: { label: string; error?: string; helper?: string; children: any }) {
  const ref = useRef<HTMLDivElement>(null)
  const id = useId()
  useEffect(() => { if (error && ref.current) animateShake(ref.current) }, [error])
  const content = isValidElement(children)
    ? cloneElement(children as any, {
        id,
        style: {
          ...((children as any).props?.style || {}),
          ...(error ? { borderColor: 'var(--t-danger-text)' } : {}),
        },
      })
    : children
  return (
    <div ref={ref} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <label htmlFor={id} style={{ fontSize: 'var(--fs-micro)', fontFamily: 'var(--font-mono)', color: 'var(--text-muted)', letterSpacing: 'var(--tracking-caps)', textTransform: 'uppercase' as const }}>{label}</label>
      {content}
      {error ? <span style={{ color: 'var(--t-danger-text)', fontSize: 'var(--fs-tag)', fontFamily: 'var(--font-mono)' }}>{error}</span> : helper ? <span style={{ color: 'var(--text-faint)', fontSize: 'var(--fs-tag)', fontFamily: 'var(--font-mono)' }}>{helper}</span> : null}
    </div>
  )
}
