import type { AuditStage } from './auditFlow'
import type { AuditLive, AuditReport } from './types'

export interface RequestFailure {
  code: string
  message: string
  retry?: () => void
}

export type AuditViewState =
  | { status: 'idle'; lastAudit: AuditReport | null }
  | { status: 'loading_stale'; previous: AuditReport; startedAt: number }
  | { status: 'running'; phase: 'repos' | 'charge' | 'vidange'; elapsedSec: number; progressPct: number }
  | { status: 'error'; code: string; message: string; retryAction: () => void }
  | { status: 'complete'; report: AuditReport }

const noop = () => undefined

function toPhase(value: string | undefined): 'repos' | 'charge' | 'vidange' {
  if (value === 'drain') return 'vidange'
  if (value === 'charge-down' || value === 'charge') return 'charge'
  return 'repos'
}

export function selectAuditViewState(input: {
  stage: AuditStage
  live: AuditLive | null
  report: AuditReport | null
  failure: RequestFailure | null
}): AuditViewState {
  if (input.failure) {
    return {
      status: 'error',
      code: input.failure.code,
      message: input.failure.message,
      retryAction: input.failure.retry ?? noop,
    }
  }
  if (input.stage === 'done' && input.report) {
    return { status: 'complete', report: input.report }
  }
  if (input.stage === 'observing' && input.live) {
    const total = input.live.total_sec ?? 60
    const elapsed = Math.max(0, Math.min(total, input.live.elapsed_sec ?? 0))
    return {
      status: 'running',
      phase: toPhase(input.live.phase),
      elapsedSec: elapsed,
      progressPct: Math.round((elapsed / total) * 100),
    }
  }
  if (input.stage === 'observing' && input.report) {
    return { status: 'loading_stale', previous: input.report, startedAt: Date.now() }
  }
  return { status: 'idle', lastAudit: input.report }
}
