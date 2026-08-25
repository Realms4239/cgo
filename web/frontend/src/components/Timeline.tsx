import * as d3 from 'd3'
import { useEffect, useRef } from 'react'

// ponytail: full d3 — switch to d3-scale/d3-selection if bundle exceeds 450KB
export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
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
      const isCurrent = currentPhase === k || (k==='recup' && currentPhase==='recup') || (k==='charge' && currentPhase==='charge') || (k==='baseline' && currentPhase==='baseline')
      svg.append('rect')
        .attr('x', rects[k].x).attr('width', rects[k].w).attr('height', h)
        .attr('fill', fill[k])
        .attr('stroke', isCurrent ? 'rgba(255,255,255,0.35)' : 'none')
        .attr('stroke-width', isCurrent ? 1 : 0)
        .style('filter', isCurrent ? 'drop-shadow(0 0 6px rgba(255,255,255,0.25))' : 'none')
    }
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{height:48, border: '1px solid #26262a'}} data-testid="timeline" aria-label={`timeline ${currentPhase}`} />
}
