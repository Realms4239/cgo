import { memo, useCallback, useEffect, useRef, useState } from 'react'
import type { EChartsOption } from 'echarts'
import Explain from '../components/Explain'
import EventBadge from '../components/EventBadge'
import { echarts } from '../lib/echarts'
import { baseOption, lineSeries, barSeries, chargeMarkArea, rawLineSeries, CRAFT } from '../lib/chartGrammar'
import { live, clearLive } from '../lib/live'
import { computeQDI } from '../lib/qdi'
import { computeJFI } from '../lib/jfi'
import { useUIStore } from '../store/ui'
import { lttb } from '../lib/lttb'
import { useRafLoop } from '../lib/hooks'
import { animateBannerPulse, animateLiveEnter } from '../lib/anime'
import { MetricCard } from '../components/ui/MetricCard'
import { LiveGuide } from '../components/LiveGuide'
import { EmptyState } from '../components/ui/EmptyState'
import { Card } from '../components/ui/Card'
import { EmptyChart } from '../components/ui/EmptyChart'
import { PeekPopover } from '../components/PeekPopover'
import { Beam } from '../components/Beam'
import { DonutJFI } from '../components/DonutJFI'
import { Timeline } from '../components/Timeline'
import { CardHead } from '../components/ui/CardHead'
import { Pill } from '../components/ui/Pill'
import { loadSettings, latencyLevel, dropsLevel, deadlineLevel, goodputLevel, jfiLevel, LEVEL_COLOR, type Settings } from '../lib/settings'
import { formatBytesFR, formatCostAR, verdictWord, truncateWords } from '../lib/format'
import { paletteOnce } from '../lib/palette'

type Craft = 'line' | 'bar' | 'area'
type Tri = { metric: boolean; chart: Craft; source: 'live' | 'frozen' | 'both' }
const DEFAULT_TRI: Tri = { metric: true, chart: 'line', source: 'both' }

// Adaptateur — une grammaire, trois rendus (pilule: ligne|barre|aire)
function craftSeries(craft: Craft, name: string, data: [number, number][], color: string) {
  if (craft === 'bar') return barSeries(name, data as any, color)
  return lineSeries(name, data, color, craft === 'area')
}

// useChart — ref de callback: l'instance ECharts suit le div que React
// mounts (re-renders at 2 Hz may replace nodes; useEffect([]) orphaned them).
// useChart — le composant surface possède le cycle de vie de l'instance (créée
// après layout, détruite à SON démontage). Les re-rendus parents à 2 Hz ne l'y touchent pas;
// l'ordre des refs React 19 ne peut plus orpheliner ni détruire le graphe.
function useChart(_title: string, _unit: string) {
  const chart = useRef<echarts.ECharts | null>(null)
  const onReady = useCallback((c: echarts.ECharts | null) => { chart.current = c }, [])
  const setData = (series: { name?: string; data?: unknown[] }[], extra?: Record<string, unknown>) => {
    if (!chart.current || chart.current.getWidth() < 10) return
    const empty = !series.some(s => ((s.data as unknown[]) ?? []).length > 1)
    const base = baseOption(_title, _unit, { idle: empty })
    // légende basse Grafana — noms des séries sous le graphe, grille remontée
    const names = series.map(s => (s as { name?: string }).name).filter((n): n is string => !!n)
    const leg = empty || names.length === 0 ? {} : {
      legend: { bottom: 0, left: 'center', textStyle: { color: paletteOnce().textMuted, fontFamily: 'var(--font-mono)', fontSize: 'var(--fs-micro)' }, itemWidth: 14, itemHeight: 2, data: names },
      grid: { left: 64, right: 32, top: 48, bottom: 58, containLabel: true },
    }
    const opt = { animation: false, ...base, ...(empty ? { dataZoom: [] } : {}), ...leg, ...extra, series } as unknown as EChartsOption
    chart.current.setOption(opt)
  }
  return { onReady, setData, chart }
}

// ChartSurface — propriétaire mémorisé : init après layout, dispose au démontage.
const ChartSurface = memo(function ChartSurface({ title, unit, domId, height, empty, hint, onReady }: {
  title: string
  unit: string
  domId?: string
  height: number | string
  empty: boolean
  hint: string
  onReady: (c: echarts.ECharts | null) => void
}) {
  const boxRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!boxRef.current) return
    const c = echarts.init(boxRef.current, undefined, { renderer: 'canvas', useDirtyRect: true, devicePixelRatio: Math.min(window.devicePixelRatio, 2) } as any)
    // amorcer le repère AVANT toute donnée — la première donnée
    // setOption on a never-painted chart crashed LineView ('coord') and
    // empoisonnait le flush partagé de zrender, gelant les autres graphes.
    // La base amorcée sert aussi de grille idle visible (vide honnête).
    try { c.setOption(baseOption(title, unit, { idle: true })) } catch (e) { console.error('[chart] prime failed', e) }
    const ro = new ResizeObserver(() => { try { c.resize() } catch { } })
    ro.observe(boxRef.current)
    onReady(c)
    return () => { ro.disconnect(); try { c.dispose() } catch { }; onReady(null) }
  }, [])
  return (
    <div className="chart-surface" style={{ height }}>
      <div ref={boxRef} id={domId} className="chart-box" />
      {empty && <EmptyChart hint={hint} />}
    </div>
  )
})

type WallGroup = { qdisc: string; profile: string; small_p95_median: number; best?: boolean; hardware_recommendation?: string }

export default function LiveView() {
  // settings — reactive: Réglages diffuse `meteolink-settings`, les seuils de
  // sévérité et la capacité du bord suivent sans rechargement (stale ref bug).
  const [settings, setSettings] = useState<Settings>(loadSettings())
  useEffect(() => {
    const on = () => setSettings({ ...loadSettings() })
    window.addEventListener('meteolink-settings', on)
    return () => window.removeEventListener('meteolink-settings', on)
  }, [])
  const rtt = useChart('RTT', 'ms')
  const small = useChart('Petits objets p95', 'ms')
  const goodput = useChart('Débit utile applicatif', 'Mbit/s')
  const iface = useChart('Interface', 'Mbit/s')
  const liveSnap = useUIStore((s: any) => s.live)
  const replayRunning = useUIStore((s: any) => s.replayRunning)
  const replayRunId = useUIStore((s: any) => s.replayRunId)
  const bannerRef = useRef<HTMLDivElement>(null)
  const [peek, setPeek] = useState<{ rect: DOMRect; value: number } | null>(null)
  const [wallGroups, setWallGroups] = useState<WallGroup[] | null>(null)
  const [wallHash, setWallHash] = useState<string>('')
  const [hovered, setHovered] = useState(false)
  const [tri, setTri] = useState<Tri>(DEFAULT_TRI)
  // Contrôle du bord — verrou de baseline + levier de façonnage
  const [locked, setLocked] = useState<{ ms: number; at: string } | null>(() => {
    try { const r = localStorage.getItem('wall-baseline'); return r ? JSON.parse(r) : null } catch { return null }
  })
  const lockedRingRef = useRef<[number, number][]>([])
  const [shape, setShape] = useState<{ applied: boolean; qdisc?: string; capacity_mbps?: number } | null>(null)
  const [shapeCap, setShapeCap] = useState(settings.shapeCap)
  // la capacité du bord suit les Réglages (le levier repart du défaut)
  useEffect(() => { setShapeCap(settings.shapeCap) }, [settings.shapeCap])
  const [shapeMsg, setShapeMsg] = useState('')
  const [journal, setJournal] = useState<{ ts: string; kind: string; msg: string }[]>([])
  const [journalTotal, setJournalTotal] = useState(0)
  const [journalOpen, setJournalOpen] = useState<boolean>(() => {
    try { return localStorage.getItem('live-journal-open') === '1' } catch { return false }
  })
  const [linkOpen, setLinkOpen] = useState(false)
  const [watching, setWatching] = useState(false)
  const [burstCc, setBurstCc] = useState('bbr')
  const [bursting, setBursting] = useState(false)

  // Q4 tri-toggle — PanelChooser dispatches {metric, chart: craft, source}
  useEffect(() => {
    const onTri = (e: Event) => {
      const d = (e as CustomEvent).detail
      if (!d) return
      setTri({ metric: !!d.metric, chart: (d.chart as Craft) || 'line', source: (d.source as Tri['source']) || 'both' })
    }
    window.addEventListener('panel-chooser-tri', onTri)
    return () => window.removeEventListener('panel-chooser-tri', onTri)
  }, [])

  // idle honnête: un nouveau run part d'anneaux vides — pas de séries périmées
  useEffect(() => { if (liveSnap?.running) clearLive() }, [liveSnap?.running])

  // historique gelé pour la baseline A/B vs CAKE hors course
  useEffect(() => {
    if (liveSnap?.running || replayRunning) return
    let cancelled = false
    fetch('/api/results').then(r => r.json()).then(j => { if (!cancelled && j.available) setWallGroups(j.groups) }).catch(() => {})
    fetch('/api/integrity').then(r => r.json()).then(j => {
      if (cancelled) return
      // triple provenance: hash8 = sha256(dernier aqm_eval.csv)[:8]; repli run-id
      const id = j?.hash8 ?? String(j?.run_ids?.[0] ?? '').slice(0, 8)
      if (id) setWallHash(String(id).slice(0, 8))
    }).catch(() => {})
    fetch('/api/shape').then(r => r.json()).then(j => { if (!cancelled) setShape(j) }).catch(() => {})
    fetch('/api/events').then(r => r.json()).then(j => { if (!cancelled && j?.events) { setJournalTotal(j.events.length); setJournal(j.events.slice(-50).reverse()) } }).catch(() => {})
    return () => { cancelled = true }
  }, [liveSnap?.running, replayRunning])

  const hash8 = wallHash || (() => {
    const src = `${liveSnap?.profile ?? ''}${liveSnap?.qdisc ?? ''}${liveSnap?.cc ?? ''}${wallGroups?.[0]?.profile ?? ''}`
    return src ? src.slice(0, 8).padEnd(8, '·').slice(0, 8) : '────────'
  })()

  const lastRef = useRef(0)
  const seqRef = useRef(-1)
  const ifaceSeqRef = useRef(-1)
  const ifaceLastRef = useRef(0)
  // la boucle ne tourne que s'il y a de la donnée à peindre — idle = zéro frame
  const rafActive = live.rtt95.length > 0 || live.small.length > 0 || live.goodput.length > 0 || lockedRingRef.current.length > 0 || live.iface.rx.length > 0
  useRafLoop((ts) => {
    // interface — porte 10 Hz propre, série brute, jamais passé par LTTB
    if (live.iface.seq !== ifaceSeqRef.current && ts - ifaceLastRef.current >= 100) {
      const charging = live.phase === 'charge'
      const cs = charging ? live.phaseSince['charge'] ?? live.iface.rx[0]?.[0] ?? Date.now() - 1000 : 0
      const ce = charging ? live.iface.rx.at(-1)?.[0] ?? Date.now() : 0
      const ma = chargeMarkArea(cs, ce, charging)
      iface.setData([
        { ...rawLineSeries('RX', live.iface.rx, CRAFT.live), markArea: ma } as any,
        { ...rawLineSeries('TX', live.iface.tx, CRAFT.ok) } as any,
      ])
      ifaceSeqRef.current = live.iface.seq
      ifaceLastRef.current = ts
    }
    // dirty-check — les anneaux n'ont pas bougé : rien à re-rendre, frame gratuite
    if (live.seq === seqRef.current) return
    if (ts - lastRef.current < 250) return
    lastRef.current = ts
    seqRef.current = live.seq
    const d = (r: [number, number][]) => r.length > 400 ? lttb(r, 400) : r
    // markArea CHARGE — une seule par graphique, pilotée par la phase.
    const charging = live.phase === 'charge'
    const cs = charging ? live.phaseSince['charge'] ?? live.rtt95[0]?.[0] ?? Date.now() - 1000 : 0
    const ce = charging ? live.rtt95.at(-1)?.[0] ?? Date.now() : 0
    const ma = chargeMarkArea(cs, ce, charging)
    rtt.setData([
      { ...craftSeries(tri.chart, 'p50', d(live.rtt50 as any), CRAFT.live), markArea: ma } as any,
      { ...craftSeries(tri.chart, 'p95', d(live.rtt95 as any), CRAFT.ok) } as any,
    ])
    const series: ReturnType<typeof lineSeries>[] = [{ ...craftSeries(tri.chart, 'small p95', d(live.small as any), CRAFT.threshold), markArea: ma } as any]
    if (locked && lockedRingRef.current.length > 1) {
      series.unshift({
        name: 'baseline verrouillée', type: 'line', showSymbol: false, smooth: 0.4, smoothMonotone: 'x', sampling: 'lttb' as const,
        lineStyle: { width: 1.5, type: 'dashed' as const, color: CRAFT.steel },
        emphasis: { focus: 'series' }, blur: { lineStyle: { opacity: 0.2 } },
        data: d(lockedRingRef.current),
      } as any)
    }
    small.setData(series)
    goodput.setData([{ ...craftSeries(tri.chart, 'goodput', d(live.goodput as any), CRAFT.bbr), markArea: ma } as any])
  }, rafActive)

  const banner = replayRunning ? `Replay — ${replayRunId}`
    : !liveSnap ? 'Hors ligne — en attente du flux'
    : liveSnap.load_status === 'bulk-on' ? 'Charge — bulk actif'
    : liveSnap.phase === 'surveil' ? 'Surveillance — sondes légères'
    : liveSnap.phase === 'baseline' ? 'Baseline'
    : liveSnap.phase === 'recup' ? 'Récupération'
    : 'Idle'

  // p95 partiel honnête — le nombre est vrai (fenêtre courante), seulement
  // la complétude est indiquée : "en cours 42 s/120 s" si la phase tourne.
  const phaseTotalHint = liveSnap?.running && liveSnap.phase && liveSnap.phase_total_s
    ? <span className="mono live-phase-hint">
        {liveSnap.small_p95_ms > 0 ? `p95 partiel ${liveSnap.small_p95_ms.toFixed(1)} ms — ` : ''}
        {liveSnap.phase} en cours
      </span>
    : null
  // hors ligne = sourdine (pas de bandeau rouge pleine largeur ; le rouge reste en-tête une fois)
  const bannerColor = replayRunning ? 'var(--t-live)'
    : !liveSnap ? 'var(--text-muted)'
    : banner.startsWith('Charge') ? 'var(--t-threshold)'
    : banner.startsWith('Surveillance') ? 'var(--t-live)'
    : banner === 'Baseline' ? 'var(--t-live)'
    : banner === 'Récupération' ? 'var(--t-ok)'
    : 'var(--text-faint)'

  const qdiVal = (() => {
    const p95 = live.rtt95.at(-1)?.[1] ?? liveSnap?.rtt_p95_ms ?? 0
    const p50 = live.rtt50.at(-1)?.[1] ?? liveSnap?.rtt_p50_ms ?? 0
    return computeQDI(p95, p50)
  })()
  const jfiClient: number | null = (() => {
    const gVals = live.goodput.slice(-20).map(([, v]) => v).filter(v => v > 0.01)
    if (gVals.length >= 2) {
      const allEqual = gVals.every(v => v === gVals[0])
      if (!allEqual) return computeJFI(gVals)
    }
    const sVals = live.small.slice(-12).map(([, v]) => v).filter(v => Number.isFinite(v) && v > 0)
    if (sVals.length >= 2) {
      const allEqual = sVals.every(v => v === sVals[0])
      if (!allEqual) return computeJFI(sVals)
    }
    return null
  })()
  // JFI du serveur (0–100, mesuré par flux) prioritaire quand il existe ;
  // sinon repli sur les anneaux — l'échelle d'affichage suit la source.
  const jfiServerPct = liveSnap?.running && (liveSnap?.jfi_pct ?? 0) > 0 ? Number(liveSnap.jfi_pct) : 0
  const jfiVal: number | null = jfiServerPct > 0 ? jfiServerPct / 100 : jfiClient

  const rttP95 = liveSnap?.rtt_p95_ms ?? live.rtt95.at(-1)?.[1] ?? 0
  const rttP50 = liveSnap?.rtt_p50_ms ?? live.rtt50.at(-1)?.[1] ?? 0
  const smallP95 = liveSnap?.small_p95_ms ?? live.small.at(-1)?.[1] ?? 0
  const goodputVal = liveSnap?.bulk_goodput_mbps ?? live.goodput.at(-1)?.[1] ?? 0
  const drops = liveSnap?.drops ?? 0
  const idle = !liveSnap?.running && live.rtt95.length === 0 && live.small.length === 0
  const wasted: number | null = idle ? null : liveSnap?.wasted_bytes ?? null
  const costAr: number | null = idle ? null : liveSnap?.cost_ar_per_h ?? null
  // Coûts entreprise : GOLD (modèle A, capacité dépréciée) et APN 150 Ar/Go
  // (modèle B, facturé au volume) — mesurés par le serveur, jamais dérivés ici.
  const costPrimary: number | null = idle ? null : liveSnap?.cost_primary_ar_per_h ?? null
  const costB2B: number | null = idle ? null : liveSnap?.cost_b2b_ar_per_h ?? null
  const deadlineOk: number | null = idle ? null : liveSnap?.deadline_ok_pct ?? null
  const spark = (r: [number, number][]) => r.map(([, v]) => v).slice(-20)
  const trendOf = (arr: number[]): 'up' | 'down' | 'flat' => {
    if (arr.length < 2) return 'flat'
    const a = arr[arr.length - 2], b = arr[arr.length - 1]
    if (b > a * 1.05) return 'up'
    if (b < a * 0.95) return 'down'
    return 'flat'
  }
  const qdiSpark = live.rtt95.slice(-20).map(([, v], i) => Math.max(0, v - (live.rtt50[i]?.[1] ?? v)))

  // médiane live de la fenêtre courante — le chiffre du comparatif
  const liveSmallMedian = (() => {
    const vals = live.small.slice(-120).map(([, v]) => v).filter(v => v > 0)
    if (vals.length < 5) return null
    return vals.reduce((a, b) => a + b, 0) / vals.length
  })()
  const liveDiff = locked && liveSmallMedian ? Math.round(((locked.ms - liveSmallMedian) / locked.ms) * 100) : null

  const lockBaseline = () => {
    if (liveSmallMedian == null) return
    const b = { ms: Math.round(liveSmallMedian * 10) / 10, at: new Date().toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' }) }
    lockedRingRef.current = [...live.small] // freeze the ring — the hero draws it grey dashed
    setLocked(b)
    try { localStorage.setItem('wall-baseline', JSON.stringify(b)) } catch { }
    useUIStore.getState().pushToast?.(`baseline verrouillée — ${b.ms} ms`, 'ok')
  }
  const applyShape = async (q: string) => {
    setShapeMsg('')
    try {
      const st = loadSettings()
      const r = await fetch('/api/shape', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ qdisc: q, capacity_mbps: shapeCap, delay_ms: st.linkDelayMs, jitter_ms: st.linkJitterMs, loss_pct: st.linkLossPct }) })
      const j = await r.json()
      if (!r.ok) { setShapeMsg(j?.error ?? `erreur ${r.status}`); useUIStore.getState().pushToast?.(j?.error ?? 'échec du façonnage', 'err') }
      else {
        setShape({ applied: q !== 'none', qdisc: q === 'none' ? undefined : q, capacity_mbps: shapeCap })
        setShapeMsg(q === 'none' ? 'façonnage retiré' : `${q} appliqué au bord @ ${shapeCap} Mbit/s`)
        useUIStore.getState().pushToast?.(q === 'none' ? 'façonnage retiré' : `bord façonné — ${q}`, 'ok')
      }
    } catch (e) { setShapeMsg(String(e)) }
  }
  const runBurst = async () => {
    setBursting(true)
    try {
      const r = await fetch('/api/burst', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ cc: burstCc, seconds: 4 }) })
      const j = await r.json().catch(() => ({}))
      if (!r.ok) useUIStore.getState().pushToast?.(j?.error ?? `échec burst (${r.status})`, 'err')
      else useUIStore.getState().pushToast?.(`burst ${burstCc} traversé — lisez goodput et RTT`, 'ok')
    } catch (e) { useUIStore.getState().pushToast?.(String(e), 'err') } finally { setBursting(false) }
  }
  const toggleWatch = async () => {
    try {
      const r = await fetch('/api/watch', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ on: !watching }) })
      const j = await r.json().catch(() => ({}))
      if (!r.ok) { useUIStore.getState().pushToast?.(j?.error ?? 'échec surveillance', 'err'); return }
      // la réponse du serveur fait foi — un double-clic ou une course 409 ne doit
      // jamais désynchroniser la pilule de ce que le backend exécute
      setWatching(!!j.watch)
      if (j.watch) { setLocked(null); lockedRingRef.current = [] }
      useUIStore.getState().pushToast?.(j.watch ? 'surveillance active — le mur est vivant' : 'surveillance arrêtée', 'ok')
    } catch (e) { useUIStore.getState().pushToast?.(String(e), 'err') }
  }
  const toggleJournal = () => setJournalOpen(o => {
    try { localStorage.setItem('live-journal-open', o ? '0' : '1') } catch { }
    return !o
  })
  const exportConstat = () => {
    if (locked == null || liveSmallMedian == null) return
    const diff = Math.round(((locked.ms - liveSmallMedian) / locked.ms) * 100)
    const md = [
      '# Constat bufferbloat — Meteolink', '',
      `- baseline verrouillée : **${locked.ms} ms** (${locked.at})`,
      `- fenêtre courante : **${liveSmallMedian.toFixed(1)} ms**`,
      `- écart : **${diff > 0 ? '-' : ''}${diff} %**`,
      `- façonnage du bord : **${shape?.applied ? `${shape.qdisc} @ ${shape.capacity_mbps} Mbit/s` : 'aucun (pfifo)'}**`,
      `- empreinte campagne : ${hash8}`,
      `- généré : ${new Date().toLocaleString('fr-FR')}`, '',
      '_Kit de diagnostic portable — métrique small p95, fenêtre 180 s._',
    ].join('\n')
    const url = URL.createObjectURL(new Blob([md], { type: 'text/markdown' }))
    const a = document.createElement('a')
    a.href = url; a.download = `constat-bufferbloat-${Date.now()}.md`; a.click()
    URL.revokeObjectURL(url)
  }

  useEffect(() => { if (bannerRef.current) animateBannerPulse(bannerRef.current) }, [banner])
  useEffect(() => { animateLiveEnter() }, [])

  // superposition A/B — médianes pfifo vs meilleure CAKE, même échelle, badge d'écart, hash 8
  const baseline = wallGroups?.find(g => g.qdisc === 'pfifo_fast')
  const cakeBest = wallGroups?.find(g => g.qdisc === 'cake' && g.best) ?? wallGroups?.find(g => g.qdisc === 'cake') ?? wallGroups?.find(g => g.best) ?? null
  // échelle figée partagée — baseline et CAKE sortent du même gel CSV, le live
  // n'y figure plus (sa comparaison vit dans avant/maintenant ci-dessus : le
  // 738 figé n'écrase plus le live en copeau)
  const maxAB = Math.max(baseline?.small_p95_median ?? 0, cakeBest?.small_p95_median ?? 0, 1)
  const diffAB = baseline && cakeBest && baseline.small_p95_median > 0 ? Math.round(((baseline.small_p95_median - cakeBest.small_p95_median) / baseline.small_p95_median) * 100) : null
  const showLiveSrc = tri.source !== 'frozen'
  const showFrozenSrc = tri.source !== 'live'
  const heroEmpty = live.small.length === 0
  const rttEmpty = live.rtt95.length === 0
  const goodputEmpty = live.goodput.length === 0
  const ifaceRx = live.iface.rx.at(-1)?.[1] ?? null
  const ifaceTx = live.iface.tx.at(-1)?.[1] ?? null
  const ifaceFmt = (v: number | null) => (v == null ? '—' : v.toFixed(1))
  const ifaceSourceLabel = live.iface.source === 'twin' ? 'twin'
    : live.iface.source === 'prod' ? 'production' : 'source non déclarée'
  const ifaceDrops = live.iface.valid
    ? `pertes IF RX ${live.iface.rxDrops} · TX ${live.iface.txDrops}`
    : 'pertes IF RX — · TX —'
  // raisons connues en français ; une raison inconnue s'affiche telle quelle
  const GAP_FR: Record<string, string> = {
    warmup: 'démarrage', absent: 'interface absente', reset: 'compteur réinitialisé',
    slow: 'fenêtre trop longue', read: 'lecture impossible', clock: 'horloge',
  }
  const ifaceGap = live.iface.gap ? (GAP_FR[live.iface.gap] ?? live.iface.gap) : ''
  const ifaceRight = `${live.iface.name || '—'} · RX ${ifaceFmt(ifaceRx)} · TX ${ifaceFmt(ifaceTx)} Mbit/s · fenêtre ${live.iface.windowMs > 0 ? live.iface.windowMs.toFixed(0) : '—'} ms · ${ifaceDrops} · ${ifaceSourceLabel}${ifaceGap ? ' · ' + ifaceGap : ''}`
  const ifaceEmpty = live.iface.rx.length === 0
  const ifaceHint = !live.iface.name
    ? 'interface non observée — déclarer CGO_IFACE côté serveur'
    : live.iface.gap === 'absent' ? 'interface absente' : "en attente d'échantillons"
  const ifaceSub = !live.iface.name ? 'compteurs noyau'
    : live.iface.windowMs > 0 && live.iface.windowMs <= 150 ? 'compteurs noyau · échantillon 10 Hz · fenêtre 60 s'
    : live.iface.windowMs > 0 ? `compteurs noyau · échantillon ${live.iface.windowMs.toFixed(0)} ms`
    : 'compteurs noyau'
  // Timeline 48 — bandes de phase depuis live.phaseSince, bornes FIGÉES :
  // recupEnd = chargeEnd + 30s nominal — la bande ne grandit pas avec now.
  const hasData = live.rtt95.length > 0
  const now = live.rtt95.at(-1)?.[0] ?? Date.now()
  const t0 = live.phaseSince['baseline'] ?? live.rtt95[0]?.[0] ?? now - 1000
  const tCharge = Math.max(t0 + 1, live.phaseSince['charge'] ?? (live.phase === 'charge' ? t0 + 1 : now))
  const tRecup = Math.max(tCharge + 1, live.phaseSince['recup'] ?? (live.phase === 'recup' ? now : tCharge + 1000))
  // borne figée — la fenêtre totale de l'event (30+120+30) connue par construction
  const tEnd = tRecup + 30000

  // couture d'instrumentation: lire options/pixels depuis les sondes
  if (typeof window !== 'undefined') (window as any).__CGO_CHARTS = { rtt: rtt.chart.current, small: small.chart.current, goodput: goodput.chart.current, iface: iface.chart.current }
  // une seule voix d'attente partout — le vide ne parle qu'une langue
  const IDLE_HINT = 'en attente — démarrez une campagne depuis Campagne'
  return (
    <div id="wall" className="panel-stack live-stage">
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono live-peek-label">live p95 — point</div>
          <div className="mono live-peek-value">{peek.value.toFixed(1)} ms</div>
          <div className="live-peek-spark">
            {live.rtt95.slice(-20).map(([, v]: [number, number], i: number) => <i key={i} style={{ flex: 1, height: `${Math.max(2, Math.min(20, (v / 300) * 20))}px`, background: CRAFT.live, borderRadius: 1, opacity: 0.6 + i * 0.02 }} />)}
          </div>
        </PeekPopover>
      )}
      <div ref={bannerRef} className="banner mono" style={{ color: bannerColor, borderColor: bannerColor + '55', display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
        <span className="live-banner-msg">{banner}</span>
        <EventBadge />
        {phaseTotalHint}
      </div>
      {showLiveSrc && (
      <Card
        head="Petits objets p95"
        sub="p95 · fenêtre 180 s"
        className="live-hero-card"
        onMouseEnter={e => { setHovered(true); const v = live.small.at(-1)?.[1] ?? smallP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); small.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.small.length - 1) }) }}
        onMouseLeave={() => { setHovered(false); setPeek(null) }}
      >
        <div className="chart-flex">
          <ChartSurface title="Petits objets p95" unit="ms" domId="chart-small" height="100%" empty={heroEmpty} hint={IDLE_HINT} onReady={small.onReady} />
        </div>
      </Card>
      )}
      {!liveSnap && !showLiveSrc && <div className="card live-empty"><EmptyState kind="idle" hint={IDLE_HINT} /></div>}
      {showLiveSrc && (
        <Card
          head="Débit interface"
          sub={ifaceSub}
          className="live-hero-card"
          right={<span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>{ifaceRight}</span>}
        >
          <div className="chart-flex">
            <ChartSurface title="Interface" unit="Mbit/s" domId="chart-iface" height="100%" empty={ifaceEmpty} hint={ifaceHint} onReady={iface.onReady} />
          </div>
        </Card>
      )}
      {/* guide de lecture (U6a) — que disent ces chiffres, repliable, au-dessus du bento */}
      <LiveGuide />
      {/* Q4 metric pill — toggles the metric groups; one 5-col bento, 10 cells, no misaligned rows */}
      {/* Q4 metric pill — responsive bento owned by CSS (container queries:
          5 col → 2 @1100 → 1 @640); the inline gridTemplateColumns variant
          overrode the breakpoints and cropped every card on mobile */}
      {showLiveSrc && (
      <div data-wall-cards="metric-groups" className={'bento-5 wall-span' + (tri.metric ? '' : ' hidden')}>
        <MetricCard term="small_p95" label="Petits objets p95" value={smallP95 ? smallP95.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(smallP95, settings)]} spark={spark(live.small)} trend={trendOf(spark(live.small))} verdict={smallP95 ? verdictWord(latencyLevel(smallP95, settings)) : undefined} verdictLevel={smallP95 ? latencyLevel(smallP95, settings) : undefined} />
        <MetricCard term="rtt_p95" label="RTT p95" value={rttP95 ? rttP95.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(rttP95, settings)]} spark={spark(live.rtt95)} trend={trendOf(spark(live.rtt95))} verdict={rttP95 ? verdictWord(latencyLevel(rttP95, settings)) : undefined} verdictLevel={rttP95 ? latencyLevel(rttP95, settings) : undefined} />
        <MetricCard term="bulk_goodput" label="Débit utile" value={goodputVal ? goodputVal.toFixed(1) : '—'} unit="Mbit/s" color={LEVEL_COLOR[goodputLevel(goodputVal, settings.shapeCap)]} spark={spark(live.goodput)} trend={trendOf(spark(live.goodput))} verdict={goodputVal ? verdictWord(goodputLevel(goodputVal, settings.shapeCap)) : undefined} verdictLevel={goodputVal ? goodputLevel(goodputVal, settings.shapeCap) : undefined} />
        <MetricCard term="deadline_ok" label="Échéances" value={deadlineOk === null ? '—' : deadlineOk.toFixed(0)} unit={deadlineOk === null ? '' : '%'} color={deadlineOk === null ? 'var(--text-faint)' : LEVEL_COLOR[deadlineLevel(deadlineOk)]} trend="flat" spark={spark(live.deadline)} verdict={deadlineOk === null ? undefined : verdictWord(deadlineLevel(deadlineOk))} verdictLevel={deadlineOk === null ? undefined : deadlineLevel(deadlineOk)} threshold={{ marker: 95, max: 100 }} />
        <MetricCard term="rtt_p50" label="RTT médian" value={rttP50 ? rttP50.toFixed(1) : '—'} unit="ms" color={LEVEL_COLOR[latencyLevel(rttP50, settings)]} spark={spark(live.rtt50)} trend={trendOf(spark(live.rtt50))} verdict={rttP50 ? verdictWord(latencyLevel(rttP50, settings)) : undefined} verdictLevel={rttP50 ? latencyLevel(rttP50, settings) : undefined} />
        <div data-testid="qdi-sparkline"><MetricCard term="QDI" label="Dispersion" value={!liveSnap || live.rtt95.length === 0 ? '—' : qdiVal.toFixed(1)} unit="ms" color={CRAFT.threshold} spark={live.rtt95.length === 0 ? undefined : qdiSpark} trend={trendOf(qdiSpark)} /></div>
        <Card head="JFI" sub={jfiServerPct > 0 ? 'équité inter-flux mesurée' : 'équité inter-flux (rings, 0–1)'} testid="qdi-card" term="JFI">
          <div className="jfi-row">
            <span className="mono jfi-value" style={{ color: jfiVal === null ? 'var(--text-body)' : LEVEL_COLOR[jfiLevel(jfiVal)] }}>{jfiVal === null ? '—' : jfiServerPct > 0 ? `${jfiServerPct.toFixed(1)} %` : jfiVal.toFixed(2)}</span>
            <DonutJFI value={jfiVal} />
            {jfiVal !== null && <span className="mono jfi-badge" style={{ background: `color-mix(in srgb, ${LEVEL_COLOR[jfiLevel(jfiVal)]} 8%, transparent)`, color: LEVEL_COLOR[jfiLevel(jfiVal)] }}>{verdictWord(jfiLevel(jfiVal))}</span>}
          </div>
        </Card>
        <MetricCard term="drops" label="Pertes" value={idle ? '—' : String(drops)} unit="" color={LEVEL_COLOR[dropsLevel(drops)]} trend={drops > 0 ? 'up' : 'flat'} spark={spark(live.drops)} verdict={idle ? undefined : verdictWord(dropsLevel(drops))} verdictLevel={idle ? undefined : dropsLevel(drops)} />
        <MetricCard term="wasted" label="Gaspillé" value={formatBytesFR(wasted).num} unit={formatBytesFR(wasted).unit} color={wasted == null ? 'var(--text-faint)' : CRAFT.threshold} trend={wasted != null && wasted > 0 ? 'up' : 'flat'} spark={spark(live.wasted)} />
        <MetricCard term="cost_primary_ar_per_h" label="Coût entreprise" value={formatCostAR(costPrimary).num} unit={formatCostAR(costPrimary).unit} color={costPrimary == null ? 'var(--text-faint)' : CRAFT.threshold} trend="flat" />
        <MetricCard term="cost_b2b_ar_per_h" label="Coût B2B" value={formatCostAR(costB2B).num} unit={formatCostAR(costB2B).unit} color={costB2B == null ? 'var(--text-faint)' : CRAFT.threshold} trend="flat" />
        <MetricCard term="cost_ar_per_h" label="Coût détail (gelé)" value={formatCostAR(costAr).num} unit={formatCostAR(costAr).unit} color={costAr == null ? 'var(--text-faint)' : CRAFT.threshold} trend={costAr != null && costAr > 0 ? 'up' : 'flat'} spark={spark(live.cost)} />
      </div>
      )}
      {showLiveSrc && (
      <div className="wall-span duo wall-duo">
        <Card
          head="RTT"
          sub="p50 · p95"
          onMouseEnter={e => { setHovered(true); const v = live.rtt95.at(-1)?.[1] ?? rttP95; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); rtt.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.rtt95.length - 1) }) }}
          onMouseLeave={() => { setHovered(false); setPeek(null) }}
        >
          <ChartSurface title="RTT" unit="ms" domId="chart-rtt" height={180} empty={rttEmpty} hint={IDLE_HINT} onReady={rtt.onReady} />
        </Card>
        <Card
          head="Débit utile applicatif"
          sub="mesuré au récepteur (compteur noyau)"
          onMouseEnter={e => { setHovered(true); const v = live.goodput.at(-1)?.[1] ?? goodputVal; setPeek({ rect: e.currentTarget.getBoundingClientRect(), value: v }); goodput.chart.current?.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: Math.max(0, live.goodput.length - 1) }) }}
          onMouseLeave={() => { setHovered(false); setPeek(null) }}
        >
          <ChartSurface title="Débit utile applicatif" unit="Mbit/s" domId="chart-goodput" height={180} empty={goodputEmpty} hint={IDLE_HINT} onReady={goodput.onReady} />
        </Card>
      </div>
      )}
      {/* Q2 — runbook pointer: a crit card always says what to do next.
          Contrôle de forme — prescrire CAKE alors que CAKE est déjà appliqué
          au bord contredit l'état propre (preuve shape=cake aux shots). */}
      {showLiveSrc && rttP95 != null && rttP95 > settings.critMs && !(shape?.applied && shape.qdisc === 'cake') && (
        <div className="mono live-runbook" data-testid="runbook-pointer">
          bufferbloat détecté (RTT p95 {rttP95.toFixed(0)} ms &gt; {settings.critMs}) → appliquez CAKE via Façonnage du bord · traduction MikroTik : queue type cake
        </div>
      )}
      {hasData && showLiveSrc && <Timeline baselineStart={t0} chargeStart={tCharge} chargeEnd={tRecup} recupEnd={tEnd} currentPhase={live.phase || 'idle'} />}

      {/* live-wall-overlay: baseline gris pointillé vs CAKE violet uni, même
          échelle figée ; le live n'y figure plus — sa comparaison vit dans
          avant/maintenant (même fenêtre), le 738 figé ne l'écrase plus */}
      <div className="live-wall-overlay card live-overlay" data-testid="live-wall-overlay">
        <CardHead
          label={`Figée vs appliqué — ${baseline?.qdisc ?? 'pfifo'} → ${shape?.applied ? shape.qdisc : 'sans façonnage'}`}
          sub={`source ${tri.source} · hash ${hash8} · ${wallGroups ? `${wallGroups.length} groupes gelés` : 'en attente'}`}
          right={diffAB != null ? (
            <span className="diff-badge mono" style={{ background: diffAB > 0 ? 'rgba(31,163,72,0.12)' : 'rgba(226,39,24,0.12)', border: '1px solid ' + (diffAB > 0 ? CRAFT.ok : CRAFT.danger), color: diffAB > 0 ? CRAFT.ok : CRAFT.danger, padding: '4px 10px', fontSize: 'var(--fs-body)', fontWeight: 700, fontVariantNumeric: 'tabular-nums' }}>{diffAB > 0 ? `-${diffAB}%` : `${diffAB}%`}</span>
          ) : undefined}
        />
        {diffAB != null && (
          <div className="mono live-overlay-note">
            {diffAB > 0
              ? `cake répond ${diffAB} % plus vite que la file simple — médianes figées, même campagne`
              : `la file simple répond ${Math.abs(diffAB)} % plus vite — médianes figées, même campagne`}
          </div>
        )}
        {/* edge control — trois sections calmes : Surveillance / Façonnage / Essai */}
        <div className="edge-control" data-idle={idle ? '1' : '0'} data-testid="edge-control">
          <section className="edge-section">
            <CardHead label="Surveillance" sub={watching ? 'sondes légères en continu' : 'sondes arrêtées'} />
            <div className="edge-row">
              <Pill label="surveillance" value={watching ? 'on' : 'off'} on={watching} onClick={toggleWatch} title="sondes légères en continu (sans bulk) — rend l'effet du façonnage visible" />
              <Pill label="conditions" value={linkOpen ? 'masquer' : `${settings.linkDelayMs} ms`} on={linkOpen} onClick={() => setLinkOpen(!linkOpen)} title="conditions du lien — délai/gigue/perte appliqués au bord" />
              <button className="btn edge-btn" onClick={lockBaseline} disabled={liveSmallMedian == null}>FIGER L'AVANT{liveSmallMedian != null ? ` — ${liveSmallMedian.toFixed(1)} ms` : ''}</button>
              {locked && <Pill label="avant" value={`${locked.ms} ms @ ${locked.at}`} on title="avant figé — cliquer pour libérer" onClick={() => { setLocked(null); try { localStorage.removeItem('wall-baseline') } catch { } }} />}
              <button className="btn btn-primary edge-btn edge-btn-end" onClick={exportConstat} disabled={locked == null || liveSmallMedian == null}>TÉLÉCHARGER LE CONSTAT</button>
            </div>
            {shapeMsg && <div className="mono edge-msg" style={{ color: shapeMsg.includes('erreur') || shapeMsg.includes('échec') ? 'var(--t-danger-text)' : 'var(--t-ok)' }}>{shapeMsg}</div>}
            {locked && showLiveSrc && liveSmallMedian != null && (
              <div className="edge-ab">
                <div className="mono edge-ab-head">
                  <span>avant → maintenant, même fenêtre</span>
                  <span style={{ fontVariantNumeric: 'tabular-nums', color: liveDiff != null && liveDiff > 0 ? 'var(--t-ok)' : 'var(--t-danger-text)' }}>{liveDiff != null ? `${liveDiff > 0 ? '-' : '+'}${Math.abs(liveDiff)} %` : '—'}</span>
                </div>
                <div className="edge-bar edge-bar-before">
                  <div className="edge-fill edge-fill-before" style={{ transform: `scaleX(${Math.min(1, locked.ms / Math.max(locked.ms, liveSmallMedian))})` }} />
                </div>
                <div className="edge-bar edge-bar-after">
                  <div className="edge-fill edge-fill-after" style={{ transform: `scaleX(${Math.min(1, liveSmallMedian / Math.max(locked.ms, liveSmallMedian))})` }} />
                </div>
              </div>
            )}
          </section>

          <section className="edge-section">
            <CardHead label="Façonnage" sub={shape?.applied ? `bord façonné : ${shape.qdisc} @ ${shape.capacity_mbps} Mbit/s` : 'bord non façonné (file simple)'} />
            <div className="edge-row">
              <Explain term="capacity">
                <input
                  type="number" min={1} max={1000} placeholder="20" value={shapeCap}
                  onChange={e => setShapeCap(parseInt(e.target.value) || 20)}
                  aria-label="capacité du bord (Mbit/s)"
                  className="c-input edge-cap"
                />
              </Explain>
              <span className="mono edge-note">Mbit/s</span>
              {/* une campagne active possède le shaper — le levier répond 409 :
                  on grise au lieu de laisser l'opérateur lever une erreur */}
              {['none', 'fq_codel', 'cake'].map(q => (
                <span
                  key={q}
                  title={liveSnap?.running ? "campagne active — le façonnage redevient disponible à l'arrêt" : `appliquer ${q} au bord`}
                  style={{ opacity: liveSnap?.running ? 'var(--disabled-opacity)' : 1, cursor: liveSnap?.running ? 'not-allowed' : 'pointer' }}
                >
                  <Pill
                    label={q === 'none' ? 'sans' : q}
                    on={shape?.applied && shape.qdisc === q}
                    onClick={liveSnap?.running ? undefined : () => applyShape(q)}
                    title={liveSnap?.running ? 'campagne active — arrêtez la mesure pour façonner' : `appliquer ${q} au bord`}
                  />
                </span>
              ))}
            </div>
          </section>

          <section className="edge-section">
            <CardHead label="Essai" sub={liveSnap?.running ? 'indisponible pendant la campagne' : 'traverse le bord façonné — regardez goodput et RTT'} />
            <div className="edge-row">
              <span className="mono edge-note"><Explain term="burst">Burst test</Explain></span>
              {['cubic', 'bbr'].map(c => (
                <button key={c} className="btn edge-btn" aria-pressed={burstCc === c} onClick={() => setBurstCc(c)}>{c}</button>
              ))}
              <button className="btn btn-primary edge-btn" onClick={runBurst} disabled={bursting || !!liveSnap?.running}>
                {bursting ? `burst ${burstCc}…` : `TESTER ${burstCc.toUpperCase()}`}
              </button>
            </div>
          </section>
        </div>
        {showFrozenSrc && (
          <div className="edge-frozen">
            <div>
              <div className="mono edge-ab-head">
                <span>baseline figée — gris pointillé</span>
                <span className="edge-ab-num">{baseline ? `${baseline.small_p95_median.toFixed(1)} ms` : '—'}</span>
              </div>
              <div className="edge-bar edge-bar-before">
                <div className="edge-fill edge-fill-before" style={{ transform: `scaleX(${Math.min(1, baseline ? baseline.small_p95_median / maxAB : 0)})` }} />
              </div>
            </div>
            <div>
              <div className="mono edge-ab-head edge-ab-head-cake">
                <span>{shape?.applied ? shape.qdisc : "cake"} — violet uni{cakeBest?.best ? ' ★' : ''}</span>
                <span className="edge-ab-num">{cakeBest ? `${cakeBest.small_p95_median.toFixed(1)} ms` : '—'}</span>
              </div>
              <div className="edge-bar edge-bar-cake">
                <div className="edge-fill edge-fill-cake" style={{ transform: `scaleX(${Math.min(1, cakeBest ? cakeBest.small_p95_median / maxAB : 0)})` }} />
              </div>
            </div>
          </div>
        )}
      </div>

      {journal.length > 0 && (
        <div className="card live-journal" data-testid="journal">
          <div className="live-journal-head">
            <CardHead label="Journal" sub={`${Math.min(journal.length, 12)} affichés sur ${journalTotal} reçus`} />
            <button className="btn live-journal-toggle" onClick={toggleJournal} aria-expanded={journalOpen} aria-controls="live-journal-body">
              {journalOpen ? 'masquer' : 'afficher'}
            </button>
          </div>
          {journalOpen && (
            <div id="live-journal-body" className="live-journal-body">
              {journal.slice(0, 12).map((e, i) => (
                <div key={i} className="mono live-journal-row">
                  <span className="live-journal-ts">{e.ts.slice(11, 19)}</span>
                  <span className="live-journal-kind">{e.kind}</span>
                  <span className="ev-msg live-journal-msg" title={e.msg}>{truncateWords(e.msg, 120)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
      {liveSnap?.running && <Beam hovered={hovered} />}
    </div>
  )
}
