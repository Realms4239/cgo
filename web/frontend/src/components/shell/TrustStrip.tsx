// Bande de confiance : cinq faits, chacun un MOT et une FORME. Une valeur
// manquante s'écrit ABSENT, jamais 0 — zéro est une mesure, pas un substitut.
export interface TrustStripProps {
  connection: 'online' | 'offline'
  source?: string | null
  freshness?: string | null
  seal?: string | null
  backpressure?: number
}

const word = (value: string | null | undefined): string =>
  value == null || value === '' ? 'ABSENT' : value

export default function TrustStrip({
  connection, source, freshness, seal, backpressure,
}: TrustStripProps) {
  const online = connection === 'online'
  const sourceWord = word(source)
  const freshnessWord = word(freshness)
  const sealWord = word(seal)
  const lossWord = backpressure == null
    ? 'ABSENT'
    : backpressure === 0
      ? 'AUCUNE PERTE'
      : `${backpressure} trames perdues`
  const abnormal = !online
    || sourceWord === 'ABSENT'
    || freshnessWord === 'ABSENT'
    || sealWord === 'ABSENT'
    || lossWord === 'ABSENT'
    || (backpressure != null && backpressure > 0)

  const slots: { id: string; shape: string; text: string }[] = [
    { id: 'connection', shape: 'trust-shape-ring', text: online ? 'EN LIGNE' : 'HORS LIGNE' },
    { id: 'source', shape: 'trust-shape-square', text: sourceWord },
    { id: 'freshness', shape: 'trust-shape-triangle', text: freshnessWord },
    { id: 'seal', shape: 'trust-shape-diamond', text: sealWord },
    { id: 'backpressure', shape: 'trust-shape-bars', text: lossWord },
  ]

  return (
    <footer className="trust-strip" data-abnormal={abnormal ? '1' : '0'} aria-label="Bande de confiance">
      {abnormal ? (
        <span className="trust-alert">
          <span className="trust-flag" aria-hidden="true" />
          <span className="trust-word">ANOMALIE</span>
        </span>
      ) : null}
      <div className="trust-slots">
        {slots.map(slot => (
          <span key={slot.id} className="trust-slot" data-slot={slot.id}>
            <span className={'trust-shape ' + slot.shape} aria-hidden="true" />
            <span className="trust-word">{slot.text}</span>
          </span>
        ))}
      </div>
    </footer>
  )
}
