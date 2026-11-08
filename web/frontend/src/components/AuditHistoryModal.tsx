import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useModalFocus } from '../lib/modalFocus'
import { useUIStore } from '../store/ui'
import { asArray } from '../lib/format'
import { paginate, Paginate } from './ui/Paginate'

// Historique des audits archivés — modale depuis la carte audit de Campagne.
// Chaque ligne porte son rapprochement au référentiel et s'importe en
// profil rejouable d'un clic (?audit_id= : pas de relance pour importer
// la fibre quand la pointe est la dernière).
export default function AuditHistoryModal({ onClose, onImported }: { onClose: () => void; onImported: (msg: string) => void }) {
  const [rows, setRows] = useState<Record<string,string>[]|null>(null)
  // pagination (U6h) — la table peut croître avec les audits gelés
  const [pg, setPg] = useState(1)
  const PG_SIZE = 10
  const load = () => {
    fetch('/api/audit/list').then(r=>r.json()).then(j=>setRows(asArray(j.audits))).catch(()=>setRows([]))
  }
  useEffect(()=>{ load() },[])
  const panelRef = useRef<HTMLDivElement>(null)
  useModalFocus(true, onClose, panelRef)

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
    <div role="dialog" aria-modal="true" aria-label="Historique des audits" className="modal-backdrop" onClick={onClose}>
      <div ref={panelRef} className="modal-panel" onClick={e=>e.stopPropagation()}>
        <div className="card-head modal-head">
          <span>Audits archivés ({rows?.length ?? '…'})</span>
          <button className="btn modal-head-btn" onClick={load}>RÉACTUALISER</button>
          <button className="btn modal-head-btn-plain" onClick={onClose}>FERMER</button>
        </div>
        {rows === null && <div className="mono muted modal-loading">chargement…</div>}
        {rows !== null && rows.length === 0 && <div className="mono muted modal-loading">aucun audit archivé — lancez le premier depuis Auditer.</div>}
        {rows !== null && rows.length > 0 && (
        <div className="modal-table-wrap">
        <table className="mono modal-table">
          <thead><tr>
            <th>site</th><th>lien</th><th>idle p50/p95</th><th>chargé p50/p95</th><th>grade</th><th>rapprochement</th><th></th>
          </tr></thead>
          <tbody>
          {rows !== null && paginate(rows, pg, PG_SIZE).map((a,i)=>(
            <tr key={i}>
              <td>{a.site}<br/><span className="modal-cell-faint">{(a.timestamp||'').slice(0,10)} {a.provider}</span></td>
              <td>{a.link_type}</td>
              <td>{a.rtt_idle_p50_ms}/{a.rtt_idle_p95_ms} ms</td>
              <td>{a.rtt_loaded_p50_ms}/{a.rtt_loaded_p95_ms} ms</td>
              <td><b style={{color:(a.bloat_grade||'').startsWith('A') ? 'var(--t-ok)' : 'var(--t-threshold)'}}>{a.bloat_grade||'—'}</b></td>
              <td>{a.profile_match ? <span title={a.match_delta}><b>{a.profile_match}</b> <span className="modal-cell-faint">{a.match_delta}</span></span> : <span className="modal-cell-dash">—</span>}</td>
              <td><button className="btn modal-btn-sm" title="importer CET audit comme profil rejouable" onClick={()=>toProfile(a.audit_id)}>→ PROFIL</button></td>
            </tr>
          ))}
          </tbody>
        </table>
        </div>
        )}
        {rows !== null && rows.length > 0 && <Paginate total={rows.length} page={pg} pageSize={PG_SIZE} onPage={setPg} />}
      </div>
    </div>,
    document.body
  )
}
