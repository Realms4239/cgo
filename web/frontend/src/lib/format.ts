// fmtIQR — médiane [IQR bas–haut] à 1 décimale ; sans IQR, médiane seule.
// L'incertitude s'affiche, ne se cache pas (DeepSWE affiche ± ; ici [–]).
export function fmtIQR(med: number, iqr?: [number, number] | null): string {
  const m = med.toFixed(1)
  if (!iqr || iqr.length !== 2) return m
  return `${m} [${iqr[0].toFixed(1)}–${iqr[1].toFixed(1)}]`
}

// improve — écart signé B vs A en % dans le SENS de la métrique (>0 =
// B meilleur). down (latences, coût) : baisser = mieux ; up (goodput,
// échéances) : monter = mieux. Source unique du signe/couleur des écarts
// (table CompareView, teintes leaderboard) — le bug "goodput -3% vert
// alors que pire" venait d'un pct brut sans direction.
export function improve(a: number, b: number, dir: 'down' | 'up'): number | null {
  if (!Number.isFinite(a) || !Number.isFinite(b) || a === 0) return null
  const pct = dir === 'down' ? ((a - b) / Math.abs(a)) * 100 : ((b - a) / Math.abs(a)) * 100
  return Math.round(pct)
}
// groupDigits — espaces fines insécables (12 400, pas 12400)
export function groupDigits(n: number, decimals = 0): string {
  if (!Number.isFinite(n)) return '—'
  const fixed = n.toFixed(decimals)
  const [int, dec] = fixed.split('.')
  const grouped = int.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
  return dec != null ? `${grouped},${dec}` : grouped
}

// formatBytesFR — octets français (o, Kio, Mio), jamais `bytes`/`MiB`
export function formatBytesFR(bytes: number | null | undefined): { num: string; unit: string } {
  if (bytes == null || !Number.isFinite(bytes)) return { num: '—', unit: '' }
  if (bytes < 1024) return { num: groupDigits(bytes), unit: 'o' }
  if (bytes < 1024 * 1024) return { num: groupDigits(bytes / 1024, 1), unit: 'Kio' }
  return { num: groupDigits(bytes / 1024 / 1024, 1), unit: 'Mio' }
}

// formatCostAR — Ariary/heure à décimales groupées
export function formatCostAR(ar: number | null | undefined): { num: string; unit: string } {
  if (ar == null || !Number.isFinite(ar)) return { num: '—', unit: '' }
  return { num: groupDigits(ar, 2), unit: 'Ar/h' }
}

// verdictWord — mot de verdict Grafana (le mot porte le jugement, jamais la
// couleur seule) ; niveaux réels des réglages, rien d'inventé
export function verdictWord(level: 'ok' | 'warn' | 'crit'): string {
  return level === 'ok' ? 'bon' : level === 'warn' ? 'fragile' : 'mauvais'
}

// truncateWords — coupe à la frontière de mot, jamais en plein mot
export function truncateWords(s: string, max: number): string {
  if (s.length <= max) return s
  const cut = s.slice(0, max)
  const sp = cut.lastIndexOf(' ')
  return (sp > max * 0.5 ? cut.slice(0, sp) : cut).trimEnd() + '…'
}
// fmtCount — compte entier (pertes, événements) : jamais de décimale
// (`pertes 0.0` venait d'un toFixed(1) générique sur des entiers).
export function fmtCount(v: number | null | undefined): string {
  if (v == null || !Number.isFinite(v)) return '—'
  return groupDigits(Math.round(v))
}

// verdictHeadline — le verdict nomme le GAGNANT, jamais le perdant seul
// (fini le `0/5 … en faveur de B` qui titrait le perdant). Égalité dite une fois.
export function verdictHeadline(winsA: number, winsB: number, total: number, nameA: string, nameB: string): string {
  if (winsA > winsB) return `${winsA}/${total} métriques en faveur de ${nameA}`
  if (winsB > winsA) return `${winsB}/${total} métriques en faveur de ${nameB}`
  return `égalité ${winsA}–${winsB} sur ${total} métriques`
}

// shouldLogScale — l'écraseur 738 : bascule log seulement si honnête —
// le log ne trace pas les zéros (min>0 exigé) et ne sert qu'à forte
// dynamique (max/min ≥ 20). Sinon linéaire + note qui nomme l'écraseur.
export function shouldLogScale(vals: (number | null | undefined)[]): boolean {
  const xs = vals.filter((v): v is number => typeof v === 'number' && Number.isFinite(v) && v > 0)
  if (xs.length < 2) return false
  return Math.max(...xs) / Math.min(...xs) >= 20
}
// asArray — garde anti-écran-d'erreur : un payload {error} ou un objet là où
// un tableau est attendu ne doit jamais atteindre .filter/.map. Vu en prod :
// /api/results?run= vide répond {available:false} SANS groups, le repli
// `j.groups??j??[]` laissait passer l'objet → O.filter is not a function.
export function asArray<T>(v: unknown): T[] {
  return Array.isArray(v) ? (v as T[]) : []
}
