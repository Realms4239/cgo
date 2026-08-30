import { useEffect, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar } from '../lib/anime'
import { PeekPopover } from '../components/PeekPopover'
import { echarts } from '../lib/echarts'
import { baseOption, scatterSeries } from '../lib/chartGrammar'
import CompareView, { type Pinned } from '../components/CompareView'
import { CRAFT } from '../lib/chartGrammar'
import Explain from '../components/Explain'

type Group = {
  profile: string; qdisc: string; cc: string
  count: number; quarantined: number
  rtt_p95_median: number; small_p95_median: number
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

export default function ResultatsView() {
  const [groups, setGroups] = useState<Group[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [peek, setPeek] = useState<{ rect: DOMRect; g: Group } | null>(null)
  const [hash8, setHash8] = useState<string>('────────')
  const [pinA, setPinA] = useState<Pinned | null>(null)
  const [pinB, setPinB] = useState<Pinned | null>(null)
  const [rankKey, setRankKey] = useState<string>('small_p95_median')
  const [fProfile, setFProfile] = useState('tous')
  const [fQdisc, setFQdisc] = useState('tous')
  const [fCc, setFCc] = useState('tous')
  const scatterRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetch('/api/results').then(r => r.json()).then(j => {
      if (j.available) setGroups(j.groups)
      else setErr(j.reason || 'pas de résultats')
    }).catch(e => setErr(String(e)))
    fetch('/api/integrity').then(r => r.json()).then(j => {
      // triple-provenance: hash8 = sha256(latest aqm_eval.csv)[:8]; fallback run id when absent
      const id = j?.hash8 ?? String(j?.run_ids?.[0] ?? '').slice(0, 8)
      if (id) setHash8(String(id).slice(0, 8))
    }).catch(() => {})
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
    const c = echarts.init(scatterRef.current, undefined, { renderer: 'canvas', useDirtyRect: true } as any)
    const ro = new ResizeObserver(() => c.resize())
    ro.observe(scatterRef.current)
    // through the grammar — same hairline base as the Wall, craft scatter, no chrome
    const base = baseOption('compromis latence / débit', 'ms')
    const bestIdx = groups.map((g, i) => g.best ? i : -1).filter(i => i >= 0)
    const opt = {
      ...base,
      // value axes override (base defaults to time) — grammar hairlines kept
      xAxis: { ...base.xAxis, type: 'value' as const, name: 'goodput (Mbit/s)' },
      yAxis: { ...base.yAxis, name: 'small p95 (ms)' },
      tooltip: { ...base.tooltip, trigger: 'item' as const },
      series: [scatterSeries('groupes', groups.map(g => [g.goodput_median, g.small_p95_median] as [number, number]), '#5ad3e3', bestIdx)],
    }
    c.setOption(opt as any)
    return () => { ro.disconnect(); c.dispose() }
  }, [groups])

  if (err) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="error" hint={err} /></div>
  if (!groups) return <div className="card"><h1 className="view-title">Résultats</h1><EmptyState kind="loading" hint="agrégation des réplications" /></div>

  const rankMeta = RANKS.find(r => r.key === rankKey) ?? RANKS[0]
  const val = (g: Group): number => {
    if (rankKey === 'cost') return g.cost_median ?? g.cost_ar_per_h ?? Number.MAX_SAFE_INTEGER
    const v = (g as any)[rankKey]
    return typeof v === 'number' && v > 0 ? v : Number.MAX_SAFE_INTEGER
  }
  const distinct = (k: 'profile' | 'qdisc' | 'cc') => Array.from(new Set(groups.map(g => g[k]))).sort()
  const filtered = groups.filter(g =>
    (fProfile === 'tous' || g.profile === fProfile) &&
    (fQdisc === 'tous' || g.qdisc === fQdisc) &&
    (fCc === 'tous' || g.cc === fCc))
  const ranked = [...filtered].sort((a, b) => rankMeta.dir === 'down' ? val(a) - val(b) : val(b) - val(a))
  const rankMax = Math.max(...filtered.map(val).filter(Number.isFinite), 1)
  const top = ranked[0]
  const baselineRow = filtered.find(g => g.qdisc === 'pfifo_fast' && (!top || g.profile === top.profile))
    ?? [...filtered].sort((a, b) => rankMeta.dir === 'down' ? val(b) - val(a) : val(a) - val(b))[0]
  const diff = top && baselineRow && Number.isFinite(val(top)) && val(baselineRow) > 0
    ? Math.round(((val(baselineRow) - val(top)) / val(baselineRow)) * 100) : null
  const hwPerProfile = Array.from(new Map(groups.filter(g => g.best).map(g => [g.profile, g.hardware_recommendation ?? '—'])).entries()).map(([p, h]) => `${p}: ${h}`).join(' · ') || '—'
  const maxSmall = Math.max(...groups.map(g => g.small_p95_median), 1)

  const chip = (label: string, active: boolean, onClick: () => void) => (
    <button key={label} className="btn" onClick={onClick} style={{
      padding: '2px 8px', fontSize: 10, fontFamily: 'var(--font-mono)',
      border: '1px solid ' + (active ? '#3a3a40' : 'var(--hairline)'),
      background: active ? 'rgba(90,211,227,0.12)' : 'transparent',
      color: active ? '#7fd6e8' : '#a8aeb7',
    }}>{label}</button>
  )

  return (
    <div className="panel-stack" style={{ position: 'relative' }}>
      {peek && (
        <PeekPopover rect={peek.rect}>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099' }}>{peek.g.profile} · {peek.g.qdisc} · {peek.g.cc}</div>
          <div style={{ display: 'flex', gap: 6, marginTop: 4, alignItems: 'end' }}>
            <span className="mono" style={{ fontSize: 10, color: '#5ad3e3' }}>{peek.g.small_p95_median.toFixed(1)} ms</span>
            <span className="mono" style={{ fontSize: 10, color: '#767b84' }}>n={peek.g.count}</span>
            <div style={{ flex: 1, height: 4, background: 'var(--hairline-faint)', borderRadius: 2, overflow: 'hidden' }}>
              <div style={{ width: `${(peek.g.small_p95_median / maxSmall) * 100}%`, height: '100%', background: peek.g.best ? 'var(--t-ok)' : '#5ad3e3' }} />
            </div>
          </div>
          {peek.g.hardware_recommendation && <div className="mono" style={{ fontSize: 9, color: '#8b9099', marginTop: 4, maxWidth: 220, whiteSpace: 'normal' }}>{peek.g.hardware_recommendation}</div>}
        </PeekPopover>
      )}
      <h1 className="view-title">Résultats — classement complet</h1>

      {/* verdict recalculé sur le critère choisi — toujours mesuré, jamais décoré */}
      {top && baselineRow && (
        <div className="card" data-testid="rank-verdict" style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: 12, alignItems: 'center', padding: 14 }}>
          <div>
            <div className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
              <Explain term="pfifo_fast">pfifo — avant</Explain>
            </div>
            <div className="mono" style={{ fontSize: 16, fontWeight: 700, color: '#c3c9d1', fontVariantNumeric: 'tabular-nums' }}>
              {val(baselineRow) === Number.MAX_SAFE_INTEGER ? '—' : `${val(baselineRow).toFixed(1)} ${rankMeta.unit}`}
            </div>
            <div className="mono" style={{ fontSize: 10, color: '#9aa0a8' }}>{baselineRow.profile} · n={baselineRow.count}</div>
          </div>
          <div className="mono" data-testid="rank-diff" style={{ fontSize: 24, fontWeight: 700, fontVariantNumeric: 'tabular-nums', color: diff != null && diff > 0 ? '#1fa348' : '#c3c9d1', textAlign: 'center' }}>
            {diff != null ? (diff > 0 ? `−${diff} %` : `+${Math.abs(diff)} %`) : '—'}
          </div>
          <div style={{ textAlign: 'right' }}>
            <div className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#7fd6e8' }}>
              <Explain term={rankMeta.term}>1er — {rankMeta.label}</Explain>
            </div>
            <div className="mono" style={{ fontSize: 16, fontWeight: 700, color: '#1fa348', fontVariantNumeric: 'tabular-nums' }}>
              {val(top) === Number.MAX_SAFE_INTEGER ? '—' : `${val(top).toFixed(1)} ${rankMeta.unit}`}
            </div>
            <div className="mono" style={{ fontSize: 10, color: '#9aa0a8' }}>{top.profile} · {top.qdisc}/{top.cc} · n={top.count}</div>
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

      <div className="card" style={{ overflowX: 'auto' }}>
        <table className='data-table' style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-mono)', fontSize: 12 }}>
          <thead>
            <tr style={{ color: '#c3c9d1', textAlign: 'left', borderBottom: '1px solid var(--hairline)' }}>
              <th style={{ padding: '6px 8px' }}>#</th>
              <th style={{ padding: '6px 8px' }}>profil</th><th>qdisc</th><th>cc</th><th>n</th>
              <th style={{ minWidth: 140 }}><Explain term="small_p95">small p95</Explain></th>
              <th><Explain term="rtt_p95">rtt p95</Explain></th>
              <th><Explain term="bulk_goodput">goodput</Explain></th>
              <th><Explain term="deadline">deadline ok</Explain></th>
              <th><Explain term="wasted">gaspillé</Explain></th>
              <th><Explain term="cost_ar_per_h">coût</Explain></th>
              <th>quar.</th>
              <th style={{ padding: '6px 8px' }}>comparer</th>
            </tr>
          </thead>
          <tbody>
            {ranked.map((g, i) => {
              const pct = Math.min(100, (val(g) / rankMax) * 100)
              const barColor = i === 0 ? 'var(--t-ok)' : g.qdisc === 'cake' ? 'var(--t-bbr)' : g.qdisc === 'fq_codel' ? 'var(--t-live)' : '#6b7078'
              const wasted: number | null = g.wasted_median ?? g.wasted_bytes ?? null
              const cost: number | null = g.cost_median ?? g.cost_ar_per_h ?? null
              const deadlineOk: number | null = g.deadline_median ?? g.deadline_ok_pct ?? null
              return (
                <tr key={`${g.profile}/${g.qdisc}/${g.cc}`} style={{ borderBottom: '1px solid var(--hairline-faint)', background: i === 0 ? 'rgba(31,163,72,0.08)' : 'transparent' }} onMouseEnter={e => setPeek({ rect: e.currentTarget.getBoundingClientRect(), g })} onMouseLeave={() => setPeek(null)}>
                  <td style={{ padding: '6px 8px', fontWeight: i === 0 ? 700 : 400, color: i === 0 ? '#1fa348' : '#a8aeb7' }}>{i + 1}</td>
                  <td style={{ padding: '6px 8px' }}>{g.profile}</td>
                  <td>{g.qdisc}</td><td>{g.cc}</td><td>{g.count}</td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <div style={{ flex: 1, height: 6, background: 'var(--hairline-faint)', position: 'relative', minWidth: 80, borderRadius: 2, overflow: 'hidden' }}>
                        <div className="leader-bar" style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: `${pct}%`, background: barColor, boxShadow: i === 0 ? `0 0 6px ${barColor}` : 'none', transformOrigin: 'left center', borderRadius: 2, filter: i === 0 ? `drop-shadow(0 0 4px ${barColor})` : 'none' }} />
                      </div>
                      <span style={{ minWidth: 45, textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>{g.small_p95_median.toFixed(1)}</span>
                    </div>
                  </td>
                  <td>{g.rtt_p95_median.toFixed(1)}</td>
                  <td>{g.goodput_median.toFixed(1)}</td>
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: deadlineOk == null ? '#9aa0a8' : deadlineOk >= 95 ? '#1fa348' : '#f4b400', textAlign: 'right' }}>{deadlineOk == null ? '—' : deadlineOk.toFixed(0) + '%'}</td>
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: wasted == null ? '#9aa0a8' : wasted > 0 ? '#e22718' : '#9aa0a8', textAlign: 'right' }}>{wasted == null ? '—' : wasted >= 1048576 ? (wasted / 1048576).toFixed(1) + ' MiB' : wasted >= 1024 ? (wasted / 1024).toFixed(0) + ' KiB' : String(wasted)}</td>
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: cost == null ? '#9aa0a8' : cost > 0 ? '#f4b400' : '#9aa0a8', textAlign: 'right' }}>{cost == null ? '—' : cost.toFixed(0)}</td>
                  <td>{g.quarantined}</td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    {(() => {
                      const pin = { profile: g.profile, qdisc: g.qdisc, cc: g.cc }
                      const isA = pinA?.profile === g.profile && pinA?.qdisc === g.qdisc && pinA?.cc === g.cc
                      const isB = pinB?.profile === g.profile && pinB?.qdisc === g.qdisc && pinB?.cc === g.cc
                      return (<>
                        <button className="btn" title="épingler comme A" onClick={() => setPinA(pin)} style={{ padding: '2px 6px', fontSize: 10, background: isA ? 'rgba(90,211,227,0.15)' : 'transparent', color: isA ? CRAFT.live : 'var(--text-muted)' }}>A</button>
                        <button className="btn" title="épingler comme B" onClick={() => setPinB(pin)} style={{ padding: '2px 6px', fontSize: 10, marginLeft: 4, background: isB ? 'rgba(31,163,72,0.15)' : 'transparent', color: isB ? CRAFT.ok : 'var(--text-muted)' }}>B</button>
                      </>)
                    })()}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {pinA && pinB && (
        <CompareView a={pinA} b={pinB} onClose={() => { setPinA(null); setPinB(null) }} />
      )}
      <div className="card" style={{ padding: 12 }}>
        <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099', marginBottom: 6 }}>goodput vs small — compromis débit/latence · hash {hash8}</div>
        <div ref={scatterRef} style={{ height: 220 }} />
      </div>
      <div className="form-row" style={{ gap: 8 }}>
        <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
        <a className="btn" href="/api/report/export?format=md" download style={{ border: '1px solid var(--hairline)', padding: '7px 16px' }}>Exporter MD</a>
        <span className="mono muted" style={{ marginLeft: 8 }}>médianes Scan réelles — ★ meilleur small p95 par profil · barres relatives au max · provenance {hash8}</span>
      </div>
      <Provenance source="data/runs/*/aqm_eval.csv" state="live" extra={`${groups.length} groupes · max small ${maxSmall.toFixed(1)} ms · ${hwPerProfile} · hash ${hash8}`} />
    </div>
  )
}
