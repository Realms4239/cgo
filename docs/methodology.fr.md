# Méthodologie de mesure — Meteolink (cgo)

Document orienté mémoire : questions de recherche, plan expérimental, conditions
de validité et chaîne de provenance. Références code : `pkg/model`,
`pkg/campagne`, `pkg/metrics`, `pkg/results`.

## 1. Questions de recherche

1. **QR1 — Effet de l'AQM** : à conditions de lien identiques, comment le choix
   du qdisc (`pfifo_fast` vs `fq_codel` vs `cake`) affecte-t-il la latence de
   queue (p95), la mortalité des petites transactions et le goodput ?
2. **QR2 — Interaction avec le contrôle de congestion** : ce effet dépend-il du
   CC de l'envoyeur (`cubic` vs `bbr`) ?
3. **QR3 — Robustesse aux conditions dégradées** : l'avantage de l'AQM tient-il
   sur les profils P2/P3 (satellite/mobile dégradé) ?
4. **QR4 — Traduction opérationnelle** : quel levier concret recommander à un
   exploitant, et avec quelle confiance mesurée (quarantaine, répétitions) ?

## 2. Profils de lien (pkg/model, valeurs exactes)

| Profil | Capacité | Délai | Gigue | Perte |
|---|---|---|---|---|
| **P1** (fibre/bon fixe) | 80 Mbit/s | 20 ms | 2 ms | 0 % |
| **P2** (accès dégradé) | 20 Mbit/s | 100 ms | 15 ms | 0,5 % |
| **P3** (satellite/mobile) | 5 Mbit/s | 600 ms | 30 ms | 1 % |

Ces profils sont des **défauts surchargeables** par import (`/api/profile/import`),
ce qui permet de caler la matrice sur un lien réel audité au préalable
(`pkg/audit` → `data/link_audit.csv`).

## 3. Plan factoriel

Matrice complète (**Tableau 3**) : profils × qdiscs × CC × répétitions.

- facteurs : 3 profils × 3 qdiscs (`pfifo_fast`, `fq_codel`, `cake`) × 2 CC
  (`cubic`, `bbr`) × 3 répétitions (défaut) = **36 événements** ;
- plan réduit (P2 seul) : 3 × 2 × 3 = **18 événements** ;
- exécution séquentielle `StartMatrixWithID` : une cellule = un événement
  `run_id/event_id` identifié, écrit dans `data/runs/<run_id>/aqm_eval.csv`
  (ordre de colonnes = Tableau 7) ; reprise possible via le writer (cellules
  déjà vues sautées) ; arrêt coopératif par `context`.

## 4. Protocole d'un événement

Chaque cellule suit **baseline 30 s → charge 120 s → récupération 30 s**
(`pkg/model` : `BaselineSec`, `ChargeSec`, `RecupSec`) :

- **baseline** : liens inactifs — stabilité de référence (porte G6) ;
- **charge** : bulk saturant (déploiement de file) + sondes ping et small
  concurrentes ; mesure de `rtt_p50/p95`, `small_p95`, `deadline_ok_pct`,
  `bulk_goodput_mbps`, `drops`, `retransmissions`, `wasted_bytes`,
  `cost_ar_per_h`, `cpu_pct` ;
- **récupération** : vidange de la file — la rapidité du retour à la baseline
  discrimine les qdiscs actifs.

Le shaper applique le profil : netem en root (délai/gigue/perte), qdisc étudié
empilé en enfant — reset-then-apply à chaque cellule pour éviter tout cumul.

## 5. Portes G0–G7 : conditions de validité

Une cellule n'entre dans l'analyse que si ses préconditions tiennent. Chaque
événement reçoit un `gate_status` : `valid`, `degraded`, `invalid`.

| Porte | Condition de validité |
|---|---|
| G0 | cible de mesure atteignable |
| G1 | bulk effectivement démarré |
| G2 | sondes productives (données reçues) |
| G3 | latence plausible (ordre de grandeur du profil) |
| G4 | goodput cohérent avec la capacité |
| G5 | pas de lignes dupliquées dans le CSV |
| G6 | baseline stable |
| G7 | CPU de la passerelle non saturé (le shaper ne fausse pas la mesure) |

## 6. Quarantaine et taxonomie de sévérité

Les cellules hors bornes sont **mises en quarantaine** (`results.Scan` →
`Quarantined`), jamais silencieusement corrigées : le compte apparaît dans
`/api/integrity` et les exports. La sévérité d'un symptôme est classée :

- **latence** small p95 : normale < 40 ms ; dégradée 40–100 ms ; critique > 100 ms ;
- **drops** : 0 ; occasionnels (> 0) ; massifs (> 100) ;
- **deadline** (respect de l'échéance p95) : ≥ 95 % ; 80–95 % ; < 80 % ;
- **goodput** bulk : ≥ 50 % de la capacité ; 20–50 % ; < 20 % ;
- **équité** (JFI) : ≥ 0,95 ; 0,8–0,95 ; < 0,8.

Les médianes par cellule (profil, qdisc, CC) et la cellule `Best` sortent de
`results.Scan` ; le seuil de deadline est paramétrable (`deadline_ms` de la
campagne, `metrics.DeadlineOKPct`).

## 7. Chaîne de provenance

1. Écriture brute des événements → `aqm_eval.csv` (ordre fixe, Tableau 7) ;
2. **gel** du run (`Freeze("config")`) + **manifest SHA-256** — immuabilité ;
3. `ProvenanceSHA` / `ProvenanceHash8` (8 premiers hex du sha256 du dernier
   `aqm_eval.csv`) — **triple provenance** : figures SVG, drawer du dashboard
   et rapports exportés portent le même `hash8`, vérifiable par
   `GET /api/integrity`.

Toute affirmation chiffrée d'un mémoire peut donc être rejouée : le `hash8`
identifie sans ambiguïté le jeu de données qui l'a produite.

## 8. Éthique de mesure — de l'écart mesuré à la prescription

Les **prescriptions ne sortent pas d'une table figée** : elles sortent d'écarts
**mesurés** porteurs de provenance. La boucle est :

> mesurer (campagne) → constater l'écart (médianes, gates, quarantaine) →
> prescrire le levier (qdisc/capacité, `hardware/translate`) → appliquer
> (`shape`) → **re-mesurer** pour vérifier l'effet.

Trois garde-fous : (i) aucune recommandation sans donnée gelée — les endpoints
répondent `available: false` plutôt que d'inventer (`F8 : honest empty`) ;
(ii) les cellules invalides sont exposées, pas masquées ; (iii) l'audit d'un
lien réel (`pkg/audit`) précède l'import d'un profil, pour que la matrice
mesure le terrain et non une abstraction.
