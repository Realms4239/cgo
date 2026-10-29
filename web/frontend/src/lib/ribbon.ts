// Faits du bandeau d'état — une ligne par monde, jamais un compte inventé.
import type { AuditReport } from './types'

export interface IntegrityFacts {
  available?: boolean
  runs?: number
  valid?: number
  degraded?: number
  quarantined?: number
  hash8?: string
  updated?: string
}

const n = (v: number | undefined): string => (typeof v === 'number' && Number.isFinite(v) ? String(v) : '—')

export function corpusFacts(i: IntegrityFacts | null | undefined): string {
  if (!i || i.available === false) return 'corpus —'
  return `${n(i.runs)} runs · ${n(i.valid)} valides · ${n(i.degraded)} dégradées · ${n(i.quarantined)} écartées`
}

export function sealFacts(hash8?: string): string {
  return hash8 ? `sceau ${hash8}` : 'sceau —'
}

export function auditFacts(report: AuditReport | null | undefined): string {
  if (!report) return 'dernier audit —'
  const site = report.site?.trim() || 'site —'
  return `dernier audit · ${site}${report.bloat_grade ? ` · ${report.bloat_grade}` : ''}`
}

export function ribbonMode(mode?: string): { text: string; observation: boolean } {
  return mode === 'observe'
    ? { text: 'observation', observation: true }
    : { text: mode || 'complet', observation: false }
}
