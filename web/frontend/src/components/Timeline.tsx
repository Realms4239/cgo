import * as d3 from 'd3'
import { useEffect, useRef } from 'react'

// Timeline 48 — bandes de phase, bornes FIGÉES par event.
// recupEnd = chargeEnd + RecupSec (nominal) : la bande ne grandit pas à
// chaque frame, le SVG n'est jamais reconstruit en cours de phase.
export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    if (!currentPhase || currentPhase === 'idle' || currentPhase === 'surveil') {
      d3.select(ref.current).html('')
      return
    }
    const w = ref.current.clientWidth || 600, h = 48
    const svg = d3.select(ref.current).html('').append('svg').attr('width', w).attr('height', h)
    const x = d3.scaleTime().domain([new Date(baselineStart), new Date(recupEnd)]).range([0, w])
    const rects: Record<string, { x: number, w: number }> = {
      baseline: { x: x(new Date(baselineStart)), w: x(new Date(chargeStart)) - x(new Date(baselineStart)) },
      charge: { x: x(new Date(chargeStart)), w: x(new Date(chargeEnd)) - x(new Date(chargeStart)) },
      recup: { x: x(new Date(chargeEnd)), w: x(new Date(recupEnd)) - x(new Date(chargeEnd)) },
    }
    const fill: Record<string,string> = { baseline:'rgba(90,211,227,0.04)', charge:'rgba(244,180,0,0.08)', recup:'rgba(31,163,72,0.06)' }
    for (const k of ['baseline','charge','recup'] as const) {
      const isCurrent = currentPhase === k
      svg.append('rect')
        .attr('x', rects[k].x).attr('width', Math.max(0, rects[k].w)).attr('height', h)
        .attr('fill', fill[k])
        .attr('stroke', isCurrent ? 'rgba(255,255,255,0.35)' : 'none')
        .attr('stroke-width', isCurrent ? 1 : 0)
        .style('filter', isCurrent ? 'drop-shadow(0 0 6px rgba(255,255,255,0.25))' : 'none')
    }
    // curseur de progression — une flèche par frame, pas de reconstruction
    const cursor = svg.append('line')
      .attr('y1', 0).attr('y2', h)
      .attr('stroke', 'rgba(255,255,255,0.5)').attr('stroke-width', 1)
    const tick = window.setInterval(() => {
      if (!ref.current) { window.clearInterval(tick); return }
      cursor.attr('x1', x(new Date())).attr('x2', x(new Date()))
    }, 500)
    return () => {
      window.clearInterval(tick)
      try { (d3 as any).select(ref.current).selectAll('*').remove() } catch {}
    }
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{height:48, border: '1px solid #26262a'}} data-testid="timeline" aria-label={`timeline ${currentPhase}`} />
}
