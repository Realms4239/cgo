// Le tiroir de preuve — deux lectures d'un même rapport scellé : ce qu'on
// déploie (commande exacte, base de dimensionnement, vérification) et ce qui le
// prouve (chiffres, sonde, empreintes). Aucun chiffre n'est recalculé ici.
import { useRef, useState, type KeyboardEvent } from 'react'
import { createPortal } from 'react-dom'
import { operatorFacts } from '../lib/dashboard'
import { directionLabel } from '../lib/auditFlow'
import { numFr } from '../lib/format'
import { loadSettings } from '../lib/settings'
import { useModalFocus } from '../lib/modalFocus'
import type { AuditRemediation, AuditReport, KernelTelemetry, RouterOSValidation } from '../lib/types'
import { useAuditStore } from '../store/audit'
import { useUIStore, type ProofTab } from '../store/ui'
import { PLATFORM_LABEL, blocksOf, carriedPlatforms, readPlatform, savePlatform, type Platform } from './RemediationCard'

const ms = (v?: number) => (v != null && Number.isFinite(v) ? `${numFr(v, 0)} ms` : 'Pas encore mesuré')

// Une absence ne pèse pas comme une mesure (même règle que le verdict).
const IS_ABSENT_VALUE = /Pas encore mesuré|Non chiffré/

// La base de dimensionnement, telle que le moteur l'a appliquée : le facteur
// n'est affiché que si le rapport porte le taux façonné.
function sizingBasis(r: AuditRemediation, rttIdleP95?: number): string {
  const parts: string[] = []
  for (const [label, measured, shaped] of [
    ['envoi', r.up_measured_mbps, r.up_shaped_mbps],
    ['réception', r.down_measured_mbps, r.down_shaped_mbps],
  ] as const) {
    if (measured && measured > 0) {
      parts.push(shaped && shaped > 0
        ? `${label} ${numFr(measured, 1)} Mbit/s × 0,95 → ${numFr(shaped, 1)} Mbit/s`
        : `${label} ${numFr(measured, 1)} Mbit/s`)
    }
  }
  if (rttIdleP95 && rttIdleP95 > 0) parts.push(`RTT repos p95 ${numFr(rttIdleP95, 0)} ms`)
  if (r.qdisc) parts.push(`qdisc ${r.qdisc}`)
  return parts.length ? `Base : ${parts.join(' · ')}.` : 'Aucune mesure de capacité — aucun taux à façonner.'
}

// Ce que la validation a prouvé, jamais ce qui a été appliqué.
type ReplayLevel = 'neutral' | 'amber' | 'red' | 'green'

export function replayState(v?: RouterOSValidation): { level: ReplayLevel; text: string } {
  if (!v) return { level: 'neutral', text: 'Rejeu RouterOS : aucune preuve' }
  if (!v.checks?.length) return { level: 'amber', text: 'Rejeu RouterOS : non validée — aucune vérification' }
  if (v.checks.some(c => !c.pass)) return { level: 'red', text: 'Rejeu RouterOS : non validée — écarts dans le rapport' }
  return { level: 'green', text: 'Rejeu RouterOS : validée par rejeu' }
}

export function replayLine(v?: RouterOSValidation): string {
  return replayState(v).text
}

function sensorLine(k?: KernelTelemetry): string {
  if (!k) return 'standard — non instrumentée'
  if (k.source === 'ebpf') return 'eBPF — compteurs noyau'
  if (k.source === 'tcpinfo') return `TCP_INFO — amont du noyau : ${k.reason || 'raison non consignée'}`
  return `standard — ${k.reason || 'sonde non instrumentée'}`
}

function copyText(label: string, text: string) {
  navigator.clipboard?.writeText(text).then(
    () => useUIStore.getState().pushToast(`${label} copié`, 'ok'),
    () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
  )
}

function CommandsTab({ report }: { report: AuditReport }) {
  const r = report.remediation
  const carried = carriedPlatforms(r)
  const [chosen, setChosen] = useState<Platform | null>(() => readPlatform())
  if (!r || carried.length === 0) {
    return <p className="mono muted">Aucune remédiation : la capacité ou le RTT repos n'a pas été mesuré.</p>
  }
  const platform = chosen && carried.includes(chosen) ? chosen : carried[0]
  const blocks = blocksOf(r, platform)
  const assumed = platform === 'routeros' && !!r.routeros_ingress && !!r.routeros_lan_assumed
  const verification = platform === 'routeros' ? '/queue type print\n/queue tree print' : 'tc qdisc show'
  const replay = replayState(report.routeros_validation)
  const pick = (p: Platform) => {
    setChosen(p)
    savePlatform(p)
  }
  return (
    <>
      <p className="proof-basis mono">{sizingBasis(r, report.measured?.rtt_idle_p95_ms)}</p>
      {carried.length > 1 ? (
        <div className="remediation-platform" role="radiogroup" aria-label="Plateforme">
          {carried.map(p => (
            <button
              key={p}
              type="button"
              role="radio"
              aria-checked={platform === p}
              className="remediation-choice"
              onClick={() => pick(p)}
            >
              {PLATFORM_LABEL[p]}
            </button>
          ))}
        </div>
      ) : (
        <p className="proof-platform mono">Plateforme : {PLATFORM_LABEL[platform]}</p>
      )}
      {assumed ? (
        <p className="remediation-assumed">
          interface supposée « {r.routeros_lan_iface || 'bridge'} » — à vérifier avant de coller
        </p>
      ) : null}
      {blocks.map(b => (
        <section className="proof-block" key={b.label}>
          <div className="proof-block-head">
            <span className="proof-block-label mono">{b.label}</span>
            <button type="button" className="btn proof-copy" onClick={() => copyText(b.label, b.text)}>Copier</button>
          </div>
          <pre className="proof-pre mono">{b.text}</pre>
        </section>
      ))}
      <section className="proof-block">
        <div className="proof-block-head">
          <span className="proof-block-label mono">Vérification — après déploiement</span>
          <button type="button" className="btn proof-copy" onClick={() => copyText('Vérification', verification)}>Copier</button>
        </div>
        <pre className="proof-pre mono">{verification}</pre>
      </section>
      <p className="proof-replay mono" data-level={replay.level}>
        <span className="proof-replay-dot" aria-hidden="true" />{replay.text}
      </p>
    </>
  )
}

function ProofTab({ report }: { report: AuditReport }) {
  const settings = loadSettings()
  const facts = operatorFacts(report, settings.warnMs, settings.critMs)
  const m = report.measured
  const p = report.provenance
  const hasMissingSource = !p?.source_sha256
  const rows: [string, string][] = [
    ['RTT repos p50 / p95', `${ms(m?.rtt_idle_p50_ms)} / ${ms(m?.rtt_idle_p95_ms)}`],
    ['RTT chargé p50 / p95', `${ms(m?.rtt_loaded_p50_ms)} / ${ms(m?.rtt_loaded_p95_ms)}`],
    ['Δ p95 chargé − repos', ms(m?.bloat_delta_ms)],
    ['Instabilité (QDI)', ms(m?.qdi_ms)],
    ['Perte', m?.loss_pct != null ? `${numFr(m.loss_pct, 2)} %` : 'Pas encore mesuré'],
    ['Débit', m?.throughput_mbps != null ? `${numFr(m.throughput_mbps, 1)} Mbit/s` : 'Pas encore mesuré'],
    ['Flux mesurés (n)', m?.flow_count != null ? String(m.flow_count) : '—'],
    ['Sens', directionLabel(m?.direction)],
  ]
  return (
    <>
      {hasMissingSource ? <p role="status">Preuve manquante — aucune empreinte source : le rapport peut être lu, pas cité.</p> : null}
      <dl className="proof-facts">
        {facts.map(f => {
          const detail = [f.target, f.band?.word].filter(Boolean).join(' · ')
          return (
            <div className="proof-fact" key={f.key} data-empty={IS_ABSENT_VALUE.test(f.value) || undefined}>
              <dt>{f.label}</dt>
              <dd>
                <b>{f.value}</b>{f.unit ? <small> {f.unit}</small> : null}
                {detail ? <em>{detail}</em> : null}
              </dd>
            </div>
          )
        })}
      </dl>
      <dl className="proof-rows">
        {rows.map(([label, value]) => (
          <div className="proof-row" key={label}>
            <dt>{label}</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
      <p className="proof-sensor mono">Sonde : {sensorLine(report.kernel_telemetry)}</p>
      <dl className="proof-hashes">
        <div className="proof-row">
          <dt>Source</dt>
          <dd>{p?.source_file || 'fichier inconnu'}</dd>
        </div>
        <div className="proof-row">
          <dt>SHA-256 source</dt>
          <dd><code>{p?.source_sha256 || 'aucune empreinte'}</code></dd>
        </div>
        <div className="proof-row">
          <dt>Télémétrie</dt>
          <dd>{p?.telemetry_file || 'aucun fichier'}</dd>
        </div>
        <div className="proof-row">
          <dt>SHA-256 télémétrie</dt>
          <dd><code>{p?.telemetry_sha256 || 'aucune empreinte'}</code></dd>
        </div>
      </dl>
      <a className="proof-link mono" href={`/api/audit/report?audit_id=${encodeURIComponent(report.audit_id)}`} target="_blank" rel="noreferrer">
        Rapport scellé — JSON
      </a>
    </>
  )
}

// Le contenu seul, sans portail : c'est lui que le test DOM rend. Sans props,
// il lit l'état courant du store ; le portail lui passe les valeurs vivantes.
export interface ProofSheetProps {
  tab?: ProofTab
  report?: AuditReport | null
  onClose?: () => void
  onTab?: (tab: ProofTab) => void
}

export function ProofSheet({ tab: tabProp, report: reportProp, onClose, onTab }: ProofSheetProps = {}) {
  const tab = tabProp ?? useUIStore.getState().proof?.tab ?? 'commands'
  const report = reportProp !== undefined ? reportProp : useAuditStore.getState().report
  const closeProof = onClose ?? useUIStore.getState().closeProof
  const openProof = onTab ?? useUIStore.getState().openProof
  const panelRef = useRef<HTMLDivElement>(null)
  useModalFocus(true, closeProof, panelRef)
  // Un vrai tablist se parcourt aux flèches, pas seulement à la tabulation.
  // Deux onglets : chaque flèche mène à l'autre, comme un cycle, en partant
  // de l'onglet qui a le focus.
  const onTabKey = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return
    e.preventDefault()
    const focused: ProofTab =
      document.activeElement?.id === 'proof-tab-commands' ? 'commands'
        : document.activeElement?.id === 'proof-tab-proof' ? 'proof'
          : tab
    const next: ProofTab = focused === 'commands' ? 'proof' : 'commands'
    openProof(next)
    requestAnimationFrame(() => document.getElementById(`proof-tab-${next}`)?.focus())
  }
  return (
    <div role="dialog" aria-modal="true" aria-label="Preuves de l'audit" className="proof-drawer" onClick={closeProof}>
      <div ref={panelRef} className="proof-drawer-panel" onClick={e => e.stopPropagation()}>
        <div className="proof-head">
          <div>
            <h2 className="proof-title">Preuves de l'audit</h2>
            <span className="proof-id mono">{report ? [report.site || 'site —', report.timestamp?.slice(0, 16).replace('T', ' ')].filter(Boolean).join(' · ') : 'aucun rapport'}</span>
          </div>
          <button type="button" className="proof-close" aria-label="Fermer" onClick={closeProof}>✕</button>
        </div>
        <div role="tablist" aria-label="Lectures du rapport" className="proof-tabs" onKeyDown={onTabKey}>
          <button
            type="button" role="tab" id="proof-tab-commands"
            aria-selected={tab === 'commands'} aria-controls="proof-panel"
            className="proof-tab" onClick={() => openProof('commands')}
          >
            Déploiement & scripts
          </button>
          <button
            type="button" role="tab" id="proof-tab-proof"
            aria-selected={tab === 'proof'} aria-controls="proof-panel"
            className="proof-tab" onClick={() => openProof('proof')}
          >
            Preuves & télémétrie brute
          </button>
        </div>
        <div role="tabpanel" id="proof-panel" aria-labelledby={`proof-tab-${tab}`} className="proof-panel">
          {report
            ? (tab === 'commands' ? <CommandsTab report={report} /> : <ProofTab report={report} />)
            : <p className="mono muted">Aucun rapport scellé — lancez un audit.</p>}
        </div>
      </div>
    </div>
  )
}

export default function ProofDrawer() {
  const proof = useUIStore(s => s.proof)
  const closeProof = useUIStore(s => s.closeProof)
  const openProof = useUIStore(s => s.openProof)
  const report = useAuditStore(s => s.report)
  return proof
    ? createPortal(
        <ProofSheet tab={proof.tab} report={report} onClose={closeProof} onTab={openProof} />,
        document.body,
      )
    : null
}
