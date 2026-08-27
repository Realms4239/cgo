export function computeQDI(rttP95: number, rttP50: number): number {
  return Math.max(0, rttP95 - rttP50)
}
