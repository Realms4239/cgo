import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { DuelCard } from './DuelCard'
import type { DuelPair } from '../lib/bench'

const usable: DuelPair = {
  profile: 'P2',
  deadline_ms: 220,
  min_reps: 3,
  base: { qdisc: 'pfifo_fast', cc: 'cubic', deadline_pct: 48.8, small_p95_ms: 328.2, n: 22 },
  best: { qdisc: 'cake', cc: 'bbr', deadline_pct: 98.1, small_p95_ms: 210.4, n: 20 },
  delta_deadline_pts: 49.3,
  delta_small_ms: -117.8,
}

const render = (pair: DuelPair | null) =>
  renderToStaticMarkup(<DuelCard pair={pair} onLaunch={() => undefined} />)

describe('DuelCard', () => {
  it('rend le face-à-face et les deltas du moteur', () => {
    const html = render(usable)
    expect(html).toContain('Le duel P2')
    expect(html).toContain('réglage par défaut')
    expect(html).toContain('réglage retenu')
    expect(html).not.toMatch(/pfifo_fast|fq_codel|cake|bbr|cubic/i)
    expect(html).toContain('48,8')
    expect(html).toContain('328,2')
    expect(html).toContain('22')
    expect(html).toContain('+49,3')
    expect(html).toContain('-117,8')
    expect(html).toContain('Lancer la comparaison')
    expect(html).toContain('P2 × 3 répétitions × 4 flux')
  })

  it('dit aucune base exploitable sans delta inventé', () => {
    const empty: DuelPair = {
      profile: 'P2',
      deadline_ms: 220,
      min_reps: 3,
      base: null,
      best: null,
      delta_deadline_pts: 0,
      delta_small_ms: 0,
      reason: 'aucune cellule valide avec au moins 3 répétitions — relancez une campagne',
    }
    for (const pair of [empty, null]) {
      const html = render(pair)
      expect(html).toContain('aucune base exploitable')
      expect(html).not.toContain('duel-delta')
      expect(html).not.toContain('+0,0')
    }
    expect(render(empty)).toContain('aucune cellule valide')
  })
})