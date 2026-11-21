// Le tiroir de preuve : deux onglets, la commande exacte d'un côté, les
// empreintes de l'autre. Les fixtures portent les champs que le moteur scelle.
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ProofSheet, replayLine } from './ProofDrawer'
import { useAuditStore } from '../store/audit'
import { useUIStore, type ProofTab } from '../store/ui'
import type { AuditReport } from '../lib/types'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const report: AuditReport = {
  audit_id: 'audit-2026-11-02',
  site: 'Département X',
  link_type: '5g',
  measured: {
    rtt_idle_p50_ms: 28,
    rtt_idle_p95_ms: 31,
    rtt_loaded_p50_ms: 44,
    rtt_loaded_p95_ms: 61,
    bloat_delta_ms: 30,
    qdi_ms: 9,
    loss_pct: 0.4,
    throughput_mbps: 18.2,
    flow_count: 4,
    direction: 'both',
    load_measured: true,
  },
  provenance: {
    source_file: 'data/audits/audit-2026-11-02/report.json',
    source_sha256: 'a'.repeat(64),
    telemetry_file: 'data/audits/audit-2026-11-02/telemetry.csv',
    telemetry_sha256: 'b'.repeat(64),
  },
  kernel_telemetry: { available: true, source: 'ebpf' },
  cap_sla: { measured: true, ok_pct: 96.7, target_pct: 95, deadline_ms: 220 },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: 'Sous charge, la latence reste sous l’objectif.' },
  remediation: {
    qdisc: 'fq_codel',
    up_measured_mbps: 5.06,
    up_shaped_mbps: 4.81,
    down_measured_mbps: 18.5,
    down_shaped_mbps: 17.58,
    routeros: '/queue type add name=cgo-shape',
    routeros_ingress: '/queue tree add name=cgoingress-lan-in parent=bridge queue=cgoingress-cake',
    routeros_lan_iface: 'bridge',
    routeros_lan_assumed: true,
    linux_tc: ['tc qdisc add dev eth0 root handle 1: htb'],
  },
  routeros_validation: { checks: [{ name: 'queue type', expected: 'cgo-shape', found: 'cgo-shape', pass: true }] },
}

function render(tab: ProofTab, r: AuditReport | null = report): string {
  useUIStore.setState({ proof: { tab } })
  useAuditStore.setState({ report: r })
  return renderToStaticMarkup(<ProofSheet />)
}

const text = (html: string) => html.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ')

beforeEach(() => { localStorage.clear() })

describe('ProofDrawer', () => {
  it('l’onglet preuve est sélectionné et porte l’empreinte source', () => {
    const html = render('proof')
    expect(html).toMatch(/id="proof-tab-proof"[^>]*aria-selected="true"/)
    expect(html).toContain(report.provenance!.source_sha256)
    expect(html).toContain(report.provenance!.telemetry_sha256)
    const out = text(html)
    expect(out).toContain('RTT repos p50 / p95')
    expect(out).toContain('les deux sens (both)')
    expect(out).toContain('eBPF — compteurs noyau')
    expect(html).toContain('href="/api/audit/report?audit_id=audit-2026-11-02"')
  })

  it('l’onglet déploiement porte la base, les blocs et la vérification', () => {
    const html = render('commands')
    expect(html).toMatch(/id="proof-tab-commands"[^>]*aria-selected="true"/)
    const out = text(html)
    expect(out).toContain('envoi 5,1 Mbit/s × 0,95 → 4,8 Mbit/s')
    expect(out).toContain('Déploiement — envoi')
    expect(out).toContain('Déploiement — réception (LAN)')
    expect(out).toContain('/queue type add name=cgo-shape')
    expect(out).toContain('/queue tree add name=cgoingress-lan-in parent=bridge')
    expect(out).toContain('Vérification — après déploiement')
    expect(out).toContain('/queue type print')
    expect((html.match(/role="radio"/g) || []).length).toBe(2)
    expect(out).toContain('interface supposée « bridge » — à vérifier avant de coller')
    expect((html.match(/>Copier</g) || []).length).toBe(3)
  })

  it('un rapport montant seul n’a pas de bloc de réception', () => {
    const uploadOnly: AuditReport = {
      ...report,
      remediation: {
        qdisc: 'fq_codel',
        up_measured_mbps: 5.06,
        up_shaped_mbps: 4.81,
        routeros: '/queue type add name=cgo-shape',
      },
    }
    const out = text(render('commands', uploadOnly))
    expect(out).toContain('Déploiement — envoi')
    expect(out).not.toContain('Déploiement — réception')
    expect(out).toContain('× 0,95 → 4,8 Mbit/s')
  })

  it('sans remédiation, l’onglet déploiement le dit', () => {
    const out = text(render('commands', { ...report, remediation: undefined }))
    expect(out).toContain('Aucune remédiation')
  })

  it('la ligne de rejeu ne dit jamais « appliquée »', () => {
    expect(replayLine(undefined)).toContain('aucune preuve')
    expect(replayLine({ checks: [] })).toContain('non validée — aucune vérification')
    expect(replayLine({ checks: [{ name: 'x', expected: 'a', found: 'b', pass: false }] })).toContain('non validée — écarts dans le rapport')
    expect(replayLine({ checks: [{ name: 'x', expected: 'a', found: 'a', pass: true }] })).toContain('validée par rejeu')
    const out = text(render('commands', { ...report, routeros_validation: undefined }))
    expect(out).toContain('Rejeu RouterOS : aucune preuve')
    expect(out).not.toContain('appliquée')
  })

  it('nomme la sonde et sa raison quand la télémétrie noyau manque', () => {
    const out = text(render('proof', {
      ...report,
      kernel_telemetry: { available: false, source: 'tcpinfo', reason: 'objet eBPF absent' },
    }))
    expect(out).toContain('TCP_INFO — amont du noyau : objet eBPF absent')
  })

  it('sans rapport, le tiroir ne fabrique aucun chiffre', () => {
    const out = text(render('proof', null))
    expect(out).toContain('Aucun rapport scellé')
    expect(out).not.toContain('Pas encore mesuré')
  })
})

describe('ProofDrawer — copie honnête', () => {
  const hadOwnClipboard = Object.prototype.hasOwnProperty.call(navigator, 'clipboard')
  const ownClipboard = Object.getOwnPropertyDescriptor(navigator, 'clipboard')

  beforeEach(() => {
    vi.useFakeTimers()
    useUIStore.setState({ toasts: [] })
  })
  afterEach(() => {
    vi.clearAllTimers()
    useUIStore.setState({ toasts: [] })
    vi.useRealTimers()
    if (hadOwnClipboard) Object.defineProperty(navigator, 'clipboard', ownClipboard!)
    else delete (navigator as unknown as { clipboard?: unknown }).clipboard
  })

  it('quand le presse-papiers est absent, « Copier » dit « Copie impossible »', () => {
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: undefined })
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    act(() => {
      root.render(<ProofSheet tab="commands" report={report} onClose={() => {}} onTab={() => {}} />)
    })
    const copy = host.querySelector<HTMLButtonElement>('.proof-copy')!
    act(() => { copy.click() })
    expect(
      useUIStore.getState().toasts.map(t => t.msg).join(' '),
      'un mot remplace le no-op silencieux',
    ).toContain('Copie impossible')
    act(() => root.unmount())
    host.remove()
  })
})
