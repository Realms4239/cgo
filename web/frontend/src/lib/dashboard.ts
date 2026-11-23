// Langage opérateur du tableau de bord : les chiffres deviennent des mots,
// les seuils restent ceux des Réglages. Aucun zéro ne se lit comme mesuré.
import type { AuditReport } from './types'
import { fmtCount, formatCostAR, numFr } from './format'

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
  const kernel = operatorKernelEvidence(report)
  const cost = report?.economics?.primary
  const alertsPct = finite(s.alertsPct) ? s.alertsPct : undefined
  // §2 — quand eBPF est primaire, le fait « Réponse sous charge » EST la
  // médiane SRTT du noyau. Le RTT userspace ne devient jamais le chiffre
  // primary ; une sous-donnée eBPF manquante reste une absence, pas un repli.
  const delayMs = kernel.primary
    ? kernel.srttMs ?? undefined
    : finite(s.delayMs) ? s.delayMs : undefined
  const delayEvidenceBand = kernel.primary
    ? kernel.srttMs != null ? delayBand(kernel.srttMs, warnMs, critMs) : null
    : s.delay
  const rawAr = cost?.ar_per_hour
  const arPerHour = finite(rawAr) ? rawAr : undefined
  const costFigure = arPerHour != null ? formatCostAR(arPerHour) : null
  // Les deux objectifs viennent du rapport scellé, jamais d'une constante
  // retapée ici : un rapport ancien n'en porte pas et n'affiche rien.
  const sla = report?.cap_sla
  const targetPct = sla?.target_pct && sla.target_pct > 0 ? `objectif ${numFr(sla.target_pct, 0)} %` : ''
  const deadlineMs = sla?.deadline_ms && sla.deadline_ms > 0 ? `objectif ${numFr(sla.deadline_ms, 0)} ms` : ''
  return [
    {
      key: 'alerts', label: 'Alertes à l’heure',
      value: alertsPct != null ? numFr(alertsPct, 1) : 'Pas encore mesuré',
      unit: alertsPct != null ? '%' : '', target: targetPct, band: s.alerts,
    },
    {
      key: 'delay', label: 'Réponse sous charge',
      value: delayMs != null ? numFr(delayMs, kernel.primary ? 1 : 0) : 'Pas encore mesuré',
      unit: delayMs != null ? 'ms' : '', target: deadlineMs, band: delayEvidenceBand,
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

// §2 — eBPF est la mesure primaire du noyau. Pertes et latence viennent des
// flux eBPF quand ils sont là ; le tc reste une ligne nommée à part, jamais
// additionnée au compteur eBPF. Sans flux valide, rien n'est fabriqué : le mot
// d'absence et la vraie raison remplacent le chiffre. tcp_retransmit_skb est une
// étiquette de regroupement, pas une cause vérifiée (HUATUO) : « observées ».
export interface KernelLine { text: string; absent: boolean }

export interface KernelEvidence {
  primary: boolean
  complete: boolean
  reason?: string
  retransEvents: number | null
  srttMs: number | null
  tcDrops: number | null
  loss: KernelLine
  latency: KernelLine
  tcCrossCheck: KernelLine
}

const KERNEL_ABSENT_WORD = 'non mesuré'

function absentKernelLine(reason?: string): KernelLine {
  return { text: `— ${KERNEL_ABSENT_WORD}${reason ? ` — ${reason}` : ''}`, absent: true }
}

function medianOf(values: number[]): number | null {
  if (values.length === 0) return null
  const sorted = [...values].sort((a, b) => a - b)
  const mid = Math.floor(sorted.length / 2)
  return sorted.length % 2 === 1 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2
}

// Une seule évidence noyau, consommée par operatorFacts — lu à la fois par le
// cockpit et par le tiroir de preuves. Aucun repli sur le RTT userspace ici :
// ce serait un vert fabriqué à partir d'une mesure qui n'est pas la mesure noyau.
export function operatorKernelEvidence(report: AuditReport | null | undefined): KernelEvidence {
  const k = report?.kernel_telemetry
  const reason = k?.reason?.trim() || undefined
  const flows = Array.isArray(k?.flows) ? k?.flows ?? [] : []
  const primary = k?.available === true && k?.source === 'ebpf' && flows.length > 0

  if (!primary) {
    const absent = absentKernelLine(reason)
    return {
      primary: false, complete: false, reason,
      retransEvents: null, srttMs: null, tcDrops: null,
      loss: absent, latency: absent, tcCrossCheck: absent,
    }
  }

  const retransValues = flows.map(f => f.retrans_events).filter(finite)
  const srttValues = flows.map(f => f.srtt_us).filter(v => finite(v) && v > 0)
  const tcValues = flows.map(f => f.qdisc_drop_packets).filter(finite)
  const reportedTcDrops = k?.qdisc_drop_packets

  const retransEvents = retransValues.length ? retransValues.reduce((a, b) => a + b, 0) : null
  const srttUs = medianOf(srttValues)
  const tcDrops = tcValues.length
    ? tcValues.reduce((a, b) => a + b, 0)
    : finite(reportedTcDrops) ? reportedTcDrops : null

  return {
    primary: true,
    complete: retransEvents != null && srttUs != null && tcDrops != null,
    retransEvents,
    srttMs: srttUs == null ? null : srttUs / 1000,
    tcDrops,
    loss: retransEvents == null
      ? absentKernelLine('retrans_events non rapportés')
      : { text: `${fmtCount(retransEvents)} retransmissions observées`, absent: false },
    latency: srttUs == null
      ? absentKernelLine('SRTT non rapporté')
      : { text: `${numFr(srttUs / 1000, 1)} ms — SRTT médian eBPF`, absent: false },
    tcCrossCheck: tcDrops == null
      ? absentKernelLine('contrôle tc/qdisc non rapporté')
      : { text: `tc/qdisc — ${fmtCount(tcDrops)} paquets écartés (contrôle croisé)`, absent: false },
  }
}

export type KernelHealthKind = 'ok' | 'attention' | 'critical' | 'absent'

export interface KernelHealth {
  kind: KernelHealthKind
  detail: string | null
}

// La santé n'est verte que si le noyau la porte. Des retransmissions observées
// ne sont pas une cause vérifiée (HUATUO), mais elles interdisent « Lien prêt ».
export function kernelHealth(evidence: KernelEvidence, warnMs: number, critMs: number): KernelHealth {
  if (!evidence.primary || evidence.srttMs == null) return { kind: 'absent', detail: null }
  const band = delayBand(evidence.srttMs, warnMs, critMs)
  let kind: KernelHealthKind =
    band?.level === 'crit' ? 'critical' : band?.level === 'warn' ? 'attention' : 'ok'
  // Une charge partielle ne porte pas le vert : le SRTT réel reste affiché,
  // mais « Lien prêt » exige le trio complet retrans/SRTT/tc.
  if (!evidence.complete) kind = 'attention'
  if (kind === 'ok' && evidence.retransEvents != null && evidence.retransEvents > 0) kind = 'attention'
  return { kind, detail: `${evidence.latency.text} · ${evidence.loss.text}` }
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
