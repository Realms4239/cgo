// Dictionnaire unique des micro-explications.
// Une ligne par terme, français clair, zéro jargon auto-référentiel: chaque
// élément essentiel de chaque vue peut expliquer ce qu'il montre.
export const EXPLAIN: Record<string, string> = {
  // profils
  P1: 'Profil fibre: 80 Mbit/s, délai 20 ms, gigue 2 ms, sans perte.',
  P2: 'Profil 4G/5G chargé: 20 Mbit/s, délai 100 ms, gigue 15 ms, perte 0,5 %.',
  P3: 'Profil VSAT: 5 Mbit/s, délai 600 ms, gigue 30 ms, perte 1 % (importable).',
  // files d'attente (AQM)
  pfifo_fast: "File simple premier entré, premier sorti — la référence sans AQM: la file grossit sans limite et la latence explose sous charge.",
  fq_codel: 'AQM qui garde la file courte en marquant les paquets des flux lents — bonnet anti-bufferbloat par défaut du noyau Linux.',
  cake: 'AQM complet (fq_codel + gestion par flux + contrôle de débit) — le plus efficace contre le bufferbloat sur un lien d\'accès.',
  none: 'Retire le façonnage: le bord retrouve sa file d\'origine.',
  // contrôles de congestion
  cubic: 'CC par défaut de Linux: il cherche la capacité en augmentant la fenêtre, ralentit sur perte.',
  bbr: 'CC de Google: il mesure le chemin (débit + RTT) et tient la file du réseau courte — insensible au bufferbloat.',
  // portes
  G0: 'La cible répond aux pings — sans elle, rien n\'est mesurable.',
  G1: 'Le transfert de masse a bien démarré — sans charge, pas de comparaison.',
  G2: 'Les sondes produisent des valeurs pendant la charge.',
  G3: 'La latence mesurée est cohérente avec le profil du lien.',
  G4: 'Le débit mesuré correspond à la capacité du profil.',
  G5: 'Aucune ligne dupliquée dans le CSV gelé.',
  G6: 'La baseline avant charge est stable — la comparaison est loyale.',
  G7: 'La machine n\'était pas saturée pendant la mesure.',
  // métriques
  rtt_p95: 'Latence aller-retour, 95e percentile: 95 % des pings sont plus rapides. La métrique du bufferbloat.',
  rtt_p50: 'Latence médiane: la valeur typique ressentie.',
  small_p95: 'Latence de petits objets (télémétrie, alertes) au 95e percentile — c\'est LE trafic critique à protéger.',
  bulk_goodput: 'Débit utile du transfert de masse, mesuré au récepteur.',
  drops: 'Palets perdus dans la file du bord pendant la charge.',
  wasted: 'Octets retransmis à cause des pertes — de la capacité gaspillée.',
  cost_ar_per_h: "Coût horaire du gaspillage en Ariary — palier Yas Net Month 4,5 Go (25 000 Ar, 5 556 Ar/Go) par défaut. Le forfait change tout : Ye'low One 1 000 Ar/Go, FTTH 490 Ar/Go. GET /api/cost/tiers liste les paliers réels (docs/data-prices.md).",
  deadline_ok: 'Part des petits objets arrivés sous la deadline choisie.',
  QDI: 'Écart p95 − médiane de latence: plus il est petit, plus le lien est régulier.',
  JFI: 'Équité de partage du lien entre flux (0–1): 1 = parfaitement équitable.',
  deadline: 'Objectif de latence pour les petits objets — il voyage avec la campagne.',
  capacity: 'Capacité du bord façonné — calquez-la sur le plan ISP réel.',
  delay: 'Délai du lien appliqué par le levier (netem).',
  jitter: 'Variation du délai appliquée par le levier.',
  loss: 'Perte appliquée par le levier (0 = aucun discard volontaire).',
  surveillance: 'Sondes légères en continu (ping + petits objets, sans bulk): le mur reste vivant hors campagne.',
  burst: 'Un transfert bref avec le CC choisi, à travers le bord façonné: comparez CUBIC et BBR en quelques secondes.',
  audit: 'Mesure votre lien réel depuis ce poste: latence, petits objets, débit. Non intrusif, aucun droit requis.',
  constat: 'La preuve exportée: CSV/JSON avec les médianes, l\'écart et la prescription.',
  run_rows: 'Chaque ligne est une cellule de la matrice, gelée et vérifiable par SHA-256.',
}

export const explain = (term: string): string => EXPLAIN[term] ?? ''
