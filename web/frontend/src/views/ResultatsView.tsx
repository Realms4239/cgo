import { useEffect, useMemo, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar, animateBarWidth, flashRowUp, prefersReducedMotion } from '../lib/anime'
import { echarts } from '../lib/echarts'
import { baseOption, scatterSeries } from '../lib/chartGrammar'
import CompareView, { type Pinned } from '../components/CompareView'
import { CRAFT } from '../lib/chartGrammar'
import ResultatsInfoModal from '../components/ResultatsInfoModal'
import InterpretationView from '../components/InterpretationView'
import { useUIStore } from '../store/ui'
import { fmtIQR } from '../lib/format'
import { asArray } from '../lib/format'
import { lienScore, components as lienComponents, type LienMetrics } from '../lib/lien'

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

// identité couleur AQM × CC — chaque file une teinte contrastée, chaque CC
// un marqueur : une ligne P1-cake/BBR se distingue de P2-pfifo/cubic d'un
// coup d'œil, sans lire. Teintes hors tricolore identité (DESIGN.md).
const QCOLOR: Record<string, string> = { cake: '#1fd4a0', fq_codel: '#5ad3e3', pfifo_fast: '#9aa3ad' }
const CCMARK: Record<string, string> = { cubic: '◆', bbr: '◇' }

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
const PAL_QDISC: Record<string, string> = { cake: '#1fd4a0', fq_codel: '#5ad3e3', pfifo_fast: '#9aa3ad' }
const PAL_CC: Record<string, string> = { cubic: '#e08a4c', bbr: '#4c9be8' }

// valeurs brutes LIEN d'un groupe (null = non mesurée, exclue de la moyenne)
// — coût au palier unique de référence (5556 Ar/Go), comme costRef ci-bas
const groupLien = (g: Group): LienMetrics => ({
  small: (() => { const v = g.small_p95_valid_median ?? g.small_p95_median; return typeof v === 'number' && v > 0 ? v : null })(),
  rtt: (() => { const v = g.rtt_p95_median; return typeof v === 'number' && v > 0 ? v : null })(),
  goodput: (() => { const v = g.goodput_median; return typeof v === 'number' && v >= 0 ? v : null })(),
  deadline: (() => { const v = g.deadline_median ?? g.deadline_ok_pct; return typeof v === 'number' ? v : null })(),
  cost: (() => { const w = g.wasted_median ?? g.wasted_bytes ?? null; if (w == null) return null; if (w <= 0) return 0; return (w / 1073741824) * 5556 * 20 })(),
})

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
  const [caption, setCaption] = useState('')
  const [infoOpen, setInfoOpen] = useState(false)
  // TradeSpace : Y = le critère du classement (un seul modèle mental),
  // X au choix + couleur paramétrables. L'indice LIEN (/100) combine les 5
  // métriques normalisées sur les groupes visibles.
  const [xKey, setXKey] = useState<TSpaceKey>('goodput')
  const [colorBy, setColorBy] = useState<'profile' | 'qdisc' | 'cc'>('qdisc')
  // Y du TradeSpace = critère du classement — un seul modèle mental
  const yKey: TSpaceKey = ({ small_p95_median: 'small', rtt_p95_median: 'rtt', goodput_median: 'goodput', cost: 'cost' } as Record<string, TSpaceKey>)[rankKey] ?? 'small'
  // X jamais confondu avec Y : repli si le critère de rang devient l'axe X
  const xk: TSpaceKey = xKey === yKey ? (yKey === 'goodput' ? 'cost' : 'goodput') : xKey
  const liveSnapRunning = useUIStore((s: any) => !!s.live?.running)
  const scatterRef = useRef<HTMLDivElement>(null)
  const reqSeq = useRef(0)
  // run touché par l'opérateur ? Le repli tous-runs sur run vide n'est légitime
  // qu'au chargement initial — un choix explicite mérite un écran explicite
  const runTouched = useRef(false)

  // runs gelés, du plus récent au plus ancien (run-smoke et scories exclus)
  const orderedRuns = useMemo(() =>
    asArray<string>(runIds).filter(id => /^run-\d+$/.test(id)).sort().reverse(),
    [runIds])
  // "En direct" = dernier run GELÉ AVEC DONNÉES — un run tué/tout-quarantaine
  // (coquille d'en-tête ou deadline de test) ne doit pas écraser la vue
  // directe d'un écran vide. Le fetch results?run=X sur la coquille répond
  // 0 groupe ; on teste donc à la volée et on descend jusqu'au premier run
  // qui répond. Cache local pour ne pas re-scanner à chaque render.
  const [newest, setNewest] = useState('')
  useEffect(() => {
    if (!orderedRuns.length) return
    let dead = 0
    let stale = false // garde d'annulation : une chaîne lente issue d'un
    // orderedRuns périmé (StrictMode, refetch) ne doit pas écraser la plus fraîche
    const probe = async (i: number) => {
      const id = orderedRuns[i]
      if (!id || stale) return
      try {
        const j = await fetch(`/api/results?run=${encodeURIComponent(id)}`).then(r => r.json())
        if (stale) return
        if (j?.available && Array.isArray(j.groups) && j.groups.length > 0 && j.groups.some((g: any) => (g.small_p95_valid_n ?? g.count) > 0)) setNewest(id)
        else { dead++; if (dead < 12) probe(i + 1) }
      } catch { /* réseau : garde le défaut */ }
    }
    probe(0)
    return () => { stale = true }
  }, [orderedRuns])
  // tant que la sonde n'a pas tranché, NE PAS demander orderedRuns[0] (peut
  // être une coquille vide) : pas de fetch, écran de chargement honnête
  const effectiveRun = runSel === '__all__' ? '' : (runSel || newest || '')

  useEffect(() => {
    // garde anti-course : StrictMode rejoue l'effet (2 fetch tous-runs) et
    // le 1er peut répondre après le 2e — seul le dernier demandé s'affiche
    const wanted = effectiveRun
    const seq = ++reqSeq.current
    fetch(`/api/results${wanted ? `?run=${encodeURIComponent(wanted)}` : ''}`).then(r => r.json()).then(j => {
      if (reqSeq.current !== seq) return
      // run vide (0 groupe : en-tête seule, campagne tuée) = run vide tout
      // autant que !available : au chargement initial → repli tous-runs,
      // choix explicite → écran vide EXPLICITE, jamais de bascule silencieuse
      const empty = !j.available || !(Array.isArray(j.groups) && j.groups.length > 0)
      if (!empty) setGroups(j.groups)
      else if (wanted && !runTouched.current) setRunSel('__all__')
      else if (wanted) setGroups([])
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
  }, [])

  // U6d/U6e : identité stable des barres (valeur animée) + rangs précédents
  // (flash de la ligne remontée — pas de FLIP, le DOM se réordonne)
  const lastWidths = useRef(new Map<string, string>())
  const prevRanks = useRef(new Map<string, number>())
  useEffect(() => {
    if (!groups) return
    const id = requestAnimationFrame(() => {
      try {
        const bars = document.querySelectorAll('.leader-bar')
        if (!bars.length) return
        const seen = new Set<string>()
        bars.forEach((el, idx) => {
          const h = el as HTMLElement
          const lid = h.dataset.leader ?? ''
          const target = h.style.width
          if (!lid) { try { animateBar(el) } catch {} return }
          seen.add(lid)
          const prevIdx = prevRanks.current.get(lid)
          if (prevIdx !== undefined && prevIdx > idx) {
            const row = h.closest('.lb-row')
            if (row) flashRowUp(row)
          }
          try {
            if (prefersReducedMotion()) { h.style.width = target }
            else {
              const prev = lastWidths.current.get(lid)
              if (prev && prev !== target) animateBarWidth(el, prev, target)
              else animateBar(el)
            }
          } catch { try { h.style.width = target } catch {} }
          lastWidths.current.set(lid, target)
          prevRanks.current.set(lid, idx)
        })
        for (const k of [...prevRanks.current.keys()]) {
          if (!seen.has(k)) { prevRanks.current.delete(k); lastWidths.current.delete(k) }
        }
      } catch { /* robustesse echarts-style : l'animation ne casse jamais le rendu */ }
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
      if (k === 'indice') return null // géré par ind() ci-bas
      return groupLien(g)[k]
    }
    // indice LIEN : moyenne des qualités normalisées 0..1 sur les groupes
    // visibles (bonté small/RTT inversée, goodput/échéance directe, coût inversé)
    const lienAll = vis.map(groupLien)
    const ind = (g: Group): number | null => lienScore(lienAll, groupLien(g))
    const valOf = (g: Group, k: TSpaceKey): number | null => (k === 'indice' ? ind(g) : raw(g, k))
    const meta = (k: TSpaceKey) => TSPACE.find(t => t.key === k) ?? TSPACE[0]
    // X auto-informative : si l'axe choisi écrase tous les points au même
    // endroit (variance quasi nulle — vu en prod : goodput 0.3–0.8 sur
    // l'échelle 0–1.4, nuage en ligne verticale), l'axe le plus DISPERSANT
    // prend le relais (effXKey, calculé au niveau composant pour le rendu).
    const xm = meta(effXKey), ym = meta(yKey)
    const pts = vis.map(g => ({ g, x: valOf(g, effXKey), y: valOf(g, yKey), name: `${g.profile}·${g.qdisc}·${g.cc}` })).filter(p => p.x != null && p.y != null) as { g: Group; x: number; y: number; name: string }[]
    // couleur par dimension — simple et lisible, une famille à la fois
    const fam = (g: Group): string => colorBy === 'profile' ? g.profile : colorBy === 'qdisc' ? g.qdisc : g.cc
    const famVals = Array.from(new Set(pts.map(p => fam(p.g)))).sort()
    const famColor = (f: string): string => {
      if (colorBy === 'qdisc') return PAL_QDISC[f] ?? '#8b9099'
      if (colorBy === 'cc') return PAL_CC[f] ?? '#8b9099'
      const i = famVals.indexOf(f)
      return PAL_PROFILE[i % PAL_PROFILE.length]
    }
    // top-3 par Y — seuls eux portent un label, NOM COURT (qdisc seul : la
    // couleur encode déjà la famille, le nom long P·q/cc recouvre les points
    // voisins), position DÉCALÉE au-dessus du point, halo sombre 3px. Un point
    // sans label n'est pas muet : le tooltip et la légende le nomment.
    const better = (p: { y: number }) => ym.dir === 'up' ? p.y : -p.y
    const nFam = new Set(pts.map(p => fam(p.g))).size
    const labeled = [...pts].sort((a, b) => better(b) - better(a))
      .slice(0, Math.max(1, Math.min(3, nFam)))
    const c = echarts.init(scatterRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(scatterRef.current)
    const base = baseOption(`${ym.label} / ${xm.label}`, ym.unit)
    // séries par la grammaire — une famille = une série, labels nominatifs
    // anti-collision par-dessus (contrainte testée : grammaire seule)
    const series = famVals.map(fv => {
      const famPts = pts.filter(p => fam(p.g) === fv)
      const bestIdx = famPts.map((p, i) => (p.g.best ? i : -1)).filter(i => i >= 0)
      const s = scatterSeries(fv, famPts.map(p => [p.x, p.y] as [number, number]), famColor(fv), bestIdx)
      return {
        ...s,
        label: {
          show: true,
          // nom court (qdisc/cc) — le nom long P·q/cc recouvre les voisins
          formatter: (par: any) => labeled.includes(famPts[par.dataIndex]) ? famPts[par.dataIndex].g.qdisc : '',
          position: 'top' as const,
          distance: 8,
          color: famColor(fv),
          fontSize: 11,
          fontFamily: 'JetBrains Mono',
          textBorderColor: '#070707',
          textBorderWidth: 3,
        },
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
      grid: { ...base.grid, left: 56, right: 28, top: 20, bottom: 44 },
      xAxis: { ...base.xAxis, type: 'value' as const },
      // pas de nom d'axe : la légende-phrase sous le graphe le dit
      yAxis: { ...base.yAxis, name: '' },
      tooltip: { ...base.tooltip, trigger: 'item' as const },
      series,
    }
    c.setOption(opt as any)
    // légende-phrase : axes + ce que le nuage dit, en une ligne sous le graphe
    if (!pts.length) setCaption('aucun point — élargissez les filtres')
    else {
      const bp = [...pts].sort((a, b) => better(b) - better(a))[0]
      setCaption(`Y : ${ym.label} (${ym.unit}) · X : ${xm.label} (${xm.unit}) — ${pts.length} groupes · ${bp.g.profile}·${bp.g.qdisc}·${bp.g.cc} en tête (${bp.y.toFixed(1)} ${ym.unit})`)
    }
    return () => { ro.disconnect(); c.dispose() }
  }, [groups, xk, yKey, colorBy, fProfile, fQdisc, fCc])

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
    // null = jamais mesuré (exclu du score, hors rang) ; 0 = mesuré, rien
    // gaspillé (le moins cher possible). Les confondre faussait le rang coût.
    if (w == null) return null
    if (w <= 0) return 0
    return (w / 1073741824) * 5556 * 20
  }
  const validN = (g: Group): number => g.small_p95_valid_n ?? g.count
  const val = (g: Group): number => {
    if (rankKey === 'cost') { const c = costRef(g); return c == null ? Number.MAX_SAFE_INTEGER : c }
    const v = (g as any)[rankKey]
    // goodput mesuré à 0 (lien mort) ≠ donnée absente : 0 se classe dernier,
    // l'absence sort du rang. Latences : seules les valeurs > 0 existent.
    if (rankKey === 'goodput_median') return typeof v === 'number' && v >= 0 ? v : Number.MAX_SAFE_INTEGER
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
  // qualités LIEN sur les groupes visibles — même échelle que le TradeSpace
  const lienAll: LienMetrics[] = filtered.map(groupLien)
  const ranked = [...filtered].sort((a, b) => rankMeta.dir === 'down' ? val(a) - val(b) : val(b) - val(a))
  // échelle de rang : la barre mesure la BONTÉ sur la plage visible
  // (min→max des valeurs finies), pas la valeur brute sur [0–max] où tout
  // semble plein. Le meilleur a toujours la barre la plus longue, dans les
  // deux sens de critère ; les moustaches gardent la position absolue.
  const finiteVals = rankKey === 'goodput_median'
    ? filtered.map(val).filter(v => Number.isFinite(v) && v >= 0 && v !== Number.MAX_SAFE_INTEGER)
    : filtered.map(val).filter(v => Number.isFinite(v) && v > 0 && v !== Number.MAX_SAFE_INTEGER)
  const rankMin = finiteVals.length ? Math.min(...finiteVals) : 0
  const rankMax = finiteVals.length ? Math.max(...finiteVals) : 1
  const rankSpan = rankMax > rankMin ? rankMax - rankMin : 0
  const rankPos = (v: number): number => rankSpan > 0
    ? Math.min(100, Math.max(0, ((v - rankMin) / rankSpan) * 100))
    : 100
  const rankFill = (v: number): number => rankMeta.dir === 'down' ? 100 - rankPos(v) : rankPos(v)
  const top = ranked[0]
  // la référence dit son vrai nom : pfifo quand présent, sinon le pire du
  // filtre (le duel reste honnête : "cake — avant" si l'opérateur a filtré
  // les files sur cake seul). baselineIsPfifo décide du libellé partout.
  const pfifoRef = filtered.find(g => g.qdisc === 'pfifo_fast' && validN(g) > 0 && (!top || g.profile === top.profile))
  // candidats à données valides UNIQUEMENT : comparer contre une ligne à 0
  // mesure valide affichait "−100 % ↓" face au vide (vu en prod, filtre cake).
  const validCands = filtered.filter(g => validN(g) > 0)
  const baselineRow = pfifoRef
    ?? [...validCands].sort((a, b) => rankMeta.dir === 'down' ? val(b) - val(a) : val(a) - val(b))[0]
  const baselineIsPfifo = !!pfifoRef
  // duel seulement si les DEUX côtés ont des mesures valides — sinon rien à comparer
  const duelReady = !!top && !!baselineRow && validN(top) > 0 && validN(baselineRow) > 0
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
    <button key={label} className="btn" onClick={onClick} title={label} style={{
      padding: '3px 10px', fontSize: 11, fontFamily: 'var(--font-mono)',
      border: '1px solid ' + (active ? '#3a3a40' : 'var(--hairline)'),
      background: active ? 'rgba(90,211,227,0.12)' : 'transparent',
      color: active ? '#7fd6e8' : '#a9aeb6',
    }}>{label}</button>
  )
  // méta source : le run qui nourrit les chiffres, pas un compte muet
  const validTotal = safeGroups.reduce((a, g) => a + validN(g), 0)
  const metaRun = effectiveRun || 'tous runs'
  const shortRun = (id: string) => id.replace(/^run-/, 'run-…').slice(-12)
  // coin du TradeSpace : où vit l'optimum (down = plus bas = mieux)
  const yd = (TSPACE.find(t => t.key === yKey)?.dir ?? 'down') === 'down'
  // l'axe X affiché peut différer du choix si le choix était dégénéré
  // (variance quasi nulle) — la flèche du coin suit l'axe RÉELlement tracé
  const xSp = (k: TSpaceKey): number => {
    const xs = filtered.map(g => (k === 'indice' ? null : (g as any)[{ small: 'small_p95_valid_median', rtt: 'rtt_p95_median', goodput: 'goodput_median', deadline: 'deadline_median', cost: 'cost_median' }[k] ?? ''])).filter((v): v is number => typeof v === 'number' && v > 0)
    if (xs.length < 2) return 0
    const mx = Math.max(...xs.map(Math.abs), 1e-9)
    return (Math.max(...xs) - Math.min(...xs)) / mx
  }
  const effXKey: TSpaceKey = xSp(xk) < 0.18
    ? (TSPACE.filter(t => t.key !== yKey && t.key !== 'indice').sort((a, b) => xSp(b.key) - xSp(a.key))[0]?.key ?? xk)
    : xk
  const xdEff = (TSPACE.find(t => t.key === effXKey)?.dir ?? 'down') === 'down'
  const cornerArrow = yd ? (xdEff ? '↙' : '↘') : (xdEff ? '↖' : '↗')

  return (
    <div className="panel-stack" style={{ position: 'relative' }}>
      <h1 className="view-title">Résultats — classement</h1>

      {/* bandeau benchmark — source, provenance, export (façon DeepSWE) */}
      <div className="card" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 14px', flexWrap: 'wrap' }}>
        <span className="mono" style={{ fontSize: 12, fontWeight: 500 }}>Meteolink Leaderboard</span>
        <span className="mono muted" style={{ fontSize: 11 }}>{safeGroups.length} groupes · {metaRun} · {validTotal} valides</span>
        {liveSnapRunning && <span className="mono" style={{ fontSize: 10, color: '#1fa348', border: '1px solid currentColor', padding: '2px 8px' }}>● CAMPAGNE EN COURS</span>}
        <select data-testid="run-select" value={runSel} onChange={e => { runTouched.current = true; setRunSel(e.target.value) }} title="source des chiffres : en direct = dernier run gelé" style={{ marginLeft: 'auto', background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'JetBrains Mono', fontSize: 11 }}>
          <option value="">{newest ? `⚡ En direct — ${shortRun(newest)}` : '⚡ En direct'}</option>
          <option value="__all__">tous runs (gelés)</option>
          {orderedRuns.map(id => <option key={id} value={id}>{id === newest ? `⚡ ${id}` : id}</option>)}
        </select>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{ border: '1px solid var(--hairline)', padding: '7px 16px' }}>MD</a>
        <button className="btn" title="Lire le classement : métriques et règles de lecture" aria-label="Lire le classement" onClick={() => setInfoOpen(true)} style={{ padding: '7px 12px', fontSize: 13 }}>ⓘ</button>
      </div>
      {infoOpen && <ResultatsInfoModal onClose={() => setInfoOpen(false)} />}

      {/* constat de campagne — le verdict en langage opérateur, l'action
          d'interprétation est un vrai bouton, plus une carte-cible géante */}
      <div className="card" style={{ padding: '10px 14px', border: '1px solid ' + (top ? CRAFT.ok : 'var(--hairline)') }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-muted)' }}>constat de campagne</span>
          <button className="btn" data-testid="constat-button" onClick={() => setInterpProfile(filtered[0]?.profile ?? 'P2')} style={{ marginLeft: 'auto', padding: '4px 12px', fontSize: 10 }}>Interprétation complète →</button>
        </div>
        <p className="mono" style={{ margin: '8px 0 0', fontSize: 13, lineHeight: 1.6, color: '#c3c9d1', maxWidth: '68ch' }}>
          {duelReady && lossRegime
            ? 'P3 : aucune discipline ne sépare — la retransmission gouverne, pas la file (échéance 0 % partout)'
            : duelReady && diff != null
              ? `${top.profile} : ${top.qdisc}/${top.cc} protège le trafic critique — ${diff > 0 ? `−${diff} %` : `+${Math.abs(diff)} %`} de small p95 vs ${baselineIsPfifo ? 'pfifo' : (baselineRow as Group).qdisc} (n=${validN(top as Group)} réplications valides)`
              : 'aucun groupe comparable — lancez une campagne ou élargissez les filtres'}
        </p>
      </div>
      {interpProfile && <InterpretationView profile={interpProfile} onClose={() => setInterpProfile(null)} />}

      {/* duel critère — référence vs 1er : deux colonnes, delta énorme au
          centre (retour à la mise en page de référence) */}
      {duelReady && (() => {
        const bv = val(baselineRow), tv = val(top)
        const bOk = Number.isFinite(bv) && bv !== Number.MAX_SAFE_INTEGER
        const tOk = Number.isFinite(tv) && tv !== Number.MAX_SAFE_INTEGER
        const col = (label: string, ok: boolean, v: number, n: number, color: string) => (
          <div style={{ flex: 1, minWidth: 180, textAlign: 'center' }}>
            <div className="mono" style={{ fontSize: 12, letterSpacing: '0.06em', textTransform: 'uppercase', color: '#a9aeb6', marginBottom: 6 }}>{label}</div>
            <div className="mono" style={{ fontSize: 26, fontWeight: 600, fontVariantNumeric: 'tabular-nums', color, lineHeight: 1 }}>
              {ok ? v.toFixed(1) : '—'}
            </div>
            <div className="mono" style={{ fontSize: 12, color: '#a9aeb6', marginTop: 4 }}>{rankMeta.unit} · {n} mesures valides</div>
          </div>
        )
        return (
          <div className="card rank-verdict" data-testid="rank-verdict" style={{ display: 'flex', alignItems: 'center', gap: 18, flexWrap: 'wrap', padding: '18px 20px' }}>
            {col(`${baselineIsPfifo ? 'pfifo' : baselineRow.qdisc} — avant`, bOk, bv, validN(baselineRow), '#a9aeb6')}
            <div style={{ textAlign: 'center' }}>
              <div className="mono" data-testid="rank-diff" style={{ fontSize: 34, fontWeight: 600, fontVariantNumeric: 'tabular-nums', color: diff != null && diff > 0 ? '#1fa348' : '#d6d8dd', lineHeight: 1 }}>
                {lossRegime ? 'égalité' : diff != null ? (diff > 0 ? `−${diff} %` : `+${Math.abs(diff)} %`) : '—'}
              </div>
              {!lossRegime && diff != null && diff > 0 && (
                <div className="mono" style={{ fontSize: 13, color: '#1fa348', marginTop: 4, lineHeight: 1 }}>↓</div>
              )}
              <div className="mono" style={{ fontSize: 11, color: '#a9aeb6', marginTop: 6, maxWidth: 130 }}>{rankMeta.label}</div>
            </div>
            {col(`1er — ${top.qdisc}/${top.cc}`, tOk, tv, validN(top), diff != null && diff > 0 ? '#1fa348' : '#d6d8dd')}
          </div>
        )
      })()}

      {/* critère + filtres — une ligne, pas de paragraphe */}
      <div className="form-row" style={{ gap: 6, alignItems: 'center', flexWrap: 'wrap', marginBottom: 8 }}>
        <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>classer par</span>
        {RANKS.map(r => chip(r.label + (r.dir === 'up' ? ' ↑' : ' ↓'), rankKey === r.key, () => setRankKey(r.key)))}
        <span style={{ width: 12 }} />
        {(['tous', ...distinct('profile')] as string[]).map(v => chip(v, fProfile === v, () => setFProfile(v)))}
        {(['tous', ...distinct('qdisc')] as string[]).map(v => chip(v, fQdisc === v, () => setFQdisc(v)))}
        {(['tous', ...distinct('cc')] as string[]).map(v => chip(v, fCc === v, () => setFCc(v)))}
        {/* clé couleur — chaque file sa teinte, chaque congestion son signe */}
        <span style={{ width: 12 }} />
        <span className="mono" style={{ fontSize: 11, color: '#a9aeb6' }} title="chaque discipline de file a sa couleur, chaque congestion son signe">
          <span style={{ color: QCOLOR.pfifo_fast }}>■</span> pfifo <span style={{ color: QCOLOR.fq_codel }}>■</span> fq_codel <span style={{ color: QCOLOR.cake }}>■</span> cake · ◆ cubic ◇ bbr
        </span>
      </div>
      {singleVerdict && (
        <div className="mono" style={{ fontSize: 11, color: '#c3c9d1', margin: '2px 0 10px', paddingLeft: 2 }}>{singleVerdict}</div>
      )}

      {filtered.length === 0 && <div className="card"><EmptyState kind="empty" hint={safeGroups.length === 0 && effectiveRun ? 'run vide ou supprimé — passez à tous runs' : 'aucun groupe pour ces filtres — élargissez la sélection'} /></div>}
      {filtered.length > 0 && <div className="card" style={{ padding: '4px 0' }}>
        {/* leaderboard façon DeepSWE : une métrique-héros + moustaches IC, le reste en sourdine */}
        {ranked.map((g, i) => {
          const h = hero(g)
          const ok = rankKey === 'goodput_median'
            ? Number.isFinite(h.v) && h.v >= 0 && h.v !== Number.MAX_SAFE_INTEGER
            : Number.isFinite(h.v) && h.v > 0 && h.v !== Number.MAX_SAFE_INTEGER
          const pct = ok ? rankFill(h.v) : 0
          const barColor = QCOLOR[g.qdisc] ?? '#6b7078'
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
            <div key={key} className="lb-row" role="button" tabIndex={0} aria-expanded={open} aria-label={`${g.profile} ${g.qdisc} ${g.cc} — détails`} style={{ padding: '10px 14px', borderBottom: '1px solid var(--hairline-faint)', background: 'transparent', cursor: 'pointer', opacity: expanded && !open ? 0.35 : 1, filter: expanded && !open ? 'saturate(0.5)' : 'none', transition: 'opacity 250ms ease, filter 250ms ease' }} onClick={() => setExpanded(open ? null : key)} onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setExpanded(open ? null : key) } }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <span className="mono" style={{ fontSize: 11, fontWeight: 400, opacity: 0.4, color: '#a8aeb7', minWidth: 22, textAlign: 'right' }}>{i + 1}</span>
                <span className="mono" style={{ fontSize: 13, fontWeight: 500, color: '#f2f2f4', minWidth: 170 }}>
                  {g.profile} · {g.qdisc} / {g.cc} <span title={g.cc === 'bbr' ? 'contrôle de congestion BBR (Google) — sonde le débit, peu agressif' : 'contrôle de congestion CUBIC — remplit les files, agressif'} style={{ color: barColor }}>{CCMARK[g.cc] ?? ''}</span>
                  {g.direction && g.direction !== 'up' ? <span title="sens download mesuré" style={{ color: '#5ad3e3' }}> ↓</span> : null}
                </span>
                <div style={{ flex: 1, height: 14, position: 'relative', minWidth: 80 }}>
                  <div style={{ position: 'absolute', inset: '2px 0', background: 'rgba(255,255,255,0.04)', borderRadius: 2, overflow: 'hidden' }}>
                    {ok && <div className="leader-bar" data-leader={key} style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: `${pct}%`, background: barColor, transformOrigin: 'left center', borderRadius: 2 }} />}
                  </div>
                  {ok && h.hi > h.lo && (
                    <div title={`IC95 [${h.lo.toFixed(1)}–${h.hi.toFixed(1)}]`} style={{ position: 'absolute', top: '50%', marginTop: -0.5, height: 1, left: `${rankPos(h.lo)}%`, width: `${Math.max(1.5, rankPos(h.hi) - rankPos(h.lo))}%`, borderLeft: '1px solid rgba(242,242,244,0.7)', borderRight: '1px solid rgba(242,242,244,0.7)', background: 'rgba(242,242,244,0.7)' }} />
                  )}
                </div>
                <span className="mono" style={{ fontSize: 13, fontWeight: 500, color: ok ? '#f2f2f4' : '#767b84', fontVariantNumeric: 'tabular-nums', minWidth: 110, textAlign: 'right' }}>
                  {ok ? `${h.v.toFixed(1)} ${rankMeta.unit}` : '—'}
                  {ok && h.hi > h.lo && <span style={{ fontSize: 10, fontWeight: 400, color: '#8b9099' }}> ±{((h.hi - h.lo) / 2).toFixed(0)}</span>}
                </span>
                {cellDelta != null && (
                  <span className="mono" title="vs run précédent, même cellule" style={{ fontSize: 10, color: cellDelta <= 0 ? '#1fa348' : '#e22718', fontVariantNumeric: 'tabular-nums' }}>
                    {cellDelta <= 0 ? '↘' : '↗'}{Math.abs(cellDelta)}%
                  </span>
                )}
              </div>
              <div className="mono" style={{ fontSize: 12, color: ok ? '#a9aeb6' : '#767b84', marginTop: 6, marginLeft: 34 }}>
                {!singleVerdict && <span style={{ color: '#c3c9d1' }}>{verdict(g, i)} · </span>}
                <span>{h.n} mesures valides · débit {(g.goodput_median ?? 0).toFixed(1)} Mb/s · échéances respectées {deadlineOk == null ? '—' : deadlineOk.toFixed(0) + '%'} · {wasted == null || wasted <= 0 ? 'rien gaspillé' : (wasted >= 1048576 ? (wasted / 1048576).toFixed(1) + ' MiB gaspillés' : wasted >= 1024 ? (wasted / 1024).toFixed(0) + ' Kio gaspillés' : wasted + ' o gaspillés')} · {cost == null || cost <= 0 ? '0 Ar' : (cost >= 1000 ? (cost / 1000).toFixed(1) + ' kAr' : cost.toFixed(0) + ' Ar')}</span>
              </div>
              {open && (
                <div style={{ marginTop: 10, marginLeft: 34, padding: '12px 14px', border: '1px solid var(--hairline)', background: 'rgba(255,255,255,0.015)' }}>
                  {/* mesures en langue opérateur — chaque ligne dit ce que c'est,
                      la valeur, et ce que ça signifie. Fini le dialecte CSV. */}
                  <div style={{ display: 'grid', gap: 8 }}>
                    <div className="mono" style={{ fontSize: 12, color: '#d6d8dd', fontVariantNumeric: 'tabular-nums' }}>
                      Latence du lien <span style={{ color: '#8b9099' }}>(RTT p95)</span> : <b style={{ fontWeight: 600 }}>{fmtIQR(g.rtt_p95_median, g.rtt_p95_iqr)}</b> ms
                      <span style={{ color: '#a9aeb6' }}> — temps d'aller-retour des paquets sous charge</span>
                    </div>
                    <div className="mono" style={{ fontSize: 12, color: '#d6d8dd', fontVariantNumeric: 'tabular-nums' }}>
                      Réactivité des petits objets <span style={{ color: '#8b9099' }}>(small p95)</span> : <b style={{ fontWeight: 600 }}>{fmtIQR(g.small_p95_median, g.small_p95_iqr)}</b> ms
                      <span style={{ color: '#a9aeb6' }}> — charger une page/mesure pendant un transfert lourd</span>
                    </div>
                    {g.small_p95_valid_ci95 && (
                      <div className="mono" style={{ fontSize: 12, color: '#a9aeb6', fontVariantNumeric: 'tabular-nums' }}>
                        Marge d'incertitude : {g.small_p95_valid_ci95[0].toFixed(0)}–{g.small_p95_valid_ci95[1].toFixed(0)} ms — fourchette où se trouve la vraie valeur 95 fois sur 100
                      </div>
                    )}
                    <div className="mono" style={{ fontSize: 12, color: g.quarantined > 0 ? '#f4b400' : '#a9aeb6' }}>
                      {g.quarantined > 0
                        ? `${g.quarantined} mesure(s) sur ${g.count} écartée(s) — incohérentes, exclues du calcul (détail : vue Provenance)`
                        : `les ${g.count} mesures sont cohérentes — aucune écartée`}
                    </div>
                    {g.hardware_recommendation && (
                      <div className="mono" style={{ fontSize: 12, color: '#5ad3e3', lineHeight: 1.5 }}>
                        Recommandation terrain : {g.hardware_recommendation}
                      </div>
                    )}
                  </div>
                  {(() => {
                    const comp = lienComponents(lienAll, groupLien(g))
                    const score = lienScore(lienAll, groupLien(g))
                    if (score == null) return null
                    const rows: { label: string; q: number | null; color: string }[] = [
                      { label: 'small p95', q: comp.small, color: '#f4b400' },
                      { label: 'RTT', q: comp.rtt, color: '#5ad3e3' },
                      { label: 'goodput', q: comp.goodput, color: '#b48ae0' },
                      { label: 'échéances', q: comp.deadline, color: '#1fa348' },
                      { label: 'coût', q: comp.cost, color: '#9aa3ad' },
                    ]
                    return (
                      <div data-testid="lien-score" style={{ marginTop: 10 }}>
                        <div className="mono" style={{ fontSize: 13, color: '#d6d8dd' }}>
                          Score LIEN : {Math.round(score)}/100 — comment il se décompose
                        </div>
                        <div style={{ display: 'grid', gap: 6, marginTop: 8 }}>
                          {rows.map(r => (
                            <div key={r.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                              <span className="mono" style={{ fontSize: 12, color: '#a9aeb6', minWidth: 80 }}>{r.label}</span>
                              <div style={{ flex: 1, height: 6, background: 'rgba(255,255,255,0.06)', borderRadius: 3, overflow: 'hidden' }}>
                                <div style={{ height: '100%', width: `${r.q == null ? 0 : Math.round(r.q * 100)}%`, background: r.color, borderRadius: 3 }} />
                              </div>
                              <span className="mono" style={{ fontSize: 12, color: '#a9aeb6', minWidth: 92, textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                                {r.q == null ? '— · 20 %' : `${Math.round(r.q * 100)}/100 · 20 %`}
                              </span>
                            </div>
                          ))}
                        </div>
                        <div className="mono" style={{ fontSize: 12, color: '#a9aeb6', marginTop: 8 }}>
                          Moyenne de 5 critères ramenés à l'échelle des groupes visibles (0 = pire visible, 100 = meilleur visible), 20 % chacun : réactivité, latence, débit, échéances, coût.
                        </div>
                      </div>
                    )
                  })()}
                  <div style={{ marginTop: 10, display: 'flex', gap: 6 }}>
                    <button className="btn" title="épingler comme A" onClick={e => { e.stopPropagation(); setPinA(pin) }} style={{ padding: '4px 10px', fontSize: 10, background: isA ? 'rgba(90,211,227,0.15)' : 'transparent', color: isA ? CRAFT.live : 'var(--text-muted)' }}>A comparer</button>
                    <button className="btn" title="épingler comme B" onClick={e => { e.stopPropagation(); setPinB(pin) }} style={{ padding: '4px 10px', fontSize: 10, background: isB ? 'rgba(31,163,72,0.15)' : 'transparent', color: isB ? CRAFT.ok : 'var(--text-muted)' }}>B comparer</button>
                    <button className="btn" onClick={e => { e.stopPropagation(); setInterpProfile(g.profile) }} style={{ padding: '4px 10px', fontSize: 10 }}>interpréter {g.profile} →</button>
                  </div>
                </div>
              )}
            </div>
          )
        })}
        {/* échelle partagée sous la dernière rangée — les barres partagent
            UNE échelle implicite (bonté : 100 % = meilleur visible).
            Sans ligne valide (coquille/quarantaine), pas d'échelle à lire. */}
        {finiteVals.length > 0 && (() => {
          const worst = rankMeta.dir === 'down' ? rankMax : rankMin
          const best = rankMeta.dir === 'down' ? rankMin : rankMax
          const fmt = (v: number) => rankSpan > 0 ? `${v.toFixed(1)} ${rankMeta.unit}` : '—'
          return (
            <div className="mono" style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 14px 12px', borderTop: '1px solid var(--hairline-faint)', fontSize: 11, color: '#a9aeb6' }}>
              <span>← {fmt(worst)}</span><span>25 %</span><span>50 %</span><span>75 %</span><span>{fmt(best)} →</span>
            </div>
          )
        })()}
      </div>}
      {pinA && pinB && (
        <CompareView a={pinA} b={pinB} onClose={() => { setPinA(null); setPinB(null) }} />
      )}
      <div className="card" style={{ padding: 12, position: 'relative' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', marginBottom: 6 }}>
          <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>TradeSpace</span>
          <span className="mono" title="Y = le critère du classement" style={{ fontSize: 10, color: '#5ad3e3' }}>Y {TSPACE.find(t => t.key === yKey)?.label}</span>
          <span className="mono" style={{ fontSize: 10, color: '#8b9099' }}>X</span>
          {TSPACE.filter(t => t.key !== yKey).map(t => chip(t.label, xk === t.key, () => setXKey(t.key)))}
          <span className="mono" style={{ fontSize: 10, color: '#8b9099' }}>couleur</span>
          {(['profile', 'qdisc', 'cc'] as const).map(v => chip({ profile: 'profil', qdisc: 'file', cc: 'CC' }[v], colorBy === v, () => setColorBy(v)))}
          <span className="mono muted" style={{ marginLeft: 'auto', fontSize: 10 }}>hash {hash8}</span>
        </div>
        <div style={{ position: 'relative' }}>
          <span className="mono" style={{ position: 'absolute', top: 2, right: 4, fontSize: 10, color: 'var(--text-faint)', pointerEvents: 'none', zIndex: 1 }}>
            optimal {cornerArrow}
          </span>
          <div ref={scatterRef} style={{ height: 320 }} />
        </div>
        <div className="mono" style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 6 }}>{caption}</div>
      </div>
      <Provenance source="data/runs/*/aqm_eval.csv" state="live" extra={`${safeGroups.length} groupes · hash ${hash8}`} />
    </div>
  )
}
