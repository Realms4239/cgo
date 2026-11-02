// Tableau de bord — deux mondes séparés : ce qui est mesuré sur le lien réel,
// et ce que la simulation du banc projette. Aucune étiquette `twin` ou `gelé`
// en premier plan ; les statistiques vivent dans le tiroir scientifique.
import { useEffect, useState } from 'react'
import { useUIStore, type ModeId } from '../store/ui'
import { useAuditStore } from '../store/audit'
import { Badge } from '../components/ui/StatusPip'
import { CardHead } from '../components/ui/CardHead'
import { Provenance } from '../components/ui/Provenance'
import { NavIcon } from '../components/NavIcon'
import { ScienceDetails } from '../components/ScienceDetails'
import { gradeTone } from '../components/AuditReportCard'
import { OperatorStats } from '../components/OperatorStats'
import { RemediationCard } from '../components/RemediationCard'
import { VerdictHero } from '../components/VerdictHero'
import { nextStep, pickHeadline, type SynGroup } from '../lib/synthese'
import { bothAbsent, remediationView, retailBound } from '../lib/dashboard'
import { compareAudits, summaryOf } from '../lib/auditSummary'
import { formatCostAR, groupDigits, relativeTime } from '../lib/format'
import { loadSettings } from '../lib/settings'
import { useModeActive } from '../lib/hooks'

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

const ms = (v?: number) => (v != null && Number.isFinite(v) && v > 0 ? v.toFixed(0) : '—')
const pct = (v?: number) => (v != null && Number.isFinite(v) ? v.toFixed(1) : '—')
const signed = (v: number) => `${v > 0 ? '+' : ''}${v}`

export default function DashboardView() {
  const setMode = useUIStore(s => s.setMode)
  const setBenchStage = useUIStore(s => s.setBenchStage)
  const active = useModeActive('dashboard')
  const running = useUIStore(s => !!s.live?.running)
  const report = useAuditStore(s => s.report)
  const previous = useAuditStore(s => s.previous)
  const showReport = useAuditStore(s => s.showReport)
  const settings = loadSettings()

  const [integrity, setIntegrity] = useState<Integrity | null>(null)
  const [groups, setGroups] = useState<SynGroup[]>([])
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    if (!active) return
    let alive = true
    Promise.all([
      fetch('/api/integrity').then(r => r.json()).catch(() => null),
      fetch('/api/results').then(r => r.json()).catch(() => null),
    ]).then(([i, r]) => {
      if (!alive) return
      if (i?.available) setIntegrity(i)
      else setFailed(true)
      if (Array.isArray(r?.groups)) setGroups(r.groups)
    })
    return () => { alive = false }
  }, [active, running])

  const m = report?.measured
  const primary = report?.economics?.primary
  const cost = formatCostAR(primary?.ar_per_hour)
  const retail = retailBound(report)
  const sim = remediationView(groups)
  const headline = pickHeadline(groups)
  const next = nextStep(running, integrity?.runs)
  const currentSummary = report ? summaryOf(report) : null
  const change = compareAudits(previous, currentSummary)

  const actions: { mode: ModeId; stage?: 'results'; icon: string; title: string }[] = [
    { mode: 'audit', icon: 'audit', title: 'Auditer un lien' },
    { mode: 'bench', icon: 'bench', title: 'Lancer une campagne' },
    { mode: 'bench', stage: 'results', icon: 'dashboard', title: 'Lire le classement' },
    { mode: 'evidence', icon: 'evidence', title: 'Vérifier la preuve' },
  ]

  return (
    <div className="syn dash">
      <header className="syn-hero">
        <VerdictHero report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
        <div className="syn-hero-meta">
          <Badge tone={failed ? 'danger' : integrity ? 'success' : 'steel'}>
            {failed ? 'preuve indisponible' : integrity ? 'preuve vérifiable' : 'chargement…'}
          </Badge>
        </div>
      </header>

      <RemediationCard remediation={report?.remediation} />

      <div className="dash-grid">
        <section className="card dash-card" data-testid="dash-real" aria-labelledby="dash-real-title">
          <CardHead label="État du lien réel — audit" sub="dernier rapport scellé" />
          <span className="dash-tag real">Mesuré sur le lien réel, sans modification</span>
          {report ? (
            <>
              <div className="dash-meta mono">
                <span>{report.site || 'site —'}</span>
                <span>{report.link_type || 'lien —'}{report.provider ? ` · ${report.provider}` : ''}</span>
                {report.timestamp ? <span>{report.timestamp.slice(0, 16).replace('T', ' ')}</span> : null}
                {report.timestamp ? <span className="dash-ago">{relativeTime(report.timestamp)}</span> : null}
                <Badge tone={gradeTone(report.bloat_grade)}>{report.bloat_grade || 'non noté'}</Badge>
              </div>
              <OperatorStats report={report} warnMs={settings.warnMs} critMs={settings.critMs} />
              {report.bloat_verdict ? <p className="syn-read">{report.bloat_verdict}</p> : null}
              <div className="dash-actions">
                <button className="btn btn-primary" onClick={() => { showReport(); setMode('audit') }}>Ouvrir le rapport</button>
                <button className="btn" onClick={() => setMode('audit')}>Auditer de nouveau</button>
              </div>
              <ScienceDetails label="Détails de l'audit réel">
                <dl className="dash-dl mono">
                  <dt>RTT repos p50 / p95</dt><dd>{ms(m?.rtt_idle_p50_ms)} / {ms(m?.rtt_idle_p95_ms)} ms</dd>
                  <dt>RTT chargé p50 / p95</dt><dd>{ms(m?.rtt_loaded_p50_ms)} / {ms(m?.rtt_loaded_p95_ms)} ms</dd>
                  <dt>Δ p95 chargé − repos</dt><dd>{ms(m?.bloat_delta_ms)} ms</dd>
                  <dt>Instabilité (QDI)</dt><dd>{ms(m?.qdi_ms)} ms</dd>
                  <dt>Drain</dt><dd>{m?.drain_measured ? `${ms(m?.drain_ms)} ms` : 'non mesuré'}</dd>
                  <dt>JFI inter-flux</dt><dd>{m?.flow_count && m.flow_count > 1 ? `${pct(m?.jfi_pct)} %` : 'mono-flux / n.a.'}</dd>
                  <dt>Perte</dt><dd>{m?.loss_pct != null ? `${Number(m.loss_pct).toFixed(2)} %` : '—'}</dd>
                  <dt>Complétions 16 Kio</dt><dd>{m?.small_measured ? `${ms(m?.http_small_p95_ms)} ms p95` : 'non mesuré'}</dd>
                  <dt>Lignes de preuve</dt><dd>{report.provenance?.row_count ?? '—'}</dd>
                </dl>
                {report.provenance?.source_sha256 ? (
                  <Provenance
                    source={report.provenance.source_file || 'rapport d’audit'}
                    refresh="au gel"
                    state="live"
                    n={report.provenance.row_count}
                    extra={`sha256 ${report.provenance.source_sha256.slice(0, 12)}…`}
                  />
                ) : null}
              </ScienceDetails>
            </>
          ) : (
            <div className="dash-empty">
              <p className="syn-read">Aucun audit sur ce poste. Un audit de 60 s mesure le lien sans droits administrateur et scelle son rapport.</p>
              <button className="btn btn-primary" onClick={() => setMode('audit')}>Auditer un lien</button>
            </div>
          )}
        </section>

        <section className="card dash-card" data-testid="dash-sim" aria-labelledby="dash-sim-title">
          <CardHead label="Remédiation recommandée — simulation" sub="profil P2 du banc" />
          <span className="dash-tag sim">Simulation de remédiation sur profil P2 — pas une mesure du lien</span>
          {sim ? (
            <>
              <div className="dash-sim-pair">
                <div>
                  <span>file simple{sim.simple.cc ? ` · ${sim.simple.cc.toUpperCase()}` : ''}</span>
                  <b>{sim.simple.deadlinePct != null ? `${pct(sim.simple.deadlinePct)} %` : '—'}</b>
                  <small>des alertes à l'heure · n={sim.simple.n}</small>
                </div>
                <div>
                  <span>{sim.best.qdisc}{sim.best.cc ? ` · ${sim.best.cc.toUpperCase()}` : ''}</span>
                  <b>{sim.best.deadlinePct != null ? `${pct(sim.best.deadlinePct)} %` : '—'}</b>
                  <small>des alertes à l'heure · n={sim.best.n}</small>
                </div>
              </div>
              <p className="syn-read">
                {sim.simple.smallP95Ms != null && sim.best.smallP95Ms != null
                  ? `Délai sous charge : ${ms(sim.simple.smallP95Ms)} ms → ${ms(sim.best.smallP95Ms)} ms sur les répétitions valides. `
                  : ''}
                Projection sur le profil P2 ; le lien audité n'a pas été façonné.
              </p>
              <div className="dash-actions">
                <button className="btn btn-primary" onClick={() => { setBenchStage('results'); setMode('bench') }}>Voir le classement</button>
              </div>
              <ScienceDetails label="Détails de la simulation">
                <dl className="dash-dl mono">
                  <dt>Cellule simple</dt><dd>{sim.simple.qdisc} · {sim.simple.cc || '—'} · n={sim.simple.n}</dd>
                  <dt>Cellule recommandée</dt><dd>{sim.best.qdisc} · {sim.best.cc || '—'} · n={sim.best.n}</dd>
                  <dt>Délai sous charge simple</dt><dd>{ms(sim.simple.smallP95Ms)} ms (médiane)</dd>
                  <dt>Délai sous charge recommandé</dt><dd>{ms(sim.best.smallP95Ms)} ms (médiane)</dd>
                  <dt>Source</dt><dd>campagnes archivées · {integrity?.runs ?? '—'} runs · hash {integrity?.hash8 ?? '—'}</dd>
                </dl>
                {headline.simple && headline.best ? (
                  <p className="syn-read">Agrégat des cellules P2 valides ; les cellules sans échantillon valide restent exclues.</p>
                ) : null}
              </ScienceDetails>
            </>
          ) : (
            <div className="dash-empty">
              <p className="syn-read">Aucune campagne archivée exploitable — la simulation n'a pas de données.</p>
              <button className="btn btn-primary" onClick={() => { setBenchStage('config'); setMode('bench') }}>Ouvrir le banc d'essai</button>
            </div>
          )}
        </section>
      </div>

      {report ? (
        bothAbsent(!!primary, !!(change.comparable && currentSummary)) ? (
          <section className="card dash-quiet" data-testid="dash-quiet">
            <p className="syn-read"><b>Ce que ça coûte</b> — non mesuré : aucun coût n’est chiffré plutôt qu’un zéro fabriqué.</p>
            <p className="syn-read"><b>Ce qui a changé</b> — aucune comparaison disponible{change.reason ? ` — ${change.reason}` : ''}. Il faut deux audits du même site et du même type de lien.</p>
          </section>
        ) : (
      <div className="dash-grid dash-grid-even">
        <section className="card dash-card" data-testid="dash-cost">
          <CardHead label="Ce que ça coûte" sub="scénario contractuel, pas une mesure" />
          {primary ? (
            <>
              <div className="dash-cost">
                <span>Perte financière</span>
                <b>{cost.num}<small> {cost.unit}</small></b>
                <em>{primary.name} · scénario primaire</em>
              </div>
              {retail != null ? (
                <p className="syn-read">
                  Borne haute hors-forfait : <b>{groupDigits(retail, 2)} Ar/h</b> — une pénalité de secours,
                  pas un modèle d'infrastructure.
                </p>
              ) : null}
              <p className="dash-disclaimer">{report?.economics?.disclaimer || 'Scénarios contractuels, pas des mesures.'}</p>
            </>
          ) : (
            <p className="syn-read">Non mesuré — aucun coût n'est chiffré plutôt qu'un zéro fabriqué.</p>
          )}
        </section>

        <section className="card dash-card" data-testid="dash-change">
          <CardHead label="Ce qui a changé" sub="comparaison locale avec l'audit précédent" />
          {change.comparable && currentSummary ? (
            <>
              <div className="dash-change">
                {change.smallDeltaPct != null ? (
                  <div>
                    <span>Délai sous charge</span>
                    <b data-dir={change.smallDeltaPct > 0 ? 'bad' : change.smallDeltaPct < 0 ? 'good' : 'same'}>
                      {signed(change.smallDeltaPct)} %
                    </b>
                  </div>
                ) : null}
                {change.deadlineDelta != null ? (
                  <div>
                    <span>Alertes à l'heure</span>
                    <b data-dir={change.deadlineDelta > 0 ? 'good' : change.deadlineDelta < 0 ? 'bad' : 'same'}>
                      {signed(change.deadlineDelta)} pts
                    </b>
                  </div>
                ) : null}
              </div>
              <p className="syn-read">
                {previous?.timestamp ? `Audit précédent du ${previous.timestamp.slice(0, 10)}` : 'Audit précédent'} ·{' '}
                {previous?.grade ? `qualité ${previous.grade} → ${currentSummary.grade || '—'}` : `qualité ${currentSummary.grade || '—'}`}.
              </p>
            </>
          ) : (
            <p className="syn-read">
              Aucune comparaison disponible{change.reason ? ` — ${change.reason}` : ''}. Il faut deux audits du
              même site et du même type de lien.
            </p>
          )}
        </section>
      </div>
        )
      ) : (
        <section className="card dash-onboarding" data-testid="dash-onboarding">
          <CardHead label="Premier pas" sub="trois gestes, aucune promesse" />
          <h2>Le tableau se remplit après le premier audit.</h2>
          <ol className="dash-steps">
            <li><b>1 — Auditer</b><span>60 s sur le lien réel, sans droits administrateur.</span></li>
            <li><b>2 — Comparer</b><span>Le banc projette la file simple et CAKE sur le profil P2.</span></li>
            <li><b>3 — Prouver</b><span>Le rapport scellé et le corpus restent citables.</span></li>
          </ol>
          <button className="btn btn-primary" onClick={() => setMode('audit')}>AUDITER UN LIEN</button>
        </section>
      )}

      <section className="syn-proof" aria-label="Preuve archivée">
        <div className="syn-proof-head">
          <span className="mono">Preuve archivée</span>
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
          extra={integrity?.hash8 ? `sceau ${integrity.hash8}` : undefined}
        />
      </section>

      <section className="syn-next syn-next-wide">
        <CardHead label="Prochaine action" />
        <h2>{next.title}</h2>
        <p>{next.body}</p>
        <button className="btn btn-primary" onClick={() => { if (next.mode === 'bench') setBenchStage('config'); setMode(next.mode) }}>{next.cta}</button>
        <button className="btn" onClick={() => setMode('evidence')}>Vérifier la preuve</button>
      </section>

      <nav className="syn-actions" aria-label="Actions rapides">
        {actions.map(a => (
          <button key={`${a.mode}-${a.title}`} className="syn-action" onClick={() => { if (a.stage) setBenchStage(a.stage); setMode(a.mode) }}>
            <NavIcon id={a.icon} size={20} />
            <span><b>{a.title}</b></span>
          </button>
        ))}
      </nav>
    </div>
  )
}
