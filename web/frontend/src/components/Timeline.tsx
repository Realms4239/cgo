import * as d3 from 'd3'
import { useEffect, useRef } from 'react'
import { live } from '../lib/live'

// ponytail: full d3 — switch to d3-scale/d3-selection if bundle exceeds 650KB
// ponytail: brushX global live.max — per-phase brush if selection throughput matters
export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    if (currentPhase === 'idle') {
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
        .attr('x', rects[k].x).attr('width', rects[k].w).attr('height', h)
        .attr('fill', fill[k])
        .attr('stroke', isCurrent ? 'rgba(255,255,255,0.35)' : 'none')
        .attr('stroke-width', isCurrent ? 1 : 0)
        .style('filter', isCurrent ? 'drop-shadow(0 0 6px rgba(255,255,255,0.25))' : 'none')
    }
    // brushX — extent [[0,0],[w,48]] sets live.max from selection width
    // ponytail: live.max is mutable ring, don't corrupt Zustand — use live.max directly
    // snap-effect handle 16px — ponytail: 16px grab area keeps brush usable on touch
    const brush = (d3 as any).brushX().extent([[0, 0], [w, 48]]).handleSize(16).on('end', (e: any) => {
      if (e.selection) {
        const a = (x.invert as any)(e.selection[0]).getTime()
        const b = (x.invert as any)(e.selection[1]).getTime()
        const span = Math.abs(b - a)
        // ponytail: live.max mutable — single source for ring window, 60..1800 clamp (rings 1800, Q53 B)
        live.max = Math.max(60, Math.min(1800, Math.round(span / 100)))
      } else {
        // reset to full window when brush cleared — prevents permanent shrink
        live.max = 1800
      }
    })
    const gBrush = svg.append('g').attr('class', 'brush').call(brush as any)
    return () => {
      try { (d3 as any).select(gBrush.node()).on('.brush', null) } catch {}
      d3.select(ref.current!).selectAll('*').remove()
    }
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{height:48, border: '1px solid #26262a'}} data-testid="timeline" aria-label={`timeline ${currentPhase}`} />
}
