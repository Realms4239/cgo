import { useEffect, useRef, useState } from 'react'
import { ArmButton } from '../components/ArmButton'
import { Timeline } from '../components/Timeline'
import { useUIStore } from '../store/ui'
import { InlineField } from '../components/InlineField'
import { validate } from '../lib/validation'
import { animateShake } from '../lib/anime'
import { PeekPopover } from '../components/PeekPopover'
import { live as liveRing } from '../lib/live'

const ALL_QDISCS = ['pfifo_fast','fq_codel','cake'] as const
const ALL_CC = ['cubic','bbr'] as const

export default function CampagneView() {
  const live = useUIStore((s: any) => s.live)
  const [profiles, setProfiles] = useState<string[]>(['P2'])
  const [reps, setReps] = useState(3)
  const [deadlineMs, setDeadlineMs] = useState(1000)
  const [msg, setMsg] = useState('')
  const [auditMsg, setAuditMsg] = useState('')
  const [auditSite, setAuditSite] = useState('Département X')
  const [auditLink, setAuditLink] = useState('5g')
  const [auditDuration, setAuditDuration] = useState(30)
  const [importMsg, setImportMsg] = useState('')
  const [importForm, setImportForm] = useState(false)
  const [impId, setImpId] = useState('')
  const [impCap, setImpCap] = useState(20)
  const [impDelay, setImpDelay] = useState(100)
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
    { id: impId, capacity: impCap, delay: impDelay },
    { id: { required: true }, capacity: { min: 0.1 }, delay: { min: 1 } }
  )

  const importProfile = () => {
    if (!importValidation.valid) return
    fetch('/api/profile/import', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: impId, capacity_mbps: impCap, delay_ms: impDelay }) })
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(() => { setImportMsg(`profil ${impId} importé`); useUIStore.getState().pushToast(`Profil ${impId} importé`, 'ok'); setImportForm(false) })
      .catch(e => { setImportMsg('échec: ' + e.message); useUIStore.getState().pushToast('Échec import: ' + e.message, 'err') })
  }

  const pushToast = useUIStore((s: any)=>s.pushToast)
  const start = async () => {
    setMsg('démarrage…')
    const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles, reps }) })
    const ok = r.ok
    setMsg(ok ? 'campagne lancée' : 'échec: ' + r.status)
    pushToast(ok ? 'Campagne lancée' : 'Échec démarrage', ok ? 'ok' : 'err')
    if (ok) {
      const st = useUIStore.getState()
      st.setPanel('live')
      // ponytail: flash handled by App wasRunning redirect — avoid duplicate
    }
  }
  const stop = async () => {
    await fetch('/api/run/stop', { method: 'POST' })
    setMsg('arrêt demandé')
    pushToast('Arrêt demandé', 'blue')
  }

  const gates = live?.gates ?? Array(8).fill(null)
  const phase = live?.phase ?? 'idle'
  const [timeline] = useState(() => {
    const n = Date.now()
    return { baselineStart: n - 90000, chargeStart: n - 60000, chargeEnd: n - 15000, recupEnd: n + 15000 }
  })

  return (
    <div className="panel-stack" style={{position:'relative'}}>
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
      {phase !== 'idle' && <Timeline baselineStart={timeline.baselineStart} chargeStart={timeline.chargeStart} chargeEnd={timeline.chargeEnd} recupEnd={timeline.recupEnd} currentPhase={phase} />}
      <div className="card">
        <div className="card-head">Campagne</div>
        <div className="form-row">
          <label>Profils</label>
          <div className="check-row">
            {(['P1','P2'] as const).map(p => (
              <label key={p}><input type="checkbox" checked={profiles.includes(p)} onChange={e => setProfiles(e.target.checked ? [...profiles,p] : profiles.filter(x=>x!==p))} /> {p}</label>
            ))}
          </div>
        </div>
        <div className="form-row">
          <label>Qdiscs</label><span className="mono muted">{ALL_QDISCS.join(' · ')}</span>
        </div>
        <div className="form-row">
          <label>CC</label><span className="mono muted">{ALL_CC.join(' · ')}</span>
        </div>
        <div className="form-row">
          <label>Répétitions</label>
          <input type="number" min={1} max={5} value={reps} onChange={e=>setReps(parseInt(e.target.value)||1)} style={{width:64}} />
        </div>
        <div className="form-row">
          <label>Deadline</label>
          <input disabled title="paramètre fixe 1000 ms côté engine (thread campagne.go:226) — contrôle retiré" type="number" min={100} max={5000} step={100} value={deadlineMs} onChange={e=>setDeadlineMs(parseInt(e.target.value)||1000)} style={{width:80, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid #26262a', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11, opacity:0.5}} />
          <span className="mono muted" style={{fontFamily:'JetBrains Mono', fontSize:10}}>ms · seuil small_p95 (fixe)</span>
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
          <ArmButton label="DÉMARRER" onConfirm={start} disabled={live?.running} />
          <ArmButton label="ARRÊTER" confirmLabel="CONFIRMER L'ARRÊT" onConfirm={stop} disabled={!live?.running} />
          <span className="mono muted">{msg}</span>
        </div>
      </div>

      <div className="card">
        <div className="card-head">État</div>
        <div className="kv"><span>phase</span><b className="mono">{phase}</b></div>
        <div className="kv"><span>profil</span><b className="mono">{live?.profile ?? '—'}</b></div>
        <div className="kv"><span>qdisc</span><b className="mono">{live?.qdisc ?? '—'}</b></div>
        <div className="kv"><span>cc</span><b className="mono">{live?.cc ?? '—'}</b></div>
        <div className="kv"><span>événement</span><b className="mono">{live?.event_id != null ? `#${live.event_id} · rép ${live.repetition ?? 0}` : '—'}</b></div>
        <div className="kv"><span>charge</span><b className="mono">{live?.load_status ?? '—'}</b></div>
      </div>

      <div className="card">
        <div className="card-head">Portes G0–G7</div>
        <div className="gates-detail">
          {gates.map((g: boolean|null, i:number) => (
            <div key={i} className={'gate-row ' + (g===null?'na':g?'ok':'fail')}>
              <span className="gate mono">G{i}</span>
              <span className="gate-lbl">{gateLabel(i)}</span>
              <span className="gate-state mono">{g===null?'—':g?'PASS':'FAIL'}</span>
            </div>
          ))}
        </div>
      </div>

      <div className="card">
        <div className="card-head">Audit lien accessible (non intrusif)</div>
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
      </div>

      <div className="card">
        <div className="card-head">Profil personnalisé</div>
        <p className="mono muted" style={{fontSize:12}}>Importer un profil réel mesuré par l'audit (LIEN III.IV — Import profile).</p>
        {!importForm ? (
          <div className="form-row" style={{gap:8}}>
            <button className="btn" onClick={()=>setImportForm(true)}>IMPORTER PROFIL</button>
            <span className="mono muted">{importMsg}</span>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 8 }}>
            <InlineField label="Identifiant" error={importValidation.errors.id} helper="ex. P3">
              <input name="import-id" value={impId} onChange={e=>setImpId(e.target.value)} placeholder="P3" style={{ background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="Capacité (Mbit/s)" error={importValidation.errors.capacity} helper="> 0">
              <input type="number" min={0.1} step={0.1} value={impCap} onChange={e=>setImpCap(parseFloat(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <InlineField label="RTT (ms)" error={importValidation.errors.delay} helper="> 0">
              <input type="number" min={1} value={impDelay} onChange={e=>setImpDelay(parseInt(e.target.value)||0)} style={{ width:100, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }} />
            </InlineField>
            <div className="form-row" style={{gap:8}}>
              <ArmButton label="CONFIRMER IMPORT" onConfirm={importProfile} disabled={!importValidation.valid} />
              <button className="btn" onClick={()=>setImportForm(false)}>ANNULER</button>
              <span className="mono muted">{importMsg}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
function gateLabel(i:number){ return ['cible joignable','bulk démarré','sondes actives','latence plausible','débit cohérent','pas de doublon','baseline stable','CPU ok'][i] }
