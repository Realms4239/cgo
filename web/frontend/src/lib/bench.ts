// Progression du banc d'essai — trois sections, jamais trois onglets.
import type { ChannelModel } from './canal'

export type BenchStage = 'config' | 'observe' | 'results'

// Une campagne qui démarre ouvre l'observation ; une campagne qui s'arrête
// ouvre les résultats. Un replay ouvre l'observation sans toucher au reste.
export function stageAfterCampaign(previous: BenchStage, running: boolean): BenchStage {
  if (running) return 'observe'
  return previous === 'observe' ? 'results' : previous
}

export function stageAfterReplay(previous: BenchStage, replay: boolean): BenchStage {
  return replay ? 'observe' : previous
}

// Configurer est toujours visible ; Observer apparaît dès qu'une exécution a
// eu lieu ; Résultats seulement quand la section est atteinte.
// Résultats du duel — miroir de `pkg/api/duel.go`. Le navigateur affiche les
// deltas du moteur ; il ne soustrait rien.
export interface DuelCell {
  qdisc: string
  cc: string
  deadline_pct: number
  small_p95_ms: number
  n: number
  runs?: string[]
}

export interface DuelPair {
  profile: string
  deadline_ms: number
  min_reps: number
  base: DuelCell | null
  best: DuelCell | null
  delta_deadline_pts: number
  delta_small_ms: number
  reason?: string
}

export class CampaignStartError extends Error {
  readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'CampaignStartError'
    this.status = status
  }
}

export async function loadDuel(profile = 'P2'): Promise<DuelPair> {
  const r = await fetch(`/api/duel?profile=${encodeURIComponent(profile)}`)
  if (!r.ok) throw new CampaignStartError(`duel indisponible (${r.status})`, r.status)
  return (await r.json()) as DuelPair
}

export interface CampaignOptions {
  profiles: string[]
  reps: number
  flows: number
  channel_model: ChannelModel
  deadline_ms: number
  target: string
}

export async function startCampaign(opts: CampaignOptions): Promise<void> {
  const r = await fetch('/api/run/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(opts),
  })
  if (!r.ok) {
    const txt = await r.text().catch(() => String(r.status))
    throw new CampaignStartError(txt.slice(0, 500), r.status)
  }
}

export function sectionOpen(stage: BenchStage, which: BenchStage): boolean {
  if (which === 'config') return true
  if (which === 'observe') return stage !== 'config'
  return stage === 'results'
}
