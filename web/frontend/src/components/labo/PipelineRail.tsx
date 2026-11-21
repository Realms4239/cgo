// Rail de lecture du laboratoire : quatre étapes d'une même méthode, pas des
// ancres dans un long document. Chaque clic amène son panneau plein écran.
import { useRef, type KeyboardEvent } from 'react'

export const CHAPTERS = [
  { id: 'configurer', n: '01', label: 'Configurer' },
  { id: 'observer', n: '02', label: 'Observer' },
  { id: 'comparer', n: '03', label: 'Comparer' },
  { id: 'preuves', n: '04', label: 'Preuves' },
] as const

interface PipelineRailProps {
  active: number
  onSelect: (index: number) => void
}

export default function PipelineRail({ active, onSelect }: PipelineRailProps) {
  const buttons = useRef<(HTMLButtonElement | null)[]>([])

  // Modèle clavier W3C emprunté aux onglets sans en adopter la sémantique :
  // flèches et Début/Fin déplacent le focus dans le rail ; Entrée/Espace
  // activent (clic natif), puis le focus rejoint le titre du panneau.
  // Le tabIndex reste sur l'étape active : au Tab suivant, le focus revient délibérément au pas courant.
  function onRailKeyDown(e: KeyboardEvent<HTMLButtonElement>, i: number) {
    let next: number | null = null
    if (e.key === 'ArrowRight') next = (i + 1) % CHAPTERS.length
    else if (e.key === 'ArrowLeft') next = (i - 1 + CHAPTERS.length) % CHAPTERS.length
    else if (e.key === 'Home') next = 0
    else if (e.key === 'End') next = CHAPTERS.length - 1
    if (next === null) return
    e.preventDefault()
    buttons.current[next]?.focus({ preventScroll: true })
  }

  return (
    <nav className="pipeline-rail" aria-label="Étapes du laboratoire">
      <p className="pr-progress" aria-live="polite">Étape {active + 1} sur {CHAPTERS.length}</p>
      {CHAPTERS.map((c, i) => (
        <button
          key={c.id}
          ref={el => { buttons.current[i] = el }}
          type="button"
          className={'pr-step' + (active === i ? ' on' : '')}
          aria-current={active === i ? 'step' : undefined}
          tabIndex={active === i ? 0 : -1}
          onKeyDown={e => onRailKeyDown(e, i)}
          onClick={() => onSelect(i)}
        >
          <span className="pr-marker" aria-hidden="true" data-current={active === i ? 'true' : 'false'} />
          <span className="pr-n mono" aria-hidden="true">{c.n}</span>
          <span className="pr-label">{c.label}</span>
          {active === i ? <span className="pr-state">en cours</span> : null}
        </button>
      ))}
    </nav>
  )
}
