import { useEffect, useState } from 'react'
import { CRAFT } from '../lib/chartGrammar'

// Interprétation d'un profil — pensée pour quelqu'un qui ne connaît PAS le
// vocabulaire du banc : verdict en français simple d'abord, chiffres avec
// leur nom complet + ce que ça veut dire, prescription copiable, étapes
// numérotées. Tout vient du gel : GET /api/hardware/translate calcule, ce
// composant rend. Zéro invention.
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
  const [copied, setCopied] = useState(false)
  useEffect(() => {
    setData(null)
    fetch(`/api/hardware/translate?profile=${encodeURIComponent(profile)}`)
      .then(r => r.json())
      .then(j => setData(j))
      .catch(() => setData({ available: false, reason: 'injoignable — réessayez' }))
  }, [profile])
  const metric = (label: string, hint: string, v: string, color = '#d6d8dd') => (
    <div style={{ border: '1px solid var(--hairline)', padding: '12px 14px' }}>
      <div style={{ fontSize: 13, color: '#a9aeb6', marginBottom: 6 }}>{label}</div>
      <div className="mono" style={{ fontSize: 20, fontWeight: 600, color, fontVariantNumeric: 'tabular-nums', lineHeight: 1 }}>{v}</div>
      <div style={{ fontSize: 12, color: '#8b9099', marginTop: 6, lineHeight: 1.5 }}>{hint}</div>
    </div>
  )
  return (
    <div className="card" data-testid="interpretation" style={{ gridColumn: '1 / -1', border: '1px solid ' + CRAFT.live, background: 'var(--surface-card)', padding: 0 }}>
      <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', borderBottom: '1px solid var(--hairline)' }}>
        <span style={{ color: CRAFT.live, fontSize: 14 }}>Interprétation — {profile}</span>
        <span className="mono" style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--text-muted)' }}>chiffres gelés, vérifiables</span>
        <button className="btn" onClick={onClose} style={{ padding: '4px 12px' }}>FERMER</button>
      </div>
      <div style={{ padding: 16, display: 'grid', gap: 14 }}>
        {!data && <span className="mono muted" style={{ fontSize: 13 }}>calcul…</span>}
        {data && !data.available && <span className="mono muted" style={{ fontSize: 13 }}>{data.reason ?? 'aucune donnée gelée pour ce profil'}</span>}
        {data?.available && (
          <>
            {/* verdict d'abord, en français, gros et lisible */}
            <div className="mono" data-testid="interp-verdict" style={{ fontSize: 15, color: '#f2f2f4', lineHeight: 1.65, maxWidth: '70ch' }}>
              {data.verdict}
              {data.throughput ? <span style={{ color: '#a9aeb6' }}> · {data.throughput}</span> : null}
            </div>
            {data.cell && (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 10 }}>
                {metric('Réactivité du trafic critique', 'temps de chargement d\'une sonde pendant un transfert lourd — plus bas = mieux',
                  `${data.cell.small_p95.toFixed(0)} ms`, data.cell.small_p95 < 1000 ? CRAFT.ok : '#f4b400')}
                {metric('Latence du lien', 'aller-retour d\'un paquet sous charge — à combien le lien "répond"',
                  `${data.cell.rtt_p95.toFixed(0)} ms`, data.cell.rtt_p95 < 300 ? CRAFT.ok : '#f4b400')}
                {metric('Débit utile du transfert lourd', 'ce qui passe réellement une fois la file gérée',
                  `${data.cell.goodput.toFixed(1)} Mb/s`, '#b48ae0')}
                {metric('Échéances respectées', 'part des sondes livrées dans le délai cible',
                  `${data.cell.deadline_ok.toFixed(0)} %`, data.cell.deadline_ok >= 95 ? CRAFT.ok : '#f4b400')}
              </div>
            )}
            {data.recommendation && (
              <div style={{ border: '1px solid rgba(244,180,0,0.45)', background: 'rgba(244,180,0,0.05)', padding: '12px 14px' }}>
                <div style={{ fontSize: 13, color: '#f4b400', marginBottom: 6, letterSpacing: '0.04em' }}>Recommandation terrain</div>
                <div style={{ fontSize: 13, color: '#d6d8dd', lineHeight: 1.6 }}>{data.recommendation}</div>
              </div>
            )}
            {data.cli && (
              <div>
                <div style={{ fontSize: 12, color: '#a9aeb6', marginBottom: 6 }}>Commande prête à appliquer sur la passerelle :</div>
                <div style={{ display: 'flex', gap: 8, alignItems: 'stretch' }}>
                  <code className="mono" style={{ flex: 1, fontSize: 12, background: 'rgba(255,255,255,0.03)', border: '1px solid var(--hairline)', padding: '10px 12px', color: CRAFT.live, whiteSpace: 'pre', overflowX: 'auto' }}>{data.cli}</code>
                  <button className="btn" onClick={() => { navigator.clipboard?.writeText(data.cli ?? ''); setCopied(true); setTimeout(() => setCopied(false), 1500) }} style={{ padding: '6px 14px', whiteSpace: 'nowrap' }}>{copied ? 'COPIÉ ✓' : 'COPIER'}</button>
                </div>
              </div>
            )}
            {data.next_steps && (
              <div>
                <div style={{ fontSize: 13, color: '#a9aeb6', marginBottom: 8 }}>Prochaines étapes</div>
                <ol style={{ paddingLeft: 20, display: 'grid', gap: 8, margin: 0 }}>
                  {data.next_steps.map((s, i) => <li key={i} style={{ fontSize: 13, color: '#d6d8dd', lineHeight: 1.6 }}>{s}</li>)}
                </ol>
              </div>
            )}
            <div style={{ fontSize: 12, color: '#8b9099', lineHeight: 1.6, borderTop: '1px solid var(--hairline-faint)', paddingTop: 10 }}>
              Contexte : liens Yas 4G/5G et fibre de Météo Madagascar — le trafic critique (télémétrie, alertes, tableaux de bord) partage le lien avec les gros transferts (imagerie satellite, modèles météo). C'est ce partage que la mesure évalue.
            </div>
          </>
        )}
      </div>
    </div>
  )
}
