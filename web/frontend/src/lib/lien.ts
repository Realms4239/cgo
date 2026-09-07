export type LienMetrics = {
  small: number | null
  rtt: number | null
  goodput: number | null
  deadline: number | null
  cost: number | null
}

export function normDown(values: (number | null)[], v: number | null): number | null {
  if (v == null) return null
  const xs = values.filter((x): x is number => x != null)
  if (!xs.length) return null
  const mn = Math.min(...xs), mx = Math.max(...xs)
  if (!(mx > mn)) return 1
  return 1 - (v - mn) / (mx - mn)
}

export function normUp(values: (number | null)[], v: number | null): number | null {
  if (v == null) return null
  const xs = values.filter((x): x is number => x != null)
  if (!xs.length) return null
  const mn = Math.min(...xs), mx = Math.max(...xs)
  if (!(mx > mn)) return 1
  return (v - mn) / (mx - mn)
}

export function components(
  all: LienMetrics[],
  self: LienMetrics,
): { small: number | null; rtt: number | null; goodput: number | null; deadline: number | null; cost: number | null } {
  return {
    small: normDown(all.map(a => a.small), self.small),
    rtt: normDown(all.map(a => a.rtt), self.rtt),
    goodput: normUp(all.map(a => a.goodput), self.goodput),
    deadline: normUp(all.map(a => a.deadline), self.deadline),
    cost: normDown(all.map(a => a.cost), self.cost),
  }
}

export function lienScore(all: LienMetrics[], self: LienMetrics): number | null {
  const c = components(all, self)
  const parts = [c.small, c.rtt, c.goodput, c.deadline, c.cost].filter((x): x is number => x != null)
  if (!parts.length) return null
  return (parts.reduce((a, b) => a + b, 0) / parts.length) * 100
}
