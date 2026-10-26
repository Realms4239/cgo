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
  drops: number
  wasted_bytes?: number
  cost_ar_per_h?: number
  deadline_ok_pct?: number
  gates: (boolean | null)[]
  running: boolean
}

export type GateIdx = 0|1|2|3|4|5|6|7

// ── Audit vivant (GET /api/audit/live) — running:false seul quand rien ne tourne
export type AuditDirection = 'up' | 'down' | 'both'

export interface AuditLive {
  running: boolean
  phase?: string
  elapsed_sec?: number
  total_sec?: number
  phase_remain_sec?: number
  last_rtt_ms?: number
  last_small_ms?: number
  samples?: { rtt_ms?: number[]; small_ms?: number[] }
}

// ── Rapport d'audit (GET /api/audit/status, /api/audit/report) — miroir des tags Go.
export interface AuditUnavailable { metric: string; reason: string }

export interface AuditMeasured {
  rtt_idle_p50_ms?: number
  rtt_idle_p95_ms?: number
  rtt_loaded_p50_ms?: number
  rtt_loaded_p95_ms?: number
  qdi_ms?: number
  bloat_delta_ms?: number
  drain_ms?: number
  drain_measured?: boolean
  load_measured?: boolean
  throughput_mbps?: number
  loss_pct?: number
  http_small_p95_ms?: number
  small_measured?: boolean
  voip_r?: number
  data_used_mb?: number
  direction?: string
  flow_count?: number
  jfi_pct?: number
  rtt_loaded_down_p95_ms?: number
  throughput_down_mbps?: number
  bloat_delta_down_ms?: number
  jfi_down_pct?: number
  load_measured_down?: boolean
  grade_down?: string
}

export interface AuditPhases {
  baseline_seconds?: number
  load_seconds?: number
  recovery_seconds?: number
}

export interface AuditProvenance {
  source_file?: string
  source_sha256?: string
  row_count?: number
  telemetry_file?: string
  telemetry_sha256?: string
  row_file_sha256?: string
}

export interface AuditRemediation {
  qdisc: string
  reason?: string
  linux_tc?: string[]
  routeros?: string
}

export interface AuditEconomicsScenario {
  model: string
  name: string
  ar_per_hour: number
  basis: string
}

export interface AuditEconomics {
  primary_link_type?: string
  primary?: AuditEconomicsScenario
  scenarios?: AuditEconomicsScenario[]
  retrans_ratio?: { ratio?: number; basis?: string }
  wasted_bytes?: number
  transported_bytes?: number
  load_seconds?: number
  disclaimer?: string
}

export interface AuditCapsla {
  measured?: boolean
  ok_pct?: number
  breach?: boolean
  label?: string
}

export interface AuditReport {
  audit_id: string
  timestamp?: string
  site?: string
  link_type?: string
  provider?: string
  measured: AuditMeasured
  phases?: AuditPhases
  unavailable?: AuditUnavailable[]
  bloat_grade?: string
  bloat_verdict?: string
  provenance?: AuditProvenance
  remediation?: AuditRemediation
  economics?: AuditEconomics
  cap_sla?: AuditCapsla
}
