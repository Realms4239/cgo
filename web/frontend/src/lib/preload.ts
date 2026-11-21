// Préchauffage de la seule destination différée : le laboratoire porte
// l'instrument (ECharts + boucle rAF). Un seul endroit nomme le chunk, donc la
// bascule, le clavier et la palette ne peuvent pas le désynchroniser.
// Deux modes seulement — aucun identifiant historique ici.
import type { ModeId } from '../store/ui'

const IMPORTERS: Partial<Record<ModeId, () => Promise<unknown>>> = {
  labo: () => import('../views/LaboView'),
}

export function preloadMode(mode: ModeId): void {
  const load = IMPORTERS[mode]
  if (load) void load()
}
