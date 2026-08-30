import { useEffect, useMemo, useRef, useState } from 'react'
import { echarts } from '../lib/echarts'
import { baseOption, lineSeries, CRAFT } from '../lib/chartGrammar'
import { CardHead } from './ui/CardHead'
import { EmptyState } from './ui/EmptyState'

// Métriques du live — Comparaison BBR×AQM « pin A/B » : deux cellules gelées, traces
// alignées par événement, table d'écart (médianes, p95, pertes, coût Ariary),
// verdict, exports CSV/JSON. Alimentée par /api/run/rows (provenance figée).
type Row = Record<string, string>
export type Pinned = { profile: string; qdisc: string; cc: string }

const METRICS = [
  { key: 'small_p95_ms', label: 'small p95 (ms)', dir: 'down' as const },
  { key: 'rtt_p95_ms', label: 'RTT p95 (ms)', dir: 'down' as const },
  { key: 'bulk_goodput_mbps', label: 'goodput (Mbit/s)', dir: 'up' as const },
  { key: 'drops', label: 'pertes', dir: 'down' as const },
  { key: 'cost_ar_per_h', label: 'coût (Ar/h)', dir: 'down' as const },
]

const cellName = (p: Pinned) => `${p.profile}/${p.qdisc}/${p.cc}`
const num = (r: Row, k: string) => parseFloat(r[k] ?? '0') || 0
const matches = (r: Row, p: Pinned) => r.profile === p.profile && r.qdisc === p.qdisc && r.cc === p.cc

function cellMedian(rows: Row[], p: Pinned, k: string): { med: number; n: number } {
  const cell = rows.filter(r => matches(r, p))
  const v = (cell.length ? cell : rows).map(r => num(r, k)).filter(x => x > 0 || k === 'drops').sort((x, y) => x - y)
  return { med: v.length ? v[Math.floor(v.length / 2)] : 0, n: cell.length }
}

export default function CompareView({ a, b, onClose }: { a: Pinned; b: Pinned; onClose: () => void }) {
  const chartRef = useRef<HTMLDivElement>(null)
  const [err, setErr] = useState<string | null>(null)
  const [runs, setRuns] = useState<string[]>([])
  const [runA, setRunA] = useState<string | null>(null)
  const [runB, setRunB] = useState<string | null>(null)
  const [rows, setRows] = useState<{ a: Row[]; b: Row[] } | null>(null)

  useEffect(() => {
    fetch('/api/replay/list').then(r => r.json()).then(j => {
      const ids: string[] = j?.runs ?? []
      setRuns(ids)
      // l'opérateur vient de lancer la campagne — les runs les plus récents
      // contiennent le plus sûrement les cellules épinglées
      setRunA(ids[ids.length - 1] ?? null)
      setRunB(ids[ids.length - 2] ?? ids[ids.length - 1] ?? null)
    }).catch(e => setErr(String(e)))
  }, [])

  useEffect(() => {
    if (!runA || !runB) return
    let cancelled = false
    // a campagne stopped early freezes a header-only CSV → 404 "run vide" —
    // ce côté est simplement vide (avertissement ambre), jamais fatal.
    const load = async (id: string): Promise<Row[]> => {
      const r = await fetch(`/api/run/rows?run=${id}`)
      if (r.status === 404) return []
      if (!r.ok) throw new Error(`rows ${id}: HTTP ${r.status}`)
      return ((await r.json()).rows ?? []) as Row[]
    }
    Promise.all([load(runA), load(runB)]).then(([ra, rb]) => {
      if (cancelled) return
      setRows({ a: ra, b: rb })
    }).catch(e => setErr(String(e)))
    return () => { cancelled = true }
  }, [runA, runB])

  // auto-résolution des runs contenant les cellules épinglées: l'épingle
  // vient de l'agrégat; sonder du plus récent au plus ancien (12 max)
  // jusqu'à un run contenant réellement la cellule.
  useEffect(() => {
    if (!rows || !runs.length) return
    const find = async (side: 'a' | 'b', pin: Pinned, current: string | null) => {
      if (!current) return
      const sideRows = side === 'a' ? rows.a : rows.b
      if (sideRows.some(r => matches(r, pin))) return
      const candidates = [...runs].reverse().filter(id => id !== current).slice(0, 12)
      for (const id of candidates) {
        try {
          const rr = await fetch(`/api/run/rows?run=${id}`)
          if (!rr.ok) continue
          const jrows = ((await rr.json()).rows ?? []) as Row[]
          if (jrows.some(r => matches(r, pin))) {
            if (side === 'a') setRunA(id)
            else setRunB(id)
            return
          }
        } catch { /* skip unreachable run */ }
      }
    }
    find('a', a, runA)
    find('b', b, runB)
  }, [rows, runs, a, b, runA, runB])

  const stats = useMemo(() => {
    if (!rows) return null
    const side = (rs: Row[], p: Pinned) =>
      Object.fromEntries(METRICS.map(m => [m.key, cellMedian(rs, p, m.key)])) as Record<string, { med: number; n: number }>
    return { a: side(rows.a, a), b: side(rows.b, b) }
  }, [rows, a, b])

  useEffect(() => {
    if (!rows || !chartRef.current) return
    const c = echarts.init(chartRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => { try { c.resize() } catch { } })
    ro.observe(chartRef.current)
    const mk = (rs: Row[]): [number, number][] => rs.map((r, i) => [i, num(r, 'small_p95_ms')])
    const base = baseOption('small p95 — événements alignés', 'ms')
    c.setOption({
      ...base,
      xAxis: { ...base.xAxis, type: 'value' as const, name: 'événement' },
      legend: { textStyle: { color: '#8b9099', fontSize: 10, fontFamily: 'JetBrains Mono' }, top: 4 },
      series: [
        { ...lineSeries(`A — ${cellName(a)}`, mk(rows.a.filter(r => matches(r, a))), CRAFT.live), name: `A — ${cellName(a)}` },
        { ...lineSeries(`B — ${cellName(b)}`, mk(rows.b.filter(r => matches(r, b))), CRAFT.ok), name: `B — ${cellName(b)}` },
      ],
    } as any)
    return () => { ro.disconnect(); try { c.dispose() } catch { } }
  }, [rows, a, b])

  if (err) return <div className="card"><EmptyState kind="error" hint={err} /></div>

  const fmt = (v: number) => v.toFixed(1)
  const verdict = (() => {
    if (!stats) return null
    const gains = METRICS.map(m => ({ m, better: m.dir === 'down' ? stats.b[m.key].med < stats.a[m.key].med : stats.b[m.key].med > stats.a[m.key].med }))
    const wins = gains.filter(g => g.better).length
    const p95a = stats.a.small_p95_ms.med, p95b = stats.b.small_p95_ms.med
    const pct = p95a > 0 ? Math.round(((p95a - p95b) / p95a) * 100) : 0
    return `${wins}/${METRICS.length} métriques en faveur de B — small p95 ${pct >= 0 ? `-${pct}` : `+${Math.abs(pct)}`} %`
  })()

  const exportCSV = () => {
    if (!stats) return
    const lines = ['metric,A,B,unite', ...METRICS.map(m => `${m.key},${fmt(stats.a[m.key].med)},${fmt(stats.b[m.key].med)},${m.label}`)].join('\n')
    const url = URL.createObjectURL(new Blob([lines], { type: 'text/csv' }))
    const el = document.createElement('a'); el.href = url; el.download = `comparaison-${cellName(a).replace(/\//g, '-')}-vs-${cellName(b).replace(/\//g, '-')}.csv`; el.click()
    URL.revokeObjectURL(url)
  }
  const prescription = (() => {
    if (!stats) return null
    // La prescription découle du verdict mesuré — jamais d'intuition.
    const p95a = stats.a.small_p95_ms.med, p95b = stats.b.small_p95_ms.med
    const winner = p95b < p95a ? b : a
    const other = p95b < p95a ? a : b
    if (winner.qdisc === other.qdisc) return null // pas de changement à prescrire
    const routeros: Record<string, string> = {
      cake: '/queue type add name=cake-edges kind=cake\n/queue simple add name=edge target=<LAN> queue=cake-edges',
      fq_codel: '/queue type add name=fqc-edges kind=fq-codel\n/queue simple add name=edge target=<LAN> queue=fqc-edges',
      pfifo_fast: '/queue simple add name=edge target=<LAN> queue=pfifo-fast',
    }
    const linux: Record<string, string> = {
      cake: 'tc qdisc replace dev <WAN> root cake bandwidth <CAP>',
      fq_codel: 'tc qdisc replace dev <WAN> root fq_codel',
      pfifo_fast: 'tc qdisc replace dev <WAN> root pfifo_fast',
    }
    return {
      gagnant: cellName(winner),
      queue_type: winner.qdisc,
      mikrotik_v7: routeros[winner.qdisc] ?? null,
      linux_tc: linux[winner.qdisc] ?? null,
      note: 'prescription calculée depuis les médianes mesurées — appliquez, puis re-mesurez (audit + comparaison avant/après)',
    }
  })()

  const exportJSON = () => {
    if (!stats) return
    const payload = {
      A: { cellule: cellName(a), ...Object.fromEntries(METRICS.map(m => [m.key, stats.a[m.key].med])) },
      B: { cellule: cellName(b), ...Object.fromEntries(METRICS.map(m => [m.key, stats.b[m.key].med])) },
      verdict, prescription, genere: new Date().toISOString(),
    }
    const url = URL.createObjectURL(new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' }))
    const el = document.createElement('a'); el.href = url; el.download = `comparaison-${cellName(a).replace(/\//g, '-')}-vs-${cellName(b).replace(/\//g, '-')}.json`; el.click()
    URL.revokeObjectURL(url)
  }

  const runSelect = (value: string | null, set: (v: string) => void, side: string) => (
    <select value={value ?? ''} onChange={e => set(e.target.value)} aria-label={`run ${side}`} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '4px 6px', fontFamily: 'var(--font-mono)', fontSize: 10, maxWidth: 200 }}>
      {runs.map(r => <option key={r} value={r}>{r}</option>)}
    </select>
  )

  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: 12 }} data-testid="compare-view">
      <CardHead
        label={`Comparaison — ${cellName(a)} vs ${cellName(b)}`}
        sub="cellules figées, même échelle — choisissez le run de chaque côté si la cellule est absente"
        right={<button className="btn" onClick={onClose} style={{ padding: '4px 10px', fontSize: 10 }}>FERMER</button>}
      />
      <div className="mono" style={{ fontSize: 10, color: 'var(--text-muted)', display: 'flex', gap: 16, flexWrap: 'wrap' }}>
        <span>A : {runA ?? '—'} {stats && stats.a.small_p95_ms.n === 0 ? <b style={{ color: CRAFT.threshold }}>(cellule absente de ce run — médiane sur tout le run)</b> : null}</span>
        <span>B : {runB ?? '—'} {stats && stats.b.small_p95_ms.n === 0 ? <b style={{ color: CRAFT.threshold }}>(cellule absente de ce run)</b> : null}</span>
      </div>
      {verdict && (
        <div className="mono" style={{ fontSize: 12, color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '8px 12px', background: 'rgba(90,211,227,0.05)' }}>
          Verdict — {verdict}
        </div>
      )}
      <div ref={chartRef} style={{ height: 240 }} />
      <table className="data-table" style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-mono)', fontSize: 12 }}>
        <thead><tr style={{ color: 'var(--text-muted)', textAlign: 'left', borderBottom: '1px solid var(--hairline)' }}>
          <th style={{ padding: '6px 8px' }}>métrique</th><th style={{ padding: '6px 8px' }}>A</th><th style={{ padding: '6px 8px' }}>B</th><th style={{ padding: '6px 8px' }}>écart</th>
        </tr></thead>
        <tbody>
          {METRICS.map(m => {
            if (!stats) return null
            const va = stats.a[m.key].med, vb = stats.b[m.key].med
            const better = m.dir === 'down' ? vb < va : vb > va
            const pct = va > 0 ? Math.round(((va - vb) / va) * 100) : 0
            return (
              <tr key={m.key} style={{ borderBottom: '1px solid var(--hairline-faint)' }}>
                <td style={{ padding: '6px 8px', textAlign: 'left' }}>{m.label}</td>
                <td style={{ padding: '6px 8px' }}>{fmt(va)}</td>
                <td style={{ padding: '6px 8px', color: better ? CRAFT.ok : 'var(--text-body)' }}>{fmt(vb)}</td>
                <td style={{ padding: '6px 8px', color: pct > 0 ? CRAFT.ok : CRAFT.danger }}>{pct > 0 ? `-${pct} %` : `+${Math.abs(pct)} %`}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
      <div className="form-row" style={{ gap: 8, alignItems: 'center' }}>
        <span className="mono" style={{ fontSize: 10, color: 'var(--text-muted)' }}>run A :</span>
        {runSelect(runA, setRunA, 'A')}
        <span className="mono" style={{ fontSize: 10, color: 'var(--text-muted)', marginLeft: 12 }}>run B :</span>
        {runSelect(runB, setRunB, 'B')}
        <button className="btn" onClick={exportCSV} style={{ padding: '4px 10px', fontSize: 10, marginLeft: 'auto' }}>CSV</button>
        <button className="btn" onClick={exportJSON} style={{ padding: '4px 10px', fontSize: 10 }}>JSON</button>
      </div>
    </div>
  )
}
