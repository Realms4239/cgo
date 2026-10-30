// Banc d'essai — une page, trois sections progressives. Aucun sous-onglet.
import { startTransition, useEffect, useState } from 'react'
import BenchConfig from '../components/BenchConfig'
import BenchProgress from '../components/BenchProgress'
import DisplayControls from '../components/DisplayControls'
import { ModeHead } from '../components/ModeHead'
import LiveView from './LiveView'
import ResultatsView from './ResultatsView'
import { useUIStore } from '../store/ui'
import { useModeActive } from '../lib/hooks'
import { scrollToId as goTo } from '../lib/scroll'

export default function BenchView() {
  const stage = useUIStore(s => s.benchStage)
  const setStage = useUIStore(s => s.setBenchStage)
  const active = useModeActive('bench')
  const running = useUIStore(s => !!s.live?.running)
  const replay = useUIStore(s => s.replayRunning)
  const [resultsOpen, setResultsOpen] = useState(stage === 'results')

  useEffect(() => {
    if (!running && !replay) return
    setStage('observe')
    goTo('bench-observe')
  }, [running, replay, setStage])

  useEffect(() => {
    if (stage === 'results') setResultsOpen(true)
  }, [stage])

  // Le montage des résultats (matrice + TradeSpace) coûte ~250 ms : hors du
  // geste, en transition non urgente — le clic reste vif et le saut vers la
  // section se fait avant que le corps n'enfle.
  const openResults = () => {
    startTransition(() => { setStage('results'); setResultsOpen(true) })
    goTo('bench-results')
  }

  return (
    <div className="panel-stack bench-page" data-stage={stage}>
      <ModeHead
        title="Banc d'essai"
        sub="Configurer la campagne, observer la charge, lire le classement — une seule page linéaire."
      />

      <nav className="bench-anchors" aria-label="Sections du banc">
        <button className="btn" onClick={() => { setStage('config'); goTo('bench-config') }}>CONFIGURER</button>
        <button className="btn" onClick={() => goTo('bench-observe')}>OBSERVER</button>
        <button className="btn" onClick={openResults}>RÉSULTATS</button>
      </nav>

      <section id="bench-config" className="bench-section" aria-labelledby="bench-config-title">
        <h2 id="bench-config-title" className="bench-section-title">1 — Configurer</h2>
        <BenchConfig />
      </section>

      <section id="bench-observe" className="bench-section" aria-labelledby="bench-observe-title">
        <div className="bench-section-head">
          <h2 id="bench-observe-title" className="bench-section-title">2 — Observer</h2>
          <DisplayControls />
        </div>
        <BenchProgress active={active} />
        <LiveView active={active} />
      </section>

      <section id="bench-results" className="bench-section" aria-labelledby="bench-results-title">
        <h2 id="bench-results-title" className="bench-section-title">3 — Résultats</h2>
        {resultsOpen ? (
          <ResultatsView embedded active={active} />
        ) : (
          <div className="card bench-held">
            <p className="mono muted">Les résultats s'ouvrent à la fin de la campagne, ou en cliquant sur RÉSULTATS.</p>
            <button className="btn btn-primary" onClick={openResults}>OUVRIR LES RÉSULTATS</button>
          </div>
        )}
      </section>
    </div>
  )
}
