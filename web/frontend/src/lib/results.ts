// results — helpers PURS de la vue Résultats (hors React, testés).
// Le classement garde un seul encodage par cellule et trois groupes de puces
// étiquetés qui wrappent : la rangée nowrap de vingt puces est morte.

export type ChipKey = 'profile' | 'qdisc' | 'cc'

export const CHIP_LABELS: Record<ChipKey, string> = {
  profile: 'PROFIL',
  qdisc: 'FILE',
  cc: 'CC',
}

export interface ChipGroup {
  key: ChipKey
  label: string
  // « tous » en tête, puis les valeurs distinctes triées (une seule source).
  values: string[]
}

// Trois groupes étiquetés, chacun wrappant sur plusieurs lignes.
export function chipGroups(distinct: (k: ChipKey) => string[]): ChipGroup[] {
  return (['profile', 'qdisc', 'cc'] as const).map(key => ({
    key,
    label: CHIP_LABELS[key],
    values: ['tous', ...distinct(key)],
  }))
}

// Un seul encodage par cellule : la primaire porte la barre + la moustache
// (valeur et son intervalle) ; les secondaires n'affichent que valeur + unité.
export function cellEncoding(primary: boolean): 'bar' | 'plain' {
  return primary ? 'bar' : 'plain'
}

// ± demi-intervalle — null quand l'API ne fournit pas d'intervalle (jamais
// une moustache inventée autour d'une valeur seule).
export function halfInterval(lo: number, hi: number): number | null {
  return hi > lo ? (hi - lo) / 2 : null
}

// Texte de la dispersion : « ±0.4 » au-dessous de 10, « ±12 » au-dessus.
export function fmtHalf(half: number): string {
  return `±${half.toFixed(half >= 10 ? 0 : 1)}`
}
