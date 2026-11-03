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
  // Une fenêtre de charge était programmée (load_attempted), pas forcément
  // établie (load_measured). Le moteur seul en tire un mot ; la console lit les
  // deux champs tels quels.
  load_attempted?: boolean
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
  linux_tc_ingress?: string[]
  routeros_ingress?: string
  up_measured_mbps?: number
  up_shaped_mbps?: number
  down_measured_mbps?: number
  down_shaped_mbps?: number
  linux_iface?: string
  routeros_wan_iface?: string
  routeros_lan_iface?: string
  routeros_lan_assumed?: boolean
}

// Les clés lues par la console ; le moteur en émet un superset (tranche 20A).
// Un modèle ne consomme qu'une partie des entrées : les autres sont omises du
// JSON plutôt que montrées mesurées.
export interface AuditEconomicsNumbers {
  model: string
  monthly_opex_mga?: number
  cir_mbps?: number
  ar_per_gb?: number
  retrans_ratio?: number
  wasted_bytes?: number
  transport_bytes?: number
  load_seconds?: number
}

export interface AuditEconomicsScenario {
  model: string
  name: string
  ar_per_hour: number
  basis: string
  numbers?: AuditEconomicsNumbers
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
  target_pct?: number
  deadline_ms?: number
}

export interface RouterOSCheck {
  name: string
  expected: string
  found: string
  pass: boolean
}

export interface RouterOSValidation {
  version?: string
  level?: string
  export_sha256?: string
  proof_path?: string
  checks?: RouterOSCheck[]
}

// ── Télémétrie noyau (miroir de pkg/audit.KernelTelemetry) — les flux ne
// viennent que de l'objet eBPF ; TCP_INFO n'a pas de lignes par flux.
export interface KernelFlow {
  local: string
  peer: string
  retrans_payload_bytes: number
  retrans_events: number
  qdisc_drop_packets: number
  srtt_us: number
}

export interface KernelTelemetry {
  available?: boolean
  source?: string
  reason?: string
  object_sha256?: string
  retrans_payload_bytes?: number
  qdisc_drop_packets?: number
  flows_seen?: number
  flows_total?: number
  flows?: KernelFlow[]
}

export type AuditVerdictLevel = 'neutral' | 'red' | 'amber' | 'green'

// Verdict opérateur calculé par le moteur (pkg/metrics.Verdict) : un mot et une
// phrase, affichés tels quels. Un rapport scellé avant la cascade rend les trois
// champs vides — omitempty ne supprime pas une structure —, d'où '' dans level.
export interface AuditVerdict {
  level: '' | AuditVerdictLevel
  word: string
  sentence: string
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
  operator_verdict?: AuditVerdict
  provenance?: AuditProvenance
  remediation?: AuditRemediation
  economics?: AuditEconomics
  cap_sla?: AuditCapsla
  routeros_validation?: RouterOSValidation
  kernel_telemetry?: KernelTelemetry
}
