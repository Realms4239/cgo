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
export const AUDIT_DURATIONS = [60] as const
export const AUDIT_DURATION_SECONDS = 60 as const
export function auditDuration(): 60 { return AUDIT_DURATION_SECONDS }
export const DEFAULT_AUDIT_DIRECTION: AuditDirection = 'up'

// La durée n'est pas négociable : 10 s repos, 40 s charge, 10 s drain.
export function normalizeDuration(value: string | number): number | null {
  const n = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(n)) return null
  return Math.trunc(n) === AUDIT_DURATION_SECONDS ? AUDIT_DURATION_SECONDS : null
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
  if (duration === null) errors.duration = 'audit limité à 60 s'
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

export function auditPhasePlan(_duration: number = AUDIT_DURATION_SECONDS): AuditPhasePlan {
  return { idle: 10, loaded: 40, drain: 10 }
}

export function phasePlanLabel(_duration: number = AUDIT_DURATION_SECONDS): string {
  return '60 s au total : 10 s repos · 40 s charge · 10 s drain';
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
