// AuditReportCard — l'étape « Lire » : le rapport complet, sans cacher une
// absence. L'économie vient de `economics.*` (le P0 ne lit plus `waste.*`) et
// chaque entrée `unavailable` est rendue telle quelle.
import { Badge } from './ui/StatusPip'
import { CardHead } from './ui/CardHead'
import { Stat } from './ui/Stat'
import { Provenance } from './ui/Provenance'
import { KernelRing } from './KernelRing'
import { directionLabel } from '../lib/auditFlow'
import { auditCitation } from '../lib/citation'
import { formatBytesFR, formatCostAR, groupDigits } from '../lib/format'
import type {
  AuditCapsla,
  AuditEconomics,
  AuditEconomicsScenario,
  AuditReport,
} from '../lib/types'

type BadgeTone = 'steel' | 'success' | 'warning' | 'danger' | 'live'

// breach ⇒ danger, pass ⇒ success, non mesuré ⇒ steel — jamais un vert par défaut.
export function capTone(cap?: AuditCapsla): BadgeTone {
  if (!cap?.measured) return 'steel'
  return cap.breach ? 'danger' : 'success'
}

type ProofState = 'valid' | 'degraded' | 'observation' | 'missing'

function proofState(report: AuditReport): ProofState {
  // L’absence d’empreinte source est l’absence de preuve, jamais un succès.
  if (!report.audit_id || !report.provenance?.source_sha256) return 'missing'
  const measured = report.measured
  const hasLoad = !!measured?.load_measured || !!measured?.load_measured_down
  if (!hasLoad) return 'observation'
  const grade = report.bloat_grade?.trim()[0]?.toUpperCase()
  if ((report.cap_sla?.measured && report.cap_sla.breach) || grade === 'D' || grade === 'F') return 'degraded'
  return 'valid'
}

const PROOF_LABEL: Record<ProofState, string> = {
  valid: 'Valide',
  degraded: 'Dégradé',
  observation: 'Observation seule',
  missing: 'Preuve manquante',
}

function proofTone(state: ProofState): BadgeTone {
  return state === 'valid' ? 'success' : state === 'degraded' ? 'warning' : 'steel'
}

export function gradeTone(grade?: string): BadgeTone {
  if (!grade) return 'steel'
  if (grade.startsWith('A')) return 'success'
  if (grade.startsWith('B')) return 'live'
  if (grade.startsWith('C')) return 'warning'
  return 'danger'
}

export function primaryARPerHour(report: Pick<AuditReport, 'economics'>): number | null {
  return report.economics?.primary?.ar_per_hour ?? null
}

export function b2bScenario(econ?: AuditEconomics): AuditEconomicsScenario | undefined {
  const scenarios = econ?.scenarios ?? []
  return scenarios.find(s => s.model === 'B' && s.name === 'b2b-m2m-apn') ?? scenarios.find(s => s.model === 'B')
}

export function b2bARPerHour(report: Pick<AuditReport, 'economics'>): number | null {
  return b2bScenario(report.economics)?.ar_per_hour ?? null
}

function ms(v?: number): string {
  return v != null && v > 0 ? v.toFixed(1) : '—'
}

function bytes(v?: number): string {
  const b = formatBytesFR(v)
  return b.unit ? `${b.num} ${b.unit}` : '—'
}

function ratioBasis(basis?: string): string {
  if (basis === 'measured') return 'mesuré (TCP_INFO)'
  if (basis === 'estimated') return 'estimé (perte × volume)'
  return 'non mesuré'
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div style={{ borderTop: '1px solid var(--hairline)', paddingTop: 'var(--sp-2)' }}>
      <CardHead label={title} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)', marginTop: 'var(--sp-2)' }}>{children}</div>
    </div>
  )
}

function Kpi({ label, value, unit }: { label: string; value: string; unit?: string }) {
  return <div className="c-kpi"><Stat label={label} value={value} unit={unit} /></div>
}

function Muted({ children }: { children: React.ReactNode }) {
  return <div className="mono" style={{ fontSize: 'var(--fs-micro)', lineHeight: 1.5, color: 'var(--text-faint)' }}>{children}</div>
}

export function AuditReportCard({ report, onCopy }: {
  report: AuditReport
  onCopy: (text: string, label: string) => void
}) {
  const m = report.measured
  const load = !!m.load_measured
  const proof = proofState(report)
  const down = !!m.load_measured_down && m.direction !== 'down'
  const econ = report.economics
  const primary = econ?.primary
  const b2b = b2bScenario(econ)
  const scenarios = econ?.scenarios ?? []
  const cap = report.cap_sla
  const mdHref = report.audit_id
    ? `/api/audit/report?audit_id=${encodeURIComponent(report.audit_id)}&format=md`
    : '/api/audit/report?format=md'

  return (
    <div data-testid="audit-report-card" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
        <span className="mono" style={{ fontSize: 'var(--fs-body)', color: 'var(--text-primary)' }}>{report.site || '—'}</span>
        <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-muted)' }}>
          {report.link_type || '—'}{report.provider ? ` · ${report.provider}` : ''}
        </span>
        <span style={{ marginLeft: 'auto' }}><Badge tone={proofTone(proof)}>{PROOF_LABEL[proof]}</Badge></span>
      </div>
      <Muted>{directionLabel(m.direction)}{report.timestamp ? ` · ${report.timestamp}` : ''}{report.bloat_grade ? ` · note ${report.bloat_grade}` : ' · non noté'}</Muted>

      {report.phases && (report.phases.baseline_seconds != null || report.phases.load_seconds != null || report.phases.recovery_seconds != null) ? (
        <Section title="Fenêtre">
          <div data-testid="audit-phases" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 'var(--sp-2)' }}>
            {report.phases.baseline_seconds != null ? <Kpi label="repos" value={String(report.phases.baseline_seconds)} unit="s" /> : null}
            {report.phases.load_seconds != null ? <Kpi label="charge" value={String(report.phases.load_seconds)} unit="s" /> : null}
            {report.phases.recovery_seconds != null ? <Kpi label="récupération" value={String(report.phases.recovery_seconds)} unit="s" /> : null}
          </div>
        </Section>
      ) : null}

      <Section title="Constat">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 'var(--sp-2)' }}>
          <Kpi label="p95 repos" value={ms(m.rtt_idle_p95_ms)} unit="ms" />
          <Kpi label="p95 chargé" value={load ? ms(m.rtt_loaded_p95_ms) : '—'} unit={load ? 'ms' : ''} />
          <Kpi label="Δ p95" value={load ? ms(m.bloat_delta_ms) : '—'} unit={load ? 'ms' : ''} />
          <Kpi label="drain" value={m.drain_measured ? Number(m.drain_ms ?? 0).toFixed(1) : '—'} unit={m.drain_measured ? 'ms' : ''} />
        </div>
        {report.bloat_verdict ? <Muted>{report.bloat_verdict}</Muted> : null}
        {!m.drain_measured ? <Muted>drain non mesuré — un audit court n'a pas de fenêtre de récupération ; ce n'est pas zéro.</Muted> : null}
      </Section>

      <Section title="Charge">
        {load ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 'var(--sp-2)' }}>
            <Kpi label="débit" value={ms(m.throughput_mbps)} unit="Mbit/s" />
            <Kpi label="perte" value={m.loss_pct == null ? '—' : Number(m.loss_pct).toFixed(2)} unit="%" />
            <Kpi label="flux" value={m.flow_count == null ? '—' : String(m.flow_count)} />
            {m.flow_count != null && m.flow_count > 1 ? <Kpi label="JFI" value={m.jfi_pct == null ? '—' : Number(m.jfi_pct).toFixed(1)} unit="%" /> : null}
            {m.small_measured ? <Kpi label="petit p95" value={ms(m.http_small_p95_ms)} unit="ms" /> : null}
            {m.voip_r ? <Kpi label="VoIP R" value={m.voip_r.toFixed(1)} /> : null}
          </div>
        ) : (
          <Muted>charge montante non mesurée — les absences sont listées plus bas.</Muted>
        )}

        {down ? (
          <div style={{ borderTop: '1px solid var(--hairline-faint)', paddingTop: 'var(--sp-2)', display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }} data-testid="audit-report-down">
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-2)' }}>
              <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-muted)' }}>miroirs descendants</span>
              <Badge tone={gradeTone(m.grade_down)}>{m.grade_down || 'note down —'}</Badge>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 'var(--sp-2)' }}>
              <Kpi label="débit down" value={ms(m.throughput_down_mbps)} unit="Mbit/s" />
              <Kpi label="p95 down" value={ms(m.rtt_loaded_down_p95_ms)} unit="ms" />
              <Kpi label="Δ down" value={ms(m.bloat_delta_down_ms)} unit="ms" />
              {m.flow_count != null && m.flow_count > 1 ? <Kpi label="JFI down" value={m.jfi_down_pct == null ? '—' : Number(m.jfi_down_pct).toFixed(1)} unit="%" /> : null}
            </div>
          </div>
        ) : null}
      </Section>

      <Section title="SLA CAP">
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-2)', flexWrap: 'wrap' }}>
          <Badge tone={capTone(cap)}>{!cap?.measured ? 'non mesuré' : cap.breach ? 'manquement SLA' : 'conforme'}</Badge>
          {cap?.measured ? <Kpi label="échéances" value={Number(cap.ok_pct ?? 0).toFixed(1)} unit="%" /> : null}
        </div>
        <Muted>objectif opérateur DGM 220 ms (cadre WIS 2.0) · complétions 16 KiB chargées</Muted>
        {cap?.measured && cap.breach && cap.label ? <Muted>{cap.label}</Muted> : null}
        {!cap?.measured ? <Muted>aucune complétion 16 KiB chargée — jamais un succès par défaut.</Muted> : null}
      </Section>

      <Section title="Économie">
        {!econ ? (
          <Muted>non mesurée — aucun scénario n'est chiffré plutôt qu'un zéro fabriqué.</Muted>
        ) : (
          <>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
              <div className="c-kpi">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-1)' }}>
                  <Badge tone="steel">scénario contractuel</Badge>
                  <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>primaire</span>
                </div>
                <Stat label={primary?.name ?? '—'} value={formatCostAR(primary?.ar_per_hour).num} unit={formatCostAR(primary?.ar_per_hour).unit} />
              </div>
              {b2b ? (
                <div className="c-kpi">
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--sp-1)' }}>
                    <Badge tone="steel">B2B</Badge>
                    <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>au volume</span>
                  </div>
                  <Stat label={b2b.name} value={formatCostAR(b2b.ar_per_hour).num} unit={formatCostAR(b2b.ar_per_hour).unit} />
                </div>
              ) : null}
            </div>
            <Muted>
              réémission {econ.retrans_ratio?.ratio != null ? Number(econ.retrans_ratio.ratio).toFixed(4) : '—'} ({ratioBasis(econ.retrans_ratio?.basis)}) · {bytes(econ.wasted_bytes)} sur {bytes(econ.transported_bytes)} en {Number(econ.load_seconds ?? 0).toFixed(0)} s de charge
            </Muted>
            {scenarios.length > 0 ? (
              <details>
                <summary className="mono" style={{ cursor: 'pointer', fontSize: 'var(--fs-micro)', color: 'var(--text-muted)' }}>
                  {scenarios.length} scénarios contractuels — voir le détail
                </summary>
                <div className="c-tbl-wrap" style={{ marginTop: 'var(--sp-1)' }}>
                  <table className="c-tbl">
                    <thead><tr><th>scénario</th><th className="r">Ar/h</th></tr></thead>
                    <tbody>
                      {scenarios.map(s => (
                        <tr key={s.name}>
                          <td>
                            {s.name}{s.name === primary?.name ? ' · primaire' : ''}
                            <br /><span style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>{s.basis}</span>
                          </td>
                          <td className="r">{groupDigits(s.ar_per_hour, 2)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </details>
            ) : null}
            <Muted>{econ.disclaimer || 'scénarios contractuels, pas des mesures.'}</Muted>
          </>
        )}
      </Section>

      <Section title="Preuves">
        <Provenance
          source={report.provenance?.source_file || 'source absente'}
          refresh="au gel"
          state="live"
          n={report.provenance?.row_count}
          extra={report.provenance?.source_sha256 ? `sha256 ${report.provenance.source_sha256.slice(0, 12)}…` : undefined}
        />
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--sp-1)' }}>
          {report.provenance?.source_sha256 ? (
            <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }} title={report.provenance.source_sha256}
              onClick={() => onCopy(report.provenance?.source_sha256 ?? '', 'Sceau source')}>
              source <code>{report.provenance.source_sha256.slice(0, 12)}…</code>
            </button>
          ) : null}
          {report.provenance?.telemetry_sha256 ? (
            <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }} title={report.provenance.telemetry_sha256}
              onClick={() => onCopy(report.provenance?.telemetry_sha256 ?? '', 'Sceau télémétrie')}>
              télémétrie <code>{report.provenance.telemetry_sha256.slice(0, 12)}…</code>
            </button>
          ) : null}
        </div>
        {!report.provenance?.source_sha256 ? <p style={{ color: 'var(--text-body)' }}>Empreinte source absente</p> : null}
        {report.provenance?.telemetry_file ? <Muted>série brute : {report.provenance.telemetry_file}</Muted> : null}
        {!report.provenance?.telemetry_sha256 ? (
          <p style={{ color: 'var(--text-body)' }}>{report.provenance?.telemetry_file ? 'Empreinte télémétrie absente' : 'Télémétrie absente'}</p>
        ) : null}
        <div style={{ display: 'flex', gap: 'var(--sp-2)', flexWrap: 'wrap', alignItems: 'baseline' }}>
          <a className="mono" href={mdHref} target="_blank" rel="noreferrer" style={{ fontSize: 'var(--fs-micro)', color: 'var(--t-live)' }}>rapport Markdown ↗</a>
          <button className="btn" disabled={proof === 'missing'} title={proof === 'missing' ? 'Citation impossible — empreinte source absente' : undefined} style={{ padding: 'var(--sp-1) var(--sp-2)' }}
            onClick={() => onCopy(auditCitation(report), 'Citation du rapport')}>CITER CE RAPPORT</button>
          {m.data_used_mb ? <Muted>données consommées par l'audit : {Number(m.data_used_mb).toFixed(2)} Mo</Muted> : null}
        </div>
      </Section>

      <Section title="Remédiation">
        {report.remediation ? (
          <>
            <Muted>{report.remediation.qdisc}{report.remediation.reason ? ` — ${report.remediation.reason}` : ''}</Muted>
            <div style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
              <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }}
                onClick={() => onCopy((report.remediation?.linux_tc ?? []).join('\n'), 'Script tc')}>COPIER LINUX TC</button>
              {report.remediation.routeros ? (
                <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }}
                  onClick={() => onCopy(report.remediation?.routeros ?? '', 'Script RouterOS')}>COPIER ROUTEROS</button>
              ) : null}
            </div>
            {(report.remediation.linux_tc_ingress?.length || report.remediation.routeros_ingress) ? (
              <div data-testid="audit-remediation-ingress" style={{ borderTop: '1px solid var(--hairline-faint)', paddingTop: 'var(--sp-2)', display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
                <Muted>téléchargement (ingress IFB)</Muted>
                <div style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
                  {report.remediation.linux_tc_ingress?.length ? (
                    <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }}
                      onClick={() => onCopy((report.remediation?.linux_tc_ingress ?? []).join('\n'), 'Script tc ingress')}>COPIER LINUX TC (INGRESS)</button>
                  ) : null}
                  {report.remediation.routeros_ingress ? (
                    <button className="btn" style={{ padding: 'var(--sp-1) var(--sp-2)' }}
                      onClick={() => onCopy(report.remediation?.routeros_ingress ?? '', 'Script RouterOS ingress')}>COPIER ROUTEROS (INGRESS)</button>
                  ) : null}
                </div>
              </div>
            ) : null}
            {report.remediation.routeros || report.remediation.routeros_ingress ? (
              <p className="mono" data-testid="ros-validation" style={{ fontSize: 'var(--fs-tag)', color: 'var(--text-muted)', marginTop: 4 }}>
                {(() => {
                  const v = report.routeros_validation
                  if (!v) return 'Configuration RouterOS non validée — aucune preuve CHR'
                  const checks = v.checks ?? []
                  // Zéro vérification n'est pas une validation : un tableau
                  // vide passait pour un succès, c'était un mensonge silencieux.
                  const ok = checks.length > 0 && checks.every(c => c.pass)
                  const level = v.level === 'linux-namespace' ? 'CHR + banc' : 'CHR'
                  const h8 = (v.export_sha256 ?? '').slice(0, 8)
                  if (checks.length === 0) return 'Configuration RouterOS non validée — aucune vérification'
                  return ok
                    ? `Configuration RouterOS validée (${v.version ?? '7.x'}, ${level}) · export ${h8}`
                    : 'Configuration RouterOS non validée — écarts dans le rapport'
                })()}
              </p>
            ) : null}
            {report.kernel_telemetry ? <KernelRing k={report.kernel_telemetry} /> : null}
          </>
        ) : (
          <Muted>non émise — capacité ou RTT repos non mesurés ; une file dimensionnée au hasard serait déployée à l'aveugle.</Muted>
        )}
      </Section>

      {(report.unavailable?.length ?? 0) > 0 ? (
        <Section title="Non mesuré">
          {report.unavailable?.map(u => (
            <div key={`${u.metric}-${u.reason}`} className="mono" style={{ fontSize: 'var(--fs-micro)', lineHeight: 1.5, color: 'var(--text-faint)' }}>
              <b style={{ color: 'var(--text-muted)' }}>{u.metric}</b> — {u.reason}
            </div>
          ))}
        </Section>
      ) : null}
    </div>
  )
}
