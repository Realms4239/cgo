// Sheet 380 backdrop Auditer — Kit cockpit ① Auditer 30s Yas 4G → ② Comparer live → ③ Exporter
import { useEffect, useRef, useState } from 'react'
import { ArmButton } from '../components/ArmButton'
import { Timeline } from '../components/Timeline'
import { loadSettings } from '../lib/settings'
import Explain from '../components/Explain'
import { EmptyState } from '../components/ui/EmptyState'
import { Card } from '../components/ui/Card'
import { AuditLive } from '../components/AuditLive'
import { AuditReportCard } from '../components/AuditReportCard'
import {
  AUDIT_DURATIONS,
  DEFAULT_AUDIT_DIRECTION,
  auditStep,
  directionLabel,
  durationChoice,
  nextAuditStage,
  validateAuditDeclaration,
  type AuditStage,
} from '../lib/auditFlow'
import type { AuditDirection, AuditLive as AuditLiveState, AuditReport } from '../lib/types'
import { formatBytesFR, groupDigits } from '../lib/format'
import { useUIStore } from '../store/ui'
import { InlineField } from '../components/InlineField'
import { validate } from '../lib/validation'
import { animateShake } from '../lib/anime'
import { PeekPopover } from '../components/PeekPopover'
import AuditHistoryModal from '../components/AuditHistoryModal'
import { live as liveRing } from '../lib/live'
import { GATE_LABELS } from '../lib/gates'
import { canalGate, effectiveChannelModel, type ChannelModel } from '../lib/canal'

const ALL_QDISCS = ['pfifo_fast','fq_codel','cake'] as const
const ALL_CC = ['cubic','bbr'] as const

export default function CampagneView() {
  const live = useUIStore((s: any) => s.live)
  const [profiles, setProfiles] = useState<string[]>(['P2'])
  const [allProfiles, setAllProfiles] = useState<{ id: string; imported: boolean }[]>([{ id: 'P1', imported: false }, { id: 'P2', imported: false }])
  const refreshProfiles = () => fetch('/api/profiles').then(r => r.json()).then(j => { if (j?.profiles?.length) setAllProfiles(j.profiles.map((x: any) => ({ id: x.id, imported: !!x.imported }))) }).catch(() => { })
  useEffect(() => { refreshProfiles() }, [])
  // le cockpit ne montre que le référentiel P1–P4 ; les imports restent
  // joignables par l'API (pkg/profile) mais n'entrent plus dans la vue.
  const isRef = (id: string) => /^P[1-4]$/.test(id)
  const refProfiles = allProfiles.filter(p => isRef(p.id))
  const profCheck = (p: { id: string; imported: boolean }) => (
    <label key={p.id}><input type="checkbox" checked={profiles.includes(p.id)} onChange={e => setProfiles(e.target.checked ? [...profiles, p.id] : profiles.filter(x => x !== p.id))} /> <Explain term={p.id}>{p.id}</Explain></label>
  )
  const [reps, setReps] = useState(3)
  // 4 flux par défaut (régime du banc durci, décision 6) — 1 = régime historique.
  const [flows, setFlows] = useState(4)
  const [channelModel, setChannelModel] = useState<ChannelModel>('clean')
  // Porte du canal : rafales légales seulement si TOUS les profils sont P2/P3
  // (même règle que le 400 serveur). Bloqué ⇒ le modèle effectif retombe clean.
  const canal = canalGate(profiles)
  const effectiveModel = effectiveChannelModel(channelModel, canal)
  useEffect(() => {
    if (!canal.allowed && channelModel === 'bursty') setChannelModel('clean')
  }, [canal.allowed, channelModel])
  // La deadline voyage avec la campagne: le champ montre la valeur
  // réelle de l'opérateur et suit les Réglages (événement meteolink-settings).
  const [deadlineMs, setDeadlineMs] = useState(loadSettings().deadlineMs)
  useEffect(() => {
    const on = () => setDeadlineMs(loadSettings().deadlineMs)
    window.addEventListener('meteolink-settings', on)
    return () => window.removeEventListener('meteolink-settings', on)
  }, [])
  const [msg, setMsg] = useState('')
  const [auditMsg, setAuditMsg] = useState('')
  const [auditSite, setAuditSite] = useState('Département X')
  const [auditLink, setAuditLink] = useState('5g')
  const [auditProvider, setAuditProvider] = useState('')
  const [auditDuration, setAuditDuration] = useState(20)
  const [auditDirection, setAuditDirection] = useState<AuditDirection>(DEFAULT_AUDIT_DIRECTION)
  // Vide = puits du banc détecté côté serveur (dualns/legacy/external) ;
  // forcer 8.8.8.8 ici contournerait la détection M4 sur le banc durci.
  const [auditTarget, setAuditTarget] = useState('')
  const [presets, setPresets] = useState<{id:string;label:string;site:string;link_type:string;provider:string;duration:number;target:string;notes:string}[]>([])
  const [presetNote, setPresetNote] = useState('')
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyN, setHistoryN] = useState<number|null>(null)
  const [importMsg, setImportMsg] = useState('')
  const [importForm, setImportForm] = useState(false)
  const [auditAdvanced, setAuditAdvanced] = useState(() => {
    try { return localStorage.getItem('audit-advanced-open') === '1' } catch { return false }
  })
  const [auditStage, setAuditStage] = useState<AuditStage>('idle')
  const [auditLiveState, setAuditLiveState] = useState<AuditLiveState | null>(null)
  const [auditReport, setAuditReport] = useState<AuditReport | null>(null)
  const [auditIface, setAuditIface] = useState('')
  const [auditRosIface, setAuditRosIface] = useState('')
  const [impId, setImpId] = useState('')
  const [impCap, setImpCap] = useState(20)
  const [impCapUp, setImpCapUp] = useState(0)
  const [impDelay, setImpDelay] = useState(100)
  const [impJitter, setImpJitter] = useState(2)
  const [impLoss, setImpLoss] = useState(0)
  const [peek, setPeek] = useState<{ rect: DOMRect; data: number[] } | null>(null)

  const auditPollRef = useRef<number | null>(null)
  const auditStoppedRef = useRef(false)
  const auditFormRef = useRef<HTMLDivElement>(null)
  useEffect(() => () => { if (auditPollRef.current) clearInterval(auditPollRef.current) }, [])

  // Esc clears flash + audit error hint
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        useUIStore.getState().setFlash(null)
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [])

  const auditValidation = validateAuditDeclaration({
    site: auditSite,
    link_type: auditLink,
    duration: auditDuration,
    direction: auditDirection,
  })
  const auditCustomDuration = durationChoice(auditDuration) === 'custom'

  const copyText = (text: string, label: string) => {
    navigator.clipboard?.writeText(text).then(
      () => useUIStore.getState().pushToast(`${label} copié`, 'ok'),
      () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
    )
  }

  // presets RQ1 + compteur d'historique : chargés à l'ouverture de la carte
  const refreshHistoryN = () => {
    fetch('/api/audit/list').then(r => r.json()).then(j => setHistoryN((j.audits || []).length)).catch(() => {})
  }

  const stopAuditPoll = () => {
    if (auditPollRef.current) {
      clearInterval(auditPollRef.current)
      auditPollRef.current = null
    }
  }

  // Observer : /api/audit/live pour l'état vivant, /api/audit/status pour la
  // fin (rapport présent ou annulation). 2 s, comme la carte le promet.
  const pollAudit = () => {
    stopAuditPoll()
    const tick = () => {
      Promise.all([
        fetch('/api/audit/live').then(res => res.ok ? res.json() : Promise.reject(new Error(String(res.status)))),
        fetch('/api/audit/status').then(res => res.ok ? res.json() : Promise.reject(new Error(String(res.status)))),
      ]).then(([lv, st]) => {
        const liveFrame = lv as AuditLiveState
        setAuditLiveState(liveFrame?.running ? liveFrame : null)
        const report = (st?.report ?? null) as AuditReport | null
        if (report) setAuditReport(report)
        if (st?.running) return
        stopAuditPoll()
        if (report) {
          setAuditStage(s => nextAuditStage(s, 'finish'))
          const p95 = report.measured?.rtt_idle_p95_ms ?? 0
          setAuditMsg(p95 > 0 ? `audit terminé — p95 repos ${p95.toFixed(1)} ms` : 'audit terminé')
        } else {
          setAuditStage(s => nextAuditStage(s, 'cancel'))
          setAuditMsg(auditStoppedRef.current ? 'audit annulé — aucun rapport écrit' : 'audit interrompu — aucun rapport écrit')
          auditStoppedRef.current = false
        }
        refreshHistoryN()
      }).catch(() => {
        stopAuditPoll()
        setAuditLiveState(null)
        setAuditStage(s => s === 'observing' ? 'declared' : s)
        setAuditMsg('suivi interrompu — vérifiez /api/audit/status')
      })
    }
    void tick()
    auditPollRef.current = window.setInterval(tick, 2000)
  }

  const startAudit = async () => {
    const declaration = auditValidation.value
    if (!auditValidation.valid || !declaration) {
      if (auditFormRef.current) animateShake(auditFormRef.current)
      return
    }
    const store = useUIStore.getState()
    setAuditMsg('démarrage…')
    setAuditReport(null)
    setAuditLiveState(null)
    auditStoppedRef.current = false
    try {
      const r = await fetch('/api/audit/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          site: declaration.site,
          link_type: declaration.link_type,
          provider: auditProvider,
          duration: declaration.duration,
          direction: declaration.direction,
          target: auditTarget || undefined,
          iface: auditIface || undefined,
          routeros_iface: auditRosIface || undefined,
        }),
      })
      if (r.status === 409) throw new Error('un audit tourne déjà')
      if (!r.ok) throw new Error(String(r.status))
      await r.json().catch(() => ({}))
      store.pushToast(`Audit lancé — ${declaration.site}`, 'ok')
      store.setFlash({ type: 'success', msg: `Audit lancé — ${declaration.site}` })
      setAuditMsg('audit en cours…')
      setAuditStage(s => nextAuditStage(s, 'start'))
      pollAudit()
    } catch (e: any) {
      store.pushToast('Échec audit: ' + e.message, 'err')
      store.setFlash({ type: 'danger', msg: 'Échec audit' })
      if (auditFormRef.current) animateShake(auditFormRef.current)
      setAuditMsg('échec: ' + e.message)
    }
  }

  const stopAudit = async () => {
    auditStoppedRef.current = true
    setAuditMsg('annulation demandée…')
    try {
      const r = await fetch('/api/audit/stop', { method: 'POST' })
      if (!r.ok) throw new Error(String(r.status))
      useUIStore.getState().pushToast('Annulation demandée', 'blue')
    } catch (e: any) {
      auditStoppedRef.current = false
      setAuditMsg('échec annulation: ' + e.message)
      useUIStore.getState().pushToast('Échec annulation: ' + e.message, 'err')
    }
  }

  // À l'ouverture, la carte rejoint l'audit déjà en cours ou le dernier rapport.
  const syncAuditStatus = () => {
    fetch('/api/audit/status').then(r => r.json()).then(j => {
      if (j.running) {
        setAuditStage(s => (s === 'idle' || s === 'declared') ? 'observing' : s)
        setAuditMsg('audit en cours…')
        pollAudit()
      } else if (j.report) {
        setAuditReport(j.report as AuditReport)
        setAuditStage(s => (s === 'idle' || s === 'declared') ? 'done' : s)
      }
    }).catch(() => {})
  }

  const toggleAudit = () => {
    if (auditStage === 'observing') return
    if (auditStage !== 'idle') {
      stopAuditPoll()
      setAuditLiveState(null)
      setAuditStage(s => nextAuditStage(s, 'close'))
      return
    }
    setAuditStage(s => nextAuditStage(s, 'open'))
    fetch('/api/audit/presets').then(r => r.json()).then(j => setPresets(j.presets || [])).catch(() => {})
    refreshHistoryN()
    syncAuditStatus()
  }

  const toggleAuditAdvanced = () => {
    setAuditAdvanced(o => {
      try { localStorage.setItem('audit-advanced-open', o ? '0' : '1') } catch { }
      return !o
    })
  }

  const restartAudit = () => {
    setAuditStage(s => nextAuditStage(s, 'again'))
    setAuditReport(null)
    setAuditLiveState(null)
    setAuditMsg('')
    auditStoppedRef.current = false
  }

  const applyPreset = (p: { site: string; link_type: string; provider: string; duration: number; target: string; notes: string; label: string }) => {
    setAuditSite(p.site); setAuditLink(p.link_type); setAuditProvider(p.provider)
    setAuditDuration(p.duration); setAuditTarget(p.target); setPresetNote(p.notes)
  }

  const importValidation = validate(
    { id: impId, capacity: impCap, delay: impDelay, jitter: impJitter, loss: impLoss },
    { id: { required: true }, capacity: { min: 0.1 }, delay: { min: 1 }, jitter: { min: 0 }, loss: { min: 0 } }
  )

  const importProfile = () => {
    if (!importValidation.valid) return
    const body: any = { id: impId, capacity_mbps: impCap, delay_ms: impDelay, jitter_ms: impJitter, loss_pct: impLoss }
    // up asymétrique : 0/vide = symétrique (défaut historique) ; les rafales
    // Gilbert-Elliott restent à l'import CSV/API (avancé, 4 paramètres couplés)
    if (impCapUp > 0) body.capacity_up_mbps = impCapUp
    fetch('/api/profile/import', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(() => { setImportMsg(`profil ${impId} importé`); useUIStore.getState().pushToast(`Profil ${impId} importé`, 'ok'); setImportForm(false) })
      .catch(e => { setImportMsg('échec: ' + e.message); useUIStore.getState().pushToast('Échec import: ' + e.message, 'err') })
  }

  const pushToast = useUIStore((s: any)=>s.pushToast)
  const start = async () => {
    setMsg('démarrage…')
    try {
      const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles, reps, flows, channel_model: effectiveModel, deadline_ms: deadlineMs, target: loadSettings().target }) })
      if (!r.ok) {
        const txt = await r.text().catch(() => String(r.status))
        setMsg('échec: ' + txt.slice(0,120))
        pushToast(txt.includes('déjà') || r.status===409 ? 'Campagne déjà en cours — arrêter d\'abord' : 'Échec démarrage: ' + r.status, 'err')
        return
      }
      setMsg('campagne lancée')
      pushToast('Campagne lancée', 'ok')
      useUIStore.getState().setPanel('live')
    } catch (e:any) {
      setMsg('échec: ' + (e?.message ?? 'réseau'))
      pushToast('Échec démarrage', 'err')
    }
  }
  const stop = async () => {
    await fetch('/api/run/stop', { method: 'POST' })
    setMsg('arrêt demandé')
    pushToast('Arrêt demandé', 'blue')
  }

  // charge en mots courants — le statut brut `bulk-on` ne sort jamais tel quel
  const loadPlain = live?.load_status === 'bulk-on' ? 'transferts en cours'
    : live?.load_status === 'bulk-off' ? 'sans transfert'
    : live?.load_status ?? '—'
  const gates = live?.gates ?? Array(8).fill(null)
  const phase = live?.phase || 'idle'
  // Phase idle = '' côté serveur — tester sur les données reçues.
  const hasData = !!(live?.running || liveRing.small.length > 0)
  // Timeline sur les VRAIES bornes de phase (live.phaseSince, ts de première
  // frame mesurée par phase) — jamais de timestamps simulés ; sans phases
  // observées, pas de Timeline (vide honnête).
  const ps = liveRing.phaseSince
  const now = Date.now()
  const timeline = {
    baselineStart: ps.baseline ?? now - 30000,
    chargeStart: ps.charge ?? ps.baseline ?? now - 15000,
    chargeEnd: ps.recup ?? (ps.charge ? now : ps.baseline ?? now),
    recupEnd: (ps.recup ?? ps.charge ?? ps.baseline ?? now) + 15000,
  }
  const hasTimeline = ps.baseline != null || ps.charge != null || ps.recup != null

  return (
    <div style={{ display: 'flex', height: '100%', minHeight: 0, minWidth: 0, paddingRight: 400, boxSizing: 'border-box' }}>
      {/* main workspace — the 75% dead zone becomes the phase stepper + live state */}
      <div className="panel-stack" style={{ flex: 1, minWidth: 0 }}>
        <h1 className="view-title" style={{ margin: '16px auto' }}>Campagne — pilotez la mesure</h1>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 'var(--gap, 24px)' }}>
          {[
            { n: '①', t: 'Auditer', d: "30 s sur le lien réel, sans droits admin" },
            { n: '②', t: 'Comparer', d: 'pfifo vs CAKE en direct, même échelle' },
            { n: '③', t: 'Exporter', d: 'récit + chiffres vérifiables' },
          ].map(x => (
            <div key={x.n} className="card" style={{ display: 'flex', flexDirection: 'column', gap: 6, padding: 16 }}>
              <span className="mono" style={{ fontSize: 'var(--fs-head)', color: 'var(--t-live)' }}>{x.n}</span>
              <span className="mono" style={{ fontSize: 'var(--fs-tag)', letterSpacing: 'var(--tracking-wide)', textTransform: 'uppercase', color: 'var(--text-body)' }}>{x.t}</span>
              <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-muted)', lineHeight: 1.6, overflowWrap: 'break-word' }}>{x.d}</span>
            </div>
          ))}
        </div>
        {hasData && hasTimeline && <Timeline baselineStart={timeline.baselineStart} chargeStart={timeline.chargeStart} chargeEnd={timeline.chargeEnd} recupEnd={timeline.recupEnd} currentPhase={phase} />}
        {!hasData && <EmptyState kind="idle" hint="en attente — démarrez une campagne depuis le kit cockpit" />}
        <div className="card">
          <div className="card-head">État — flux SSE</div>
          <div className="kv-grid-3" style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 8, minWidth: 0 }}>
            <div className="kv"><span>phase</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{phase}</b></div>
            <div className="kv"><span>profil</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{live?.profile ?? '—'}</b></div>
            <div className="kv"><span>qdisc</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{live?.qdisc ?? '—'}</b></div>
            <div className="kv"><span>cc</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{live?.cc ?? '—'}</b></div>
            <div className="kv"><span>événement</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{live?.event_id != null ? `#${live.event_id} · rép ${live.repetition ?? 0}` : '—'}</b></div>
            <div className="kv"><span>charge</span><b className="mono" style={{ whiteSpace: 'nowrap' }}>{loadPlain}</b></div>
          </div>
        </div>
        <div className="card">
          <div className="card-head">Portes G0–G7</div>
          <div className="gates-detail" style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: '2px 16px' }}>
            {gates.map((g: boolean|null, i:number) => (
              <div key={i} className={'gate-row ' + (g===null?'na':g?'ok':'fail')} style={{ alignItems: 'baseline' }}>
                <span className="gate mono">G{i}</span>
                <span className="gate-lbl" style={{ flex: 1 }}><Explain term={`G${i}`}>{GATE_LABELS[i]}</Explain></span>
                <span className="gate-state mono" style={{ whiteSpace: 'nowrap', fontVariantNumeric: 'tabular-nums' }}>{g===null?'—':g?'PASS':'FAIL'}</span>
              </div>
            ))}
          </div>
        </div>
        {profiles.length > 0 && (
        <div className="card" data-testid="matrix-table">
          <div className="card-head">Matrice — état par cellule</div>
          <div style={{ overflowX: 'auto', minWidth: 0 }}>
          <table className="data-table" style={{ width:'100%', minWidth: 340, borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize: 'var(--fs-tag)' }}>
            <thead><tr style={{ color:'var(--text-body-strong)', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
              <th style={{ padding: 'var(--sp-1) var(--sp-2)' }}>profil</th><th>file × CC</th><th>répétitions</th><th>état</th>
            </tr></thead><tbody>
            {profiles.flatMap((p, pi) => ALL_QDISCS.flatMap((q, qi) => ALL_CC.map((c, ci) => {
              // ordre serveur (matrix.go) : profils × [pfifo, fq_codel, cake] × [cubic, bbr] × reps, event_id dès 1
              const cellStart = ((pi * ALL_QDISCS.length + qi) * ALL_CC.length + ci) * reps + 1
              // campagne arrêtée/terminée : event_id restant est un fantôme —
              // la progression ne se lit que sur un run ACTIF, sinon tout est en attente
              const cur = live?.running ? live?.event_id ?? 0 : 0
              const st = cur === 0 ? 'attente' : (cellStart + reps - 1 < cur ? 'terminée' : (cellStart <= cur ? 'en cours' : 'attente'))
              const n = Math.max(0, Math.min(reps, cur - cellStart + (st === 'en cours' ? 1 : 0)))
              const color = st === 'terminée' ? 'var(--t-ok)' : st === 'en cours' ? 'var(--t-live)' : 'var(--text-faint)'
              return (<tr key={`${p}/${q}/${c}`} style={{ borderBottom:'1px solid var(--hairline-faint)' }}>
                <td style={{ padding: 'var(--sp-1) var(--sp-2)' }}>{p}</td><td>{q} × {c}</td>
                <td style={{ fontVariantNumeric:'tabular-nums', whiteSpace:'nowrap' }}>{n}/{reps}</td>
                <td style={{ color, whiteSpace:'nowrap' }}>{st}</td></tr>)
            })))}
            </tbody></table>
          </div>
          <div className="mono muted" style={{ fontSize: 'var(--fs-micro)', marginTop:6, overflowWrap: 'anywhere' }}>La campagne teste chaque combinaison dans cet ordre : profils, puis files, puis contrôles de congestion, puis répétitions{live?.running && live?.event_id != null ? ` · événement #${live.event_id}/${live?.total_events ?? '—'}` : ''}</div>
        </div>
        )}
      </div>
      {/* cockpit sheet — matrix + audit + import */}
      <aside className="panel-stack cockpit" style={{position:'fixed', right:0, top:48, bottom:28, width:380, zIndex:40, backdropFilter:'blur(12px)', WebkitBackdropFilter:'blur(12px)', borderLeft:'1px solid var(--hairline)', background:'rgba(7,7,7,0.85)', overflowY:'auto', padding:12}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-micro)', color:'var(--text-muted)', marginBottom:4}}>live small — 20 pts</div>
          <div style={{display:'flex', alignItems:'end', gap:1, height:28}}>
            {peek.data.map((v,i)=>(
              <i key={i} style={{flex:1, height: `${Math.max(2, Math.min(28, (v/200)*28))}px`, background: 'var(--t-threshold)', borderRadius:1, opacity: 0.7 + (i/peek.data.length)*0.3}} />
            ))}
          </div>
          <div className="mono" style={{fontSize: 'var(--fs-micro)', color:'var(--t-threshold)', marginTop:4}}>{peek.data.at(-1)?.toFixed(1) ?? '—'} ms</div>
        </PeekPopover>
      )}
      <div className="card">
        <div className="card-head">Campagne — Kit cockpit</div>
        <div className="form-row">
          <label>Profils</label>
          <div className="check-row">
            {refProfiles.map(p => profCheck(p))}
          </div>
        </div>
        <div className="form-row">
          <label>Files (AQM)</label><span className="mono muted">{ALL_QDISCS.map((q, i) => <span key={q}><Explain term={q}>{q}</Explain>{i < ALL_QDISCS.length - 1 ? ' · ' : ''}</span>)}</span>
        </div>
        <div className="form-row">
          <label>CC</label><span className="mono muted">{ALL_CC.map((c, i) => <span key={c}><Explain term={c}>{c}</Explain>{i < ALL_CC.length - 1 ? ' · ' : ''}</span>)}</span>
        </div>
        <div className="form-row">
          <label>Répétitions</label>
          <input type="number" min={1} max={5} placeholder="3" value={reps} onChange={e=>setReps(parseInt(e.target.value)||1)} style={{width:64}} />
        </div>
        <div className="form-row">
          <label title=">1 sature un lien avec perte où un CUBIC mono-flux s'effondre">Flux de charge</label>
          <input type="number" min={1} max={8} value={flows} onChange={e=>setFlows(parseInt(e.target.value)||1)} style={{width:64}} />
          <span className="mono muted" style={{fontSize: 'var(--fs-micro)'}}>1–8 · défaut 4</span>
        </div>
        <div className="form-row">
          <label><Explain term="canal">Canal</Explain></label>
          <div role="group" aria-label="Modèle de canal" style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
            {(['clean', 'bursty'] as const).map(m => {
              const blocked = m === 'bursty' && !canal.allowed
              const active = effectiveModel === m
              return (
                <button
                  key={m} type="button" className="btn" aria-pressed={active}
                  disabled={blocked}
                  title={blocked ? canal.reason : m === 'bursty' ? 'rafales — canal propre P2/P3' : 'canal propre, sans rafales'}
                  onClick={() => setChannelModel(m)}
                  style={{
                    padding: 'var(--sp-1) var(--sp-2)',
                    borderColor: active ? 'var(--t-live)' : 'var(--hairline)',
                    color: active ? 'var(--t-live)' : 'var(--text-muted)',
                    background: active ? 'var(--live-dim)' : 'transparent',
                    opacity: blocked ? 'var(--disabled-opacity, 0.5)' : 1,
                    cursor: blocked ? 'not-allowed' : 'pointer',
                  }}
                >{m}</button>
              )
            })}
          </div>
        </div>
        {!canal.allowed && (
          <div className="mono muted" data-testid="canal-reason" style={{ fontSize: 'var(--fs-micro)', marginTop: -4 }}>{canal.reason}</div>
        )}
        <div className="form-row">
          <label><Explain term="deadline">Deadline</Explain></label>
          <input title="objectif small p95 — voyage avec la campagne (bornes serveur 200–5000 ms)" type="number" min={200} max={5000} step={100} placeholder="1000" value={deadlineMs} onChange={e=>setDeadlineMs(parseInt(e.target.value)||1000)} style={{width:80, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)'}} />
          <span className="mono muted" style={{fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-micro)'}}>ms</span>
        </div>
        <div className="form-row" style={{gap:8, border:'1px solid var(--hairline)', background:'rgba(244,180,0,0.06)', padding: 'var(--sp-1) var(--sp-2)', marginTop:6}} onMouseEnter={e=>setPeek({rect:e.currentTarget.getBoundingClientRect(), data: liveRing.small.slice(-20).map(([,v]:[number,number])=>v)})} onMouseLeave={()=>setPeek(null)}>
          <span className="mono" style={{fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-micro)', letterSpacing: 'var(--tracking-wide)', textTransform:'uppercase', color:'var(--text-muted)'}}>cost preview</span>
          <span className="mono" style={{fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)', color: live?.wasted_bytes != null ? 'var(--t-threshold)' : 'var(--text-faint)', fontVariantNumeric:'tabular-nums'}}>
            {(() => {
              if (!live || live.wasted_bytes == null || live.cost_ar_per_h == null) return '—'
              const w = formatBytesFR(live.wasted_bytes)
              return `${w.num} ${w.unit} gâchés · ${groupDigits(live.cost_ar_per_h, 2)} Ar/h`
            })()}
          </span>
          <span className="mono muted" style={{fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-micro)', marginLeft:'auto'}}>octets gâchés × tarif unique</span>
        </div>
        <div className="form-row" style={{gap:8}}>
          <span title={live?.running ? 'campagne en cours — arrêter d\'abord' : undefined}><ArmButton label="DÉMARRER" onConfirm={start} disabled={live?.running} /></span>
          <ArmButton label="ARRÊTER" confirmLabel="CONFIRMER L'ARRÊT" onConfirm={stop} disabled={!live?.running} />
          <span className="mono muted">{msg}</span>
        </div>
      </div>

      <Card
        head="Audit du lien"
        term="audit"
        sub="audit standard · non intrusif"
        testid="audit-card"
        right={(
          <div style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
            <button className="btn" onClick={() => setHistoryOpen(true)} style={{ padding: 'var(--sp-1) var(--sp-2)' }} title="audits gelés : rapprochement au référentiel, import par ligne">HISTORIQUE{historyN !== null ? ` (${historyN})` : ''}</button>
            <button className="btn" onClick={toggleAudit} disabled={auditStage === 'observing'} style={{ padding: 'var(--sp-1) var(--sp-2)' }} title={auditStage === 'observing' ? 'audit en cours — annuler d’abord' : undefined}>{auditStage === 'idle' ? 'OUVRIR' : 'FERMER'}</button>
          </div>
        )}
      >
        {auditStage !== 'idle' && (
          <div aria-label="étapes de l'audit" role="list" className="mono" style={{ display: 'flex', gap: 'var(--sp-2)', flexWrap: 'wrap', fontSize: 'var(--fs-micro)' }}>
            {['Déclarer', 'Observer', 'Lire'].map((label, i) => {
              const active = auditStep(auditStage) === i + 1
              return (
                <span key={label} role="listitem" aria-current={active ? 'step' : undefined} style={{ color: active ? 'var(--t-live)' : 'var(--text-faint)' }}>
                  {i + 1}. {label}
                </span>
              )
            })}
          </div>
        )}

        {auditStage === 'declared' && <>
          {presets.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--sp-1)' }}>
              {presets.map(p => (
                <button key={p.id} className="c-chip" style={{ cursor: 'pointer' }} title={`${p.site} · ${p.link_type} · ${p.duration}s → ${p.target} — ${p.notes}`} onClick={() => applyPreset(p)}>⚑ {p.label}</button>
              ))}
            </div>
          )}
          {presetNote && <div className="mono muted" style={{ fontSize: 'var(--fs-micro)' }}>ⓘ {presetNote}</div>}
          <div ref={auditFormRef} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
            <InlineField label="Site" error={auditValidation.errors.site} helper={!auditSite ? 'requis' : undefined}>
              <input className="c-input" name="site" value={auditSite} onChange={e => setAuditSite(e.target.value)} />
            </InlineField>
            <InlineField label="Type de lien" error={auditValidation.errors.link_type}>
              <select className="c-input" value={auditLink} onChange={e => setAuditLink(e.target.value)}>
                {['fiber', '5g', '4g', 'vsat', 'other'].map(t => <option key={t} value={t}>{t}</option>)}
              </select>
            </InlineField>
            <InlineField label="Durée" error={auditValidation.errors.duration} helper={auditCustomDuration ? `durée libre : ${auditDuration} s` : '20 ou 60 s — durée libre dans avancé'}>
              <div role="group" aria-label="Durée de l'audit" style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
                {AUDIT_DURATIONS.map(d => (
                  <button
                    key={d} type="button" className="btn" aria-pressed={auditDuration === d}
                    onClick={() => setAuditDuration(d)}
                    style={{
                      padding: 'var(--sp-1) var(--sp-2)',
                      borderColor: auditDuration === d ? 'var(--t-live)' : 'var(--hairline)',
                      color: auditDuration === d ? 'var(--t-live)' : 'var(--text-muted)',
                      background: auditDuration === d ? 'var(--live-dim)' : 'transparent',
                    }}
                  >{d} s</button>
                ))}
              </div>
            </InlineField>
            <InlineField label="Sens" helper="fenêtre de charge exercée">
              <div role="group" aria-label="Sens de l'audit" style={{ display: 'flex', gap: 'var(--sp-1)', flexWrap: 'wrap' }}>
                {(['up', 'down', 'both'] as const).map(d => (
                  <button
                    key={d} type="button" className="btn" aria-pressed={auditDirection === d}
                    onClick={() => setAuditDirection(d)}
                    style={{
                      padding: 'var(--sp-1) var(--sp-2)',
                      borderColor: auditDirection === d ? 'var(--t-live)' : 'var(--hairline)',
                      color: auditDirection === d ? 'var(--t-live)' : 'var(--text-muted)',
                      background: auditDirection === d ? 'var(--live-dim)' : 'transparent',
                    }}
                  >{directionLabel(d)}</button>
                ))}
              </div>
            </InlineField>

            <button className="btn" aria-expanded={auditAdvanced} onClick={toggleAuditAdvanced} style={{ alignSelf: 'flex-start', padding: 'var(--sp-1) var(--sp-2)' }}>
              {auditAdvanced ? 'MASQUER LES CHAMPS AVANCÉS' : 'CHAMPS AVANCÉS'}
            </button>
            {auditAdvanced && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-2)' }}>
                <InlineField label="Fournisseur" helper="ex. yas — gelé avec la ligne">
                  <input className="c-input" name="audit-provider" value={auditProvider} onChange={e => setAuditProvider(e.target.value)} placeholder="yas" />
                </InlineField>
                <InlineField label="Cible ping" helper="vide = puits du banc détecté">
                  <input className="c-input" name="audit-target" value={auditTarget} onChange={e => setAuditTarget(e.target.value)} placeholder="détecté" />
                </InlineField>
                <InlineField label="Interface Linux" helper="optionnel — ex. eth0, pour le script tc">
                  <input className="c-input" name="audit-iface" value={auditIface} onChange={e => setAuditIface(e.target.value)} placeholder="eth0" />
                </InlineField>
                <InlineField label="Interface RouterOS" helper="optionnel — ex. ether1-wan">
                  <input className="c-input" name="audit-ros-iface" value={auditRosIface} onChange={e => setAuditRosIface(e.target.value)} placeholder="ether1-wan" />
                </InlineField>
                <InlineField label="Durée libre (s)" helper="10–600 s">
                  <input className="c-input" type="number" min={10} max={600} value={auditDuration} onChange={e => setAuditDuration(parseInt(e.target.value) || 0)} />
                </InlineField>
              </div>
            )}
          </div>
          <div className="form-row" style={{ gap: 'var(--sp-2)', marginTop: 'var(--sp-2)' }}>
            <ArmButton label="LANCER AUDIT" confirmLabel="CONFIRMER AUDIT" onConfirm={startAudit} disabled={!auditValidation.valid} />
            <span className="mono muted">{auditMsg}</span>
          </div>
        </>}

        {auditStage === 'observing' && (
          <>
            <AuditLive live={auditLiveState} direction={auditDirection} onStop={stopAudit} />
            {auditMsg && <span className="mono muted" style={{ fontSize: 'var(--fs-micro)' }}>{auditMsg}</span>}
          </>
        )}

        {auditStage === 'done' && auditReport && (
          <>
            <AuditReportCard report={auditReport} onCopy={copyText} />
            <div className="form-row" style={{ gap: 'var(--sp-2)', marginTop: 'var(--sp-2)' }}>
              <button className="btn" onClick={restartAudit}>NOUVEL AUDIT</button>
              <span className="mono muted" style={{ fontSize: 'var(--fs-micro)' }}>{auditMsg}</span>
            </div>
          </>
        )}
        {auditStage === 'done' && !auditReport && (
          <div className="mono muted" style={{ fontSize: 'var(--fs-micro)' }}>rapport absent — relancez un audit.</div>
        )}

        {historyOpen && <AuditHistoryModal onClose={() => { setHistoryOpen(false); refreshHistoryN() }} onImported={m => setAuditMsg(m)} />}
      </Card>

      <div className="card">
        <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', minWidth: 0 }}>
          <span style={{ minWidth: 0, overflowWrap: 'anywhere' }}>Profil personnalisé</span>
          {!importForm && <>
            <button className="btn" onClick={async () => {
              try {
                const r = await fetch('/api/audit/toprofile', { method: 'POST' })
                const j = await r.json().catch(() => ({}))
                if (!r.ok) { useUIStore.getState().pushToast(j?.error ?? 'échec import audit', 'err'); return }
                useUIStore.getState().pushToast(`Profil ${j.profile?.id ?? 'P-audit'} importé depuis l'audit — ${j.profile?.capacity_mbps} Mbit/s, ${j.profile?.delay_ms} ms, perte ${j.profile?.loss_pct ?? 0} %`, 'ok')
              } catch { useUIStore.getState().pushToast('échec import audit', 'err') }
            }} title="transforme le dernier audit du lien réel en profil rejouable sur le banc" style={{ marginLeft: 'auto', padding: '2px 10px' }}>AUDIT → PROFIL</button>
            <button className="btn" onClick={()=>setImportForm(true)} style={{ padding: '2px 10px' }}>IMPORTER</button>
            <span className="mono muted" style={{ fontSize: 'var(--fs-micro)' }}>{importMsg}</span>
          </>}
        </div>
        {importForm && <>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 8 }}>
            <InlineField label="Identifiant" error={importValidation.errors.id} helper="jamais P1–P4 (référentiel protégé) — ex. P-site-fibre">
              <input name="import-id" value={impId} onChange={e=>setImpId(e.target.value)} placeholder="P-site-fibre" style={{ background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <InlineField label="Capacité (Mbit/s)" error={importValidation.errors.capacity} helper="> 0">
              <input type="number" min={0.1} step={0.1} value={impCap} onChange={e=>setImpCap(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <InlineField label="Cap. montante (Mbit/s)" error={undefined} helper="0 = symétrique — ex. 5 pour 4G 20/5">
              <input type="number" min={0} step={0.1} value={impCapUp} onChange={e=>setImpCapUp(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <InlineField label="RTT (ms)" error={importValidation.errors.delay} helper="> 0">
              <input type="number" min={1} value={impDelay} onChange={e=>setImpDelay(parseInt(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <InlineField label="Gigue (ms)" error={importValidation.errors.jitter} helper="≥ 0 — ex. 30 pour VSAT">
              <input type="number" min={0} step={0.5} value={impJitter} onChange={e=>setImpJitter(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <InlineField label="Perte (%)" error={importValidation.errors.loss} helper="≥ 0 — ex. 1 pour VSAT">
              <input type="number" min={0} step={0.1} value={impLoss} onChange={e=>setImpLoss(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding: 'var(--sp-1) var(--sp-2)', fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-tag)' }} />
            </InlineField>
            <div className="form-row" style={{gap:8}}>
              <ArmButton label="CONFIRMER IMPORT" onConfirm={importProfile} disabled={!importValidation.valid} />
              <button className="btn" onClick={()=>setImportForm(false)}>ANNULER</button>
            </div>
          </div>
        </>}
      </div>
      </aside>
    </div>
  )
}
