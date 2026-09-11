// LiveGuide — guide de lecture des métriques du mur (U6a/U6i): chaque
// chiffre dit ce qu'un HAUT/MOYEN/BAS signifie pour le trafic, seuils
// branchés sur les Réglages réels. Matériel à lire, pas de la déco.
import { useEffect, useState } from 'react'
import Explain from './Explain'
import { loadSettings, type Settings } from '../lib/settings'

type Entry = { term: string; label: string; haut: string; moyen: string; bas: string; impact: string }

function entries(s: Settings): Entry[] {
  return [
    {
      term: 'small_p95', label: 'Petits objets p95',
      haut: `les pages et sondes traînent pendant les transferts — critique au-delà de ${s.critMs} ms, le lien est inutilisable en charge`,
      moyen: `réactivité dégradée sous charge — dégradé au-delà de ${s.warnMs} ms, CAKE/fq_codel s'imposent`,
      bas: `le petit trafic passe immédiatement, même pendant un gros transfert`,
      impact: 'impact trafic : DNS, voix et télémétrie patientent quand ça monte',
    },
    {
      term: 'rtt_p95', label: 'RTT p95',
      haut: `l'aller-retour gonfle (bufferbloat) — critique au-delà de ${s.critMs} ms, chaque clic attend la file`,
      moyen: `latence sensible sous charge — dégradé au-delà de ${s.warnMs} ms`,
      bas: 'le lien répond à plat, le buffer ne retient personne',
      impact: 'impact trafic : jeux et voix saccadent, les sessions TCP multiplient les retransmissions',
    },
    {
      term: 'rtt_p50', label: 'RTT médian',
      haut: `la latence typique elle-même est haute — critique au-delà de ${s.critMs} ms, pas seulement les extrêmes`,
      moyen: `latence typique en hausse — à surveiller au-delà de ${s.warnMs} ms`,
      bas: 'la moitié des allers-retours passe sous ce temps — ressenti courant sain',
      impact: 'impact trafic : ce que ressentent la plupart des clics, pas les cas extrêmes',
    },
    {
      term: 'bulk_goodput', label: 'Débit utile',
      haut: `le transfert lourd sature le bord (≥ 50 % de la capacité ${s.shapeCap} Mbit/s = sain)`,
      moyen: `débit correct mais pas au rendez-vous (20–50 % de ${s.shapeCap} Mbit/s)`,
      bas: `sous 20 % de ${s.shapeCap} Mbit/s — la capacité promise n'est pas livrée`,
      impact: `impact trafic : les gros téléchargements s'éternisent, la fenêtre TCP plafonne`,
    },
    {
      term: 'deadline_ok', label: 'Échéances',
      haut: '≥ 95 % des paquets sondes arrivent à l\'échéance — le trafic temps réel tient',
      moyen: '80–95 % — quelques paquets ratent, la voix grésille par à-coups',
      bas: '< 80 % — les paquets arrivent trop tard, retransmissions et voix hachée',
      impact: 'impact trafic : ce qui rate l\'échéance est retransmis — doublon de trafic et gaspillage',
    },
    {
      term: 'QDI', label: 'Dispersion (QDI)',
      haut: 'le spread p95−p50 est grand — la file piétine certains paquets pendant que d\'autres passent',
      moyen: 'dispersion modérée — expérience incohérente selon l\'instant',
      bas: 'p95 proche de la médiane — uniforme, chaque paquet subit la même file',
      impact: 'impact trafic : un QDI qui monte signale des micro-bursts de latence — télémétrie et ACK arrivent par vagues',
    },
    {
      term: 'JFI', label: 'Équité (JFI)',
      haut: 'proche de 1 — les flux se partagent le lien équitablement',
      moyen: '0,8–0,95 — un flux prend un peu plus que sa part',
      bas: 'sous 0,8 — un flux affame les autres, le façonnage par flux (CAKE) s\'impose',
      impact: 'impact trafic : un JFI bas = un gros transfert écrase la voix et la télémétrie',
    },
    {
      term: 'drops', label: 'Pertes',
      haut: `> 100 paquets perdus — retransmissions en cascade, le débit utile s'effondre`,
      moyen: '1–100 pertes — TCP ralentit (congestion), flux sensibles dégradés',
      bas: `zéro perte — rien à renvoyer`,
      impact: 'impact trafic : chaque perte force une retransmission — trafic doublé et délai doublé pour ces octets',
    },
    {
      term: 'wasted', label: 'Gaspillé',
      haut: 'des mégaoctets retransmis à chaque heure — la capacité payée part en doublons',
      moyen: 'quelques centaines de kilooctets — retransmissions occasionnelles',
      bas: `proche de zéro — presque aucun octet renvoyé`,
      impact: 'impact trafic : ces octets traversent le lien deux fois — ils coûtent et ralentissent',
    },
    {
      term: 'cost_ar_per_h', label: 'Coût',
      haut: 'le gaspillage horaire chiffre en Ariary — renégocier le façonnage coûte moins cher',
      moyen: 'quelques Ariary par heure — le prix des pertes courantes',
      bas: `proche de zéro — les pertes ne coûtent presque rien`,
      impact: 'impact trafic : tarif unique 5556 Ar/Go — les forfaits changent tout, voir GET /api/cost/tiers',
    },
  ]
}

// Guide de lecture du mur — repliable, mémorisé (REPLIÉ par défaut : le mur
// respire, l'opérateur déplie au besoin).
export function LiveGuide() {
  const [settings, setSettings] = useState<Settings>(loadSettings())
  const [open, setOpen] = useState<boolean>(() => {
    try { const r = localStorage.getItem('live-guide-open'); return r === '1' } catch { return false }
  })
  useEffect(() => {
    const on = () => setSettings({ ...loadSettings() })
    window.addEventListener('meteolink-settings', on)
    return () => window.removeEventListener('meteolink-settings', on)
  }, [])
  const toggle = () => setOpen(o => {
    try { localStorage.setItem('live-guide-open', o ? '0' : '1') } catch { }
    return !o
  })
  const L = (t: string, c: string) => <span style={{ display: 'block', color: c }}>{t}</span>
  return (
    <div className="card" data-testid="live-guide" style={{ gridColumn: '1 / -1', border: '1px solid var(--hairline)', background: 'var(--surface-card)', padding: '10px 14px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span className="mono" style={{ fontSize: 12, fontWeight: 500, color: '#d6d8dd' }}>Guide de lecture — que disent ces chiffres ?</span>
        <button
          className="btn" onClick={toggle} aria-expanded={open}
          title={open ? 'masquer le guide' : 'afficher le guide'}
          style={{
            marginLeft: 'auto', padding: '3px 10px', fontSize: 11, fontFamily: 'var(--font-mono)',
            border: '1px solid ' + (open ? '#3a3a40' : 'var(--hairline)'),
            background: open ? 'rgba(90,211,227,0.12)' : 'transparent',
            color: open ? '#7fd6e8' : '#a9aeb6',
          }}
        >{open ? 'masquer' : 'afficher'}</button>
      </div>
      {open && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '10px 18px', marginTop: 10 }}>
          {entries(settings).map(e => (
            <div key={e.term} style={{ minWidth: 0 }}>
              <div className="mono" style={{ fontSize: 13, color: '#d6d8dd', marginBottom: 2 }}><Explain term={e.term}>{e.label}</Explain></div>
              <div className="mono" style={{ fontSize: 12, lineHeight: 1.6, color: '#a9aeb6' }}>
                {L(`HAUT : ${e.haut}`, '#f4b400')}
                {L(`MOYEN : ${e.moyen}`, '#a9aeb6')}
                {L(`BAS : ${e.bas}`, '#1fa348')}
                {L(e.impact, '#a9aeb6')}
              </div>
            </div>
          ))}
          <div style={{ minWidth: 0 }}>
            <div className="mono" style={{ fontSize: 13, color: '#d6d8dd', marginBottom: 2 }}>Portes G0–G7 et quarantaine</div>
            <div className="mono" style={{ fontSize: 12, lineHeight: 1.6, color: '#a9aeb6' }}>
              {L('G0 cible joignable · G1 bulk démarré · G2 sondes actives · G3 latence plausible · G4 débit cohérent · G5 pas de doublon · G6 baseline stable · G7 CPU ok — détail au survol dans Campagne.', '#a9aeb6')}
              {L('Quarantaine : les runs douteux (G3/G4) sont exclus des médianes, pas moyennés — voir Archives.', '#f4b400')}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
