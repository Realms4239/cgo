export interface SynGroup {
  profile: string
  qdisc: string
  cc: string
  deadline_median?: number
  deadline_valid_n?: number
  small_p95_median?: number
}

export interface Headline {
  simple?: SynGroup
  best?: SynGroup
}

// Le constat d'accueil ne sélectionne que des cellules P2 réellement
// exploitables : une ligne sans échantillon valide ne devient jamais un chiffre.
export function pickHeadline(groups: SynGroup[]): Headline {
  const p2 = groups.filter(g =>
    g.profile === 'P2' &&
    (g.deadline_valid_n ?? 0) > 0 &&
    Number.isFinite(g.deadline_median),
  )
  if (p2.length === 0) return {}
  const simple = p2.find(g => g.qdisc === 'pfifo_fast' && g.cc === 'bbr') ?? p2.find(g => g.qdisc === 'pfifo_fast')
  const best = [...p2].sort((a, b) => (b.deadline_median ?? 0) - (a.deadline_median ?? 0))[0]
  return { simple, best }
}

export type NextMode = 'audit' | 'bench' | 'evidence'

export function nextStep(running: boolean, runs: number | undefined): {
  title: string
  body: string
  cta: string
  mode: NextMode
} {
  if (running) {
    return {
      title: 'La campagne tourne',
      body: 'Suivez la charge, les cartes et les portes en direct. Le gel s’écrit cellule par cellule.',
      cta: "Voir l'observation",
      mode: 'bench',
    }
  }
  if (!runs) {
    return {
      title: 'Aucun corpus lisible',
      body: 'Lancez une première campagne sur le banc, ou vérifiez le chemin data/runs avant de conclure.',
      cta: "Ouvrir le banc d'essai",
      mode: 'bench',
    }
  }
  return {
    title: 'Auditer un lien ou relire le corpus',
    body: `${runs} runs sont gelés. Un audit de 60 s ajoute une ligne de terrain sans toucher au réseau de production.`,
    cta: 'Auditer le lien',
    mode: 'audit',
  }
}
