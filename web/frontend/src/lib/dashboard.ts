// Langage opérateur du tableau de bord : les chiffres deviennent des mots,
// les seuils restent ceux des Réglages. Aucun zéro ne se lit comme mesuré.
import type { AuditReport } from './types'
import { formatCostAR, numFr } from './format'
import { pickHeadline, type SynGroup } from './synthese'

export type PlainLevel = 'ok' | 'warn' | 'crit'
export interface Band { word: string; level: PlainLevel }

const finite = (v: number | null | undefined): v is number =>
  typeof v === 'number' && Number.isFinite(v)

// Quatre mots pour un opérateur : faible, modéré, élevé, critique. Le dernier
// palier commence à 4 × le seuil critique — 328 ms reste « élevé », pas
// « critique » quand le seuil est 100 ms.
export function delayBand(ms: number | null | undefined, warnMs: number, critMs: number): Band | null {
  if (!finite(ms) || ms <= 0) return null
  if (ms <= warnMs) return { word: 'faible', level: 'ok' }
  if (ms <= critMs) return { word: 'modéré', level: 'warn' }
  if (ms <= critMs * 4) return { word: 'élevé', level: 'warn' }
  return { word: 'critique', level: 'crit' }
}

export function alertsBand(pct: number | null | undefined): Band | null {
  if (!finite(pct)) return null
  if (pct >= 95) return { word: "à l'heure", level: 'ok' }
  if (pct >= 80) return { word: 'limite', level: 'warn' }
  return { word: 'en retard', level: 'crit' }
}

export function instabilityBand(qdiMs: number | null | undefined): Band | null {
  if (!finite(qdiMs)) return null
  if (qdiMs <= 50) return { word: 'stable', level: 'ok' }
  if (qdiMs <= 200) return { word: 'modérée', level: 'warn' }
  return { word: 'sévère', level: 'crit' }
}

export function qualityWord(grade: string | undefined): string {
  if (!grade) return 'non noté'
  const g = grade.charAt(0).toUpperCase()
  if (g === 'A') return 'excellent'
  if (g === 'B') return 'bon'
  if (g === 'C') return 'moyen'
  if (g === 'D' || g === 'E' || g === 'F') return 'faible'
  return 'non noté'
}

export interface OperatorSummary {
  delayMs?: number
  delay: Band | null
  alertsPct?: number
  alerts: Band | null
  grade?: string
  quality: string
}

// Le même résumé alimente le tableau de bord et la lecture d'audit : un seul
// calcul, deux présentations — plus de divergence de seuil ou de libellé.
export function operatorSummary(
  report: Pick<AuditReport, 'measured' | 'cap_sla' | 'bloat_grade'> | null | undefined,
  warnMs: number,
  critMs: number,
): OperatorSummary {
  const m = report?.measured
  const raw = m?.small_measured ? m?.http_small_p95_ms : m?.rtt_loaded_p95_ms
  const delayMs = typeof raw === 'number' && raw > 0 ? raw : undefined
  const alertsPct = report?.cap_sla?.measured ? report.cap_sla.ok_pct : undefined
  return {
    delayMs,
    delay: delayBand(delayMs, warnMs, critMs),
    alertsPct,
    alerts: alertsBand(alertsPct),
    grade: report?.bloat_grade,
    quality: qualityWord(report?.bloat_grade),
  }
}

export interface RemediationCell {
  qdisc: string
  cc: string
  deadlinePct?: number
  smallP95Ms?: number
  n: number
}

export interface RemediationView {
  simple: RemediationCell
  best: RemediationCell
}

const cellOf = (g: SynGroup): RemediationCell => ({
  qdisc: g.qdisc,
  cc: g.cc,
  deadlinePct: finite(g.deadline_median) ? g.deadline_median : undefined,
  smallP95Ms: finite(g.small_p95_median) && g.small_p95_median > 0 ? g.small_p95_median : undefined,
  n: g.deadline_valid_n ?? 0,
})

// Seules les cellules P2 avec échantillon valide entrent dans la projection.
export function remediationView(groups: SynGroup[] | null | undefined): RemediationView | null {
  const h = pickHeadline(Array.isArray(groups) ? groups : [])
  if (!h.simple || !h.best) return null
  return { simple: cellOf(h.simple), best: cellOf(h.best) }
}

export interface Fact { key: string; label: string; value: string; unit: string; target: string; band: Band | null }

// Les quatre faits de la console. Un seul calcul, partagé par le tableau de
// bord et l'audit : le verdict et les faits lisent les mêmes champs.
// Une valeur absente OU illisible devient un mot : une cellule vide, comme un
// zéro, se lirait comme une mesure. Rien ici ne dépasse la lecture d'un champ.
export function operatorFacts(
  report: AuditReport | null | undefined,
  warnMs: number,
  critMs: number,
): Fact[] {
  const s = operatorSummary(report, warnMs, critMs)
  const cost = report?.economics?.primary
  const alertsPct = finite(s.alertsPct) ? s.alertsPct : undefined
  const delayMs = finite(s.delayMs) ? s.delayMs : undefined
  const rawAr = cost?.ar_per_hour
  const arPerHour = finite(rawAr) ? rawAr : undefined
  const costFigure = arPerHour != null ? formatCostAR(arPerHour) : null
  return [
    {
      key: 'alerts', label: 'Alertes à l’heure',
      value: alertsPct != null ? numFr(alertsPct, 1) : 'Pas encore mesuré',
      unit: alertsPct != null ? '%' : '', target: 'objectif 95 %', band: s.alerts,
    },
    {
      key: 'delay', label: 'Réponse sous charge',
      value: delayMs != null ? numFr(delayMs, 0) : 'Pas encore mesuré',
      unit: delayMs != null ? 'ms' : '', target: 'objectif 220 ms', band: s.delay,
    },
    {
      key: 'quality', label: 'Qualité du lien',
      value: s.grade || 'Pas encore mesuré', unit: '',
      target: s.grade ? `note ${s.grade}` : 'aucune note', band: s.grade ? { word: s.quality, level: qualityLevel(s.grade) } : null,
    },
    {
      key: 'cost', label: 'Capacité gaspillée',
      value: costFigure ? costFigure.num : 'Non chiffré',
      unit: costFigure ? costFigure.unit : '',
      target: cost?.name ? `scénario ${scenarioWord(cost.model, cost.name)}` : 'aucun scénario retenu',
      band: costFigure ? { word: 'estimation contractuelle', level: 'warn' as PlainLevel } : null,
    },
  ]
}

// La lettre du moteur devient un mot, au même palier que la note affichée.
function qualityLevel(grade: string): PlainLevel {
  const g = grade.charAt(0).toUpperCase()
  if (g === 'A' || g === 'B') return 'ok'
  if (g === 'C') return 'warn'
  return 'crit'
}

// Le nom du scénario reste technique dans le tiroir ; ici il devient lisible.
function scenarioWord(model: string | undefined, name: string): string {
  if (model === 'A') return 'ligne dédiée'
  if (model === 'B') return 'forfait entreprise'
  if (model === 'emergency') return 'hors forfait'
  return name
}

// Borne de secours détail — visible, jamais présentée comme un modèle d'infra.
export function retailBound(report: Pick<AuditReport, 'economics'> | null | undefined): number | null {
  const s = report?.economics?.scenarios?.find(x => x.model === 'emergency')
  return s && finite(s.ar_per_hour) ? s.ar_per_hour : null
}

// Deux cartes vides côte à côte pèsent plus qu'un fait : quand ni le coût ni
// la comparaison ne portent de valeur, une seule bande les dit — même règle
// que les cellules compactes sur téléphone.
export function bothAbsent(hasCost: boolean, hasChange: boolean): boolean {
  return !hasCost && !hasChange
}
