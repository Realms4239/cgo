import { useEffect } from 'react'
import { createPortal } from 'react-dom'

// Mode d'emploi du classement — modale ⓘ depuis la barre benchmark.
// Tout ce qu'il faut pour lire une ligne sans ouvrir le mémoire : les 5
// métriques en une phrase chacune, les règles de lecture, le régime perte.
// Concis par construction : une ligne par concept, pas de paragraphes.
const ROWS: [string, string][] = [
  ['small p95', 'Latence des petits objets (télémétrie, alertes) — LE trafic critique à protéger.'],
  ['RTT p95', 'Aller-retour ping — la métrique du bufferbloat : elle gonfle quand la file déborde.'],
  ['goodput', 'Débit utile mesuré au récepteur pendant la charge.'],
  ['échéance', '% de petits objets sous la deadline — ne se compare qu\u2019à deadline fixée, les runs historiques la mélangent.'],
  ['gaspillé · coût', 'Octets retransmis → Ar/h au palier unique 5556 Ar/Go.'],
  ['n=', 'Réplications valides (valid-only strict : degraded G2/G6 et quarantaine exclus, jamais moyennés).'],
  ['± IC95', 'Intervalle bootstrap : deux barres qui se chevauchent = égalité, pas de hiérarchie.'],
  ['indice LIEN', 'Note /100 qui combine les 5 métriques (TradeSpace) : une seule note par lien.'],
  ['régime perte', 'Sur VSAT (600 ms, 1 % perte), une retransmission coûte un aller simple (+640 ms) : la perte gouverne la sonde, pas la file. Échéance médiane 0 partout = normal, pas un échec.'],
];

export default function ResultatsInfoModal({ onClose }: { onClose: () => void }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])
  return createPortal(
    <div role="dialog" aria-label="Lire les résultats" style={{ position: 'fixed', inset: 0, zIndex: 650, background: 'rgba(0,0,0,0.55)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }} onClick={onClose}>
      <div style={{ maxWidth: 620, width: '100%', maxHeight: '85vh', overflowY: 'auto', background: 'var(--surface-soft, #0b0b0c)', border: '1px solid var(--hairline)', padding: '16px 18px' }} onClick={e => e.stopPropagation()}>
        <div className="card-head" style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
          <span>Lire le classement</span>
          <button className="btn" style={{ marginLeft: 'auto', padding: '2px 10px' }} onClick={onClose}>FERMER</button>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          {ROWS.map(([k, v]) => (
            <div key={k} style={{ display: 'flex', gap: 10, padding: '6px 0', borderTop: '1px solid var(--hairline-faint)', fontSize: 12, lineHeight: 1.55 }}>
              <span className="mono" style={{ minWidth: 92, color: '#7fd6e8' }}>{k}</span>
              <span style={{ color: '#c3c9d1' }}>{v}</span>
            </div>
          ))}
        </div>
        <div className="mono muted" style={{ fontSize: 10, marginTop: 10 }}>Gel SHA-256 vérifiable : Provenance &amp; archives · Mémoire §3–4.</div>
      </div>
    </div>,
    document.body
  )
}
