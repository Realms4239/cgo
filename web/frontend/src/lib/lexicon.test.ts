// Porte lexicon — les quatre surfaces par défaut parlent opérateur.
// Ce qu'il prouve : aucun mot banni n'apparaît chez l'opérateur qui ouvre
// l'outil sans rien toucher. Ce qu'il ne prouve pas : les tiroirs et le mode
// recherche, qui gardent leur règle dans DESIGN.md §5.
import { describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { createElement } from 'react'
import DashboardView from '../views/DashboardView'
import AuditView from '../views/AuditView'
import BenchView from '../views/BenchView'
import EvidenceView from '../views/EvidenceView'

const REPORT = vi.hoisted(() => ({
  audit_id: 'audit-lexicon',
  site: 'site-nord',
  link_type: '4g',
  provider: 'yas',
  timestamp: '2026-11-02T10:00:00Z',
  bloat_grade: 'B',
  bloat_verdict: 'Le lien tient sous charge.',
  operator_verdict: { level: 'green', word: 'Lien stable', sentence: 'La mesure tient sous la charge déclarée.' },
  measured: {
    direction: 'down',
    load_measured: true,
    load_attempted: true,
    throughput_mbps: 18.2,
    http_small_p95_ms: 180,
    small_measured: true,
    rtt_idle_p50_ms: 30,
    rtt_idle_p95_ms: 40,
    rtt_loaded_p50_ms: 48,
    rtt_loaded_p95_ms: 63,
  },
  economics: {
    disclaimer: 'Scénarios contractuels, pas des mesures.',
    primary: { model: 'A', name: 'ligne dédiée', ar_per_hour: 1388.89 },
    scenarios: [{ model: 'emergency', name: 'hors forfait', ar_per_hour: 5556 }],
  },
  cap_sla: { target_pct: 95, deadline_ms: 220, breach: false },
  remediation: {
    qdisc: 'cake',
    cc: 'bbr',
    up_measured_mbps: 20,
    up_shaped_mbps: 19,
    down_measured_mbps: 100,
    down_shaped_mbps: 95,
    linux_tc: ['tc qdisc add dev eth0 root cake'],
    routeros: '/queue type add name=cake',
  },
  provenance: { source_file: 'audit_report.json', row_count: 1, source_sha256: 'b'.repeat(64) },
} as any))

vi.mock('../store/audit', () => {
  const state = {
    stage: 'done',
    report: REPORT,
    previous: null,
    comparison: null,
    msg: '',
    historyN: null,
    live: null,
    direction: 'down',
    restart: () => undefined,
    showReport: () => undefined,
    open: () => undefined,
    stop: () => undefined,
  }
  return {
    useAuditStore: (selector?: (s: any) => any) => (selector ? selector(state) : state),
  }
})

const SURFACES: { name: string; render: () => string }[] = [
  { name: 'tableau de bord', render: () => renderToStaticMarkup(createElement(DashboardView)) },
  { name: 'auditer', render: () => renderToStaticMarkup(createElement(AuditView)) },
  { name: 'banc', render: () => renderToStaticMarkup(createElement(BenchView)) },
  { name: 'preuves', render: () => renderToStaticMarkup(createElement(EvidenceView)) },
]

const BANNED: { label: string; pattern: RegExp }[] = [
  { label: 'p50', pattern: /p50/i },
  { label: 'p95', pattern: /p95/i },
  { label: 'qdisc', pattern: /qdisc/i },
  { label: 'CAKE', pattern: /cake/i },
  { label: 'fq_codel', pattern: /fq[_ ]?codel/i },
  { label: 'BBR', pattern: /bbr/i },
  { label: 'CUBIC', pattern: /cubic/i },
  { label: 'JFI', pattern: /\bjfi\b/i },
  { label: 'QDI', pattern: /\bqdi\b/i },
  { label: 'SLA', pattern: /\bsla\b/i },
  { label: 'bufferbloat', pattern: /bufferbloat/i },
  { label: 'eBPF', pattern: /ebpf/i },
  { label: 'manifeste', pattern: /manifeste/i },
  { label: 'corpus', pattern: /corpus/i },
  { label: 'gel', pattern: /gel/i },
  { label: 'twin', pattern: /\btwin\b/i },
  { label: 'run', pattern: /\bruns?\b/i },
  { label: 'n=', pattern: /n=/i },
  { label: 'empreinte hexadécimale', pattern: /\b[0-9a-f]{12,}\b/i },
  { label: 'garanti', pattern: /garanti/i },
  { label: 'prouvé', pattern: /prouvé/i },
]

export function scanOperatorText(text: string): string[] {
  const found: string[] = []
  for (const { label, pattern } of BANNED) {
    const match = pattern.exec(text)
    if (match) {
      const start = Math.max(0, match.index - 24)
      const end = Math.min(text.length, match.index + label.length + 24)
      found.push(`${label} — « …${text.slice(start, end).trim()}… »`)
    }
  }
  return found
}

export function visibleText(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<[^>]*>/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&nbsp;/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

describe('lexicon opérateur', () => {
  for (const surface of SURFACES) {
    it(`${surface.name} ne parle pas le lexique de laboratoire`, () => {
      const offenders = scanOperatorText(visibleText(surface.render()))
      expect(offenders, offenders.join('\n')).toEqual([])
    })
  }

  it('le contrôle détecte bien un texte interdit', () => {
    expect(scanOperatorText('le corpus gelé').length).toBeGreaterThan(0)
  })
})