# Méthodologie de mesure — Meteolink (cgo)

Document orienté mémoire : questions de recherche, plan expérimental, conditions
de validité et chaîne de provenance. Références code : `pkg/model`,
`pkg/campagne`, `pkg/metrics`, `pkg/results`.

## 0. Ancrage méthodologique

La chaîne de mesure s'ancre dans des standards et outils publics :

- **RFC 8290** ( fq_codel) et **RFC 7928** (cake) — les disciplines testées,
  spécifiées ; **RFC 8033** (PIE) pour la famille AQM alternative.
- **DSCP EF, RFC 3246** — la sonde petit objet est marquée Expedited
  Forwarding (banc Linux) : la mesure porte la classe temps réel que l'AQM
  doit protéger, pas un flux best-effort anonyme.
- **Flent / RRUL** (Toke Høiland-Jørgensen et al.) — la référence
  académique du test de latence sous charge (flux montants + descendants
  simultanés + ping) ; le banc en applique la forme mono-sens (charge
  montante), la forme combinée étant une perspective documentée.
- **Waveform Bufferbloat Test** (waveform.com/tools/bufferbloat) — la note
  A+..F de l'audit et des cellules reprend ses bandes verbatim
  (A+ <5 ms · A <30 · B <60 · C <200 · D <400 · F ≥400) sur l'écart de
  latence moyenne idle→chargée, pire sens mesuré ; source citée dans le
  verdict.
- **FCC Measuring Broadband America** — précédent réglementaire de la
  mesure de latence sous charge auprès du grand public.

## 0.1 Asymétrie TCP mesurée (résultat du banc, 2026-09-04)

Sur lien 100 ms / perte 0,5 %, un flux TCP unique plafonne par l'équation
de Mathis (~2,5 Mb/s cubic) **dès que la perte touche la boucle** — y
compris la perte des seuls ACK (sens download : données intactes en
descente, ACK perdus à 0,5 % en montée → BBR mesuré à ~2,5 Mb/s contre
~19 Mb/s en montée où les ACK sont propres). Conséquences verrouillées
dans le code :

- la CC de la cellule pilote l'ÉMETTEUR du sens mesuré (protocole `D:<cc>`
  côté sink, `DialWithCC` côté client) — jamais la CC par-défaut de l'hôte ;
- la porte G4 juge le sens mesuré (CapUp/CapDown) avec un plancher
  direction-aware : en download mono-flux, le goodput est horlogé par les
  ACK (qdisc-indépendant) — le plancher (1 % de CapDown) ne détecte que le
  tuyau mort, et **le verdict download se joue sur la latence**
  (small_p95, deadline), pas sur le débit ;
- le download multi-flux (`flows` > 1) agrège au-delà du plafond ACK
  mono-flux et montre l'équité inter-flux (JFI gelé per-flow).

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
| **P4** (Starlink LEO) | 100 Mbit/s | 40 ms | 20 ms | 0,3 % |

P4 fige la classe LEO (débit élevé, délai bas par rapport au VSAT
géostationnaire, mais gigue marquée des handovers et perte ponctuelle).
Deux dimensions optionnelles affinent un profil (`pkg/model.Profile`,
importables par `/api/profile/import`) :

- **asymétrie** (`capacity_up_mbps`) : scinde le sens montant (4G 20/5,
  VSAT 5/1 — la règle, pas l'exception). Absent → up = down
  (symétrique historique) ; `CapUp()` plafonne à la capacité déclarée.
  Le shaper et la porte G4 jugent le sens mesuré (voir §4–§5) ;
- **perte en rafales** (`loss_burst_{p,r,h,k}`, Gilbert-Elliott via
  `netem gemodel`, kernel du banc 6.8 vérifié) : la vraie vie
  mobile/satellite où la perte arrive en salves. Zéro/absent = perte
  uniforme (comportement historique).

Ces profils sont des **défauts surchargeables** par import (`/api/profile/import`),
ce qui permet de caler la matrice sur un lien réel audité au préalable
(`pkg/audit` → `data/link_audit.csv`).

### 2.1 Distribution empirique (`.dist`)

Le délai d'un profil peut être rejoué depuis une trace d'audit plutôt que
depuis le couple déclaré `delay/jitter`. La calibration ne retient que la
phase **baseline** de `data/audits/<id>/telemetry.csv` : les échantillons
sous charge décrivent l'AQM testé, pas le lien.

```
cgo calibrate --audit audit-XXXX --profile P2
```

Le profil doit exister (`pkg/model`, import compris) et sa médiane RTT doit
dépasser son `delay_ms`. La commande écrit `data/calibrations/P2.dist`
(table netem native, 1024 entrées, échelle 8192) et `P2.json` (provenance
scellée : audit source, SHA-256 de la télémétrie et du `.dist`, site, type
de lien, percentiles source et reconstruits).

Convention : `delay_ms` est un **RTT**. Le mode dist pose `plancher/2` sur
chaque saut et la table empirique sur le seul saut mesuré (`w-edge` en
montée, `c-wan` en descente), donc RTT ≈ plancher + résidu. Le mode est
opt-in :

```
CGO_DELAY_MODEL=dist CGO_CALIBRATION_DIR=$PWD/data/calibrations   cgo run --bench dualns --profiles P2 --direction up
```

Une calibration absente, périmée (profil modifié) ou dont le `.dist` ne
correspond plus au sceau arrête l'événement **avant** toute mutation qdisc :
jamais de repli synthétique. `both` est refusé (une trace = un sens). Le
scalaire reste le défaut. Les traces `SMOKE-TEST` locales ne sont pas une
preuve terrain et aucun `.dist` qui en dérive n'est cité.

## 3. Plan factoriel

Matrice complète (**Tableau 3**) : profils × qdiscs × CC × répétitions.

- facteurs : 4 profils × 3 qdiscs (`pfifo_fast`, `fq_codel`, `cake`) × 2 CC
  (`cubic`, `bbr`) × 3 répétitions (défaut) = **72 événements** ;
- plan réduit (P2 seul) : 3 × 2 × 3 = **18 événements** ;
- **sous-matrice** (`StartMatrixFiltered`, `qdiscs`/`ccs` de
  `POST /api/run/start` ou `--qdiscs` + `--cc` ensemble en CLI) : axes
  explicites pour rejouer une cellule n=1 en 1 événement au lieu de la
  matrice pleine. Axes vides ou valeurs inconnues = refus explicite, pas
  de matrice fantôme ;
- **sens de charge** (`direction` : `up` défaut historique, `down`,
  `both`) : `up` = upload client→sink (shaper sur l'émission cliente
  veth-c) ; `down` = download source→client (shaper sur l'émission
  serveur veth-s via le netns `cgo-srv`, voir §4) ; `both` = RRUL
  séquentiel — matrice up gélée puis matrice down enchaînée dans la même
  campagne (le banc n'a qu'un shaper actif à la fois ; Flent/RRUL
  simultané reste la référence, §0). La direction est gelée avec chaque
  ligne et scinde les groupes d'analyse (voir §6) : les sens ne se
  comparent jamais ;
- exécution séquentielle `StartMatrixWithID` : une cellule = un événement
  `run_id/event_id` identifié, écrit dans `data/runs/<run_id>/aqm_eval.csv`
  (ordre de colonnes = Tableau 7) ; reprise possible via le writer (cellules
  déjà vues sautées) ; arrêt coopératif par `context`.

## 4. Protocole d'un événement

Chaque cellule suit **baseline 30 s → charge 120 s → récupération 30 s**
(`pkg/model` : `BaselineSec`, `ChargeSec`, `RecupSec`) :

- **baseline** : liens inactifs — stabilité de référence (porte G6) ;
  la latence au repos (`rtt_base_p50_ms`, `rtt_base_p95_ms`) est gelée
  avec la ligne : la note bufferbloat Waveform (écart chargé − repos)
  exige la référence, pas seulement la charge ;
- **charge** : bulk saturant (déploiement de file) + sondes ping et small
  concurrentes ; mesure de `rtt_p50/p95`, `small_p95`, `deadline_ok_pct`,
  `bulk_goodput_mbps`, `drops`, `retransmissions`, `wasted_bytes`,
  `cost_ar_per_h`, `cpu_pct` ;
- **récupération** : vidange de la file — la rapidité du retour à la baseline
  discrimine les qdiscs actifs.

Le shaper applique le profil : netem en root (délai/gigue/perte, rafales
incluses) sur veth-c, qdisc étudié empilé en enfant — reset-then-apply à
chaque cellule pour éviter tout cumul.

- **sonde EF** : le petit objet part marqué DSCP EF (46, RFC 3246) sur le
  banc Linux (`SmallClient`) — la mesure porte la classe temps réel que
  l'AQM doit protéger (diffserv exercée, pas supposée). Hors Linux, la
  sonde part best-effort (documenté, mesure EF non interprétable) ;
- **multi-flux** (`flows` > 1, `BulkN` câblé) : N connexions concurrentes,
  octets per-flow — l'équité inter-flux (JFI de Jain, gelé 0..100 avec la
  ligne, 0 = mono-flux/n.a.) ne se mesure que per-flow : un total
  identique peut cacher la famine d'un flux. En download, le multi-flux
  agrège au-delà du plafond ACK mono-flux (voir §0.1) ;
- **download** (`direction: down`) : le shaper vit sur l'émission serveur
  (veth-s, netns `cgo-srv`) — le congestionnement du download naît côté
  source ; le netem reste sur veth-c (le délai s'applique aux deux sens).
  La CC de la cellule pilote l'émetteur (protocole sink `D:<cc>`,
  `DialWithCC`), jamais la CC par-défaut de l'hôte — sinon bbr/cubic sont
  indiscernables. Les compteurs (`tc -s`) sont lus sur le saut du sens
  mesuré : sonder veth-c en download ne compterait que les ACK (gel à 0,0
  malgré des Mo reçus, prouvé 2026-09-04) ;

## 5. Portes G0–G7 : conditions de validité

Une cellule n'entre dans l'analyse que si ses préconditions tiennent. Chaque
événement reçoit un `gate_status` : `valid`, `degraded`, `invalid`.

| Porte | Condition de validité |
|---|---|
| G0 | cible de mesure atteignable |
| G1 | bulk effectivement démarré |
| G2 | sondes productives (données reçues) |
| G3 | latence plausible (ordre de grandeur du profil) |
| G4 | goodput cohérent avec la capacité **du sens mesuré** : montant → `CapUp`, plancher 50 % ; descendant → `CapDown`, plancher 1 % (tuyau mort seul — le goodput down mono-flux est horlogé par les ACK, qdisc-indépendant, voir §0.1). Le verdict download se joue sur la latence, pas le débit |
| G5 | pas de lignes dupliquées dans le CSV |
| G6 | baseline stable |
| G7 | CPU de la passerelle non saturé (le shaper ne fausse pas la mesure) |

## 6. Quarantaine et taxonomie de sévérité

Les cellules hors bornes sont **mises en quarantaine** (`results.Scan` →
`Quarantined`), jamais silencieusement corrigées : le compte apparaît dans
`/api/integrity` et les exports. Seules les lignes `invalid` sont exclues
des médianes — `degraded` reste exploitable (inclus). La sévérité d'un symptôme est classée :

- **latence** small p95 : normale < 40 ms ; dégradée 40–100 ms ; critique > 100 ms ;
- **drops** : 0 ; occasionnels (> 0) ; massifs (> 100) ;
- **deadline** (respect de l'échéance p95) : ≥ 95 % ; 80–95 % ; < 80 % ;
- **goodput** bulk : ≥ 50 % de la capacité ; 20–50 % ; < 20 % ;
- **équité** (JFI) : ≥ 0,95 ; 0,8–0,95 ; < 0,8.

Les médianes par cellule (profil, qdisc, CC, **direction**) et la cellule
`Best` sortent de `results.Scan` ; le seuil de deadline est paramétrable
(`deadline_ms` de la campagne, `metrics.DeadlineOKPct`). Règles de
regroupement : la clé inclut la direction (défaut `up` pour les runs
historiques sans la colonne) — up et down ne fusionnent jamais ; les runs
`run-smoke*` (fixtures du pipeline, pas des mesures) sont exclus du corpus
des deux côtés Go/JS (parité prouvée).

## 7. Chaîne de provenance

1. Écriture brute des événements → `aqm_eval.csv` (ordre fixe,
   `model.AQMEvalHeader` — **23 colonnes** : `run_id`, `event_id`,
   `profile`, `qdisc`, `cc`, `direction`, `repetition`, `rtt_p50_ms`,
   `rtt_p95_ms`, `qdi_ms`, `voip_r`, `jfi_pct`, `small_p95_ms`,
   `deadline_ok_pct`, `bulk_goodput_mbps`, `drops`, `retransmissions`,
   `wasted_bytes`, `cost_ar_per_h`, `cpu_pct`, `rtt_base_p50_ms`,
   `rtt_base_p95_ms`, `gate_status`). Lecture **par nom**
   (`results.ReadAQM`), jamais positionnelle : les schémas 17/18/19/21/22
   des runs antérieurs restent lisibles, colonnes absentes = mesure
   indisponible honnête ;
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

## 9. Façonnage descendant : IFB (Linux) et file LAN (RouterOS)

Le façonnage egress ne corrige pas un téléchargement : quand le paquet arrive,
il a déjà traversé le lien. Linux le résout par un **IFB** : l'ingress de
l'interface WAN est redirigé par `matchall`/`mirred` vers `ifb0`, puis un qdisc
racine ordinaire met le paquet en file sur `ifb0` — côté réception.

### 9.1 Banc `dualns`

- `ifb0` vit dans `ns_client` ; le redirect et le qdisc ingress vivent sur
  `c-cli`, l'interface d'accès mesurée.
- Le CAKE descendant de `e-wan` est **absent** en mode ingress. S'il restait,
  le goulot serait en amont de l'IFB et la file de l'IFB ne se formerait
  jamais.
- Le CAKE montant de `e-srv` reste en place : il n'est pas sur le chemin
  descendant.
- Le CAKE par défaut sur `ifb0` est un confort de banc (`up --ingress on`) ;
  une cellule de campagne remplace la racine `ifb0` puis la supprime en
  sortie. Après une campagne, `check` peut donc signaler l'absence de CAKE
  sur `ifb0` jusqu'au prochain `up --ingress on`.
- Le mode du banc et celui de la campagne doivent coïncider :
  `CGO_SHAPE_MODE=ingress` sert aussi de défaut à `up`/`check`/`show` du banc
  lorsque `--ingress` est absent, pour que le pré-contrôle de `cgo run` voie
  la même topologie.
- Le banc ne modélise pas le LAN derrière le client : le téléchargement
  aboutit dans le même namespace que les sondes. C'est un raccourci de banc
  documenté, pas une topologie CPE complète.

### 9.2 Comptabilité

- Les pertes à l'IFB sont des pertes **hôtes**, après traversée du lien
  émulé ; elles ne prouvent pas une perte au lien distant et ne libèrent pas
  de capacité.
- En mode `ingress`, CAKE compte les paquets jetés comme transférés et
  maintient au moins deux paquets par flux. Un compteur de pertes ingress
  n'est donc pas comparable tel quel à un compteur egress ; toute table qui
  les compare doit le dire.
- Le RTT mesuré inclut le lien netem et la file IFB. Aucun horodatage noyau
  ne sépare les deux : on ne présente pas `RTT − repos` comme un temps de
  séjour dans la file du lien distant.
- Le mode `ingress` est opt-in (`CGO_SHAPE_MODE=ingress`) et limité à
  `bench=dualns` + `direction=down`. `up` garde le chemin egress ; `both` est
  refusé plutôt qu'approché.

### 9.3 RouterOS

RouterOS n'a pas d'IFB. Le script descendant est une **file sur l'interface
LAN/privée**, là où le trafic descendant sort vers les clients. Ce n'est pas
un équivalent octet pour octet de l'IFB Linux : le point de passage diffère.
Les objets portent `cgoingress-*` et `comment="cgo-managed-ingress"`, disjoints
du nettoyage montant (`^cgo-`, `comment="cgo-managed"`).

### 9.4 Preuve

Les tests Windows prouvent la génération des commandes, les champs du rapport
et l'ordre fermé-échec du préflight. Ils ne prouvent pas qu'un noyau 6.8
accepte les commandes ni que l'IFB jette réellement sous charge. Cette preuve
est la tâche d'acceptation Linux, distincte et bloquée.

## 10. BBRv3 : noyau, étiquette et matrice 3×3

La section 9 est réservée au façonnage descendant IFB (plan distinct). Cette
section décrit la troisième ligne de contrôle du plan factoriel.

### 10.1 Pourquoi un noyau

BBRv3 n'est pas un module de contrôle de congestion : Google le publie comme
une arborescence de noyau complète (`google/bbr`), avec des changements dans
le cœur TCP. Le noyau 6.8 d'Ubuntu n'expose que `bbr`, qui est **BBRv1**.
La décision #8 du projet reste donc vraie : `bbr` signifie BBRv1. BBRv3
entre sous une étiquette additive, `bbr3`, et seulement sur un hôte dont le
noyau enregistre ce nom.

### 10.2 Provisionnement opt-in

`kit/install-bbr3.sh` construit la branche `bbr-v3-2026-09-16-01` (Linux 7.1)
sur la VM, avec `CONFIG_TCP_CONG_BBR=y` et `CONFIG_TCP_CONG_BBR3=y`. Les deux
contrôles vivent alors dans le même noyau : la comparaison v1/v3 ne change
qu'une variable. Le script n'est jamais appelé par `cgo run`, `cgo deploy` ou
`cgo kit bootstrap` ; il s'exécute sur la console ou via `cgo kit ssh`. Le
redémarrage est manuel par défaut. Le noyau précédent reste dans GRUB.

### 10.3 Matrice

```bash
./kit/run-dualns.sh up
sudo ip netns exec ns_client env CGO_BENCH=dualns ~/cgo/cgo-linux run \
  --bench dualns --direction up --profiles P2 \
  --qdiscs pfifo_fast,fq_codel,cake \
  --cc cubic,bbr,bbr3 \
  --reps 3 --flows 4 --target 10.30.0.2
```

Neuf cellules, trois répétitions, 27 événements. Le sens montant reproduit
la campagne du mémoire. Ce jeu de runs est **autonome** : il ne se fusionne
pas numériquement avec les cellules du corpus 6.8 ni avec le Tableau 15,
qui viennent d'un autre banc et d'un autre régime de flux. Chaque manifeste
gele le noyau (`kernel`) avec la configuration ; une ligne `cc=bbr3` n'est
citable qu'avec ce noyau.

### 10.4 Portée

La matrice n'active ni ECN ni L4S : les cellules comparent le comportement
sous perte et délai, pas les signaux ECN de BBRv3. Le façonnage descendant
reste celui de la section 9 ; la matrice ne dépend pas de l'IFB.
