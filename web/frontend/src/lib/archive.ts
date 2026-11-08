// Archive des preuves — fenêtre de rendu et mots scellés uniquement.
// La liste ne porte que ce que le dossier scellé prouve ; une ligne sans
// dossier est annoncée comme telle avant le clic, jamais colorée.
export type ArchiveRow = Record<string, string>

export interface WindowInput {
  total: number
  scrollTop: number
  rowHeight: number
  viewportHeight: number
  overscan: number
}

// La fenêtre suit le haut visible puis étire jusqu'à l'overscan du bas ;
// au-delà des 179 lignes elle se recadre pour ne jamais dépasser la fin.
export function computeWindow(input: WindowInput): { start: number; end: number } {
  const total = Math.max(0, Math.trunc(input.total))
  const rowHeight = Math.max(1, input.rowHeight)
  const viewportHeight = Math.max(0, input.viewportHeight)
  const overscan = Math.max(0, Math.trunc(input.overscan))

  if (total === 0) return { start: 0, end: 0 }

  const scrollTop = Math.max(0, input.scrollTop)
  const count = Math.ceil(viewportHeight / rowHeight)
  const windowSize = count + overscan * 2

  let start = Math.max(0, Math.floor(scrollTop / rowHeight) - overscan)
  const end = Math.min(total, Math.ceil((scrollTop + viewportHeight) / rowHeight) + overscan)
  if (start >= end) start = Math.max(0, end - windowSize)

  return { start, end }
}

// Ce que le dossier scellé autorise à dire. Absent = absent : trois mots
// seulement, et « — » quand la source n'existe pas.
export function archiveStatus(row: ArchiveRow): string {
  if (row.has_report !== 'true') return '—'
  switch (row.verdict_level) {
    case 'green':
      return 'Valide'
    case 'amber':
    case 'red':
      return 'Dégradé'
    case 'neutral':
      return 'Observation seule'
    default:
      return '—'
  }
}

export function archiveVerdictWord(row: ArchiveRow): string {
  if (row.has_report !== 'true') return 'Non noté'
  const word = row.verdict_word?.trim()
  return word ? word : 'Non noté'
}

// YYYY-MM-DD HH:mm — l'archive parle locale, pas UTC ambiguë.
export function formatArchiveDate(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return '—'
  const date = new Date(trimmed)
  if (Number.isNaN(date.getTime())) return '—'
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
