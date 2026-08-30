import { useEffect, useState } from 'react'
import { CRAFT } from '../lib/chartGrammar'

// Interprétation riche d'une cellule — verdict, chiffres clés, prescription
// réseau prête à copier, prochaines étapes opérateur. Tout vient du gel :
// GET /api/hardware/translate calcule, ce composant rend. Zéro invention.
type Interp = {
  available: boolean
  reason?: string
  recommendation?: string
  verdict?: string
  throughput?: string
  cli?: string
  next_steps?: string[]
  cell?: { qdisc: string; cc: string; small_p95: number; rtt_p95: number; goodput: number; deadline_ok: number }
}

export default function InterpretationView({ profile, onClose }: { profile: string; onClose: () => void }) {
  const [data, setData] = useState<Interp | null>(null)
  useEffect(() => {
    fetch(`/api/hardware/translate?profile=${encodeURIComponent(profile)}`)
      .then(r => r.json())
      .then(j => setData(j))
      .catch(() => setData({ available: false, reason: 'réseau' }))
  }, [profile])
  return (
    <div className="card" data-testid="interpretation" style={{ gridColumn: '1 / -1', border: '1px solid ' + CRAFT.live, background: 'var(--surface-card)', padding: 0 }}>
      <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '10px 14px', borderBottom: '1px solid var(--hairline)' }}>
        <span style={{ color: CRAFT.live }}>Interprétation — profil {profile}</span>
        <span className="mono" style={{ marginLeft: 'auto', fontSize: 9, color: 'var(--text-faint)' }}>calculée depuis les CSV gelés</span>
        <button className="btn" onClick={onClose} style={{ padding: '2px 10px' }}>FERMER</button>
      </div>
      <div style={{ padding: 14, display: 'grid', gap: 12 }}>
        {!data && <span className="mono muted">calcul…</span>}
        {data && !data.available && <span className="mono muted">{data.reason ?? 'aucune donnée'}</span>}
        {data?.available && (
          <>
            <div className="mono" data-testid="interp-verdict" style={{ fontSize: 12, color: '#d6dade', lineHeight: 1.6 }}>
              {data.verdict}
              {data.throughput ? <span style={{ color: 'var(--text-muted)' }}> · {data.throughput}</span> : null}
            </div>
            {data.cell && (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: 8 }}>
                {[
                  { l: 'small p95', v: `${data.cell.small_p95.toFixed(1)} ms` },
                  { l: 'rtt p95', v: `${data.cell.rtt_p95.toFixed(1)} ms` },
                  { l: 'goodput', v: `${data.cell.goodput.toFixed(1)} Mbit/s` },
                  { l: 'deadline ok', v: `${data.cell.deadline_ok.toFixed(0)}%` },
                ].map(x => (
                  <div key={x.l} style={{ border: '1px solid var(--hairline)', padding: '8px 10px' }}>
                    <div className="mono" style={{ fontSize: 9, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{x.l}</div>
                    <div className="mono" style={{ fontSize: 15, fontWeight: 700, color: CRAFT.ok, fontVariantNumeric: 'tabular-nums' }}>{x.v}</div>
                  </div>
                ))}
              </div>
            )}
            <div className="mono" style={{ fontSize: 10, color: 'var(--t-warn, #f4b400)', lineHeight: 1.6, border: '1px dashed rgba(244,180,0,0.4)', padding: '6px 10px' }}>
              {data.recommendation}
            </div>
            {data.cli && (
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <code className="mono" style={{ flex: 1, fontSize: 11, background: 'rgba(255,255,255,0.03)', border: '1px solid var(--hairline)', padding: '8px 10px', color: CRAFT.live }}>{data.cli}</code>
                <button className="btn" onClick={() => navigator.clipboard?.writeText(data.cli ?? '')} style={{ padding: '6px 10px', whiteSpace: 'nowrap' }}>COPIER</button>
              </div>
            )}
            {data.next_steps && (
              <ol style={{ paddingLeft: 18, display: 'grid', gap: 4 }}>
                {data.next_steps.map((s, i) => <li key={i} className="mono" style={{ fontSize: 11, color: 'var(--text-body)' }}>{s}</li>)}
              </ol>
            )}
            <div className="mono" style={{ fontSize: 10, color: 'var(--text-faint)' }}>
              Contexte : liens Yas 4G/5G et fibre Météo Madagascar — le trafic critique (télémétrie AWS, tableaux de bord, alertes) partage le lien avec les transferts de masse (imagerie satellite, modèles).
            </div>
          </>
        )}
      </div>
    </div>
  )
}
