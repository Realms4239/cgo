import { useEffect, useRef } from 'react'

// Timeline 48 — bandes de phase, bornes FIGÉES par event, zéro dépendance :
// le scale est un map linéaire, le SVG un template. d3 complet (90 KB gz)
// servait trois appels — retiré.
const PHASES = [
  { key: 'baseline', fill: 'rgba(90,211,227,0.04)' },
  { key: 'charge', fill: 'rgba(244,180,0,0.08)' },
  { key: 'recup', fill: 'rgba(31,163,72,0.06)' },
] as const

export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    if (!currentPhase || currentPhase === 'idle' || currentPhase === 'surveil') {
      el.innerHTML = ''
      return
    }
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
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{ height: 48, border: '1px solid #26262a' }} data-testid="timeline" aria-label={`timeline ${currentPhase}`} />
}
