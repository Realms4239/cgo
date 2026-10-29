import { useEffect, useState } from 'react'

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
  // Échap referme — comme la modale ⓘ, le duel et le tiroir Réglages
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])
  const metric = (label: string, hint: string, v: string, color = 'var(--text-body-strong)') => (
    <div className="interp-metric">
      <div className="interp-metric-label">{label}</div>
      <div className="mono interp-metric-value" style={{ color }}>{v}</div>
      <div className="interp-metric-hint">{hint}</div>
    </div>
  )
  return (
    <div className="card interp-card" data-testid="interpretation">
      <div className="card-head interp-head">
        <span className="interp-head-title">Interprétation — {profile}</span>
        <span className="mono interp-head-note">chiffres archivés, vérifiables</span>
        <button className="btn interp-close" onClick={onClose}>FERMER</button>
      </div>
      <div className="interp-body">
        {!data && <span className="mono muted interp-wait">calcul…</span>}
        {data && !data.available && <span className="mono muted interp-wait">{data.reason ?? 'aucune donnée archivée pour ce profil'}</span>}
        {data?.available && (
          <>
            {/* verdict d'abord, en français, gros et lisible */}
            <div className="mono interp-verdict" data-testid="interp-verdict">
              {data.verdict}
              {data.throughput ? <span className="interp-verdict-sub"> · {data.throughput}</span> : null}
            </div>
            {data.cell && (
              <div className="interp-metrics">
                {metric('Réactivité du trafic critique', "temps de chargement d'une sonde pendant un transfert lourd — plus bas = mieux",
                  `${data.cell.small_p95.toFixed(0)} ms`, data.cell.small_p95 < 1000 ? 'var(--t-ok)' : 'var(--t-threshold)')}
                {metric('Latence du lien', "aller-retour d'un paquet sous charge — à combien le lien \"répond\"",
                  `${data.cell.rtt_p95.toFixed(0)} ms`, data.cell.rtt_p95 < 300 ? 'var(--t-ok)' : 'var(--t-threshold)')}
                {metric('Débit utile du transfert lourd', 'ce qui passe réellement une fois la file gérée',
                  `${data.cell.goodput.toFixed(1)} Mb/s`, 'var(--t-bbr)')}
                {metric('Échéances respectées', 'part des sondes livrées dans le délai cible',
                  `${data.cell.deadline_ok.toFixed(0)} %`, data.cell.deadline_ok >= 95 ? 'var(--t-ok)' : 'var(--t-threshold)')}
              </div>
            )}
            {data.recommendation && (
              <div className="interp-reco">
                <div className="interp-reco-title">Recommandation terrain</div>
                <div className="interp-reco-body">{data.recommendation}</div>
              </div>
            )}
            {data.cli && (
              <div>
                <div className="interp-cli-label">Commande prête à appliquer sur la passerelle :</div>
                <div className="interp-cli-row">
                  <code className="mono interp-cli">{data.cli}</code>
                  <button className="btn interp-copy" onClick={() => { navigator.clipboard?.writeText(data.cli ?? ''); setCopied(true); setTimeout(() => setCopied(false), 1500) }}>{copied ? 'COPIÉ ✓' : 'COPIER'}</button>
                </div>
              </div>
            )}
            {data.next_steps && (
              <div>
                <div className="interp-steps-title">Prochaines étapes</div>
                <ol className="interp-steps">
                  {data.next_steps.map((s, i) => <li key={i} className="interp-step">{s}</li>)}
                </ol>
              </div>
            )}
            <div className="interp-context">
              Contexte : liens Yas 4G/5G et fibre de Météo Madagascar — le trafic critique (télémétrie, alertes, tableaux de bord) partage le lien avec les gros transferts (imagerie satellite, modèles météo). C'est ce partage que la mesure évalue.
            </div>
          </>
        )}
      </div>
    </div>
  )
}
