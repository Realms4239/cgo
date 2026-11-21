// Un chapitre est un panneau de la même méthode : seul le panneau actif est
// exposé au clavier et aux lecteurs d'écran ; les autres sont inertes.
import type { ReactNode } from 'react'

export interface ChapterProps {
  id: string
  n: string
  title: string
  lede: string
  active: boolean
  children?: ReactNode
}

export default function Chapter({ id, n, title, lede, active, children }: ChapterProps) {
  return (
    <section className="chapter" id={`chapter-${id}`} aria-labelledby={`chapter-${id}-title`}
      data-active={active ? 'true' : 'false'} inert={!active} aria-hidden={active ? undefined : true}>
      <div className="chapter-head">
        <span className="chapter-n mono" aria-hidden="true">{n}</span>
        <div className="chapter-heading">
          <h2 id={`chapter-${id}-title`} className="chapter-title" tabIndex={-1}>{title}</h2>
          <p className="chapter-lede">{lede}</p>
        </div>
      </div>
      <div className="chapter-body">{children}</div>
    </section>
  )
}
