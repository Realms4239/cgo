// Une seule description courte par profil de lien, partagée par le banc et
// le classement — plus de table locale divergente.
export const PROFILE_SHORT: Record<string, string> = {
  P1: 'fibre 80 Mb/s, 20 ms',
  P2: '4G chargée 20 Mb/s, 100 ms',
  P3: 'VSAT 5 Mb/s, 600 ms',
  P4: 'Starlink 100 Mb/s, 40 ms',
}

export const profileShort = (id: string): string => PROFILE_SHORT[id] ?? id
