import { useEffect, useState } from 'react'
import { startReplay, stopReplay } from '../lib/replay'
import { useUIStore } from '../store/ui'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import Explain from '../components/Explain'
import { PeekPopover } from '../components/PeekPopover'
import { hardwareRecommendation } from '../lib/hardware'
import { asArray } from '../lib/format'
import { paginate, Paginate } from '../components/ui/Paginate'

type Integrity = {
  available: boolean; reason?: string
  runs?: number; manifests?: number; valid?: number; degraded?: number; quarantined?: number
  run_ids?: string[]
  breakdown?: { run: string; rows: number; valid: number; degraded?: number; quarantined: number }[]
  updated?: string
  hash8?: string; sha256?: string
}

export default function IntegriteView({ embedded = false, active = true }: { embedded?: boolean; active?: boolean }) {
  const [data, setData] = useState<Integrity|null>(null)
  const [err, setErr] = useState<string|null>(null)
  const [msg, setMsg] = useState('')
  const [replayRuns, setReplayRuns] = useState<string[]>([])
  const [regenOk, setRegenOk] = useState(false)
  const [peek, setPeek] = useState<{ rect: DOMRect; run: string } | null>(null)
  const replayRunning = useUIStore(s=>s.replayRunning)
  const setMode = useUIStore(s=>s.setMode)
  const setBenchStage = useUIStore(s=>s.setBenchStage)
  const [groups, setGroups] = useState<any[]|null>(null)
  const [peekGroups, setPeekGroups] = useState<any[]|null>(null)
  const [quar, setQuar] = useState<{run:string;event_id:number;profile:string;qdisc:string;cc:string;gate_status:string;failed_gates?:string[]}[]>([])
  // journal opérateur — vit ici (archives), plus dans Résultats (classement)
  const [events, setEvents] = useState<{ ts: string; kind: string; msg: string }[]>([])
  const [evPage, setEvPage] = useState(1)
  const EV_PAGE_SIZE = 8
  const [quarSum, setQuarSum] = useState<{total_invalid?:number;g4_low?:number;g4_high?:number;empty_probes?:number;g3_implausible?:number}|null>(null)
  // pagination runs (U6h) — 407 lignes rendues d'un bloc noyait la vue
  const [runsPage, setRunsPage] = useState(1)
  const [quarPage, setQuarPage] = useState(1)
  const QUAR_PAGE_SIZE = 12
  const RUNS_PAGE_SIZE = 12

  const load = () => {
    fetch('/api/integrity').then(r=>r.json()).then(j=>setData(j)).catch(e=>setErr(String(e)))
    fetch('/api/replay/list').then(r=>r.json()).then(j=>setReplayRuns(asArray<string>(j.runs))).catch(()=>{})
    fetch('/api/results').then(r=>r.json()).then(j=>setGroups(asArray(j.groups))).catch(()=>setGroups([]))
    fetch('/api/quarantine').then(r=>r.json()).then(j=>setQuar(asArray(j.quarantines))).catch(()=>{})
    fetch('/api/events').then(r=>r.json()).then(j=>setEvents(asArray<{ts:string;kind:string;msg:string}>(j.events).reverse())).catch(()=>{})
    fetch('/api/quarantine/summary').then(r=>r.json()).then(j=>setQuarSum(j)).catch(()=>{})
  }
  useEffect(()=>{
    if (!active) return
    load()
    const t=setInterval(load,10000)
    return ()=>clearInterval(t)
  }, [active])
  useEffect(()=>{
    if (!peek) return
    setPeekGroups(null)
    fetch(`/api/results?run=${encodeURIComponent(peek.run)}`).then(r=>r.json()).then(j=>setPeekGroups(asArray(j.groups))).catch(()=>setPeekGroups([]))
  },[peek?.run])

  const verify = async () => {
    setMsg('vérification…')
    const r = await fetch('/api/integrity')
    const j = await r.json()
    setData(j); setMsg(j.available ? 'intégrité vérifiée' : j.reason)
  }
  const copyFingerprint = () => {
    if (!data?.sha256) { setMsg('empreinte SHA non exposée par /api/integrity'); return }
    navigator.clipboard?.writeText(data.sha256).then(
      () => setMsg('empreinte copiée'),
      () => setMsg('copie impossible — sélectionnez le texte'),
    )
  }
  const regen = async () => {
    setMsg('génération…')
    const r = await fetch('/api/figures/regen', {method:'POST'})
    if (r.ok) { setMsg('figures régénérées'); setRegenOk(true) }
    else setMsg('échec: '+r.status)
  }

  if (err) return <div className="card">{!embedded && <h1 className="view-title">Intégrité</h1>}<EmptyState kind="error" hint={err} /></div>
  if (!data) return <div className="card">{!embedded && <h1 className="view-title">Intégrité</h1>}<EmptyState kind="loading" hint="vérification des archives" /></div>
  if (!data.available) return <div className="card">{!embedded && <h1 className="view-title">Intégrité</h1>}<EmptyState kind="empty" hint={data.reason} /><button className="btn btn-primary int-retry" onClick={load}>Réessayer</button></div>

  // Quarantaine synthétique retirée — gate_status réel depuis Scan.
  // single scroll: main défile déjà (double-scroll supprimé).
  return (
    <div className="panel-stack int-stack">
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono int-peek-run">run — {peek.run}</div>
          {(() => {
            const src = asArray<any>(peekGroups).length ? peekGroups as any[] : asArray<any>(groups)
            if (src.length===0) return <div className="mono int-peek-empty">aucun groupe — gel d'abord</div>
            // Meilleur par profil pour ce run, repli sur les 4 premiers.
            const bests = src.filter((g:any)=>g.best)
            const show = bests.length ? bests.slice(0,4) : src.slice(0,4)
            return show.map((g: any, i: number) => (
              <div key={i} className="int-peek-row">
                <span className="int-peek-id">{g.profile}·{g.qdisc}·{g.cc}</span>
                <span className="int-peek-val" data-best={g.best ? '1' : undefined}>{g.best ? '★ ' : ''}{Number(g.small_p95_median ?? 0).toFixed(1)} ms</span>
              </div>
            ))
          })()}
        </PeekPopover>
      )}
      {!embedded && <h1 className="view-title">Intégrité — archives scellées</h1>}
      {/* quatre comptes du corpus — 32 px, l'empreinte complète reste copiable */}
      <section className="card int-summary" aria-label="Synthèse du corpus scellé">
        <dl className="int-counts">
          {[
            ['runs archivés', data.runs ?? 0],
            ['mesures valides', data.valid ?? 0],
            ['dégradées', data.degraded ?? 0],
            ['écartées', data.quarantined ?? 0],
          ].map(([label, value]) => (
            <div className="int-count" key={label}>
              <dt>{label}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
        <p className="mono int-seal">
          <span>sceau {data.hash8 ?? '—'} · {data.manifests ?? '—'} manifestes · maj {data.updated ? data.updated.slice(0, 16).replace('T', ' ') : '—'}</span>
          <button type="button" className="btn int-copy" onClick={copyFingerprint}>Copier l'empreinte</button>
        </p>
        <div className="int-summary-actions">
          <button className="btn btn-primary" onClick={verify}>Vérifier l'intégrité</button>
          <a className="btn int-export" href="/api/report/export?format=md" download>Rapport MD</a>
          <span className="mono muted" role="status" aria-live="polite">{msg}</span>
        </div>
      </section>
      <details className="int-corpus-details" data-testid="corpus-details">
        <summary className="card-head">Détails du corpus</summary>
      <details className="card int-sources-details" data-testid="sources-card">
        <summary className="card-head">D'où viennent ces chiffres</summary>
        <div className="int-sources">
          <div className="mono int-source-line"><b>Manifestes signés</b> — la liste scellée de chaque run archivé (fichiers + empreinte) · dans le dossier de chaque run · mis à jour {data.updated || '—'}</div>
          <div className="mono int-source-line"><b>Quarantaine</b> — les mesures écartées par les <Explain term="G3">portes de validation</Explain> · dans le registre de quarantaine du serveur · à chaque gel de campagne</div>
          <div className="mono int-source-line"><b>Mesures</b> — le tableur des mesures archivées, une ligne par sondage · dans le fichier de mesures de chaque run · mis à jour {data.updated || '—'}</div>
          <div className="mono int-source-line"><b>Journal</b> — les opérations (campagnes, audits) racontées au fil de l'eau · dans le journal opérateur ci-dessous · en continu</div>
        </div>
      </details>
      {/* RDF provenance — important look, not debug dump */}
      <div className="card int-rdf">
        <div className="card-head int-rdf-head">
          <span>RDF — provenance scellée</span>
          <span className="mono int-rdf-tag">frozen-wave</span>
        </div>
        <div className="int-rdf-body">
          <div className="mono int-rdf-cap">sha256 manifest</div>
          {/* SHA réel exposé par /api/integrity */}
          {data.sha256 ? (
            <div data-testid="provenance-hash" className="mono int-hash">
              sha256:{data.sha256}
            </div>
          ) : (
            <div title="empreinte SHA non exposée par /api/integrity — afficher le manifeste archivé">
              <EmptyState kind="empty" hint="empreinte SHA non exposée par /api/integrity — afficher le manifeste archivé" />
            </div>
          )}
          <div className="mono int-rdf-note">manifestes signés · registre de quarantaine · tableur des mesures — les trois couvrent les mêmes runs archivés, l'empreinte ci-dessus les scelle</div>
          <div className="int-badges">
            <span className="mono int-badge int-badge-live">Mémoire, tableau 7</span>
            <span className="mono int-badge int-badge-ok">archivé • vérifiable</span>
          </div>
        </div>
      </div>
      <div className="card" data-testid="runs-table">
        <div className="card-head">Runs archivés</div>
        <div className="tbl-scroll int-scroll" data-testid="runs-scroll">
        <table className="data-table int-table">
          <thead><tr className="int-thead-row">
            <th>run</th><th>lignes</th><th>valides</th><th>dégradées</th><th>quar.</th><th>résultats</th><th>rejouer</th>
          </tr></thead><tbody>
          {paginate(Array.isArray(data.breakdown) ? data.breakdown : (asArray<string>(data.run_ids).map(id=>({run:id,rows:0,valid:0,degraded:0,quarantined:0}))), runsPage, RUNS_PAGE_SIZE).map(b=>(
            <tr key={b.run} className="int-row" onMouseEnter={e=>setPeek({rect:(e.currentTarget as HTMLElement).getBoundingClientRect(), run:b.run})} onMouseLeave={()=>setPeek(null)}>
              <td className="int-td-run"><span title={b.run} className="int-run-id">{b.run}</span></td><td>{b.rows}</td><td>{b.valid}</td><td>{b.degraded ?? 0}</td><td>{b.quarantined}</td>
              <td><a href={`/api/results?run=${b.run}`} target="_blank" rel="noreferrer" className="int-link">résultats</a></td>
              <td>{b.rows > 0 ? (
                <button className="btn btn-primary int-replay" data-testid="run-rejouer" onClick={()=>{ startReplay(b.run); setBenchStage('observe'); setMode('bench') }}>Rejouer</button>
              ) : (
                <span className="mono data-missing" title="aucune ligne archivée — rien à rejouer">non rejouable</span>
              )}</td>
            </tr>
          ))}
          </tbody></table>
        </div>
          <Paginate
            total={(Array.isArray(data.breakdown) ? data.breakdown : asArray<string>(data.run_ids)).length}
            page={runsPage} pageSize={RUNS_PAGE_SIZE} onPage={setRunsPage}
          />
      </div>
      <div className="card" data-testid="quarantine-table">
        <div className="card-head">Quarantaine — lignes invalidées par les portes</div>
        {quar.length===0 ? <div className="int-empty-row"><EmptyState kind="empty" hint="Aucune mise en quarantaine" /></div> :
        <div className="tbl-scroll int-scroll" data-testid="quarantine-scroll">
        <table className="data-table int-table">
          <thead><tr className="int-thead-row">
            <th>run</th><th>événement</th><th>cellule</th><th>statut</th><th>portes</th>
          </tr></thead><tbody>
          {paginate(quar, quarPage, QUAR_PAGE_SIZE).map((q,i)=><tr key={i} className="int-row">
            <td className="int-td-run"><span title={q.run} className="int-run-id">{q.run}</span></td><td>#{q.event_id}</td><td className="int-quar-cell">{q.profile}·{q.qdisc}·{q.cc}</td><td className="int-quar-status">{q.gate_status}</td><td className="int-quar-gates">{(q.failed_gates ?? []).join(' ') || '—'}</td>
          </tr>)}
          </tbody></table>
        </div>}
        <Paginate
          total={quar.length}
          page={quarPage} pageSize={QUAR_PAGE_SIZE} onPage={setQuarPage}
        />
        <div className="mono int-note">{quarSum && quarSum.total_invalid != null ? (<>{quarSum.total_invalid} mesures écartées au total — débit trop bas (<Explain term="G4">porte G4</Explain>) {quarSum.g4_low} · débit trop haut (<Explain term="G4">porte G4</Explain>) {quarSum.g4_high} · sondes vides {quarSum.empty_probes} · latence incohérente (<Explain term="G3">porte G3</Explain>) {quarSum.g3_implausible} <span className="int-note-muted">(un compteur peut cumuler)</span></>) : 'mesures écartées par les portes de validation — détail par ligne ci-dessus'}</div>
        <div className="mono int-note-fine">registre de quarantaine du serveur — une ligne par mesure écartée, avec sa cellule et ses portes</div>
      </div>
      <div className="card int-reco-card">
        <div className="card-head">Recommandations — Traduction Matérielle</div>
        <p className="mono int-reco-lede">
          Linux prouvé en lab → matériel DSI, sans réécrire l'infra. <span className="int-reco-lede-live">Observé côté client</span>.
        </p>
        <div className="int-reco-scroll">
          <table className="int-hw-table">
            <thead>
              <tr className="int-hw-thead">
                <th className="int-hw-th">Principe Linux</th>
                <th className="int-hw-th">MikroTik</th>
                <th className="int-hw-th">ISP / mini-PC gateway</th>
              </tr>
            </thead>
            <tbody>
              <tr className="int-row">
                <td className="int-hw-td">pfifo_fast <span className="int-hw-faint">baseline FIFO</span></td>
                <td className="int-hw-td">— ne pas reproduire</td>
                <td className="int-hw-td">— ne pas reproduire</td>
              </tr>
              <tr className="int-row">
                <td className="int-hw-td">fq_codel <span className="int-hw-ok">prouvé lab</span></td>
                <td className="int-hw-td">Queue Tree PCQ / CAKE RouterOS v7+</td>
                <td className="int-hw-td">mini-PC transparent bridge CAKE en amont CPE</td>
              </tr>
              <tr>
                <td className="int-hw-td">cake <span className="int-hw-ok">prouvé lab</span></td>
                <td className="int-hw-td">CAKE RouterOS v7+ Queue Tree</td>
                <td className="int-hw-td">mini-PC gateway CAKE transparent</td>
              </tr>
            </tbody>
          </table>
        </div>
        {asArray<any>(groups).filter((g:any)=>g.best).length>0 && (
          <div className="int-hw-byprofile">
            <div className="mono int-hw-cap">Traduction matérielle par profil (depuis Scan ★ best)</div>
            {asArray<any>(groups).filter((g:any)=>g.best).slice(0,4).map((g:any)=>(
              <div key={`${g.profile}-${g.qdisc}`} className="mono int-hw-row">
                <span className="int-hw-profile">{g.profile}</span> · {g.qdisc} → {g.hardware_recommendation || hardwareRecommendation(g.qdisc, g.profile)}
              </div>
            ))}
          </div>
        )}
        <div className="int-hw-kits">
          <div className="int-kit int-kit-live">
            <b>Kit de Diagnostic Portable DSI</b> <span className="mono int-chip int-chip-live">5-min branch</span>
            <div className="int-kit-body">Branchement sans <code className="int-code-primary">tc</code> : laptop → Yas/Telma 4G → <code className="int-code-primary">1.1.1.1</code> ping 5 Hz + Small 4–32 K + bulk GRIB. Zéro install, client-side observé depuis les locaux — pas contrôleur réseau.</div>
          </div>
          <div className="int-kit int-kit-warn">
            <b>Pilote isolé</b> <span className="mono int-chip int-chip-warn">1 site → 1 département 4G</span>
            <div className="int-kit-body">Un seul site 4G à la fois, mesure avant/après. Si CAKE prouvé en lab réduit <code className="int-code-ok">small_p95 738 → 20 ms</code>, alors seulement généraliser. Pas de déploiement global sans pilote.</div>
          </div>
          <div className="int-kit int-kit-plain">
            <b>Why Go: zero-dependency vs Flent</b>
            <div className="int-kit-body">Un seul binaire <code className="int-code-primary">cgo</code> posable sur laptop DSI — pas de Python, pas de netperf, pas de dépendance. Flent exige stack complète et lab ; <code className="int-code-live">Go 1.25</code> embarque <code className="int-code-primary">embed dist</code> + csv/metrics/probe campagne en un fichier. Reproductible offline, <code className="int-code-faint">manifest.json sha256 + quarantine.json</code> vérifiable.</div>
          </div>
        </div>
        <div className="mono int-foot">
          Traduction matérielle : <b className="mono">{(() => { try { const g = (window as any).__CGO_GROUPS ?? null; return g?.best?.profile ?? 'P2' } catch { return 'P2' } })()}</b> — routeur Linux/pfSense : tc direct · MikroTik : Queue Tree PCQ/CAKE (RouterOS v7+) · accès ISP : passerelle dédiée en amont
        </div>
      </div>
      <div className="card">
        <div className="card-head">Figures</div>
        <p className="mono muted int-fig-note">Générées par <span className="mono int-fig-code">make figures</span> depuis les CSV archivés.</p>
        <div className="int-fig-actions">
          <button className="btn btn-primary" onClick={regen}>Régénérer</button>
          <span className="mono muted">{regenOk ? '✓ disponible' : ''}</span>
        </div>
        {regenOk && (
          <div className="int-figures">
            <div>
              <img src="/api/figures/small_p95.svg" alt="small p95" className="int-figure" />
              <div className="mono muted int-figure-cap">small p95 par configuration · hash {data.hash8 ?? '—'}</div>
            </div>
            <div>
              <img src="/api/figures/scatter.svg" alt="scatter" className="int-figure" />
              <div className="mono muted int-figure-cap">compromis latence / débit · hash {data.hash8 ?? '—'}</div>
            </div>
          </div>
        )}
      </div>
      <Provenance source="manifestes signés des runs archivés" refresh="au gel" state="live" />
      {events.length > 0 && (
        <div className="card" data-testid="changelog">
          <div className="card-head">Changelog — journal opérateur</div>
          <ul className="int-changelog">
            {paginate(events, evPage, EV_PAGE_SIZE).map((e, i) => {
              // pastille de nature — campagne cyan, profil violet, audit ambre, reste acier
              const kc = /campagne/i.test(e.kind) ? 'var(--t-live)' : /profil/i.test(e.kind) ? 'var(--t-bbr)' : /audit/i.test(e.kind) ? 'var(--t-threshold)' : 'var(--t-neutral)'
              return (
                <li key={i} className="mono int-event">
                  <span className="int-event-ts">{e.ts}</span>
                  <span title={e.kind} className="int-event-dot" style={{ color: kc }}>●</span>
                  <span className="ev-msg int-event-msg">{e.msg.slice(0, 120)}</span>
                  <span className="int-event-kind" style={{ color: kc }}>{e.kind}</span>
                </li>
              )
            })}
          </ul>
          <Paginate total={events.length} page={evPage} pageSize={EV_PAGE_SIZE} onPage={setEvPage} />
        </div>
      )}
      <div className="card">
        <div className="card-head">Replay {replayRunning && <span className="mono int-replay-live">● en cours</span>}</div>
        {replayRunning ? (
          <div className="int-replay-row">
            <span className="mono muted">replay actif — voir l'observation du banc</span>
            <button className="btn int-replay-stop" onClick={()=>stopReplay()}>Arrêter</button>
            <button className="btn btn-primary" onClick={()=>{ setBenchStage('observe'); setMode('bench') }}>Voir l'observation</button>
          </div>
        ) : replayRuns.length===0 ? <p className="mono muted">aucun run à rejouer</p> :
          <ul className="int-replay-list">
            {replayRuns.map(id=>(
              <li key={id} className="int-replay-item">
                <span title={id} className="int-replay-id">{id}</span>
                <button className="btn btn-primary int-replay" onClick={()=>{ startReplay(id); setBenchStage('observe'); setMode('bench')}}>Rejouer</button>
              </li>
            ))}
          </ul>
        }
        {/* pas de portée à choisir : le replay rejoue tout le run — l'API
            n'expose aucun filtre par cellule ou métrique (run seul). */}
        {!replayRunning && replayRuns.length>0 && (
          <p className="mono int-replay-note">le replay rejoue tout le run — aucun filtre par cellule ou métrique côté serveur, il n'y a pas de portée à choisir</p>
        )}
      </div>
      </details>
    </div>
  )
}
