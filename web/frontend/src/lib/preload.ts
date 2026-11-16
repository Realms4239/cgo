// Le Labo est différé : le Cockpit seul entre dans le premier rendu. Le
// préchauffage au survol/focus suit la même table que `lazy()` dans App.tsx —
// un seul endroit nomne le chunk, donc un mode ajouté ne peut pas être oublié ici.
import type { ModeId } from '../store/ui'

const IMPORTERS: Partial<Record<ModeId, () => Promise<unknown>>> = {
  labo: () => import('../views/LaboView'),
}

export function preloadMode(mode: ModeId): void {
  const load = IMPORTERS[mode]
  if (load) void load()
}
