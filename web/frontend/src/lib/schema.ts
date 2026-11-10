// Schéma des paramètres — servi par GET /api/schema (source unique: le
// serveur valide dessus, le client rend les inputs dessus).
//
// Aucun repli local : un contrat devenu illisible se rend honnêtement
// (liste vide, champs sans bornes), jamais avec des bornes fausses.
export type Param = {
  key: string
  label: string
  min?: number
  max?: number
  default: any
  unit?: string
  step?: number
  enum?: string[]
  desc: string
}

let cache: Param[] | null = null

export async function loadSchema(): Promise<Param[]> {
  if (cache) return cache
  try {
    const j = await fetch('/api/schema').then(r => r.json())
    cache = (j?.params ?? []) as Param[]
  } catch { cache = [] }
  return cache
}

export function param(key: string, params: Param[]): Param | undefined {
  return params.find(p => p.key === key)
}
