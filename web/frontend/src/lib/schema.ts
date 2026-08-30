// Schéma des paramètres — servi par GET /api/schema (source unique: le
// serveur valide dessus, le client rend les inputs dessus). Repli local si
// le serveur est joignable mais l'endpoint absent (vieux binaire).
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

export const LOCAL_FALLBACK: Param[] = [
  { key: 'reps', label: 'Répétitions', min: 1, max: 5, default: 3, unit: '×', desc: 'répétitions de chaque cellule' },
  { key: 'deadline_ms', label: 'Deadline small p95', min: 200, max: 5000, default: 1000, unit: 'ms', step: 100, desc: 'objectif de latence qui voyage avec la campagne' },
  { key: 'target', label: 'Cible de mesure', max: 64, default: '1.1.1.1', desc: 'hôte sondé' },
  { key: 'audit_duration_s', label: "Durée d'audit", min: 10, max: 600, default: 30, unit: 's', step: 10, desc: "durée de l'audit du lien" },
  { key: 'warn_ms', label: 'Seuil dégradé', min: 10, max: 200, default: 40, unit: 'ms', step: 5, desc: 'seuil jaune' },
  { key: 'crit_ms', label: 'Seuil critique', min: 50, max: 500, default: 100, unit: 'ms', step: 5, desc: 'seuil rouge' },
  { key: 'ring_max', label: 'Fenêtre live', min: 600, max: 3600, default: 1800, unit: 'pts', step: 100, desc: 'historique du tableau live' },
  { key: 'watch_ms', label: 'Cadence surveillance', min: 200, max: 2000, default: 500, unit: 'ms', step: 100, desc: 'intervalle des sondes' },
  { key: 'shape_capacity_mbps', label: 'Capacité du bord', min: 1, max: 1000, default: 20, unit: 'Mbit/s', desc: 'capacité appliquée au bord' },
  { key: 'link_delay_ms', label: 'Délai du lien', min: 0, max: 600, default: 20, unit: 'ms', desc: 'délai du levier' },
  { key: 'link_jitter_ms', label: 'Gigue', min: 0, max: 100, default: 2, unit: 'ms', desc: 'gigue du levier' },
  { key: 'link_loss_pct', label: 'Perte', min: 0, max: 10, default: 0, unit: '%', step: 0.1, desc: 'perte du levier' },
  { key: 'burst_seconds', label: 'Durée du burst', min: 2, max: 10, default: 4, unit: 's', desc: "durée d'un burst de test" },
]

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
