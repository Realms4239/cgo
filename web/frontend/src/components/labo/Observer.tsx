// Observer — le vrai instrument : une seule règle de temps pour le délai, le
// débit utile et l'interface, plus le journal des faits. Les anneaux sont la
// source ; aucune valeur n'est recopiée dans React à chaque trame.
import { instrument } from '../../lib/ringBuffer'
import TimeField from '../cockpit/TimeField'
import JournalRail from '../cockpit/JournalRail'
import { CardHead } from '../ui/CardHead'

export default function Observer({ running = false }: { running?: boolean }) {
  return (
    <div className="observer">
      <CardHead label="Instrument" sub="Latence, débit et état du lien au fil du temps." />
      <div className="observer-field">
        <TimeField rings={instrument} objectiveMs={220} />
        <JournalRail events={instrument.events} running={running} />
      </div>
    </div>
  )
}
