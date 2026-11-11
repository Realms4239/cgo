// Fixtures de rapport scellé pour les tests — UN jeu complet par type ; chaque test
// n'échange que les champs qui le concernent. Les valeurs par défaut disent « rien
// n'a été mesuré » : elles ne prétendent pas à des mesures inventées.
// Ces fixtures existent parce que le miroir TS est GÉNÉRÉ (types.gen.ts) : un champ
// ajouté en Go manque ici d'abord, et le compilateur le dit.
import type {
  AuditCapsla,
  AuditEconomics,
  AuditEconomicsNumbers,
  AuditEconomicsScenario,
  AuditMeasured,
  AuditPhases,
  AuditProvenance,
  AuditRemediation,
  AuditReport,
  AuditVerdict,
  KernelTelemetry,
  RouterOSValidation,
} from './types'

export function measured(p: Partial<AuditMeasured> = {}): AuditMeasured {
  return {
    rtt_idle_p50_ms: 0,
    rtt_idle_p95_ms: 0,
    rtt_loaded_p50_ms: 0,
    rtt_loaded_p95_ms: 0,
    qdi_ms: 0,
    bloat_delta_ms: 0,
    drain_ms: 0,
    drain_measured: false,
    load_attempted: false,
    load_measured: false,
    throughput_mbps: 0,
    loss_pct: 0,
    http_small_p95_ms: 0,
    small_measured: false,
    voip_r: 0,
    data_used_mb: 0,
    direction: 'up',
    flow_count: 1,
    jfi_pct: 0,
    jfi_measured: false,
    rtt_loaded_down_p95_ms: 0,
    throughput_down_mbps: 0,
    bloat_delta_down_ms: 0,
    jfi_down_pct: 0,
    load_measured_down: false,
    grade_down: '',
    ...p,
  }
}

export function phases(p: Partial<AuditPhases> = {}): AuditPhases {
  return { baseline_seconds: 10, load_seconds: 40, recovery_seconds: 10, ...p }
}

export function provenance(p: Partial<AuditProvenance> = {}): AuditProvenance {
  return {
    source_file: 'data/runs/run-1/aqm_eval.csv',
    source_sha256: 'a'.repeat(64),
    row_count: 100,
    ...p,
  }
}

export function capsla(p: Partial<AuditCapsla> = {}): AuditCapsla {
  return {
    measured: false,
    ok_pct: 0,
    breach: false,
    target_pct: 95,
    deadline_ms: 220,
    ...p,
  }
}

export function verdict(p: Partial<AuditVerdict> = {}): AuditVerdict {
  return { level: '', word: '', sentence: '', ...p }
}

export function economicsNumbers(p: Partial<AuditEconomicsNumbers> = {}): AuditEconomicsNumbers {
  return {
    model: 'B',
    retrans_ratio: 0,
    wasted_bytes: 0,
    transport_bytes: 0,
    load_seconds: 0,
    ...p,
  }
}

export function economicsScenario(p: Partial<AuditEconomicsScenario> = {}): AuditEconomicsScenario {
  return { model: 'B', name: 'B2B APN', ar_per_hour: 0, basis: 'base', numbers: economicsNumbers(), ...p }
}

export function economics(p: Partial<AuditEconomics> = {}): AuditEconomics {
  return {
    primary_link_type: 'fiber',
    scenarios: [economicsScenario()],
    retrans_ratio: { ratio: 0, basis: '' },
    wasted_bytes: 0,
    transported_bytes: 0,
    load_seconds: 0,
    disclaimer: 'scénario, pas une affirmation sur le contrat réel du site',
    ...p,
  }
}

export function remediation(p: Partial<AuditRemediation> = {}): AuditRemediation {
  return { qdisc: 'cake', reason: 'fq_codel en défaut', linux_tc: ['tc qdisc replace dev eth0 root handle 1: cake'], ...p }
}

export function routerosValidation(p: Partial<RouterOSValidation> = {}): RouterOSValidation {
  return { version: '7.18.2', level: 'vérifiée', export_sha256: 'b'.repeat(64), checks: [], ...p }
}

export function kernelTelemetry(p: Partial<KernelTelemetry> = {}): KernelTelemetry {
  return { available: false, ...p }
}

export function report(p: Partial<AuditReport> = {}): AuditReport {
  return {
    audit_id: 'audit-fixture',
    timestamp: '2026-11-01T10:00:00Z',
    site: 'Toamasina',
    link_type: 'fiber',
    provider: 'blueline',
    measured: measured(),
    phases: phases(),
    unavailable: [],
    bloat_grade: '',
    bloat_verdict: '',
    operator_verdict: verdict(),
    provenance: provenance(),
    remediation_state: 'observation_only',
    remediation_available: false,
    cap_sla: capsla(),
    ...p,
  }
}
