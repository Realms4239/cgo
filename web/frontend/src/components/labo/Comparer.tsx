// Comparer — constat, duel, classement, espace de compromis. Une mesure absente
// se classe en dernier et affiche « — » ; un duel à une mesure vide refuse avec
// des mots, jamais une équivalence inventée.
import { useEffect, useState } from 'react'
import { CardHead } from '../ui/CardHead'
import { EmptyState } from '../ui/EmptyState'
import { EmptyChart } from '../ui/EmptyChart'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { StateSlot } from '../ui/StateSlot'
import { numFr, improve } from '../../lib/format'
import { paretoFrontier } from '../../lib/chartGrammar'
import { profileShort } from '../../lib/profiles'
import { useUIStore } from '../../store/ui'

export interface RankRow {
  id: string
  label: string
  delayMs: number | null
  goodputMbps: number | null
}

export type SortKey = 'label' | 'delay' | 'goodput'
export type SortDir = 'asc' | 'desc'

const measured = (v: number | null | undefined): v is number =>
  typeof v === 'number' && Number.isFinite(v) && v > 0

/**
 * Tri honnête : les non mesurés restent en dernier pour toute colonne et tout
 * sens — les trier « à l'envers » les ferait passer pour les meilleurs.
 */
export function sortRows(rows: RankRow[], key: SortKey, dir: SortDir): RankRow[] {
  const sign = dir === 'asc' ? 1 : -1
  const unmeasured = (r: RankRow) => !measured(r.delayMs)
  return [...rows].sort((a, b) => {
    const au = unmeasured(a)
    const bu = unmeasured(b)
    if (au !== bu) return au ? 1 : -1
    if (au && bu) return a.label.localeCompare(b.label, 'fr')
    let cmp = 0
    if (key === 'delay') {
      cmp = (a.delayMs as number) - (b.delayMs as number)
    } else if (key === 'goodput') {
      const ag = measured(a.goodputMbps)
      const bg = measured(b.goodputMbps)
      if (ag !== bg) return ag ? -1 : 1
      cmp = (a.goodputMbps as number) - (b.goodputMbps as number)
    } else {
      cmp = a.label.localeCompare(b.label, 'fr')
    }
    return cmp * sign
  })
}

/** Mesurés d'abord (du délai le plus court au plus long), absents en dernier. */
export function rankRows(rows: RankRow[]): RankRow[] {
  return sortRows(rows, 'delay', 'asc')
}

const signed = (v: number, digits = 1): string =>
  `${v > 0 ? '+' : ''}${numFr(v, digits)}`

interface ApiGroup {
  profile?: string
  small_p95_valid_median?: number
  small_p95_median?: number
  goodput_median?: number
  runs?: unknown[] | number
}
interface ApiResults { available?: boolean; groups?: ApiGroup[] }

function fromApi(j: ApiResults): RankRow[] {
  if (!j || j.available === false || !Array.isArray(j.groups)) return []
  return j.groups.map((g, i) => {
    const delay = g.small_p95_valid_median ?? g.small_p95_median
    return {
      id: `${g.profile ?? 'profil'}-${i}`,
      label: profileShort(g.profile ?? ''),
      delayMs: measured(delay) && (delay as number) > 0 ? delay as number : null,
      goodputMbps: measured(g.goodput_median) ? g.goodput_median as number : null,
    }
  })
}

function Duel({
  a, b, onProof,
}: {
  a?: RankRow | null
  b?: RankRow | null
  onProof: () => void
}) {
  const pins = (
    <div className="duel-pins" aria-label="Épingles du duel">
      <span className="duel-pin" data-slot="A">
        <span className="duel-pin-key mono">A</span>
        <span className="duel-pin-value">{a ? a.label : 'référence — à choisir'}</span>
      </span>
      <span className="duel-pin" data-slot="B">
        <span className="duel-pin-key mono">B</span>
        <span className="duel-pin-value">{b ? b.label : 'candidat — à choisir'}</span>
      </span>
      <button type="button" className="btn duel-proof" onClick={onProof}>PREUVE</button>
    </div>
  )

  const both = !!a && !!b && measured(a.delayMs) && measured(b.delayMs)
  if (!both) {
    return (
      <div className="duel duel-refused" data-state="refused" role="status">
        {pins}
        <p>Sélectionnez deux mesures pour les comparer — leur écart de délai s'affichera ici. Une mesure absente ne vaut pas une mesure égale.</p>
      </div>
    )
  }

  const da = a!.delayMs as number
  const db = b!.delayMs as number
  const scale = Math.max(da, db)
  // Sens unique : référence − candidat, positif quand le candidat est plus
  // rapide ; le pourcentage improve() suit le même sens (down = mieux).
  const deltaMs = da - db
  const deltaPct = improve(da, db, 'down')
  const winner = deltaMs > 0 ? b! : deltaMs < 0 ? a! : null
  const loser = winner === b! ? a! : b!
  const word = winner
    ? `${winner.label} est plus rapide que ${loser.label}`
    : 'résultats proches'

  return (
    <div className="duel" data-state="compared">
      {pins}
      <div className="duel-scale" aria-hidden="true">
        <span className="duel-bar" data-slot="A" style={{ width: `${(da / scale) * 100}%` }} />
        <span className="duel-bar" data-slot="B" style={{ width: `${(db / scale) * 100}%` }} />
      </div>
      <p className="duel-scale-note">échelle commune du délai — repère {numFr(scale, 1)} ms</p>
      <div className="duel-side">
        <span className="duel-role mono">référence · A</span>
        <span className="duel-label">{a!.label}</span>
        <span className="duel-value mono">{numFr(da, 1)} ms</span>
      </div>
      <div className="duel-side">
        <span className="duel-role mono">candidat · B</span>
        <span className="duel-label">{b!.label}</span>
        <span className="duel-value mono">{numFr(db, 1)} ms</span>
      </div>
      <p className="duel-delta" data-testid="duel-delta">
        <span>écart référence − candidat</span>
        <b className="mono">{signed(deltaMs)}<small> ms</small></b>
        <b className="mono">{deltaPct === null ? '—' : `${deltaPct > 0 ? '+' : ''}${deltaPct}`}<small> %</small></b>
      </p>
      <p className="duel-word" role="status">{word}</p>
    </div>
  )
}

export interface ComparerProps {
  rows?: RankRow[]
  a?: RankRow | null
  b?: RankRow | null
}

export default function Comparer({ rows, a = null, b = null }: ComparerProps) {
  const [fetched, setFetched] = useState<RankRow[]>([])
  const [duel, setDuel] = useState<{ a: RankRow | null; b: RankRow | null }>({ a, b })
  const [sort, setSort] = useState<{ key: SortKey; dir: SortDir }>({ key: 'delay', dir: 'asc' })
  const openProof = useUIStore(s => s.openProof)
  useEffect(() => {
    let alive = true
    fetch('/api/results').then(r => r.json()).then(j => { if (alive) setFetched(fromApi(j)) }).catch(() => {})
    return () => { alive = false }
  }, [])

  // Deux gestes suffisent : la première ligne pose la référence, la seconde le
  // candidat ; une troisième recommence. Une ligne sans mesure est refusée par
  // le duel lui-même, pas masquée ici.
  const pin = (r: RankRow) => setDuel(p =>
    (p.a && p.b) || !p.a ? { a: r, b: null } : { a: p.a, b: r })

  const toggle = (key: SortKey) => setSort(s => ({
    key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc',
  }))
  const sortAttr = (key: SortKey) =>
    sort.key === key ? (sort.dir === 'asc' ? 'ascending' : 'descending') : 'none'
  const header = (key: SortKey, label: string) => (
    <th scope="col" aria-sort={sortAttr(key)}>
      <button type="button" className="rank-sort" onClick={() => toggle(key)}>
        {label}
        <span className="rank-sort-mark mono" aria-hidden="true">
          {sort.key === key ? (sort.dir === 'asc' ? '▴' : '▾') : '↕'}
        </span>
      </button>
    </th>
  )

  const ranked = sortRows(rows ?? fetched, sort.key, sort.dir)
  const pts = ranked
    .filter(r => measured(r.delayMs) && measured(r.goodputMbps))
    .map(r => ({ x: r.delayMs as number, y: r.goodputMbps as number }))
  const frontier = pts.length ? paretoFrontier(pts, 'down', 'up') : []

  return (
    <div className="comparer">
      <p className="comparer-lede">
        Le constat vient de la mesure. Ce qui n'a pas été mesuré n'a pas de valeur et se classe en dernier.
      </p>

      <div className="comparer-duel">
        <CardHead label="Duel" sub="Cliquez deux lignes du classement : la première est la référence" />
        <Duel a={duel.a} b={duel.b} onProof={() => openProof('proof')} />
      </div>

      <div className="comparer-rank">
        <CardHead label="Classement" sub="Les colonnes se trient ; les non mesurés restent en dernier" />
        <div className="rank-scroll">
          <table className="rank-table">
            <thead>
              <tr>
                <th scope="col">Rang</th>
                {header('label', 'Configuration')}
                {header('delay', 'Délai sous charge')}
                {header('goodput', 'Débit utile')}
              </tr>
            </thead>
            <tbody>
              {ranked.length === 0 ? (
                <tr className="rank-row rank-empty">
                  <td colSpan={4}>
                    <EmptyState kind="empty" hint="Aucune mesure — lancez la campagne : chaque configuration apparaîtra ici, du délai le plus court au plus long." />
                  </td>
                </tr>
              ) : ranked.map((r, i) => (
                <tr key={r.id} className="rank-row">
                  <td className="rank-n mono">
                    <button
                      type="button" className="rank-pin" title="Sélectionner pour le duel"
                      aria-pressed={duel.a?.id === r.id || duel.b?.id === r.id}
                      onClick={() => pin(r)}
                    >
                      {measured(r.delayMs) ? i + 1 : '—'}
                    </button>
                  </td>
                  <th scope="row">{r.label}</th>
                  <td>
                    <StateSlot
                      label="Délai"
                      value={measured(r.delayMs) ? numFr(r.delayMs, 1) : null}
                      unit={measured(r.delayMs) ? 'ms' : undefined}
                      word="non mesuré"
                      empty={!measured(r.delayMs)}
                    />
                  </td>
                  <td>
                    <StateSlot
                      label="Débit"
                      value={measured(r.goodputMbps) ? numFr(r.goodputMbps, 1) : null}
                      unit={measured(r.goodputMbps) ? 'Mbit/s' : undefined}
                      word="non mesuré"
                      empty={!measured(r.goodputMbps)}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {ranked.length === 0 ? (
            <div className="rank-empty-frame">
              <EmptyChart hint="aucune mesure à classer" />
            </div>
          ) : null}
        </div>
      </div>

      <InlineDisclosure label="Espace de compromis">
        {frontier.length ? (
          <>
            <p className="frontier-summary">
              {frontier.length} configuration{frontier.length > 1 ? 's' : ''} sur l'enveloppe —
              {' '}délai le plus court {numFr(Math.min(...frontier.map(i => pts[i].x)), 1)} ms,
              {' '}débit le plus élevé {numFr(Math.max(...frontier.map(i => pts[i].y)), 1)} Mbit/s.
            </p>
            <ul className="frontier-list">
              {frontier.map(i => (
                <li key={i} className="mono">
                  {numFr(pts[i].x, 1)} ms · {numFr(pts[i].y, 1)} Mbit/s
                </li>
              ))}
            </ul>
          </>
        ) : (
          <p className="muted">Aucun compromis calculable : il faut au moins deux mesures complètes.</p>
        )}
      </InlineDisclosure>
    </div>
  )
}
