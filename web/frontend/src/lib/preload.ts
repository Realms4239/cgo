// Les trois destinations non-attractées sont différées : le tableau de bord
// seul entre dans le premier rendu. Le préchauffage au survol/focus suit la
// même table que `lazy()` dans App.tsx — un seul endroit nomne le chunk, donc
// un mode ajouté ne peut pas être oublié ici.
import type { ModeId } from '../store/ui'

const IMPORTERS: Partial<Record<ModeId, () => Promise<unknown>>> = {
  audit: () => import('../views/AuditView'),
  bench: () => import('../views/BenchView'),
  evidence: () => import('../views/EvidenceView'),
}

export function preloadMode(mode: ModeId): void {
  const load = IMPORTERS[mode]
  if (load) void load()
}
