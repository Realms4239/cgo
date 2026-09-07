// LiveGuide — guide de lecture des métriques du mur (U6a/U6i): chaque
// chiffre dit ce qu'un HAUT/MOYEN/BAS signifie pour le trafic, seuils
// branchés sur les Réglages réels. Matériel à lire, pas de la déco.
import { useEffect, useState } from 'react'
import { loadSettings, type Settings } from '../lib/settings'

type Entry = { name: string; haut: string; moyen: string; bas: string; impact: string }

function entries(s: Settings): Entry[] {
  return [
    {
      name: 'small_p95',
      haut: `les pages et sondes traînent pendant les transferts — critique au-delà de ${s.critMs} ms, le lien est inutilisable en charge`,
      moyen: `réactivité dégradée sous charge — dégradé au-delà de ${s.warnMs} ms, CAKE/fq_codel s'imposent`,
      bas: `le petit trafic passe immédiatement, même pendant un gros transfert`,
      impact: 'impact trafic : DNS, voix et télémétrie stallent quand ça monte',
    },
    {
      name: 'rtt_p95',
      haut: `l'aller-retour gonfle (bufferbloat) — critique au-delà de ${s.critMs} ms, chaque clic attend la file`,
      moyen: `latence sensible sous charge — dégradé au-delà de ${s.warnMs} ms`,
      bas: 'le lien répond à plat, le buffer ne retient personne',
      impact: 'impact trafic : jeux et voix bégayent, les sessions TCP multiplient les retransmissions',
    },
    {
      name: 'bulk_goodput',
      haut: `le transfert lourd sature le bord (≥ 50 % de la capacité ${s.shapeCap} Mbit/s = sain)`,
      moyen: `débit correct mais pas au rendez-vous (20–50 % de ${s.shapeCap} Mbit/s)`,
      bas: `sous 20 % de ${s.shapeCap} Mbit/s — la capacité promised n'est pas livrée`,
      impact: `impact trafic : les gros téléchargements s'éternisent, la fenêtre TCP plafonne`,
    },
    {
      name: 'deadline_ok',
      haut: '≥ 95 % des paquets sondes arrivent à l\'échéance — le trafic temps réel tient',
      moyen: '80–95 % — quelques paquets ratent, la voix grésille par à-coups',
      bas: '< 80 % — les paquets arrivent trop tard, retransmissions et voix hachée',
      impact: 'impact trafic : ce qui rate l\'échéance est retransmis — doublon de trafic et gaspillage',
    },
    {
      name: 'QDI',
      haut: 'le spread p95−p50 est grand — le file piétine certains paquets pendant que d\'autres passent',
      moyen: 'dispersion modérée — expérience incohérente selon l\'instant',
      bas: 'p95 proche de p50 — uniforme, chaque paquet subit la même file',
      impact: 'impact trafic : un QDI qui monte signale des micro-bursts de latence — télémétrie et ACK arrivent par vagues',
    },
    {
      name: 'drops',
      haut: `> 100 paquets perdus — retransmissions en cascade, goodput s'effondre`,
      moyen: '1–100 pertes — TCP ralentit (congestion), flux sensibles dégradés',
      bas: `zéro perte — rien à renvoyer`,
      impact: 'impact trafic : chaque drop force une retransmission — trafic doublé et délai doublé pour ces octets',
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
            <div key={e.name} style={{ minWidth: 0 }}>
              <div className="mono" style={{ fontSize: 13, color: '#d6d8dd', marginBottom: 2 }}>{e.name}</div>
              <div className="mono" style={{ fontSize: 12, lineHeight: 1.6, color: '#a9aeb6' }}>
                {L(`HAUT : ${e.haut}`, '#f4b400')}
                {L(`MOYEN : ${e.moyen}`, '#a9aeb6')}
                {L(`BAS : ${e.bas}`, '#1fa348')}
                {L(e.impact, '#a9aeb6')}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
