// Observer — l'état de la campagne : phase, cellules, portes. La progression
// ne se lit que sur un run actif ; au repos, la matrice reste un aperçu replié.
import { Timeline } from './Timeline'
import Explain from './Explain'
import { EmptyState } from './ui/EmptyState'
import { ScienceDetails } from './ScienceDetails'
import { useUIStore } from '../store/ui'
import { useBenchStore } from '../store/bench'
import { live as liveRing } from '../lib/live'
import { GATE_LABELS } from '../lib/gates'

const ALL_QDISCS = ['pfifo_fast', 'fq_codel', 'cake'] as const
const ALL_CC = ['cubic', 'bbr'] as const

export default function BenchProgress({ active = true }: { active?: boolean }) {
  const live = useUIStore(s => (active ? s.live : null))
  const profiles = useBenchStore(s => s.profiles)
  const reps = useBenchStore(s => s.reps)

  const loadPlain = live?.load_status === 'bulk-on' ? 'transferts en cours'
    : live?.load_status === 'bulk-off' ? 'sans transfert'
    : live?.load_status ?? '—'
  const gates = live?.gates ?? Array(8).fill(null)
  const phase = live?.phase || 'idle'
  const hasData = !!(live?.running || liveRing.small.length > 0)
  const ps = liveRing.phaseSince
  const now = Date.now()
  const timeline = {
    baselineStart: ps.baseline ?? now - 30000,
    chargeStart: ps.charge ?? ps.baseline ?? now - 15000,
    chargeEnd: ps.recup ?? (ps.charge ? now : ps.baseline ?? now),
    recupEnd: (ps.recup ?? ps.charge ?? ps.baseline ?? now) + 15000,
  }
  const hasTimeline = ps.baseline != null || ps.charge != null || ps.recup != null

  const matrix = profiles.length > 0 ? (
    <div className="card" data-testid="matrix-table">
      <div className="card-head">Matrice — état par cellule</div>
      <div className="bench-matrix-scroll">
        <table className="data-table bench-matrix">
          <thead>
            <tr><th>profil</th><th>file × contrôle</th><th>répétitions</th><th>état</th></tr>
          </thead>
          <tbody>
            {profiles.flatMap((p, pi) => ALL_QDISCS.flatMap((q, qi) => ALL_CC.map((c, ci) => {
              // ordre serveur (matrix.go) : profils × [pfifo, fq_codel, cake] × [cubic, bbr] × reps
              const cellStart = ((pi * ALL_QDISCS.length + qi) * ALL_CC.length + ci) * reps + 1
              const cur = live?.running ? live?.event_id ?? 0 : 0
              const st = cur === 0 ? 'attente' : (cellStart + reps - 1 < cur ? 'terminée' : (cellStart <= cur ? 'en cours' : 'attente'))
              const n = Math.max(0, Math.min(reps, cur - cellStart + (st === 'en cours' ? 1 : 0)))
              const color = st === 'terminée' ? 'var(--t-ok)' : st === 'en cours' ? 'var(--t-live)' : 'var(--text-faint)'
              return (
                <tr key={`${p}/${q}/${c}`}>
                  <td>{p}</td>
                  <td>{q} × {c}</td>
                  <td className="bench-num-col">{n}/{reps}</td>
                  <td style={{ color }}>{st}</td>
                </tr>
              )
            })))}
          </tbody>
        </table>
      </div>
      <div className="mono muted bench-matrix-note">
        La campagne teste chaque combinaison dans cet ordre : profils, puis files, puis contrôles, puis répétitions
        {live?.running && live?.event_id != null ? ` · événement #${live.event_id}/${live?.total_events ?? '—'}` : ''}
      </div>
    </div>
  ) : null

  const gatesCard = (
    <div className="gates-detail">
      {gates.map((g: boolean | null, i: number) => (
        <div key={i} className={'gate-row ' + (g === null ? 'na' : g ? 'ok' : 'fail')}>
          <span className="gate mono">G{i}</span>
          <span className="gate-lbl"><Explain term={`G${i}`}>{GATE_LABELS[i]}</Explain></span>
          <span className="gate-state mono">{g === null ? '—' : g ? 'PASS' : 'FAIL'}</span>
        </div>
      ))}
    </div>
  )

  // Au repos, on ne peint ni phase vide ni matrice « attente » : un aperçu
  // replié dit ce qui va tourner, sans simuler une progression.
  if (!hasData) {
    return (
      <div className="bench-progress">
        <EmptyState kind="idle" hint="en attente — démarrez une campagne dans Configurer" />
        <details className="bench-preview">
          <summary className="mono">APERÇU — MATRICE ET PORTES QUALITÉ</summary>
          <div className="bench-preview-body">
            {matrix}
            <ScienceDetails label="Portes G0–G7 — contrôle qualité de la mesure">{gatesCard}</ScienceDetails>
          </div>
        </details>
      </div>
    )
  }

  return (
    <div className="bench-progress">
      {hasTimeline ? (
        <Timeline baselineStart={timeline.baselineStart} chargeStart={timeline.chargeStart} chargeEnd={timeline.chargeEnd} recupEnd={timeline.recupEnd} currentPhase={phase} />
      ) : null}

      <div className="card">
        <div className="card-head">État — flux SSE</div>
        <div className="kv-grid-3">
          <div className="kv"><span>phase</span><b className="mono">{phase}</b></div>
          <div className="kv"><span>profil</span><b className="mono">{live?.profile ?? '—'}</b></div>
          <div className="kv"><span>file</span><b className="mono">{live?.qdisc ?? '—'}</b></div>
          <div className="kv"><span>contrôle</span><b className="mono">{live?.cc ?? '—'}</b></div>
          <div className="kv"><span>événement</span><b className="mono">{live?.event_id != null ? `#${live.event_id} · rép ${live.repetition ?? 0}` : '—'}</b></div>
          <div className="kv"><span>charge</span><b className="mono">{loadPlain}</b></div>
        </div>
      </div>

      {matrix}

      <ScienceDetails label="Portes G0–G7 — contrôle qualité de la mesure">{gatesCard}</ScienceDetails>
    </div>
  )
}
