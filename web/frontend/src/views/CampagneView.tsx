import { useState } from 'react'
import { ArmButton } from '../components/ArmButton'
import { useUIStore } from '../store/ui'

const ALL_QDISCS = ['pfifo_fast','fq_codel','cake'] as const
const ALL_CC = ['cubic','bbr'] as const

export default function CampagneView() {
  const live = useUIStore(s => s.live)
  const [profiles, setProfiles] = useState<string[]>(['P2'])
  const [reps, setReps] = useState(3)
  const [msg, setMsg] = useState('')

  const start = async () => {
    setMsg('démarrage…')
    const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles, reps }) })
    setMsg(r.ok ? 'campagne lancée' : 'échec: ' + r.status)
  }
  const stop = async () => {
    await fetch('/api/run/stop', { method: 'POST' })
    setMsg('arrêt demandé')
  }

  const gates = live?.gates ?? Array(8).fill(null)
  const phase = live?.phase ?? 'idle'

  return (
    <div className="panel-stack">
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
    </div>
  )
}
function gateLabel(i:number){ return ['cible joignable','bulk démarré','sondes actives','latence plausible','débit cohérent','pas de doublon','baseline stable','CPU ok'][i] }
