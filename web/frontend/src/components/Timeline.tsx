import { useEffect, useRef } from 'react'

// Timeline 48 — bandes de phase, bornes FIGÉES par event, zéro dépendance :
// le scale est un map linéaire, le SVG un template. d3 complet (90 KB gz)
// servait trois appels — retiré.
const PHASES = [
  { key: 'baseline', label: 'référence', fill: 'rgba(90,211,227,0.04)', text: '#8b9099' },
  { key: 'charge', label: 'charge', fill: 'rgba(244,180,0,0.14)', text: '#f4b400' },
  { key: 'recup', label: 'récupération', fill: 'rgba(31,163,72,0.08)', text: '#1fa348' },
] as const

const secs = (a: number, b: number) => Math.max(0, Math.round((b - a) / 1000))

export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  // vide honnête : hors phase, pas de boîte fantôme bordée de 48 px
  const idle = !currentPhase || currentPhase === 'idle' || currentPhase === 'surveil'
  useEffect(() => {
    const el = ref.current
    if (!el || idle) return
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
      rects += `<rect x="${rx.toFixed(1)}" width="${rw.toFixed(1)}" height="${h}" fill="${PHASES.find(p => p.key === k)!.fill}"${cur ? ' stroke="rgba(255,255,255,0.35)" stroke-width="1"' : ''}/>`
    }
    el.innerHTML = `<svg width="${w}" height="${h}" style="display:block">${rects}<line class="tl-cursor" y1="0" y2="${h}" stroke="rgba(255,255,255,0.5)" stroke-width="1"/></svg>`
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
      <div ref={ref} style={{ height: 48, border: '1px solid #26262a' }} />
      <div className="mono" style={{ display: 'flex', gap: 12, flexWrap: 'wrap', fontSize: 10 }}>
        {PHASES.map(p => {
          const b = p.key === 'baseline' ? [baselineStart, chargeStart] : p.key === 'charge' ? [chargeStart, chargeEnd] : [chargeEnd, recupEnd]
          const cur = currentPhase === p.key
          return (
            <span key={p.key} style={{ color: cur ? '#f2f2f4' : p.text, fontWeight: cur ? 700 : 400 }}>
              {cur ? '● ' : ''}{p.label} · {secs(b[0], b[1])} s
            </span>
          )
        })}
      </div>
    </div>
  )
}
