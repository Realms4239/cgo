import { Fragment, useEffect, useMemo, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar, animateBarWidth, flashRowUp, prefersReducedMotion } from '../lib/anime'
import { echarts } from '../lib/echarts'
import { baseOption, paretoFrontier, scatterSeries } from '../lib/chartGrammar'
import CompareView, { type Pinned } from '../components/CompareView'
import { CRAFT } from '../lib/chartGrammar'
import ResultatsInfoModal from '../components/ResultatsInfoModal'
import InterpretationView from '../components/InterpretationView'
import { useUIStore } from '../store/ui'
import { fmtIQR } from '../lib/format'
import { asArray, improve, shouldLogScale } from '../lib/format'
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
  { key: 'deadline_median', label: 'échéance', dir: 'up' as const, unit: '%', term: 'deadline_ok' },
  { key: 'cost', label: 'coût', dir: 'down' as const, unit: 'Ar/h', term: 'cost_ar_per_h' },
]

// décodage config en mots simples — mêmes faits que lib/explain.ts,
// raccourcis pour la sous-ligne d'identité (une idée : ce que c'est).
const PROF_SHORT: Record<string, string> = {
  P1: 'fibre 80 Mb/s, 20 ms', P2: '4G chargée 20 Mb/s, 100 ms',
  P3: 'VSAT 5 Mb/s, 600 ms', P4: 'Starlink 100 Mb/s, 40 ms',
}
const Q_SHORT: Record<string, string> = {
  cake: 'file qui protège le petit trafic', fq_codel: 'file qui reste courte',
  pfifo_fast: 'file simple, sans protection',
}
const CC_SHORT: Record<string, string> = { cubic: 'remplit les files', bbr: 'garde les files courtes' }
const decodeSub = (g: { profile: string; qdisc: string; cc: string }): string =>
  `${PROF_SHORT[g.profile] ?? g.profile} · ${Q_SHORT[g.qdisc] ?? g.qdisc} · ${CC_SHORT[g.cc] ?? g.cc}`

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
  // pré-gel (aucun gel) = écran neutre, jamais "erreur" : le genre d'erreur
  // suit la cause — raison de vide → empty, fetch qui jette → error.
  const [errKind, setErrKind] = useState<'empty' | 'error'>('empty')
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
  const yKey: TSpaceKey = ({ small_p95_median: 'small', rtt_p95_median: 'rtt', goodput_median: 'goodput', deadline_median: 'deadline', cost: 'cost' } as Record<string, TSpaceKey>)[rankKey] ?? 'small'
  // X jamais confondu avec Y : repli si le critère de rang devient l'axe X
  const xk: TSpaceKey = xKey === yKey ? (yKey === 'goodput' ? 'cost' : 'goodput') : xKey
  // synchro X — l'état suit l'affichage : si le critère de rang atterrit sur
  // X, X migre vers le repli (plus de substitution silencieuse au rendu)
  useEffect(() => {
    if (xKey === yKey) setXKey(yKey === 'goodput' ? 'cost' : 'goodput')
  }, [xKey, yKey])
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
  // nœuds de l'enveloppe TradeSpace — le rail droit reflète le tracé
  const [frontier, setFrontier] = useState<{ name: string; x: number; y: number }[]>([])
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

  // refresh — charge les groupes (+hash). quiet=true (polling campagne) :
  // jamais de bascule silencieuse ni d'erreur écrasante — on garde les
  // chiffres affichés si le fetch rate ou revient vide entre deux cellules.
  const refresh = (quiet: boolean) => {
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
      if (!empty) { setGroups(j.groups); setErr(null) }
      else if (quiet) return
      else if (wanted && !runTouched.current) setRunSel('__all__')
      else if (wanted) { setGroups([]); setErr(null) }
      else { setErr(j.reason || 'pas de résultats'); setErrKind('empty') }
    }).catch(() => { if (reqSeq.current === seq && !quiet) { setErr('backend injoignable — réessayez'); setErrKind('error') } })
    fetch('/api/integrity').then(r => r.json()).then(j => {
      // triple provenance: hash8 = sha256(dernier aqm_eval.csv)[:8]; repli run-id
      const id = j?.hash8 ?? String(j?.run_ids?.[0] ?? '').slice(0, 8)
      if (id) setHash8(String(id).slice(0, 8))
    }).catch(() => {})
  }
  useEffect(() => { refresh(false) }, [effectiveRun])
  // campagne en cours : les cellules gèlent au fil de l'eau — recharger
  // toutes les 10 s pour voir le classement BOUGER (barres animées par
  // valeur, flash des lignes remontées). Arrêt net avec la campagne.
  useEffect(() => {
    if (!liveSnapRunning) return
    const t = setInterval(() => refresh(true), 10000)
    return () => clearInterval(t)
  }, [liveSnapRunning, effectiveRun])
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
    // valeur brute d'une métrique (null = non mesurée, exclue du calcul)
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
    // X = le choix de l'opérateur, TOUJOURS (le repli silencieux est mort :
    // Gate A #4 — un axe sans dispersion s'affiche tel quel + une note le
    // dit sous le graphe, jamais de substitution cachée).
    const xm = meta(xk), ym = meta(yKey)
    const pts = vis.map((g, idx) => ({ g, idx, x: valOf(g, xk), y: valOf(g, yKey), name: `${g.profile}·${g.qdisc}·${g.cc}` })).filter(p => p.x != null && p.y != null) as { g: Group; idx: number; x: number; y: number; name: string }[]
    // écraseur 738 : Y passe en log quand la dynamique l'exige (zéros →
    // linéaire, jamais de point rogné en silence — voir shouldLogScale)
    const yLog = shouldLogScale(pts.map(p => p.y))
    const yStretch = (() => {
      if (pts.length < 2 || yLog) return null
      const ys = [...pts.map(p => p.y)].sort((a, b) => a - b)
      const med = ys[Math.floor(ys.length / 2)]
      const top = [...pts].sort((a, b) => b.y - a.y)[0]
      return med > 0 && top.y > 4 * med ? top : null
    })()
    // enveloppe réelle — points non dominés (même algo testé que la lib) ;
    // le rail droit et les labels reflètent exactement cet ensemble.
    const frontierIdx = paretoFrontier(pts.map(p => ({ x: p.x, y: p.y })), xm.dir, ym.dir)
    const frontierSet = new Set(frontierIdx)
    const rail = frontierIdx
      .map(i => pts[i])
      .sort((a, b) => (xm.dir === 'down' ? a.x - b.x : b.x - a.x))
    setFrontier(rail.map(p => ({ name: p.name, x: p.x, y: p.y })))
    // couleur par dimension — simple et lisible, une famille à la fois
    const fam = (g: Group): string => colorBy === 'profile' ? g.profile : colorBy === 'qdisc' ? g.qdisc : g.cc
    const famVals = Array.from(new Set(pts.map(p => fam(p.g)))).sort()
    const famColor = (f: string): string => {
      if (colorBy === 'qdisc') return PAL_QDISC[f] ?? '#8b9099'
      if (colorBy === 'cc') return PAL_CC[f] ?? '#8b9099'
      const i = famVals.indexOf(f)
      return PAL_PROFILE[i % PAL_PROFILE.length]
    }
    // labels = nœuds de l'enveloppe seuls (nom court profil·file : la
    // couleur encode déjà la famille) — fini le top-3 qui se recouvrait.
    // Un point sans label n'est pas muet : tooltip + rail le nomment.
    const c = echarts.init(scatterRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(scatterRef.current)
    const base = baseOption('', '')
    // séries par la grammaire — une famille = une série ; hors-enveloppe
    // atténués (opacité par point), enveloppe en polyligne verte par-dessus
    const series = famVals.map(fv => {
      const famPts = pts.filter(p => fam(p.g) === fv)
      const bestIdx = famPts.map((p, i) => (p.g.best ? i : -1)).filter(i => i >= 0)
      const s = scatterSeries(fv, famPts.map(p => [p.x, p.y] as [number, number]), famColor(fv), bestIdx)
      return {
        ...s,
        data: famPts.map(p => ({ value: [p.x, p.y] as [number, number], itemStyle: { opacity: frontierSet.has(p.idx) ? 1 : 0.22 } })),
        label: {
          show: true,
          formatter: (par: any) => {
            const p = famPts[par.dataIndex]
            return p && frontierSet.has(p.idx) ? `${p.g.profile}·${p.g.qdisc}` : ''
          },
          position: 'top' as const,
          distance: 10,
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
            const v = (par.value?.value ?? par.value) as [number, number]
            const who = famPts[par.dataIndex]?.name ?? String(fv)
            return `${who}<br/>${xm.label} : ${v[0].toFixed(1)} ${xm.unit}<br/>${ym.label} : ${v[1].toFixed(1)} ${ym.unit}`
          },
        },
      }
    })
    // l'enveloppe : polyligne réelle par les nœuds non dominés triés en X
    if (rail.length > 1) {
      series.push({
        name: 'enveloppe',
        type: 'line',
        data: rail.map(p => [p.x, p.y] as [number, number]),
        symbol: 'circle',
        symbolSize: 5,
        lineStyle: { color: '#1fa348', width: 2 },
        itemStyle: { color: '#1fa348', borderColor: '#0b0b0c', borderWidth: 1 },
        z: 10,
        tooltip: { trigger: 'item' as const, formatter: (par: any) => `enveloppe — ${(par.value as [number, number])[0].toFixed(1)} ${xm.unit} · ${(par.value as [number, number])[1].toFixed(1)} ${ym.unit}` },
      } as any)
    }
    const opt = {
      ...base,
      grid: { ...base.grid, left: 56, right: 28, top: 20, bottom: 52 },
      // mots sur les axes, pas seulement en légende (Gate A : labels OK)
      xAxis: { ...base.xAxis, type: 'value' as const, name: `${xm.label} (${xm.unit})`, nameLocation: 'middle' as const, nameGap: 30, nameTextStyle: { color: '#7e838c', fontSize: 10, fontFamily: 'JetBrains Mono' } },
      yAxis: { ...base.yAxis, type: (yLog ? 'log' : 'value') as 'log' | 'value', name: `${ym.label} (${ym.unit})${yLog ? ' — échelle log' : ''}`, nameTextStyle: { color: '#7e838c', fontSize: 10, fontFamily: 'JetBrains Mono' } },
      tooltip: { ...base.tooltip, trigger: 'item' as const },
      series,
    }
    c.setOption(opt as any)
    // légende-phrase : axes + ce que le nuage dit, en une ligne sous le graphe
    if (!pts.length) { setCaption('aucun point — élargissez les filtres'); setFrontier([]) }
    else {
      const score = (p: { y: number }) => (ym.dir === 'up' ? p.y : -p.y)
      const bp = [...pts].sort((a, b) => score(b) - score(a))[0]
      const stretch = yLog
        ? (() => { const t = [...pts].sort((a, b) => b.y - a.y)[0]; return ` · échelle log — ${t.name} à ${t.y.toFixed(1)} ${ym.unit} tire l'échelle` })()
        : yStretch ? ` · ${yStretch.name} à ${yStretch.y.toFixed(1)} ${ym.unit} tire l'échelle` : ''
      setCaption(`Y : ${ym.label} (${ym.unit}) · X : ${xm.label} (${xm.unit}) — ${pts.length} groupes · ${bp.g.profile}·${bp.g.qdisc}·${bp.g.cc} en tête (${bp.y.toFixed(1)} ${ym.unit})${stretch}`)
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

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind={errKind} hint={err} /></div>
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
  // colonnes du classement — UNE métrique par colonne, valeur + incertitude
  // où l'API en a une (small/RTT : CI/IQR gelés ; goodput/échéance/coût :
  // médiane + n, jamais d'intervalle inventé). get=null = non mesurée.
  type MetricCol = {
    key: TSpaceKey; rankKey: string; label: string; unit: string; dir: 'down' | 'up'
    tip: string; get: (g: Group) => number | null; ci: (g: Group) => [number, number] | null
  }
  const COLS: MetricCol[] = [
    {
      key: 'small', rankKey: 'small_p95_median', label: 'small p95', unit: 'ms', dir: 'down',
      tip: "Temps de chargement d'une sonde pendant un transfert lourd — plus bas = mieux",
      get: g => { const v = g.small_p95_valid_median ?? g.small_p95_median; return typeof v === 'number' && v > 0 ? v : null },
      ci: g => g.small_p95_valid_ci95 ?? g.small_p95_valid_iqr ?? g.small_p95_iqr ?? null,
    },
    {
      key: 'rtt', rankKey: 'rtt_p95_median', label: 'RTT p95', unit: 'ms', dir: 'down',
      tip: 'Aller-retour d\u2019un paquet sous charge — plus bas = mieux',
      get: g => (typeof g.rtt_p95_median === 'number' && g.rtt_p95_median > 0 ? g.rtt_p95_median : null),
      ci: g => g.rtt_p95_iqr ?? null,
    },
    {
      key: 'goodput', rankKey: 'goodput_median', label: 'goodput', unit: 'Mbit/s', dir: 'up',
      tip: 'Débit utile mesuré au récepteur — plus haut = mieux',
      get: g => (typeof g.goodput_median === 'number' && g.goodput_median >= 0 ? g.goodput_median : null),
      ci: () => null,
    },
    {
      key: 'deadline', rankKey: 'deadline_median', label: 'échéance', unit: '%', dir: 'up',
      tip: 'Part des sondes arrivées à temps — plus haut = mieux',
      get: g => { const v = g.deadline_median ?? g.deadline_ok_pct; return typeof v === 'number' && v >= 0 ? v : null },
      ci: () => null,
    },
    {
      key: 'cost', rankKey: 'cost', label: 'coût', unit: 'Ar/h', dir: 'down',
      tip: 'Argent perdu en retransmissions, au tarif unique — plus bas = mieux',
      get: g => costRef(g),
      ci: () => null,
    },
  ]
  const colOf = (rankKey: string): MetricCol => COLS.find(c => c.rankKey === rankKey) ?? COLS[0]
  const val = (g: Group): number => {
    const v = colOf(rankKey).get(g)
    return v == null ? Number.MAX_SAFE_INTEGER : v
  }
  const hero = (g: Group, c: MetricCol = colOf(rankKey)): { v: number; lo: number; hi: number; n: number } => {
    const n = validN(g)
    const v = c.get(g)
    const iv = c.ci(g)
    // v null (non mesuré) → sentinelle hors rang, comme val()
    const vv = v == null ? Number.MAX_SAFE_INTEGER : v
    return { v: vv, lo: iv?.[0] ?? vv, hi: iv?.[1] ?? vv, n }
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
  // échelle de VALEUR partagée : la barre mesure v sur [0–max visible],
  // dans les deux sens de critère (la plus longue = la plus grande valeur,
  // PAS la meilleure — le rang, lui, dit qui gagne). Les moustaches gardent
  // leur position absolue sur la même échelle.
  const finiteVals = filtered.map(val).filter(v => Number.isFinite(v) && v !== Number.MAX_SAFE_INTEGER)
  const vmax = finiteVals.length ? Math.max(...finiteVals, 0) : 1
  const vpos = (v: number): number => vmax > 0 ? Math.min(100, Math.max(0, (v / vmax) * 100)) : 0
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
  const rv = (g: Group): number | null => { const v = val(g); return v === Number.MAX_SAFE_INTEGER ? null : v }
  // écart 1er vs référence DANS LE SENS du critère (>0 = le 1er est meilleur)
  const duelImp = duelReady ? (() => { const t = rv(top as Group), b = rv(baselineRow as Group); return t != null && b != null ? improve(b, t, rankMeta.dir) : null })() : null
  // convention d'affichage (A2) : le % montré = amélioration improve()
  // (>0 = mieux, vert) + les mots « mieux »/« moins bon » — fini le
  // « −93 % vert » illisible. La flèche garde le sens de la valeur (↓/↑).
  const duelArrow = duelImp == null || duelImp === 0 ? '' : ((rankMeta.dir === 'down') === (duelImp > 0) ? '↓' : '↑')
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
    if (n <= 0 || !Number.isFinite(val(g))) return 'Pas de mesure valide — ligne non classée (détail : Quarantaine)'
    if (lossRegime) return 'Toutes les sondes ratent l\u2019échéance — la perte décide, pas la file'
    if (g.profile === 'P3' && (g.small_p95_median ?? 0) > 1500)
      return 'Toutes les sondes ratent l\u2019échéance — la perte décide, pas la file'
    if (i === 0 && filtered.length > 1) return `1er sur ${rankMeta.label} — ${n} réplications valides`
    const b = bestOf(g)
    if (b === g) return `meilleur ${g.profile} sur ce critère — ${n} réplications`
    // retard vs le meilleur du profil, en % dans le sens du critère
    const behind = improve(val(b), val(g), rankMeta.dir)
    return behind != null && behind < 0
      ? `+${Math.abs(behind)} % vs ${b.qdisc}, meilleur ${g.profile} — ${n} réplications`
      : `au coude-à-coude avec ${b.qdisc} — ${n} réplications`
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
  // méta source : le run qui nourrit les chiffres, en clair
  const validTotal = safeGroups.reduce((a, g) => a + validN(g), 0)
  const metaRun = effectiveRun ? `run gelé ${effectiveRun}` : 'tous runs (gelés)'
  // coin du TradeSpace : où vit l'optimum (down = plus bas = mieux) —
  // suit l'axe X CHOISI (plus de substitution silencieuse)
  const yd = (TSPACE.find(t => t.key === yKey)?.dir ?? 'down') === 'down'
  const xd = (TSPACE.find(t => t.key === xk)?.dir ?? 'down') === 'down'
  const cornerArrow = yd ? (xd ? '↙' : '↘') : (xd ? '↖' : '↗')
  // dispersion relative d'un axe (0 = tous les points au même endroit) —
  // sert la note "X varie peu", jamais un repli caché
  const xSpread = (k: TSpaceKey): number => {
    const xs = (k === 'indice'
      ? filtered.map(g => lienScore(lienAll, groupLien(g)))
      : filtered.map(g => COLS.find(c => c.key === k)?.get(g) ?? null)
    ).filter((v): v is number => typeof v === 'number' && Number.isFinite(v))
    if (xs.length < 2) return 0
    const mx = Math.max(...xs.map(Math.abs), 1e-9)
    return (Math.max(...xs) - Math.min(...xs)) / mx
  }

  // texte d'une cellule métrique (coût compacté en kAr, % sans décimale)
  const fmtCell = (c: MetricCol, v: number): string => {
    if (c.key === 'cost') return v >= 1000 ? `${(v / 1000).toFixed(1)} k` : v.toFixed(0)
    if (c.key === 'deadline') return v.toFixed(0)
    return v.toFixed(1)
  }
  // teinte d'une cellule secondaire : écart signé vs la référence du duel
  // (même baseline partout — une seule histoire). alpha ∝ |écart|.
  const heatOf = (c: MetricCol, g: Group): { bg: string; fg: string; title?: string } => {
    const b = baselineRow ? c.get(baselineRow) : null
    const v = c.get(g)
    const imp = b != null && v != null ? improve(b, v, c.dir) : null
    if (imp == null || imp === 0) return { bg: 'transparent', fg: '#f2f2f4' }
    const alpha = Math.min(0.22, Math.abs(imp) / 150)
    return {
      bg: imp > 0 ? `rgba(31,163,72,${alpha.toFixed(2)})` : `rgba(226,39,24,${alpha.toFixed(2)})`,
      fg: imp > 0 ? '#1fa348' : 'var(--t-danger-text, #e84a3a)',
      title: `${imp > 0 ? 'mieux' : 'moins bon'} que ${baselineIsPfifo ? 'pfifo' : (baselineRow as Group).qdisc} de ${Math.abs(imp)} % sur ${c.label}`,
    }
  }

  return (
    <div className="panel-stack" style={{ position: 'relative' }}>
      <h1 className="view-title">Résultats — classement</h1>

      {/* bandeau benchmark — source, provenance, export (façon DeepSWE) */}
      <div className="card bench-bar" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 14px', flexWrap: 'wrap' }}>
        <span className="mono" style={{ fontSize: 12, fontWeight: 500 }}>Meteolink Leaderboard</span>
        <span className="mono muted" style={{ fontSize: 11 }}>{safeGroups.length} groupes · {metaRun} · {validTotal} mesures valides</span>
        {liveSnapRunning && <span className="mono" style={{ fontSize: 10, color: '#1fa348', border: '1px solid currentColor', padding: '2px 8px' }}>● CAMPAGNE EN COURS</span>}
        <select data-testid="run-select" value={runSel} onChange={e => { runTouched.current = true; setRunSel(e.target.value) }} title="source des chiffres : en direct = dernier run gelé — identifiant complet dans la liste" style={{ marginLeft: 'auto', background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'JetBrains Mono', fontSize: 11, maxWidth: '100%' }}>
          <option value="">{newest ? `⚡ En direct — ${newest}` : '⚡ En direct'}</option>
          <option value="__all__">tous runs (gelés)</option>
          {orderedRuns.map(id => <option key={id} value={id} title={id}>{id === newest ? `⚡ ${id}` : id}</option>)}
        </select>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{ border: '1px solid var(--hairline)', padding: '7px 16px' }}>MD</a>
        <button className="btn" title="Lire le classement : métriques et règles de lecture" aria-label="Lire le classement" onClick={() => setInfoOpen(true)} style={{ padding: '7px 12px', fontSize: 13 }}>ⓘ</button>
      </div>
      {infoOpen && <ResultatsInfoModal onClose={() => setInfoOpen(false)} />}

      {/* constat de campagne — UNE phrase, sans chiffre : le chiffre vit
          dans le duel ci-dessous (dédupliqué : constat/duel/verdict). */}
      <div className="card" style={{ padding: '10px 14px', border: '1px solid ' + (top ? CRAFT.ok : 'var(--hairline)') }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-muted)' }}>constat de campagne</span>
          <button className="btn" data-testid="constat-button" onClick={() => setInterpProfile(filtered[0]?.profile ?? 'P2')} style={{ marginLeft: 'auto', padding: '4px 12px', fontSize: 10 }}>Interprétation complète →</button>
        </div>
        <p className="mono" style={{ margin: '8px 0 0', fontSize: 13, lineHeight: 1.6, color: '#c3c9d1', maxWidth: '68ch' }}>
          {duelReady && lossRegime
            ? 'Sur P3, toutes les files se valent : aucune sonde n\u2019arrive à temps (0 % partout)'
            : duelReady
              ? <>{top.profile} : {top.qdisc}/{top.cc} en tête sur {rankMeta.label} (n={validN(top as Group)} réplications valides) — <button className="btn" onClick={() => document.getElementById('duel')?.scrollIntoView({ behavior: 'smooth', block: 'center' })} style={{ padding: '2px 10px', fontSize: 11 }}>Voir le duel ↓</button></>
              : 'aucun groupe comparable — lancez une campagne ou élargissez les filtres'}
        </p>
      </div>
      {interpProfile && <InterpretationView profile={interpProfile} onClose={() => setInterpProfile(null)} />}

      {/* duel critère — référence vs 1er : LE chiffre vit ici et
          nulle part ailleurs (constat renvoie ici, verdicts jamais).
          Grille .rank-verdict : 3 colonnes desktop, empilée ≤640px. */}
      {duelReady && (() => {
        const bv = rv(baselineRow as Group), tv = rv(top as Group)
        const good = duelImp != null && duelImp > 0
        const col = (label: string, v: number | null, n: number, color: string) => (
          <div style={{ minWidth: 0, textAlign: 'center' }}>
            <div className="mono" style={{ fontSize: 12, letterSpacing: '0.06em', textTransform: 'uppercase', color: '#a9aeb6', marginBottom: 6 }}>{label}</div>
            <div className="mono" style={{ fontSize: 26, fontWeight: 600, fontVariantNumeric: 'tabular-nums', color, lineHeight: 1 }}>
              {v != null ? v.toFixed(1) : '—'}
            </div>
            <div className="mono" style={{ fontSize: 12, color: '#a9aeb6', marginTop: 4 }}>{rankMeta.unit} · {n} mesures valides</div>
          </div>
        )
        return (
          <div className="card rank-verdict" id="duel" data-testid="rank-verdict" style={{ padding: '18px 20px' }}>
            {col(`${baselineIsPfifo ? 'pfifo' : (baselineRow as Group).qdisc} — avant`, bv, validN(baselineRow as Group), '#a9aeb6')}
            <div style={{ textAlign: 'center' }}>
              <div className="mono" data-testid="rank-diff" style={{ fontSize: 34, fontWeight: 600, fontVariantNumeric: 'tabular-nums', color: good ? '#1fa348' : '#d6d8dd', lineHeight: 1 }}>
                {lossRegime ? 'égalité' : duelImp != null ? (duelImp > 0 ? `+${duelImp} %` : duelImp < 0 ? `${duelImp} %` : '0 %') : '—'}
              </div>
              {!lossRegime && duelImp != null && duelImp !== 0 && (
                <div className="mono" style={{ fontSize: 13, color: duelImp > 0 ? '#1fa348' : 'var(--t-danger-text, #e84a3a)', marginTop: 4, lineHeight: 1.4 }}>{duelArrow} {duelImp > 0 ? 'mieux' : 'moins bon'}</div>
              )}
              <div className="mono" style={{ fontSize: 11, color: '#a9aeb6', marginTop: 6, maxWidth: 130 }}>{rankMeta.label}</div>
            </div>
            {col(`1er — ${top.qdisc}/${top.cc}`, tv, validN(top as Group), good ? '#1fa348' : '#d6d8dd')}
          </div>
        )
      })()}

      {/* critère + filtres — rangée scrollable (jamais 3 lignes à 390px) */}
      <div className="form-row chip-row" data-testid="chip-row" style={{ gap: 6, alignItems: 'center', marginBottom: 8 }}>
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
      {filtered.length > 0 && <div className="card lb-card" style={{ padding: '4px 0' }}>
        {/* classement grammaire arène : en-tête sticky triable + ⓘ par
            colonne, identité + sous-ligne décodée, valeur + ±CI par cellule,
            teinte par signe sur les secondaires, moustaches sur la primaire.
            ≤640px : secondaires repliées dans le détail (CSS .lb-sec). */}
        <div className="lb-scroll">
        <table className="data-table lb-table">
          <thead><tr>
            <th className="lb-rank" title="Rang sur ce critère — 1er en haut">#</th>
            <th className="lb-id" title="Profil de lien · file d'attente · contrôle de congestion">Configuration <span aria-hidden="true">ⓘ</span></th>
            {COLS.map(c => (
              <th key={c.key} className={c.rankKey === rankKey ? 'lb-prim' : 'lb-sec'}>
                <button className="lb-sort" onClick={() => setRankKey(c.rankKey)} title={`${c.tip} — cliquer pour classer par ${c.label}`} aria-label={`Classer par ${c.label}`}>
                  {c.label}{c.rankKey === rankKey ? (c.dir === 'up' ? ' ↑' : ' ↓') : ''} <span aria-hidden="true" className="lb-info">ⓘ</span>
                </button>
              </th>
            ))}
          </tr></thead>
          <tbody>
        {ranked.map((g, i) => {
          const barColor = QCOLOR[g.qdisc] ?? '#6b7078'
          const key = `${g.profile}/${g.qdisc}/${g.cc}/${g.direction ?? 'up'}`
          const open = expanded === key
          const cellDelta = deltas[`${g.profile}|${g.qdisc}|${g.cc}${g.direction && g.direction !== 'up' ? `|${g.direction}` : ''}`]?.small_p95_pct
          const wasted: number | null = g.wasted_median ?? g.wasted_bytes ?? null
          const cost: number | null = costRef(g)
          const pin = { profile: g.profile, qdisc: g.qdisc, cc: g.cc }
          const isA = pinA?.profile === g.profile && pinA?.qdisc === g.qdisc && pinA?.cc === g.cc
          const isB = pinB?.profile === g.profile && pinB?.qdisc === g.qdisc && pinB?.cc === g.cc
          const metricCell = (c: MetricCol) => {
            const h = hero(g, c)
            const ok = h.v !== Number.MAX_SAFE_INTEGER && (c.key === 'cost' || c.dir === 'up' ? h.v >= 0 : h.v > 0)
            const primary = c.rankKey === rankKey
            const half = h.hi > h.lo ? ((h.hi - h.lo) / 2) : null
            const heat = !primary ? heatOf(c, g) : null
            return (
              <td key={c.key} className={primary ? 'lb-prim' : 'lb-sec'} title={heat?.title} style={heat ? { background: heat.bg, ...(c.key === 'cost' ? { whiteSpace: 'nowrap' as const } : null) } : (c.key === 'cost' ? { whiteSpace: 'nowrap' as const } : undefined)}>
                <div className="mono" style={{ fontSize: primary ? 13 : 12, fontWeight: primary ? 600 : 500, color: !ok ? 'var(--text-faint, #7e838c)' : heat ? heat.fg : '#f2f2f4', fontVariantNumeric: 'tabular-nums', textAlign: 'right', overflowWrap: 'anywhere' }}>
                  {ok ? <>{fmtCell(c, h.v)} <span style={{ fontWeight: 400, fontSize: 10, color: '#8b9099' }}>{c.unit}</span></> : '—'}
                </div>
                <div className="mono" style={{ fontSize: 10, color: '#8b9099', marginTop: 2, textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                  {ok ? <>{half != null ? `±${half.toFixed(half >= 10 ? 0 : 1)} · ` : ''}n={h.n}</> : 'non mesuré'}
                </div>
                {primary && (
                  <div style={{ height: 12, position: 'relative', marginTop: 4, minWidth: 80 }}>
                    <div style={{ position: 'absolute', inset: '2px 0', background: 'rgba(255,255,255,0.04)', borderRadius: 2, overflow: 'hidden' }}>
                      {ok && <div className="leader-bar" data-leader={key} style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: `${vpos(h.v)}%`, background: barColor, transformOrigin: 'left center', borderRadius: 2 }} />}
                    </div>
                    {ok && half != null && (
                      <div title={`intervalle [${h.lo.toFixed(1)}–${h.hi.toFixed(1)}] ${c.unit}`} style={{ position: 'absolute', top: '50%', marginTop: -1, height: 2, left: `${vpos(h.lo)}%`, width: `${Math.max(1.5, vpos(h.hi) - vpos(h.lo))}%`, borderLeft: '2px solid #f2f2f4', borderRight: '2px solid #f2f2f4', background: '#f2f2f4' }} />
                    )}
                  </div>
                )}
                {c.key === 'small' && cellDelta != null && (
                  <div className="mono" title={`vs run précédent, même cellule — ${cellDelta <= 0 ? 'mieux' : 'moins bon'} (la valeur ${cellDelta <= 0 ? 'baisse' : 'monte'})`} style={{ fontSize: 10, color: cellDelta <= 0 ? '#1fa348' : 'var(--t-danger-text, #e84a3a)', fontVariantNumeric: 'tabular-nums', textAlign: 'right', marginTop: 2 }}>
                    {cellDelta <= 0 ? '↘' : '↗'}{Math.abs(cellDelta)} %
                  </div>
                )}
              </td>
            )
          }
          return (
            <Fragment key={key}>
            <tr className="lb-row" tabIndex={0} aria-expanded={open} title={`${g.profile} ${g.qdisc} ${g.cc} — cliquer pour le détail`} style={{ background: 'transparent', cursor: 'pointer', opacity: expanded && !open ? 0.35 : 1, filter: expanded && !open ? 'saturate(0.5)' : 'none', transition: 'opacity 250ms ease, filter 250ms ease' }} onClick={() => setExpanded(open ? null : key)} onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setExpanded(open ? null : key) } }}>
              <td className="lb-rank mono">{i + 1}</td>
              <td className="lb-id">
                <div className="mono" style={{ fontSize: 13, fontWeight: 500, color: '#f2f2f4', overflowWrap: 'anywhere' }}>
                  {g.profile} · {g.qdisc} / {g.cc} <span title={g.cc === 'bbr' ? 'contrôle de congestion BBR (Google) — sonde le débit, peu agressif' : 'contrôle de congestion CUBIC — remplit les files, agressif'} style={{ color: barColor }}>{CCMARK[g.cc] ?? ''}</span>
                  {g.direction && g.direction !== 'up' ? <span title="sens download mesuré" style={{ color: '#5ad3e3' }}> ↓</span> : null}
                </div>
                <div className="mono" style={{ fontSize: 11, color: '#a9aeb6', marginTop: 2, overflowWrap: 'anywhere' }}>{decodeSub(g)}</div>
                {!singleVerdict && <div className="mono" style={{ fontSize: 11, color: '#c3c9d1', marginTop: 2, overflowWrap: 'anywhere' }}>{rowVerdicts[i]}</div>}
                <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
                  <button className="btn" title="épingler comme A pour le duel" aria-label={`épingler ${g.profile} ${g.qdisc} ${g.cc} comme A`} onClick={e => { e.stopPropagation(); setPinA(pin) }} style={{ padding: '2px 8px', fontSize: 10, background: isA ? 'rgba(90,211,227,0.15)' : 'transparent', color: isA ? CRAFT.live : 'var(--text-muted)' }}>A</button>
                  <button className="btn" title="épingler comme B pour le duel" aria-label={`épingler ${g.profile} ${g.qdisc} ${g.cc} comme B`} onClick={e => { e.stopPropagation(); setPinB(pin) }} style={{ padding: '2px 8px', fontSize: 10, background: isB ? 'rgba(31,163,72,0.15)' : 'transparent', color: isB ? CRAFT.ok : 'var(--text-muted)' }}>B</button>
                </div>
              </td>
              {COLS.map(c => metricCell(c))}
            </tr>
            {open && (
            <tr className="lb-detail"><td colSpan={2 + COLS.length}>
                {/* repli 390px : les secondaires masquées se relisent ici */}
                <div className="mono" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: 8, marginBottom: 10 }}>
                  {COLS.map(c => {
                    const dh = hero(g, c)
                    const dv = c.get(g)
                    const dhalf = dh.hi > dh.lo ? ((dh.hi - dh.lo) / 2) : null
                    return (
                      <div key={c.key} style={{ border: '1px solid var(--hairline-faint)', padding: '8px 10px' }}>
                        <div style={{ fontSize: 10, color: '#8b9099' }}>{c.label}{c.rankKey === rankKey ? ' — critère' : ''}</div>
                        <div style={{ fontSize: 15, fontWeight: 600, color: dv != null ? '#f2f2f4' : 'var(--text-faint, #7e838c)', fontVariantNumeric: 'tabular-nums', marginTop: 2 }}>
                          {dv != null ? <>{fmtCell(c, dv)} <span style={{ fontSize: 10, fontWeight: 400 }}>{c.unit}</span></> : '—'}
                        </div>
                        <div style={{ fontSize: 10, color: '#8b9099', marginTop: 2 }}>{dhalf != null ? `±${dhalf.toFixed(dhalf >= 10 ? 0 : 1)} ${c.unit} · ` : ''}n={dh.n}</div>
                      </div>
                    )
                  })}
                </div>
                <div style={{ padding: '12px 14px', border: '1px solid var(--hairline)', background: 'rgba(255,255,255,0.015)' }}>
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
                        IC95 : {g.small_p95_valid_ci95[0].toFixed(0)}–{g.small_p95_valid_ci95[1].toFixed(0)} ms
                      </div>
                    )}
                    <div className="mono" style={{ fontSize: 12, color: g.quarantined > 0 ? '#f4b400' : '#a9aeb6' }}>
                      {g.quarantined > 0
                        ? `${g.quarantined} mesure(s) sur ${g.count} écartée(s) — incohérentes, exclues du calcul (détail : vue Provenance)`
                        : `les ${g.count} mesures sont cohérentes — aucune écartée`}
                    </div>
                    <div className="mono" style={{ fontSize: 12, color: '#a9aeb6', fontVariantNumeric: 'tabular-nums' }}>
                      Coût : {wasted == null || wasted <= 0 ? 'rien gaspillé' : (wasted >= 1048576 ? (wasted / 1048576).toFixed(1) + ' MiB gaspillés' : wasted >= 1024 ? (wasted / 1024).toFixed(0) + ' Kio gaspillés' : wasted + ' o gaspillés')} · {cost == null || cost <= 0 ? '0 Ar' : (cost >= 1000 ? (cost / 1000).toFixed(1) + ' kAr' : cost.toFixed(0) + ' Ar')}
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
                          Note sur 100 = moyenne des 5 notes, une par critère.
                        </div>
                      </div>
                    )
                  })()}
                  <div style={{ marginTop: 10, display: 'flex', gap: 6 }}>
                    <button className="btn" onClick={e => { e.stopPropagation(); setInterpProfile(g.profile) }} style={{ padding: '4px 10px', fontSize: 10 }}>interpréter {g.profile} →</button>
                  </div>
                </div>
              </td></tr>
              )}
            </Fragment>
          )
        })}
          </tbody>
        </table>
        </div>
        {/* échelle de valeur partagée — les barres mesurent v sur [0–max],
            pas la bonté. Sans ligne valide, pas d'échelle à lire. */}
        {finiteVals.length > 0 && (
          <div className="mono" style={{ display: 'flex', justifyContent: 'space-between', gap: 8, padding: '8px 14px 12px', borderTop: '1px solid var(--hairline-faint)', fontSize: 11, color: '#a9aeb6' }}>
            <span>0 {rankMeta.unit}</span><span>les barres mesurent la valeur, pas le rang</span><span>{vmax.toFixed(1)} {rankMeta.unit} →</span>
          </div>
        )}
      </div>}
      {pinA && pinB && (
        <CompareView a={pinA} b={pinB} onClose={() => { setPinA(null); setPinB(null) }} />
      )}
      <div className="card" style={{ padding: 12, position: 'relative' }}>
        <div style={{ marginBottom: 4 }}>
          <span className="mono" style={{ fontSize: 13, fontWeight: 600, color: '#f2f2f4' }}>TradeSpace — compromis mesurés</span>
          {(() => {
            const xmR = TSPACE.find(t => t.key === xk) ?? TSPACE[0]
            const ymR = TSPACE.find(t => t.key === yKey) ?? TSPACE[0]
            return <div className="mono" style={{ fontSize: 11, color: '#a9aeb6', marginTop: 2 }}>{ymR.label} ({ymR.unit}) en fonction de {xmR.label} ({xmR.unit}) — médianes gelées, {filtered.length} groupes · {metaRun}</div>
          })()}
        </div>
        <div className="form-row chip-row" style={{ gap: 6, alignItems: 'center', marginBottom: 6 }}>
          <span className="mono" title="Y = le critère du classement — un seul modèle mental" style={{ fontSize: 10, color: '#5ad3e3' }}>Y {(() => TSPACE.find(t => t.key === yKey)?.label)()}</span>
          <span className="mono" style={{ fontSize: 10, color: '#8b9099' }}>X</span>
          {TSPACE.filter(t => t.key !== yKey).map(t => chip(t.label, xk === t.key, () => setXKey(t.key)))}
          <span className="mono" style={{ fontSize: 10, color: '#8b9099' }}>couleur</span>
          {(['profile', 'qdisc', 'cc'] as const).map(v => chip({ profile: 'profil', qdisc: 'file', cc: 'CC' }[v], colorBy === v, () => setColorBy(v)))}
        </div>
        <div className="ts-wrap">
          <div style={{ flex: 1, minWidth: 0, position: 'relative' }}>
            <span className="mono" style={{ position: 'absolute', top: 2, right: 4, fontSize: 10, color: 'var(--text-faint)', pointerEvents: 'none', zIndex: 1 }}>
              optimal {cornerArrow}
            </span>
            <div ref={scatterRef} style={{ height: 320 }} />
          </div>
          <aside className="ts-rail" aria-label="Enveloppe — points non dominés">
            <div className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#1fa348', marginBottom: 6 }}>Enveloppe — {frontier.length} point{frontier.length > 1 ? 's' : ''}</div>
            {frontier.length === 0 && <div className="mono" style={{ fontSize: 11, color: '#8b9099' }}>aucun point — élargissez les filtres</div>}
            {(() => {
              const xmR = TSPACE.find(t => t.key === xk) ?? TSPACE[0]
              const ymR = TSPACE.find(t => t.key === yKey) ?? TSPACE[0]
              return frontier.map(p => (
                <div key={p.name} className="mono" style={{ fontSize: 11, padding: '5px 0 5px 10px', borderLeft: '2px solid #1fa348', marginBottom: 4, color: '#d6d8dd', overflowWrap: 'anywhere' }}>
                  <div style={{ color: '#f2f2f4' }}>{p.name}</div>
                  <div style={{ color: '#1fa348', fontVariantNumeric: 'tabular-nums' }}>{p.y.toFixed(1)} {ymR.unit} · {p.x.toFixed(1)} {xmR.unit}</div>
                </div>
              ))
            })()}
          </aside>
        </div>
        {(() => {
          const sp = xSpread(xk)
          if (sp >= 0.18 || filtered.length < 2) return null
          const xmR = TSPACE.find(t => t.key === xk) ?? TSPACE[0]
          const alt = TSPACE.filter(t => t.key !== yKey && t.key !== xk).sort((a, b) => xSpread(b.key) - xSpread(a.key))[0]
          return <div className="mono" style={{ fontSize: 11, color: '#f4b400', marginTop: 6 }}>« {xmR.label} » varie peu sur ces groupes — les points s'alignent{alt ? `, essayez ${alt.label}` : ''}.</div>
        })()}
        <div className="mono" style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 6 }}>{caption}</div>
      </div>
      <Provenance source="mesures gelées des campagnes" refresh="au gel" state="live" extra={`${safeGroups.length} groupes · hash ${hash8}`} />
    </div>
  )
}
