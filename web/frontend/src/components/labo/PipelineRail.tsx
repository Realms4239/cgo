// Rail de lecture du laboratoire : quatre étapes d'une même méthode, pas des
// ancres dans un long document. Chaque clic amène son panneau plein écran.
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
  return (
    <nav className="pipeline-rail" aria-label="Étapes du laboratoire">
      {CHAPTERS.map((c, i) => (
        <button
          key={c.id}
          type="button"
          className={'pr-step' + (active === i ? ' on' : '')}
          aria-current={active === i ? 'step' : undefined}
          onClick={() => onSelect(i)}
        >
          <span className="pr-n mono" aria-hidden="true">{c.n}</span>
          <span className="pr-label">{c.label}</span>
        </button>
      ))}
    </nav>
  )
}
