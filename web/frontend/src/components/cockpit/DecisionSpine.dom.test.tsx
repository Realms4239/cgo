// Le spine porte le jugement du moteur, jamais un jugement inventé par l'interface.
// Ces assertions lisent le DOM rendu : un seul mot, une seule phrase, des faits
// qui disent leur absence, et quatre états d'action exclusifs.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SynthèseView from '../../views/SynthèseView'
import AnswerBand from './AnswerBand'
import DecisionSpine, { selectActState } from './DecisionSpine'
import type { AuditReport } from '../../lib/types'

const REPORT = {
  audit_id: 'audit-t4',
  site: 'Siège',
  link_type: 'fiber',
  measured: {
    rtt_idle_p50_ms: 20,
    rtt_loaded_p95_ms: 78,
    small_measured: true,
    http_small_p95_ms: 64,
    load_measured: true,
  },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: "Sous charge, la latence reste sous l'objectif." },
  cap_sla: { measured: true, ok_pct: 99.1, target_pct: 95, deadline_ms: 220 },
  bloat_grade: 'A',
  economics: { primary: { model: 'A', name: 'ligne dédiée', ar_per_hour: 1388.89 } },
  remediation: { qdisc: 'cake', cc: 'bbr' },
  remediation_state: 'available',
  remediation_available: true,
} as unknown as AuditReport

let auditState: Record<string, unknown>
let uiState: Record<string, unknown>

vi.mock('../../store/audit', () => ({
  useAuditStore: (selector: (s: Record<string, unknown>) => unknown) => selector(auditState),
}))
vi.mock('../../store/ui', () => ({
  useUIStore: (selector: (s: Record<string, unknown>) => unknown) => selector(uiState),
}))

function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}

beforeEach(() => {
  auditState = { report: null, stage: 'idle', failure: null, open: vi.fn(), start: vi.fn(), stop: vi.fn() }
  uiState = { setMode: vi.fn() }
})

describe('DecisionSpine — une seule source de vérité', () => {
  it('SynthèseView mesuré : un seul mot, rendu mot pour mot', () => {
    auditState = { ...auditState, report: REPORT, stage: 'done' }
    const host = mount(<SynthèseView />)
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
    expect(host.querySelector('#verdict-word')?.textContent).toContain('Lien sain')
  })

  it('aucun kicker au-dessus du verdict — un seul mot couronné', () => {
    const host = mount(<DecisionSpine report={REPORT} />)
    expect(host.querySelector('.answer-kicker')).toBeNull()
    expect(host.querySelector('.answer-band')?.textContent).not.toMatch(/réponse/i)
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
  })

  it('absence : MESURE ABSENTE en neutre, aucun chiffre fabriqué', () => {
    const host = mount(<SynthèseView />)
    const word = host.querySelector('#verdict-word')!
    expect(word.textContent).toContain('MESURE ABSENTE')
    expect(word.getAttribute('data-state')).toBe('absent')
    expect(word.innerHTML).toContain('var(--text-faint)')
    expect(word.innerHTML).not.toContain('var(--t-ok)')
    expect(host.textContent).not.toMatch(/\d+\s*ms\b/)
  })

  it('faits absents : un tiret et un mot, jamais zéro', () => {
    const host = mount(<DecisionSpine report={null} />)
    const slots = host.querySelectorAll('.state-slot[data-state="absent"]')
    expect(slots.length).toBe(4)
    for (const slot of slots) {
      expect(slot.textContent).toContain('—')
      expect(slot.textContent).not.toMatch(/^0\b/)
    }
  })

  it('AnswerBand ne duplique pas le mot du verdict', () => {
    const host = mount(<AnswerBand><DecisionSpine report={REPORT} /></AnswerBand>)
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
  })

  it('le Δ est signé : hausse et amélioration', () => {
    expect(mount(<DecisionSpine report={REPORT} />).textContent).toContain('+58 ms')
    const improved = {
      ...REPORT,
      measured: { ...REPORT.measured, rtt_idle_p50_ms: 100, rtt_loaded_p95_ms: 60 },
    } as unknown as AuditReport
    expect(mount(<DecisionSpine report={improved} />).textContent).toContain('-40 ms')
  })

  it('la phrase du moteur traverse sans réécriture ni coupe', () => {
    const sentence = "Sous charge, la latence reste sous l'objectif et 99,1 % des alertes passent à l'heure."
    const report = { ...REPORT, operator_verdict: { ...REPORT.operator_verdict!, sentence } } as unknown as AuditReport
    expect(mount(<DecisionSpine report={report} />).querySelector('.answer-sentence')?.textContent).toBe(sentence)
  })

  it('sans verdict du moteur, aucune couleur ni phrase de verdict', () => {
    for (const operator_verdict of [undefined, { level: '', word: '', sentence: '' }]) {
      const report = { ...REPORT, operator_verdict } as unknown as AuditReport
      const host = mount(<DecisionSpine report={report} />)
      const word = host.querySelector('#verdict-word')!
      expect(word.textContent).toContain('VERDICT ABSENT')
      expect(word.innerHTML).toContain('var(--text-faint)')
      expect(host.textContent).not.toContain("Sous charge, la latence reste sous l'objectif.")
    }
  })

  it('un niveau hors vocabulaire retombe sur le neutre, le mot reste entier', () => {
    const report = {
      ...REPORT,
      operator_verdict: { level: 'violet', word: 'Lien nouveau', sentence: 'Phrase du moteur.' },
    } as unknown as AuditReport
    const word = mount(<DecisionSpine report={report} />).querySelector('#verdict-word')!
    expect(word.textContent).toContain('Lien nouveau')
    expect(word.getAttribute('data-level')).toBe('neutral')
    expect(word.innerHTML).toContain('var(--text-faint)')
  })

  it('chaque couleur de verdict porte son mot', () => {
    const cases = [
      { level: 'neutral', dot: 'var(--text-faint)' },
      { level: 'red', dot: 'var(--t-danger-text)' },
      { level: 'amber', dot: 'var(--t-threshold)' },
      { level: 'green', dot: 'var(--t-ok)' },
    ] as const
    for (const c of cases) {
      const report = {
        ...REPORT,
        operator_verdict: { level: c.level, word: 'Mot du moteur', sentence: 'Phrase du moteur.' },
      } as unknown as AuditReport
      const word = mount(<DecisionSpine report={report} />).querySelector('#verdict-word')!
      expect(word.textContent).toContain('Mot du moteur')
      expect(word.innerHTML).toContain(c.dot)
    }
  })
})

describe('DecisionSpine — Act, quatre états exclusifs', () => {
  it('inactif : invitation, déclaration et armement', () => {
    expect(selectActState({ stage: 'idle', report: null })).toEqual({ kind: 'idle' })
    const host = mount(<DecisionSpine report={null} stage="idle" actions={{ declare: () => {}, arm: () => {} }} />)
    expect(host.textContent).toContain('Configurer le lien')
    expect(host.textContent).toContain('Démarrer la mesure')
    expect(host.textContent).not.toContain('ANNULER')
    expect(host.querySelector('.act-slot')?.getAttribute('data-state')).toBe('idle')
  })

  it('en mesure : ANNULER seul, aucune phase marquée terminée, aucun verdict — même avec un rapport périmé', () => {
    expect(selectActState({ stage: 'observing', report: REPORT })).toEqual({ kind: 'measuring' })
    // Hostile : un rapport terminé peut rester en mémoire quand une nouvelle
    // mesure démarre. Il ne doit jamais réapparaître comme verdict courant.
    const host = mount(<DecisionSpine report={REPORT} stage="observing" actions={{ cancel: () => {} }} />)
    expect(host.querySelector('.act-slot')?.getAttribute('data-state')).toBe('measuring')
    expect(host.querySelector('[role="status"]')?.textContent).toContain('Mesure en cours')
    expect(host.textContent).toContain('ANNULER')
    expect(host.textContent).not.toContain('Démarrer la mesure')
    expect(host.querySelector('#verdict-word')?.textContent).toContain('MESURE ABSENTE')
    expect(host.textContent).not.toContain('Lien sain')
    expect(host.querySelector('.answer-sentence')?.textContent).toContain('Mesure en cours')
    expect(host.textContent).not.toContain('Lancez une mesure pour obtenir un verdict.')
  })

  it('mesuré : la remédiation est gate sur la mesure', () => {
    expect(selectActState({ stage: 'done', report: REPORT })).toEqual({ kind: 'remediable' })
    const host = mount(<DecisionSpine report={REPORT} stage="done" actions={{ remediate: () => {} }} />)
    expect(host.querySelector('.act-slot')?.getAttribute('data-state')).toBe('remediable')
    expect(host.textContent).toContain('VOIR LA REMÉDIATION')
  })

  it('mesuré sans remédiation : phrase calme, aucune action', () => {
    const quiet = {
      ...REPORT,
      remediation: undefined,
      remediation_state: 'observation_only',
      remediation_available: false,
    } as unknown as AuditReport
    expect(selectActState({ stage: 'done', report: quiet })).toEqual({ kind: 'quiet' })
    const host = mount(<DecisionSpine report={quiet} stage="done" />)
    expect(host.querySelector('.act-slot')?.getAttribute('data-state')).toBe('quiet')
    expect(host.textContent).toContain("Aucun réglage proposé")
    expect(host.textContent).not.toContain('VOIR LA REMÉDIATION')
    expect(host.textContent).not.toContain('Démarrer la mesure')
  })

  it('un échec ne présente jamais un verdict précédent comme courant', () => {
    expect(selectActState({ stage: 'done', report: REPORT, failure: { message: 'lecture impossible' } }))
      .toEqual({ kind: 'quiet', reason: 'lecture impossible' })
    const host = mount(<DecisionSpine report={REPORT} stage="done" failure={{ message: 'lecture impossible' }} />)
    expect(host.querySelector('[role="status"]')?.textContent).toBe('lecture impossible')
    expect(host.querySelector('#verdict-word')?.textContent).not.toContain('Lien sain')
  })
})

describe('DecisionSpine — échelle', () => {
  const sheet = readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')
  const index = readFileSync(join(process.cwd(), 'src/styles/index.css'), 'utf8')

  it('le mot du verdict occupe l’étage hero, les faits l’étage metric', () => {
    expect(sheet).toMatch(/\.answer-word\s*\{[^}]*var\(--fs-hero\)/)
    expect(index).toMatch(/\.state-slot-num[^}]*\{[^}]*var\(--fs-metric\)/)
  })

  it('l’ancien hero a disparu des deux feuilles', () => {
    expect(sheet).not.toMatch(/\.verdict-word\s*\{/)
    expect(index).not.toMatch(/\.verdict-word\s*\{/)
  })
})

// §2 — la mesure noyau est primaire ; le tc reste une ligne nommée, jamais
// fusionnée. Aucun libellé ne prétend identifier une cause vérifiée.
describe('DecisionSpine — mesure noyau eBPF primaire, tc en ligne séparée', () => {
  const ebpfReport = {
    ...REPORT,
    kernel_telemetry: {
      available: true, source: 'ebpf', object_sha256: 'sha-abc', qdisc_drop_packets: 3,
      flows: [
        { local: '10.0.0.1:40000', peer: '10.0.0.2:5201', retrans_payload_bytes: 2048, retrans_events: 1, qdisc_drop_packets: 1, srtt_us: 12000 },
        { local: '10.0.0.1:40002', peer: '10.0.0.2:5201', retrans_payload_bytes: 2048, retrans_events: 225, qdisc_drop_packets: 2, srtt_us: 15000 },
      ],
    },
  } as unknown as AuditReport

  it('rend la somme eBPF, la médiane SRTT et une ligne tc séparée, sans cause confirmée', () => {
    const host = mount(<DecisionSpine report={ebpfReport} />)
    const loss = host.querySelector('[data-testid="kernel-loss"]')!.textContent!
    const latency = host.querySelector('[data-testid="kernel-latency"]')!.textContent!
    const tc = host.querySelector('[data-testid="kernel-tc"]')!.textContent!
    expect(loss).toContain('226 retransmissions observées')
    expect(latency).toContain('13,5 ms')
    expect(tc).toContain('tc/qdisc')
    expect(tc).toContain('3 paquets écartés')
    expect(loss).not.toContain('tc/qdisc')
    expect(host.textContent).not.toContain('cause confirmée')
  })

  it('un repli tcpinfo dit son absence et sa vraie raison, sans vert fabriqué', () => {
    const report = {
      ...REPORT,
      kernel_telemetry: { available: false, source: 'tcpinfo', reason: 'objet eBPF absent' },
    } as unknown as AuditReport
    const host = mount(<DecisionSpine report={report} />)
    for (const id of ['kernel-loss', 'kernel-latency', 'kernel-tc']) {
      const line = host.querySelector(`[data-testid="${id}"]`)!.textContent!
      expect(line).toContain('—')
      expect(line).toContain('non mesuré')
      expect(line).toContain('objet eBPF absent')
    }
    expect(host.textContent).not.toContain('226 retransmissions')
  })

  it('fait primary et santé viennent de l’eBPF, jamais du RTT userspace 900', () => {
    const hostile = {
      ...ebpfReport,
      measured: { ...REPORT.measured, small_measured: true, http_small_p95_ms: 900, rtt_loaded_p95_ms: 900 },
    } as unknown as AuditReport
    const host = mount(<DecisionSpine report={hostile} />)
    const delay = [...host.querySelectorAll('.state-slot')]
      .find(s => s.textContent?.includes('Réponse sous charge'))!
    expect(delay.textContent).toContain('13,5')
    expect(delay.textContent).not.toContain('900')
    expect(host.querySelector('.health-word')?.getAttribute('data-level')).toBe('attention')
    expect(host.querySelector('.health-word')?.textContent).toContain('À surveiller')
    expect(host.querySelector('.band-status')?.getAttribute('data-level')).toBe('attention')
    expect(host.querySelector('.band-number')?.textContent).toContain('13,5')
  })

  it('absence : data-primary=false, chaque ligne absente, aucune santé verte fabriquée', () => {
    const report = {
      ...REPORT,
      measured: { link_type: 'fiber' },
      kernel_telemetry: { available: false, source: 'tcpinfo', reason: 'objet eBPF absent' },
    } as unknown as AuditReport
    const host = mount(<DecisionSpine report={report} />)
    expect(host.querySelector('.kernel-evidence')?.getAttribute('data-primary')).toBe('false')
    for (const id of ['kernel-loss', 'kernel-latency', 'kernel-tc']) {
      expect(host.querySelector(`[data-testid="${id}"]`)?.getAttribute('data-absent')).toBe('true')
    }
    expect(host.querySelector('.health-word')?.getAttribute('data-level')).toBe('absent')
    expect(host.querySelector('.health-word')?.textContent).not.toContain('Lien prêt')
  })

  it('une sous-donnée manquante n’est pas un vert global : data-complete=false', () => {
    const report = {
      ...ebpfReport,
      kernel_telemetry: {
        ...ebpfReport.kernel_telemetry!,
        qdisc_drop_packets: undefined,
        flows: ebpfReport.kernel_telemetry!.flows!.map(f => ({ ...f, qdisc_drop_packets: Number.NaN })),
      },
    } as unknown as AuditReport
    const host = mount(<DecisionSpine report={report} />)
    expect(host.querySelector('.kernel-evidence')?.getAttribute('data-primary')).toBe('true')
    expect(host.querySelector('.kernel-evidence')?.getAttribute('data-complete')).toBe('false')
    expect(host.querySelector('[data-testid="kernel-tc"]')?.getAttribute('data-absent')).toBe('true')
    expect(host.querySelector('[data-testid="kernel-loss"]')?.getAttribute('data-absent')).toBe('false')
  })

  it('SRTT valide mais retrans illisible : « À surveiller », jamais le vert « Lien prêt »', () => {
    const report = {
      ...ebpfReport,
      kernel_telemetry: {
        ...ebpfReport.kernel_telemetry!,
        flows: [{ ...ebpfReport.kernel_telemetry!.flows![0], retrans_events: Number.NaN }],
      },
    } as unknown as AuditReport
    const host = mount(<DecisionSpine report={report} />)
    const health = host.querySelector('.health-word')!
    expect(health.getAttribute('data-level')).toBe('attention')
    expect(health.textContent).toContain('À surveiller')
    expect(health.textContent).not.toContain('Lien prêt')
    expect(health.getAttribute('data-level')).not.toBe('ok')
    expect(host.querySelector('.band-status')?.getAttribute('data-level')).toBe('attention')
  })
})
