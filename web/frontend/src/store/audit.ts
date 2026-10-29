// Session d'audit à la racine de l'application. Quitter la vue Auditer ne
// coupe ni le suivi ni le rapport : le minuteur vit ici, pas dans le composant.
import { create } from 'zustand'
import {
  DEFAULT_AUDIT_DIRECTION,
  nextAuditStage,
  validateAuditDeclaration,
  type AuditStage,
} from '../lib/auditFlow'
import type { AuditDirection, AuditLive, AuditReport } from '../lib/types'
import {
  compareAudits,
  loadAuditPair,
  saveLastSummary,
  summaryOf,
  type AuditComparison,
  type AuditSummary,
} from '../lib/auditSummary'
import { useUIStore } from './ui'

export interface AuditPreset {
  id: string
  label: string
  site: string
  link_type: string
  provider: string
  duration: number
  target: string
  notes: string
}

interface AuditFields {
  site: string
  link: string
  provider: string
  duration: number
  direction: AuditDirection
  target: string
  iface: string
  rosIface: string
  advanced: boolean
}

interface AuditState extends AuditFields {
  stage: AuditStage
  presets: AuditPreset[]
  presetNote: string
  historyN: number | null
  msg: string
  live: AuditLive | null
  report: AuditReport | null
  previous: AuditSummary | null
  comparison: AuditComparison | null
  patch: (p: Partial<AuditFields>) => void
  toggleAdvanced: () => void
  open: () => void
  close: () => void
  restart: () => void
  applyPreset: (p: AuditPreset) => void
  start: () => Promise<void>
  stop: () => Promise<void>
  restore: () => void
  showReport: () => void
  _live: (live: AuditLive | null) => void
  _finish: (report: AuditReport) => void
  _restore: (report: AuditReport) => void
  _cancel: (byOperator: boolean) => void
  _interrupted: () => void
  _history: () => void
}

let pollTimer: number | null = null
let stoppedByOperator = false

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function poll() {
  stopPolling()
  const tick = () => {
    Promise.all([
      fetch('/api/audit/live').then(r => (r.ok ? r.json() : Promise.reject(new Error(String(r.status))))),
      fetch('/api/audit/status').then(r => (r.ok ? r.json() : Promise.reject(new Error(String(r.status))))),
    ])
      .then(([lv, st]) => {
        const s = useAuditStore.getState()
        s._live(lv?.running ? (lv as AuditLive) : null)
        const report = (st?.report ?? null) as AuditReport | null
        if (report) s._finish(report)
        if (st?.running) return
        stopPolling()
        if (report) {
          useUIStore.getState().pushToast('Audit terminé', 'ok')
        } else {
          s._cancel(stoppedByOperator)
        }
        stoppedByOperator = false
        s._history()
      })
      .catch(() => {
        stopPolling()
        useAuditStore.getState()._interrupted()
      })
  }
  void tick()
  pollTimer = window.setInterval(tick, 2000)
}

const advancedOpen = () => {
  try {
    return localStorage.getItem('audit-advanced-open') === '1'
  } catch {
    return false
  }
}

export const useAuditStore = create<AuditState>((set, get) => ({
  stage: 'idle',
  site: 'Département X',
  link: '5g',
  provider: '',
  duration: 20,
  direction: DEFAULT_AUDIT_DIRECTION,
  target: '',
  iface: '',
  rosIface: '',
  advanced: advancedOpen(),
  presets: [],
  presetNote: '',
  historyN: null,
  msg: '',
  live: null,
  report: null,
  previous: null,
  comparison: null,

  patch: p => set(p),

  toggleAdvanced: () =>
    set(s => {
      try {
        localStorage.setItem('audit-advanced-open', s.advanced ? '0' : '1')
      } catch {}
      return { advanced: !s.advanced }
    }),

  open: () => {
    set(s => ({ stage: s.stage === 'idle' ? nextAuditStage(s.stage, 'open') : s.stage }))
    fetch('/api/audit/presets')
      .then(r => r.json())
      .then(j => set({ presets: Array.isArray(j?.presets) ? j.presets : [] }))
      .catch(() => {})
    get()._history()
    fetch('/api/audit/status')
      .then(r => r.json())
      .then(j => {
        if (j?.running) {
          set({ stage: 'observing', msg: 'audit en cours…' })
          poll()
        }
      })
      .catch(() => {})
  },

  close: () => {
    if (get().stage === 'observing') return
    stopPolling()
    set(s => ({ stage: nextAuditStage(s.stage, 'close'), live: null, msg: '' }))
  },

  restart: () => {
    stopPolling()
    set(s => ({
      stage: nextAuditStage(s.stage, 'again'),
      report: null,
      live: null,
      msg: '',
    }))
    stoppedByOperator = false
  },

  applyPreset: p =>
    set({
      site: p.site,
      link: p.link_type,
      provider: p.provider,
      duration: p.duration,
      target: p.target,
      presetNote: p.notes,
    }),

  start: async () => {
    const s = get()
    const v = validateAuditDeclaration({
      site: s.site,
      link_type: s.link,
      duration: s.duration,
      direction: s.direction,
    })
    if (!v.valid || !v.value) return
    const store = useUIStore.getState()
    stoppedByOperator = false
    set({ msg: 'démarrage…', report: null, live: null })
    try {
      const r = await fetch('/api/audit/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          site: v.value.site,
          link_type: v.value.link_type,
          provider: s.provider,
          duration: v.value.duration,
          direction: v.value.direction,
          target: s.target.trim() || undefined,
          iface: s.iface || undefined,
          routeros_iface: s.rosIface || undefined,
        }),
      })
      if (r.status === 409) throw new Error('un audit tourne déjà')
      if (!r.ok) throw new Error(`échec ${r.status}`)
      await r.json().catch(() => ({}))
      store.pushToast(`Audit lancé — ${v.value.site}`, 'ok')
      store.setFlash({ type: 'success', msg: `Audit lancé — ${v.value.site}` })
      set({ stage: 'observing', msg: 'audit en cours…' })
      poll()
    } catch (e: any) {
      const reason = e?.message ?? 'réseau'
      store.pushToast(`Échec audit : ${reason}`, 'err')
      store.setFlash({ type: 'danger', msg: 'Échec audit' })
      set({ msg: `échec : ${reason}` })
    }
  },

  stop: async () => {
    stoppedByOperator = true
    set({ msg: 'annulation demandée…' })
    try {
      const r = await fetch('/api/audit/stop', { method: 'POST' })
      if (!r.ok) throw new Error(String(r.status))
      useUIStore.getState().pushToast('Annulation demandée', 'blue')
    } catch (e: any) {
      stoppedByOperator = false
      const reason = e?.message ?? 'réseau'
      set({ msg: `échec annulation : ${reason}` })
      useUIStore.getState().pushToast(`Échec annulation : ${reason}`, 'err')
    }
  },

  restore: () => {
    fetch('/api/audit/status')
      .then(r => r.json())
      .then(j => {
        if (j?.running) {
          set({ stage: 'observing', msg: 'audit en cours…' })
          poll()
          return
        }
        if (j?.report) {
          get()._restore(j.report as AuditReport)
          return
        }
        fetch('/api/audit/report')
          .then(r => (r.ok ? r.json() : null))
          .then(rep => {
            if (rep?.audit_id) get()._restore(rep as AuditReport)
          })
          .catch(() => {})
      })
      .catch(() => {})
  },

  showReport: () => {
    if (get().report) set({ stage: 'done', msg: '' })
  },

  _live: live => set({ live }),

  _finish: report =>
    set(s => {
      if (s.report?.audit_id === report.audit_id && s.stage === 'done') return { live: null }
      const current = summaryOf(report)
      const pair = loadAuditPair()
      const previous =
        pair.last && pair.last.audit_id !== current.audit_id ? pair.last : pair.previous
      saveLastSummary(current)
      const p95 = report.measured?.rtt_idle_p95_ms ?? 0
      return {
        report,
        stage: 'done',
        live: null,
        previous,
        comparison: compareAudits(previous, current),
        msg: p95 > 0 ? `audit terminé — p95 repos ${p95.toFixed(1)} ms` : 'audit terminé',
      }
    }),

  _restore: report =>
    set(s => {
      if (s.report?.audit_id === report.audit_id) return {}
      const current = summaryOf(report)
      const pair = loadAuditPair()
      const previous =
        pair.last && pair.last.audit_id !== current.audit_id ? pair.last : pair.previous
      return { report, previous, comparison: compareAudits(previous, current) }
    }),

  _cancel: byOperator =>
    set(s => ({
      stage: nextAuditStage(s.stage, 'cancel'),
      live: null,
      msg: byOperator ? 'audit annulé — aucun rapport écrit' : 'audit interrompu — aucun rapport écrit',
    })),

  _interrupted: () =>
    set(s => ({
      stage: s.stage === 'observing' ? 'declared' : s.stage,
      live: null,
      msg: 'suivi interrompu — vérifiez /api/audit/status',
    })),

  _history: () =>
    fetch('/api/audit/list')
      .then(r => r.json())
      .then(j => set({ historyN: (j.audits || []).length }))
      .catch(() => {}),
}))
