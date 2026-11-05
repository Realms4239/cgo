// Le canvas de l'étape Observer — une seule courbe, rompue à chaque seconde
// muette. Aucun ECharts, aucun nœud par échantillon, aucune animation.
import { useEffect, useRef } from 'react'
import type { AuditTimelinePoint } from '../lib/types'
import { alpha, paletteOnce } from '../lib/palette'
import { buildTrace, TRACE_HEIGHT } from '../lib/trace'

const labelMs = (v: number | null | undefined) => {
  if (v == null || !Number.isFinite(v) || v <= 0) return 'non mesuré'
  return `${v >= 100 ? v.toFixed(0) : v.toFixed(1).replace(/\.0$/, '')} ms`
}

function band(phase: string, p: ReturnType<typeof paletteOnce>): string {
  switch (phase) {
    case 'charge':
    case 'charge-down':
      return alpha(p.threshold, 0.08)
    case 'drain':
      return alpha(p.neutral, 0.12)
    default:
      return alpha(p.hairline, 0.45)
  }
}

export function TraceCanvas({
  points = [],
  totalSec,
  deadlineMs,
  responseMs,
  loadResponseMs,
}: {
  points?: AuditTimelinePoint[]
  totalSec?: number
  deadlineMs: number
  responseMs?: number | null
  loadResponseMs?: number | null
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const sorted = [...points].sort((a, b) => a.t - b.t)
  const maxT = sorted.reduce((m, p) => Math.max(m, p.t), 0)
  const span = Number.isFinite(totalSec) && totalSec != null && totalSec > 0 ? totalSec : maxT + 1
  const accessibleName = `Trace des sondes — temps de réponse ${labelMs(responseMs)} ; réponse sous charge ${labelMs(loadResponseMs)}, objectif ${deadlineMs} ms`

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const dpr = Math.min(2, window.devicePixelRatio || 1)
    const width = canvas.clientWidth
    if (width <= 0) return
    canvas.width = Math.max(1, Math.floor(width * dpr))
    canvas.height = TRACE_HEIGHT * dpr
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, width, TRACE_HEIGHT)

    const palette = paletteOnce()
    const x = (t: number) => (Math.max(0, t) / span) * width
    if (sorted.length > 0) {
      let start = sorted[0].t
      let phase = sorted[0].phase
      for (let i = 1; i < sorted.length; i++) {
        const p = sorted[i]
        if (p.phase === phase) continue
        ctx.fillStyle = band(phase, palette)
        ctx.fillRect(x(start), 0, Math.max(1, x(p.t) - x(start)), TRACE_HEIGHT)
        start = p.t
        phase = p.phase
      }
      ctx.fillStyle = band(phase, palette)
      ctx.fillRect(x(start), 0, Math.max(1, width - x(start)), TRACE_HEIGHT)
    }

    const g = buildTrace(sorted, deadlineMs)
    ctx.setLineDash([4, 3])
    ctx.strokeStyle = palette.threshold
    ctx.globalAlpha = 0.8
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(0, g.objectiveY)
    ctx.lineTo(width, g.objectiveY)
    ctx.stroke()
    ctx.globalAlpha = 1
    ctx.setLineDash([])
    ctx.fillStyle = palette.threshold
    ctx.font = '11px "JetBrains Mono", monospace'
    ctx.textAlign = 'right'
    ctx.fillText(`objectif ${deadlineMs} ms`, width - 4, g.objectiveY <= 16 ? g.objectiveY + 13 : g.objectiveY - 4)

    ctx.strokeStyle = palette.live
    ctx.lineWidth = 2
    for (const segment of g.segments) {
      if (segment.points.length < 2) continue
      ctx.beginPath()
      ctx.moveTo(x(segment.points[0].t), segment.points[0].y)
      for (const point of segment.points.slice(1)) ctx.lineTo(x(point.t), point.y)
      ctx.stroke()
    }
  }, [points, totalSec, deadlineMs]) // points, totalSec, deadlineMs : la courbe se redessine à chaque sonde

  return (
    <div className="live-trace" data-testid="trace-canvas">
      <canvas ref={canvasRef} role="img" aria-label={accessibleName} className="live-trace-canvas" />
      <div className="live-trace-legend mono">
        <span className="live-trace-gap" aria-hidden="true" />
        sonde sans réponse
      </div>
    </div>
  )
}
