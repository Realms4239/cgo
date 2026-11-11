export interface LiveFrame {
  ts: number
  phase: string
  profile: string
  qdisc: string
  cc: string
  repetition: number
  event_id: number
  total_events?: number
  done_events?: number
  phase_total_s?: number
  phaseSince?: Record<string, number>
  load_status: string
  rtt_p50_ms: number
  rtt_p95_ms: number
  small_p95_ms: number
  bulk_goodput_mbps: number
  // Interface observée (additif, absent quand CGO_IFACE est vide).
  // iface_valid=false invalide tous les champs numériques.
  iface_name?: string
  iface_source?: 'twin' | 'prod' | 'unknown'
  iface_rx_mbps?: number
  iface_tx_mbps?: number
  iface_rx_drops?: number
  iface_tx_drops?: number
  iface_window_ms?: number
  iface_sample_ts?: number
  iface_valid?: boolean
  iface_gap?: string
  drops: number
  wasted_bytes?: number
  cost_ar_per_h?: number
  // Coûts entreprise (miroir de pkg/campagne.Snapshot) : scénario GOLD (modèle A)
  // et B2B APN 150 Ar/Go (modèle B) — le cost_ar_per_h détail reste à côté.
  cost_primary_ar_per_h?: number
  cost_b2b_ar_per_h?: number
  // JFI mesuré par le serveur, 0–100 (0 = mono-flux/n.a.) — préféré aux rings
  // client quand > 0, sinon computeJFI côté mur.
  jfi_pct?: number
  deadline_ok_pct?: number
  gates: (boolean | null)[]
  running: boolean
}

export type GateIdx = 0|1|2|3|4|5|6|7

// ── Audit vivant (GET /api/audit/live) — running:false seul quand rien ne tourne
export type AuditDirection = 'up' | 'down' | 'both'

// Une seconde muette porte n=0 et des valeurs null : elle ne se lit jamais
// comme un 0 ms.
export interface AuditTimelinePoint {
  t: number
  phase: string
  n: number
  rtt_p50: number | null
  rtt_p95: number | null
  small_p95: number | null
}

export interface AuditLive {
  running: boolean
  phase?: string
  elapsed_sec?: number
  total_sec?: number
  phase_remain_sec?: number
  last_rtt_ms?: number
  last_small_ms?: number
  samples?: {
    rtt_ms?: number[]
    small_ms?: number[]
  }
  timeline?: AuditTimelinePoint[]
  deadline_ms?: number
}

// ── Rapport d'audit scellé (GET /api/audit/report) — types GÉNÉRÉS, jamais à la main.
// `audit/cmd/typesgen` les régénère depuis les structs Go tagées json ; un champ
// ajouté en Go paraît en TS à la génération, sinon le générateur s'arrête.
// Sémantique d'absence (voir la doc des structs Go) :
//   un float64 sans omitempty est TOUJOURS émis — 0 signifie « non mesuré pour ce
//   champ-là » (drain_ms, jfi_pct…), jamais un zéro vrai ; les drapeaux de mesure
//   (load_measured, drain_measured, small_measured, jfi_measured) disent ce qui a
//   été mesuré ; les champs omitempty (remediation, economics, invocation) sont
//   ABSENTS d'un rapport scellé avant ces champs.
export type {
  AuditReport,
  AuditMeasured,
  AuditPhases,
  AuditUnavailable,
  AuditProvenance,
  AuditInvocation,
  AuditRemediation,
  AuditRemediationState,
  AuditEconomics,
  AuditEconomicsScenario,
  AuditEconomicsNumbers,
  AuditRetransRatio,
  AuditCapsla,
  AuditVerdict,
  AuditVerdictLevel,
  RouterOSCheck,
  RouterOSValidation,
  KernelTelemetry,
  KernelFlow,
} from './types.gen'
