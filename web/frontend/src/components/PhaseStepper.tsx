type AuditViewPhase = 'repos' | 'charge' | 'vidange'

const PHASES: Array<{ id: AuditViewPhase; label: string }> = [
  { id: 'repos', label: 'Repos' },
  { id: 'charge', label: 'Charge' },
  { id: 'vidange', label: 'Vidange' },
]

export default function PhaseStepper({ phase, error }: { phase: AuditViewPhase; error?: string }) {
  const activeIndex = PHASES.findIndex(item => item.id === phase)
  return (
    <section aria-label="Phases de l’audit" className="phase-stepper">
      <ol role="group" aria-label="Phases de l’audit">
        {PHASES.map((item, index) => {
          const state = index < activeIndex ? 'done' : index === activeIndex ? 'current' : 'pending'
          return (
            <li
              key={item.id}
              className="phase-step"
              data-state={state}
              aria-current={state === 'current' ? 'step' : undefined}
            >
              <span className="phase-step-index" aria-hidden="true">{index + 1}</span>
              <span aria-label={item.label}>{item.label}</span>
            </li>
          )
        })}
      </ol>
      {error ? <p role="status">{error}</p> : null}
    </section>
  )
}
