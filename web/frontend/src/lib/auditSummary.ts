// Résumé d'audit conservé localement — ce que le tableau de bord compare
// entre deux audits du MÊME site et du MÊME type de lien. Aucun chiffre
// absent n'est remplacé par zéro.
import type { AuditReport } from './types'

export interface AuditSummary {
  audit_id: string
  site: string
  link_type: string
  timestamp?: string
  small_p95_ms?: number
  deadline_ok_pct?: number
  grade?: string
}

export interface AuditComparison {
  comparable: boolean
  reason?: string
  smallDeltaPct?: number
  deadlineDelta?: number
}

const KEY = 'cgo-audit-last'

const positive = (v: number | undefined | null): number | undefined =>
  typeof v === 'number' && Number.isFinite(v) && v > 0 ? v : undefined

export function summaryOf(report: AuditReport): AuditSummary {
  const m = report.measured
  const small = m.small_measured ? m.http_small_p95_ms : m.rtt_loaded_p95_ms
  const cap = report.cap_sla
  return {
    audit_id: report.audit_id,
    site: (report.site ?? '').trim(),
    link_type: report.link_type ?? '',
    timestamp: report.timestamp,
    small_p95_ms: positive(small),
    deadline_ok_pct: cap?.measured ? cap.ok_pct : undefined,
    grade: report.bloat_grade,
  }
}

export function compareAudits(
  previous: AuditSummary | null | undefined,
  current: AuditSummary | null | undefined,
): AuditComparison {
  if (!previous || !current) return { comparable: false, reason: 'aucun audit précédent' }
  if (!previous.site || previous.site !== current.site) {
    return { comparable: false, reason: 'site différent' }
  }
  if (!previous.link_type || previous.link_type !== current.link_type) {
    return { comparable: false, reason: 'type de lien différent' }
  }
  const out: AuditComparison = { comparable: true }
  if (previous.small_p95_ms && current.small_p95_ms) {
    const pct = ((current.small_p95_ms - previous.small_p95_ms) / previous.small_p95_ms) * 100
    out.smallDeltaPct = Math.round(pct * 10) / 10
  }
  if (previous.deadline_ok_pct != null && current.deadline_ok_pct != null) {
    out.deadlineDelta = Math.round((current.deadline_ok_pct - previous.deadline_ok_pct) * 10) / 10
  }
  return out
}

function storage(): Storage | null {
  try {
    return typeof localStorage === 'undefined' ? null : localStorage
  } catch {
    return null
  }
}

export function loadLastSummary(): AuditSummary | null {
  const s = storage()
  if (!s) return null
  try {
    const raw = s.getItem(KEY)
    if (!raw) return null
    const v = JSON.parse(raw) as AuditSummary
    return v && typeof v.audit_id === 'string' && typeof v.site === 'string' ? v : null
  } catch {
    return null
  }
}

export function saveLastSummary(summary: AuditSummary): void {
  const s = storage()
  if (!s) return
  try {
    s.setItem(KEY, JSON.stringify(summary))
  } catch {
    /* quota / mode privé : la comparaison reste absente, jamais inventée */
  }
}
