// Rail de lecture : quatre chapters dans l'ordre du geste. Ce sont des ancres
// dans le flux, pas des onglets — rien ne masque le reste du document.
const CHAPTERS = [
  { id: 'configurer', n: '01', label: 'Configurer' },
  { id: 'observer', n: '02', label: 'Observer' },
  { id: 'comparer', n: '03', label: 'Comparer' },
  { id: 'preuves', n: '04', label: 'Preuves' },
] as const

export default function PipelineRail() {
  return (
    <nav className="pipeline-rail" aria-label="Étapes du laboratoire">
      {CHAPTERS.map(c => (
        <a key={c.id} className="pr-step" href={`#chapter-${c.id}`}>
          <span className="pr-n mono" aria-hidden="true">{c.n}</span>
          <span className="pr-label">{c.label}</span>
        </a>
      ))}
    </nav>
  )
}
