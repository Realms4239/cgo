// AuditReportCard — l'étape « Lire » : le rapport complet, sans cacher une
// absence. L'économie vient de `economics.*` (le P0 ne lit plus `waste.*`) et
// chaque entrée `unavailable` est rendue telle quelle.
import { Badge } from './ui/StatusPip'
import { CardHead } from './ui/CardHead'
import { Stat } from './ui/Stat'
import { directionLabel } from '../lib/auditFlow'
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
        <span style={{ marginLeft: 'auto' }}><Badge tone={gradeTone(report.bloat_grade)}>{report.bloat_grade || 'non mesuré'}</Badge></span>
      </div>
      <Muted>{directionLabel(m.direction)}{report.timestamp ? ` · ${report.timestamp}` : ''}</Muted>

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
        {report.provenance?.source_file ? <Muted>source : {report.provenance.source_file} · {report.provenance.row_count ?? 0} lignes</Muted> : null}
        {report.provenance?.telemetry_file ? <Muted>série brute : {report.provenance.telemetry_file}</Muted> : null}
        <div style={{ display: 'flex', gap: 'var(--sp-2)', flexWrap: 'wrap', alignItems: 'baseline' }}>
          <a className="mono" href={mdHref} target="_blank" rel="noreferrer" style={{ fontSize: 'var(--fs-micro)', color: 'var(--t-live)' }}>rapport Markdown ↗</a>
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
