// StateSlot — une valeur ou une absence, jamais un zéro fabriqué.
// La hauteur du créneau est réservée en permanence : le passage de l'absence
// à la mesure ne fait pas sauter la rangée. Mot + forme portent l'état.
export interface StateSlotProps {
  label: string
  value?: string | number | null
  unit?: string
  word: string
  empty?: boolean
}

export function StateSlot({ label, value, unit, word, empty = false }: StateSlotProps) {
  const absent = empty || value === null || value === undefined
  return (
    <div className="state-slot" data-state={absent ? 'absent' : 'measured'}>
      <span className="state-slot-label">{label}</span>
      <span className="state-slot-value" style={{ minHeight: 'var(--fs-metric)' }}>
        {absent ? (
          <>
            <span className="state-slot-dash" aria-hidden="true">—</span>
            <span className="state-slot-word">{word}</span>
          </>
        ) : (
          <>
            <span className="state-slot-num">{String(value)}</span>
            {unit ? <span className="state-slot-unit">{unit}</span> : null}
          </>
        )}
      </span>
    </div>
  )
}
