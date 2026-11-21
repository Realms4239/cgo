import { directionLabel } from './auditFlow'
import type { AuditReport } from './types'

// Citation d'un rapport scellé : uniquement des champs du rapport, aucune
// valeur inventée. Les champs absents sont omis, jamais remplacés par zéro.
export function auditCitation(report: AuditReport): string {
  const id = report.audit_id?.trim()
  if (!id) return ''

  const site = report.site?.trim()
  const link = report.link_type?.trim()
  const provider = report.provider?.trim()
  const linkLabel = [link, provider].filter(Boolean).join('/')
  const dir = directionLabel(report.measured?.direction)

  const details = [site, linkLabel, report.timestamp?.trim()].filter(Boolean).join(', ')
  const loadMeasured = !!report.measured?.load_measured || !!report.measured?.load_measured_down
  const measurement = loadMeasured
    ? (dir !== '—' ? `direction ${dir}` : 'charge mesurée')
    : (dir !== '—' ? `direction ${dir} · charge non mesurée` : 'charge non mesurée')
  const tail = [details, measurement].filter(Boolean).join(', ')
  let out = tail ? `Audit CGO ${id} — ${tail}.` : `Audit CGO ${id}.`

  const seal = report.provenance?.source_sha256?.trim()
  if (seal) out += ` Rapport scellé : source_sha256 ${seal}.`
  return out
}
