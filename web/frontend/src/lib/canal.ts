// canal — rafales (bursty) sur canal propre. Le serveur refuse la campagne
// AVANT toute mutation (400 nommant le profil fautif) : l'UI applique la même
// règle pour ne jamais offrir un choix que le backend rejetterait.
export type ChannelModel = 'clean' | 'bursty'

const BURSTY_PROFILES = ['P2', 'P3'] as const

export interface CanalGate {
  allowed: boolean
  // Vide quand autorisé ; sinon nomme les profils fautifs ET la règle.
  reason: string
}

export function canalGate(profiles: string[]): CanalGate {
  if (profiles.length === 0) {
    return { allowed: false, reason: 'aucun profil sélectionné — rafales sur P2/P3 uniquement' }
  }
  const hors = profiles.filter(p => !(BURSTY_PROFILES as readonly string[]).includes(p))
  if (hors.length > 0) {
    return {
      allowed: false,
      reason: `bursty invalide pour ${hors.join(', ')} — rafales sur P2/P3 uniquement (choisir clean)`,
    }
  }
  return { allowed: true, reason: '' }
}

// Le modèle effectif retombe sur clean quand la porte refuse bursty : le POST
// ne part jamais avec une combinaison que le serveur refuserait.
export function effectiveChannelModel(model: ChannelModel, gate: CanalGate): ChannelModel {
  return model === 'bursty' && !gate.allowed ? 'clean' : model
}
