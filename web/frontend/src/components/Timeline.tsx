import * as d3 from 'd3'
import { useEffect, useRef } from 'react'

export function Timeline({ baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase }: { baselineStart: number, chargeStart: number, chargeEnd: number, recupEnd: number, currentPhase: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current) return
    const w = ref.current.clientWidth || 600, h = 48
    const svg = d3.select(ref.current).html('').append('svg').attr('width', w).attr('height', h)
    const x = d3.scaleTime().domain([new Date(baselineStart), new Date(recupEnd)]).range([0, w])
    svg.append('rect').attr('x', x(new Date(baselineStart))).attr('width', x(new Date(chargeStart)) - x(new Date(baselineStart))).attr('height', h).attr('fill', 'rgba(90,211,227,0.04)')
    svg.append('rect').attr('x', x(new Date(chargeStart))).attr('width', x(new Date(chargeEnd)) - x(new Date(chargeStart))).attr('height', h).attr('fill', 'rgba(244,180,0,0.08)')
    svg.append('rect').attr('x', x(new Date(chargeEnd))).attr('width', x(new Date(recupEnd)) - x(new Date(chargeEnd))).attr('height', h).attr('fill', 'rgba(31,163,72,0.06)')
    // animate via Task 4: pulse current phase rect with anime stagger
  }, [baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase])
  return <div ref={ref} style={{height:48, border: '1px solid #26262a'}} data-testid="timeline" aria-label={`timeline ${currentPhase}`} />
}
