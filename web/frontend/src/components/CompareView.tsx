import { useEffect, useMemo, useRef, useState } from 'react'
import { echarts } from '../lib/echarts'
import { baseOption, lineSeries, CRAFT } from '../lib/chartGrammar'
import { CardHead } from './ui/CardHead'
import { EmptyState } from './ui/EmptyState'

// Q14 — Comparaison BBR×AQM « pin A/B » : deux cellules gelées, traces
// phase-alignées, table d'écart (médianes, p95, pertes, coût Ariary), verdict,
// exports JSON/CSV/MD. Alimentée par /api/run/rows (provenance figée).
type Row = Record<string, string>
type Pinned = { run: string; row: Row }

const METRICS = [
  { key: 'small_p95_ms', label: 'small p95 (ms)', dir: 'down' as const },
  { key: 'rtt_p95_ms', label: 'RTT p95 (ms)', dir: 'down' as const },
  { key: 'bulk_goodput_mbps', label: 'goodput (Mbit/s)', dir: 'up' as const },
  { key: 'drops', label: 'pertes', dir: 'down' as const },
  { key: 'cost_ar_per_h', label: 'coût (Ar/h)', dir: 'down' as const },
]

const cellName = (r: Row) => `${r.profile}/${r.qdisc}/${r.cc}`
const num = (r: Row, k: string) => parseFloat(r[k] ?? '0') || 0

export default function CompareView({ a, b, onClose }: { a: Pinned; b: Pinned; onClose: () => void }) {
  const chartRef = useRef<HTMLDivElement>(null)
  const [err, setErr] = useState<string | null>(null)
  const [rows, setRows] = useState<{ a: Row[]; b: Row[] } | null>(null)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      fetch(`/api/run/rows?run=${a.run}`).then(r => r.json()),
      fetch(`/api/run/rows?run=${b.run}`).then(r => r.json()),
    ]).then(([ja, jb]) => {
      if (cancelled) return
      setRows({ a: (ja.rows ?? []) as Row[], b: (jb.rows ?? []) as Row[] })
    }).catch(e => setErr(String(e)))
    return () => { cancelled = true }
  }, [a.run, b.run])

  // médianes par métrique, uniquement sur les cellules épinglées
  const stats = useMemo(() => {
    if (!rows) return null
    const med = (rs: Row[], k: string) => {
      const v = rs.map(r => num(r, k)).filter(v => v > 0 || k === 'drops').sort((x, y) => x - y)
      return v.length ? v[Math.floor(v.length / 2)] : 0
    }
    return { a: Object.fromEntries(METRICS.map(m => [m.key, med(rows.a, m.key)])), b: Object.fromEntries(METRICS.map(m => [m.key, med(rows.b, m.key)])) }
  }, [rows])

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
        { ...lineSeries(`A — ${cellName(a.row)}`, mk(rows.a), CRAFT.live), name: `A — ${cellName(a.row)}` },
        { ...lineSeries(`B — ${cellName(b.row)}`, mk(rows.b), CRAFT.ok), name: `B — ${cellName(b.row)}` },
      ],
    } as any)
    return () => { ro.disconnect(); try { c.dispose() } catch { } }
  }, [rows, a, b])

  if (err) return <div className="card"><EmptyState kind="error" hint={err} /></div>

  const fmt = (v: number) => v.toFixed(1)
  const verdict = (() => {
    if (!stats) return null
    const gains = METRICS.filter(m => m.key !== 'bulk_goodput_mbps')
      .map(m => ({ m, better: m.dir === 'down' ? stats.b[m.key] < stats.a[m.key] : stats.b[m.key] > stats.a[m.key] }))
    const wins = gains.filter(g => g.better).length
    const p95a = stats.a.small_p95_ms, p95b = stats.b.small_p95_ms
    const pct = p95a > 0 ? Math.round(((p95a - p95b) / p95a) * 100) : 0
    return `${wins}/${METRICS.length} métriques en faveur de B — small p95 ${pct > 0 ? `-${pct} %` : `+${Math.abs(pct)} %`}`
  })()

  const exportCSV = () => {
    if (!stats) return
    const lines = ['metric,A,B,unit', ...METRICS.map(m => `${m.key},${fmt(stats.a[m.key])},${fmt(stats.b[m.key])},`)].join('\n')
    const url = URL.createObjectURL(new Blob([lines], { type: 'text/csv' }))
    const el = document.createElement('a'); el.href = url; el.download = `compare-${a.run}-${b.run}.csv`; el.click()
    URL.revokeObjectURL(url)
  }
  const exportJSON = () => {
    if (!stats) return
    const payload = { A: { run: a.run, cell: cellName(a.row), ...stats.a }, B: { run: b.run, cell: cellName(b.row), ...stats.b }, verdict }
    const url = URL.createObjectURL(new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' }))
    const el = document.createElement('a'); el.href = url; el.download = `compare-${a.run}-${b.run}.json`; el.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: 12 }} data-testid="compare-view">
      <CardHead
        label={`Comparaison — ${cellName(a.row)} vs ${cellName(b.row)}`}
        sub={`runs gelés : ${a.run} · ${b.run}`}
        right={<button className="btn" onClick={onClose} style={{ padding: '4px 10px', fontSize: 10 }}>FERMER</button>}
      />
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
            const va = stats.a[m.key], vb = stats.b[m.key]
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
      <div className="form-row" style={{ gap: 8 }}>
        <button className="btn" onClick={exportCSV} style={{ padding: '4px 10px', fontSize: 10 }}>CSV</button>
        <button className="btn" onClick={exportJSON} style={{ padding: '4px 10px', fontSize: 10 }}>JSON</button>
        <span className="mono muted" style={{ fontSize: 10 }}>métrique small p95 alignée par événement — médianes sur les cellules épinglées</span>
      </div>
    </div>
  )
}
