// En-tête de commande épinglé : la recette, l'état et les deux gestes restent
// visibles sans défilement, pendant toute la campagne. Aucun chiffre inventé.
import { CardHead } from '../ui/CardHead'

export type CommandState = 'idle' | 'observing' | 'done'

export interface CommandHeaderProps {
  recipe: string
  state: CommandState
  running: boolean
  objectiveMs?: number
  onArm: () => void
  onStop: () => void
}

const STATE_WORD: Record<CommandState, string> = {
  idle: 'En attente',
  observing: 'Mesure en cours',
  done: 'Mesure terminée',
}

export default function CommandHeader({
  recipe, state, running, objectiveMs = 220, onArm, onStop,
}: CommandHeaderProps) {
  return (
    <header className="command-header" data-pinned="true" data-state={state}>
      <div className="ch-line">
        <CardHead
          label="Laboratoire"
          sub={recipe}
          right={<span className="ch-state" data-state={state}>{STATE_WORD[state]}</span>}
        />
      </div>
      <div className="ch-actions">
        <button type="button" className="btn btn-primary cmd-arm" onClick={onArm} disabled={running}>
          ARMER
        </button>
        <button type="button" className="btn btn-danger cmd-stop" onClick={onStop} disabled={!running}>
          ARRÊTER
        </button>
        <span className="ch-objective mono">objectif {objectiveMs} ms</span>
      </div>
    </header>
  )
}
