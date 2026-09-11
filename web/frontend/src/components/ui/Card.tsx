// Card — coque unique: bordure hairline, fond carte, marge 12.
// Les vues ne redéfinissent ni bordures ni fonds en inline (tokens uniquement).
import { CardHead } from './CardHead'
import Explain from '../Explain'

export function Card({ head, sub, right, children, testid, term, style, className, onMouseEnter, onMouseLeave }: {
  head?: string
  sub?: string
  right?: React.ReactNode
  children: React.ReactNode
  testid?: string
  // terme du dictionnaire explain.ts — la tête gagne son ⓘ une-ligne
  term?: string
  style?: React.CSSProperties
  className?: string
  onMouseEnter?: (e: React.MouseEvent<HTMLDivElement>) => void
  onMouseLeave?: (e: React.MouseEvent<HTMLDivElement>) => void
}) {
  const headNode = head && term ? <Explain term={term}>{head}</Explain> : head
  return (
    <div
      data-testid={testid}
      className={'card' + (className ? ' ' + className : '')}
      style={{ border: '1px solid var(--hairline, #26262a)', background: 'var(--surface-card, #101012)', padding: 12, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6, ...style }}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      {(headNode || right) && <CardHead label={headNode ?? ''} sub={sub} right={right} />}
      {children}
    </div>
  )
}
