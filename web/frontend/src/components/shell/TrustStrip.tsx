// Bande de confiance : une ligne calme. Chaque fait reste un MOT + une FORME ;
// une valeur absente s'écrit ABSENT, jamais 0 — zéro est une mesure, pas un substitut.
// Une seule pastille s'allume en faute, jamais tout le bandeau.
export interface TrustStripProps {
  connection: 'online' | 'offline'
  source?: string | null
  freshness?: string | null
  seal?: string | null
  backpressure?: number
}

type SlotState = 'ok' | 'absent' | 'alarm'
interface Slot {
  id: string
  shape: string
  text: string
  state: SlotState
}

// Miroir de VERSION à la racine ; à resynchroniser lors d'un bump.
const VERSION = '2.0.0'

const word = (value: string | null | undefined): string =>
  value == null || value === '' ? 'ABSENT' : value

const fact = (id: string, shape: string, value: string | null | undefined): Slot => {
  const text = word(value)
  return { id, shape, text, state: text === 'ABSENT' ? 'absent' : 'ok' }
}

export default function TrustStrip({
  connection, source, freshness, seal, backpressure,
}: TrustStripProps) {
  const online = connection === 'online'
  const slots: Slot[] = [
    {
      id: 'connection', shape: 'trust-shape-ring',
      text: online ? 'EN LIGNE' : 'HORS LIGNE', state: online ? 'ok' : 'alarm',
    },
    fact('source', 'trust-shape-square', source),
    fact('freshness', 'trust-shape-triangle', freshness),
    fact('seal', 'trust-shape-diamond', seal),
    backpressure == null
      ? { id: 'backpressure', shape: 'trust-shape-bars', text: 'ABSENT', state: 'absent' }
      : backpressure === 0
        ? { id: 'backpressure', shape: 'trust-shape-bars', text: 'AUCUNE PERTE', state: 'ok' }
        : { id: 'backpressure', shape: 'trust-shape-bars', text: `${backpressure} trames perdues`, state: 'alarm' },
  ]

  // Plusieurs faits peuvent être en faute : une seule pastille s'allume, le lien d'abord.
  const lit = slots.find(s => s.state === 'alarm')?.id ?? null
  const abnormal = lit !== null

  return (
    <footer className="trust-strip" data-abnormal={abnormal ? '1' : '0'} aria-label="Bande de confiance">
      <div className="trust-id">
        <span className="trust-brand">CGO</span>
        <span className="trust-version">v{VERSION}</span>
      </div>
      <div className="trust-slots">
        {slots.map(slot => (
          <span
            key={slot.id}
            className="trust-slot"
            data-slot={slot.id}
            data-state={slot.state}
            data-lit={slot.id === lit ? '1' : undefined}
          >
            <span className={'trust-shape ' + slot.shape} aria-hidden="true" />
            <span className="trust-word">{slot.text}</span>
          </span>
        ))}
      </div>
    </footer>
  )
}
