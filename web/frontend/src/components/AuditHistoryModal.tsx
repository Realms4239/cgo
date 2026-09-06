import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUIStore } from '../store/ui'
import { asArray } from '../lib/format'

// Historique des audits gelés — modale depuis la carte audit de Campagne.
// Chaque ligne porte son rapprochement au référentiel et s'importe en
// profil rejouable d'un clic (?audit_id= : pas de relance pour importer
// la fibre quand la pointe est la dernière).
export default function AuditHistoryModal({ onClose, onImported }: { onClose: () => void; onImported: (msg: string) => void }) {
  const [rows, setRows] = useState<Record<string,string>[]|null>(null)
  const load = () => {
    fetch('/api/audit/list').then(r=>r.json()).then(j=>setRows(asArray(j.audits))).catch(()=>setRows([]))
  }
  useEffect(()=>{ load() },[])
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  const toProfile = async (auditId: string) => {
    const store = useUIStore.getState()
    try {
      const r = await fetch('/api/audit/toprofile?audit_id='+encodeURIComponent(auditId), { method: 'POST' })
      const j = await r.json().catch(()=>({}))
      if (!r.ok) throw new Error(j?.error ?? String(r.status))
      store.pushToast(`Profil ${j.profile?.id} importé — ${j.profile?.capacity_mbps} Mbit/s, ${j.profile?.delay_ms} ms, perte ${j.profile?.loss_pct} %`, 'ok')
      onImported(`profil ${j.profile?.id} importé depuis ${auditId}`)
      onClose()
    } catch(e:any) { store.pushToast('Échec import: '+e.message, 'err') }
  }

  // portail racine : la section .view porte une animation transform qui
  // ferait de fixed un positionnement relatif à la carte (modale invisible
  // coincée dans la barre latérale — vu en prod). Même pattern que PeekPopover.
  return createPortal(
    <div role="dialog" aria-label="Historique des audits" style={{ position:'fixed', inset:0, zIndex:650, background:'rgba(0,0,0,0.55)', display:'flex', alignItems:'center', justifyContent:'center', padding:16 }} onClick={onClose}>
      <div style={{ maxWidth:860, width:'100%', maxHeight:'85vh', overflowY:'auto', background:'var(--surface-soft, #0b0b0c)', border:'1px solid var(--hairline)', padding:16 }} onClick={e=>e.stopPropagation()}>
        <div className="card-head" style={{ display:'flex', alignItems:'center', gap:8 }}>
          <span>Audits gelés ({rows?.length ?? '…'})</span>
          <button className="btn" style={{ marginLeft:'auto', padding:'2px 10px' }} onClick={load}>RÉACTUALISER</button>
          <button className="btn" style={{ padding:'2px 10px' }} onClick={onClose}>FERMER</button>
        </div>
        {rows === null && <div className="mono muted" style={{fontSize:11, marginTop:8}}>chargement…</div>}
        {rows !== null && rows.length === 0 && <div className="mono muted" style={{fontSize:11, marginTop:8}}>aucun audit gelé — le premier se lance depuis le protocole RQ1 ci-dessous.</div>}
        {rows !== null && rows.length > 0 && (
        <div style={{overflowX:'auto', marginTop:8}}>
        <table className="mono" style={{fontSize:11, borderCollapse:'collapse', width:'100%'}}>
          <thead><tr style={{textAlign:'left', color:'#8b9099'}}>
            <th>site</th><th>lien</th><th>idle p50/p95</th><th>chargé p50/p95</th><th>grade</th><th>rapprochement</th><th></th>
          </tr></thead>
          <tbody>
          {rows.map((a,i)=>(
            <tr key={i} style={{borderTop:'1px solid var(--hairline)'}}>
              <td>{a.site}<br/><span style={{color:'#767b84', fontSize:10}}>{(a.timestamp||'').slice(0,10)} {a.provider}</span></td>
              <td>{a.link_type}</td>
              <td>{a.rtt_idle_p50_ms}/{a.rtt_idle_p95_ms} ms</td>
              <td>{a.rtt_loaded_p50_ms}/{a.rtt_loaded_p95_ms} ms</td>
              <td><b style={{color:(a.bloat_grade||'').startsWith('A') ? '#1fa348' : '#f4b400'}}>{a.bloat_grade||'—'}</b></td>
              <td>{a.profile_match ? <span title={a.match_delta}><b>{a.profile_match}</b> <span style={{color:'#767b84', fontSize:10}}>{a.match_delta}</span></span> : <span style={{color:'#767b84'}}>—</span>}</td>
              <td><button className="btn" style={{padding:'2px 8px', fontSize:10}} title="importer CET audit comme profil rejouable" onClick={()=>toProfile(a.audit_id)}>→ PROFIL</button></td>
            </tr>
          ))}
          </tbody>
        </table>
        </div>
        )}
      </div>
    </div>,
    document.body
  )
}
