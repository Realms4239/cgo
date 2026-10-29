// Progression du banc d'essai — trois sections, jamais trois onglets.
export type BenchStage = 'config' | 'observe' | 'results'

// Une campagne qui démarre ouvre l'observation ; une campagne qui s'arrête
// ouvre les résultats. Un replay ouvre l'observation sans toucher au reste.
export function stageAfterCampaign(previous: BenchStage, running: boolean): BenchStage {
  if (running) return 'observe'
  return previous === 'observe' ? 'results' : previous
}

export function stageAfterReplay(previous: BenchStage, replay: boolean): BenchStage {
  return replay ? 'observe' : previous
}

// Configurer est toujours visible ; Observer apparaît dès qu'une exécution a
// eu lieu ; Résultats seulement quand la section est atteinte.
export function sectionOpen(stage: BenchStage, which: BenchStage): boolean {
  if (which === 'config') return true
  if (which === 'observe') return stage !== 'config'
  return stage === 'results'
}
