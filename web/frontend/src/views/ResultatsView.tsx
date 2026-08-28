import { useEffect, useRef, useState } from 'react'
import { EmptyState } from '../components/ui/EmptyState'
import { Provenance } from '../components/ui/Provenance'
import { animateBar } from '../lib/anime'
import { computeJFI } from '../lib/jfi'
import { PeekPopover } from '../components/PeekPopover'
import { echarts } from '../lib/echarts'
import { baseOption, scatterSeries } from '../lib/chartGrammar'

type Group = {
  profile: string; qdisc: string; cc: string
  count: number; quarantined: number
  rtt_p95_median: number; small_p95_median: number
  goodput_median: number
  deadline_median?: number; deadline_ok_pct?: number
  wasted_median?: number; cost_median?: number; wasted_bytes?: number; cost_ar_per_h?: number
  best?: boolean; hardware_recommendation?: string
}

export default function ResultatsView() {
  const [groups, setGroups] = useState<Group[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [showCosts, setShowCosts] = useState(true)
  const [peek, setPeek] = useState<{ rect: DOMRect; g: Group } | null>(null)
  const [hash8, setHash8] = useState<string>('────────')
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

  const maxSmall = Math.max(...groups.map(g => g.small_p95_median), 1)
  const best = groups.find(g => g.best) ?? groups[0]
  const baseline = groups.find(g => g.qdisc === 'pfifo_fast' && g.profile === best.profile) ?? [...groups].sort((a, b) => b.small_p95_median - a.small_p95_median)[0]
  const diff = baseline && best && baseline.small_p95_median > 0 ? Math.round(((baseline.small_p95_median - best.small_p95_median) / baseline.small_p95_median) * 100) : null
  // hardware_recommendation per profile best (from Scan) — not per row
  const hwPerProfile = Array.from(new Map(groups.filter(g => g.best).map(g => [g.profile, g.hardware_recommendation ?? '—'])).entries()).map(([p, h]) => `${p}: ${h}`).join(' · ') || best.hardware_recommendation || '—'

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
      <h1 className="view-title">Résultats — comparaison AQM/BBR</h1>

      {/* ab-bento 3-col Avant/Après diff badge — hardware_recommendation provenance per profile best */}
      <div className="ab-bento card" data-testid="ab-bento" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: 12, alignItems: 'center', border: '1px solid #26262a', background: 'var(--surface-card)', padding: 16 }}>
        <div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>Avant — baseline</div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 14, fontWeight: 700, color: '#767b84', fontVariantNumeric: 'tabular-nums' }}>{baseline.qdisc} {baseline.small_p95_median.toFixed(1)} ms</div>
          <svg width="100%" height={4} style={{ display: 'block', marginTop: 6 }} aria-hidden><line x1={0} y1={2} x2="100%" y2={2} stroke="#767b84" strokeWidth={2} strokeDasharray="6 4" /></svg>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#767b84', marginTop: 4 }}>pfifo_fast — gris pointillé · Scan median réel</div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#8b9099', marginTop: 2 }}>profil {baseline.profile} · n={baseline.count}</div>
        </div>
        <div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#5ad3e3' }}>Après — CAKE</div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 14, fontWeight: 700, color: '#1fa348', fontVariantNumeric: 'tabular-nums' }}>{best.qdisc} {best.small_p95_median.toFixed(1)} ms</div>
          <svg width="100%" height={4} style={{ display: 'block', marginTop: 6 }} aria-hidden><line x1={0} y1={2} x2="100%" y2={2} stroke="#5ad3e3" strokeWidth={2} /><line x1={0} y1={2} x2="100%" y2={2} stroke="#1fa348" strokeWidth={1} strokeDasharray="12 6" opacity={0.7} /></svg>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#5ad3e3', marginTop: 4 }}>CAKE — cyan solide · best median · Scan median réel</div>
          <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 9, color: '#8b9099', marginTop: 2 }}>profil {best.profile} · n={best.count} · {best.best ? '★' : ''}</div>
        </div>
        {diff != null && <div className="diff-badge mono" style={{ background: diff > 0 ? 'rgba(31,163,72,0.12)' : 'rgba(226,39,24,0.12)', border: '1px solid ' + (diff > 0 ? '#1fa348' : '#e22718'), color: diff > 0 ? '#1fa348' : '#e22718', padding: '8px 12px', fontSize: 18, fontWeight: 700, fontVariantNumeric: 'tabular-nums', textAlign: 'center', minWidth: 80 }}>{diff > 0 ? `-${diff}%` : `${diff}%`}<div style={{ fontSize: 9, fontWeight: 400, color: '#8b9099', marginTop: 2 }}>-(baseline-best)/baseline</div></div>}
      </div>
      <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099', background: 'var(--surface-card)', border: '1px solid #26262a', padding: '8px 12px', display: 'flex', justifyContent: 'space-between', gap: 8, flexWrap: 'wrap' }}>
        <span>provenance hash {hash8} · {groups.length} groupes · max small {maxSmall.toFixed(1)} ms</span>
        <span style={{ color: '#767b84', maxWidth: 480, textAlign: 'right', whiteSpace: 'normal' }}>{hwPerProfile}</span>
      </div>

      <div className="form-row" style={{ justifyContent: 'flex-end', gap: 8, marginBottom: 8 }}>
        <span className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#8b9099' }}>affichage</span>
        <button className="btn" onClick={() => setShowCosts(v => !v)} style={{ border: '1px solid #26262a', padding: '4px 10px', font: '700 10px JetBrains Mono', letterSpacing: '0.08em', textTransform: 'uppercase', background: showCosts ? 'rgba(90,211,227,0.08)' : 'transparent', color: showCosts ? '#5ad3e3' : '#8b9099' }}>
          {showCosts ? 'masquer coûts' : 'afficher coûts'}
        </button>
        <span className="mono muted" style={{ fontFamily: 'JetBrains Mono', fontSize: 10 }}>deadline_ok · wasted · cost · hash {hash8}</span>
      </div>

      <div className="card" style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-mono)', fontSize: 12 }}>
          <thead>
            <tr style={{ color: 'var(--text-muted)', textAlign: 'left', borderBottom: '1px solid var(--hairline)' }}>
              <th style={{ padding: '6px 8px' }}>profil</th><th>qdisc</th><th>cc</th><th>n</th><th style={{ minWidth: 140 }}>small p95</th><th>rtt p95</th><th>goodput</th><th title="Jain's fairness 0–1" style={{ width: 52, fontSize: 11, fontFamily: 'JetBrains Mono', color: '#9aa3ad' }}>JFI</th>
              {showCosts && <><th style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#8b9099' }}>deadline_ok</th><th style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#8b9099' }}>wasted</th><th style={{ fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f4b400' }}>cost</th></>}
              <th>quar.</th><th>best</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((g, i) => {
              const pct = (g.small_p95_median / maxSmall) * 100
              const barColor = g.best ? 'var(--t-ok)' : g.qdisc === 'cake' ? 'var(--t-bbr)' : g.qdisc === 'fq_codel' ? 'var(--t-live)' : 'var(--text-faint)'
              const jfiVals = Array.from({ length: g.count }, () => g.small_p95_median)
              const jfiDegenerate = g.count < 2 || jfiVals.every(v => v === jfiVals[0])
              const jfi: number | null = jfiDegenerate ? null : computeJFI(jfiVals)
              const wasted: number | null = g.wasted_median ?? g.wasted_bytes ?? null
              const cost: number | null = g.cost_median ?? g.cost_ar_per_h ?? null
              const deadlineOk: number | null = g.deadline_median ?? g.deadline_ok_pct ?? null
              return (
                <tr key={i} style={{ borderBottom: '1px solid var(--hairline-faint)', background: g.best ? 'rgba(31,163,72,0.08)' : 'transparent' }} onMouseEnter={e => setPeek({ rect: e.currentTarget.getBoundingClientRect(), g })} onMouseLeave={() => setPeek(null)}>
                  <td style={{ padding: '6px 8px', fontWeight: g.best ? 700 : 400 }}>{g.profile}</td>
                  <td>{g.qdisc}</td><td>{g.cc}</td><td>{g.count}</td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <div style={{ flex: 1, height: 6, background: 'var(--hairline-faint)', position: 'relative', minWidth: 80, borderRadius: 2, overflow: 'hidden' }}>
                        <div className="leader-bar" style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: `${pct}%`, background: barColor, boxShadow: g.best ? `0 0 6px ${barColor}` : 'none', transformOrigin: 'left center', borderRadius: 2, filter: g.best ? `drop-shadow(0 0 4px ${barColor})` : 'none' }} />
                      </div>
                      <span style={{ minWidth: 45, textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>{g.small_p95_median.toFixed(1)}</span>
                    </div>
                  </td>
                  <td>{g.rtt_p95_median.toFixed(1)}</td>
                  <td>{g.goodput_median.toFixed(1)}</td>
                  <td title={jfi === null ? 'JFI requiert détail par répétition (detail=1)' : undefined} style={{ width: 52, fontFamily: 'JetBrains Mono', fontSize: 11, fontVariantNumeric: 'tabular-nums', color: jfi === null ? '#767b84' : jfi > 0.95 ? '#1fa348' : '#9aa3ad', textAlign: 'right' }}>{jfi === null ? '—' : jfi.toFixed(2)}</td>
                  {showCosts && <>
                    <td style={{ fontFamily: 'JetBrains Mono', fontSize: 11, fontVariantNumeric: 'tabular-nums', color: deadlineOk == null ? '#767b84' : deadlineOk >= 95 ? '#1fa348' : '#f4b400', textAlign: 'right' }}>{deadlineOk == null ? '—' : deadlineOk.toFixed(0) + '%'}</td>
                    <td style={{ fontFamily: 'JetBrains Mono', fontSize: 11, fontVariantNumeric: 'tabular-nums', color: wasted == null ? '#767b84' : wasted > 0 ? '#e22718' : '#767b84', textAlign: 'right' }}>{wasted == null ? '—' : wasted ? (wasted > 1024 * 1024 ? (wasted / 1024 / 1024).toFixed(1) + 'M' : String(wasted)) : '0'}</td>
                    <td style={{ fontFamily: 'JetBrains Mono', fontSize: 11, fontVariantNumeric: 'tabular-nums', color: cost == null ? '#767b84' : cost > 0 ? '#f4b400' : '#767b84', textAlign: 'right' }}>{cost == null ? '—' : cost ? cost.toFixed(0) : '0'}</td>
                  </>}
                  <td>{g.quarantined}</td>
                  <td>{g.best ? '★' : ''}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      <div className="card" style={{ padding: 12 }}>
        <div className="mono" style={{ fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099', marginBottom: 6 }}>goodput vs small — brush rect pour comparer · hash {hash8}</div>
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
