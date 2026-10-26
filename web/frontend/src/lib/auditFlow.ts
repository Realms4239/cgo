// auditFlow — la carte audit en trois temps : déclarer, observer, lire.
// Tout le calcul est ici (pur, testé) ; les composants ne font que rendre.
import type { AuditDirection, AuditLive } from './types'

export type AuditStage = 'idle' | 'declared' | 'observing' | 'done'
export type AuditStageEvent = 'open' | 'close' | 'start' | 'finish' | 'cancel' | 'again'

// La machine ne saute pas d'étape : un audit annulé n'a pas de rapport à lire,
// il retourne donc à la déclaration ; « again » repart d'un rapport terminé.
export function nextAuditStage(stage: AuditStage, event: AuditStageEvent): AuditStage {
  switch (stage) {
    case 'idle':
      return event === 'open' ? 'declared' : stage
    case 'declared':
      return event === 'start' ? 'observing' : event === 'close' ? 'idle' : stage
    case 'observing':
      return event === 'finish' ? 'done' : event === 'cancel' ? 'declared' : stage
    case 'done':
      return event === 'again' ? 'declared' : event === 'close' ? 'idle' : stage
  }
}

// 0 = replié, 1 = déclarer, 2 = observer, 3 = lire.
export function auditStep(stage: AuditStage): 0 | 1 | 2 | 3 {
  switch (stage) {
    case 'declared':
      return 1
    case 'observing':
      return 2
    case 'done':
      return 3
    default:
      return 0
  }
}

export const AUDIT_LINK_TYPES = ['fiber', '5g', '4g', 'vsat', 'other'] as const
export const AUDIT_DURATIONS = [20, 60] as const
export const DEFAULT_AUDIT_DIRECTION: AuditDirection = 'up'

export type DurationChoice = '20' | '60' | 'custom'

// 20/60 sont les deux durées de la segmentation ; « custom » ouvre 10–600.
export function durationChoice(seconds: number): DurationChoice {
  if (seconds === 20) return '20'
  if (seconds === 60) return '60'
  return 'custom'
}

// Normalisation : jamais de clamp silencieux — hors 10–600 ou illisible ⇒ null.
export function normalizeDuration(value: string | number): number | null {
  const n = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(n)) return null
  const seconds = Math.trunc(n)
  return seconds >= 10 && seconds <= 600 ? seconds : null
}

export function durationFor(choice: DurationChoice, custom: string | number): number | null {
  if (choice === '20') return 20
  if (choice === '60') return 60
  return normalizeDuration(custom)
}

export function normalizeDirection(value: unknown): AuditDirection {
  return value === 'down' || value === 'both' || value === 'up' ? value : DEFAULT_AUDIT_DIRECTION
}

export function directionLabel(direction: string | undefined): string {
  switch (direction) {
    case 'down':
      return 'descendant (down)'
    case 'both':
      return 'les deux sens (both)'
    case 'up':
      return 'montant (up)'
    default:
      return '—'
  }
}

export interface AuditDeclaration {
  site: string
  link_type: string
  duration: number
  direction: AuditDirection
}

export interface AuditDeclarationInput {
  site: string
  link_type: string
  duration: string | number
  direction: string
}

export type AuditDeclarationErrors = Partial<Record<'site' | 'link_type' | 'duration', string>>

export function validateAuditDeclaration(input: AuditDeclarationInput): {
  valid: boolean
  errors: AuditDeclarationErrors
  value: AuditDeclaration | null
} {
  const errors: AuditDeclarationErrors = {}
  const site = input.site.trim()
  if (!site) errors.site = 'requis'
  if (!(AUDIT_LINK_TYPES as readonly string[]).includes(input.link_type)) errors.link_type = 'type de lien inconnu'
  const duration = normalizeDuration(input.duration)
  if (duration === null) errors.duration = '10–600 s'
  const valid = Object.keys(errors).length === 0
  return {
    valid,
    errors,
    value: valid
      ? { site, link_type: input.link_type, duration: duration as number, direction: normalizeDirection(input.direction) }
      : null,
  }
}

// Le chemin des phases suit le sens déclaré : un audit descendant n'a que la
// charge descendante, un « both » montre les deux fenêtres distinctement.
export function auditPhases(direction: AuditDirection): string[] {
  if (direction === 'both') return ['repos', 'charge', 'charge-down', 'drain']
  return direction === 'down' ? ['repos', 'charge-down', 'drain'] : ['repos', 'charge', 'drain']
}

export function phaseLabel(phase: string | undefined): string {
  switch (phase) {
    case 'repos':
      return 'Repos'
    case 'charge':
      return 'Charge'
    case 'charge-down':
      return 'Charge descendante'
    case 'drain':
      return 'Drain'
    default:
      return phase || '—'
  }
}

// Percentiles client : même interpolation linéaire que metrics.Percentile.
// Un 0 est une sonde perdue, pas une mesure nulle — il ne tire pas le p95 vers
// le bas (même convention que lastNonZero côté Go).
export function percentile(samples: number[], p: number): number | null {
  const xs = samples.filter(v => Number.isFinite(v) && v > 0).sort((a, b) => a - b)
  if (xs.length === 0) return null
  if (xs.length === 1) return xs[0]
  const rank = (p / 100) * (xs.length - 1)
  const lo = Math.floor(rank)
  const hi = Math.ceil(rank)
  return xs[lo] * (1 - (rank - lo)) + xs[hi] * (rank - lo)
}

export interface LiveStats {
  rttP50: number | null
  rttP95: number | null
  smallP50: number | null
  smallP95: number | null
}

export function liveStats(rtt: number[] | undefined, small: number[] | undefined): LiveStats {
  return {
    rttP50: percentile(rtt ?? [], 50),
    rttP95: percentile(rtt ?? [], 95),
    smallP50: percentile(small ?? [], 50),
    smallP95: percentile(small ?? [], 95),
  }
}

export function remainingSeconds(live: AuditLive | null): number | null {
  if (!live?.running || live.total_sec == null || live.elapsed_sec == null) return null
  return Math.max(0, Math.round(live.total_sec - live.elapsed_sec))
}
