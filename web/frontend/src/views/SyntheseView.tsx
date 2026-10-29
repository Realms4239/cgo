import { useEffect, useState } from 'react'
import { useUIStore } from '../store/ui'
import { Badge } from '../components/ui/StatusPip'
import { CardHead } from '../components/ui/CardHead'
import { Provenance } from '../components/ui/Provenance'
import { NavIcon } from '../components/NavIcon'
import { nextStep, pickHeadline, type SynGroup } from '../lib/synthese'

type Integrity = {
  available?: boolean
  runs?: number
  manifests?: number
  valid?: number
  degraded?: number
  quarantined?: number
  updated?: string
  hash8?: string
}

const pct = (v?: number) => v == null || !Number.isFinite(v) ? '—' : v.toFixed(1)

export default function SyntheseView() {
  const setPanel = useUIStore(s => s.setPanel)
  const connected = useUIStore(s => s.connected)
  const running = useUIStore(s => !!s.live?.running)
  const [integrity, setIntegrity] = useState<Integrity | null>(null)
  const [groups, setGroups] = useState<SynGroup[]>([])
  const [mode, setMode] = useState('')
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let alive = true
    Promise.all([
      fetch('/api/integrity').then(r => r.json()).catch(() => null),
      fetch('/api/results').then(r => r.json()).catch(() => null),
      fetch('/api/health').then(r => r.json()).catch(() => null),
    ]).then(([i, r, h]) => {
      if (!alive) return
      if (i?.available) setIntegrity(i)
      else setFailed(true)
      if (Array.isArray(r?.groups)) setGroups(r.groups)
      if (h?.mode) setMode(h.mode)
    })
    return () => { alive = false }
  }, [])

  const headline = pickHeadline(groups)
  const next = nextStep(running, integrity?.runs)
  const actions = [
    { panel: 'campagne' as const, icon: 'campagne', title: 'Auditer un lien', sub: '60 s, sans droits admin' },
    { panel: 'campagne' as const, icon: 'live', title: 'Lancer une campagne', sub: 'matrice P1–P4, 4 flux' },
    { panel: 'resultats' as const, icon: 'resultats', title: 'Lire le classement', sub: 'cellules gelées, IQR, coût' },
    { panel: 'integrite' as const, icon: 'integrite', title: 'Vérifier la preuve', sub: 'SHA-256, runs, quarantaine' },
  ]

  return (
    <div className="syn">
      <header className="syn-hero">
        <div>
          <h1>Synthèse du lien</h1>
          <p>Audit client-side, banc dual-ns et preuve gelée — Météo Madagascar.</p>
        </div>
        <div className="syn-hero-meta">
          <Badge tone={mode === 'observe' ? 'steel' : 'success'}>{mode === 'observe' ? 'observation' : mode || 'mode —'}</Badge>
          <span className="mono">{connected ? 'SSE connecté' : 'SSE déconnecté'}</span>
        </div>
      </header>

      <section className="syn-proof" aria-label="Preuve gelée">
        <div className="syn-proof-head">
          <span className="mono">Preuve gelée</span>
          <span className="mono">{integrity?.updated ? `maj ${integrity.updated.replace('T', ' ').replace('Z', '')}` : failed ? 'indisponible' : 'chargement…'}</span>
        </div>
        <div className="syn-proof-grid">
          <div className="syn-proof-cell"><b>{integrity ? integrity.runs ?? 0 : '—'}</b><span>runs</span></div>
          <div className="syn-proof-cell"><b>{integrity ? integrity.valid ?? 0 : '—'}</b><span>valides</span></div>
          <div className="syn-proof-cell"><b>{integrity ? integrity.degraded ?? 0 : '—'}</b><span>dégradées</span></div>
          <div className="syn-proof-cell"><b>{integrity ? integrity.quarantined ?? 0 : '—'}</b><span>écartées</span></div>
        </div>
        <Provenance
          source="manifestes signés des runs archivés"
          refresh="au gel"
          state={integrity ? 'live' : 'wait'}
          n={integrity?.manifests}
          extra={integrity?.hash8 ? `hash ${integrity.hash8}` : undefined}
        />
      </section>

      <div className="syn-split">
        <section className="syn-constat">
          <CardHead label="Constat agrégé — P2 cellulaire" />
          {headline.simple && headline.best ? (
            <>
              <div className="syn-pair">
                <div>
                  <span>file simple · BBR</span>
                  <b>{pct(headline.simple.deadline_median)}<small>%</small></b>
                  <small>échéances tenues · n={headline.simple.deadline_valid_n}</small>
                </div>
                <div>
                  <span>{headline.best.qdisc} · {headline.best.cc.toUpperCase()}</span>
                  <b>{pct(headline.best.deadline_median)}<small>%</small></b>
                  <small>meilleure cellule · n={headline.best.deadline_valid_n}</small>
                </div>
              </div>
              <p className="syn-read">
                {headline.simple.small_p95_median != null && headline.best.small_p95_median != null
                  ? `small p95 ${pct(headline.simple.small_p95_median)} ms → ${pct(headline.best.small_p95_median)} ms sur les répétitions valides. `
                  : ''}
                Agrégat de tout le corpus gelé, pas la campagne finale ; pas une mesure de production. Les cellules sans échantillon valide restent hors du constat.
              </p>
            </>
          ) : (
            <p className="syn-read">Aucun agrégat P2 exploitable dans le corpus. Le classement s’ouvre sur les lignes gelées, sans chiffre inventé.</p>
          )}
          <button className="btn btn-primary" onClick={() => setPanel('resultats')}>Ouvrir le classement</button>
        </section>

        <aside className="syn-next">
          <CardHead label="Prochaine action" />
          <h2>{next.title}</h2>
          <p>{next.body}</p>
          <button className="btn btn-primary" onClick={() => setPanel(next.panel)}>{next.cta}</button>
          <button className="btn" onClick={() => setPanel('integrite')}>Vérifier la preuve</button>
        </aside>
      </div>

      <nav className="syn-actions" aria-label="Actions rapides">
        {actions.map(a => (
          <button key={`${a.panel}-${a.title}`} className="syn-action" onClick={() => setPanel(a.panel)}>
            <NavIcon id={a.icon} size={20} />
            <span><b>{a.title}</b><small>{a.sub}</small></span>
          </button>
        ))}
      </nav>
    </div>
  )
}
