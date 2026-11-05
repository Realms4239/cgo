import { useEffect } from 'react'
import { createPortal } from 'react-dom'

// Mode d'emploi du classement — modale ⓘ depuis la barre benchmark.
// Tout ce qu'il faut pour lire une ligne sans ouvrir le mémoire : les 5
// métriques en une phrase chacune, les règles de lecture, le régime perte.
// Concis par construction : une ligne par concept, pas de paragraphes.
const ROWS: [string, string][] = [
  ['small p95', 'Latence des petits objets (télémétrie, alertes) — LE trafic critique à protéger.'],
  ['RTT p95', 'Aller-retour ping — la métrique du bufferbloat : elle augmente quand la file d\u2019attente déborde.'],
  ['goodput', 'Débit utile mesuré au récepteur pendant la charge.'],
  ['échéance', '% de petits objets sous la deadline — ne se compare qu\u2019à deadline fixée, les runs historiques la mélangent.'],
  ['gaspillé · coût', 'Octets retransmis → Ar/h au palier unique 5556 Ar/Go.'],
  ['n=', 'Réplications valides : seules les mesures valides comptent — les douteuses sont exclues, jamais mélangées.'],
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
    <div role="dialog" aria-label="Lire les résultats" className="modal-backdrop" onClick={onClose}>
      <div className="modal-panel modal-panel-sm" onClick={e => e.stopPropagation()}>
        <div className="card-head modal-head modal-head-spaced">
          <span>Lire le classement</span>
          <button className="btn modal-head-btn" onClick={onClose}>FERMER</button>
        </div>
        <div className="info-rows">
          {ROWS.map(([k, v]) => (
            <div key={k} className="info-row">
              <span className="mono info-row-key">{k}</span>
              <span className="info-row-val">{v}</span>
            </div>
          ))}
        </div>
        <div className="mono muted info-foot">Gel SHA-256 vérifiable : Provenance &amp; archives · Mémoire §3–4.</div>
      </div>
    </div>,
    document.body
  )
}
