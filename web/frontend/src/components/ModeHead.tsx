// En-tête d'une destination — même rythme, même mesure, mêmes actions.
import type { ReactNode } from 'react'

export function ModeHead({ title, sub, actions }: { title: string; sub?: string; actions?: ReactNode }) {
  return (
    <header className="mode-head">
      <div>
        <h1>{title}</h1>
        {sub ? <p>{sub}</p> : null}
      </div>
      {actions ? <div className="mode-head-actions">{actions}</div> : null}
    </header>
  )
}
