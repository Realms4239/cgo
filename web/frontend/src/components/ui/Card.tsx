// Card — the single surface shell: hairline border, card surface, 12 padding.
// Views must not re-roll borders/backgrounds inline (NOC rebuild, tokens only).
import { CardHead } from './CardHead'

export function Card({ head, sub, right, children, testid, style, className, onMouseEnter, onMouseLeave }: {
  head?: string
  sub?: string
  right?: React.ReactNode
  children: React.ReactNode
  testid?: string
  style?: React.CSSProperties
  className?: string
  onMouseEnter?: (e: React.MouseEvent<HTMLDivElement>) => void
  onMouseLeave?: (e: React.MouseEvent<HTMLDivElement>) => void
}) {
  return (
    <div
      data-testid={testid}
      className={'card' + (className ? ' ' + className : '')}
      style={{ border: '1px solid var(--hairline, #26262a)', background: 'var(--surface-card, #101012)', padding: 12, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6, ...style }}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      {(head || right) && <CardHead label={head ?? ''} sub={sub} right={right} />}
      {children}
    </div>
  )
}
