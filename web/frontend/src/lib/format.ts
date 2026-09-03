// fmtIQR — médiane [IQR bas–haut] à 1 décimale ; sans IQR, médiane seule.
// L'incertitude s'affiche, ne se cache pas (DeepSWE affiche ± ; ici [–]).
export function fmtIQR(med: number, iqr?: [number, number] | null): string {
  const m = med.toFixed(1)
  if (!iqr || iqr.length !== 2) return m
  return `${m} [${iqr[0].toFixed(1)}–${iqr[1].toFixed(1)}]`
}
