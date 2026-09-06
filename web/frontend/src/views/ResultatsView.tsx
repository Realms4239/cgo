import { useEffect, useMemo, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar } from '../lib/anime'
import { echarts } from '../lib/echarts'
import { baseOption, scatterSeries } from '../lib/chartGrammar'
import CompareView, { type Pinned } from '../components/CompareView'
import { CRAFT } from '../lib/chartGrammar'
import Explain from '../components/Explain'
import InterpretationView from '../components/InterpretationView'
import { useUIStore } from '../store/ui'
import { fmtIQR } from '../lib/format'
import { asArray } from '../lib/format'

type Group = {
  profile: string; qdisc: string; cc: string; direction?: string
  count: number; quarantined: number
  rtt_p95_median: number; rtt_p95_iqr?: [number, number]
  small_p95_median: number; small_p95_iqr?: [number, number]
  // valid-only strict (degraded exclus) — preuve pilote, parité mémoire.
  small_p95_valid_n?: number; small_p95_valid_median?: number
  small_p95_valid_iqr?: [number, number]; small_p95_valid_ci95?: [number, number]
  deadline_valid_n?: number; deadline_valid_median?: number; deadline_valid_ci95?: [number, number]
  goodput_median: number
  deadline_median?: number; deadline_ok_pct?: number
  wasted_median?: number; cost_median?: number; wasted_bytes?: number; cost_ar_per_h?: number
  best?: boolean; hardware_recommendation?: string
}

// le classement EST la comparaison toutes cellules (Q1/Q2): critère choisi,
// filtres profil/file/CC, verdict recalculé — tout depuis les CSV gelés.
const RANKS = [
  { key: 'small_p95_median', label: 'small p95', dir: 'down' as const, unit: 'ms', term: 'small_p95' },
  { key: 'rtt_p95_median', label: 'RTT p95', dir: 'down' as const, unit: 'ms', term: 'rtt_p95' },
  { key: 'goodput_median', label: 'goodput', dir: 'up' as const, unit: 'Mbit/s', term: 'bulk_goodput' },
  { key: 'cost', label: 'coût', dir: 'down' as const, unit: 'Ar/h', term: 'cost_ar_per_h' },
]

const QCOLOR: Record<string, string> = { cake: 'var(--t-bbr)', fq_codel: 'var(--t-live)', pfifo_fast: '#6b7078' }

// TradeSpace — le nuage paramétrable façon DeepSWE (score vs coût) : axes X/Y
// au choix parmi les métriques + indice composite, couleur par dimension.
// L'indice LIEN (/100) combine les 5 métriques normalisées sur les groupes
// visibles : une seule note qui dit « ce lien tient-il ses promesses ».
type TSpaceKey = 'small' | 'rtt' | 'goodput' | 'deadline' | 'cost' | 'indice'
const TSPACE: { key: TSpaceKey; label: string; unit: string; dir: 'down' | 'up' }[] = [
  { key: 'small', label: 'small p95', unit: 'ms', dir: 'down' },
  { key: 'rtt', label: 'RTT p95', unit: 'ms', dir: 'down' },
  { key: 'goodput', label: 'goodput', unit: 'Mb/s', dir: 'up' },
  { key: 'deadline', label: 'échéance', unit: '%', dir: 'up' },
  { key: 'cost', label: 'coût', unit: 'Ar/h', dir: 'down' },
  { key: 'indice', label: 'indice LIEN', unit: '/100', dir: 'up' },
]
const PAL_PROFILE = ['#5ad3e3', '#1fa348', '#f4b400', '#b48ce8', '#e2635c', '#8b9099']
const PAL_QDISC: Record<string, string> = { cake: '#6aa8ff', fq_codel: '#5ad3e3', pfifo_fast: '#8b9099' }
const PAL_CC: Record<string, string> = { cubic: '#e08a4c', bbr: '#4c9be8' }

export default function ResultatsView() {
  const [groups, setGroups] = useState<Group[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [hash8, setHash8] = useState<string>('────────')
  const [pinA, setPinA] = useState<Pinned | null>(null)
  const [pinB, setPinB] = useState<Pinned | null>(null)
  const [rankKey, setRankKey] = useState<string>('small_p95_median')
  const [fProfile, setFProfile] = useState('tous')
  const [fQdisc, setFQdisc] = useState('tous')
  const [fCc, setFCc] = useState('tous')
  const [interpProfile, setInterpProfile] = useState<string | null>(null)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [deltas, setDeltas] = useState<Record<string, { small_p95_pct?: number }>>({})
  // source : '' = EN DIRECT (dernier run gelé), '__all__' = tous runs, sinon run figé
  const [runSel, setRunSel] = useState('')
  const [runIds, setRunIds] = useState<string[]>([])
  const [events, setEvents] = useState<{ ts: string; kind: string; msg: string }[]>([])
  const [caption, setCaption] = useState('')
  // TradeSpace : axes + couleur paramétrables, défaut compromis débit/latence
  const [xKey, setXKey] = useState<TSpaceKey>('goodput')
  const [yKey, setYKey] = useState<TSpaceKey>('small')
  const [colorBy, setColorBy] = useState<'profile' | 'qdisc' | 'cc'>('qdisc')
  const liveSnapRunning = useUIStore((s: any) => !!s.live?.running)
  const scatterRef = useRef<HTMLDivElement>(null)
  const reqSeq = useRef(0)

  // runs gelés, du plus récent au plus ancien (run-smoke et scories exclus)
  const orderedRuns = useMemo(() =>
    asArray<string>(runIds).filter(id => /^run-\d+$/.test(id)).sort().reverse(),
    [runIds])
  const newest = orderedRuns[0] ?? ''
  const effectiveRun = runSel === '__all__' ? '' : (runSel || newest)

  useEffect(() => {
    // garde anti-course : StrictMode rejoue l'effet (2 fetch tous-runs) et
    // le 1er peut répondre après le 2e — seul le dernier demandé s'affiche
    const wanted = effectiveRun
    const seq = ++reqSeq.current
    fetch(`/api/results${wanted ? `?run=${encodeURIComponent(wanted)}` : ''}`).then(r => r.json()).then(j => {
      if (reqSeq.current !== seq) return
      if (j.available && Array.isArray(j.groups)) setGroups(j.groups)
      else if (j.available) setGroups([])
      // run vide (en-tête seule, campagne tuée) : repli sur tous runs plutôt
      // qu'un écran d'erreur — la source reste affichée et explicite
      else if (wanted) setRunSel('__all__')
      else setErr(j.reason || 'pas de résultats')
    }).catch(e => { if (reqSeq.current === seq) setErr(String(e)) })
    fetch('/api/integrity').then(r => r.json()).then(j => {
      // triple provenance: hash8 = sha256(dernier aqm_eval.csv)[:8]; repli run-id
      const id = j?.hash8 ?? String(j?.run_ids?.[0] ?? '').slice(0, 8)
      if (id) setHash8(String(id).slice(0, 8))
    }).catch(() => {})
  }, [effectiveRun])
  useEffect(() => {
    fetch('/api/replay/list').then(r => r.json()).then(j => setRunIds(asArray<string>(j.runs))).catch(() => {})
    fetch('/api/events').then(r => r.json()).then(j => setEvents(asArray<{ts:string;kind:string;msg:string}>(j.events).slice(-20).reverse())).catch(() => {})
  }, [])

  useEffect(() => {
    if (!groups) return
    const id = requestAnimationFrame(() => {
      document.querySelectorAll('.leader-bar').forEach(el => animateBar(el))
    })
    return () => cancelAnimationFrame(id)
  }, [groups])

  useEffect(() => {
    if (!groups || !scatterRef.current) return
    const sg: Group[] = Array.isArray(groups) ? groups : []
    const vis = sg.filter(g =>
      (fProfile === 'tous' || g.profile === fProfile) &&
      (fQdisc === 'tous' || g.qdisc === fQdisc) &&
      (fCc === 'tous' || g.cc === fCc))
    // valeur brute d'une métrique (null = non mesurée, exclue de la moyenne)
    const raw = (g: Group, k: TSpaceKey): number | null => {
      if (k === 'small') { const v = g.small_p95_valid_median ?? g.small_p95_median; return typeof v === 'number' && v > 0 ? v : null }
      if (k === 'rtt') { const v = g.rtt_p95_median; return typeof v === 'number' && v > 0 ? v : null }
      if (k === 'goodput') { const v = g.goodput_median; return typeof v === 'number' && v >= 0 ? v : null }
      if (k === 'deadline') { const v = g.deadline_median ?? g.deadline_ok_pct; return typeof v === 'number' ? v : null }
      if (k === 'cost') {
        const w = g.wasted_median ?? g.wasted_bytes ?? null
        if (w == null || w <= 0) return 0
        return (w / 1073741824) * 5556 * 20
      }
      return null
    }
    // indice LIEN : moyenne des qualités normalisées 0..1 sur les groupes
    // visibles (bonté small/RTT inversée, goodput/échéance directe, coût inversé)
    const ind = (g: Group): number | null => {
      const parts: number[] = []
      const norm = (vals: (number | null)[], v: number | null, down: boolean): number | null => {
        if (v == null) return null
        const xs = vals.filter((x): x is number => x != null)
        if (!xs.length) return null
        const mn = Math.min(...xs), mx = Math.max(...xs)
        if (!(mx > mn)) return 1
        return down ? 1 - (v - mn) / (mx - mn) : (v - mn) / (mx - mn)
      }
      const cols: [TSpaceKey, boolean][] = [['small', true], ['rtt', true], ['goodput', false], ['deadline', false], ['cost', true]]
      for (const [k, down] of cols) {
        const q = norm(vis.map(x => raw(x, k)), raw(g, k), down)
        if (q != null) parts.push(q)
      }
      if (!parts.length) return null
      return (parts.reduce((a, b) => a + b, 0) / parts.length) * 100
    }
    const valOf = (g: Group, k: TSpaceKey): number | null => (k === 'indice' ? ind(g) : raw(g, k))
    const meta = (k: TSpaceKey) => TSPACE.find(t => t.key === k) ?? TSPACE[0]
    const xm = meta(xKey), ym = meta(yKey)
    const pts = vis.map(g => ({ g, x: valOf(g, xKey), y: valOf(g, yKey), name: `${g.profile}·${g.qdisc}·${g.cc}` })).filter(p => p.x != null && p.y != null) as { g: Group; x: number; y: number; name: string }[]
    // couleur par dimension — simple et lisible, une famille à la fois
    const fam = (g: Group): string => colorBy === 'profile' ? g.profile : colorBy === 'qdisc' ? g.qdisc : g.cc
    const famVals = Array.from(new Set(pts.map(p => fam(p.g)))).sort()
    const famColor = (f: string): string => {
      if (colorBy === 'qdisc') return PAL_QDISC[f] ?? '#8b9099'
      if (colorBy === 'cc') return PAL_CC[f] ?? '#8b9099'
      const i = famVals.indexOf(f)
      return PAL_PROFILE[i % PAL_PROFILE.length]
    }
    const c = echarts.init(scatterRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(scatterRef.current)
    const base = baseOption(`${ym.label} / ${xm.label}`, ym.unit)
    // séries par la grammaire — une famille = une série, labels nominatifs
    // anti-collision par-dessus (contrainte testée : grammaire seule)
    const series = famVals.map(fv => {
      const famPts = pts.filter(p => fam(p.g) === fv)
      const s = scatterSeries(fv, famPts.map(p => [p.x, p.y] as [number, number]), famColor(fv),
        famPts.map((p, i) => (p.g.best ? i : -1)).filter(i => i >= 0))
      return {
        ...s,
        label: { show: true, formatter: (par: any) => famPts[par.dataIndex]?.name ?? '', color: famColor(fv), fontSize: 10, fontFamily: 'JetBrains Mono' },
        labelLayout: { hideOverlap: true },
        tooltip: {
          trigger: 'item' as const,
          formatter: (par: any) => {
            const v = par.value as [number, number]
            const who = famPts[par.dataIndex]?.name ?? String(fv)
            return `${who}<br/>${xm.label} : ${v[0].toFixed(1)} ${xm.unit}<br/>${ym.label} : ${v[1].toFixed(1)} ${ym.unit}`
          },
        },
      }
    })
    const opt = {
      ...base,
      // marge droite élargie : le nom de l'axe X ne doit pas être tronqué
      grid: { ...base.grid, right: 96 },
      xAxis: { ...base.xAxis, type: 'value' as const, name: `${xm.label} (${xm.unit})` },
      yAxis: { ...base.yAxis, name: `${ym.label} (${ym.unit})` },
      tooltip: { ...base.tooltip, trigger: 'item' as const },
      series,
    }
    c.setOption(opt as any)
    // légende-phrase : ce que le nuage dit, en une ligne
    if (!pts.length) setCaption('aucun point — élargissez les filtres')
    else {
      const better = ym.dir === 'up'
        ? pts.reduce((a, b) => (b.y > a.y ? b : a))
        : pts.reduce((a, b) => (b.y < a.y ? b : a))
      const bg = better.g
      setCaption(`${ym.label} en fonction de ${xm.label} — ${pts.length} groupes · ${bg.profile}·${bg.qdisc}·${bg.cc} en tête (${better.y.toFixed(1)} ${ym.unit})`)
    }
    return () => { ro.disconnect(); c.dispose() }
  }, [groups, xKey, yKey, colorBy, fProfile, fQdisc, fCc])

  // delta vs run précédent — la dérive temporelle depuis l'historique gelé
  useEffect(() => {
    if (!groups) return
    let cancelled = false
    Promise.all((Array.isArray(groups) ? groups : []).map(async (g) => {
      // cellule 4 parties quand download : up et down ne se comparent jamais
      const dir = g.direction && g.direction !== 'up' ? `|${g.direction}` : ''
      const cell = `${g.profile}|${g.qdisc}|${g.cc}${dir}`
      const r = await fetch(`/api/results/delta?cell=${encodeURIComponent(cell)}`).then(x => x.json()).catch(() => null)
      return [cell, r] as const
    })).then(rows => {
      if (cancelled) return
      const d: Record<string, { small_p95_pct?: number }> = {}
      for (const [cell, r] of rows) {
        if (r?.available) d[cell] = r.delta
      }
      setDeltas(d)
    })
    return () => { cancelled = true }
  }, [groups])

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="error" hint={err} /></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="loading" hint="agrégation des réplications" /></div>

  const rankMeta = RANKS.find(r => r.key === rankKey) ?? RANKS[0]
  // coût au palier unique de référence (5556 Ar/Go) depuis le gaspillage
  // médian gelé — les runs historiques mélangent les paliers (×16,7).
  const costRef = (g: Group): number | null => {
    const w = g.wasted_median ?? g.wasted_bytes ?? null
    if (w == null || w <= 0) return 0
    return (w / 1073741824) * 5556 * 20
  }
  const validN = (g: Group): number => g.small_p95_valid_n ?? g.count
  const val = (g: Group): number => {
    if (rankKey === 'cost') { const c = costRef(g); return c == null ? Number.MAX_SAFE_INTEGER : c }
    const v = (g as any)[rankKey]
    return typeof v === 'number' && v > 0 ? v : Number.MAX_SAFE_INTEGER
  }
  // héros + incertitude du critère (façon DeepSWE : valeur ±, moustaches IC)
  const hero = (g: Group): { v: number; lo: number; hi: number; n: number } => {
    const n = validN(g)
    if (rankKey === 'small_p95_median') {
      const v = g.small_p95_valid_median ?? g.small_p95_median
      const ci = g.small_p95_valid_ci95 ?? g.small_p95_valid_iqr ?? g.small_p95_iqr
      return { v, lo: ci?.[0] ?? v, hi: ci?.[1] ?? v, n }
    }
    if (rankKey === 'rtt_p95_median') {
      const v = g.rtt_p95_median, iqr = g.rtt_p95_iqr
      return { v, lo: iqr?.[0] ?? v, hi: iqr?.[1] ?? v, n }
    }
    if (rankKey === 'goodput_median') return { v: g.goodput_median, lo: g.goodput_median, hi: g.goodput_median, n }
    const c = costRef(g) ?? Number.MAX_SAFE_INTEGER
    return { v: c, lo: c, hi: c, n }
  }
  // groupes sûrs — évite l'écran d'erreur si la charge est inattendue
  const safeGroups: Group[] = Array.isArray(groups) ? groups : []
  const distinct = (k: 'profile' | 'qdisc' | 'cc') => Array.from(new Set(safeGroups.map(g => g[k]))).sort()
  const filtered = safeGroups.filter(g =>
    (fProfile === 'tous' || g.profile === fProfile) &&
    (fQdisc === 'tous' || g.qdisc === fQdisc) &&
    (fCc === 'tous' || g.cc === fCc))
  const ranked = [...filtered].sort((a, b) => rankMeta.dir === 'down' ? val(a) - val(b) : val(b) - val(a))
  const rankMax = Math.max(...filtered.map(g => hero(g).hi).filter(v => Number.isFinite(v) && v > 0), 1)
  const top = ranked[0]
  const baselineRow = filtered.find(g => g.qdisc === 'pfifo_fast' && (!top || g.profile === top.profile))
    ?? [...filtered].sort((a, b) => rankMeta.dir === 'down' ? val(b) - val(a) : val(a) - val(b))[0]
  const diff = top && baselineRow && Number.isFinite(val(top)) && val(baselineRow) > 0
    ? Math.round(((val(baselineRow) - val(top)) / val(baselineRow)) * 100) : null
  // régime perte : toutes les cellules valides affichées ratent l'échéance —
  // la retransmission gouverne la sonde, pas la file. Constaté sur P3 aval
  // (n=3, dl_med 0 partout) : comparer des disciplines indiscernables serait
  // du décor. Le constat le dit une fois, pas une fois par ligne.
  const lossRegime = (() => {
    const ds = filtered
      .filter(g => validN(g) > 0)
      .map(g => g.deadline_median ?? g.deadline_ok_pct ?? null)
      .filter((v): v is number => v != null)
    return ds.length >= 2 && ds.every(d => d === 0)
  })()
  // meilleur par profil sur le critère courant — le verdict s'y compare
  const bestOf = (g: Group): Group => {
    const same = filtered.filter(x => x.profile === g.profile && Number.isFinite(val(x)))
    if (!same.length) return g
    return same.sort((a, b) => rankMeta.dir === 'down' ? val(a) - val(b) : val(b) - val(a))[0]
  }
  // verdict vivant : une phrase qui dit ce que la ligne signifie. En régime
  // perte, toutes les lignes disent la même chose — le verdict factorisé
  // s'affiche une fois au-dessus de la liste, pas N fois dedans.
  const verdict = (g: Group, i: number): string => {
    const n = validN(g)
    if (n <= 0 || !Number.isFinite(val(g))) return 'aucune ligne valide — hors rang, voir quarantaine'
    if (lossRegime) return 'régime perte — la retransmission gouverne, pas la file'
    if (g.profile === 'P3' && (g.small_p95_median ?? 0) > 1500)
      return 'régime perte — la retransmission gouverne, pas la file'
    if (i === 0 && filtered.length > 1) return `référence ${rankMeta.label} — ${n} réplications valides`
    const b = bestOf(g)
    if (b === g) return `meilleur ${g.profile} sur ce critère — ${n} réplications`
    const gap = rankMeta.dir === 'down'
      ? Math.round((val(g) / val(b) - 1) * 100)
      : Math.round((1 - val(g) / val(b)) * 100)
    return gap <= 0 ? `au coude-à-coude avec ${b.qdisc} — ${n} réplications` : `+${gap} % vs ${b.qdisc}, meilleur ${g.profile} — ${n} réplications`
  }
  // verdict factorisé : s'il est identique partout, une seule banderole
  const rowVerdicts = ranked.map((g, i) => verdict(g, i))
  const singleVerdict = ranked.length > 1 && rowVerdicts.every(v => v === rowVerdicts[0]) ? rowVerdicts[0] as string : null

  const chip = (label: string, active: boolean, onClick: () => void) => (
    <button key={label} className="btn" onClick={onClick} style={{
      padding: '2px 8px', fontSize: 10, fontFamily: 'var(--font-mono)',
      border: '1px solid ' + (active ? '#3a3a40' : 'var(--hairline)'),
      background: active ? 'rgba(90,211,227,0.12)' : 'transparent',
      color: active ? '#7fd6e8' : '#a8aeb7',
    }}>{label}</button>
  )
  const shortRun = (id: string) => id.replace(/^run-/, 'run-…').slice(-12)

  return (
    <div className="panel-stack" style={{ position: 'relative' }}>
      <h1 className="view-title">Résultats — classement</h1>

      {/* bandeau benchmark — source, provenance, export (façon DeepSWE) */}
      <div className="card" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 14px', flexWrap: 'wrap' }}>
        <span className="mono" style={{ fontSize: 12, fontWeight: 700 }}>Meteolink Leaderboard</span>
        <span className="mono muted" style={{ fontSize: 11 }}>{safeGroups.length} groupes</span>
        {liveSnapRunning && <span className="mono" style={{ fontSize: 10, color: '#1fa348', border: '1px solid currentColor', padding: '2px 8px' }}>● CAMPAGNE EN COURS</span>}
        <select data-testid="run-select" value={runSel} onChange={e => setRunSel(e.target.value)} title="source des chiffres : en direct = dernier run gelé" style={{ marginLeft: 'auto', background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'JetBrains Mono', fontSize: 11 }}>
          <option value="">{newest ? `⚡ En direct — ${shortRun(newest)}` : '⚡ En direct'}</option>
          <option value="__all__">tous runs (gelés)</option>
          {orderedRuns.map(id => <option key={id} value={id}>{id === newest ? `⚡ ${id}` : id}</option>)}
        </select>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{ border: '1px solid var(--hairline)', padding: '7px 16px' }}>MD</a>
      </div>
      {/* avertissement : ce qu'il faut savoir avant de conclure — court, visible, pas replié */}
      <div className="mono" data-testid="disclaimer" style={{ fontSize: 11, lineHeight: 1.7, color: '#9aa3ad', padding: '8px 14px', border: '1px dashed var(--hairline)', background: 'rgba(255,255,255,0.015)' }}>
        Lecture : <Explain term="valid_only">médianes valid-only</Explain> (n = réplications valides, IC95) ·
        l'<Explain term="mixed_deadline">échéance</Explain> ne se compare qu'à deadline fixée — les runs historiques la mélangent ·
        coût au palier unique 5556 Ar/Go · gel SHA-256, hash {hash8}.
      </div>
      {events.length > 0 && (
        <div className="card" data-testid="changelog">
          <div className="card-head">Changelog — journal opérateur</div>
          <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {events.map((e, i) => <li key={i} className="mono" style={{ fontSize: 11, padding: '3px 0', borderBottom: '1px solid var(--hairline-faint)' }}><span style={{ color: '#767b84' }}>{e.ts}</span> <span style={{ color: '#5ad3e3' }}>{e.kind}</span> {e.msg.slice(0, 120)}</li>)}
          </ul>
        </div>
      )}

      {/* constat de campagne — le verdict en langage opérateur, cliquable pour l'interprétation riche */}
      <button onClick={() => setInterpProfile(filtered[0]?.profile ?? 'P2')} data-testid="constat-button" style={{ all: 'unset', cursor: 'pointer', display: 'block', width: '100%' }}>
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 14px', border: '1px solid ' + (top ? CRAFT.ok : 'var(--hairline)') }}>
          <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-muted)' }}>constat de campagne</span>
          {top && baselineRow && diff != null && !lossRegime && (
            <span className="mono" style={{ fontSize: 12, color: diff > 0 ? CRAFT.ok : '#c3c9d1' }}>
              {top.profile} : {top.qdisc}/{top.cc} protège le trafic critique — {diff > 0 ? `−${diff} %` : `+${Math.abs(diff)} %`} de small p95 vs pfifo
            </span>
          )}
          {top && baselineRow && lossRegime && (
            <span className="mono" style={{ fontSize: 12, color: '#c3c9d1' }}>
              {top.profile} : aucune discipline ne sépare — la perte gouverne, pas la file · cliquez pour la prescription quand même applicable
            </span>
          )}
          {(!top || !baselineRow) && <span className="mono muted" style={{ fontSize: 11 }}>cliquez pour l'interprétation complète</span>}
          <span className="mono" style={{ marginLeft: 'auto', fontSize: 10, color: CRAFT.live }}>interpréter →</span>
        </div>
      </button>
      {interpProfile && <InterpretationView profile={interpProfile} onClose={() => setInterpProfile(null)} />}

      {/* duel critère — avant/après lisible d'un coup d'œil */}
      {top && baselineRow && (
        <div className="card rank-verdict" data-testid="rank-verdict">
          <div>
            <div className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
              pfifo — avant
            </div>
            <div className="mono" style={{ fontSize: 16, fontWeight: 700, color: '#c3c9d1', fontVariantNumeric: 'tabular-nums' }}>
              {val(baselineRow) === Number.MAX_SAFE_INTEGER ? '—' : `${val(baselineRow).toFixed(1)} ${rankMeta.unit}`}
            </div>
            <div className="mono" style={{ fontSize: 10, color: '#9aa0a8' }}>{baselineRow.profile} · n={validN(baselineRow)}v</div>
          </div>
          <div className="mono" data-testid="rank-diff" style={{ fontSize: 24, fontWeight: 700, fontVariantNumeric: 'tabular-nums', color: diff != null && diff > 0 ? '#1fa348' : '#c3c9d1', textAlign: 'center' }}>
            {lossRegime ? 'égalité' : diff != null ? (diff > 0 ? `−${diff} %` : `+${Math.abs(diff)} %`) : '—'}
          </div>
          <div style={{ textAlign: 'right' }}>
            <div className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#7fd6e8' }}>
              1er — {rankMeta.label}
            </div>
            <div className="mono" style={{ fontSize: 16, fontWeight: 700, color: '#1fa348', fontVariantNumeric: 'tabular-nums' }}>
              {val(top) === Number.MAX_SAFE_INTEGER ? '—' : `${val(top).toFixed(1)} ${rankMeta.unit}`}
            </div>
            <div className="mono" style={{ fontSize: 10, color: '#9aa0a8' }}>{top.profile} · {top.qdisc}/{top.cc} · n={validN(top)}v</div>
          </div>
        </div>
      )}

      {/* critère + filtres — une ligne, pas de paragraphe */}
      <div className="form-row" style={{ gap: 6, alignItems: 'center', flexWrap: 'wrap', marginBottom: 8 }}>
        <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>classer par</span>
        {RANKS.map(r => chip(r.label + (r.dir === 'up' ? ' ↑' : ' ↓'), rankKey === r.key, () => setRankKey(r.key)))}
        <span style={{ width: 12 }} />
        {(['tous', ...distinct('profile')] as string[]).map(v => chip(v, fProfile === v, () => setFProfile(v)))}
        {(['tous', ...distinct('qdisc')] as string[]).map(v => chip(v, fQdisc === v, () => setFQdisc(v)))}
        {(['tous', ...distinct('cc')] as string[]).map(v => chip(v, fCc === v, () => setFCc(v)))}
      </div>
      <div className="mono" style={{ fontSize: 10, color: 'var(--text-faint)', marginBottom: 8 }}>
        source : {runSel === '__all__' ? 'tous runs gelés' : <>⚡ en direct · {shortRun(effectiveRun || newest) || '…'}</>} · <Explain term="valid_only">médianes valid-only</Explain>{hash8 !== '────────' ? ` · hash ${hash8}` : ''}
        {liveSnapRunning ? ' — campagne en cours, rafraîchi au gel' : ''}
      </div>
      {singleVerdict && (
        <div className="mono" style={{ fontSize: 11, color: '#c3c9d1', margin: '-2px 0 10px', paddingLeft: 2 }}>{singleVerdict}</div>
      )}

      {filtered.length === 0 && <div className="card"><EmptyState kind="empty" hint="aucun groupe pour ces filtres — élargissez la sélection" /></div>}
      {filtered.length > 0 && <div className="card" style={{ padding: '4px 0' }}>
        {/* leaderboard façon DeepSWE : une métrique-héros + moustaches IC, le reste en sourdine */}
        {ranked.map((g, i) => {
          const h = hero(g)
          const ok = Number.isFinite(h.v) && h.v > 0 && h.v !== Number.MAX_SAFE_INTEGER
          const pct = ok ? Math.min(100, (h.v / rankMax) * 100) : 0
          const barColor = i === 0 ? 'var(--t-ok)' : (QCOLOR[g.qdisc] ?? '#6b7078')
          const key = `${g.profile}/${g.qdisc}/${g.cc}/${g.direction ?? 'up'}`
          const open = expanded === key
          const cellDelta = deltas[`${g.profile}|${g.qdisc}|${g.cc}${g.direction && g.direction !== 'up' ? `|${g.direction}` : ''}`]?.small_p95_pct
          const wasted: number | null = g.wasted_median ?? g.wasted_bytes ?? null
          const cost: number | null = costRef(g)
          const deadlineOk: number | null = g.deadline_median ?? g.deadline_ok_pct ?? null
          const pin = { profile: g.profile, qdisc: g.qdisc, cc: g.cc }
          const isA = pinA?.profile === g.profile && pinA?.qdisc === g.qdisc && pinA?.cc === g.cc
          const isB = pinB?.profile === g.profile && pinB?.qdisc === g.qdisc && pinB?.cc === g.cc
          return (
            <div key={key} style={{ padding: '10px 14px', borderBottom: '1px solid var(--hairline-faint)', background: i === 0 && ok ? 'rgba(31,163,72,0.07)' : 'transparent', cursor: 'pointer' }} onClick={() => setExpanded(open ? null : key)}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <span className="mono" style={{ fontSize: 12, fontWeight: i === 0 ? 700 : 400, color: i === 0 && ok ? '#1fa348' : '#a8aeb7', minWidth: 22 }}>{i + 1}</span>
                <span className="mono" style={{ fontSize: 12, fontWeight: 600, color: '#f2f2f4', minWidth: 150 }}>
                  {g.profile} · {g.qdisc} / {g.cc}
                  {g.direction && g.direction !== 'up' ? <span title="sens download mesuré" style={{ color: '#5ad3e3' }}> ↓</span> : null}
                </span>
                <div style={{ flex: 1, height: 10, position: 'relative', minWidth: 80 }}>
                  <div style={{ position: 'absolute', inset: '3px 0', background: 'var(--hairline-faint)', borderRadius: 2, overflow: 'hidden' }}>
                    {ok && <div className="leader-bar" style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: `${pct}%`, background: barColor, boxShadow: i === 0 ? `0 0 6px ${barColor}` : 'none', transformOrigin: 'left center', borderRadius: 2 }} />}
                  </div>
                  {ok && h.hi > h.lo && (
                    <div title={`IC95 [${h.lo.toFixed(1)}–${h.hi.toFixed(1)}]`} style={{ position: 'absolute', top: 0, bottom: 0, left: `${Math.min(100, (h.lo / rankMax) * 100)}%`, width: `${Math.max(1, Math.min(100, (h.hi / rankMax) * 100) - Math.min(100, (h.lo / rankMax) * 100))}%`, borderLeft: '2px solid #f2f2f4', borderRight: '2px solid #f2f2f4' }} />
                  )}
                </div>
                <span className="mono" style={{ fontSize: 14, fontWeight: 700, color: ok ? (i === 0 ? '#1fa348' : '#f2f2f4') : '#767b84', fontVariantNumeric: 'tabular-nums', minWidth: 120, textAlign: 'right' }}>
                  {ok ? `${h.v.toFixed(1)} ${rankMeta.unit}` : '—'}
                  {ok && h.hi > h.lo && <span style={{ fontSize: 10, fontWeight: 400, color: '#8b9099' }}> ±{((h.hi - h.lo) / 2).toFixed(0)}</span>}
                </span>
                {cellDelta != null && (
                  <span className="mono" title="vs run précédent, même cellule" style={{ fontSize: 10, color: cellDelta <= 0 ? '#1fa348' : '#e22718', fontVariantNumeric: 'tabular-nums' }}>
                    {cellDelta <= 0 ? '↘' : '↗'}{Math.abs(cellDelta)}%
                  </span>
                )}
              </div>
              <div className="mono" style={{ fontSize: 11, color: ok ? '#9aa3ad' : '#767b84', marginTop: 4, marginLeft: 34 }}>
                {!singleVerdict && <>{verdict(g, i)} · </>}
                <span style={{ color: '#5c6169' }}>n={h.n}v · goodput {(g.goodput_median ?? 0).toFixed(1)} Mb/s · échéance {deadlineOk == null ? '—' : deadlineOk.toFixed(0) + '%'} · {wasted == null || wasted <= 0 ? '0 gaspillé' : 'gaspillé'} · {cost == null || cost <= 0 ? '0 Ar' : (cost >= 1000 ? (cost / 1000).toFixed(1) + ' kAr' : cost.toFixed(0) + ' Ar')}</span>
              </div>
              {open && (
                <div className="mono" style={{ fontSize: 11, color: '#9aa3ad', marginTop: 8, marginLeft: 34, padding: '8px 10px', border: '1px solid var(--hairline)', background: 'rgba(255,255,255,0.015)' }}>
                  <div>rtt p95 {fmtIQR(g.rtt_p95_median, g.rtt_p95_iqr)} · small {fmtIQR(g.small_p95_median, g.small_p95_iqr)}{g.small_p95_valid_ci95 ? ` · IC95 valid-only [${g.small_p95_valid_ci95[0].toFixed(1)}–${g.small_p95_valid_ci95[1].toFixed(1)}]` : ''}</div>
                  <div style={{ marginTop: 4 }}>quarantaine {g.quarantined}/{g.count}{g.quarantined > 0 ? ' — voir Provenance' : ''}{g.hardware_recommendation ? ` · ${g.hardware_recommendation}` : ''}</div>
                  <div style={{ marginTop: 6, display: 'flex', gap: 6 }}>
                    <button className="btn" title="épingler comme A" onClick={e => { e.stopPropagation(); setPinA(pin) }} style={{ padding: '2px 8px', fontSize: 10, background: isA ? 'rgba(90,211,227,0.15)' : 'transparent', color: isA ? CRAFT.live : 'var(--text-muted)' }}>A comparer</button>
                    <button className="btn" title="épingler comme B" onClick={e => { e.stopPropagation(); setPinB(pin) }} style={{ padding: '2px 8px', fontSize: 10, background: isB ? 'rgba(31,163,72,0.15)' : 'transparent', color: isB ? CRAFT.ok : 'var(--text-muted)' }}>B comparer</button>
                    <button className="btn" onClick={e => { e.stopPropagation(); setInterpProfile(g.profile) }} style={{ padding: '2px 8px', fontSize: 10 }}>interpréter {g.profile} →</button>
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>}
      {pinA && pinB && (
        <CompareView a={pinA} b={pinB} onClose={() => { setPinA(null); setPinB(null) }} />
      )}
      <div className="card" style={{ padding: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', marginBottom: 6 }}>
          <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>TradeSpace</span>
          <label className="mono" style={{ fontSize: 10, color: '#8b9099' }}>X <select value={xKey} onChange={e => setXKey(e.target.value as TSpaceKey)} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '4px 6px', fontFamily: 'JetBrains Mono', fontSize: 11 }}>
            {TSPACE.map(t => <option key={t.key} value={t.key}>{t.label}</option>)}
          </select></label>
          <label className="mono" style={{ fontSize: 10, color: '#8b9099' }}>Y <select value={yKey} onChange={e => setYKey(e.target.value as TSpaceKey)} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '4px 6px', fontFamily: 'JetBrains Mono', fontSize: 11 }}>
            {TSPACE.map(t => <option key={t.key} value={t.key}>{t.label}</option>)}
          </select></label>
          <span className="mono" style={{ fontSize: 10, color: '#8b9099' }}>couleur</span>
          {(['profile', 'qdisc', 'cc'] as const).map(v => chip({ profile: 'profil', qdisc: 'file', cc: 'CC' }[v], colorBy === v, () => setColorBy(v)))}
          <span className="mono muted" style={{ marginLeft: 'auto', fontSize: 10 }}>hash {hash8}</span>
        </div>
        <div ref={scatterRef} style={{ height: 260 }} />
        <div className="mono" style={{ fontSize: 11, color: '#c3c9d1', marginTop: 6 }}>{caption}</div>
      </div>
      <Provenance source="data/runs/*/aqm_eval.csv" state="live" extra={`${safeGroups.length} groupes · hash ${hash8}`} />
    </div>
  )
}
