// La trace d'un audit — la réponse des petites sondes sous charge. Une
// seconde muette rompt le tracé ; elle n'est jamais dessinée comme 0 ms.
import type { AuditTimelinePoint } from './types'

export const TRACE_HEIGHT = 120
const TOP = 10
const BOTTOM = 16
const PLOT_HEIGHT = TRACE_HEIGHT - TOP - BOTTOM

export interface TraceSegment {
  points: { t: number; y: number }[]
}

export interface TraceGeometry {
  segments: TraceSegment[]
  yMax: number
  objectiveY: number
}

const answered = (p: AuditTimelinePoint): number | null => {
  const v = p.small_p95
  return Number.isFinite(v) && v != null && v > 0 ? v : null
}

export function buildTrace(points: AuditTimelinePoint[], deadlineMs: number): TraceGeometry {
  const sorted = [...points].sort((a, b) => a.t - b.t)
  const goal = Number.isFinite(deadlineMs) && deadlineMs > 0 ? deadlineMs : 1
  const observed = sorted.map(answered).filter((v): v is number => v != null)
  const yMax = Math.max(goal, ...observed, 1)
  const y = (v: number) => TOP + (1 - v / yMax) * PLOT_HEIGHT

  const segments: TraceSegment[] = []
  let current: TraceSegment | null = null
  for (const point of sorted) {
    const value = answered(point)
    if (value == null) {
      if (current) segments.push(current)
      current = null
      continue
    }
    current ??= { points: [] }
    current.points.push({ t: point.t, y: y(value) })
  }
  if (current) segments.push(current)

  return { segments, yMax, objectiveY: y(goal) }
}
