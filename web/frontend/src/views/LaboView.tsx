// Laboratoire — quatre états d'une même méthode, un seul panneau visible à la
// fois sous un en-tête de commande. La mesure vient du moteur et des anneaux ;
// l'interface ne fabrique ni chiffre, ni sceau, ni verdict.
import { useEffect, useRef, useState } from 'react'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'
import CommandHeader, { type CommandState } from '../components/labo/CommandHeader'
import PipelineRail, { CHAPTERS } from '../components/labo/PipelineRail'
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
  const [active, setActive] = useState(0)
  const firstRender = useRef(true)

  useEffect(() => {
    if (firstRender.current) { firstRender.current = false; return }
    document.getElementById(`chapter-${CHAPTERS[active].id}-title`)?.focus({ preventScroll: true })
  }, [active])

  const state: CommandState = running || stage === 'observing'
    ? 'observing'
    : stage === 'done' && report ? 'done' : 'idle'
  const recipe = report?.site
    ? `${report.site} · ${report.link_type ?? 'lien non précisé'}`
    : 'Aucun lien déclaré'

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
        <PipelineRail active={active} onSelect={setActive} />
        <div className="panel-viewport">
          <div className="panel-track" style={{ transform: `translateX(-${active * 100}%)` }}>
            <Chapter
              id="configurer" n="01" title="Configurer" active={active === 0}
              lede="Quel lien voulez-vous étudier ?"
              outcome="Vous verrez la déclaration réelle du lien et les profils disponibles."
            >
              <Configurer site={report?.site} link={report?.link_type} />
            </Chapter>
            <Chapter
              id="observer" n="02" title="Observer" active={active === 1}
              lede="Que se passe-t-il sur le lien pendant la mesure ?"
              outcome="Vous verrez la matrice de progression, les portes et la chronologie d'une campagne."
            >
              <Observer running={running} />
            </Chapter>
            <Chapter
              id="comparer" n="03" title="Comparer" active={active === 2}
              lede="Quelle configuration est la plus rapide ?"
              outcome="Vous verrez les mesures classées, le duel et l'espace de décision."
            >
              <Comparer />
            </Chapter>
            <Chapter
              id="preuves" n="04" title="Preuves" active={active === 3}
              lede="Quelles preuves peut-on vérifier ?"
              outcome="Vous verrez l'état des fichiers de preuve, les archives scellées et les preuves vérifiables."
            >
              <Preuves report={report} />
            </Chapter>
          </div>
        </div>
      </div>
    </div>
  )
}
