export function computeJFI(values: number[]): number {
  if (values.length === 0) return 0
  const sum = values.reduce((a, b) => a + b, 0)
  const sumSq = values.reduce((a, b) => a + b * b, 0)
  return sumSq === 0 ? 0 : (sum * sum) / (values.length * sumSq)
}
