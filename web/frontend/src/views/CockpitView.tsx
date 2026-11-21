// Cockpit : une seule lecture, épinglée sur la décision. Aucune mesure n'est
// fabriquée ici : les créneaux attendent la source réelle et disent leur absence.
import { useState } from 'react'
import { useAuditStore } from '../store/audit'
import { useUIStore } from '../store/ui'
import AnswerBand from '../components/cockpit/AnswerBand'
import DecisionSpine from '../components/cockpit/DecisionSpine'
import JournalRail from '../components/cockpit/JournalRail'
import TimeField from '../components/cockpit/TimeField'
import { instrument } from '../lib/ringBuffer'
import { Drawer } from '../components/ui/Drawer'

export default function CockpitView() {
  const [proofOpen, setProofOpen] = useState(false)
  const report = useAuditStore(s => s.report)
  const stage = useAuditStore(s => s.stage)
  const failure = useAuditStore(s => s.failure)
  const start = useAuditStore(s => s.start)
  const stop = useAuditStore(s => s.stop)
  const setMode = useUIStore(s => s.setMode)
  const running = useUIStore(s => !!s.live?.running)

  return (
    <section className="spine-view" aria-label="Cockpit">
      <AnswerBand>
        <DecisionSpine
          report={report}
          stage={stage ?? (report ? 'done' : 'idle')}
          failure={failure}
          actions={{
            declare: () => setMode('labo'),
            arm: () => { void start() },
            cancel: () => { void stop() },
            remediate: () => setMode('labo'),
            proof: () => setProofOpen(true),
          }}
        />
      </AnswerBand>

      {/* T5 — l'instrument lit les anneaux réels écrits par le flux SSE. */}
      <div className="cockpit-field">
        <TimeField rings={instrument} objectiveMs={220} />
        <JournalRail events={instrument.events} running={running} />
      </div>

      <Drawer open={proofOpen} onClose={() => setProofOpen(false)} title="Preuve">
        <p className="muted">Aucune preuve scellée.</p>
      </Drawer>
    </section>
  )
}
