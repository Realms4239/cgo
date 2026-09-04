// Sheet 380 backdrop Auditer — Kit cockpit ① Auditer 30s Yas 4G → ② Comparer live → ③ Exporter
import { useEffect, useRef, useState } from 'react'
import { ArmButton } from '../components/ArmButton'
import { Timeline } from '../components/Timeline'
import { loadSettings } from '../lib/settings'
import Explain from '../components/Explain'
import { useUIStore } from '../store/ui'
import { InlineField } from '../components/InlineField'
import { validate } from '../lib/validation'
import { animateShake } from '../lib/anime'
import { PeekPopover } from '../components/PeekPopover'
import { live as liveRing } from '../lib/live'
import { GATE_LABELS } from '../lib/gates'

const ALL_QDISCS = ['pfifo_fast','fq_codel','cake'] as const
const ALL_CC = ['cubic','bbr'] as const

export default function CampagneView() {
  const live = useUIStore((s: any) => s.live)
  const [profiles, setProfiles] = useState<string[]>(['P2'])
  const [allProfiles, setAllProfiles] = useState<{ id: string; imported: boolean }[]>([{ id: 'P1', imported: false }, { id: 'P2', imported: false }])
  useEffect(() => { fetch('/api/profiles').then(r => r.json()).then(j => { if (j?.profiles?.length) setAllProfiles(j.profiles.map((x: any) => ({ id: x.id, imported: !!x.imported }))) }).catch(() => { }) }, [])
  const [reps, setReps] = useState(3)
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
  const [auditDuration, setAuditDuration] = useState(30)
  const [importMsg, setImportMsg] = useState('')
  const [importForm, setImportForm] = useState(false)
  const [auditOpen, setAuditOpen] = useState(false)
  const [auditLast, setAuditLast] = useState<any>(null)
  const [impId, setImpId] = useState('')
  const [impCap, setImpCap] = useState(20)
  const [impCapUp, setImpCapUp] = useState(0)
  const [impDelay, setImpDelay] = useState(100)
  const [impJitter, setImpJitter] = useState(2)
  const [impLoss, setImpLoss] = useState(0)
  const [peek, setPeek] = useState<{ rect: DOMRect; data: number[] } | null>(null)

  const auditPollRef = useRef<number | null>(null)
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

  const auditValidation = validate(
    { site: auditSite, link_type: auditLink, duration: auditDuration },
    { site: { required: true }, link_type: { required: true }, duration: { min: 10, max: 600 } }
  )

  const startAudit = async () => {
    if (!auditValidation.valid) {
      if (auditFormRef.current) animateShake(auditFormRef.current)
      return
    }
    const store = useUIStore.getState()
    // All+optimistic: optimistic Toast+Flash success
    store.pushToast(`Audit lancé — ${auditSite}`, 'ok')
    store.setFlash({ type: 'success', msg: `Audit lancé — ${auditSite}` })
    setAuditMsg('audit en cours…')
    try {
      const r = await fetch('/api/audit/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ site: auditSite, link_type: auditLink, duration: auditDuration }),
      })
      if (!r.ok) throw new Error(String(r.status))
      await r.json().catch(() => ({}))
      const poll = window.setInterval(() => {
        fetch('/api/audit/status')
          .then((res) => res.json())
          .then((j) => {
            setAuditLast(j.last ?? null)
            if (!j.running) {
              clearInterval(poll)
              auditPollRef.current = null
              setAuditMsg(j.last ? `audit terminé — p95 ${Number(j.last.rtt_idle_p95_ms).toFixed(1)} ms` : 'audit terminé')
            }
          })
          .catch(() => {
            clearInterval(poll)
            auditPollRef.current = null
          })
      }, 2000)
      auditPollRef.current = poll
    } catch (e: any) {
      // rollback: err toast + danger flash + shake
      store.pushToast('Échec audit: ' + e.message, 'err')
      store.setFlash({ type: 'danger', msg: 'Échec audit' })
      if (auditFormRef.current) animateShake(auditFormRef.current)
      setAuditMsg('échec: ' + e.message)
    }
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
      const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles, reps, deadline_ms: deadlineMs, target: loadSettings().target }) })
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
        <h1 className="view-title">Campagne — pilotez la mesure</h1>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 'var(--gap, 24px)' }}>
          {[
            { n: '①', t: 'Auditer', d: "30 s sur le lien réel, sans droits admin" },
            { n: '②', t: 'Comparer', d: 'pfifo vs CAKE en direct, même échelle' },
            { n: '③', t: 'Exporter', d: 'constat MD/CSV signé' },
          ].map(x => (
            <div key={x.n} className="card" style={{ display: 'flex', flexDirection: 'column', gap: 6, padding: 16 }}>
              <span className="mono" style={{ fontSize: 18, color: 'var(--t-live, #5ad3e3)' }}>{x.n}</span>
              <span className="mono" style={{ fontSize: 11, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-body, #f2f2f4)' }}>{x.t}</span>
              <span className="mono" style={{ fontSize: 10, color: 'var(--text-muted, #8b9099)', lineHeight: 1.6 }}>{x.d}</span>
            </div>
          ))}
        </div>
        {hasData && hasTimeline && <Timeline baselineStart={timeline.baselineStart} chargeStart={timeline.chargeStart} chargeEnd={timeline.chargeEnd} recupEnd={timeline.recupEnd} currentPhase={phase} />}
        <div className="card">
          <div className="card-head">État — flux SSE</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 8 }}>
            <div className="kv"><span>phase</span><b className="mono">{phase}</b></div>
            <div className="kv"><span>profil</span><b className="mono">{live?.profile ?? '—'}</b></div>
            <div className="kv"><span>qdisc</span><b className="mono">{live?.qdisc ?? '—'}</b></div>
            <div className="kv"><span>cc</span><b className="mono">{live?.cc ?? '—'}</b></div>
            <div className="kv"><span>événement</span><b className="mono">{live?.event_id != null ? `#${live.event_id} · rép ${live.repetition ?? 0}` : '—'}</b></div>
            <div className="kv"><span>charge</span><b className="mono">{live?.load_status ?? '—'}</b></div>
          </div>
        </div>
        <div className="card">
          <div className="card-head">Portes G0–G7</div>
          <div className="gates-detail" style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: '2px 16px' }}>
            {gates.map((g: boolean|null, i:number) => (
              <div key={i} className={'gate-row ' + (g===null?'na':g?'ok':'fail')}>
                <span className="gate mono">G{i}</span>
                <span className="gate-lbl">{GATE_LABELS[i]}</span>
                <span className="gate-state mono">{g===null?'—':g?'PASS':'FAIL'}</span>
              </div>
            ))}
          </div>
        </div>
        {profiles.length > 0 && (
        <div className="card" data-testid="matrix-table">
          <div className="card-head">Matrice — état par cellule</div>
          <table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:11 }}>
            <thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
              <th style={{ padding:'6px 8px' }}>profil</th><th>file × CC</th><th>répétitions</th><th>état</th>
            </tr></thead><tbody>
            {profiles.flatMap((p, pi) => ALL_QDISCS.flatMap((q, qi) => ALL_CC.map((c, ci) => {
              // ordre serveur (matrix.go) : profils × [pfifo, fq_codel, cake] × [cubic, bbr] × reps, event_id dès 1
              const cellStart = ((pi * ALL_QDISCS.length + qi) * ALL_CC.length + ci) * reps + 1
              const cur = live?.event_id ?? 0
              const st = cur === 0 ? 'attente' : (cellStart + reps - 1 < cur ? 'terminée' : (cellStart <= cur ? 'en cours' : 'attente'))
              const n = Math.max(0, Math.min(reps, cur - cellStart + (st === 'en cours' ? 1 : 0)))
              const color = st === 'terminée' ? '#1fa348' : st === 'en cours' ? '#5ad3e3' : '#767b84'
              return (<tr key={`${p}/${q}/${c}`} style={{ borderBottom:'1px solid var(--hairline-faint)' }}>
                <td style={{ padding:'6px 8px' }}>{p}</td><td>{q} × {c}</td>
                <td style={{ fontVariantNumeric:'tabular-nums' }}>{n}/{reps}</td>
                <td style={{ color }}>{st}</td></tr>)
            })))}
            </tbody></table>
          <div className="mono muted" style={{ fontSize:10, marginTop:6 }}>ordre serveur : profils × files × CC × répétitions · événement #{live?.event_id ?? '—'}/{live?.total_events ?? '—'}</div>
        </div>
        )}
      </div>
      {/* cockpit sheet — matrix + audit + import */}
      <aside className="panel-stack cockpit" style={{position:'fixed', right:0, top:48, bottom:28, width:380, zIndex:40, backdropFilter:'blur(12px)', WebkitBackdropFilter:'blur(12px)', borderLeft:'1px solid var(--hairline)', background:'rgba(7,7,7,0.85)', overflowY:'auto', padding:12}}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, color:'#8b9099', marginBottom:4}}>live small — 20 pts</div>
          <div style={{display:'flex', alignItems:'end', gap:1, height:28}}>
            {peek.data.map((v,i)=>(
              <i key={i} style={{flex:1, height: `${Math.max(2, Math.min(28, (v/200)*28))}px`, background: '#f4b400', borderRadius:1, opacity: 0.7 + (i/peek.data.length)*0.3}} />
            ))}
          </div>
          <div className="mono" style={{fontSize:10, color:'#f4b400', marginTop:4}}>{peek.data.at(-1)?.toFixed(1) ?? '—'} ms</div>
        </PeekPopover>
      )}
      <div className="card">
        <div className="card-head">Campagne — Kit cockpit</div>
        <div className="form-row">
          <label>Profils</label>
          <div className="check-row">
            {allProfiles.map(p => (
              <label key={p.id}><input type="checkbox" checked={profiles.includes(p.id)} onChange={e => setProfiles(e.target.checked ? [...profiles,p.id] : profiles.filter(x=>x!==p.id))} /> <Explain term={p.id}>{p.id}</Explain>{p.imported ? <span className="mono muted" title="issu d'un audit réel"> ↧</span> : null}</label>
            ))}
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
          <label><Explain term="deadline">Deadline</Explain></label>
          <input title="objectif small p95 — voyage avec la campagne (bornes serveur 200–5000 ms)" type="number" min={200} max={5000} step={100} placeholder="1000" value={deadlineMs} onChange={e=>setDeadlineMs(parseInt(e.target.value)||1000)} style={{width:80, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid #26262a', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11}} />
          <span className="mono muted" style={{fontFamily:'JetBrains Mono', fontSize:10}}>ms</span>
        </div>
        <div className="form-row" style={{gap:8, border:'1px solid #26262a', background:'rgba(244,180,0,0.06)', padding:'6px 8px', marginTop:6}} onMouseEnter={e=>setPeek({rect:e.currentTarget.getBoundingClientRect(), data: liveRing.small.slice(-20).map(([,v]:[number,number])=>v)})} onMouseLeave={()=>setPeek(null)}>
          <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.08em', textTransform:'uppercase', color:'#8b9099'}}>cost preview</span>
          <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, color: live?.wasted_bytes != null ? '#f4b400' : '#767b84', fontVariantNumeric:'tabular-nums'}}>
            {(() => {
              if (!live || live.wasted_bytes == null || live.cost_ar_per_h == null) return '—'
              return `${live.wasted_bytes} o gâchés · ${live.cost_ar_per_h.toFixed(2)} Ar/h`
            })()}
          </span>
          <span className="mono muted" style={{fontFamily:'JetBrains Mono', fontSize:10, marginLeft:'auto'}}>wasted × cost_per_h</span>
        </div>
        <div className="form-row" style={{gap:8}}>
          <span title={live?.running ? 'campagne en cours — arrêter d\'abord' : undefined}><ArmButton label="DÉMARRER" onConfirm={start} disabled={live?.running} /></span>
          <ArmButton label="ARRÊTER" confirmLabel="CONFIRMER L'ARRÊT" onConfirm={stop} disabled={!live?.running} />
          <span className="mono muted">{msg}</span>
        </div>
      </div>

      <div className="card">
        <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span><Explain term="audit">Audit du lien</Explain></span>
          <span className="mono muted" style={{ marginLeft: 'auto', fontSize: 10 }}>30 s, non intrusif</span>
          <button className="btn" onClick={() => setAuditOpen(o => !o)} style={{ padding: '2px 10px' }}>{auditOpen ? 'FERMER' : 'LANCER'}</button>
        </div>
        {auditOpen && <>
        <div ref={auditFormRef} style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <InlineField label="Site" error={auditValidation.errors.site} helper={!auditSite ? 'requis' : undefined}>
            <input name="site" value={auditSite} onChange={e=>setAuditSite(e.target.value)} style={{ flex:1, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
          </InlineField>
          <InlineField label="Type de lien" error={auditValidation.errors.link_type}>
            <select value={auditLink} onChange={e=>setAuditLink(e.target.value)} style={{ width:'100%', background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }}>
              {['fiber','5g','4g','vsat','other'].map(t=><option key={t} value={t}>{t}</option>)}
            </select>
          </InlineField>
          <InlineField label="Durée (s)" error={auditValidation.errors.duration} helper="10–600 s">
            <input type="number" min={10} max={600} value={auditDuration} onChange={e=>setAuditDuration(parseInt(e.target.value)||0)} style={{ width:80, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
          </InlineField>
        </div>
        <div className="form-row" style={{ gap:8, marginTop:8 }}>
          <ArmButton label="LANCER AUDIT" confirmLabel="CONFIRMER AUDIT" onConfirm={startAudit} disabled={!auditValidation.valid} />
          <span className="mono muted">{auditMsg}</span>
        </div>
        </>}
        {/* latence de travail — 3 cases, vides honnêtes : idle et montée
            chargée mesurées par l'audit ; la descente chargée attend la
            campagne download (le manque s'affiche, il ne se cache pas) */}
        {auditLast && (
        <div style={{ display:'grid', gridTemplateColumns:'repeat(3, minmax(0,1fr))', gap:8, marginTop:10 }} data-testid="work-latency">
          <div style={{ padding:'6px 8px', border:'1px solid var(--hairline)', background:'rgba(90,211,227,0.04)' }}>
            <div className="mono" style={{ fontSize:9, letterSpacing:'0.08em', textTransform:'uppercase', color:'#8b9099' }}>idle</div>
            <div className="mono" style={{ fontSize:16, color:'#f2f2f4', fontVariantNumeric:'tabular-nums' }}>{Number(auditLast.rtt_idle_p50_ms ?? 0).toFixed(1)} ms</div>
          </div>
          <div style={{ padding:'6px 8px', border:'1px solid var(--hairline)', background:'rgba(90,211,227,0.04)' }}>
            <div className="mono" style={{ fontSize:9, letterSpacing:'0.08em', textTransform:'uppercase', color:'#8b9099' }}>montée chargée</div>
            <div className="mono" style={{ fontSize:16, color:'#5ad3e3', fontVariantNumeric:'tabular-nums' }}>{Number(auditLast.rtt_loaded_p50_ms ?? 0).toFixed(1)} ms</div>
          </div>
          <div style={{ padding:'6px 8px', border:'1px dashed var(--hairline)' }} title="mesurable après la campagne download (bulk inversé)">
            <div className="mono" style={{ fontSize:9, letterSpacing:'0.08em', textTransform:'uppercase', color:'#767b84' }}>descente chargée</div>
            <div className="mono" style={{ fontSize:16, color:'#767b84' }}>— après campagne download</div>
          </div>
        </div>
        )}
        {auditLast?.bloat_grade && (
        <div className="mono" style={{ fontSize:11, marginTop:8, padding:'6px 8px', border:'1px solid #26262a', background:'rgba(244,180,0,0.06)' }} data-testid="bloat-grade">
          note bufferbloat : <b style={{ color: auditLast.bloat_grade.startsWith('A') ? '#1fa348' : auditLast.bloat_grade === 'B' ? '#5ad3e3' : '#f4b400' }}>{auditLast.bloat_grade}</b>
          {' '}({Number(auditLast.bloat_delta_ms ?? 0).toFixed(1)} ms sous charge) — {auditLast.bloat_verdict}
          <span style={{ color:'#767b84' }}> · bandes Waveform</span>
        </div>
        )}
      </div>

      <div className="card">
        <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span>Profil personnalisé</span>
          {!importForm && <>
            <button className="btn" onClick={async () => {
              try {
                const r = await fetch('/api/audit/toprofile', { method: 'POST' })
                const j = await r.json().catch(() => ({}))
                if (!r.ok) { useUIStore.getState().pushToast(j?.error ?? 'échec import audit', 'err'); return }
                useUIStore.getState().pushToast(`Profil ${j.profile?.id ?? 'P-audit'} importé depuis l'audit — ${j.profile?.capacity_mbps} Mbit/s, ${j.profile?.delay_ms} ms`, 'ok')
              } catch { useUIStore.getState().pushToast('échec import audit', 'err') }
            }} title="transforme le dernier audit du lien réel en profil rejouable sur le banc" style={{ marginLeft: 'auto', padding: '2px 10px' }}>AUDIT → PROFIL</button>
            <button className="btn" onClick={()=>setImportForm(true)} style={{ padding: '2px 10px' }}>IMPORTER</button>
            <span className="mono muted" style={{ fontSize: 10 }}>{importMsg}</span>
          </>}
        </div>
        {importForm && <>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 8 }}>
            <InlineField label="Identifiant" error={importValidation.errors.id} helper="ex. P3">
              <input name="import-id" value={impId} onChange={e=>setImpId(e.target.value)} placeholder="P3" style={{ background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="Capacité (Mbit/s)" error={importValidation.errors.capacity} helper="> 0">
              <input type="number" min={0.1} step={0.1} value={impCap} onChange={e=>setImpCap(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="Cap. montante (Mbit/s)" error={undefined} helper="0 = symétrique — ex. 5 pour 4G 20/5">
              <input type="number" min={0} step={0.1} value={impCapUp} onChange={e=>setImpCapUp(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="RTT (ms)" error={importValidation.errors.delay} helper="> 0">
              <input type="number" min={1} value={impDelay} onChange={e=>setImpDelay(parseInt(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="Gigue (ms)" error={importValidation.errors.jitter} helper="≥ 0 — ex. 30 pour VSAT">
              <input type="number" min={0} step={0.5} value={impJitter} onChange={e=>setImpJitter(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="Perte (%)" error={importValidation.errors.loss} helper="≥ 0 — ex. 1 pour VSAT">
              <input type="number" min={0} step={0.1} value={impLoss} onChange={e=>setImpLoss(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
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
