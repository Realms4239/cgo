// DuelCard — la décision du banc en une carte : le réglage retenu face à la
// base, sur les mesures archivées. Aucun chiffre n'est calculé ici.
import { ArmButton } from './ArmButton'
import { numFr } from '../lib/format'
import type { DuelCell, DuelPair } from '../lib/bench'

const signed = (v: number, digits = 1) => `${v > 0 ? '+' : ''}${numFr(v, digits)}`

function Cell({ label, name, cell }: { label: string; name: string; cell: DuelCell }) {
  return (
    <div className="duel-cell">
      <span className="duel-cell-label">{label}</span>
      <b className="duel-cell-name">{name}</b>
      <dl className="duel-facts">
        <div>
          <dt>réponse sous charge respectée</dt>
          <dd>{numFr(cell.deadline_pct, 1)}<small> %</small></dd>
        </div>
        <div>
          <dt>réponse des petits objets</dt>
          <dd>{numFr(cell.small_p95_ms, 1)}<small> ms</small></dd>
        </div>
        <div>
          <dt>mesures</dt>
          <dd>{cell.n}</dd>
        </div>
      </dl>
    </div>
  )
}

export function DuelCard({
  pair,
  onLaunch,
  disabled = false,
}: {
  pair: DuelPair | null
  onLaunch: () => void
  disabled?: boolean
}) {
  const base = pair?.base ?? null
  const best = pair?.best ?? null
  const deadline = pair?.deadline_ms ?? 220

  return (
    <section className="card duel" aria-labelledby="duel-title">
      <div className="duel-head">
        <div>
          <h2 id="duel-title">Le duel P2</h2>
          <p className="duel-sub">Même lien, même charge : la base mesurée face au réglage retenu.</p>
        </div>
        <span className="mono duel-source">mesures archivées</span>
      </div>

      {pair && base && best ? (
        <>
          <div className="duel-cells">
            <Cell label="Base mesurée" name="réglage par défaut" cell={base} />
            <div className="duel-vs" aria-hidden="true">contre</div>
            <Cell label="Réglage retenu" name="réglage retenu" cell={best} />
          </div>
          <p className="duel-delta" data-testid="duel-delta">
            <span>écart du réglage retenu</span>
            <b>{signed(pair.delta_deadline_pts, 1)}<small> pts de délai</small></b>
            <b>{signed(pair.delta_small_ms, 1)}<small> ms de réponse</small></b>
          </p>
        </>
      ) : (
        <div className="duel-empty">
          <p>aucune base exploitable</p>
          {pair?.reason ? <p className="duel-reason">{pair.reason}</p> : null}
        </div>
      )}

      <div className="duel-launch">
        <ArmButton label="Lancer la comparaison" confirmLabel="CONFIRMER ?" onConfirm={onLaunch} disabled={disabled} />
        <span className="mono muted duel-preset">P2 × 3 répétitions × 4 flux × canal propre × objectif {deadline} ms</span>
      </div>
    </section>
  )
}
