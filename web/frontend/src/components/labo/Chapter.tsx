// Un chapter est une section nommée du même document : numéro, titre, lede,
// puis le contenu. La lede dit ce que la section peut et ne peut pas affirmer.
import type { ReactNode } from 'react'

export interface ChapterProps {
  id: string
  n: string
  title: string
  lede: string
  children?: ReactNode
}

export default function Chapter({ id, n, title, lede, children }: ChapterProps) {
  return (
    <section className="chapter" id={`chapter-${id}`} aria-labelledby={`chapter-${id}-title`}>
      <div className="chapter-head">
        <span className="chapter-n mono" aria-hidden="true">{n}</span>
        <div className="chapter-heading">
          <h2 id={`chapter-${id}-title`} className="chapter-title">{title}</h2>
          <p className="chapter-lede">{lede}</p>
        </div>
      </div>
      <div className="chapter-body">{children}</div>
    </section>
  )
}
