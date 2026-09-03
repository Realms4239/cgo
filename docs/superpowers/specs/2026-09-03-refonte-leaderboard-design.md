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
- Table des runs : run | lignes | valides | quar. | hash8 | résultats ·
  rejouer. Tri chronologique inverse (aujourd'hui croissant). Données :
  champ additif `breakdown` dans `/api/integrity` (un seul appel, pas N+1).
- Table quarantaine réelle via NOUVELLE route `GET /api/quarantine?run=`
  (lecture `quarantine.json` gelés, même garde anti-traversal que run/rows).
- Recommandations matérielles : fond inchangé, recartées.
- Figures : régénération conservée, hash8 en légende des SVG.
- Replay : boutons conservés vers Live.

## Section 3 — Live, Campagne, transverse (APPROUVÉE)

- Live : structure inchangée, MetricCards resserrées (unités + seuils du
  schéma), bannière + journal conservés, contrôles (shape/burst/watch)
  inchangés. Timeline : branchée sur `phase`/`phase_total_s` réels, ou
  retirée si impossible honnêtement (aujourd'hui timestamps simulés).
- Campagne : cockpit conservé ; matrice complète en tableau
  profil × qdisc × CC avec état par cellule (attente/encours/gelée/skippée)
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
