import { useEffect, useRef } from 'react'
import { paletteOnce, alpha } from '../lib/palette'

// Timeline 48 — bandes de phase, bornes FIGÉES par event, zéro dépendance :
// le scale est un map linéaire, le SVG un template. d3 complet (90 KB gz)
// servait trois appels — retiré.
const PHASES = [
  { key: 'baseline', label: 'référence', color: 'live', wash: 0.04, text: 'var(--text-muted)' },
  { key: 'charge', label: 'charge', color: 'threshold', wash: 0.14, text: 'var(--t-threshold)' },
  { key: 'recup', label: 'récupération', color: 'ok', wash: 0.08, text: 'var(--t-ok)' },
] as const

const secs = (a: number, b: number) => Math.max(0, Math.round((b - a) / 1000))

export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  // vide honnête : hors phase, pas de boîte fantôme bordée de 48 px
  const idle = !currentPhase || currentPhase === 'idle' || currentPhase === 'surveil'
  useEffect(() => {
    const el = ref.current
    if (!el || idle) return
    const p = paletteOnce()
    const w = el.clientWidth || 600, h = 48
    const x = (t: number) => ((t - baselineStart) / Math.max(1, recupEnd - baselineStart)) * w
    const bands: [string, number, number][] = [
      ['baseline', baselineStart, chargeStart],
      ['charge', chargeStart, chargeEnd],
      ['recup', chargeEnd, recupEnd],
    ]
    let rects = ''
    for (const [k, a, b] of bands) {
      const rx = Math.max(0, x(a)), rw = Math.max(0, x(b) - x(a))
      const cur = currentPhase === k
      const ph = PHASES.find(x => x.key === k)!
      rects += `<rect x="${rx.toFixed(1)}" width="${rw.toFixed(1)}" height="${h}" fill="${alpha(p[ph.color], ph.wash)}"${cur ? ` stroke="${alpha(p.textPrimary, 0.35)}" stroke-width="1"` : ''}/>`
    }
    el.innerHTML = `<svg width="${w}" height="${h}" style="display:block">${rects}<line class="tl-cursor" y1="0" y2="${h}" stroke="${alpha(p.textPrimary, 0.5)}" stroke-width="1"/></svg>`
    // curseur de progression — une position par 500 ms, jamais de rebuild
    const cursor = el.querySelector<SVGLineElement>('.tl-cursor')
    const tick = window.setInterval(() => {
      if (!cursor) { window.clearInterval(tick); return }
      cursor.setAttribute('x1', String(x(Date.now())))
      cursor.setAttribute('x2', String(x(Date.now())))
    }, 500)
    return () => {
      window.clearInterval(tick)
      if (el.isConnected) el.innerHTML = ''
    }
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase, idle])
  if (idle) return null
  return (
    <div data-testid="timeline" aria-label={`timeline ${currentPhase}`} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div ref={ref} style={{ height: 48, border: '1px solid var(--hairline)' }} />
      <div className="mono" style={{ display: 'flex', gap: 12, flexWrap: 'wrap', fontSize: 'var(--fs-micro)' }}>
        {PHASES.map(p => {
          const b = p.key === 'baseline' ? [baselineStart, chargeStart] : p.key === 'charge' ? [chargeStart, chargeEnd] : [chargeEnd, recupEnd]
          const cur = currentPhase === p.key
          return (
            <span key={p.key} style={{ color: cur ? 'var(--text-primary)' : p.text, fontWeight: cur ? 700 : 400 }}>
              {cur ? '● ' : ''}{p.label} · {secs(b[0], b[1])} s
            </span>
          )
        })}
      </div>
    </div>
  )
}
