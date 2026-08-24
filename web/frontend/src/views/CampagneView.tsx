import { useState } from 'react'
import { ArmButton } from '../components/ArmButton'
import { Timeline } from '../components/Timeline'
import { useUIStore } from '../store/ui'

const ALL_QDISCS = ['pfifo_fast','fq_codel','cake'] as const
const ALL_CC = ['cubic','bbr'] as const

export default function CampagneView() {
  const live = useUIStore(s => s.live)
  const [profiles, setProfiles] = useState<string[]>(['P2'])
  const [reps, setReps] = useState(3)
  const [deadlineMs, setDeadlineMs] = useState(1000)
  const [msg, setMsg] = useState('')
  const [auditMsg, setAuditMsg] = useState('')
  const [auditSite, setAuditSite] = useState('Département X')
  const [auditLink, setAuditLink] = useState('5g')
  const [importMsg, setImportMsg] = useState('')

  const startAudit = () => {
    setAuditMsg('audit en cours…')
    fetch('/api/audit/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ site: auditSite, link_type: auditLink, duration: 30 }) })
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(() => {
        const poll = setInterval(() => {
          fetch('/api/audit/status').then(r=>r.json()).then(j=>{
            if (!j.running) {
              clearInterval(poll)
              setAuditMsg(j.last ? `audit terminé — p95 ${Number(j.last.rtt_idle_p95_ms).toFixed(1)} ms` : 'audit terminé')
            }
          }).catch(()=>clearInterval(poll))
        }, 2000)
      })
      .catch(e => setAuditMsg('échec: ' + e.message))
  }

  const importProfile = () => {
    const id = prompt('Identifiant du profil (ex. P3):')
    if (!id) return
    const cap = parseFloat(prompt('Capacité Mbit/s:', '20') || '20')
    const delay = parseFloat(prompt('RTT ms:', '100') || '100')
    fetch('/api/profile/import', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id, capacity_mbps: cap, delay_ms: delay }) })
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(() => setImportMsg(`profil ${id} importé`))
      .catch(e => setImportMsg('échec: ' + e.message))
  }

  const pushToast = useUIStore(s=>s.pushToast)
  const start = async () => {
    setMsg('démarrage…')
    const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles, reps }) })
    const ok = r.ok
    setMsg(ok ? 'campagne lancée' : 'échec: ' + r.status)
    pushToast(ok ? 'Campagne lancée' : 'Échec démarrage', ok ? 'ok' : 'err')
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
    <div className="panel-stack">
      <Timeline baselineStart={timeline.baselineStart} chargeStart={timeline.chargeStart} chargeEnd={timeline.chargeEnd} recupEnd={timeline.recupEnd} currentPhase={phase} />
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
          <input type="number" min={100} max={5000} step={100} value={deadlineMs} onChange={e=>setDeadlineMs(parseInt(e.target.value)||1000)} style={{width:80, background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid #26262a', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11}} />
          <span className="mono muted" style={{fontFamily:'JetBrains Mono', fontSize:10}}>ms · seuil small_p95</span>
        </div>
        {/* ponytail: cost forecast linear, non-linear if thesis needs */}
        <div className="form-row" style={{gap:8, border:'1px solid #26262a', background:'rgba(244,180,0,0.06)', padding:'6px 8px', marginTop:6}}>
          <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:10, letterSpacing:'0.08em', textTransform:'uppercase', color:'#8b9099'}}>cost preview</span>
          <span className="mono" style={{fontFamily:'JetBrains Mono', fontSize:11, color:'#f4b400', fontVariantNumeric:'tabular-nums'}}>
            {(() => {
              const wasted = (live?.drops ?? 0) * 1448
              const cost = wasted / (4.5*1024*1024*1024) * 30000
              return `${wasted} o gâchés · ${cost.toFixed(2)} Ar/h`
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
        <div className="kv"><span>événement</span><b className="mono">{live ? `#${live.event_id} · rép ${live.repetition}` : '—'}</b></div>
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
        <div className="form-row">
          <label>Site</label>
          <input value={auditSite} onChange={e=>setAuditSite(e.target.value)} style={{flex:1}} />
        </div>
        <div className="form-row">
          <label>Type de lien</label>
          <select value={auditLink} onChange={e=>setAuditLink(e.target.value)} style={{background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)'}}>
            {['fiber','5g','4g','vsat','other'].map(t=><option key={t} value={t}>{t}</option>)}
          </select>
        </div>
        <div className="form-row" style={{gap:8}}>
          <ArmButton label="LANCER AUDIT" confirmLabel="CONFIRMER AUDIT" onConfirm={startAudit} />
          <span className="mono muted">{auditMsg}</span>
        </div>
      </div>

      <div className="card">
        <div className="card-head">Profil personnalisé</div>
        <p className="mono muted" style={{fontSize:12}}>Importer un profil réel mesuré par l'audit (LIEN III.IV — Import profile).</p>
        <div className="form-row" style={{gap:8}}>
          <ArmButton label="IMPORTER PROFIL" confirmLabel="CONFIRMER IMPORT" onConfirm={importProfile} />
          <span className="mono muted">{importMsg}</span>
        </div>
      </div>
    </div>
  )
}
function gateLabel(i:number){ return ['cible joignable','bulk démarré','sondes actives','latence plausible','débit cohérent','pas de doublon','baseline stable','CPU ok'][i] }
