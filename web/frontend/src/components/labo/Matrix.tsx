// Matrice CH2 — le plan réel (profils × files × contrôles) et l'état de chaque
// cellule dans l'ordre du serveur. Sans campagne : « — » + attente, jamais un
// compte zéro fabriqué. Après campagne : mots FINALE / terminée, jamais retour
// à attente. Aucune cellule n'est inventée : sans plan, il n'y a rien à montrer.
import { CardHead } from '../ui/CardHead'
import { EmptyState } from '../ui/EmptyState'
import { StatusPip } from '../ui/StatusPip'

export interface MatrixPlan {
  profiles: string[]
  qdiscs: string[]
  ccs: string[]
  reps: number
}

export interface CampaignProgress {
  running: boolean
  eventId: number
  doneEvents: number
  totalEvents: number
}

export type CellState = 'attente' | 'en cours' | 'terminée'

export interface MatrixCell {
  profile: string
  qdisc: string
  cc: string
  n: number | null // null = aucune donnée réelle → « — »
  reps: number
  state: CellState
}

const GLYPH: Record<CellState, string> = { attente: '○', 'en cours': '◐', terminée: '●' }

/** Ordre serveur (campagne.go) : profils → files → contrôles → répétitions. */
export function matrixCells(plan: MatrixPlan, p: CampaignProgress | null): MatrixCell[] {
  const { profiles, qdiscs, ccs, reps } = plan
  if (!profiles.length || !qdiscs.length || !ccs.length || !reps) return []
  const started = !!p && (p.eventId > 0 || p.totalEvents > 0 || p.doneEvents > 0)
  // event_id est l'identifiant 1-based de l'événement courant (donnée réelle) ;
  // done_events sert de repli quand l'identifiant n'est pas encore monté.
  const cur = started ? (p!.eventId > 0 ? p!.eventId : p!.doneEvents) : 0
  return profiles.flatMap((profile, pi) =>
    qdiscs.flatMap((qdisc, qi) =>
      ccs.map((cc, ci) => {
        const first = ((pi * qdiscs.length + qi) * ccs.length + ci) * reps + 1
        if (!cur) return { profile, qdisc, cc, n: null, reps, state: 'attente' as CellState }
        const state: CellState = first + reps - 1 <= cur
          ? 'terminée' : first <= cur ? 'en cours' : 'attente'
        // Une cellule non commencée n'a aucun compte à afficher : « — », pas 0.
        const n = state === 'attente' ? null : Math.min(reps, cur - first + 1)
        return { profile, qdisc, cc, n, reps, state }
      })))
}

export default function Matrix({ plan, progress }: { plan: MatrixPlan; progress: CampaignProgress | null }) {
  const cells = matrixCells(plan, progress)
  const final = !!progress && !progress.running && progress.totalEvents > 0 && progress.eventId >= progress.totalEvents
  const phase = !cells.length ? 'absent' : final ? 'final' : progress ? 'live' : 'attente'
  const sentence = !cells.length
    ? 'Plan non disponible : les files et les contrôles déclarés par l’hôte sont nécessaires.'
    : final ? 'Campagne terminée : les mots ci-dessous sont définitifs.'
      : progress ? 'Campagne en cours : les mots suivent l’avancement réel du serveur.'
        : 'En attente d’une campagne : chaque combinaison sera testée dans cet ordre, avec son état et ses répétitions.'
  const word = phase === 'absent' ? '— indisponible' : phase === 'final' ? 'FINALE'
    : phase === 'live' ? 'EN COURS' : 'ATTENTE'

  return (
    <section className="labo-matrix" data-state={phase}>
      <CardHead label="Matrice de progression" sub={`${cells.length || '—'} combinaisons · ordre du serveur`}
        right={<StatusPip state={phase === 'final' ? 'ok' : phase === 'live' ? 'on' : 'idle'} label={word} title={word} />} />
      <p className="labo-matrix-sentence">{sentence}</p>
      {cells.length ? <p className="labo-matrix-note">n déduit de l’avancement serveur</p> : null}
      {cells.length ? (
        <table className="labo-matrix-table">
          <thead>
            <tr>
              <th scope="col">profil</th>
              <th scope="col">file et contrôle</th>
              <th scope="col">répétitions</th>
              <th scope="col">état</th>
            </tr>
          </thead>
          <tbody>
            {cells.map((c, i) => (
              <tr key={`${c.profile}/${c.qdisc}/${c.cc}/${i}`} data-state={c.state}>
                <td data-label="profil">{c.profile}</td>
                <td data-label="file et contrôle">{c.qdisc} · {c.cc}</td>
                <td data-label="répétitions" className="num">
                  {c.n === null ? <span className="labo-absent">— non lancée</span> : `${c.n}/${c.reps}`}
                </td>
                <td data-label="état" className="state">
                  <span className="labo-cell-glyph" aria-hidden="true">{GLYPH[c.state]}</span> {c.state}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <EmptyState kind="empty" hint="Plan non disponible." />
      )}
    </section>
  )
}
