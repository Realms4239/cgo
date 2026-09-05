# Guide opérateur — Meteolink 1.2.2

Ce guide suit la boucle opérateur, de bout en bout :

**mesure → prescription → application → re-mesure.**

Règle d'or : toute recommandation sort d'un **écart mesuré** (seuils de
sévérité, table d'écart de la comparaison), jamais d'une intuition. On
audite le lien, on applique une prescription de façonnage, on
re-mesure, et on ne conclut que sur des chiffres.

L'interface est en français et s'organise en quatre vues :

| Vue | Rôle |
|---|---|
| **Campagne de mesure** | auditer le lien, lancer les campagnes |
| **Tableau live (Temps réel)** | observer la mesure en cours |
| **Résultats + Comparaison** | médianes par cellule, épinglage A/B, exports |
| **Provenance & archives (Intégrité)** | SHA-256, vérification des manifestes |

> **Mode observation (Windows)** : sous Windows, Meteolink tourne en
> mode observation. Vous pouvez auditer le lien (audit client-side) et
> consulter résultats, tableau live, provenance et journal. Les
> fonctions qui exigent le noyau Linux — campagne, façonnage du bord,
> surveillance continue — renvoient **HTTP 501** avec un message clair.
> Le démarrage sélectionne le mode via `--mode observe|full|auto`
> (défaut `auto` : Linux = complet, Windows = observation).
> `GET /api/health` renvoie `{ok, version, mode}` — vérifiez-y le mode.

---

## 1. Auditer le lien

**Quoi faire.** Dans la vue **Campagne de mesure**, panneau *Audit lien
accessible* : renseigner le site, le type de lien (`fiber`, `5g`, `4g`,
`vsat`, `other`), le fournisseur, la durée et la cible. Équivalent CLI :

```bash
cgo audit --link-type 5g --site "Site X" --duration 300
```

**Ce qu'on obtient.** Une ligne ajoutée à `data/link_audit.csv` : le
profil réel du lien (latence, gigue, perte observées vers la cible), plus
la **note bufferbloat** (`bloat_grade` A+..F, `bloat_delta_ms`,
`bloat_verdict` — bandes Waveform) et la **latence de travail en
3 cases** : idle, montée chargée (mesurées), descente chargée (vide
honnête « — après campagne download » tant qu'aucune campagne download
n'a gelé le sens inverse — le manque s'affiche, il ne se cache pas).
Le bouton **AUDIT → PROFIL** transforme le dernier audit en profil
rejouable sur le banc (`POST /api/audit/toprofile`).

**Ce que ça signifie.** C'est la référence avant toute prescription :
on ne façonne jamais un bord sans avoir mesuré ce que le lien fait à
vide. Cet audit est disponible même en **mode observation** (Windows) —
c'est le point d'entrée de la boucle.

## 2. Préparer le banc

**VM propre sans SSH (première fois).** Le kit diagnostique la cause exacte : `cgo kit doctor` teste le port 22 et l'auth séparément, puis `cgo kit ensure` tente boot + découverte d'IP. Sur une VM fraîche, ouvrir la console de l'hyperviseur une fois : `sudo apt install -y openssh-server && sudo systemctl enable --now ssh`, puis autoriser la clé de l'hôte : `mkdir -p ~/.ssh && echo '<clé publique>' >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys`. Ensuite tout passe par le kit.

**Quoi faire.** Sur la machine Linux de mesure, vérifier la santé de
l'environnement :

```bash
cgo doctor
```

`cgo doctor` vérifie : le mode OS, la présence de `tc` dans le PATH,
les droits CAP_NET_ADMIN (ou root), la disponibilité de BBR, et `ping`.

**Ce qu'on obtient.** Un rapport clair de ce qui manque avant de
pouvoir lancer une campagne.

**Ce que ça signifie.** Sans `tc` ni CAP_NET_ADMIN, le façonnage et les
campagnes ne peuvent pas tourner — autant le savoir avant, pas au
milieu d'une mesure. Sous Windows, cette étape se limite à vérifier le
mode via `GET /api/health`.

## 3. Lancer une campagne

**Quoi faire.** Vue **Campagne de mesure** : sélectionner les profils
et les répétitions, puis `DÉMARRER` → `CONFIRMER ?`. Chaque événement
enchaîne baseline 30 s → charge 120 s → récupération 30 s. Un arrêt est
gracieux : relancer la même campagne reprend après les événements déjà
terminés.

**Campagnes filtrées (sous-matrice).** Le formulaire UI lance la matrice
pleine (tous qdiscs × toutes CC) ; pour rejouer une cellule ciblée sans
54 min de matrice complète, passer les axes explicitement (API/CLI) :

```bash
cgo run --profiles P2 --qdiscs cake --cc bbr --reps 1 --direction down
# ou : POST /api/run/start {"profiles":["P2"],"qdiscs":["cake"],"ccs":["bbr"],"reps":1,"direction":"down"}
```

Axes vides = matrice pleine ; axes inconnus = refus explicite.

**Sens de charge (download / RRUL).** `up` (défaut) = upload
client→sink ; `down` = download source→client (shaper côté serveur) ;
`both` = RRUL séquentiel (matrice up gélée puis matrice down enchaînée).
Dans **Résultats**, les cellules download portent le badge **↓** et ne
se comparent jamais aux cellules up (voir §6).

**Ce qu'on obtient.** Les lignes figées dans
`data/runs/<run_id>/aqm_eval.csv`, une ligne par cellule
`{profil, qdisc, cc, direction}`.

**Ce que ça signifie.** La campagne produit la matrice de mesure qui
alimente Résultats et Comparaison. Impossible en mode observation
(HTTP 501) — le banc doit être sous Linux.

## 4. Façonner le bord

**Quoi faire.** Deux leviers, dans le panneau de façonnage :

- **AQM + capacité** : `qdisc` parmi `cake`, `fq_codel`,
  `pfifo_fast`, `none`, et capacité de **1 à 1000 Mbit/s**.
- **Conditions du lien** : délai **0–600 ms**, gigue **0–100 ms**,
  perte **0–10 %**.

La limite haute de capacité et les bornes des conditions viennent des
**Réglages** (voir §8) — le serveur reste autoritaire.

**Ce qu'on obtient.** Un bord façonné (netem + AQM) sur lequel la
prochaine mesure est re-productible.


## 5. Surveiller en continu

**Quoi faire.** Activer la surveillance : `POST /api/watch {on}`. Elle
ping + charge un petit objet **toutes les 500 ms** (cadence réglable,
`watchMs`).

**Ce qu'on obtient.** Une série continue de RTT et de bons débits sur
petits objets, visible en direct sur le **Tableau live (Temps réel)**.

**Ce que ça signifie.** La surveillance révèle les dégradations
transitoires (gigue, pertes, queues AQM) qu'une campagne par blocs de
30/120 s peut lisser. Désactiver avec `POST /api/watch {off}`.

## 6. Comparer

**Quoi faire.** Vue **Résultats + Comparaison** : épingler deux
cellules **A/B par clic sur une cellule** `{profil, qdisc, cc}` (les
cellules download `↓` portent leur sens : up et down ne se comparent
jamais, `GET /api/results/delta?cell=P|q|cc|dir` refuse le mélange).

**Ce qu'on obtient.** Une **table d'écart** entre A et B : p95 RTT,
small-object, goodput, pertes, coût Ar/h — plus un **verdict** et des
exports **CSV + MD** (boutons Résultats/Intégrité). La **recette sh**
(`GET /api/report/export?format=sh`, API uniquement) rejoue la meilleure
cellule gelée de chaque profil en script `tc` (`aqm-recipe.sh` : reset +
netem aux conditions du profil + shaper au goodput mesuré ×0,9 — même
règle que la suggestion CLI, sans `sudo` préfixé).

**Ce que ça signifie.** C'est ici que la prescription naît : l'écart
mesuré entre deux cellules justifie — ou non — de changer le façonnage.
Pas d'écart mesuré, pas de recommandation.

## 7. Télécharger le constat

**Quoi faire.** Depuis Résultats : **« Figer l'avant »** pour capturer
l'état de référence avant une modification, puis après re-mesure
**« Télécharger le constat »**.

**Ce qu'on obtient.** Un document d'archive (avec provenance SHA-256)
qui fait foi : avant / après, écarts, verdict.

**Ce que ça signifie.** Le constat boucle le cycle mesure →
prescription → application → re-mesure et le rend auditable par un
tiers.

## 8. Réglages

Les **10 paramètres** sont persistés en `localStorage` ; les bornes
serveur restent **autoritaires** (une valeur hors bornes est rejetée à
l'application, pas à la saisie).

| Paramètre | Défaut | Rôle |
|---|---|---|
| `warnMs` | 40 | seuil latence « dégradé » (ms) |
| `critMs` | 100 | seuil latence « critique » (ms) |
| `deadlineMs` | 1000 | deadline small p95 (ms), passée à run/start |
| `ringMax` | 1800 | fenêtre du tableau live (points) |
| `watchMs` | 500 | cadence de la surveillance (ms) |
| `shapeCap` | 20 | capacité du bord par défaut (Mbit/s) |
| `linkDelayMs` | 20 | délai du lien — levier (ms) |
| `linkJitterMs` | 2 | gigue — levier (ms) |
| `linkLossPct` | 0 | perte — levier (%) |
| `target` | `1.1.1.1` | cible de mesure |

**Palier tarifaire** (état **serveur**, pas `localStorage`) : le select
**Réglages** liste `GET /api/cost/tiers` et fige le forfait actif via
`POST /api/cost/tier` — tous les coûts affichés/calculés suivent
(en-mémoire : redémarrage = défaut `yas-month-4.5gb`). La fibre est ~10×
moins chère au Go que le mobile : le palier change le verdict
économique.

**Sévérité absolue** (uniforme sur toutes les cartes, indépendante des
réglages locaux) :

| Métrique | ok | warn | crit |
|---|---|---|---|
| Latence | ≤ 40 ms | ≤ 100 ms | > 100 ms |
| Drops | 0 | 1–100 | > 100 |
| Deadline (small p95) | ≥ 95 % | ≥ 80 % | < 80 % |
| Goodput vs capacité | ≥ 50 % | ≥ 20 % | < 20 % |
| JFI (équité) | > 0,95 | ≥ 0,8 | < 0,8 |

## 9. Journal

**Quoi faire.** Consulter le journal opérateur : `GET /api/events`, ou
la **bande des 12 derniers événements** affichée sous le Tableau live.

**Ce qu'on obtient.** Un anneau de **50 événements** : démarrages de
campagne, applications de façonnage, refus 409, bascules de
surveillance.

**Ce que ça signifie.** Le journal est la mémoire d'exploitation : il
permet de rattacher un résultat à la séquence exacte d'actions qui l'a
produit — indispensable pour un constat crédible.

