// Comparer — constat, duel, classement, espace de compromis. Une mesure absente
// se classe en dernier et affiche « — » ; un duel à une mesure vide refuse avec
// des mots, jamais une équivalence inventée.
import { useEffect, useState } from 'react'
import { CardHead } from '../ui/CardHead'
import { EmptyChart } from '../ui/EmptyChart'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { StateSlot } from '../ui/StateSlot'
import { numFr, improve } from '../../lib/format'
import { paretoFrontier } from '../../lib/chartGrammar'
import { profileShort } from '../../lib/profiles'

export interface RankRow {
  id: string
  label: string
  delayMs: number | null
  goodputMbps: number | null
}

const measured = (v: number | null | undefined): v is number =>
  typeof v === 'number' && Number.isFinite(v) && v > 0

/** Mesurés d'abord (du délai le plus court au plus long), absents en dernier. */
export function rankRows(rows: RankRow[]): RankRow[] {
  return [...rows].sort((a, b) => {
    const am = measured(a.delayMs)
    const bm = measured(b.delayMs)
    if (am !== bm) return am ? -1 : 1
    if (!am && !bm) return a.label.localeCompare(b.label, 'fr')
    return (a.delayMs as number) - (b.delayMs as number)
  })
}

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

function Duel({ a, b }: { a?: RankRow | null; b?: RankRow | null }) {
  const both = !!a && !!b && measured(a.delayMs) && measured(b.delayMs)
  if (!both) {
    return (
      <div className="duel duel-refused" data-state="refused" role="status">
        <p>Sélectionnez deux mesures pour les comparer — leur écart de délai s'affichera ici. Une mesure absente ne vaut pas une mesure égale.</p>
      </div>
    )
  }
  const delta = improve(a!.delayMs as number, b!.delayMs as number, 'down')
  const word = delta === null ? 'écart non calculable'
    : delta === 0 ? 'résultats proches'
      : delta > 0 ? 'le candidat est plus rapide' : 'la référence est plus rapide'
  return (
    <div className="duel" data-state="compared">
      <div className="duel-side">
        <span className="duel-role mono">référence</span>
        <span className="duel-label">{a!.label}</span>
        <span className="duel-value mono">{numFr(a!.delayMs, 1)} ms</span>
      </div>
      <div className="duel-side">
        <span className="duel-role mono">candidat</span>
        <span className="duel-label">{b!.label}</span>
        <span className="duel-value mono">{numFr(b!.delayMs, 1)} ms</span>
      </div>
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

  const ranked = rankRows(rows ?? fetched)
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
        <CardHead
          label="Duel"
          sub="Cliquez deux lignes du classement : la première est la référence"
          right={duel.a ? (
            <button type="button" className="btn" onClick={() => setDuel({ a: null, b: null })}>
              Réinitialiser
            </button>
          ) : undefined}
        />
        <Duel a={duel.a} b={duel.b} />
      </div>

      <div className="comparer-rank">
        <CardHead label="Classement" sub="Délai sous charge, du plus court au plus long" />
        <div className="rank-scroll">
          <table className="rank-table">
            <thead>
              <tr>
                <th scope="col">Rang</th>
                <th scope="col">Configuration</th>
                <th scope="col" aria-sort="ascending">Délai sous charge</th>
                <th scope="col">Débit utile</th>
              </tr>
            </thead>
            <tbody>
              {ranked.length === 0 ? (
                <tr className="rank-row rank-empty">
                  <td colSpan={4}>Aucune mesure — lancez la campagne : chaque configuration apparaîtra ici, du délai le plus court au plus long.</td>
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
          <ul className="frontier-list">
            {frontier.map(i => (
              <li key={i} className="mono">
                {numFr(pts[i].x, 1)} ms · {numFr(pts[i].y, 1)} Mbit/s
              </li>
            ))}
          </ul>
        ) : (
          <p className="muted">Aucun compromis calculable : il faut au moins deux mesures complètes.</p>
        )}
      </InlineDisclosure>
    </div>
  )
}
