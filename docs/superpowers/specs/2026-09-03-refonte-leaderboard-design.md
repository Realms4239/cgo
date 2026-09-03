# Refonte leaderboard — spec design (2026-09-03)

Approche A « DeepSWE sombre », standard propreté maximal.
Références : artificialanalysis.ai (rigueur-index), deepswe.datacurve.ai
(densité + incertitude + limites), Kaggle benchmarks (export + versions + CI).

## Décisions verrouillées

1. Périmètre : refonte totale de l'app, sans réserve. `book/`, `thesis-v1.docx`,
   `shots-thesis/`, figures : INTOUCHÉS (les chiffres ne changent pas).
2. IA : live en landing, rail 5 panneaux conservé (campagne, live,
   resultats, integrite + import).
3. Profondeur : tout, un passage. Thème sombre conservé. Zéro dépendance.
4. Aucune métrique hors thèse (pas d'indices composites).

## Section 1 — Résultats (APPROUVÉE)

- Bandeau benchmark : runs · lignes · valides · quarantaine · hash8 ·
  updated + sélecteur version run (`/api/replay/list` → `/api/results?run=`,
  défaut = tous runs, comportement actuel inchangé)
  + lien changelog (`/api/events` surfacé).
- Tableau dense : # rang | cellule mono | deadline % | small p95 (med [IQR])
  | RTT p95 (med [IQR]) | goodput | coût | quar. | n | Δ run précédent.
- Backend additif : `Group.Smallp95IQR` (l'IQR RTT existe déjà, jamais
  affiché). n et IQR visibles : l'incertitude s'affiche, ne se cache pas.
- Best ★, tri, filtres, verdict, scatter compromis : conservés.
- Tiroir « Méthode & limites » repliable (portes G0–G7, médianes, n=1).
- Export CSV/MD remonté en tête façon « Download leaderboard ».
- Sélecteurs e2e inchangés : `.data-table`, `[data-testid]`, rank.

## Section 2 — Intégrité (APPROUVÉE, avec micro-route)

- Bandeau preuve resserré (une ligne mono, pas des cartes).
- Table des runs : run | lignes | valides | quar. | résultats · rejouer.
  Tri chronologique inverse (aujourd'hui croissant). Données :
  champ additif `breakdown` dans `/api/integrity` (un seul appel, pas N+1).
  Pas de hash8 par run (inexistant dans les gels — le hash8 global reste
  en tête).
- Table quarantaine réelle via NOUVELLE route `GET /api/quarantine?run=`
  (lecture `quarantine.json` gelés, même garde anti-traversal que run/rows ;
  `run` vide = tous runs, récents d'abord).
- Recommandations matérielles : fond inchangé, recartées.
- Figures : régénération conservée, hash8 en légende des SVG.
- Replay : boutons conservés vers Live.

## Section 3 — Live, Campagne, transverse (APPROUVÉE)

- Live : structure inchangée, MetricCards resserrées (unités + seuils du
  schéma), bannière + journal conservés, contrôles (shape/burst/watch)
  inchangés. Timeline : branchée sur `phase`/`phase_total_s` réels, ou
  retirée si impossible honnêtement (aujourd'hui timestamps simulés).
- Campagne : cockpit conservé ; matrice complète en tableau
  profil × qdisc × CC avec état par cellule (attente/en cours/terminée —
  « skippée » inobservable depuis les données : abandonnée, pas simulée)
  depuis `live.event_id` + `total_events`.
- Transverse : tokens existants uniquement, tabulaire partout, français,
  responsive 390/1366/1920, sélecteurs `[data-panel]`, `[data-metric]`
  intacts, bundle ≤ budget (315,8 KB gz actuel).

## Hors scope

Thème clair, nouvelles libs, nouvelles métriques, retouche livrables,
changement de contrats API existants (tout ajout est additif et documenté
dans `docs/api.md`).

## Vérification

`npm run typecheck`, `vitest run`, `npm run build`, tour Edge 5 vues +
specs existantes (smoke, wall, chart-probe, compare-probe, archives),
console zéro erreur. Commits `fix(frontend):` / `feat(frontend):` en français.

---

## Backlog grill Q1–Q20 (décisions verrouillées 2026-09-03, implémentation phasée)

Références vérifiées : Waveform (note = pire écart idle vs down-loaded vs
up-loaded ; A+<5/A<30/B<60/C<200/D<400/F≥400 ms, source waveform.com),
Flent RRUL (up+down simultanés + ping + flux temps réel), FCC MBA, RFC
8290/7928/8033.

### Phase A — réalisme mesure (local, sans redéploiement bench)

- **Q8** profils asymétriques : `capacity_down/up_mbps` (défaut up=down,
  rétrocompatible).
- **Q2** filtrage sous-matrice qdisc/CC (API + CLI) — les flags --qdiscs/--cc
  du handoff n'existent pas ; indispensable aux cellules n=1.
- **Q15** note A+..F : moteur de grade (bandes Waveform verbatim citées) +
  note audit immédiate (idle/loaded déjà gelés) + gel `rtt_base_p50/p95`
  (nouvelles colonnes, lecteurs par nom OK, vieux runs = note indisponible
  honnête).
- **Q16** panneau triple latence idle/up-loaded/down-loaded (vides honnêtes
  avant Q9).
- **Q10** marquage DSCP EF (small) / BE (bulk) via IP_TOS.
- **Q11** paramètre `flows` (défaut 1 ; JFI réel à 4).
- **Q12** `loss_burst` optionnel (netem gemodel, défaut off — vérifier kernel).
- **Q13** export script tc (`/api/report/export?format=sh`, args déjà connus).
- **Q6** skip journalisé (event + statut `skipped`).
- **Q20** références RFC/FCC/Flent/Waveform dans methodology + verdicts.
- **Q1** P1–P4 = classes ; écart Starlink en limites. **Q3** durées uniformes
  30/120/30. **Q5** QDI chargé documenté tel quel. **Q4** plafond réel du
  banc à vérifier (P1/pfifo invalid en série) ; règle invalid/degraded gardée.
- **Q14** alerting : perspective post-soutenance (NON implémenté).

### Phase B — download (après Q9 : bulk inversé + shaper émission serveur)

- **Q9a** bulk inversé + cellules download ; **Q17** cellule combinée
  RRUL-style (up+down+ping, grade sur pire écart) ; grade down-loaded.

### Phase C — ops bench (campagne P2 en cours terminée + backup + déploy)

- Déploiement binaire neuf, `kit schedule` (**Q19**, timer systemd via kit),
  campagnes filtrées P1/P4/P3 (+download), STOP si contradiction deck
  (60,2/96,2/98,1 % · QDI 57→13 · R 92,9/71,5).
