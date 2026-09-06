// fmtIQR — médiane [IQR bas–haut] à 1 décimale ; sans IQR, médiane seule.
// L'incertitude s'affiche, ne se cache pas (DeepSWE affiche ± ; ici [–]).
export function fmtIQR(med: number, iqr?: [number, number] | null): string {
  const m = med.toFixed(1)
  if (!iqr || iqr.length !== 2) return m
  return `${m} [${iqr[0].toFixed(1)}–${iqr[1].toFixed(1)}]`
}

// asArray — garde anti-écran-d'erreur : un payload {error} ou un objet là où
// un tableau est attendu ne doit jamais atteindre .filter/.map. Vu en prod :
// /api/results?run= vide répond {available:false} SANS groups, le repli
// `j.groups??j??[]` laissait passer l'objet → O.filter is not a function.
export function asArray<T>(v: unknown): T[] {
  return Array.isArray(v) ? (v as T[]) : []
}
