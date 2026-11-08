export type HealthInput = {
  linkType: string
  idleP50Ms?: number
  loadedP95Ms?: number
  targetMs?: number | null
}

export type HealthVerdict =
  | { kind: 'absent' }
  | { kind: 'unconfigured'; bloatMs: number }
  | { kind: 'ok'; bloatMs: number }
  | { kind: 'attention'; bloatMs: number }
  | { kind: 'critical'; bloatMs: number }

type LinkClass = 'fiber' | 'cellular' | 'geo' | 'leo' | 'other'

const RED_BLOAT_MS: Record<LinkClass, number | null> = {
  fiber: 30,
  cellular: 100,
  geo: 150,
  leo: null,
  other: null,
}

function classOf(linkType: string): LinkClass {
  const v = linkType.toLowerCase()
  if (v === 'fiber') return 'fiber'
  if (v === '4g' || v === '5g') return 'cellular'
  if (v === 'vsat') return 'geo'
  if (v === 'leo' || v === 'p4') return 'leo'
  return 'other'
}

export function evaluateLoadedHealth(input: HealthInput): HealthVerdict {
  if (input.idleP50Ms == null || input.loadedP95Ms == null) return { kind: 'absent' }
  const bloatMs = Math.max(0, input.loadedP95Ms - input.idleP50Ms)
  const linkClass = classOf(input.linkType)
  const red = RED_BLOAT_MS[linkClass]
  if (red == null) return { kind: 'unconfigured', bloatMs }
  if (bloatMs > red) return { kind: 'critical', bloatMs }
  if (bloatMs >= red * 0.8) return { kind: 'attention', bloatMs }
  return { kind: 'ok', bloatMs }
}
