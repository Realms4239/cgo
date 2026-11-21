// Laboratoire — un seul document en quatre chapters, sous un en-tête de
// commande épinglé. La mesure vient du moteur et des anneaux ; l'interface ne
// fabrique ni chiffre, ni sceau, ni verdict.
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'
import CommandHeader, { type CommandState } from '../components/labo/CommandHeader'
import PipelineRail from '../components/labo/PipelineRail'
import Chapter from '../components/labo/Chapter'
import Configurer from '../components/labo/Configurer'
import Observer from '../components/labo/Observer'
import Comparer from '../components/labo/Comparer'
import Preuves from '../components/labo/Preuves'

export default function LaboView() {
  const report = useAuditStore(s => s.report)
  const stage = useAuditStore(s => s.stage)
  const start = useAuditStore(s => s.start)
  const stop = useAuditStore(s => s.stop)
  const running = useUIStore(s => !!s.live?.running)

  const state: CommandState = running || stage === 'observing'
    ? 'observing'
    : stage === 'done' && report ? 'done' : 'idle'
  const recipe = report?.site
    ? `${report.site} · ${report.link_type ?? 'lien non précisé'}`
    : 'Lien non déclaré'

  return (
    <div className="labo-view">
      <CommandHeader
        recipe={recipe}
        state={state}
        running={running}
        objectiveMs={220}
        onArm={() => { void start() }}
        onStop={() => { void stop() }}
      />
      <div className="labo-body">
      <PipelineRail />
      <div className="labo-chapters">
        <Chapter
          id="configurer" n="01" title="Configurer"
          lede="Un réglage reste une hypothèse jusqu'à sa mesure."
        >
          <Configurer site={report?.site} link={report?.link_type} />
        </Chapter>
        <Chapter
          id="observer" n="02" title="Observer"
          lede="La trace attend la première sonde ; une absence n'est jamais un zéro."
        >
          <Observer running={running} />
        </Chapter>
        <Chapter
          id="comparer" n="03" title="Comparer"
          lede="Deux mesures complètes seulement peuvent être comparées."
        >
          <Comparer />
        </Chapter>
        <Chapter
          id="preuves" n="04" title="Preuves"
          lede="Le sceau n'est dit vérifié qu'après une vérification réelle."
        >
          <Preuves report={report} />
        </Chapter>
      </div>
      </div>
    </div>
  )
}
