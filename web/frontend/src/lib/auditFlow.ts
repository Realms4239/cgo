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

// Miroir exact de `phasePlan` (pkg/audit/audit.go) : la durée déclarée est la
// durée TOTALE de l'audit, partagée en repos / charge / drain. L'écran le dit
// avant de lancer, au lieu de laisser l'opérateur deviner ce que « 20 s » veut dire.
export interface AuditPhasePlan { idle: number; loaded: number; drain: number }

export function auditPhasePlan(duration: number): AuditPhasePlan {
  const d = duration > 0 ? duration : 10
  let idle: number
  let drain: number
  if (d >= 60) { idle = 10; drain = 10 }
  else if (d >= 20) { idle = Math.floor(d / 4); drain = Math.floor(d / 4) }
  else { idle = Math.floor(d / 2); drain = 0 }
  if (idle < 1) idle = 1
  const loaded = Math.max(0, d - idle - drain)
  return { idle, loaded, drain }
}

export function phasePlanLabel(duration: number): string {
  const p = auditPhasePlan(duration)
  return `${duration} s au total : ${p.idle} s repos · ${p.loaded} s charge${p.drain ? ` · ${p.drain} s drain` : ''}`
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

export function remainingSeconds(live: AuditLive | null): number | null {
  if (!live?.running || live.total_sec == null || live.elapsed_sec == null) return null
  return Math.max(0, Math.round(live.total_sec - live.elapsed_sec))
}
