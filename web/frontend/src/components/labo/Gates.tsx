// Portes G0–G7 — texte tri-state. Le mot (PASS / FAIL / —) est lu sur la
// valeur réelle transmise par le serveur, jamais sur le rang : aucune pastille
// verte par index, aucune porte inventée quand la valeur manque. La science
// est une divulgation inline, même tâche, même endroit.
import { GATE_LABELS } from '../../lib/gates'
import { ScienceDetails } from '../ScienceDetails'
import { CardHead } from '../ui/CardHead'

export type GateState = 'pass' | 'fail' | 'na'

export function gateState(value: boolean | null | undefined): GateState {
  return value === true ? 'pass' : value === false ? 'fail' : 'na'
}

export const GATE_TOKEN: Record<GateState, string> = { pass: 'PASS', fail: 'FAIL', na: '—' }
const GATE_WORD: Record<GateState, string> = { pass: 'évaluée', fail: 'évaluée', na: 'non évaluée' }

export default function Gates({ gates }: { gates: (boolean | null)[] }) {
  const states = GATE_LABELS.map((_, i) => gateState(gates[i]))
  const assessed = states.filter(s => s !== 'na').length
  return (
    <section className="labo-gates" data-state={assessed ? 'assessed' : 'absent'}>
      <CardHead label="Portes qualité" sub={`G0 – G7 · ${assessed || '—'} évaluées · le mot porte le résultat, jamais le rang`} />
      <ol className="labo-gate-list">
        {states.map((st, i) => (
          <li key={i} className="labo-gate-row" data-state={st}>
            <span className="labo-gate-id mono">G{i}</span>
            <span className="labo-gate-token mono">{GATE_TOKEN[st]}</span>
            <span className="labo-gate-word">{GATE_WORD[st]}</span>
          </li>
        ))}
      </ol>
      <ScienceDetails label="Portes G0–G7 — science des contrôles">
        <ol className="labo-gate-science">
          {GATE_LABELS.map((label, i) => (
            <li key={i}><b className="mono">G{i}</b> — {label}</li>
          ))}
        </ol>
      </ScienceDetails>
    </section>
  )
}
