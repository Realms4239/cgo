// Les quatre profils de référence restent visibles, lisibles et complets.
// Les valeurs sont celles du référentiel produit ; elles ne sont pas inventées.
export const PROFILE_SHORT: Record<string, string> = {
  P1: 'fibre 80 Mb/s · 20 ms RTT · référence fibre',
  P2: '4G dégradée 20 Mb/s · 100 ms RTT · cellulaire contrainte',
  P3: 'VSAT géostationnaire 5 Mb/s · 600 ms RTT · satellite',
  P4: 'Orbite basse 100 Mb/s · 40 ms RTT · LEO',
}

export const profileShort = (id: string): string => PROFILE_SHORT[id] ?? id

export const PROFILE_TILES = [
  { id: 'P1', title: 'P1 — Fibre dédiée', spec: '80 Mbit/s · 20 ms RTT', loss: '0 % perte', use: 'Référence fibre' },
  { id: 'P2', title: 'P2 — 4G dégradée', spec: '20 Mbit/s · 100 ms RTT', loss: '0,5 % perte', use: 'Cellulaire contrainte' },
  { id: 'P3', title: 'P3 — VSAT géostationnaire', spec: '5 Mbit/s · 600 ms RTT', loss: '1 % perte', use: 'Satellite' },
  { id: 'P4', title: 'P4 — Orbite basse', spec: '100 Mbit/s · 40 ms RTT', loss: 'jitter orbital', use: 'LEO' },
] as const
